package sharelink

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/application/usecases/invitation"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func newGenerateFixture(t *testing.T, now time.Time) (
	*GenerateMatchShareLinkUseCase, *fakeShareLinkRepository,
) {
	t.Helper()
	links := newFakeShareLinkRepository(now)
	matchRepo := newFakeMatchRepository()
	groupRepo := newFakeGroupRepository()
	matchRepo.seedMatch(t, "match-1", "group-1")
	groupRepo.seedGroup(t, "group-1", "org-1", "Sunday League")
	uc := NewGenerateMatchShareLinkUseCase(
		links, matchRepo, groupRepo,
		newFakeTokenService("plain-share-token"),
		newFakeIDGenerator("link-1"),
		newFakeClock(now),
	)
	return uc, links
}

func TestGenerateMatchShareLinkUseCase_ReturnsPlainTokenAndLink_WhenMatchHasNoActiveLink(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	uc, links := newGenerateFixture(t, now)

	result, err := uc.Execute(context.Background(), "match-1", "org-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PlainToken != "plain-share-token" {
		t.Errorf("expected plain token 'plain-share-token', got %q", result.PlainToken)
	}
	if result.ShareLink.MatchID() != "match-1" {
		t.Errorf("expected matchID 'match-1', got %q", result.ShareLink.MatchID())
	}
	expectedHash := newFakeTokenService().HashToken("plain-share-token")
	if result.ShareLink.TokenHash() != expectedHash {
		t.Errorf("expected stored hash of the plain token, got %q", result.ShareLink.TokenHash())
	}
	expectedExpiry := now.Add(invitation.DefaultInvitationValidityDuration)
	if !result.ShareLink.ExpiresAt().Equal(expectedExpiry) {
		t.Errorf("expected expiresAt %v, got %v", expectedExpiry, result.ShareLink.ExpiresAt())
	}
	if _, ok := links.links["link-1"]; !ok {
		t.Error("expected link to be persisted under id 'link-1'")
	}
}

func TestGenerateMatchShareLinkUseCase_RevokesPreviousActiveLink_WhenRegenerating(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	uc, links := newGenerateFixture(t, now)
	links.seedLink(t, "link-old", "old-hash", now.Add(48*time.Hour), nil, now.Add(-time.Hour))

	result, err := uc.Execute(context.Background(), "match-1", "org-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if links.links["link-old"].RevokedAt() == nil {
		t.Error("expected previous link to be revoked")
	}
	active, findErr := links.FindActiveByMatchID(context.Background(), "match-1")
	if findErr != nil {
		t.Fatalf("unexpected error finding active link: %v", findErr)
	}
	if active.ID() != result.ShareLink.ID() {
		t.Errorf("expected the new link %q to be the only active one, got %q",
			result.ShareLink.ID(), active.ID())
	}
}

func TestGenerateMatchShareLinkUseCase_ReturnsErrMatchNotFound_WhenMatchDoesNotExist(t *testing.T) {
	t.Parallel()
	uc, _ := newGenerateFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))

	_, err := uc.Execute(context.Background(), "ghost-match", "org-1")

	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}

func TestGenerateMatchShareLinkUseCase_ReturnsErrMatchNotFound_WhenRequesterIsNotTheOwner(t *testing.T) {
	t.Parallel()
	uc, links := newGenerateFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))

	_, err := uc.Execute(context.Background(), "match-1", "org-intruder")

	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound (anti-oracle), got %v", err)
	}
	if len(links.links) != 0 {
		t.Errorf("link must not be persisted on ownership rejection, found %d", len(links.links))
	}
}

func TestGenerateMatchShareLinkUseCase_ReturnsErrGroupNotFound_WhenMatchReferencesMissingGroup(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	links := newFakeShareLinkRepository(now)
	matchRepo := newFakeMatchRepository()
	matchRepo.seedMatch(t, "match-orphan", "ghost-group")
	uc := NewGenerateMatchShareLinkUseCase(
		links, matchRepo, newFakeGroupRepository(),
		newFakeTokenService("tok"), newFakeIDGenerator("link-1"), newFakeClock(now),
	)

	_, err := uc.Execute(context.Background(), "match-orphan", "org-1")

	if !errors.Is(err, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound on dangling reference, got %v", err)
	}
}

func TestGenerateMatchShareLinkUseCase_PropagatesFindActiveLinkError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("db down")
	uc, links := newGenerateFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	links.findActiveErr = repoErr

	_, err := uc.Execute(context.Background(), "match-1", "org-1")

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}

func TestGenerateMatchShareLinkUseCase_PropagatesUpdateError_WhenRevokingPreviousLink(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("db down")
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	uc, links := newGenerateFixture(t, now)
	links.seedLink(t, "link-old", "old-hash", now.Add(48*time.Hour), nil, now.Add(-time.Hour))
	links.updateErr = repoErr

	_, err := uc.Execute(context.Background(), "match-1", "org-1")

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}

func TestGenerateMatchShareLinkUseCase_PropagatesCreateError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("disk full")
	uc, links := newGenerateFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	links.createErr = repoErr

	_, err := uc.Execute(context.Background(), "match-1", "org-1")

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}
