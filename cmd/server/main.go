// Package main is the HTTP server entry point for the SMO backend.
//
// This file currently wires a minimal Gin router with a single /health
// endpoint. It will grow into the full dependency wiring location as use
// cases, repositories, and adapters are introduced.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
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

	router := gin.New()
	router.Use(gin.Recovery())

	// TODO: extract to infrastructure/http/handlers/health.go once the
	// first real use case is introduced and the handler layer is set up.
	router.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	address := ":" + port
	// #nosec G706 -- address is built from a port validated by parsePort()
	// to be an integer in [1, 65535], so it cannot contain injection chars.
	slog.Info("starting http server", "address", address)

	if err := router.Run(address); err != nil {
		slog.Error("http server stopped with error", "error", err)
		os.Exit(1)
	}
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
