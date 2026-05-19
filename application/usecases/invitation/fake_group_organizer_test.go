package invitation

import (
	"context"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// fakeGroupRepo is a minimal GroupRepository for invitation use case
// tests. Only FindByID is exercised; the rest of the port is stubbed.
type fakeGroupRepo struct {
	groups map[entities.GroupID]*entities.Group
}

func newFakeGroupRepo() *fakeGroupRepo {
	return &fakeGroupRepo{groups: make(map[entities.GroupID]*entities.Group)}
}

func (r *fakeGroupRepo) seedGroup(t testHelper, id entities.GroupID, organizerID entities.OrganizerID, name string) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	g, err := entities.NewGroup(id, name, organizerID, "", now)
	if err != nil {
		t.Fatalf("seedGroup: %v", err)
	}
	r.groups[id] = g
}

func (r *fakeGroupRepo) FindByID(_ context.Context, id entities.GroupID) (*entities.Group, error) {
	g, ok := r.groups[id]
	if !ok {
		return nil, domainerrors.ErrGroupNotFound
	}
	return g, nil
}

func (r *fakeGroupRepo) Save(context.Context, *entities.Group) error    { return nil }
func (r *fakeGroupRepo) Update(context.Context, *entities.Group) error  { return nil }
func (r *fakeGroupRepo) Delete(context.Context, entities.GroupID) error { return nil }
func (r *fakeGroupRepo) ListByOrganizer(context.Context, entities.OrganizerID) ([]*entities.Group, error) {
	return nil, nil
}

// fakeOrganizerRepo is a minimal OrganizerRepository for invitation use
// case tests. Only FindByID is exercised; the rest of the port is stubbed.
type fakeOrganizerRepo struct {
	organizers map[entities.OrganizerID]*entities.Organizer
}

func newFakeOrganizerRepo() *fakeOrganizerRepo {
	return &fakeOrganizerRepo{organizers: make(map[entities.OrganizerID]*entities.Organizer)}
}

func (r *fakeOrganizerRepo) seedOrganizer(t testHelper, id entities.OrganizerID, displayName string) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	o, err := entities.NewOrganizer(id, "organizer@example.com", displayName, "hash", now)
	if err != nil {
		t.Fatalf("seedOrganizer: %v", err)
	}
	r.organizers[id] = o
}

func (r *fakeOrganizerRepo) FindByID(_ context.Context, id entities.OrganizerID) (*entities.Organizer, error) {
	o, ok := r.organizers[id]
	if !ok {
		return nil, domainerrors.ErrOrganizerNotFound
	}
	return o, nil
}

func (r *fakeOrganizerRepo) Save(context.Context, *entities.Organizer) error { return nil }
func (r *fakeOrganizerRepo) FindByEmail(_ context.Context, _ string) (*entities.Organizer, error) {
	return nil, domainerrors.ErrOrganizerNotFound
}
