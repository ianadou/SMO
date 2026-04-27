package match

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// This file contains the four pure-status transition use cases for the
// Match aggregate. They all follow the same three-step pattern:
//
//   1. Load the match from the repository.
//   2. Invoke the domain transition method on the entity. The entity
//      validates that the transition is allowed from the current state.
//   3. Persist the new status via the repository.
//
// The full lifecycle is:
//
//   Draft → Open → TeamsReady → InProgress → Completed → Closed
//
// The terminal Completed → Closed transition is NOT in this file: it is
// owned by FinalizeMatchUseCase, which also computes the MVP and the
// post-match rankings before persisting the closed state.

// OpenMatchUseCase transitions a match from Draft to Open, making it
// visible to players who can accept invitations.
type OpenMatchUseCase struct {
	matchRepo ports.MatchRepository
}

// NewOpenMatchUseCase builds an OpenMatchUseCase.
func NewOpenMatchUseCase(matchRepo ports.MatchRepository) *OpenMatchUseCase {
	return &OpenMatchUseCase{matchRepo: matchRepo}
}

// Execute opens the match with the given ID.
func (uc *OpenMatchUseCase) Execute(ctx context.Context, id entities.MatchID) (*entities.Match, error) {
	return runTransition(ctx, uc.matchRepo, id, "open", func(m *entities.Match) error {
		return m.Open()
	})
}

// MarkTeamsReadyUseCase transitions a match from Open to TeamsReady,
// signaling that team assignment is complete.
type MarkTeamsReadyUseCase struct {
	matchRepo ports.MatchRepository
}

// NewMarkTeamsReadyUseCase builds a MarkTeamsReadyUseCase.
func NewMarkTeamsReadyUseCase(matchRepo ports.MatchRepository) *MarkTeamsReadyUseCase {
	return &MarkTeamsReadyUseCase{matchRepo: matchRepo}
}

// Execute marks the teams as ready for the match with the given ID.
func (uc *MarkTeamsReadyUseCase) Execute(ctx context.Context, id entities.MatchID) (*entities.Match, error) {
	return runTransition(ctx, uc.matchRepo, id, "mark teams ready", func(m *entities.Match) error {
		return m.MarkTeamsReady()
	})
}

// StartMatchUseCase transitions a match from TeamsReady to InProgress,
// signaling that the match is being played.
type StartMatchUseCase struct {
	matchRepo ports.MatchRepository
}

// NewStartMatchUseCase builds a StartMatchUseCase.
func NewStartMatchUseCase(matchRepo ports.MatchRepository) *StartMatchUseCase {
	return &StartMatchUseCase{matchRepo: matchRepo}
}

// Execute starts the match with the given ID.
func (uc *StartMatchUseCase) Execute(ctx context.Context, id entities.MatchID) (*entities.Match, error) {
	return runTransition(ctx, uc.matchRepo, id, "start", func(m *entities.Match) error {
		return m.Start()
	})
}

// CompleteMatchUseCase transitions a match from InProgress to Completed,
// signaling that the match has ended and voting may open.
type CompleteMatchUseCase struct {
	matchRepo ports.MatchRepository
}

// NewCompleteMatchUseCase builds a CompleteMatchUseCase.
func NewCompleteMatchUseCase(matchRepo ports.MatchRepository) *CompleteMatchUseCase {
	return &CompleteMatchUseCase{matchRepo: matchRepo}
}

// Execute completes the match with the given ID.
func (uc *CompleteMatchUseCase) Execute(ctx context.Context, id entities.MatchID) (*entities.Match, error) {
	return runTransition(ctx, uc.matchRepo, id, "complete", func(m *entities.Match) error {
		return m.Complete()
	})
}

// runTransition is the shared helper that implements the three-step
// pattern used by every transition use case in this file. Having it
// here keeps each use case down to a one-line Execute body and avoids
// duplicating the error-wrapping logic five times.
func runTransition(
	ctx context.Context,
	matchRepo ports.MatchRepository,
	id entities.MatchID,
	verb string,
	transition func(*entities.Match) error,
) (*entities.Match, error) {
	match, err := matchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s match use case: find match %q: %w", verb, id, err)
	}

	if transErr := transition(match); transErr != nil {
		return nil, fmt.Errorf("%s match use case: apply transition on %q: %w", verb, id, transErr)
	}

	if saveErr := matchRepo.UpdateStatus(ctx, match); saveErr != nil {
		return nil, fmt.Errorf("%s match use case: persist status for %q: %w", verb, id, saveErr)
	}

	return match, nil
}
