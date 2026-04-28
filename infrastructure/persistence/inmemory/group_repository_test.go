package inmemory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func newGroup(t *testing.T, id entities.GroupID, organizerID entities.OrganizerID) *entities.Group {
	t.Helper()
	g, err := entities.NewGroup(id, "Test group", organizerID, "", time.Now())
	if err != nil {
		t.Fatalf("NewGroup: %v", err)
	}
	return g
}

func TestGroupRepository_Save_PersistsGroup_AndFindByIDReturnsIt(t *testing.T) {
	t.Parallel()

	repo := NewGroupRepository()
	g := newGroup(t, "g-1", "org-1")

	if err := repo.Save(context.Background(), g); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(context.Background(), "g-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.ID() != "g-1" || got.OrganizerID() != "org-1" {
		t.Errorf("unexpected group: %+v", got)
	}
}

func TestGroupRepository_FindByID_ReturnsErrGroupNotFound_WhenMissing(t *testing.T) {
	t.Parallel()

	repo := NewGroupRepository()

	_, err := repo.FindByID(context.Background(), "nope")

	if !errors.Is(err, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestGroupRepository_Save_OverwritesExistingGroup(t *testing.T) {
	t.Parallel()

	repo := NewGroupRepository()
	original := newGroup(t, "g-1", "org-1")
	if err := repo.Save(context.Background(), original); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	// A second Save with the same ID is allowed and overwrites the
	// previous entry — that's the documented behaviour of Save.
	updated, err := entities.NewGroup("g-1", "Renamed", "org-1", "", time.Now())
	if err != nil {
		t.Fatalf("NewGroup: %v", err)
	}
	if err := repo.Save(context.Background(), updated); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	got, _ := repo.FindByID(context.Background(), "g-1")
	if got.Name() != "Renamed" {
		t.Errorf("expected name 'Renamed' after overwrite, got %q", got.Name())
	}
}

func TestGroupRepository_ListByOrganizer_ReturnsOnlyOwnedGroups(t *testing.T) {
	t.Parallel()

	repo := NewGroupRepository()
	_ = repo.Save(context.Background(), newGroup(t, "g-1", "org-1"))
	_ = repo.Save(context.Background(), newGroup(t, "g-2", "org-1"))
	_ = repo.Save(context.Background(), newGroup(t, "g-3", "org-2"))

	got, err := repo.ListByOrganizer(context.Background(), "org-1")
	if err != nil {
		t.Fatalf("ListByOrganizer: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 groups for org-1, got %d", len(got))
	}
	for _, g := range got {
		if g.OrganizerID() != "org-1" {
			t.Errorf("returned a group from a different organizer: %+v", g)
		}
	}
}

func TestGroupRepository_ListByOrganizer_ReturnsEmpty_WhenNoMatch(t *testing.T) {
	t.Parallel()

	repo := NewGroupRepository()
	_ = repo.Save(context.Background(), newGroup(t, "g-1", "org-1"))

	got, err := repo.ListByOrganizer(context.Background(), "org-other")
	if err != nil {
		t.Fatalf("ListByOrganizer: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 groups, got %d", len(got))
	}
}

func TestGroupRepository_Update_OverwritesGroup_WhenItExists(t *testing.T) {
	t.Parallel()

	repo := NewGroupRepository()
	original := newGroup(t, "g-1", "org-1")
	_ = repo.Save(context.Background(), original)

	updated, err := entities.NewGroup("g-1", "New name", "org-1", "", time.Now())
	if err != nil {
		t.Fatalf("NewGroup: %v", err)
	}
	if err := repo.Update(context.Background(), updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.FindByID(context.Background(), "g-1")
	if got.Name() != "New name" {
		t.Errorf("expected name 'New name', got %q", got.Name())
	}
}

func TestGroupRepository_Update_ReturnsErrGroupNotFound_WhenMissing(t *testing.T) {
	t.Parallel()

	repo := NewGroupRepository()
	g := newGroup(t, "g-1", "org-1")

	err := repo.Update(context.Background(), g)

	if !errors.Is(err, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestGroupRepository_Delete_RemovesGroup(t *testing.T) {
	t.Parallel()

	repo := NewGroupRepository()
	_ = repo.Save(context.Background(), newGroup(t, "g-1", "org-1"))

	if err := repo.Delete(context.Background(), "g-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(context.Background(), "g-1")
	if !errors.Is(err, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound after delete, got %v", err)
	}
}

func TestGroupRepository_Delete_IsIdempotent_OnMissingID(t *testing.T) {
	t.Parallel()

	repo := NewGroupRepository()

	// Documented behaviour: Delete on a missing ID is a no-op, not
	// an error. The contract matches the Postgres adapter's
	// idempotent delete.
	if err := repo.Delete(context.Background(), "nope"); err != nil {
		t.Errorf("expected nil for missing-ID delete, got %v", err)
	}
}
