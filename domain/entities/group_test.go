package entities

import (
	"errors"
	"strings"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestNewGroup_ReturnsGroup_WhenInputsAreValid(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	group, err := NewGroup("group-123", "Foot du jeudi", "org-456", createdAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if group.ID() != "group-123" {
		t.Errorf("expected ID 'group-123', got %q", group.ID())
	}
	if group.Name() != "Foot du jeudi" {
		t.Errorf("expected name 'Foot du jeudi', got %q", group.Name())
	}
	if group.OrganizerID() != "org-456" {
		t.Errorf("expected organizer ID 'org-456', got %q", group.OrganizerID())
	}
	if !group.CreatedAt().Equal(createdAt) {
		t.Errorf("expected createdAt %v, got %v", createdAt, group.CreatedAt())
	}
}

func TestNewGroup_TrimsWhitespaceAroundName(t *testing.T) {
	t.Parallel()

	group, err := NewGroup("group-1", "  Mon Groupe  ", "org-1", time.Now())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if group.Name() != "Mon Groupe" {
		t.Errorf("expected trimmed name 'Mon Groupe', got %q", group.Name())
	}
}

func TestNewGroup_ReturnsError_WhenInputsAreInvalid(t *testing.T) {
	t.Parallel()

	validTime := time.Now()
	longName := strings.Repeat("a", 101)

	cases := []struct {
		name        string
		id          GroupID
		groupName   string
		organizerID OrganizerID
		createdAt   time.Time
		wantErr     error
	}{
		{
			name:        "empty id",
			id:          "",
			groupName:   "Valid",
			organizerID: "org-1",
			createdAt:   validTime,
			wantErr:     domainerrors.ErrInvalidID,
		},
		{
			name:        "empty name",
			id:          "group-1",
			groupName:   "",
			organizerID: "org-1",
			createdAt:   validTime,
			wantErr:     domainerrors.ErrInvalidName,
		},
		{
			name:        "whitespace-only name",
			id:          "group-1",
			groupName:   "    ",
			organizerID: "org-1",
			createdAt:   validTime,
			wantErr:     domainerrors.ErrInvalidName,
		},
		{
			name:        "name too long",
			id:          "group-1",
			groupName:   longName,
			organizerID: "org-1",
			createdAt:   validTime,
			wantErr:     domainerrors.ErrInvalidName,
		},
		{
			name:        "empty organizer id",
			id:          "group-1",
			groupName:   "Valid",
			organizerID: "",
			createdAt:   validTime,
			wantErr:     domainerrors.ErrInvalidID,
		},
		{
			name:        "zero createdAt",
			id:          "group-1",
			groupName:   "Valid",
			organizerID: "org-1",
			createdAt:   time.Time{},
			wantErr:     domainerrors.ErrInvalidDate,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			group, err := NewGroup(testCase.id, testCase.groupName, testCase.organizerID, testCase.createdAt)

			if group != nil {
				t.Errorf("expected nil group, got %+v", group)
			}
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("expected error %v, got %v", testCase.wantErr, err)
			}
		})
	}
}
