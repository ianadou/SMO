package invitation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestGetInvitationUseCase_Execute_ReturnsInvitation_WhenExists(t *testing.T) {
	t.Parallel()
	repo := newFakeInvitationRepository()
	createdAt := time.Now()
	inv, _ := entities.NewInvitation("inv-1", "m-1", "p-1", "hash", createdAt.Add(5*24*time.Hour), nil, createdAt)
	_ = repo.Save(context.Background(), inv)

	uc := NewGetInvitationUseCase(repo)
	found, err := uc.Execute(context.Background(), "inv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID() != "inv-1" {
		t.Errorf("expected 'inv-1', got %q", found.ID())
	}
}

func TestGetInvitationUseCase_Execute_ReturnsErrInvitationNotFound(t *testing.T) {
	t.Parallel()
	uc := NewGetInvitationUseCase(newFakeInvitationRepository())
	_, err := uc.Execute(context.Background(), "nonexistent")
	if !errors.Is(err, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}
