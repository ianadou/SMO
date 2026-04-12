package dto

import (
	"testing"

	"github.com/ianadou/smo/domain/entities"
)

func TestPlayerResponseFromEntity_MapsAllFields(t *testing.T) {
	t.Parallel()
	player, err := entities.NewPlayer("p-1", "g-1", "Alice", 1200)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	resp := PlayerResponseFromEntity(player)

	if resp.ID != "p-1" {
		t.Errorf("expected 'p-1', got %q", resp.ID)
	}
	if resp.Name != "Alice" {
		t.Errorf("expected 'Alice', got %q", resp.Name)
	}
	if resp.Ranking != 1200 {
		t.Errorf("expected 1200, got %d", resp.Ranking)
	}
}
