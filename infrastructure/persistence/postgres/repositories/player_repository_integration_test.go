//go:build integration

package repositories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/repositories"
)

// newTestPlayerRepository returns a fresh PostgresPlayerRepository, cleaning
// the players table and re-seeding the test-group fixture if needed.
func newTestPlayerRepository(t *testing.T) *repositories.PostgresPlayerRepository {
	t.Helper()
	ctx := context.Background()
	if _, err := sharedPool.Exec(ctx, "DELETE FROM players"); err != nil {
		t.Fatalf("failed to clean players: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO groups (id, organizer_id, name)
		VALUES ('test-group', 'test-org', 'Test Group')
		ON CONFLICT (id) DO NOTHING
	`); err != nil {
		t.Fatalf("failed to re-seed test-group: %v", err)
	}
	return repositories.NewPostgresPlayerRepository(sharedPool)
}

func TestPostgresPlayerRepository_Save_PersistsPlayer(t *testing.T) {
	repo := newTestPlayerRepository(t)
	ctx := context.Background()

	player, err := entities.NewPlayer("player-1", "test-group", "Alice", 1000)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if saveErr := repo.Save(ctx, player); saveErr != nil {
		t.Fatalf("expected Save to succeed, got: %v", saveErr)
	}

	found, findErr := repo.FindByID(ctx, "player-1")
	if findErr != nil {
		t.Fatalf("expected to find saved player, got: %v", findErr)
	}
	if found.Name() != "Alice" {
		t.Errorf("expected name 'Alice', got %q", found.Name())
	}
}

func TestPostgresPlayerRepository_Save_ReturnsErrReferencedEntityNotFound_WhenGroupMissing(t *testing.T) {
	repo := newTestPlayerRepository(t)
	ctx := context.Background()

	player, _ := entities.NewPlayer("player-1", "nonexistent-group", "Alice", 1000)
	err := repo.Save(ctx, player)

	if !errors.Is(err, domainerrors.ErrReferencedEntityNotFound) {
		t.Errorf("expected ErrReferencedEntityNotFound, got %v", err)
	}
}

func TestPostgresPlayerRepository_FindByID_ReturnsErrPlayerNotFound_WhenMissing(t *testing.T) {
	repo := newTestPlayerRepository(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "does-not-exist")

	if !errors.Is(err, domainerrors.ErrPlayerNotFound) {
		t.Errorf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestPostgresPlayerRepository_ListByGroup_ReturnsPlayersOrderedByRanking(t *testing.T) {
	repo := newTestPlayerRepository(t)
	ctx := context.Background()

	for _, data := range []struct {
		id      string
		name    string
		ranking int
	}{
		{"p-1", "Alice", 1200},
		{"p-2", "Bob", 1000},
		{"p-3", "Carol", 1500},
	} {
		p, _ := entities.NewPlayer(entities.PlayerID(data.id), "test-group", data.name, data.ranking)
		if err := repo.Save(ctx, p); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	players, err := repo.ListByGroup(ctx, "test-group")
	if err != nil {
		t.Fatalf("expected ListByGroup to succeed, got: %v", err)
	}
	if len(players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(players))
	}
	// Verify DESC ranking order.
	if players[0].Ranking() != 1500 || players[2].Ranking() != 1000 {
		t.Errorf("expected order by ranking DESC, got %d, %d, %d",
			players[0].Ranking(), players[1].Ranking(), players[2].Ranking())
	}
}

func TestPostgresPlayerRepository_UpdateRanking_PersistsNewValue(t *testing.T) {
	repo := newTestPlayerRepository(t)
	ctx := context.Background()

	player, _ := entities.NewPlayer("player-1", "test-group", "Alice", 1000)
	if err := repo.Save(ctx, player); err != nil {
		t.Fatalf("setup: %v", err)
	}

	updated, _ := entities.NewPlayer("player-1", "test-group", "Alice", 1200)
	if err := repo.UpdateRanking(ctx, updated); err != nil {
		t.Fatalf("expected UpdateRanking to succeed, got: %v", err)
	}

	found, _ := repo.FindByID(ctx, "player-1")
	if found.Ranking() != 1200 {
		t.Errorf("expected ranking 1200, got %d", found.Ranking())
	}
}

func TestPostgresPlayerRepository_Delete_RemovesPlayer(t *testing.T) {
	repo := newTestPlayerRepository(t)
	ctx := context.Background()

	player, _ := entities.NewPlayer("player-1", "test-group", "Alice", 1000)
	if err := repo.Save(ctx, player); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := repo.Delete(ctx, "player-1"); err != nil {
		t.Fatalf("expected Delete to succeed, got: %v", err)
	}

	_, findErr := repo.FindByID(ctx, "player-1")
	if !errors.Is(findErr, domainerrors.ErrPlayerNotFound) {
		t.Errorf("expected ErrPlayerNotFound after delete, got %v", findErr)
	}
}
