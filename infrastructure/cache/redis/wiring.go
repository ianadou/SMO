package redis

import (
	rdb "github.com/redis/go-redis/v9"

	"github.com/ianadou/smo/domain/ports"
)

// WrapGroupRepository returns inner unchanged when the Redis client is
// nil (cache disabled, ADR 0002 state 1), or a caching decorator
// wrapping it otherwise. Centralizing this here keeps main.go free of
// nil-checks and lets tests verify the 3-state config with a spy
// instead of reflection.
func WrapGroupRepository(inner ports.GroupRepository, client *rdb.Client) ports.GroupRepository {
	if client == nil {
		return inner
	}
	return NewCachingGroupRepository(inner, client)
}

// WrapPlayerRepository is the player counterpart of WrapGroupRepository.
func WrapPlayerRepository(inner ports.PlayerRepository, client *rdb.Client) ports.PlayerRepository {
	if client == nil {
		return inner
	}
	return NewCachingPlayerRepository(inner, client)
}
