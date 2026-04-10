package group

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// GetGroupUseCase retrieves a Group by its identifier.
//
// This use case is intentionally trivial — it just delegates to the
// repository — but having it as a dedicated use case keeps the
// architecture consistent and provides a stable API for callers
// (HTTP handlers, CLI tools, etc.) that should not depend directly
// on the repository.
type GetGroupUseCase struct {
	groupRepo ports.GroupRepository
}

// NewGetGroupUseCase builds a GetGroupUseCase with the given repository.
func NewGetGroupUseCase(groupRepo ports.GroupRepository) *GetGroupUseCase {
	return &GetGroupUseCase{groupRepo: groupRepo}
}

// Execute returns the group with the given ID.
//
// Returns errors.ErrGroupNotFound (wrapped) if no group exists with
// that ID.
func (uc *GetGroupUseCase) Execute(ctx context.Context, id entities.GroupID) (*entities.Group, error) {
	group, err := uc.groupRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get group use case: find group %q: %w", id, err)
	}
	return group, nil
}
