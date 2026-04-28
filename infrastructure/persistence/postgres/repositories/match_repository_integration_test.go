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
		scheduledAt, time.Now(),
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
		time.Now(), time.Now(),
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
		time.Now().Add(time.Hour), time.Now(),
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

func TestPostgresMatchRepository_Finalize_PersistsMVPAndStatus(t *testing.T) {
	repo := newTestMatchRepository(t)
	ctx := context.Background()

	match, _ := entities.NewMatch(
		"match-1", "test-group", "Title", "Venue",
		time.Now().Add(time.Hour), time.Now(),
	)
	if err := repo.Save(ctx, match); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Walk the state machine to Completed via the entity's transitions,
	// then finalize with a non-nil MVP. The MVP value is just a string
	// here — the FK is ON DELETE SET NULL, so this test does not require
	// a real player row to exist (we are testing the repo, not the FK).
	if err := match.Open(); err != nil {
		t.Fatalf("setup: Open: %v", err)
	}
	if err := match.MarkTeamsReady(); err != nil {
		t.Fatalf("setup: MarkTeamsReady: %v", err)
	}
	if err := match.Start(); err != nil {
		t.Fatalf("setup: Start: %v", err)
	}
	if err := match.Complete(); err != nil {
		t.Fatalf("setup: Complete: %v", err)
	}
	if err := repo.UpdateStatus(ctx, match); err != nil {
		t.Fatalf("setup: persist completed: %v", err)
	}

	// Finalize with nil MVP: exercises the nullable column path. A non-nil
	// MVP would require a real player row to satisfy the FK; that case is
	// covered end-to-end by the FinalizeMatchUseCase tests.
	if err := match.Finalize(nil); err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if err := repo.Finalize(ctx, match); err != nil {
		t.Fatalf("repo.Finalize: %v", err)
	}

	found, _ := repo.FindByID(ctx, "match-1")
	if found.Status() != entities.MatchStatusClosed {
		t.Errorf("expected status closed, got %q", found.Status())
	}
	if found.MVP() != nil {
		t.Errorf("expected nil MVP, got %v", found.MVP())
	}
}

// TestPostgresMatchRepository_Finalize_PersistsNonNilMVP covers the
// other half of the Finalize path: the FK matches.mvp_player_id →
// players.id is exercised with a real player row, so a non-nil MVP
// round-trips through the database.
//
// Tracked in issue #34: previously the only test for Finalize used
// MVP=nil because seeding a player required FK plumbing.
func TestPostgresMatchRepository_Finalize_PersistsNonNilMVP(t *testing.T) {
	repo := newTestMatchRepository(t)
	ctx := context.Background()

	// Seed a player whose ID we can hand to Finalize. test-group is
	// re-seeded by newTestMatchRepository, so the FK
	// players.group_id → groups.id is also satisfied.
	const mvpID = "player-mvp"
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO players (id, group_id, name, ranking)
		VALUES ($1, 'test-group', 'MVP Player', 1500)
		ON CONFLICT (id) DO NOTHING
	`, mvpID); err != nil {
		t.Fatalf("setup: insert player: %v", err)
	}
	t.Cleanup(func() {
		_, _ = sharedPool.Exec(context.Background(), `DELETE FROM players WHERE id = $1`, mvpID)
	})

	match, _ := entities.NewMatch(
		"match-mvp", "test-group", "Title", "Venue",
		time.Now().Add(time.Hour), time.Now(),
	)
	if err := repo.Save(ctx, match); err != nil {
		t.Fatalf("setup: save match: %v", err)
	}

	// Walk the state machine to Completed, then Finalize with the
	// real player as MVP.
	for i, step := range []func() error{
		match.Open, match.MarkTeamsReady, match.Start, match.Complete,
	} {
		if err := step(); err != nil {
			t.Fatalf("setup: transition %d: %v", i, err)
		}
	}
	if err := repo.UpdateStatus(ctx, match); err != nil {
		t.Fatalf("setup: persist completed: %v", err)
	}

	mvpPlayerID := entities.PlayerID(mvpID)
	if err := match.Finalize(&mvpPlayerID); err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if err := repo.Finalize(ctx, match); err != nil {
		t.Fatalf("repo.Finalize: %v", err)
	}

	found, err := repo.FindByID(ctx, "match-mvp")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Status() != entities.MatchStatusClosed {
		t.Errorf("expected status closed, got %q", found.Status())
	}
	if found.MVP() == nil {
		t.Fatalf("expected non-nil MVP, got nil")
	}
	if *found.MVP() != mvpPlayerID {
		t.Errorf("expected MVP %q, got %q", mvpPlayerID, *found.MVP())
	}
}

func TestPostgresMatchRepository_Delete_RemovesMatch(t *testing.T) {
	repo := newTestMatchRepository(t)
	ctx := context.Background()

	match, _ := entities.NewMatch(
		"match-1", "test-group", "Title", "Venue",
		time.Now(), time.Now(),
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
