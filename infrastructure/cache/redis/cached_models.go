package redis

import (
	"time"

	"github.com/ianadou/smo/domain/entities"
)

// cachedGroup is the JSON-serializable shadow of entities.Group used
// for Redis storage. JSON tags live HERE, never on the domain entity
// (CLAUDE.md architecture rule 1: the domain has no infrastructure
// concerns, including serialization).
type cachedGroup struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OrganizerID string    `json:"organizer_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func cachedGroupFromDomain(g *entities.Group) cachedGroup {
	return cachedGroup{
		ID:          string(g.ID()),
		Name:        g.Name(),
		OrganizerID: string(g.OrganizerID()),
		CreatedAt:   g.CreatedAt(),
	}
}

func cachedGroupToDomain(c cachedGroup) (*entities.Group, error) {
	return entities.NewGroup(
		entities.GroupID(c.ID),
		c.Name,
		entities.OrganizerID(c.OrganizerID),
		c.CreatedAt,
	)
}

// cachedPlayer is the JSON-serializable shadow of entities.Player.
type cachedPlayer struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
	Name    string `json:"name"`
	Ranking int    `json:"ranking"`
}

func cachedPlayerFromDomain(p *entities.Player) cachedPlayer {
	return cachedPlayer{
		ID:      string(p.ID()),
		GroupID: string(p.GroupID()),
		Name:    p.Name(),
		Ranking: p.Ranking(),
	}
}

func cachedPlayerToDomain(c cachedPlayer) (*entities.Player, error) {
	return entities.NewPlayer(
		entities.PlayerID(c.ID),
		entities.GroupID(c.GroupID),
		c.Name,
		c.Ranking,
	)
}

// cachedPlayersFromDomain converts a slice of domain Players into a
// slice of cached players for storage. Used by ListByGroup caching.
func cachedPlayersFromDomain(players []*entities.Player) []cachedPlayer {
	out := make([]cachedPlayer, 0, len(players))
	for _, p := range players {
		out = append(out, cachedPlayerFromDomain(p))
	}
	return out
}

// cachedPlayersToDomain rebuilds a slice of domain Players from cached
// rows. Returns the first construction error encountered (if any), so
// a corrupted cache entry triggers a fall-through to the database.
func cachedPlayersToDomain(rows []cachedPlayer) ([]*entities.Player, error) {
	out := make([]*entities.Player, 0, len(rows))
	for _, row := range rows {
		p, err := cachedPlayerToDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}
