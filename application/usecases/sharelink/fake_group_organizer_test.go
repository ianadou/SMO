package sharelink

import (
	"context"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// fakeGroupRepository is a minimal GroupRepository for share link use
// case tests. Only FindByID is exercised; the rest of the port is stubbed.
type fakeGroupRepository struct {
	groups map[entities.GroupID]*entities.Group
}

func newFakeGroupRepository() *fakeGroupRepository {
	return &fakeGroupRepository{groups: make(map[entities.GroupID]*entities.Group)}
}

// seedGroup seeds group "group-1" owned by "org-1". The ids are fixed
// because every test in the package revolves around that single group;
// ownership rejections pass a different organizer id to Execute instead.
func (r *fakeGroupRepository) seedGroup(t testHelper) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	g, err := entities.NewGroup("group-1", "Sunday League", "org-1", "", now)
	if err != nil {
		t.Fatalf("seedGroup: %v", err)
	}
	r.groups[g.ID()] = g
}

func (r *fakeGroupRepository) FindByID(_ context.Context, id entities.GroupID) (*entities.Group, error) {
	g, ok := r.groups[id]
	if !ok {
		return nil, domainerrors.ErrGroupNotFound
	}
	return g, nil
}

func (r *fakeGroupRepository) Save(context.Context, *entities.Group) error    { return nil }
func (r *fakeGroupRepository) Update(context.Context, *entities.Group) error  { return nil }
func (r *fakeGroupRepository) Delete(context.Context, entities.GroupID) error { return nil }
func (r *fakeGroupRepository) ListByOrganizer(context.Context, entities.OrganizerID) ([]*entities.Group, error) {
	return nil, nil
}

// fakeOrganizerRepository is a minimal OrganizerRepository for share
// link use case tests. Only FindByID is exercised.
type fakeOrganizerRepository struct {
	organizers map[entities.OrganizerID]*entities.Organizer
}

func newFakeOrganizerRepository() *fakeOrganizerRepository {
	return &fakeOrganizerRepository{organizers: make(map[entities.OrganizerID]*entities.Organizer)}
}

func (r *fakeOrganizerRepository) seedOrganizer(t testHelper, id entities.OrganizerID, displayName string) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	o, err := entities.NewOrganizer(id, "organizer@example.com", displayName, "hash", now)
	if err != nil {
		t.Fatalf("seedOrganizer: %v", err)
	}
	r.organizers[id] = o
}

func (r *fakeOrganizerRepository) FindByID(_ context.Context, id entities.OrganizerID) (*entities.Organizer, error) {
	o, ok := r.organizers[id]
	if !ok {
		return nil, domainerrors.ErrOrganizerNotFound
	}
	return o, nil
}

func (r *fakeOrganizerRepository) Save(context.Context, *entities.Organizer) error { return nil }
func (r *fakeOrganizerRepository) FindByEmail(context.Context, string) (*entities.Organizer, error) {
	return nil, domainerrors.ErrOrganizerNotFound
}
