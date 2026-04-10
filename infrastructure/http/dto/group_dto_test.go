package dto

import (
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

func TestGroupResponseFromEntity_BuildsResponseFromGroup(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 4, 10, 14, 0, 0, 0, time.UTC)
	group, err := entities.NewGroup("group-1", "Foot du jeudi", "org-1", createdAt)
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}

	response := GroupResponseFromEntity(group)

	if response.ID != "group-1" {
		t.Errorf("expected ID 'group-1', got %q", response.ID)
	}
	if response.Name != "Foot du jeudi" {
		t.Errorf("expected name 'Foot du jeudi', got %q", response.Name)
	}
	if response.OrganizerID != "org-1" {
		t.Errorf("expected OrganizerID 'org-1', got %q", response.OrganizerID)
	}
	if !response.CreatedAt.Equal(createdAt) {
		t.Errorf("expected CreatedAt %v, got %v", createdAt, response.CreatedAt)
	}
}
