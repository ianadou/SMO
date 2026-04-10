package group

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestCreateGroupUseCase_Execute_ReturnsGroup_WhenInputIsValid(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2026, 4, 10, 14, 0, 0, 0, time.UTC)
	repo := newFakeGroupRepository()
	idGen := newFakeIDGenerator("group-fixed-id")
	clock := newFakeClock(fixedTime)
	useCase := NewCreateGroupUseCase(repo, idGen, clock)

	input := CreateGroupInput{
		Name:        "Foot du jeudi",
		OrganizerID: "org-1",
	}

	group, err := useCase.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if group.ID() != "group-fixed-id" {
		t.Errorf("expected ID 'group-fixed-id', got %q", group.ID())
	}
	if group.Name() != "Foot du jeudi" {
		t.Errorf("expected name 'Foot du jeudi', got %q", group.Name())
	}
	if group.OrganizerID() != "org-1" {
		t.Errorf("expected OrganizerID 'org-1', got %q", group.OrganizerID())
	}
	if !group.CreatedAt().Equal(fixedTime) {
		t.Errorf("expected createdAt %v, got %v", fixedTime, group.CreatedAt())
	}
}

func TestCreateGroupUseCase_Execute_PersistsGroup(t *testing.T) {
	t.Parallel()

	repo := newFakeGroupRepository()
	idGen := newFakeIDGenerator("group-persisted")
	clock := newFakeClock(time.Now())
	useCase := NewCreateGroupUseCase(repo, idGen, clock)

	_, err := useCase.Execute(context.Background(), CreateGroupInput{
		Name:        "Persisted Group",
		OrganizerID: "org-1",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stored, findErr := repo.FindByID(context.Background(), "group-persisted")
	if findErr != nil {
		t.Fatalf("expected group to be persisted, got error: %v", findErr)
	}
	if stored.Name() != "Persisted Group" {
		t.Errorf("expected stored name 'Persisted Group', got %q", stored.Name())
	}
}

func TestCreateGroupUseCase_Execute_ReturnsError_WhenNameIsInvalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		groupName string
	}{
		{name: "empty name", groupName: ""},
		{name: "whitespace-only name", groupName: "   "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Each sub-test gets its own fakes so the use case is fully
			// isolated. Sharing a fakeIDGenerator across sub-tests would
			// cause it to run out of IDs since each Execute call consumes
			// one regardless of validation outcome.
			repo := newFakeGroupRepository()
			idGen := newFakeIDGenerator("group-1")
			clock := newFakeClock(time.Now())
			useCase := NewCreateGroupUseCase(repo, idGen, clock)

			group, err := useCase.Execute(context.Background(), CreateGroupInput{
				Name:        testCase.groupName,
				OrganizerID: "org-1",
			})

			if group != nil {
				t.Errorf("expected nil group, got %+v", group)
			}
			if !errors.Is(err, domainerrors.ErrInvalidName) {
				t.Errorf("expected ErrInvalidName, got %v", err)
			}
		})
	}
}

func TestCreateGroupUseCase_Execute_ReturnsError_WhenOrganizerIDIsEmpty(t *testing.T) {
	t.Parallel()

	repo := newFakeGroupRepository()
	idGen := newFakeIDGenerator("group-1")
	clock := newFakeClock(time.Now())
	useCase := NewCreateGroupUseCase(repo, idGen, clock)

	group, err := useCase.Execute(context.Background(), CreateGroupInput{
		Name:        "Valid",
		OrganizerID: "",
	})

	if group != nil {
		t.Errorf("expected nil group, got %+v", group)
	}
	if !errors.Is(err, domainerrors.ErrInvalidID) {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}
}

func TestCreateGroupUseCase_Execute_DoesNotPersistOnValidationError(t *testing.T) {
	t.Parallel()

	repo := newFakeGroupRepository()
	idGen := newFakeIDGenerator("group-should-not-exist")
	clock := newFakeClock(time.Now())
	useCase := NewCreateGroupUseCase(repo, idGen, clock)

	_, _ = useCase.Execute(context.Background(), CreateGroupInput{
		Name:        "",
		OrganizerID: "org-1",
	})

	// Verify nothing was persisted despite the ID being generated.
	_, findErr := repo.FindByID(context.Background(), "group-should-not-exist")
	if !errors.Is(findErr, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", findErr)
	}
}

// failingGroupRepository is a repository that always fails on Save,
// used to verify error propagation from the persistence layer.
type failingGroupRepository struct {
	*fakeGroupRepository
	saveErr error
}

func (r *failingGroupRepository) Save(_ context.Context, _ *entities.Group) error {
	return r.saveErr
}

func TestCreateGroupUseCase_Execute_PropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database is unreachable")
	repo := &failingGroupRepository{
		fakeGroupRepository: newFakeGroupRepository(),
		saveErr:             expectedErr,
	}
	idGen := newFakeIDGenerator("group-1")
	clock := newFakeClock(time.Now())
	useCase := NewCreateGroupUseCase(repo, idGen, clock)

	group, err := useCase.Execute(context.Background(), CreateGroupInput{
		Name:        "Valid",
		OrganizerID: "org-1",
	})

	if group != nil {
		t.Errorf("expected nil group, got %+v", group)
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap %v, got %v", expectedErr, err)
	}
}
