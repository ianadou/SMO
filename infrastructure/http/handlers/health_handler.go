package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	dbPingTimeout    = 2 * time.Second
	cachePingTimeout = 1 * time.Second
)

// DBPinger is the small interface this handler needs to assess database
// readiness. *pgxpool.Pool implements it natively; a fake stub is used
// in unit tests. Defining the interface on the consumer side (rather
// than re-importing pgx) keeps this handler free of database-driver
// concerns and decoupled from the pgx version.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// CachePinger is the small interface this handler needs to assess cache
// readiness. *redis.Client implements it natively (Ping returns a
// *StatusCmd whose Err() reports reachability); a fake stub is used in
// unit tests.
//
// The interface is satisfied by both the concrete *redis.Client (whose
// Ping returns a *StatusCmd) via a small adapter at the wiring site,
// and by test fakes.
type CachePinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler exposes liveness and readiness probes used by Docker
// HEALTHCHECK, Dockhand, and any future load balancer or orchestrator.
//
// Liveness ("is the process up?") and readiness ("can the process serve
// real traffic?") are deliberately split: liveness is shallow and never
// fails so the container is not killed during transient backend issues,
// readiness reflects actual ability to handle requests so a load
// balancer can drain the instance when its database is unreachable.
type HealthHandler struct {
	db    DBPinger
	cache CachePinger
}

// NewHealthHandler builds a HealthHandler with the given DB pinger and
// optional cache pinger. Pass cache=nil when the cache is disabled
// (REDIS_URL unset, ADR 0002 state 1); the readiness probe will then
// report cache="disabled" without affecting the overall HTTP status.
func NewHealthHandler(db DBPinger, cache CachePinger) *HealthHandler {
	return &HealthHandler{db: db, cache: cache}
}

// Register attaches the two health endpoints under /health.
//
// Note: these endpoints are NOT under /api/v1 because they belong to
// the infrastructure layer (used by orchestrators, monitoring agents),
// not to the business contract. Their shape can evolve independently
// of the API version.
func (h *HealthHandler) Register(router *gin.Engine) {
	router.GET("/health/live", h.Live)
	router.GET("/health/ready", h.Ready)
}

// Live handles GET /health/live. Always returns 200: as long as the Go
// runtime can answer this handler, the process is alive.
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready handles GET /health/ready. Returns 200 if every external
// dependency the app needs to serve traffic is reachable, 503 if a
// hard dependency (the database) is down.
//
// The cache is a soft dependency per ADR 0002: a cache outage degrades
// performance but the app keeps serving correct results from Postgres.
// So a Redis ping failure surfaces in the body as cache="down" and
// status="degraded" but the HTTP code stays 200 — only DB failure
// flips to 503.
func (h *HealthHandler) Ready(c *gin.Context) {
	databaseStatus, dbHealthy := h.checkDatabase(c.Request.Context())
	cacheStatus := h.checkCache(c.Request.Context())

	httpStatus := http.StatusOK
	overall := "ready"

	switch {
	case !dbHealthy:
		httpStatus = http.StatusServiceUnavailable
		overall = "degraded"
	case cacheStatus == "down":
		// Soft degradation: still 200, but the body tells the truth.
		overall = "degraded"
	}

	c.JSON(httpStatus, gin.H{
		"status":    overall,
		"database":  databaseStatus,
		"cache":     cacheStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HealthHandler) checkDatabase(parent context.Context) (string, bool) {
	ctx, cancel := context.WithTimeout(parent, dbPingTimeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return "down", false
	}
	return "ok", true
}

func (h *HealthHandler) checkCache(parent context.Context) string {
	if h.cache == nil {
		return "disabled"
	}

	ctx, cancel := context.WithTimeout(parent, cachePingTimeout)
	defer cancel()

	if err := h.cache.Ping(ctx); err != nil {
		return "down"
	}
	return "ok"
}
