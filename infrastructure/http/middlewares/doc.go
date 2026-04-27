// Package middlewares contains Gin middlewares used by the SMO HTTP
// server. They are wired in cmd/server/main.go and form the request
// pipeline: request ID → structured logging → recovery → handlers.
package middlewares
