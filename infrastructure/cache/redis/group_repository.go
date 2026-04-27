package redis

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	rdb "github.com/redis/go-redis/v9"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

const (
	defaultCacheTTL = 5 * time.Minute
	groupKeyPrefix  = "smo:group:"
)

// CachingGroupRepository is a cache-aside decorator over any
// GroupRepository implementation (typically the Postgres one). Reads
// of FindByID hit Redis first; misses fall through and populate the
// cache. Writes (Save, Update, Delete) invalidate the cached entry
// AFTER the underlying write succeeds.
//
// Redis is a soft dependency: any Redis error is logged at WARN level
// and the operation falls through to the inner repository. The cache
// is an optimization, never a source of truth.
type CachingGroupRepository struct {
	inner  ports.GroupRepository
	client *rdb.Client
	ttl    time.Duration
}

// NewCachingGroupRepository wraps inner with cache-aside Redis caching
// using the default 5-minute TTL.
func NewCachingGroupRepository(inner ports.GroupRepository, client *rdb.Client) *CachingGroupRepository {
	return &CachingGroupRepository{
		inner:  inner,
		client: client,
		ttl:    defaultCacheTTL,
	}
}

func groupKey(id entities.GroupID) string {
	return groupKeyPrefix + string(id)
}

// Save delegates to the inner repository, then invalidates the cache.
func (r *CachingGroupRepository) Save(ctx context.Context, g *entities.Group) error {
	if err := r.inner.Save(ctx, g); err != nil {
		return err
	}
	r.invalidate(ctx, groupKey(g.ID()))
	return nil
}

// FindByID looks up the group in Redis first. On hit, the cached value
// is returned. On miss (or any Redis error), the inner repository is
// queried and the result is written back to Redis with the TTL.
func (r *CachingGroupRepository) FindByID(ctx context.Context, id entities.GroupID) (*entities.Group, error) {
	key := groupKey(id)

	if g, hit := r.lookup(ctx, key); hit {
		return g, nil
	}

	g, err := r.inner.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	r.populate(ctx, key, cachedGroupFromDomain(g))
	return g, nil
}

// ListByOrganizer is not cached in this PR. Lists are harder to keep
// fresh and the use case is rarer than FindByID; revisit later.
func (r *CachingGroupRepository) ListByOrganizer(ctx context.Context, organizerID entities.OrganizerID) ([]*entities.Group, error) {
	return r.inner.ListByOrganizer(ctx, organizerID)
}

// Update delegates to the inner repository, then invalidates the cache.
func (r *CachingGroupRepository) Update(ctx context.Context, g *entities.Group) error {
	if err := r.inner.Update(ctx, g); err != nil {
		return err
	}
	r.invalidate(ctx, groupKey(g.ID()))
	return nil
}

// Delete delegates to the inner repository, then invalidates the cache.
func (r *CachingGroupRepository) Delete(ctx context.Context, id entities.GroupID) error {
	if err := r.inner.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx, groupKey(id))
	return nil
}

// lookup attempts to fetch and decode a group from Redis. Returns
// (nil, false) on any miss or error: callers must always fall through
// to the inner repository. A corrupted cache entry is treated as a
// miss; the next populate will overwrite it.
func (r *CachingGroupRepository) lookup(ctx context.Context, key string) (*entities.Group, bool) {
	raw, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if !errors.Is(err, rdb.Nil) {
			slog.WarnContext(ctx, "cache get failed, falling back to db",
				slog.String("key", key), slog.String("error", err.Error()))
		}
		return nil, false
	}

	var c cachedGroup
	if jsonErr := json.Unmarshal(raw, &c); jsonErr != nil {
		slog.WarnContext(ctx, "cache entry corrupted, falling back to db",
			slog.String("key", key), slog.String("error", jsonErr.Error()))
		return nil, false
	}

	g, mapErr := cachedGroupToDomain(c)
	if mapErr != nil {
		slog.WarnContext(ctx, "cached group failed domain validation, falling back to db",
			slog.String("key", key), slog.String("error", mapErr.Error()))
		return nil, false
	}
	return g, true
}

func (r *CachingGroupRepository) populate(ctx context.Context, key string, c cachedGroup) {
	raw, err := json.Marshal(c)
	if err != nil {
		slog.WarnContext(ctx, "cache marshal failed",
			slog.String("key", key), slog.String("error", err.Error()))
		return
	}
	if setErr := r.client.Set(ctx, key, raw, r.ttl).Err(); setErr != nil {
		slog.WarnContext(ctx, "cache set failed",
			slog.String("key", key), slog.String("error", setErr.Error()))
	}
}

func (r *CachingGroupRepository) invalidate(ctx context.Context, key string) {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		slog.WarnContext(ctx, "cache invalidation failed",
			slog.String("key", key), slog.String("error", err.Error()))
	}
}
