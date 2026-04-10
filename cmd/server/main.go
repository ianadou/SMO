// Package main is the HTTP server entry point for the SMO backend.
//
// This file is the composition root of the application: the only
// place where concrete implementations are instantiated and wired
// together. Every other layer depends on interfaces (ports) and
// receives its dependencies via constructor injection.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	groupusecase "github.com/ianadou/smo/application/usecases/group"
	"github.com/ianadou/smo/infrastructure/clock"
	"github.com/ianadou/smo/infrastructure/http/handlers"
	"github.com/ianadou/smo/infrastructure/idgen"
	"github.com/ianadou/smo/infrastructure/persistence/inmemory"
)

const (
	defaultPort = "8081"
	minPort     = 1
	maxPort     = 65535
)

// errInvalidPort is returned by parsePort when the provided value is not a
// valid TCP port number.
var errInvalidPort = errors.New("invalid port")

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port, err := parsePort(os.Getenv("PORT"))
	if err != nil {
		slog.Error("invalid port configuration", "error", err)
		os.Exit(1)
	}

	router := buildRouter()

	address := ":" + port
	// #nosec G706 -- address is built from a port validated by parsePort()
	// to be an integer in [1, 65535], so it cannot contain injection chars.
	slog.Info("starting http server", "address", address)
	if runErr := router.Run(address); runErr != nil {
		slog.Error("http server stopped with error", "error", runErr)
		os.Exit(1)
	}
}

// buildRouter assembles all the application dependencies and returns a
// fully configured Gin router ready to serve HTTP requests.
//
// This is the composition root: every concrete adapter is instantiated
// here, and the resulting wiring is the only place where the application
// is coupled to specific implementations (UUID, system clock, in-memory
// repository, etc.).
//
// TODO: replace inmemory.NewGroupRepository() with a PostgreSQL-backed
// implementation in PR #16 once Docker Compose makes Postgres available
// locally.
func buildRouter() *gin.Engine {
	// Infrastructure adapters (concrete implementations of domain ports).
	groupRepo := inmemory.NewGroupRepository()
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
