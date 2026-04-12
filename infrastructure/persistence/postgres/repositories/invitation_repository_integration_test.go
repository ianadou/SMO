//go:build integration

package repositories_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/repositories"
)

// newTestInvitationRepository cleans the invitations table and re-seeds a
// match so invitation rows can reference it via FK.
func newTestInvitationRepository(t *testing.T) *repositories.PostgresInvitationRepository {
	t.Helper()
	ctx := context.Background()

	if _, err := sharedPool.Exec(ctx, "DELETE FROM invitations"); err != nil {
		t.Fatalf("clean invitations: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, "DELETE FROM matches"); err != nil {
		t.Fatalf("clean matches: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO groups (id, organizer_id, name)
		VALUES ('test-group', 'test-org', 'Test Group')
		ON CONFLICT (id) DO NOTHING
	`); err != nil {
		t.Fatalf("re-seed group: %v", err)
	}
	if _, err := sharedPool.Exec(ctx, `
		INSERT INTO matches (id, group_id, title, venue, scheduled_at, status)
		VALUES ('test-match', 'test-group', 'Test Match', 'Venue', NOW() + INTERVAL '1 day', 'draft')
	`); err != nil {
		t.Fatalf("seed match: %v", err)
	}
	return repositories.NewPostgresInvitationRepository(sharedPool)
}

func buildTestInvitation(t *testing.T, id, matchID, hash string) *entities.Invitation {
	t.Helper()
	createdAt := time.Now()
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	inv, err := entities.NewInvitation(
		entities.InvitationID(id), entities.MatchID(matchID), hash,
		expiresAt, nil, createdAt,
	)
	if err != nil {
		t.Fatalf("NewInvitation: %v", err)
	}
	return inv
}

func TestPostgresInvitationRepository_Save_PersistsInvitation(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "test-match", "hash-1")
	if err := repo.Save(ctx, inv); err != nil {
		t.Fatalf("expected Save to succeed: %v", err)
	}

	found, findErr := repo.FindByID(ctx, "inv-1")
	if findErr != nil {
		t.Fatalf("expected to find saved invitation: %v", findErr)
	}
	if found.TokenHash() != "hash-1" {
		t.Errorf("expected hash 'hash-1', got %q", found.TokenHash())
	}
}

func TestPostgresInvitationRepository_Save_ReturnsErrReferencedEntityNotFound(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "nonexistent-match", "hash")
	err := repo.Save(ctx, inv)
	if !errors.Is(err, domainerrors.ErrReferencedEntityNotFound) {
		t.Errorf("expected ErrReferencedEntityNotFound, got %v", err)
	}
}

func TestPostgresInvitationRepository_FindByID_ReturnsErrInvitationNotFound(t *testing.T) {
	repo := newTestInvitationRepository(t)
	_, err := repo.FindByID(context.Background(), "does-not-exist")
	if !errors.Is(err, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestPostgresInvitationRepository_FindByTokenHash_ReturnsInvitation(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "test-match", "unique-hash-xyz")
	_ = repo.Save(ctx, inv)

	found, err := repo.FindByTokenHash(ctx, "unique-hash-xyz")
	if err != nil {
		t.Fatalf("expected to find by hash, got: %v", err)
	}
	if found.ID() != "inv-1" {
		t.Errorf("expected inv-1, got %q", found.ID())
	}
}

func TestPostgresInvitationRepository_FindByTokenHash_ReturnsErrInvitationNotFound(t *testing.T) {
	repo := newTestInvitationRepository(t)
	_, err := repo.FindByTokenHash(context.Background(), "nonexistent-hash")
	if !errors.Is(err, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestPostgresInvitationRepository_ListByMatch_ReturnsAll(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	for i, hash := range []string{"h1", "h2", "h3"} {
		inv := buildTestInvitation(t, "inv-"+string(rune('1'+i)), "test-match", hash)
		if err := repo.Save(ctx, inv); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	invitations, err := repo.ListByMatch(ctx, "test-match")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(invitations) != 3 {
		t.Errorf("expected 3, got %d", len(invitations))
	}
}

func TestPostgresInvitationRepository_MarkAsUsed_PersistsUsedAt(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "test-match", "hash")
	_ = repo.Save(ctx, inv)

	if err := inv.MarkAsUsed(time.Now()); err != nil {
		t.Fatalf("MarkAsUsed: %v", err)
	}
	if err := repo.MarkAsUsed(ctx, inv); err != nil {
		t.Fatalf("repo.MarkAsUsed: %v", err)
	}

	found, _ := repo.FindByID(ctx, "inv-1")
	if found.UsedAt() == nil {
		t.Error("expected UsedAt to be set after MarkAsUsed")
	}
}

func TestPostgresInvitationRepository_Delete_RemovesInvitation(t *testing.T) {
	repo := newTestInvitationRepository(t)
	ctx := context.Background()

	inv := buildTestInvitation(t, "inv-1", "test-match", "hash")
	_ = repo.Save(ctx, inv)
	if err := repo.Delete(ctx, "inv-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, findErr := repo.FindByID(ctx, "inv-1")
	if !errors.Is(findErr, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound after delete, got %v", findErr)
	}
}
