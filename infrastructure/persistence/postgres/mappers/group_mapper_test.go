package mappers

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

func TestGroupToDomain_ReturnsEntity_WhenRowIsValid(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)
	row := generated.Groups{
		ID:          "group-1",
		OrganizerID: "org-1",
		Name:        "Foot du jeudi",
		CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
	}

	group, err := GroupToDomain(row)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if group.ID() != "group-1" {
		t.Errorf("expected ID 'group-1', got %q", group.ID())
	}
	if group.OrganizerID() != "org-1" {
		t.Errorf("expected OrganizerID 'org-1', got %q", group.OrganizerID())
	}
	if group.Name() != "Foot du jeudi" {
		t.Errorf("expected name 'Foot du jeudi', got %q", group.Name())
	}
	if !group.CreatedAt().Equal(createdAt) {
		t.Errorf("expected createdAt %v, got %v", createdAt, group.CreatedAt())
	}
}

func TestGroupToDomain_ReturnsError_WhenRowHasEmptyName(t *testing.T) {
	t.Parallel()

	row := generated.Groups{
		ID:          "group-1",
		OrganizerID: "org-1",
		Name:        "",
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	group, err := GroupToDomain(row)

	if group != nil {
		t.Errorf("expected nil group, got %+v", group)
	}
	if !errors.Is(err, domainerrors.ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestGroupToDomain_ReturnsError_WhenRowHasZeroCreatedAt(t *testing.T) {
	t.Parallel()

	row := generated.Groups{
		ID:          "group-1",
		OrganizerID: "org-1",
		Name:        "Valid name",
		CreatedAt:   pgtype.Timestamptz{Time: time.Time{}, Valid: true},
	}

	group, err := GroupToDomain(row)

	if group != nil {
		t.Errorf("expected nil group, got %+v", group)
	}
	if !errors.Is(err, domainerrors.ErrInvalidDate) {
		t.Errorf("expected ErrInvalidDate, got %v", err)
	}
}

func TestGroupToCreateParams_BuildsParamsFromEntity(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)
	group, err := entities.NewGroup("group-1", "Foot du jeudi", "org-1", "", createdAt)
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}

	params := GroupToCreateParams(group)

	if params.ID != "group-1" {
		t.Errorf("expected ID 'group-1', got %q", params.ID)
	}
	if params.OrganizerID != "org-1" {
		t.Errorf("expected OrganizerID 'org-1', got %q", params.OrganizerID)
	}
	if params.Name != "Foot du jeudi" {
		t.Errorf("expected name 'Foot du jeudi', got %q", params.Name)
	}
	if !params.CreatedAt.Valid {
		t.Errorf("expected CreatedAt to be marked Valid")
	}
	if !params.CreatedAt.Time.Equal(createdAt) {
		t.Errorf("expected CreatedAt %v, got %v", createdAt, params.CreatedAt.Time)
	}
}

func TestGroupToUpdateParams_BuildsParamsFromEntity(t *testing.T) {
	t.Parallel()

	group, err := entities.NewGroup("group-1", "New name", "org-1", "", time.Now())
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}

	params := GroupToUpdateParams(group)

	if params.ID != "group-1" {
		t.Errorf("expected ID 'group-1', got %q", params.ID)
	}
	if params.Name != "New name" {
		t.Errorf("expected name 'New name', got %q", params.Name)
	}
}

func TestGroupRoundTrip_DomainToParamsToDomain(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)
	original, _ := entities.NewGroup("group-1", "Foot du jeudi", "org-1", "https://discord.com/api/webhooks/123456789/round-trip-token", createdAt)

	// Domain → CreateParams → Groups (via direct conversion since both
	// structs have identical fields) → Domain. The round trip must
	// preserve all data.
	params := GroupToCreateParams(original)
	simulatedRow := generated.Groups(params)

	roundTripped, err := GroupToDomain(simulatedRow)
	if err != nil {
		t.Fatalf("round trip failed: %v", err)
	}

	if roundTripped.ID() != original.ID() {
		t.Errorf("ID changed: %q → %q", original.ID(), roundTripped.ID())
	}
	if roundTripped.OrganizerID() != original.OrganizerID() {
		t.Errorf("OrganizerID changed: %q → %q", original.OrganizerID(), roundTripped.OrganizerID())
	}
	if roundTripped.Name() != original.Name() {
		t.Errorf("Name changed: %q → %q", original.Name(), roundTripped.Name())
	}
	if !roundTripped.CreatedAt().Equal(original.CreatedAt()) {
		t.Errorf("CreatedAt changed: %v → %v", original.CreatedAt(), roundTripped.CreatedAt())
	}
	if roundTripped.WebhookURL() != original.WebhookURL() {
		t.Errorf("WebhookURL changed: %q → %q", original.WebhookURL(), roundTripped.WebhookURL())
	}
}
