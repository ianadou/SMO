package match

import (
	"context"
	"fmt"
	"time"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// CreateMatchUseCase orchestrates the creation of a new Match.
//
// A new match always starts in MatchStatusDraft. The match becomes
// visible to players only after the Open transition (see the
// transitions use cases).
type CreateMatchUseCase struct {
	matchRepo ports.MatchRepository
	idGen     ports.IDGenerator
	clock     ports.Clock
}

// CreateMatchInput is the parameter struct for CreateMatchUseCase.Execute.
type CreateMatchInput struct {
	GroupID     entities.GroupID
	Title       string
	Venue       string
	ScheduledAt time.Time
}

// NewCreateMatchUseCase builds a CreateMatchUseCase with the given
// dependencies.
func NewCreateMatchUseCase(
	matchRepo ports.MatchRepository,
	idGen ports.IDGenerator,
	clock ports.Clock,
) *CreateMatchUseCase {
	return &CreateMatchUseCase{
		matchRepo: matchRepo,
		idGen:     idGen,
		clock:     clock,
	}
}

// Execute creates a new Match from the given input.
func (uc *CreateMatchUseCase) Execute(ctx context.Context, input CreateMatchInput) (*entities.Match, error) {
	id := entities.MatchID(uc.idGen.Generate())
	now := uc.clock.Now()

	match, err := entities.NewMatch(
		id,
		input.GroupID,
		input.Title,
		input.Venue,
		input.ScheduledAt,
		entities.MatchStatusDraft,
		nil,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("create match use case: build match: %w", err)
	}

	if saveErr := uc.matchRepo.Save(ctx, match); saveErr != nil {
		return nil, fmt.Errorf("create match use case: save match %q: %w", match.ID(), saveErr)
	}

	return match, nil
}
