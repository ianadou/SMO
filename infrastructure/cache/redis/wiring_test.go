package redis_test

import (
	"context"
	"testing"

	rdb "github.com/redis/go-redis/v9"

	"github.com/ianadou/smo/domain/entities"
	cacheredis "github.com/ianadou/smo/infrastructure/cache/redis"
)

// spyGroupRepo records calls to verify the wrapper passed through.
type spyGroupRepo struct {
	findCalls int
}

func (s *spyGroupRepo) Save(context.Context, *entities.Group) error { return nil }
func (s *spyGroupRepo) FindByID(context.Context, entities.GroupID) (*entities.Group, error) {
	s.findCalls++
	return nil, nil
}

func (s *spyGroupRepo) ListByOrganizer(context.Context, entities.OrganizerID) ([]*entities.Group, error) {
	return nil, nil
}
func (s *spyGroupRepo) Update(context.Context, *entities.Group) error  { return nil }
func (s *spyGroupRepo) Delete(context.Context, entities.GroupID) error { return nil }

type spyPlayerRepo struct{}

func (spyPlayerRepo) Save(context.Context, *entities.Player) error { return nil }
func (spyPlayerRepo) FindByID(context.Context, entities.PlayerID) (*entities.Player, error) {
	return nil, nil
}

func (spyPlayerRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Player, error) {
	return nil, nil
}
func (spyPlayerRepo) UpdateRanking(context.Context, *entities.Player) error { return nil }
func (spyPlayerRepo) Delete(context.Context, entities.PlayerID) error       { return nil }

func TestWrapGroupRepository_ReturnsInnerUnchanged_WhenClientIsNil(t *testing.T) {
	t.Parallel()
	spy := &spyGroupRepo{}

	wrapped := cacheredis.WrapGroupRepository(spy, nil)

	// Same pointer means no decoration: the spy is the wired repo.
	if wrapped != spy {
		t.Errorf("expected wrap with nil client to return inner unchanged, got %T", wrapped)
	}
}

func TestWrapGroupRepository_ReturnsCachingDecorator_WhenClientIsSet(t *testing.T) {
	t.Parallel()
	spy := &spyGroupRepo{}
	client := rdb.NewClient(&rdb.Options{Addr: "localhost:1"}) // never used here, no Ping

	wrapped := cacheredis.WrapGroupRepository(spy, client)

	if wrapped == spy {
		t.Errorf("expected wrap with client to return a decorator, got the inner repo")
	}
	_ = client.Close()
}

func TestWrapPlayerRepository_ReturnsInnerUnchanged_WhenClientIsNil(t *testing.T) {
	t.Parallel()
	spy := spyPlayerRepo{}

	wrapped := cacheredis.WrapPlayerRepository(spy, nil)

	if wrapped != spy {
		t.Errorf("expected wrap with nil client to return inner unchanged, got %T", wrapped)
	}
}
