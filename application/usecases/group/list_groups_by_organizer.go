package group

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// ListGroupsByOrganizerUseCase returns every group owned by a given organizer.
//
// Like GetGroupUseCase this is intentionally a thin pass-through: the
// repository contract guarantees the ordering (newest first) and the
// scoping (only groups owned by the organizer). Keeping it as its own
// use case provides a stable boundary the HTTP handler can depend on
// without reaching into the port directly.
type ListGroupsByOrganizerUseCase struct {
	groupRepo ports.GroupRepository
}

// NewListGroupsByOrganizerUseCase builds the use case with the given
// repository.
func NewListGroupsByOrganizerUseCase(groupRepo ports.GroupRepository) *ListGroupsByOrganizerUseCase {
	return &ListGroupsByOrganizerUseCase{groupRepo: groupRepo}
}

// Execute returns the groups owned by organizerID, newest first.
//
// Returns an empty slice (never nil) when the organizer has no groups.
func (uc *ListGroupsByOrganizerUseCase) Execute(ctx context.Context, organizerID entities.OrganizerID) ([]*entities.Group, error) {
	groups, err := uc.groupRepo.ListByOrganizer(ctx, organizerID)
	if err != nil {
		return nil, fmt.Errorf("list groups by organizer use case: organizer %q: %w", organizerID, err)
	}
	return groups, nil
}
