// Package main is the HTTP server entry point for the SMO backend.
//
// This file is the composition root of the application: the only
// place where concrete implementations are instantiated and wired
// together. Every other layer depends on interfaces (ports) and
// receives its dependencies via constructor injection.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	groupusecase "github.com/ianadou/smo/application/usecases/group"
	invitationusecase "github.com/ianadou/smo/application/usecases/invitation"
	matchusecase "github.com/ianadou/smo/application/usecases/match"
	playerusecase "github.com/ianadou/smo/application/usecases/player"
	voteusecase "github.com/ianadou/smo/application/usecases/vote"
	"github.com/ianadou/smo/domain/ranking"
	"github.com/ianadou/smo/infrastructure/clock"
	"github.com/ianadou/smo/infrastructure/http/handlers"
	"github.com/ianadou/smo/infrastructure/idgen"
	"github.com/ianadou/smo/infrastructure/persistence"
	"github.com/ianadou/smo/infrastructure/persistence/postgres"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/repositories"
	"github.com/ianadou/smo/infrastructure/token"
)

const (
	defaultPort = "8081"
	// #nosec G101 -- this is a development-only default used when
	// DATABASE_URL is not set; in production, the real connection
	// string (including real credentials) is always injected via the
	// DATABASE_URL environment variable from a secret store.
	defaultDatabaseURL = "postgres://smo:smo@localhost:5433/smo_dev?sslmode=disable"
	minPort            = 1
	maxPort            = 65535
	connectTimeout     = 30 * time.Second
)

// errInvalidPort is returned by parsePort when the provided value is not a
// valid TCP port number.
var errInvalidPort = errors.New("invalid port")

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	port, err := parsePort(os.Getenv("PORT"))
	if err != nil {
		return fmt.Errorf("invalid port configuration: %w", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
		slog.Info("DATABASE_URL not set, using default", "url", defaultDatabaseURL)
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	pool, err := postgres.Connect(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()
	slog.Info("database connection established")

	if migrateErr := postgres.RunMigrations(pool, persistence.MigrationsFS); migrateErr != nil {
		return fmt.Errorf("failed to apply migrations: %w", migrateErr)
	}
	slog.Info("database migrations applied")

	router := buildRouter(pool)

	address := ":" + port
	// #nosec G706 -- address is built from a port validated by parsePort()
	slog.Info("starting http server", "address", address)
	if runErr := router.Run(address); runErr != nil {
		return fmt.Errorf("http server stopped with error: %w", runErr)
	}
	return nil
}

// buildRouter assembles all application dependencies and returns a
// fully configured Gin router ready to serve HTTP requests.
func buildRouter(pool *pgxpool.Pool) *gin.Engine {
	// Infrastructure adapters.
	groupRepo := repositories.NewPostgresGroupRepository(pool)
	matchRepo := repositories.NewPostgresMatchRepository(pool)
	playerRepo := repositories.NewPostgresPlayerRepository(pool)
	invitationRepo := repositories.NewPostgresInvitationRepository(pool)
	voteRepo := repositories.NewPostgresVoteRepository(pool)
	tokenService := token.New()
	idGenerator := idgen.New()
	systemClock := clock.New()

	// Group use cases.
	createGroupUC := groupusecase.NewCreateGroupUseCase(groupRepo, idGenerator, systemClock)
	getGroupUC := groupusecase.NewGetGroupUseCase(groupRepo)

	// Ranking calculator. The default learning rate is the right
	// default for now; making it configurable is on the backlog.
	rankingCalculator, rankingErr := ranking.NewCalculator(ranking.DefaultLearningRate())
	if rankingErr != nil {
		// NewCalculator only fails on out-of-range learning rate, so a
		// failure here means the default constant itself is invalid —
		// a programming error, not a configuration issue.
		panic(fmt.Sprintf("invalid default learning rate: %v", rankingErr))
	}

	// Match use cases.
	createMatchUC := matchusecase.NewCreateMatchUseCase(matchRepo, idGenerator, systemClock)
	getMatchUC := matchusecase.NewGetMatchUseCase(matchRepo)
	listMatchesByGroupUC := matchusecase.NewListMatchesByGroupUseCase(matchRepo)
	openMatchUC := matchusecase.NewOpenMatchUseCase(matchRepo)
	markTeamsReadyUC := matchusecase.NewMarkTeamsReadyUseCase(matchRepo)
	startMatchUC := matchusecase.NewStartMatchUseCase(matchRepo)
	completeMatchUC := matchusecase.NewCompleteMatchUseCase(matchRepo)
	finalizeMatchUC := matchusecase.NewFinalizeMatchUseCase(matchRepo, voteRepo, playerRepo, rankingCalculator)

	// Player use cases.
	createPlayerUC := playerusecase.NewCreatePlayerUseCase(playerRepo, idGenerator)
	getPlayerUC := playerusecase.NewGetPlayerUseCase(playerRepo)
	listPlayersByGroupUC := playerusecase.NewListPlayersByGroupUseCase(playerRepo)
	updatePlayerRankingUC := playerusecase.NewUpdatePlayerRankingUseCase(playerRepo)

	// Invitation use cases.
	createInvitationUC := invitationusecase.NewCreateInvitationUseCase(invitationRepo, tokenService, idGenerator, systemClock)
	getInvitationUC := invitationusecase.NewGetInvitationUseCase(invitationRepo)
	listInvitationsByMatchUC := invitationusecase.NewListInvitationsByMatchUseCase(invitationRepo)
	acceptInvitationUC := invitationusecase.NewAcceptInvitationUseCase(invitationRepo, tokenService, systemClock)

	// Vote use cases.
	castVoteUC := voteusecase.NewCastVoteUseCase(voteRepo, matchRepo, idGenerator, systemClock)
	getVoteUC := voteusecase.NewGetVoteUseCase(voteRepo)
	listVotesByMatchUC := voteusecase.NewListVotesByMatchUseCase(voteRepo)

	// HTTP handlers.
	groupHandler := handlers.NewGroupHandler(createGroupUC, getGroupUC)
	matchHandler := handlers.NewMatchHandler(
		createMatchUC,
		getMatchUC,
		listMatchesByGroupUC,
		openMatchUC,
		markTeamsReadyUC,
		startMatchUC,
		completeMatchUC,
		finalizeMatchUC,
	)

	// Router configuration.
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	groupHandler.Register(api)
	matchHandler.Register(api)

	playerHandler := handlers.NewPlayerHandler(createPlayerUC, getPlayerUC, listPlayersByGroupUC, updatePlayerRankingUC)
	playerHandler.Register(api)

	invitationHandler := handlers.NewInvitationHandler(createInvitationUC, getInvitationUC, listInvitationsByMatchUC, acceptInvitationUC)
	invitationHandler.Register(api)

	voteHandler := handlers.NewVoteHandler(castVoteUC, getVoteUC, listVotesByMatchUC)
	voteHandler.Register(api)

	return router
}

func parsePort(raw string) (string, error) {
	if raw == "" {
		return defaultPort, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %q is not a number", errInvalidPort, raw)
	}

	if value < minPort || value > maxPort {
		return "", fmt.Errorf("%w: %d is out of range [%d, %d]", errInvalidPort, value, minPort, maxPort)
	}

	return strconv.Itoa(value), nil
}
