package entities

import (
	"strings"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

const (
	maxPlayerNameLength  = 50
	defaultPlayerRanking = 1000
)

// PlayerID is the unique identifier of a Player.
type PlayerID string

// Player represents a participant in a match. Players do not have an
// account; they are referenced by an organizer who adds them to a group
// or invites them via a token.
//
// Player is distinct from Organizer: an organizer who participates in
// their own match exists as both an Organizer entity (for authentication)
// and a Player entity (for the match logic).
type Player struct {
	id      PlayerID
	groupID GroupID
	name    string
	ranking int
}

// NewPlayer builds a Player after validating its inputs.
//
// The ranking parameter is the player's current ranking score. New players
// who have never played should be created with DefaultPlayerRanking.
func NewPlayer(id PlayerID, groupID GroupID, name string, ranking int) (*Player, error) {
	if id == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if groupID == "" {
		return nil, domainerrors.ErrInvalidID
	}

	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > maxPlayerNameLength {
		return nil, domainerrors.ErrInvalidName
	}

	return &Player{
		id:      id,
		groupID: groupID,
		name:    trimmedName,
		ranking: ranking,
	}, nil
}

// DefaultPlayerRanking returns the starting ranking for a brand new player.
func DefaultPlayerRanking() int { return defaultPlayerRanking }

// ID returns the player identifier.
func (p *Player) ID() PlayerID { return p.id }

// GroupID returns the identifier of the group this player belongs to.
func (p *Player) GroupID() GroupID { return p.groupID }

// Name returns the player's display name.
func (p *Player) Name() string { return p.name }

// Ranking returns the player's current ranking score.
func (p *Player) Ranking() int { return p.ranking }
