package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// LoginOrganizerUseCase orchestrates the login flow: look up the
// organizer by email, verify the password, sign a JWT.
//
// Both "email not found" and "wrong password" surface as
// ErrInvalidCredentials so the caller cannot distinguish the two — this
// prevents enumeration of valid emails.
type LoginOrganizerUseCase struct {
	organizerRepo ports.OrganizerRepository
	hasher        ports.PasswordHasher
	signer        ports.JWTSigner
}

// LoginOrganizerInput is the parameter struct for Execute.
type LoginOrganizerInput struct {
	Email    string
	Password string
}

// LoginOrganizerOutput is the success result of Execute: the issued
// token and the matching organizer entity. Handlers typically expose
// only the token + a small projection of the organizer (id, email,
// display name) — never the password hash.
type LoginOrganizerOutput struct {
	Token     string
	Organizer *entities.Organizer
}

// NewLoginOrganizerUseCase builds a LoginOrganizerUseCase.
func NewLoginOrganizerUseCase(
	organizerRepo ports.OrganizerRepository,
	hasher ports.PasswordHasher,
	signer ports.JWTSigner,
) *LoginOrganizerUseCase {
	return &LoginOrganizerUseCase{
		organizerRepo: organizerRepo,
		hasher:        hasher,
		signer:        signer,
	}
}

// Execute logs the organizer in.
func (uc *LoginOrganizerUseCase) Execute(ctx context.Context, in LoginOrganizerInput) (*LoginOrganizerOutput, error) {
	organizer, err := uc.organizerRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrOrganizerNotFound) {
			return nil, fmt.Errorf("login organizer: %w", domainerrors.ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("login organizer: find: %w", err)
	}

	if compareErr := uc.hasher.Compare(organizer.PasswordHash(), in.Password); compareErr != nil {
		if errors.Is(compareErr, domainerrors.ErrInvalidCredentials) {
			return nil, fmt.Errorf("login organizer: %w", domainerrors.ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("login organizer: compare password: %w", compareErr)
	}

	token, err := uc.signer.Sign(organizer.ID())
	if err != nil {
		return nil, fmt.Errorf("login organizer: sign token: %w", err)
	}

	return &LoginOrganizerOutput{Token: token, Organizer: organizer}, nil
}
