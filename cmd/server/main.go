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
	"github.com/ianadou/smo/infrastructure/clock"
	"github.com/ianadou/smo/infrastructure/http/handlers"
	"github.com/ianadou/smo/infrastructure/idgen"
	"github.com/ianadou/smo/infrastructure/persistence"
	"github.com/ianadou/smo/infrastructure/persistence/postgres"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/repositories"
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

	// run() owns all resources that need deferred cleanup (context
	// cancellation, connection pool close). main() only exits on the
	// result, which guarantees that defers inside run() actually fire
	// before the process terminates.
	if err := run(); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

// run performs the full server lifecycle: read configuration, connect
// to the database, apply migrations, build the router, and serve HTTP.
//
// It returns an error instead of calling os.Exit so that defers can
// clean up resources properly.
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

	// Connection + migrations happen before the router is built so that
	// any startup failure surfaces as a fast exit, not as a broken server.
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
	// to be an integer in [1, 65535], so it cannot contain injection chars.
	slog.Info("starting http server", "address", address)
	if runErr := router.Run(address); runErr != nil {
		return fmt.Errorf("http server stopped with error: %w", runErr)
	}
	return nil
}

// buildRouter assembles all the application dependencies and returns a
// fully configured Gin router ready to serve HTTP requests.
//
// This is the composition root: every concrete adapter is instantiated
// here, and the resulting wiring is the only place where the application
// is coupled to specific implementations (UUID, system clock, Postgres
// repository, etc.).
func buildRouter(pool *pgxpool.Pool) *gin.Engine {
	// Infrastructure adapters (concrete implementations of domain ports).
	groupRepo := repositories.NewPostgresGroupRepository(pool)
	idGenerator := idgen.New()
	systemClock := clock.New()

	// Application use cases (orchestrators that depend only on domain ports).
	createGroupUC := groupusecase.NewCreateGroupUseCase(groupRepo, idGenerator, systemClock)
	getGroupUC := groupusecase.NewGetGroupUseCase(groupRepo)

	// HTTP handlers (thin wrappers around use cases).
	groupHandler := handlers.NewGroupHandler(createGroupUC, getGroupUC)

	// Router configuration.
	router := gin.New()
	router.Use(gin.Recovery())

	// System endpoints (no /api prefix).
	// TODO: extract to infrastructure/http/handlers/health.go once we
	// add more system endpoints (readiness, metrics, etc.).
	router.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API endpoints (under /api prefix).
	api := router.Group("/api")
	groupHandler.Register(api)

	return router
}

// parsePort validates that the given raw value is a usable TCP port number
// and returns it as a string ready to be used by the HTTP server.
//
// An empty input is treated as "use the default port" and returns defaultPort
// without error. Any other input must parse as a positive integer in the
// valid TCP range [1, 65535]; otherwise an error is returned so the caller
// can fail fast at startup.
//
// TODO: move to infrastructure/config/ once the full configuration layer
// is introduced.
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
