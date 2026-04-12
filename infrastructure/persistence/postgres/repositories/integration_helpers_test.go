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
// Ryuk fails to start on Fedora 43 with Docker 29 (known issue), and
// our deferred container.Terminate() in TestMain handles cleanup.
func init() {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

var sharedPool *pgxpool.Pool

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

	// Pre-seed fixtures that satisfy the foreign key constraints used
	// by the tests: an organizer (for groups) and a group (for matches).
	if _, seedErr := pool.Exec(ctx, `
		INSERT INTO organizers (id, email, password_hash, display_name)
		VALUES ('test-org', 'test@example.com', 'fake-hash', 'Test Organizer')
		ON CONFLICT (id) DO NOTHING
	`); seedErr != nil {
		return container, nil, seedErr
	}

	if _, seedErr := pool.Exec(ctx, `
		INSERT INTO groups (id, organizer_id, name)
		VALUES ('test-group', 'test-org', 'Test Group')
		ON CONFLICT (id) DO NOTHING
	`); seedErr != nil {
		return container, nil, seedErr
	}

	return container, pool, nil
}

// newTestGroupRepository returns a fresh PostgresGroupRepository wired
// to the shared pool, after deleting any existing groups.
//
// Note: this also re-seeds the test-group fixture because the DELETE
// above removes it. Individual tests can then reference test-group in
// their match rows without worrying about FK violations.
func newTestGroupRepository(t *testing.T) *repositories.PostgresGroupRepository {
	t.Helper()

	if sharedPool == nil {
		t.Fatal("sharedPool is nil; TestMain did not run setupTestContainer")
	}

	if _, err := sharedPool.Exec(context.Background(), "DELETE FROM groups"); err != nil {
		t.Fatalf("failed to clean groups table: %v", err)
	}

	return repositories.NewPostgresGroupRepository(sharedPool)
}

// newTestMatchRepository returns a fresh PostgresMatchRepository wired
// to the shared pool, after deleting any existing matches and ensuring
// the test-group fixture exists (recreating it if a prior group test
// deleted the groups table).
func newTestMatchRepository(t *testing.T) *repositories.PostgresMatchRepository {
	t.Helper()

	if sharedPool == nil {
		t.Fatal("sharedPool is nil; TestMain did not run setupTestContainer")
	}

	ctx := context.Background()

	if _, err := sharedPool.Exec(ctx, "DELETE FROM matches"); err != nil {
		t.Fatalf("failed to clean matches table: %v", err)
	}

	// Ensure test-group exists (a previous group test may have deleted it).
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO groups (id, organizer_id, name)
		VALUES ('test-group', 'test-org', 'Test Group')
		ON CONFLICT (id) DO NOTHING
	`); err != nil {
		t.Fatalf("failed to re-seed test-group: %v", err)
	}

	return repositories.NewPostgresMatchRepository(sharedPool)
}
