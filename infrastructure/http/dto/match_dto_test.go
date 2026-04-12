package dto

import (
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

func TestMatchResponseFromEntity_MapsAllFields(t *testing.T) {
	t.Parallel()

	scheduledAt := time.Date(2026, 5, 1, 18, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 12, 10, 30, 0, 0, time.UTC)

	match, err := entities.NewMatch(
		"match-1", "group-1", "Friday football", "Stadium A",
		scheduledAt, entities.MatchStatusOpen, createdAt,
	)
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}

	resp := MatchResponseFromEntity(match)

	if resp.ID != "match-1" {
		t.Errorf("expected ID 'match-1', got %q", resp.ID)
	}
	if resp.GroupID != "group-1" {
		t.Errorf("expected GroupID 'group-1', got %q", resp.GroupID)
	}
	if resp.Title != "Friday football" {
		t.Errorf("expected title 'Friday football', got %q", resp.Title)
	}
	if resp.Venue != "Stadium A" {
		t.Errorf("expected venue 'Stadium A', got %q", resp.Venue)
	}
	if !resp.ScheduledAt.Equal(scheduledAt) {
		t.Errorf("expected scheduledAt %v, got %v", scheduledAt, resp.ScheduledAt)
	}
	if resp.Status != "open" {
		t.Errorf("expected status 'open', got %q", resp.Status)
	}
	if !resp.CreatedAt.Equal(createdAt) {
		t.Errorf("expected createdAt %v, got %v", createdAt, resp.CreatedAt)
	}
}
