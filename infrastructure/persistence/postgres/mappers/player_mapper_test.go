package mappers

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

func TestPlayerToDomain_ReturnsEntity_WhenRowIsValid(t *testing.T) {
	t.Parallel()
	row := generated.Players{
		ID:        "p-1",
		GroupID:   "g-1",
		Name:      "Alice",
		Ranking:   1200,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	player, err := PlayerToDomain(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player.Name() != "Alice" {
		t.Errorf("expected 'Alice', got %q", player.Name())
	}
	if player.Ranking() != 1200 {
		t.Errorf("expected ranking 1200, got %d", player.Ranking())
	}
}

func TestPlayerToDomain_ReturnsError_WhenNameIsEmpty(t *testing.T) {
	t.Parallel()
	row := generated.Players{
		ID: "p-1", GroupID: "g-1", Name: "", Ranking: 1000,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	_, err := PlayerToDomain(row)
	if !errors.Is(err, domainerrors.ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestPlayerToCreateParams_BuildsParams(t *testing.T) {
	t.Parallel()
	player, _ := entities.NewPlayer("p-1", "g-1", "Alice", 1000)
	createdAt := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)

	params := PlayerToCreateParams(player, createdAt)

	if params.ID != "p-1" {
		t.Errorf("expected 'p-1', got %q", params.ID)
	}
	if params.Ranking != 1000 {
		t.Errorf("expected 1000, got %d", params.Ranking)
	}
	if !params.CreatedAt.Time.Equal(createdAt) {
		t.Errorf("expected createdAt %v, got %v", createdAt, params.CreatedAt.Time)
	}
}

func TestPlayerToUpdateRankingParams_BuildsParams(t *testing.T) {
	t.Parallel()
	player, _ := entities.NewPlayer("p-1", "g-1", "Alice", 1500)
	params := PlayerToUpdateRankingParams(player)
	if params.ID != "p-1" || params.Ranking != 1500 {
		t.Errorf("unexpected params: %+v", params)
	}
}
