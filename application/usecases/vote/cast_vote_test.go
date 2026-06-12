package vote

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

const votableMatchID = entities.MatchID("test-match")

var voteTestKickoff = time.Date(2026, 6, 4, 19, 0, 0, 0, time.UTC)

// seedMatchInStatus stores a match with rosters A=[p-1,p-2] B=[p-3,p-4]
// in the given status. Completed and closed snapshots carry a 3-2 score
// so WinningSide resolves.
func seedMatchInStatus(t *testing.T, repo *fakeMatchRepository, status entities.MatchStatus) *entities.Match {
	t.Helper()
	scoreA, scoreB := 3, 2
	snapshot := entities.MatchSnapshot{
		ID:          votableMatchID,
		GroupID:     "g-1",
		Title:       "Match du jeudi",
		Venue:       "Stade",
		ScheduledAt: voteTestKickoff,
		Status:      status,
		TeamA:       []entities.PlayerID{"p-1", "p-2"},
		TeamB:       []entities.PlayerID{"p-3", "p-4"},
		CreatedAt:   voteTestKickoff.Add(-48 * time.Hour),
	}
	if status == entities.MatchStatusCompleted || status == entities.MatchStatusClosed {
		snapshot.ScoreA = &scoreA
		snapshot.ScoreB = &scoreB
	}
	m, err := entities.RehydrateMatch(snapshot)
	if err != nil {
		t.Fatalf("seed match: %v", err)
	}
	repo.addMatch(m)
	return m
}

// seedBearerInvitation registers an invitation whose stored hash matches
// the plain token "tok-<player>" under the fake token service, and
// returns that plain token.
func seedBearerInvitation(
	t *testing.T,
	repo *fakeInvitationRepository,
	response entities.InvitationResponse,
) string {
	t.Helper()
	const playerID = entities.PlayerID("p-1")
	plain := "tok-" + string(playerID)
	respondedAt := voteTestKickoff.Add(-24 * time.Hour)
	inv, err := entities.NewInvitation(
		entities.InvitationID("inv-"+playerID), votableMatchID, playerID,
		fakeTokenService{}.HashToken(plain),
		voteTestKickoff.Add(24*time.Hour), response, &respondedAt, nil,
		voteTestKickoff.Add(-48*time.Hour),
	)
	if err != nil {
		t.Fatalf("seed invitation: %v", err)
	}
	repo.addInvitation(inv)
	return plain
}

func newCastVoteUseCase(t *testing.T) (*CastVoteUseCase, *fakeVoteRepository, *fakeMatchRepository, *fakeInvitationRepository) {
	t.Helper()
	voteRepo := newFakeVoteRepository()
	matchRepo := newFakeMatchRepository()
	invRepo := newFakeInvitationRepository()
	uc := NewCastVoteUseCase(voteRepo, matchRepo, invRepo, fakeTokenService{},
		newFakeIDGenerator("v-1", "v-2"), newFakeClock(voteTestKickoff.Add(2*time.Hour)))
	return uc, voteRepo, matchRepo, invRepo
}

func TestCastVoteUseCase_ReturnsVote_WhenBearerRatesTeammate(t *testing.T) {
	t.Parallel()
	uc, _, matchRepo, invRepo := newCastVoteUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	vote, err := uc.Execute(context.Background(), CastVoteInput{
		PlainToken: token, VotedID: "p-2", Score: 4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vote.VoterID() != "p-1" {
		t.Errorf("expected voter derived from token to be p-1, got %q", vote.VoterID())
	}
	if vote.Score() != 4 {
		t.Errorf("expected score 4, got %d", vote.Score())
	}
}

func TestCastVoteUseCase_AllowsVoting_WhenInvitationExpired(t *testing.T) {
	t.Parallel()
	uc, _, matchRepo, invRepo := newCastVoteUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	respondedAt := voteTestKickoff.Add(-24 * time.Hour)
	inv, err := entities.NewInvitation(
		"inv-p-1", votableMatchID, "p-1",
		fakeTokenService{}.HashToken("tok-p-1"),
		voteTestKickoff.Add(-time.Hour), entities.InvitationResponseYes, &respondedAt, nil,
		voteTestKickoff.Add(-48*time.Hour),
	)
	if err != nil {
		t.Fatalf("seed invitation: %v", err)
	}
	invRepo.addInvitation(inv)

	_, err = uc.Execute(context.Background(), CastVoteInput{
		PlainToken: "tok-p-1", VotedID: "p-2", Score: 5,
	})
	if err != nil {
		t.Fatalf("expiry must gate the RSVP only, never the vote: %v", err)
	}
}

func TestCastVoteUseCase_ReturnsErrInvitationNotFound_WhenTokenUnknown(t *testing.T) {
	t.Parallel()
	uc, _, matchRepo, _ := newCastVoteUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)

	_, err := uc.Execute(context.Background(), CastVoteInput{
		PlainToken: "tok-forged", VotedID: "p-2", Score: 4,
	})

	if !errors.Is(err, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestCastVoteUseCase_ReturnsErrNotConfirmedParticipant_WhenBearerDeclined(t *testing.T) {
	t.Parallel()
	uc, _, matchRepo, invRepo := newCastVoteUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseNo)

	_, err := uc.Execute(context.Background(), CastVoteInput{
		PlainToken: token, VotedID: "p-2", Score: 4,
	})

	if !errors.Is(err, domainerrors.ErrNotConfirmedParticipant) {
		t.Errorf("expected ErrNotConfirmedParticipant, got %v", err)
	}
}

func TestCastVoteUseCase_ReturnsErrMatchNotCompleted_WhenMatchInProgress(t *testing.T) {
	t.Parallel()
	uc, _, matchRepo, invRepo := newCastVoteUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusInProgress)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	_, err := uc.Execute(context.Background(), CastVoteInput{
		PlainToken: token, VotedID: "p-2", Score: 4,
	})

	if !errors.Is(err, domainerrors.ErrMatchNotCompleted) {
		t.Errorf("expected ErrMatchNotCompleted, got %v", err)
	}
}

func TestCastVoteUseCase_ReturnsErrMatchNotCompleted_WhenMatchClosed(t *testing.T) {
	t.Parallel()
	uc, _, matchRepo, invRepo := newCastVoteUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusClosed)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	_, err := uc.Execute(context.Background(), CastVoteInput{
		PlainToken: token, VotedID: "p-2", Score: 4,
	})

	if !errors.Is(err, domainerrors.ErrMatchNotCompleted) {
		t.Errorf("expected ErrMatchNotCompleted once voting window closed, got %v", err)
	}
}

func TestCastVoteUseCase_ReturnsErrNotTeammates_WhenTargetIsOpponent(t *testing.T) {
	t.Parallel()
	uc, _, matchRepo, invRepo := newCastVoteUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	_, err := uc.Execute(context.Background(), CastVoteInput{
		PlainToken: token, VotedID: "p-3", Score: 4,
	})

	if !errors.Is(err, domainerrors.ErrNotTeammates) {
		t.Errorf("expected ErrNotTeammates, got %v", err)
	}
}

func TestCastVoteUseCase_ReturnsErrNotTeammates_WhenTargetNotInMatch(t *testing.T) {
	t.Parallel()
	uc, _, matchRepo, invRepo := newCastVoteUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	_, err := uc.Execute(context.Background(), CastVoteInput{
		PlainToken: token, VotedID: "stranger", Score: 4,
	})

	if !errors.Is(err, domainerrors.ErrNotTeammates) {
		t.Errorf("expected ErrNotTeammates, got %v", err)
	}
}

func TestCastVoteUseCase_ReturnsErrSelfVote_WhenBearerRatesThemselves(t *testing.T) {
	t.Parallel()
	uc, _, matchRepo, invRepo := newCastVoteUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	_, err := uc.Execute(context.Background(), CastVoteInput{
		PlainToken: token, VotedID: "p-1", Score: 4,
	})

	if !errors.Is(err, domainerrors.ErrSelfVote) {
		t.Errorf("expected ErrSelfVote, got %v", err)
	}
}

func TestCastVoteUseCase_ReturnsErrAlreadyVoted_WhenDuplicate(t *testing.T) {
	t.Parallel()
	uc, _, matchRepo, invRepo := newCastVoteUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	if _, err := uc.Execute(context.Background(), CastVoteInput{
		PlainToken: token, VotedID: "p-2", Score: 4,
	}); err != nil {
		t.Fatalf("first vote unexpected error: %v", err)
	}

	_, err := uc.Execute(context.Background(), CastVoteInput{
		PlainToken: token, VotedID: "p-2", Score: 5,
	})

	if !errors.Is(err, domainerrors.ErrAlreadyVoted) {
		t.Errorf("expected ErrAlreadyVoted, got %v", err)
	}
}

func TestCastVoteUseCase_PropagatesGenericSaveError(t *testing.T) {
	t.Parallel()
	uc, voteRepo, matchRepo, invRepo := newCastVoteUseCase(t)
	saveErr := errors.New("disk full")
	voteRepo.saveErr = saveErr
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	_, err := uc.Execute(context.Background(), CastVoteInput{
		PlainToken: token, VotedID: "p-2", Score: 4,
	})

	if !errors.Is(err, saveErr) {
		t.Errorf("expected wrapped save error, got %v", err)
	}
}
