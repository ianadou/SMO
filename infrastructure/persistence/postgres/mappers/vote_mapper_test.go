package mappers

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

func TestVoteToDomain_ReturnsEntity_WhenValid(t *testing.T) {
	t.Parallel()
	row := generated.Votes{
		ID:        "v-1",
		MatchID:   "m-1",
		VoterID:   "p-1",
		VotedID:   "p-2",
		Score:     4,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	v, err := VoteToDomain(row)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if v.Score() != 4 {
		t.Errorf("expected score 4, got %d", v.Score())
	}
}

func TestVoteToDomain_ReturnsError_WhenSelfVote(t *testing.T) {
	t.Parallel()
	row := generated.Votes{
		ID:        "v-1",
		MatchID:   "m-1",
		VoterID:   "p-1",
		VotedID:   "p-1",
		Score:     3,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	_, err := VoteToDomain(row)
	if !errors.Is(err, domainerrors.ErrSelfVote) {
		t.Errorf("expected ErrSelfVote, got %v", err)
	}
}

func TestVoteToCreateParams_BuildsParams(t *testing.T) {
	t.Parallel()
	v, _ := entities.NewVote("v-1", "m-1", "p-1", "p-2", 4, time.Now())
	params := VoteToCreateParams(v)
	if params.Score != 4 {
		t.Errorf("expected 4, got %d", params.Score)
	}
	if params.VoterID != "p-1" || params.VotedID != "p-2" {
		t.Errorf("unexpected voter/voted: %+v", params)
	}
}
