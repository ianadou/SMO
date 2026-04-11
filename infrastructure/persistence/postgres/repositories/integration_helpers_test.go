//go:build integration

package repositories_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/ianadou/smo/infrastructure/persistence"
	"github.com/ianadou/smo/infrastructure/persistence/postgres"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/repositories"
)

// init disables the Ryuk reaper container before any test starts.
//
// Ryuk is a "garbage collector" container that testcontainers normally
// spins up alongside test containers to clean them up if the Go process
// crashes. On some Linux configurations (notably Fedora 43 with Docker
// 29), the Ryuk container fails to start its health endpoint and the
// whole testcontainers setup aborts.
//
// We can safely disable it because TestMain has a deferred
// container.Terminate() that handles the normal cleanup path. The only
// case Ryuk would catch that we don't is a hard process kill (SIGKILL),
// which is rare in CI and easy to clean up manually if it ever happens.
func init() {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

// sharedPool is the pgxpool used by all integration tests in this package.
// It is initialized once by setupTestContainer and reused across tests for
// performance: spinning up a Postgres container takes ~5-10 seconds, so
// doing it per-test would make the suite painfully slow.
var sharedPool *pgxpool.Pool

// setupTestContainer starts a single Postgres container, applies the goose
// migrations, and pre-seeds a test organizer to satisfy foreign key
// constraints. It is called once via TestMain and the resulting pool is
// stored in sharedPool.
func setupTestContainer(ctx context.Context) (testcontainers.Container, *pgxpool.Pool, error) {
	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("smo_test"),
		tcpostgres.WithUsername("smo"),
		tcpostgres.WithPassword("smo-test-password"),
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

	// Pre-seed an organizer so tests can create groups without manually
	// inserting an organizer row in every single test.
	if _, seedErr := pool.Exec(ctx, `
		INSERT INTO organizers (id, email, password_hash, display_name)
		VALUES ('test-org', 'test@example.com', 'fake-hash', 'Test Organizer')
		ON CONFLICT (id) DO NOTHING
	`); seedErr != nil {
		return container, nil, seedErr
	}

	return container, pool, nil
}

// newTestRepository returns a fresh PostgresGroupRepository wired to the
// shared pool, after deleting any existing groups so each test starts
// from a clean slate.
func newTestRepository(t *testing.T) *repositories.PostgresGroupRepository {
	t.Helper()

	if sharedPool == nil {
		t.Fatal("sharedPool is nil; TestMain did not run setupTestContainer")
	}

	if _, err := sharedPool.Exec(context.Background(), "DELETE FROM groups"); err != nil {
		t.Fatalf("failed to clean groups table: %v", err)
	}

	return repositories.NewPostgresGroupRepository(sharedPool)
}
