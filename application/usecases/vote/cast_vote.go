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
// The voter is never trusted from the request: their identity is derived
// from the invitation token (the same capability used to accept the
// invitation), which makes vote spoofing impossible by construction.
// The use case enforces that the bearer confirmed attendance, that the
// match is in the Completed state, and that voter and voted player are
// on the same team (players rate teammates, never opponents). The
// UNIQUE constraint (match, voter, voted) is enforced by the repository
// and surfaces as ErrAlreadyVoted.
type CastVoteUseCase struct {
	voteRepo  ports.VoteRepository
	matchRepo ports.MatchRepository
	invRepo   ports.InvitationRepository
	tokens    ports.InvitationTokenService
	idGen     ports.IDGenerator
	clock     ports.Clock
}

// CastVoteInput is the parameter struct. PlainToken identifies (and
// authenticates) the voter; the match is the one the token was issued
// for.
type CastVoteInput struct {
	PlainToken string
	VotedID    entities.PlayerID
	Score      int
}

// NewCastVoteUseCase builds the use case.
func NewCastVoteUseCase(
	voteRepo ports.VoteRepository,
	matchRepo ports.MatchRepository,
	invRepo ports.InvitationRepository,
	tokens ports.InvitationTokenService,
	idGen ports.IDGenerator,
	clock ports.Clock,
) *CastVoteUseCase {
	return &CastVoteUseCase{
		voteRepo:  voteRepo,
		matchRepo: matchRepo,
		invRepo:   invRepo,
		tokens:    tokens,
		idGen:     idGen,
		clock:     clock,
	}
}

// Execute resolves the token bearer and casts their vote. Returns
// ErrInvitationNotFound for an unknown token, ErrNotConfirmedParticipant
// when the bearer never confirmed attendance, ErrMatchNotCompleted
// outside the voting window, ErrNotTeammates when the target is not on
// the bearer's team, and ErrAlreadyVoted on duplicates.
func (uc *CastVoteUseCase) Execute(ctx context.Context, input CastVoteInput) (*entities.Vote, error) {
	invitation, err := uc.invRepo.FindByTokenHash(ctx, uc.tokens.HashToken(input.PlainToken))
	if err != nil {
		return nil, fmt.Errorf("cast vote use case: find invitation by hash: %w", err)
	}
	if !invitation.IsConfirmed() {
		return nil, fmt.Errorf("cast vote use case: bearer declined or never answered: %w",
			domainerrors.ErrNotConfirmedParticipant)
	}

	match, err := uc.matchRepo.FindByID(ctx, invitation.MatchID())
	if err != nil {
		return nil, fmt.Errorf("cast vote use case: find match %q: %w", invitation.MatchID(), err)
	}
	if match.Status() != entities.MatchStatusCompleted {
		return nil, fmt.Errorf("cast vote use case: match %q status is %q: %w",
			match.ID(), match.Status(), domainerrors.ErrMatchNotCompleted)
	}

	voterID := invitation.PlayerID()
	if teammateErr := requireTeammates(match, voterID, input.VotedID); teammateErr != nil {
		return nil, fmt.Errorf("cast vote use case: %w", teammateErr)
	}

	vote, err := entities.NewVote(
		entities.VoteID(uc.idGen.Generate()),
		match.ID(), voterID, input.VotedID, input.Score,
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

// requireTeammates rejects a vote whose target is not on the voter's
// team. Self-votes pass this check (same side) and are rejected by
// entities.NewVote, which owns that invariant.
func requireTeammates(match *entities.Match, voterID, votedID entities.PlayerID) error {
	voterSide, voterAssigned := match.TeamOf(voterID)
	votedSide, votedAssigned := match.TeamOf(votedID)
	if !voterAssigned || !votedAssigned || voterSide != votedSide {
		return fmt.Errorf("player %q cannot rate %q: %w",
			voterID, votedID, domainerrors.ErrNotTeammates)
	}
	return nil
}
