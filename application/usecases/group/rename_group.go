package group

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// RenameGroupUseCase changes a group's display name on behalf of its
// owner. Ownership is enforced here, not in the handler: a group that
// belongs to another organizer is reported as not found rather than
// forbidden, so the endpoint never reveals whether a group ID exists.
type RenameGroupUseCase struct {
	groupRepo ports.GroupRepository
}

// NewRenameGroupUseCase builds the use case.
func NewRenameGroupUseCase(groupRepo ports.GroupRepository) *RenameGroupUseCase {
	return &RenameGroupUseCase{groupRepo: groupRepo}
}

// Execute renames the group after checking the requester owns it.
// Returns ErrGroupNotFound for an unknown OR foreign group, and
// ErrInvalidName when the new name violates the entity rules.
func (uc *RenameGroupUseCase) Execute(
	ctx context.Context,
	groupID entities.GroupID,
	organizerID entities.OrganizerID,
	name string,
) (*entities.Group, error) {
	groupToRename, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("rename group use case: find group: %w", err)
	}

	if groupToRename.OrganizerID() != organizerID {
		return nil, fmt.Errorf("rename group use case: group %q not owned by requester: %w",
			groupID, domainerrors.ErrGroupNotFound)
	}

	if renameErr := groupToRename.Rename(name); renameErr != nil {
		return nil, fmt.Errorf("rename group use case: %w", renameErr)
	}

	if updateErr := uc.groupRepo.Update(ctx, groupToRename); updateErr != nil {
		return nil, fmt.Errorf("rename group use case: persist: %w", updateErr)
	}

	return groupToRename, nil
}
