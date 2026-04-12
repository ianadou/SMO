//go:build integration

package repositories_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestPostgresMatchRepository_Save_PersistsMatch(t *testing.T) {
	repo := newTestMatchRepository(t)
	ctx := context.Background()

	scheduledAt := time.Date(2026, 5, 1, 18, 0, 0, 0, time.UTC)
	match, err := entities.NewMatch(
		"match-1", "test-group", "Friday football", "Stadium A",
		scheduledAt, entities.MatchStatusDraft, time.Now(),
	)
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}

	if saveErr := repo.Save(ctx, match); saveErr != nil {
		t.Fatalf("expected Save to succeed, got: %v", saveErr)
	}

	found, findErr := repo.FindByID(ctx, "match-1")
	if findErr != nil {
		t.Fatalf("expected to find saved match, got: %v", findErr)
	}
	if found.Title() != "Friday football" {
		t.Errorf("expected title 'Friday football', got %q", found.Title())
	}
	if found.Status() != entities.MatchStatusDraft {
		t.Errorf("expected status draft, got %q", found.Status())
	}
}

func TestPostgresMatchRepository_Save_ReturnsErrReferencedEntityNotFound_WhenGroupMissing(t *testing.T) {
	repo := newTestMatchRepository(t)
	ctx := context.Background()

	match, _ := entities.NewMatch(
		"match-1", "nonexistent-group", "Title", "Venue",
		time.Now(), entities.MatchStatusDraft, time.Now(),
	)

	err := repo.Save(ctx, match)

	if !errors.Is(err, domainerrors.ErrReferencedEntityNotFound) {
		t.Errorf("expected ErrReferencedEntityNotFound, got %v", err)
	}
}

func TestPostgresMatchRepository_FindByID_ReturnsErrMatchNotFound_WhenMissing(t *testing.T) {
	repo := newTestMatchRepository(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "does-not-exist")

	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}

func TestPostgresMatchRepository_ListByGroup_ReturnsAllMatches(t *testing.T) {
	repo := newTestMatchRepository(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		match, _ := entities.NewMatch(
			entities.MatchID([]string{"m-1", "m-2", "m-3"}[i]),
			"test-group",
			"Match",
			"Venue",
			time.Now().Add(time.Duration(i)*time.Hour),
			entities.MatchStatusDraft,
			time.Now(),
		)
		if err := repo.Save(ctx, match); err != nil {
			t.Fatalf("setup: save %d failed: %v", i, err)
		}
	}

	matches, err := repo.ListByGroup(ctx, "test-group")
	if err != nil {
		t.Fatalf("expected ListByGroup to succeed, got: %v", err)
	}
	if len(matches) != 3 {
		t.Errorf("expected 3 matches, got %d", len(matches))
	}
}

func TestPostgresMatchRepository_UpdateStatus_PersistsTransition(t *testing.T) {
	repo := newTestMatchRepository(t)
	ctx := context.Background()

	match, _ := entities.NewMatch(
		"match-1", "test-group", "Title", "Venue",
		time.Now().Add(time.Hour), entities.MatchStatusDraft, time.Now(),
	)
	if err := repo.Save(ctx, match); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Use the domain's state machine to transition the status.
	if err := match.Open(); err != nil {
		t.Fatalf("setup: Open failed: %v", err)
	}

	if err := repo.UpdateStatus(ctx, match); err != nil {
		t.Fatalf("expected UpdateStatus to succeed, got: %v", err)
	}

	found, _ := repo.FindByID(ctx, "match-1")
	if found.Status() != entities.MatchStatusOpen {
		t.Errorf("expected status open after update, got %q", found.Status())
	}
}

func TestPostgresMatchRepository_Delete_RemovesMatch(t *testing.T) {
	repo := newTestMatchRepository(t)
	ctx := context.Background()

	match, _ := entities.NewMatch(
		"match-1", "test-group", "Title", "Venue",
		time.Now(), entities.MatchStatusDraft, time.Now(),
	)
	if err := repo.Save(ctx, match); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := repo.Delete(ctx, "match-1"); err != nil {
		t.Fatalf("expected Delete to succeed, got: %v", err)
	}

	_, findErr := repo.FindByID(ctx, "match-1")
	if !errors.Is(findErr, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound after delete, got %v", findErr)
	}
}
