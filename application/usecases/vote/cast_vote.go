package vote

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// CastVoteUseCase records a peer rating on a completed match.
//
// It enforces the cross-aggregate rule that votes can only be cast on
// matches in the Completed state. The UNIQUE constraint (match, voter,
// voted) is enforced by the repository (and surfaces as ErrAlreadyVoted).
type CastVoteUseCase struct {
	voteRepo  ports.VoteRepository
	matchRepo ports.MatchRepository
	idGen     ports.IDGenerator
	clock     ports.Clock
}

// CastVoteInput is the parameter struct.
type CastVoteInput struct {
	MatchID entities.MatchID
	VoterID entities.PlayerID
	VotedID entities.PlayerID
	Score   int
}

// NewCastVoteUseCase builds the use case.
func NewCastVoteUseCase(
	voteRepo ports.VoteRepository,
	matchRepo ports.MatchRepository,
	idGen ports.IDGenerator,
	clock ports.Clock,
) *CastVoteUseCase {
	return &CastVoteUseCase{voteRepo: voteRepo, matchRepo: matchRepo, idGen: idGen, clock: clock}
}

// Execute casts a new vote if the match is completed.
func (uc *CastVoteUseCase) Execute(ctx context.Context, input CastVoteInput) (*entities.Vote, error) {
	match, err := uc.matchRepo.FindByID(ctx, input.MatchID)
	if err != nil {
		return nil, fmt.Errorf("cast vote use case: find match %q: %w", input.MatchID, err)
	}
	if match.Status() != entities.MatchStatusCompleted {
		return nil, fmt.Errorf("cast vote use case: match %q status is %q: %w",
			input.MatchID, match.Status(), domainerrors.ErrMatchNotCompleted)
	}

	vote, err := entities.NewVote(
		entities.VoteID(uc.idGen.Generate()),
		input.MatchID, input.VoterID, input.VotedID, input.Score,
		uc.clock.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("cast vote use case: build vote: %w", err)
	}

	if saveErr := uc.voteRepo.Save(ctx, vote); saveErr != nil {
		return nil, fmt.Errorf("cast vote use case: save %q: %w", vote.ID(), saveErr)
	}
	return vote, nil
}
