//go:build integration

// Package main_test holds the integration smoke test for the server's
// composition root.
//
// The test boots the real router via buildRouter() with a Postgres
// testcontainer, then probes /health/ready to assert the wiring works
// end to end. This is a single high-value test: it covers the entire
// dependency injection block, every use case constructor, every
// handler registration, and the middleware chain — paths that are
// not exercised by any unit test.
//
// The signal handling and graceful shutdown loop in runServer remain
// uncovered here. Process-level testing of those paths would require
// spawning a real binary and sending SIGTERM, which is out of scope
// for this smoke test.
package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/ianadou/smo/infrastructure/persistence"
	"github.com/ianadou/smo/infrastructure/persistence/postgres"
)

// init disables the Ryuk reaper container before any test starts —
// see infrastructure/persistence/postgres/repositories/integration_helpers_test.go
// for the rationale (Ryuk fails on Fedora 43 + Docker 29; deferred
// Terminate handles cleanup).
func init() {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

var sharedPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, pool, err := setupBootContainer(ctx)
	if err != nil {
		log.Fatalf("boot smoke setup: %v", err)
	}
	sharedPool = pool

	code := m.Run()

	pool.Close()
	if termErr := container.Terminate(ctx); termErr != nil {
		log.Printf("terminate container: %v", termErr)
	}
	os.Exit(code)
}

func setupBootContainer(ctx context.Context) (testcontainers.Container, *pgxpool.Pool, error) {
	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("smo_boot_test"),
		tcpostgres.WithUsername("smo"),
		tcpostgres.WithPassword("smo-boot-password"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return container, nil, err
	}

	pool, err := postgres.Connect(ctx, connStr)
	if err != nil {
		return container, nil, err
	}

	if migErr := postgres.RunMigrations(pool, persistence.MigrationsFS); migErr != nil {
		return container, nil, migErr
	}

	return container, pool, nil
}

// TestBuildRouter_HealthReady_Returns200_WhenAllDepsAreUp is the
// composition-root smoke test: it builds the full Gin router with a
// real Postgres pool (no Redis — that's an optional dependency, ADR
// 0002), then probes /health/ready. A 200 response means every
// constructor and every handler registration in buildRouter ran
// without panicking, AND the database is reachable through the wired
// HealthHandler.
func TestBuildRouter_HealthReady_Returns200_WhenAllDepsAreUp(t *testing.T) {
	if sharedPool == nil {
		t.Fatal("sharedPool is nil; TestMain did not run setupBootContainer")
	}

	router := buildRouter(sharedPool, nil, "test-jwt-secret-for-boot-smoke")
	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health/ready") //nolint:gosec // test target
	if err != nil {
		t.Fatalf("GET /health/ready: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// TestBuildRouter_HealthLive_Returns200 mirrors the above for the
// liveness endpoint, which must succeed even when the database is
// unreachable. Together the two assertions exercise both branches of
// the health handler wiring.
func TestBuildRouter_HealthLive_Returns200(t *testing.T) {
	if sharedPool == nil {
		t.Fatal("sharedPool is nil; TestMain did not run setupBootContainer")
	}

	router := buildRouter(sharedPool, nil, "test-jwt-secret-for-boot-smoke")
	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health/live") //nolint:gosec // test target
	if err != nil {
		t.Fatalf("GET /health/live: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
