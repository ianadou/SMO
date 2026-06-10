package vote

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func newVoteContextUseCase(t *testing.T) (
	*GetVoteContextUseCase,
	*fakeInvitationRepository,
	*fakeMatchRepository,
	*fakeVoteRepository,
) {
	t.Helper()
	invRepo := newFakeInvitationRepository()
	matchRepo := newFakeMatchRepository()
	voteRepo := newFakeVoteRepository()
	groupRepo := newFakeGroupRepository()

	group, err := entities.NewGroup("g-1", "Foot du jeudi", "org-1", "", voteTestKickoff.Add(-30*24*time.Hour))
	if err != nil {
		t.Fatalf("seed group: %v", err)
	}
	groupRepo.addGroup(group)

	uc := NewGetVoteContextUseCase(invRepo, matchRepo, groupRepo, voteRepo, fakeTokenService{})
	return uc, invRepo, matchRepo, voteRepo
}

func seedRoster(matchRepo *fakeMatchRepository) {
	matchRepo.members = []entities.MatchTeamMember{
		{PlayerID: "p-1", PlayerName: "Alice", Team: "A", Slot: 0},
		{PlayerID: "p-2", PlayerName: "Bob", Team: "A", Slot: 1},
		{PlayerID: "p-3", PlayerName: "Carol", Team: "B", Slot: 0},
		{PlayerID: "p-4", PlayerName: "Dan", Team: "B", Slot: 1},
	}
}

func mustVote(t *testing.T, id entities.VoteID, matchID entities.MatchID, voter, voted entities.PlayerID, score int) *entities.Vote {
	t.Helper()
	v, err := entities.NewVote(id, matchID, voter, voted, score, voteTestKickoff.Add(3*time.Hour))
	if err != nil {
		t.Fatalf("seed vote: %v", err)
	}
	return v
}

func TestGetVoteContext_ReturnsMatchBlockOnly_WhenMatchNotCompletedYet(t *testing.T) {
	t.Parallel()
	uc, invRepo, matchRepo, _ := newVoteContextUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusOpen)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	pageContext, err := uc.Execute(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageContext.Status != entities.MatchStatusOpen {
		t.Errorf("expected status open, got %q", pageContext.Status)
	}
	if pageContext.GroupName != "Foot du jeudi" {
		t.Errorf("expected group name, got %q", pageContext.GroupName)
	}
	if pageContext.Teammates != nil {
		t.Errorf("expected no teammates before completion, got %v", pageContext.Teammates)
	}
	if pageContext.Results != nil {
		t.Error("expected nil results before completion")
	}
}

func TestGetVoteContext_ListsTeammatesWithNamesAndCounts_WhenCompleted(t *testing.T) {
	t.Parallel()
	uc, invRepo, matchRepo, _ := newVoteContextUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	seedRoster(matchRepo)
	matchRepo.togetherCounts = map[entities.PlayerID]int{"p-2": 7}
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	pageContext, err := uc.Execute(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageContext.Voter.Name != "Alice" || pageContext.Voter.Team != entities.TeamSideA {
		t.Errorf("expected voter Alice on side A, got %+v", pageContext.Voter)
	}
	if len(pageContext.Teammates) != 1 {
		t.Fatalf("expected exactly 1 teammate in a 2v2, got %d", len(pageContext.Teammates))
	}
	teammate := pageContext.Teammates[0]
	if teammate.PlayerID != "p-2" || teammate.Name != "Bob" || teammate.MatchesTogether != 7 {
		t.Errorf("expected Bob with 7 matches together, got %+v", teammate)
	}
	if teammate.YourScore != nil {
		t.Errorf("expected no score before voting, got %v", *teammate.YourScore)
	}
	if pageContext.VotersTotal != 4 || pageContext.VotersDone != 0 {
		t.Errorf("expected progress 0/4, got %d/%d", pageContext.VotersDone, pageContext.VotersTotal)
	}
}

func TestGetVoteContext_MarksAlreadyRatedTeammates_WithBearerScore(t *testing.T) {
	t.Parallel()
	uc, invRepo, matchRepo, voteRepo := newVoteContextUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	seedRoster(matchRepo)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)
	if err := voteRepo.Save(context.Background(),
		mustVote(t, "v-1", votableMatchID, "p-1", "p-2", 4)); err != nil {
		t.Fatalf("seed vote: %v", err)
	}

	pageContext, err := uc.Execute(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageContext.Teammates[0].YourScore == nil || *pageContext.Teammates[0].YourScore != 4 {
		t.Errorf("expected your score 4 on Bob, got %+v", pageContext.Teammates[0])
	}
	if pageContext.VotersDone != 1 {
		t.Errorf("expected 1 distinct voter, got %d", pageContext.VotersDone)
	}
}

func TestGetVoteContext_CountsDistinctVoters_ForProgress(t *testing.T) {
	t.Parallel()
	uc, invRepo, matchRepo, voteRepo := newVoteContextUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	seedRoster(matchRepo)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)
	for _, v := range []*entities.Vote{
		mustVote(t, "v-1", votableMatchID, "p-1", "p-2", 4),
		mustVote(t, "v-2", votableMatchID, "p-3", "p-4", 5),
		mustVote(t, "v-3", votableMatchID, "p-4", "p-3", 2),
	} {
		if err := voteRepo.Save(context.Background(), v); err != nil {
			t.Fatalf("seed vote: %v", err)
		}
	}

	pageContext, err := uc.Execute(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageContext.VotersDone != 3 {
		t.Errorf("expected 3 distinct voters, got %d", pageContext.VotersDone)
	}
}

func TestGetVoteContext_ComputesResultsWithDelta_WhenClosed(t *testing.T) {
	t.Parallel()
	uc, invRepo, matchRepo, voteRepo := newVoteContextUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusClosed)
	seedRoster(matchRepo)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	previousScore := 1
	previous, err := entities.RehydrateMatch(entities.MatchSnapshot{
		ID: "previous-match", GroupID: "g-1", Title: "Avant", Venue: "Stade",
		ScheduledAt: voteTestKickoff.Add(-7 * 24 * time.Hour),
		Status:      entities.MatchStatusClosed,
		ScoreA:      &previousScore, ScoreB: new(int),
		CreatedAt: voteTestKickoff.Add(-8 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("seed previous match: %v", err)
	}
	matchRepo.previousMatch = previous

	for _, v := range []*entities.Vote{
		mustVote(t, "v-1", votableMatchID, "p-1", "p-2", 4),
		mustVote(t, "v-2", votableMatchID, "p-2", "p-1", 3),
		mustVote(t, "v-3", "previous-match", "p-1", "p-2", 3),
	} {
		if err := voteRepo.Save(context.Background(), v); err != nil {
			t.Fatalf("seed vote: %v", err)
		}
	}

	pageContext, err := uc.Execute(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageContext.Results == nil {
		t.Fatal("expected results on a closed match")
	}
	bob := pageContext.Results.Teammates[0]
	if bob.Average != 4 || bob.VotesCount != 1 {
		t.Errorf("expected Bob average 4 from 1 vote, got %+v", bob)
	}
	if bob.Delta == nil || math.Abs(*bob.Delta-1.0) > 1e-9 {
		t.Errorf("expected Bob delta +1.0 (4 now vs 3 before), got %v", bob.Delta)
	}
	if pageContext.Results.Self.Average == nil || *pageContext.Results.Self.Average != 3 {
		t.Errorf("expected self average 3, got %v", pageContext.Results.Self.Average)
	}
	if pageContext.Results.Self.VotesCount != 1 {
		t.Errorf("expected self votes count 1, got %d", pageContext.Results.Self.VotesCount)
	}
}

func TestGetVoteContext_DeltaNil_WhenNoPreviousMatch(t *testing.T) {
	t.Parallel()
	uc, invRepo, matchRepo, voteRepo := newVoteContextUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusClosed)
	seedRoster(matchRepo)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)
	if err := voteRepo.Save(context.Background(),
		mustVote(t, "v-1", votableMatchID, "p-1", "p-2", 4)); err != nil {
		t.Fatalf("seed vote: %v", err)
	}

	pageContext, err := uc.Execute(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageContext.Results.Teammates[0].Delta != nil {
		t.Errorf("expected nil delta without a previous match, got %v",
			*pageContext.Results.Teammates[0].Delta)
	}
}

func TestGetVoteContext_SelfAverageNil_WhenBearerReceivedNoVotes(t *testing.T) {
	t.Parallel()
	uc, invRepo, matchRepo, _ := newVoteContextUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusClosed)
	seedRoster(matchRepo)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseYes)

	pageContext, err := uc.Execute(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageContext.Results.Self.Average != nil {
		t.Errorf("expected nil self average with no votes, got %v", *pageContext.Results.Self.Average)
	}
	if pageContext.Results.Self.VotesCount != 0 {
		t.Errorf("expected 0 self votes, got %d", pageContext.Results.Self.VotesCount)
	}
}

func TestGetVoteContext_ReturnsErrNotConfirmedParticipant_WhenBearerDeclined(t *testing.T) {
	t.Parallel()
	uc, invRepo, matchRepo, _ := newVoteContextUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)
	token := seedBearerInvitation(t, invRepo, entities.InvitationResponseNo)

	_, err := uc.Execute(context.Background(), token)

	if !errors.Is(err, domainerrors.ErrNotConfirmedParticipant) {
		t.Errorf("expected ErrNotConfirmedParticipant, got %v", err)
	}
}

func TestGetVoteContext_ReturnsErrInvitationNotFound_WhenTokenUnknown(t *testing.T) {
	t.Parallel()
	uc, _, matchRepo, _ := newVoteContextUseCase(t)
	seedMatchInStatus(t, matchRepo, entities.MatchStatusCompleted)

	_, err := uc.Execute(context.Background(), "tok-forged")

	if !errors.Is(err, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}
