package entities

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestNewInvitation_ReturnsPendingInvitation_WhenInputsAreValid(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)

	inv, err := NewInvitation("inv-1", "match-1", "p-1", "hash-abc-123", expiresAt, InvitationResponsePending, nil, nil, createdAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if inv.ID() != "inv-1" {
		t.Errorf("expected ID 'inv-1', got %q", inv.ID())
	}
	if inv.MatchID() != "match-1" {
		t.Errorf("expected MatchID 'match-1', got %q", inv.MatchID())
	}
	if inv.PlayerID() != "p-1" {
		t.Errorf("expected PlayerID 'p-1', got %q", inv.PlayerID())
	}
	if inv.TokenHash() != "hash-abc-123" {
		t.Errorf("expected token hash, got %q", inv.TokenHash())
	}
	if !inv.ExpiresAt().Equal(expiresAt) {
		t.Errorf("expected expiresAt %v, got %v", expiresAt, inv.ExpiresAt())
	}
	if inv.Response() != InvitationResponsePending {
		t.Errorf("expected pending response, got %q", inv.Response())
	}
	if inv.RespondedAt() != nil {
		t.Errorf("expected nil RespondedAt for pending invitation, got %v", inv.RespondedAt())
	}
	if inv.IsConfirmed() {
		t.Errorf("expected new invitation to not be confirmed")
	}
}

func TestNewInvitation_AcceptsSettledResponse_WhenRespondedAtProvided(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	respondedAt := createdAt.Add(2 * time.Hour)

	inv, err := NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, InvitationResponseYes, &respondedAt, nil, createdAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !inv.IsConfirmed() {
		t.Errorf("expected invitation to be confirmed when response is yes")
	}
	if inv.RespondedAt() == nil || !inv.RespondedAt().Equal(respondedAt) {
		t.Errorf("expected RespondedAt %v, got %v", respondedAt, inv.RespondedAt())
	}
}

func TestNewInvitation_AllowsDeclinedResponse_WithRespondedAt(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	respondedAt := createdAt.Add(time.Hour)

	inv, err := NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, InvitationResponseNo, &respondedAt, nil, createdAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if inv.IsConfirmed() {
		t.Errorf("a declined invitation must not be confirmed")
	}
	if inv.Response() != InvitationResponseNo {
		t.Errorf("expected response 'no', got %q", inv.Response())
	}
}

func TestInvitation_IsExpired(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	inv, _ := NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, InvitationResponsePending, nil, nil, createdAt)

	cases := []struct {
		name string
		now  time.Time
		want bool
	}{
		{name: "before expiration", now: time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC), want: false},
		{name: "exactly at expiration", now: expiresAt, want: true},
		{name: "after expiration", now: time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC), want: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := inv.IsExpired(testCase.now)

			if got != testCase.want {
				t.Errorf("expected IsExpired=%v, got %v", testCase.want, got)
			}
		})
	}
}

func TestNewInvitation_ReturnsError_WhenInputsAreInvalid(t *testing.T) {
	t.Parallel()

	validCreatedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	validExpiresAt := validCreatedAt.Add(24 * time.Hour)
	settledAt := validCreatedAt.Add(time.Hour)

	cases := []struct {
		name        string
		id          InvitationID
		matchID     MatchID
		playerID    PlayerID
		tokenHash   string
		expiresAt   time.Time
		response    InvitationResponse
		respondedAt *time.Time
		createdAt   time.Time
		wantErr     error
	}{
		{name: "empty id", id: "", matchID: "m-1", playerID: "p-1", tokenHash: "h", expiresAt: validExpiresAt, response: InvitationResponsePending, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidID},
		{name: "empty match id", id: "i-1", matchID: "", playerID: "p-1", tokenHash: "h", expiresAt: validExpiresAt, response: InvitationResponsePending, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidID},
		{name: "empty player id", id: "i-1", matchID: "m-1", playerID: "", tokenHash: "h", expiresAt: validExpiresAt, response: InvitationResponsePending, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidID},
		{name: "empty token hash", id: "i-1", matchID: "m-1", playerID: "p-1", tokenHash: "", expiresAt: validExpiresAt, response: InvitationResponsePending, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidID},
		{name: "zero createdAt", id: "i-1", matchID: "m-1", playerID: "p-1", tokenHash: "h", expiresAt: validExpiresAt, response: InvitationResponsePending, createdAt: time.Time{}, wantErr: domainerrors.ErrInvalidDate},
		{name: "zero expiresAt", id: "i-1", matchID: "m-1", playerID: "p-1", tokenHash: "h", expiresAt: time.Time{}, response: InvitationResponsePending, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidDate},
		{name: "expiresAt before createdAt", id: "i-1", matchID: "m-1", playerID: "p-1", tokenHash: "h", expiresAt: validCreatedAt.Add(-time.Hour), response: InvitationResponsePending, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidDate},
		{name: "expiresAt equals createdAt", id: "i-1", matchID: "m-1", playerID: "p-1", tokenHash: "h", expiresAt: validCreatedAt, response: InvitationResponsePending, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidDate},
		{name: "unknown response value", id: "i-1", matchID: "m-1", playerID: "p-1", tokenHash: "h", expiresAt: validExpiresAt, response: InvitationResponse("maybe"), createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidInvitationResponse},
		{name: "pending with respondedAt set", id: "i-1", matchID: "m-1", playerID: "p-1", tokenHash: "h", expiresAt: validExpiresAt, response: InvitationResponsePending, respondedAt: &settledAt, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidInvitationResponse},
		{name: "yes without respondedAt", id: "i-1", matchID: "m-1", playerID: "p-1", tokenHash: "h", expiresAt: validExpiresAt, response: InvitationResponseYes, respondedAt: nil, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidInvitationResponse},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			inv, err := NewInvitation(testCase.id, testCase.matchID, testCase.playerID, testCase.tokenHash, testCase.expiresAt, testCase.response, testCase.respondedAt, nil, testCase.createdAt)

			if inv != nil {
				t.Errorf("expected nil invitation, got %+v", inv)
			}
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("expected error %v, got %v", testCase.wantErr, err)
			}
		})
	}
}

func TestInvitation_Respond_TransitionsResponse(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	respondAt := createdAt.Add(2 * time.Hour)

	cases := []struct {
		name          string
		initialAnswer InvitationResponse
		initialAt     *time.Time
		answer        InvitationResponse
		wantConfirmed bool
		wantResponse  InvitationResponse
	}{
		{name: "pending to yes", initialAnswer: InvitationResponsePending, initialAt: nil, answer: InvitationResponseYes, wantConfirmed: true, wantResponse: InvitationResponseYes},
		{name: "pending to no", initialAnswer: InvitationResponsePending, initialAt: nil, answer: InvitationResponseNo, wantConfirmed: false, wantResponse: InvitationResponseNo},
		{name: "yes to no", initialAnswer: InvitationResponseYes, initialAt: &createdAt, answer: InvitationResponseNo, wantConfirmed: false, wantResponse: InvitationResponseNo},
		{name: "no to yes", initialAnswer: InvitationResponseNo, initialAt: &createdAt, answer: InvitationResponseYes, wantConfirmed: true, wantResponse: InvitationResponseYes},
		{name: "idempotent yes", initialAnswer: InvitationResponseYes, initialAt: &createdAt, answer: InvitationResponseYes, wantConfirmed: true, wantResponse: InvitationResponseYes},
		{name: "idempotent no", initialAnswer: InvitationResponseNo, initialAt: &createdAt, answer: InvitationResponseNo, wantConfirmed: false, wantResponse: InvitationResponseNo},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			inv, _ := NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, testCase.initialAnswer, testCase.initialAt, nil, createdAt)

			err := inv.Respond(testCase.answer, respondAt, false)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if inv.Response() != testCase.wantResponse {
				t.Errorf("expected response %q, got %q", testCase.wantResponse, inv.Response())
			}
			if inv.IsConfirmed() != testCase.wantConfirmed {
				t.Errorf("expected IsConfirmed=%v, got %v", testCase.wantConfirmed, inv.IsConfirmed())
			}
			if inv.RespondedAt() == nil || !inv.RespondedAt().Equal(respondAt) {
				t.Errorf("expected RespondedAt %v, got %v", respondAt, inv.RespondedAt())
			}
		})
	}
}

func TestInvitation_Respond_RejectsPendingAsAnswer(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	inv, _ := NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, InvitationResponsePending, nil, nil, createdAt)

	err := inv.Respond(InvitationResponsePending, createdAt.Add(time.Hour), false)

	if !errors.Is(err, domainerrors.ErrInvalidInvitationResponse) {
		t.Errorf("expected ErrInvalidInvitationResponse, got %v", err)
	}
}

func TestInvitation_Respond_RejectsUnknownAnswer(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	inv, _ := NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, InvitationResponsePending, nil, nil, createdAt)

	err := inv.Respond(InvitationResponse("perhaps"), createdAt.Add(time.Hour), false)

	if !errors.Is(err, domainerrors.ErrInvalidInvitationResponse) {
		t.Errorf("expected ErrInvalidInvitationResponse, got %v", err)
	}
}

func TestInvitation_Respond_ReturnsErrInvitationExpired_WhenExpired(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)
	inv, _ := NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, InvitationResponsePending, nil, nil, createdAt)

	err := inv.Respond(InvitationResponseYes, createdAt.Add(2*time.Hour), false)

	if !errors.Is(err, domainerrors.ErrInvitationExpired) {
		t.Errorf("expected ErrInvitationExpired, got %v", err)
	}
}

func TestInvitation_Respond_ReturnsErrInvitationLocked_WhenLocked(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	inv, _ := NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, InvitationResponsePending, nil, nil, createdAt)

	err := inv.Respond(InvitationResponseYes, createdAt.Add(time.Hour), true)

	if !errors.Is(err, domainerrors.ErrInvitationLocked) {
		t.Errorf("expected ErrInvitationLocked, got %v", err)
	}
}

func TestInvitation_Respond_PrefersExpiredOverLocked(t *testing.T) {
	t.Parallel()
	// Expired is the more actionable signal for the player: even if the
	// match also locked attendance, the invitation itself is dead.
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)
	inv, _ := NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, InvitationResponsePending, nil, nil, createdAt)

	err := inv.Respond(InvitationResponseYes, createdAt.Add(2*time.Hour), true)

	if !errors.Is(err, domainerrors.ErrInvitationExpired) {
		t.Errorf("expected ErrInvitationExpired (priority), got %v", err)
	}
}

func TestInvitation_Claim_RotatesTokenHashAndSetsClaimedAt_WhenPendingAndUnclaimed(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	claimAt := createdAt.Add(2 * time.Hour)
	inv, _ := NewInvitation("inv-1", "match-1", "p-1", "old-hash", expiresAt, InvitationResponsePending, nil, nil, createdAt)

	err := inv.Claim("new-hash", claimAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if inv.TokenHash() != "new-hash" {
		t.Errorf("expected token hash rotated to 'new-hash', got %q", inv.TokenHash())
	}
	if inv.ClaimedAt() == nil || !inv.ClaimedAt().Equal(claimAt) {
		t.Errorf("expected ClaimedAt %v, got %v", claimAt, inv.ClaimedAt())
	}
	if inv.Response() != InvitationResponsePending {
		t.Errorf("expected response to stay pending after claim, got %q", inv.Response())
	}
}

func TestInvitation_Claim_ReturnsErrInvitationAlreadyClaimed_WhenAlreadyClaimed(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	inv, _ := NewInvitation("inv-1", "match-1", "p-1", "old-hash", expiresAt, InvitationResponsePending, nil, nil, createdAt)
	firstClaimAt := createdAt.Add(time.Hour)
	if err := inv.Claim("first-hash", firstClaimAt); err != nil {
		t.Fatalf("first claim must succeed, got: %v", err)
	}

	err := inv.Claim("second-hash", createdAt.Add(2*time.Hour))

	if !errors.Is(err, domainerrors.ErrInvitationAlreadyClaimed) {
		t.Errorf("expected ErrInvitationAlreadyClaimed, got %v", err)
	}
	if inv.TokenHash() != "first-hash" {
		t.Errorf("expected first claimer's hash to be preserved, got %q", inv.TokenHash())
	}
	if !inv.ClaimedAt().Equal(firstClaimAt) {
		t.Errorf("expected first ClaimedAt %v to be preserved, got %v", firstClaimAt, inv.ClaimedAt())
	}
}

func TestInvitation_Claim_ReturnsErrInvitationAlreadyClaimed_WhenAlreadyRespondedYes(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	respondedAt := createdAt.Add(time.Hour)
	inv, _ := NewInvitation("inv-1", "match-1", "p-1", "old-hash", expiresAt, InvitationResponseYes, &respondedAt, nil, createdAt)

	err := inv.Claim("new-hash", createdAt.Add(2*time.Hour))

	if !errors.Is(err, domainerrors.ErrInvitationAlreadyClaimed) {
		t.Errorf("expected ErrInvitationAlreadyClaimed, got %v", err)
	}
	if inv.TokenHash() != "old-hash" {
		t.Errorf("expected token hash to be preserved on refused claim, got %q", inv.TokenHash())
	}
	if inv.ClaimedAt() != nil {
		t.Errorf("expected ClaimedAt to stay nil on refused claim, got %v", inv.ClaimedAt())
	}
}

func TestInvitation_Claim_ReturnsErrInvitationExpired_WhenExpired(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)
	inv, _ := NewInvitation("inv-1", "match-1", "p-1", "old-hash", expiresAt, InvitationResponsePending, nil, nil, createdAt)

	err := inv.Claim("new-hash", createdAt.Add(2*time.Hour))

	if !errors.Is(err, domainerrors.ErrInvitationExpired) {
		t.Errorf("expected ErrInvitationExpired, got %v", err)
	}
	if inv.TokenHash() != "old-hash" {
		t.Errorf("expected token hash to be preserved on refused claim, got %q", inv.TokenHash())
	}
}

func TestInvitation_Claim_ReturnsErrInvalidID_WhenNewTokenHashIsEmpty(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	inv, _ := NewInvitation("inv-1", "match-1", "p-1", "old-hash", expiresAt, InvitationResponsePending, nil, nil, createdAt)

	err := inv.Claim("", createdAt.Add(time.Hour))

	if !errors.Is(err, domainerrors.ErrInvalidID) {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}
	if inv.ClaimedAt() != nil {
		t.Errorf("expected ClaimedAt to stay nil on refused claim, got %v", inv.ClaimedAt())
	}
}
