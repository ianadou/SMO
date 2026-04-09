package entities

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestNewVote_ReturnsVote_WhenInputsAreValid(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 15, 21, 0, 0, 0, time.UTC)

	vote, err := NewVote("vote-1", "match-1", "voter-1", "voted-1", 4, createdAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if vote.ID() != "vote-1" {
		t.Errorf("expected ID 'vote-1', got %q", vote.ID())
	}
	if vote.MatchID() != "match-1" {
		t.Errorf("expected MatchID 'match-1', got %q", vote.MatchID())
	}
	if vote.VoterID() != "voter-1" {
		t.Errorf("expected VoterID 'voter-1', got %q", vote.VoterID())
	}
	if vote.VotedID() != "voted-1" {
		t.Errorf("expected VotedID 'voted-1', got %q", vote.VotedID())
	}
	if vote.Score() != 4 {
		t.Errorf("expected score 4, got %d", vote.Score())
	}
	if !vote.CreatedAt().Equal(createdAt) {
		t.Errorf("expected createdAt %v, got %v", createdAt, vote.CreatedAt())
	}
}

func TestNewVote_AcceptsAllValidScores(t *testing.T) {
	t.Parallel()

	for score := MinVoteScore(); score <= MaxVoteScore(); score++ {
		_, err := NewVote("v-1", "m-1", "voter-1", "voted-1", score, time.Now())
		if err != nil {
			t.Errorf("expected score %d to be valid, got: %v", score, err)
		}
	}
}

func TestNewVote_RejectsSelfVote(t *testing.T) {
	t.Parallel()

	vote, err := NewVote("v-1", "m-1", "player-1", "player-1", 5, time.Now())

	if vote != nil {
		t.Errorf("expected nil vote, got %+v", vote)
	}
	if !errors.Is(err, domainerrors.ErrSelfVote) {
		t.Errorf("expected ErrSelfVote, got %v", err)
	}
}

func TestNewVote_ReturnsError_WhenInputsAreInvalid(t *testing.T) {
	t.Parallel()

	validTime := time.Now()

	cases := []struct {
		name      string
		id        VoteID
		matchID   MatchID
		voterID   PlayerID
		votedID   PlayerID
		score     int
		createdAt time.Time
		wantErr   error
	}{
		{name: "empty id", id: "", matchID: "m-1", voterID: "v-1", votedID: "v-2", score: 3, createdAt: validTime, wantErr: domainerrors.ErrInvalidID},
		{name: "empty match id", id: "v-1", matchID: "", voterID: "v-1", votedID: "v-2", score: 3, createdAt: validTime, wantErr: domainerrors.ErrInvalidID},
		{name: "empty voter id", id: "v-1", matchID: "m-1", voterID: "", votedID: "v-2", score: 3, createdAt: validTime, wantErr: domainerrors.ErrInvalidID},
		{name: "empty voted id", id: "v-1", matchID: "m-1", voterID: "v-1", votedID: "", score: 3, createdAt: validTime, wantErr: domainerrors.ErrInvalidID},
		{name: "score too low", id: "v-1", matchID: "m-1", voterID: "v-1", votedID: "v-2", score: 0, createdAt: validTime, wantErr: domainerrors.ErrInvalidScore},
		{name: "score too high", id: "v-1", matchID: "m-1", voterID: "v-1", votedID: "v-2", score: 6, createdAt: validTime, wantErr: domainerrors.ErrInvalidScore},
		{name: "negative score", id: "v-1", matchID: "m-1", voterID: "v-1", votedID: "v-2", score: -1, createdAt: validTime, wantErr: domainerrors.ErrInvalidScore},
		{name: "zero createdAt", id: "v-1", matchID: "m-1", voterID: "v-1", votedID: "v-2", score: 3, createdAt: time.Time{}, wantErr: domainerrors.ErrInvalidDate},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			vote, err := NewVote(
				testCase.id,
				testCase.matchID,
				testCase.voterID,
				testCase.votedID,
				testCase.score,
				testCase.createdAt,
			)

			if vote != nil {
				t.Errorf("expected nil vote, got %+v", vote)
			}
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("expected error %v, got %v", testCase.wantErr, err)
			}
		})
	}
}
