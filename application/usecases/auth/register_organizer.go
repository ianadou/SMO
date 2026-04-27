package auth

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// minPasswordLength is enforced at the application boundary. The domain
// entity does not enforce password rules because by the time a hash
// reaches NewOrganizer the plain text is gone — only this use case sees
// the plain password.
const minPasswordLength = 12

// RegisterOrganizerUseCase orchestrates registration of a new organizer:
// validate the password, hash it, build the entity, and persist it.
type RegisterOrganizerUseCase struct {
	organizerRepo ports.OrganizerRepository
	hasher        ports.PasswordHasher
	idGen         ports.IDGenerator
	clock         ports.Clock
}

// RegisterOrganizerInput is the parameter struct for Execute.
type RegisterOrganizerInput struct {
	Email       string
	Password    string
	DisplayName string
}

// NewRegisterOrganizerUseCase builds a RegisterOrganizerUseCase.
func NewRegisterOrganizerUseCase(
	organizerRepo ports.OrganizerRepository,
	hasher ports.PasswordHasher,
	idGen ports.IDGenerator,
	clock ports.Clock,
) *RegisterOrganizerUseCase {
	return &RegisterOrganizerUseCase{
		organizerRepo: organizerRepo,
		hasher:        hasher,
		idGen:         idGen,
		clock:         clock,
	}
}

// Execute registers a new organizer.
func (uc *RegisterOrganizerUseCase) Execute(ctx context.Context, in RegisterOrganizerInput) (*entities.Organizer, error) {
	if len(in.Password) < minPasswordLength {
		return nil, fmt.Errorf("register organizer: %w: minimum %d characters", domainerrors.ErrInvalidPassword, minPasswordLength)
	}

	hash, err := uc.hasher.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("register organizer: hash password: %w", err)
	}

	organizer, err := entities.NewOrganizer(
		entities.OrganizerID(uc.idGen.Generate()),
		in.Email,
		in.DisplayName,
		hash,
		uc.clock.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("register organizer: build entity: %w", err)
	}

	if saveErr := uc.organizerRepo.Save(ctx, organizer); saveErr != nil {
		return nil, fmt.Errorf("register organizer: save: %w", saveErr)
	}

	return organizer, nil
}
