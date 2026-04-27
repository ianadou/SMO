package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const dbPingTimeout = 2 * time.Second

// DBPinger is the small interface this handler needs to assess database
// readiness. *pgxpool.Pool implements it natively; a fake stub is used
// in unit tests. Defining the interface on the consumer side (rather
// than re-importing pgx) keeps this handler free of database-driver
// concerns and decoupled from the pgx version.
type DBPinger interface {
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
	db DBPinger
}

// NewHealthHandler builds a HealthHandler with the given DB pinger.
func NewHealthHandler(db DBPinger) *HealthHandler {
	return &HealthHandler{db: db}
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
// dependency the app needs to serve traffic is reachable, 503
// otherwise. Today only the Postgres pool is checked; Redis will be
// added when the cache layer lands.
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), dbPingTimeout)
	defer cancel()

	databaseStatus := "ok"
	httpStatus := http.StatusOK
	overall := "ready"

	if err := h.db.Ping(ctx); err != nil {
		databaseStatus = "down"
		httpStatus = http.StatusServiceUnavailable
		overall = "degraded"
	}

	c.JSON(httpStatus, gin.H{
		"status":    overall,
		"database":  databaseStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
