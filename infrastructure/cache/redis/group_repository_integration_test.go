//go:build integration

package redis_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	rdb "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	cacheredis "github.com/ianadou/smo/infrastructure/cache/redis"
)

// init disables the Ryuk reaper container before any test in this
// package starts.
//
// Ryuk fails to start on Fedora 43 + Docker 29 (known upstream bug),
// causing every testcontainers Run() call to error with "container
// is not running". Disabling Ryuk works around the local issue; the
// per-test t.Cleanup(_ = container.Terminate(ctx)) handles the
// cleanup that Ryuk would otherwise have done.
//
// Mirrors the same workaround used in
// infrastructure/persistence/postgres/repositories/integration_helpers_test.go
// and cmd/server/main_integration_test.go.
func init() {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

// fakeGroupRepo is a controllable in-memory GroupRepository used as
// the inner repo behind the caching decorator. It records call counts
// so tests can assert that cache hits avoid the inner call.
type fakeGroupRepo struct {
	groups    map[entities.GroupID]*entities.Group
	findCalls int
}

func newFakeGroupRepo() *fakeGroupRepo {
	return &fakeGroupRepo{groups: make(map[entities.GroupID]*entities.Group)}
}

func (r *fakeGroupRepo) Save(_ context.Context, g *entities.Group) error {
	r.groups[g.ID()] = g
	return nil
}

func (r *fakeGroupRepo) FindByID(_ context.Context, id entities.GroupID) (*entities.Group, error) {
	r.findCalls++
	g, ok := r.groups[id]
	if !ok {
		return nil, domainerrors.ErrGroupNotFound
	}
	return g, nil
}

func (r *fakeGroupRepo) ListByOrganizer(context.Context, entities.OrganizerID) ([]*entities.Group, error) {
	return nil, nil
}

func (r *fakeGroupRepo) Update(_ context.Context, g *entities.Group) error {
	r.groups[g.ID()] = g
	return nil
}

func (r *fakeGroupRepo) Delete(_ context.Context, id entities.GroupID) error {
	delete(r.groups, id)
	return nil
}

func startRedis(t *testing.T) (*rdb.Client, testcontainers.Container) {
	t.Helper()
	ctx := context.Background()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("start redis container: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	url, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get redis URL: %v", err)
	}

	client, err := cacheredis.Connect(ctx, url)
	if err != nil {
		t.Fatalf("connect to redis: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return client, container
}

func seedGroup(t *testing.T, repo *fakeGroupRepo, id entities.GroupID) {
	t.Helper()
	g, err := entities.NewGroup(id, "Test", "org-1", "", time.Now())
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	_ = repo.Save(context.Background(), g)
}

func TestCachingGroupRepository_FindByID_HitAfterMiss_DoesNotCallInnerTwice(t *testing.T) {
	client, _ := startRedis(t)
	inner := newFakeGroupRepo()
	seedGroup(t, inner, "g-1")
	caching := cacheredis.NewCachingGroupRepository(inner, client)
	ctx := context.Background()

	// First call: miss → inner is hit.
	if _, err := caching.FindByID(ctx, "g-1"); err != nil {
		t.Fatalf("first FindByID: %v", err)
	}
	if inner.findCalls != 1 {
		t.Fatalf("expected 1 inner call after miss, got %d", inner.findCalls)
	}

	// Second call: hit → inner is NOT hit again.
	if _, err := caching.FindByID(ctx, "g-1"); err != nil {
		t.Fatalf("second FindByID: %v", err)
	}
	if inner.findCalls != 1 {
		t.Errorf("expected cached read to skip inner, but inner.findCalls=%d", inner.findCalls)
	}
}

func TestCachingGroupRepository_Save_InvalidatesCachedEntry(t *testing.T) {
	client, _ := startRedis(t)
	inner := newFakeGroupRepo()
	seedGroup(t, inner, "g-1")
	caching := cacheredis.NewCachingGroupRepository(inner, client)
	ctx := context.Background()

	// Warm the cache.
	if _, err := caching.FindByID(ctx, "g-1"); err != nil {
		t.Fatalf("warm: %v", err)
	}
	innerCallsBefore := inner.findCalls

	// Save a different group with the same ID — invalidation must drop
	// the cached value so the next read goes back to the inner repo.
	updated, _ := entities.NewGroup("g-1", "Renamed", "org-1", "", time.Now())
	if err := caching.Save(ctx, updated); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := caching.FindByID(ctx, "g-1")
	if err != nil {
		t.Fatalf("post-save FindByID: %v", err)
	}
	if got.Name() != "Renamed" {
		t.Errorf("expected fresh value 'Renamed' after invalidation, got %q", got.Name())
	}
	if inner.findCalls <= innerCallsBefore {
		t.Errorf("expected inner to be re-hit after invalidation")
	}
}

func TestCachingGroupRepository_FallsThroughToInner_WhenRedisGoesDown(t *testing.T) {
	client, container := startRedis(t)
	inner := newFakeGroupRepo()
	seedGroup(t, inner, "g-1")
	caching := cacheredis.NewCachingGroupRepository(inner, client)
	ctx := context.Background()

	// Verify the wiring works while Redis is up.
	if _, err := caching.FindByID(ctx, "g-1"); err != nil {
		t.Fatalf("warm: %v", err)
	}

	// Kill Redis: simulate runtime failure (ADR 0002 state 3).
	if err := container.Terminate(ctx); err != nil {
		t.Fatalf("terminate redis: %v", err)
	}

	// Reads must still succeed via the inner repo, with no error
	// surfaced to the caller. A WARN log is expected internally.
	got, err := caching.FindByID(ctx, "g-1")
	if err != nil {
		t.Fatalf("expected fall-through to succeed, got %v", err)
	}
	if got == nil || got.ID() != "g-1" {
		t.Errorf("expected group g-1 from inner, got %v", got)
	}
}

func TestCachingGroupRepository_FindByID_PropagatesNotFound_FromInner(t *testing.T) {
	client, _ := startRedis(t)
	inner := newFakeGroupRepo()
	caching := cacheredis.NewCachingGroupRepository(inner, client)

	_, err := caching.FindByID(context.Background(), "missing")

	if !errors.Is(err, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound to propagate, got %v", err)
	}
}
