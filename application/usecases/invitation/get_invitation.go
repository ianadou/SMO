package invitation

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// GetInvitationUseCase retrieves an invitation by ID. Returns the
// stored hash, never a plain token.
type GetInvitationUseCase struct {
	repo ports.InvitationRepository
}

// NewGetInvitationUseCase builds the use case.
func NewGetInvitationUseCase(repo ports.InvitationRepository) *GetInvitationUseCase {
	return &GetInvitationUseCase{repo: repo}
}

// Execute returns the invitation with the given ID.
func (uc *GetInvitationUseCase) Execute(ctx context.Context, id entities.InvitationID) (*entities.Invitation, error) {
	inv, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get invitation use case: find %q: %w", id, err)
	}
	return inv, nil
}
