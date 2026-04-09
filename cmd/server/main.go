// Package main is the HTTP server entry point for the SMO backend.
//
// This file currently wires a minimal Gin router with a single /health
// endpoint. It will grow into the full dependency wiring location as use
// cases, repositories, and adapters are introduced.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const defaultPort = "8081"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	router := gin.New()
	router.Use(gin.Recovery())

	// TODO: extract to infrastructure/http/handlers/health.go once the
	// first real use case is introduced and the handler layer is set up.
	router.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	address := ":" + port
	slog.Info("starting http server", "address", address)

	if err := router.Run(address); err != nil {
		slog.Error("http server stopped with error", "error", err)
		os.Exit(1)
	}
}
