package group

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

type erroringGroupRepository struct {
	fakeGroupRepository
	listErr error
}

func (r *erroringGroupRepository) ListByOrganizer(_ context.Context, _ entities.OrganizerID) ([]*entities.Group, error) {
	return nil, r.listErr
}

func TestListGroupsByOrganizerUseCase_Execute_ReturnsOnlyTheOrganizerGroups(t *testing.T) {
	t.Parallel()

	repo := newFakeGroupRepository()
	mine, _ := entities.NewGroup("g-1", "Mine", "org-1", "", time.Now())
	other, _ := entities.NewGroup("g-2", "Other", "org-2", "", time.Now())
	alsoMine, _ := entities.NewGroup("g-3", "AlsoMine", "org-1", "", time.Now())
	_ = repo.Save(context.Background(), mine)
	_ = repo.Save(context.Background(), other)
	_ = repo.Save(context.Background(), alsoMine)

	useCase := NewListGroupsByOrganizerUseCase(repo)

	got, err := useCase.Execute(context.Background(), "org-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 groups for org-1, got %d", len(got))
	}
	for _, g := range got {
		if g.OrganizerID() != "org-1" {
			t.Errorf("expected only org-1 groups, got organizer %q", g.OrganizerID())
		}
	}
}

func TestListGroupsByOrganizerUseCase_Execute_WrapsRepositoryError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("postgres connection lost")
	repo := &erroringGroupRepository{listErr: repoErr}
	useCase := NewListGroupsByOrganizerUseCase(repo)

	got, err := useCase.Execute(context.Background(), "org-1")
	if got != nil {
		t.Errorf("expected nil result on error, got %+v", got)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
	if !strings.Contains(err.Error(), "list groups by organizer use case") {
		t.Errorf("expected use case prefix in error message, got %q", err.Error())
	}
}

func TestListGroupsByOrganizerUseCase_Execute_ReturnsEmptySlice_WhenOrganizerHasNoGroups(t *testing.T) {
	t.Parallel()

	repo := newFakeGroupRepository()
	useCase := NewListGroupsByOrganizerUseCase(repo)

	got, err := useCase.Execute(context.Background(), "org-with-nothing")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got == nil {
		t.Errorf("expected non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d groups", len(got))
	}
}
