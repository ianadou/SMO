package invitation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

type contextFixture struct {
	uc   *GetInvitationContextUseCase
	repo *fakeInvitationRepository
}

// contextFixtureToken is the plain token seeded by newContextFixture;
// tests resolve it through Execute and use any other value to exercise
// the not-found branch.
const contextFixtureToken = "tok"

func newContextFixture(
	t *testing.T,
	expiresAt time.Time,
	response entities.InvitationResponse,
	respondedAt *time.Time,
	matchStatus entities.MatchStatus,
	clockNow time.Time,
) contextFixture {
	t.Helper()

	repo := newFakeInvitationRepository()
	matchRepo := newFakeMatchRepo()
	matchRepo.seedMatchWithStatus(t, matchStatus)
	groupRepo := newFakeGroupRepo()
	groupRepo.seedGroup(t, "group-1", "org-1", "Les Bras Cassés")
	organizerRepo := newFakeOrganizerRepo()
	organizerRepo.seedOrganizer(t, "org-1", "Eddin")
	tokens := newFakeTokenService()
	hash := tokens.HashToken(contextFixtureToken)

	createdAt := expiresAt.Add(-7 * 24 * time.Hour)
	inv, err := entities.NewInvitation("inv-1", "match-1", "p-1", hash, expiresAt, response, respondedAt, nil, createdAt)
	if err != nil {
		t.Fatalf("setup: NewInvitation: %v", err)
	}
	_ = repo.Save(context.Background(), inv)

	uc := NewGetInvitationContextUseCase(repo, matchRepo, groupRepo, organizerRepo, tokens, newFakeClock(clockNow))
	return contextFixture{uc: uc, repo: repo}
}

func TestGetInvitationContextUseCase_Execute_AssemblesContext_WhenTokenResolves(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(48 * time.Hour)
	f := newContextFixture(t, expires, entities.InvitationResponsePending, nil, entities.MatchStatusOpen, now)

	got, err := f.uc.Execute(context.Background(), contextFixtureToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.OrganizerName != "Eddin" {
		t.Errorf("OrganizerName = %q, want %q", got.OrganizerName, "Eddin")
	}
	if got.GroupName != "Les Bras Cassés" {
		t.Errorf("GroupName = %q, want %q", got.GroupName, "Les Bras Cassés")
	}
	if got.MatchTitle != "Match" {
		t.Errorf("MatchTitle = %q, want %q", got.MatchTitle, "Match")
	}
	if got.Venue != "Venue" {
		t.Errorf("Venue = %q, want %q", got.Venue, "Venue")
	}
	if got.MaxParticipants != entities.MaxParticipantsPerMatch {
		t.Errorf("MaxParticipants = %d, want %d", got.MaxParticipants, entities.MaxParticipantsPerMatch)
	}
	if got.Response != entities.InvitationResponsePending {
		t.Errorf("Response = %q, want pending", got.Response)
	}
	if !got.ExpiresAt.Equal(expires) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, expires)
	}
	if got.Locked {
		t.Error("Locked = true, want false for an open match")
	}
	if got.Expired {
		t.Error("Expired = true, want false before expiry")
	}
	if len(got.ConfirmedNames) != 0 {
		t.Errorf("ConfirmedNames len = %d, want 0", len(got.ConfirmedNames))
	}
}

func TestGetInvitationContextUseCase_Execute_CountsConfirmedParticipants(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(48 * time.Hour)
	f := newContextFixture(t, expires, entities.InvitationResponsePending, nil, entities.MatchStatusOpen, now)

	respondedAt := now
	confirmed, err := entities.NewInvitation(
		"inv-2", "match-1", "p-2", "other-hash", expires,
		entities.InvitationResponseYes, &respondedAt, nil, now.Add(-time.Hour),
	)
	if err != nil {
		t.Fatalf("setup: confirmed invitation: %v", err)
	}
	_ = f.repo.Save(context.Background(), confirmed)

	got, err := f.uc.Execute(context.Background(), contextFixtureToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got.ConfirmedNames) != 1 {
		t.Fatalf("ConfirmedNames len = %d, want 1", len(got.ConfirmedNames))
	}
	if got.ConfirmedNames[0] != "Fake p-2" {
		t.Errorf("ConfirmedNames[0] = %q, want %q", got.ConfirmedNames[0], "Fake p-2")
	}
}

func TestGetInvitationContextUseCase_Execute_ReportsExpired_WhenPastExpiry(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(-time.Hour)
	f := newContextFixture(t, expires, entities.InvitationResponsePending, nil, entities.MatchStatusOpen, now)

	got, err := f.uc.Execute(context.Background(), contextFixtureToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !got.Expired {
		t.Error("Expired = false, want true for an invitation past its expiry")
	}
}

func TestGetInvitationContextUseCase_Execute_ReportsLocked_WhenMatchPastTeamsReady(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(48 * time.Hour)
	f := newContextFixture(t, expires, entities.InvitationResponseYes, &now, entities.MatchStatusTeamsReady, now)

	got, err := f.uc.Execute(context.Background(), contextFixtureToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !got.Locked {
		t.Error("Locked = false, want true for a teams_ready match")
	}
}

func TestGetInvitationContextUseCase_Execute_ReturnsNotFound_WhenTokenDoesNotMatch(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(48 * time.Hour)
	f := newContextFixture(t, expires, entities.InvitationResponsePending, nil, entities.MatchStatusOpen, now)

	_, err := f.uc.Execute(context.Background(), "wrong-token")
	if !errors.Is(err, domainerrors.ErrInvitationNotFound) {
		t.Fatalf("error = %v, want ErrInvitationNotFound", err)
	}
}
