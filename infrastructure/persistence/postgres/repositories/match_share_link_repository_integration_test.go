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

// newTestMatchShareLinkRepository cleans the match_share_links table and
// re-seeds a match so link rows can reference it via FK.
func newTestMatchShareLinkRepository(t *testing.T) *repositories.PostgresMatchShareLinkRepository {
	t.Helper()
	ctx := context.Background()

	if _, err := sharedPool.Exec(ctx, "DELETE FROM match_share_links"); err != nil {
		t.Fatalf("clean match_share_links: %v", err)
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
		VALUES ('test-match', 'test-group', 'Test Match', 'Venue', NOW() + INTERVAL '1 day', 'open')
	`); err != nil {
		t.Fatalf("seed match: %v", err)
	}
	return repositories.NewPostgresMatchShareLinkRepository(sharedPool)
}

func buildTestShareLink(t *testing.T, id, matchID, hash string) *entities.MatchShareLink {
	t.Helper()
	createdAt := time.Now()
	link, err := entities.NewMatchShareLink(
		entities.MatchShareLinkID(id), entities.MatchID(matchID), hash,
		createdAt.Add(5*24*time.Hour), nil, createdAt,
	)
	if err != nil {
		t.Fatalf("NewMatchShareLink: %v", err)
	}
	return link
}

func TestPostgresMatchShareLinkRepository_Create_PersistsLink(t *testing.T) {
	repo := newTestMatchShareLinkRepository(t)
	ctx := context.Background()

	link := buildTestShareLink(t, "link-1", "test-match", "hash-1")
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("expected Create to succeed: %v", err)
	}

	found, findErr := repo.FindByTokenHash(ctx, "hash-1")
	if findErr != nil {
		t.Fatalf("expected to find created link: %v", findErr)
	}
	if found.ID() != "link-1" {
		t.Errorf("expected 'link-1', got %q", found.ID())
	}
	if found.MatchID() != "test-match" {
		t.Errorf("expected match 'test-match', got %q", found.MatchID())
	}
}

func TestPostgresMatchShareLinkRepository_Create_ReturnsErrReferencedEntityNotFound(t *testing.T) {
	repo := newTestMatchShareLinkRepository(t)
	ctx := context.Background()

	link := buildTestShareLink(t, "link-1", "nonexistent-match", "hash")
	err := repo.Create(ctx, link)
	if !errors.Is(err, domainerrors.ErrReferencedEntityNotFound) {
		t.Errorf("expected ErrReferencedEntityNotFound, got %v", err)
	}
}

func TestPostgresMatchShareLinkRepository_FindByTokenHash_ReturnsErrShareLinkNotFound(t *testing.T) {
	repo := newTestMatchShareLinkRepository(t)
	_, err := repo.FindByTokenHash(context.Background(), "nonexistent-hash")
	if !errors.Is(err, domainerrors.ErrShareLinkNotFound) {
		t.Errorf("expected ErrShareLinkNotFound, got %v", err)
	}
}

func TestPostgresMatchShareLinkRepository_FindActiveByMatchID_ReturnsActiveLink(t *testing.T) {
	repo := newTestMatchShareLinkRepository(t)
	ctx := context.Background()

	link := buildTestShareLink(t, "link-1", "test-match", "hash-1")
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.FindActiveByMatchID(ctx, "test-match")
	if err != nil {
		t.Fatalf("expected an active link: %v", err)
	}
	if found.ID() != "link-1" {
		t.Errorf("expected 'link-1', got %q", found.ID())
	}
}

func TestPostgresMatchShareLinkRepository_FindActiveByMatchID_ReturnsErrShareLinkNotFound_WhenLinkIsRevoked(t *testing.T) {
	repo := newTestMatchShareLinkRepository(t)
	ctx := context.Background()

	link := buildTestShareLink(t, "link-1", "test-match", "hash-1")
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := link.Revoke(time.Now()); err != nil {
		t.Fatalf("entity Revoke: %v", err)
	}
	if err := repo.Update(ctx, link); err != nil {
		t.Fatalf("Update: %v", err)
	}

	_, err := repo.FindActiveByMatchID(ctx, "test-match")
	if !errors.Is(err, domainerrors.ErrShareLinkNotFound) {
		t.Errorf("expected ErrShareLinkNotFound for a revoked link, got %v", err)
	}
}

func TestPostgresMatchShareLinkRepository_FindActiveByMatchID_ReturnsErrShareLinkNotFound_WhenLinkIsExpired(t *testing.T) {
	repo := newTestMatchShareLinkRepository(t)
	ctx := context.Background()

	createdAt := time.Now().Add(-2 * time.Hour)
	expired, buildErr := entities.NewMatchShareLink(
		"link-expired", "test-match", "hash-expired",
		createdAt.Add(time.Hour), nil, createdAt,
	)
	if buildErr != nil {
		t.Fatalf("NewMatchShareLink: %v", buildErr)
	}
	if err := repo.Create(ctx, expired); err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err := repo.FindActiveByMatchID(ctx, "test-match")
	if !errors.Is(err, domainerrors.ErrShareLinkNotFound) {
		t.Errorf("expected ErrShareLinkNotFound for an expired link, got %v", err)
	}
}

func TestPostgresMatchShareLinkRepository_Update_PersistsRevocation(t *testing.T) {
	repo := newTestMatchShareLinkRepository(t)
	ctx := context.Background()

	link := buildTestShareLink(t, "link-1", "test-match", "hash-1")
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := link.Revoke(time.Now()); err != nil {
		t.Fatalf("entity Revoke: %v", err)
	}
	if err := repo.Update(ctx, link); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// FindByTokenHash returns the link regardless of active state: the
	// use case decides what an inactive link means.
	found, findErr := repo.FindByTokenHash(ctx, "hash-1")
	if findErr != nil {
		t.Fatalf("expected revoked link to remain findable by hash: %v", findErr)
	}
	if found.RevokedAt() == nil {
		t.Error("expected revoked_at to be persisted")
	}
}

func TestPostgresMatchShareLinkRepository_Update_ReturnsErrShareLinkNotFound(t *testing.T) {
	repo := newTestMatchShareLinkRepository(t)
	ctx := context.Background()

	// Built but never created: the UPDATE matches zero rows.
	link := buildTestShareLink(t, "link-missing", "test-match", "hash")
	if err := link.Revoke(time.Now()); err != nil {
		t.Fatalf("entity Revoke: %v", err)
	}

	err := repo.Update(ctx, link)
	if !errors.Is(err, domainerrors.ErrShareLinkNotFound) {
		t.Errorf("expected ErrShareLinkNotFound, got %v", err)
	}
}
