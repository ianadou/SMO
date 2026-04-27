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
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	authusecase "github.com/ianadou/smo/application/usecases/auth"
	groupusecase "github.com/ianadou/smo/application/usecases/group"
	invitationusecase "github.com/ianadou/smo/application/usecases/invitation"
	matchusecase "github.com/ianadou/smo/application/usecases/match"
	playerusecase "github.com/ianadou/smo/application/usecases/player"
	voteusecase "github.com/ianadou/smo/application/usecases/vote"
	"github.com/ianadou/smo/domain/ranking"
	bcryptauth "github.com/ianadou/smo/infrastructure/auth/bcrypt"
	jwtauth "github.com/ianadou/smo/infrastructure/auth/jwt"
	cacheredis "github.com/ianadou/smo/infrastructure/cache/redis"
	"github.com/ianadou/smo/infrastructure/clock"
	"github.com/ianadou/smo/infrastructure/http/handlers"
	"github.com/ianadou/smo/infrastructure/http/middlewares"
	"github.com/ianadou/smo/infrastructure/idgen"
	"github.com/ianadou/smo/infrastructure/persistence"
	"github.com/ianadou/smo/infrastructure/persistence/postgres"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/repositories"
	"github.com/ianadou/smo/infrastructure/token"

	rdb "github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
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
	jwtLifetime        = 24 * time.Hour

	// HTTP server timeouts. Aligned with conservative production
	// defaults: header read fast (10s) to drop slowloris, write
	// generous (30s) for slow clients on JSON, idle large to amortize
	// keep-alive across short request bursts.
	httpReadHeaderTimeout = 10 * time.Second
	httpWriteTimeout      = 30 * time.Second
	httpIdleTimeout       = 60 * time.Second

	// Maximum time the server will spend draining in-flight requests
	// after a SIGTERM/SIGINT before giving up and exiting.
	shutdownTimeout = 30 * time.Second
)

// errMissingJWTSecret is returned when JWT_SECRET is not set. We refuse
// to fall back to a hardcoded secret: an unset JWT_SECRET in production
// would silently degrade to a known-weak key.
var errMissingJWTSecret = errors.New("JWT_SECRET environment variable is required")

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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return errMissingJWTSecret
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

	// Redis: optional. Empty REDIS_URL → cache disabled (caching wrappers
	// not installed). Set + reachable → cache enabled. Set + unreachable
	// → fail boot. See docs/adr/0002-cache-aside-with-redis.md.
	redisClient, redisErr := cacheredis.Connect(ctx, os.Getenv("REDIS_URL"))
	if redisErr != nil {
		return fmt.Errorf("failed to connect to redis: %w", redisErr)
	}
	if redisClient == nil {
		slog.Info("redis cache disabled (REDIS_URL not set)")
	} else {
		slog.Info("redis cache enabled")
		defer func() { _ = redisClient.Close() }()
	}

	router := buildRouter(pool, redisClient, jwtSecret)

	address := ":" + port
	server := &http.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: httpReadHeaderTimeout,
		WriteTimeout:      httpWriteTimeout,
		IdleTimeout:       httpIdleTimeout,
	}

	return runServer(server)
}

// runServer starts the HTTP server in a goroutine and blocks until
// either ListenAndServe fails or a SIGTERM/SIGINT arrives. On signal,
// in-flight requests are drained up to shutdownTimeout. The database
// pool is closed by the deferred pool.Close() in the caller, not here:
// that keeps pool ownership single-responsibility and avoids a
// double-close when ListenAndServe errors out.
func runServer(server *http.Server) error {
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting http server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("http server stopped with error: %w", err)
		}
		return nil
	case sig := <-signals:
		slog.Info("shutdown initiated", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed, forcing close", "error", err)
		_ = server.Close()
	}
	slog.Info("shutdown complete")
	return nil
}

// buildRouter assembles all application dependencies and returns a
// fully configured Gin router ready to serve HTTP requests.
//
// This function is intentionally long: per CLAUDE.md "Dependency wiring
// is manual and explicit in cmd/server/main.go. A 50-line wiring block
// is normal and desirable — the flow of dependencies must be readable
// top-to-bottom." Extracting helpers here would scatter the wiring
// across the file and hide the composition root.
//
//nolint:funlen,cyclop // composition root — see CLAUDE.md project rules
func buildRouter(pool *pgxpool.Pool, redisClient *rdb.Client, jwtSecret string) *gin.Engine {
	// Infrastructure adapters. Group and Player are wrapped in their
	// caching variants when redisClient is non-nil; otherwise the
	// wrappers pass through (ADR 0002 state 1).
	groupRepo := cacheredis.WrapGroupRepository(repositories.NewPostgresGroupRepository(pool), redisClient)
	matchRepo := repositories.NewPostgresMatchRepository(pool)
	playerRepo := cacheredis.WrapPlayerRepository(repositories.NewPostgresPlayerRepository(pool), redisClient)
	invitationRepo := repositories.NewPostgresInvitationRepository(pool)
	voteRepo := repositories.NewPostgresVoteRepository(pool)
	organizerRepo := repositories.NewPostgresOrganizerRepository(pool)
	tokenService := token.New()
	idGenerator := idgen.New()
	systemClock := clock.New()
	passwordHasher := bcryptauth.New(bcrypt.DefaultCost)
	jwtSigner := jwtauth.New(jwtSecret, jwtLifetime)

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

	// Auth use cases.
	registerOrganizerUC := authusecase.NewRegisterOrganizerUseCase(organizerRepo, passwordHasher, idGenerator, systemClock)
	loginOrganizerUC := authusecase.NewLoginOrganizerUseCase(organizerRepo, passwordHasher, jwtSigner)

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
	//
	// Middleware chain (outer → inner):
	//   1. RequestID — assigns/propagates X-Request-ID, exposes it via
	//      context.Context for downstream use cases.
	//   2. SLogLogger — emits one structured JSON log per completed
	//      request, with the request_id field for correlation.
	//   3. Recovery — catches panics from handlers and turns them into
	//      500 responses; sits innermost so SLogLogger above logs them.
	router := gin.New()
	router.Use(middlewares.RequestID())
	router.Use(middlewares.SLogLogger(slog.Default()))
	router.Use(gin.Recovery())

	// Health endpoints sit at the root, outside /api/v1: they belong to
	// the infrastructure layer (Docker HEALTHCHECK, Dockhand, future
	// orchestrator probes), not the business contract.
	handlers.NewHealthHandler(pool).Register(router)

	// All business endpoints live under /api/v1. The /v1 prefix locks
	// in a versioned contract from day one: future breaking changes can
	// ship as /v2 alongside without breaking existing clients. The
	// rationale (URL-based vs header-based versioning) is documented in
	// docs/adr/0001-api-url-versioning.md.
	//
	// Two parallel groups are created on the same /api/v1 path: `public`
	// without auth, `protected` with the JWTAuth middleware. Each
	// handler decides on a per-route basis which group to attach to —
	// reads typically go on `public`, mutations on `protected`, and a
	// few special cases (token-authed accept invitation, cast vote)
	// stay on `public` despite being mutations.
	public := router.Group("/api/v1")
	protected := router.Group("/api/v1")
	protected.Use(middlewares.JWTAuth(jwtSigner))

	groupHandler.Register(public, protected)
	matchHandler.Register(public, protected)

	playerHandler := handlers.NewPlayerHandler(createPlayerUC, getPlayerUC, listPlayersByGroupUC, updatePlayerRankingUC)
	playerHandler.Register(public, protected)

	invitationHandler := handlers.NewInvitationHandler(createInvitationUC, getInvitationUC, listInvitationsByMatchUC, acceptInvitationUC)
	invitationHandler.Register(public, protected)

	voteHandler := handlers.NewVoteHandler(castVoteUC, getVoteUC, listVotesByMatchUC)
	voteHandler.Register(public, protected)

	authHandler := handlers.NewAuthHandler(registerOrganizerUC, loginOrganizerUC)
	authHandler.Register(public, protected)

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
