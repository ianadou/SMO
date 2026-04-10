package group

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// CreateGroupUseCase orchestrates the creation of a new Group.
//
// It generates an ID via the injected IDGenerator, gets the current
// time via the injected Clock, builds the entity through the
// NewGroup constructor (which validates business invariants), and
// persists it via the GroupRepository port.
type CreateGroupUseCase struct {
	groupRepo ports.GroupRepository
	idGen     ports.IDGenerator
	clock     ports.Clock
}

// CreateGroupInput is the parameter struct for CreateGroupUseCase.Execute.
//
// Encapsulating parameters in a struct keeps the use case API
// extensible: new fields can be added without breaking existing
// callers.
type CreateGroupInput struct {
	Name        string
	OrganizerID entities.OrganizerID
}

// NewCreateGroupUseCase builds a CreateGroupUseCase with the given
// dependencies. All dependencies are required; passing nil for any of
// them is a programming error and will cause a panic at first use.
func NewCreateGroupUseCase(
	groupRepo ports.GroupRepository,
	idGen ports.IDGenerator,
	clock ports.Clock,
) *CreateGroupUseCase {
	return &CreateGroupUseCase{
		groupRepo: groupRepo,
		idGen:     idGen,
		clock:     clock,
	}
}

// Execute creates a new Group from the given input.
//
// Returns an error if:
//   - the input fails domain validation (empty name, etc.) — wrapped
//     from entities.NewGroup
//   - the repository fails to persist the group
func (uc *CreateGroupUseCase) Execute(ctx context.Context, input CreateGroupInput) (*entities.Group, error) {
	id := entities.GroupID(uc.idGen.Generate())
	now := uc.clock.Now()

	group, err := entities.NewGroup(id, input.Name, input.OrganizerID, now)
	if err != nil {
		return nil, fmt.Errorf("create group use case: build group: %w", err)
	}

	if saveErr := uc.groupRepo.Save(ctx, group); saveErr != nil {
		return nil, fmt.Errorf("create group use case: save group %q: %w", group.ID(), saveErr)
	}

	return group, nil
}
