//go:build integration

package repositories_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// TestMain bootstraps the shared Postgres container once for the whole
// integration test package. Spinning up the container takes ~5-10 seconds,
// so reusing it across tests is critical for a usable test suite.
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, pool, err := setupTestContainer(ctx)
	if err != nil {
		log.Fatalf("failed to setup test container: %v", err)
	}
	sharedPool = pool

	code := m.Run()

	pool.Close()
	if termErr := container.Terminate(ctx); termErr != nil {
		log.Printf("failed to terminate container: %v", termErr)
	}
	os.Exit(code)
}

func TestPostgresGroupRepository_Save_PersistsGroup(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	createdAt := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	group, err := entities.NewGroup("group-1", "Test Group", "test-org", createdAt)
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}

	if saveErr := repo.Save(ctx, group); saveErr != nil {
		t.Fatalf("expected Save to succeed, got: %v", saveErr)
	}

	// Verify the group was actually persisted by reading it back.
	found, findErr := repo.FindByID(ctx, "group-1")
	if findErr != nil {
		t.Fatalf("expected to find saved group, got: %v", findErr)
	}
	if found.Name() != "Test Group" {
		t.Errorf("expected name 'Test Group', got %q", found.Name())
	}
	if found.OrganizerID() != "test-org" {
		t.Errorf("expected organizer 'test-org', got %q", found.OrganizerID())
	}
}

func TestPostgresGroupRepository_Save_ReturnsErrReferencedEntityNotFound_WhenOrganizerMissing(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	group, _ := entities.NewGroup("group-1", "Orphan Group", "nonexistent-org", time.Now())

	err := repo.Save(ctx, group)

	if !errors.Is(err, domainerrors.ErrReferencedEntityNotFound) {
		t.Errorf("expected ErrReferencedEntityNotFound, got %v", err)
	}
}

func TestPostgresGroupRepository_FindByID_ReturnsErrGroupNotFound_WhenMissing(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "does-not-exist")

	if !errors.Is(err, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestPostgresGroupRepository_ListByOrganizer_ReturnsAllGroupsForOrganizer(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	for i, name := range []string{"Group A", "Group B", "Group C"} {
		group, _ := entities.NewGroup(
			entities.GroupID([]string{"g-1", "g-2", "g-3"}[i]),
			name,
			"test-org",
			time.Now(),
		)
		if err := repo.Save(ctx, group); err != nil {
			t.Fatalf("setup: failed to save group %d: %v", i, err)
		}
	}

	groups, err := repo.ListByOrganizer(ctx, "test-org")
	if err != nil {
		t.Fatalf("expected ListByOrganizer to succeed, got: %v", err)
	}
	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}
}

func TestPostgresGroupRepository_Update_ChangesGroupName(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	original, _ := entities.NewGroup("group-1", "Old Name", "test-org", time.Now())
	if err := repo.Save(ctx, original); err != nil {
		t.Fatalf("setup: %v", err)
	}

	updated, _ := entities.NewGroup("group-1", "New Name", "test-org", original.CreatedAt())
	if err := repo.Update(ctx, updated); err != nil {
		t.Fatalf("expected Update to succeed, got: %v", err)
	}

	found, _ := repo.FindByID(ctx, "group-1")
	if found.Name() != "New Name" {
		t.Errorf("expected name 'New Name', got %q", found.Name())
	}
}

func TestPostgresGroupRepository_Delete_RemovesGroup(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	group, _ := entities.NewGroup("group-1", "To Delete", "test-org", time.Now())
	if err := repo.Save(ctx, group); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := repo.Delete(ctx, "group-1"); err != nil {
		t.Fatalf("expected Delete to succeed, got: %v", err)
	}

	_, findErr := repo.FindByID(ctx, "group-1")
	if !errors.Is(findErr, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound after delete, got %v", findErr)
	}
}
