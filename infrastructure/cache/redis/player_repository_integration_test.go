//go:build integration

package redis_test

import (
	"context"
	"testing"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	cacheredis "github.com/ianadou/smo/infrastructure/cache/redis"
)

type fakePlayerRepo struct {
	players       map[entities.PlayerID]*entities.Player
	listCalls     int
	groupContents map[entities.GroupID][]*entities.Player
}

func newFakePlayerRepo() *fakePlayerRepo {
	return &fakePlayerRepo{
		players:       make(map[entities.PlayerID]*entities.Player),
		groupContents: make(map[entities.GroupID][]*entities.Player),
	}
}

func (r *fakePlayerRepo) Save(_ context.Context, p *entities.Player) error {
	r.players[p.ID()] = p
	r.groupContents[p.GroupID()] = append(r.groupContents[p.GroupID()], p)
	return nil
}

func (r *fakePlayerRepo) FindByID(_ context.Context, id entities.PlayerID) (*entities.Player, error) {
	p, ok := r.players[id]
	if !ok {
		return nil, domainerrors.ErrPlayerNotFound
	}
	return p, nil
}

func (r *fakePlayerRepo) ListByGroup(_ context.Context, groupID entities.GroupID) ([]*entities.Player, error) {
	r.listCalls++
	return r.groupContents[groupID], nil
}

func (r *fakePlayerRepo) UpdateRanking(_ context.Context, p *entities.Player) error {
	r.players[p.ID()] = p
	// Replace in groupContents: rebuild the slice with the new ranking.
	updated := make([]*entities.Player, 0, len(r.groupContents[p.GroupID()]))
	for _, existing := range r.groupContents[p.GroupID()] {
		if existing.ID() == p.ID() {
			updated = append(updated, p)
		} else {
			updated = append(updated, existing)
		}
	}
	r.groupContents[p.GroupID()] = updated
	return nil
}

func (r *fakePlayerRepo) Delete(_ context.Context, id entities.PlayerID) error {
	delete(r.players, id)
	return nil
}

func TestCachingPlayerRepository_ListByGroup_HitAfterMiss(t *testing.T) {
	client, _ := startRedis(t)
	inner := newFakePlayerRepo()
	p, _ := entities.NewPlayer("p-1", "g-1", "Alice", 1000)
	_ = inner.Save(context.Background(), p)
	caching := cacheredis.NewCachingPlayerRepository(inner, client)
	ctx := context.Background()

	if _, err := caching.ListByGroup(ctx, "g-1"); err != nil {
		t.Fatalf("first ListByGroup: %v", err)
	}
	if inner.listCalls != 1 {
		t.Fatalf("expected 1 inner call after miss, got %d", inner.listCalls)
	}

	if _, err := caching.ListByGroup(ctx, "g-1"); err != nil {
		t.Fatalf("second ListByGroup: %v", err)
	}
	if inner.listCalls != 1 {
		t.Errorf("expected cached read to skip inner, got %d", inner.listCalls)
	}
}

func TestCachingPlayerRepository_UpdateRanking_InvalidatesGroupListCache(t *testing.T) {
	client, _ := startRedis(t)
	inner := newFakePlayerRepo()
	p, _ := entities.NewPlayer("p-1", "g-1", "Alice", 1000)
	_ = inner.Save(context.Background(), p)
	caching := cacheredis.NewCachingPlayerRepository(inner, client)
	ctx := context.Background()

	// Warm cache.
	if _, err := caching.ListByGroup(ctx, "g-1"); err != nil {
		t.Fatalf("warm: %v", err)
	}

	// FinalizeMatch-like update of the player's ranking.
	updated, _ := entities.NewPlayer("p-1", "g-1", "Alice", 1200)
	if err := caching.UpdateRanking(ctx, updated); err != nil {
		t.Fatalf("UpdateRanking: %v", err)
	}

	// The next list MUST reflect the new ranking, not the cached 1000.
	got, err := caching.ListByGroup(ctx, "g-1")
	if err != nil {
		t.Fatalf("post-update ListByGroup: %v", err)
	}
	if len(got) != 1 || got[0].Ranking() != 1200 {
		t.Errorf("expected ranking 1200 after invalidation, got %v", got)
	}
}
