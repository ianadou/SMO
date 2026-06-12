package sharelink

import (
	"context"
	"errors"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func newRevokeFixture(t *testing.T, now time.Time) (
	*RevokeMatchShareLinkUseCase, *fakeShareLinkRepository,
) {
	t.Helper()
	links := newFakeShareLinkRepository(now)
	matchRepo := newFakeMatchRepository()
	groupRepo := newFakeGroupRepository()
	matchRepo.seedMatch(t, "match-1", "group-1")
	groupRepo.seedGroup(t)
	uc := NewRevokeMatchShareLinkUseCase(links, matchRepo, groupRepo, newFakeClock(now))
	return uc, links
}

func TestRevokeMatchShareLinkUseCase_RevokesActiveLink(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	uc, links := newRevokeFixture(t, now)
	links.seedLink(t, "link-1", "hash", now.Add(48*time.Hour), nil, now.Add(-time.Hour))

	err := uc.Execute(context.Background(), "match-1", "org-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	revokedAt := links.links["link-1"].RevokedAt()
	if revokedAt == nil {
		t.Fatal("expected link to be revoked")
	}
	if !revokedAt.Equal(now) {
		t.Errorf("expected revokedAt %v, got %v", now, *revokedAt)
	}
}

func TestRevokeMatchShareLinkUseCase_ReturnsErrShareLinkNotFound_WhenNoActiveLinkExists(t *testing.T) {
	t.Parallel()
	uc, _ := newRevokeFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))

	err := uc.Execute(context.Background(), "match-1", "org-1")

	if !errors.Is(err, domainerrors.ErrShareLinkNotFound) {
		t.Errorf("expected ErrShareLinkNotFound, got %v", err)
	}
}

func TestRevokeMatchShareLinkUseCase_ReturnsErrMatchNotFound_WhenMatchDoesNotExist(t *testing.T) {
	t.Parallel()
	uc, _ := newRevokeFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))

	err := uc.Execute(context.Background(), "ghost-match", "org-1")

	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}

func TestRevokeMatchShareLinkUseCase_ReturnsErrMatchNotFound_WhenRequesterIsNotTheOwner(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	uc, links := newRevokeFixture(t, now)
	links.seedLink(t, "link-1", "hash", now.Add(48*time.Hour), nil, now.Add(-time.Hour))

	err := uc.Execute(context.Background(), "match-1", "org-intruder")

	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound (anti-oracle), got %v", err)
	}
	if links.links["link-1"].RevokedAt() != nil {
		t.Error("link must not be revoked on ownership rejection")
	}
}

func TestRevokeMatchShareLinkUseCase_PropagatesUpdateError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("disk full")
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	uc, links := newRevokeFixture(t, now)
	links.seedLink(t, "link-1", "hash", now.Add(48*time.Hour), nil, now.Add(-time.Hour))
	links.updateErr = repoErr

	err := uc.Execute(context.Background(), "match-1", "org-1")

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}

func TestRevokeMatchShareLinkUseCase_PropagatesError_WhenAdapterHandsBackDeadActiveLink(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	uc, links := newRevokeFixture(t, now)
	revokedAt := now.Add(-time.Hour)
	dead := links.seedLink(t, "link-dead", "dead-hash",
		now.Add(48*time.Hour), &revokedAt, now.Add(-2*time.Hour))
	links.findActiveOverride = dead

	err := uc.Execute(context.Background(), "match-1", "org-1")

	if !errors.Is(err, domainerrors.ErrShareLinkInactive) {
		t.Errorf("expected ErrShareLinkInactive when revoking a dead link, got %v", err)
	}
}
