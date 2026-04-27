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

const playersByGroupKeyPrefix = "smo:players:group:"

// CachingPlayerRepository caches the per-group player list, which is
// the hottest read path (every match page loads it). Save, UpdateRanking
// and Delete invalidate the affected group's list AFTER the inner
// write succeeds — UpdateRanking matters in particular because the
// FinalizeMatchUseCase calls it for every participant: without
// invalidation, the next ListByGroup would return stale rankings for
// up to the TTL.
type CachingPlayerRepository struct {
	inner  ports.PlayerRepository
	client *rdb.Client
	ttl    time.Duration
}

// NewCachingPlayerRepository wraps inner with cache-aside Redis caching.
func NewCachingPlayerRepository(inner ports.PlayerRepository, client *rdb.Client) *CachingPlayerRepository {
	return &CachingPlayerRepository{
		inner:  inner,
		client: client,
		ttl:    defaultCacheTTL,
	}
}

func playersByGroupKey(groupID entities.GroupID) string {
	return playersByGroupKeyPrefix + string(groupID)
}

// Save delegates and invalidates the affected group's player list.
func (r *CachingPlayerRepository) Save(ctx context.Context, p *entities.Player) error {
	if err := r.inner.Save(ctx, p); err != nil {
		return err
	}
	r.invalidate(ctx, playersByGroupKey(p.GroupID()))
	return nil
}

// FindByID is not cached in this PR. A FindByID is rarely the hot path
// (the frontend goes through groups → players list). Revisit if profiling
// shows otherwise.
func (r *CachingPlayerRepository) FindByID(ctx context.Context, id entities.PlayerID) (*entities.Player, error) {
	return r.inner.FindByID(ctx, id)
}

// ListByGroup looks up Redis first. On miss (or any error), falls
// through to the inner repository and populates the cache.
func (r *CachingPlayerRepository) ListByGroup(ctx context.Context, groupID entities.GroupID) ([]*entities.Player, error) {
	key := playersByGroupKey(groupID)

	if players, hit := r.lookup(ctx, key); hit {
		return players, nil
	}

	players, err := r.inner.ListByGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	r.populate(ctx, key, cachedPlayersFromDomain(players))
	return players, nil
}

// UpdateRanking delegates and invalidates the cache. Critical: without
// this, FinalizeMatchUseCase's ranking updates would be invisible to
// readers for up to the TTL.
func (r *CachingPlayerRepository) UpdateRanking(ctx context.Context, p *entities.Player) error {
	if err := r.inner.UpdateRanking(ctx, p); err != nil {
		return err
	}
	r.invalidate(ctx, playersByGroupKey(p.GroupID()))
	return nil
}

// Delete fetches the player first to know its groupID for invalidation,
// then delegates the actual deletion. If the player does not exist, no
// cache entry to invalidate so a missing-row error is returned as-is.
func (r *CachingPlayerRepository) Delete(ctx context.Context, id entities.PlayerID) error {
	player, findErr := r.inner.FindByID(ctx, id)
	if delErr := r.inner.Delete(ctx, id); delErr != nil {
		return delErr
	}
	if findErr == nil && player != nil {
		r.invalidate(ctx, playersByGroupKey(player.GroupID()))
	}
	return nil
}

func (r *CachingPlayerRepository) lookup(ctx context.Context, key string) ([]*entities.Player, bool) {
	raw, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if !errors.Is(err, rdb.Nil) {
			slog.WarnContext(ctx, "cache get failed, falling back to db",
				slog.String("key", key), slog.String("error", err.Error()))
		}
		return nil, false
	}

	var rows []cachedPlayer
	if jsonErr := json.Unmarshal(raw, &rows); jsonErr != nil {
		slog.WarnContext(ctx, "cache entry corrupted, falling back to db",
			slog.String("key", key), slog.String("error", jsonErr.Error()))
		return nil, false
	}

	players, mapErr := cachedPlayersToDomain(rows)
	if mapErr != nil {
		slog.WarnContext(ctx, "cached players failed domain validation, falling back to db",
			slog.String("key", key), slog.String("error", mapErr.Error()))
		return nil, false
	}
	return players, true
}

func (r *CachingPlayerRepository) populate(ctx context.Context, key string, rows []cachedPlayer) {
	raw, err := json.Marshal(rows)
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

func (r *CachingPlayerRepository) invalidate(ctx context.Context, key string) {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		slog.WarnContext(ctx, "cache invalidation failed",
			slog.String("key", key), slog.String("error", err.Error()))
	}
}
