package entities

import (
	"strings"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

const maxTeamLabelLength = 20

// TeamID is the unique identifier of a Team within a match.
type TeamID string

// Team represents one of the two sides of a match. A team is owned by
// exactly one Match and contains a list of player IDs.
//
// We store player IDs rather than full Player entities to keep the team
// independent of the player aggregate: the team only knows who plays for it,
// not the full player data, which is fetched separately when needed.
type Team struct {
	id      TeamID
	matchID MatchID
	label   string
	players []PlayerID
}

// NewTeam builds a Team after validating its inputs.
//
// The players slice is copied defensively to prevent the caller from
// mutating the team's internal state after construction.
func NewTeam(id TeamID, matchID MatchID, label string, players []PlayerID) (*Team, error) {
	if id == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if matchID == "" {
		return nil, domainerrors.ErrInvalidID
	}

	trimmedLabel := strings.TrimSpace(label)
	if trimmedLabel == "" || len(trimmedLabel) > maxTeamLabelLength {
		return nil, domainerrors.ErrInvalidName
	}

	// Defensive copy: caller cannot mutate the team's internal slice
	// by holding a reference to the original.
	playersCopy := make([]PlayerID, len(players))
	copy(playersCopy, players)

	return &Team{
		id:      id,
		matchID: matchID,
		label:   trimmedLabel,
		players: playersCopy,
	}, nil
}

// ID returns the team identifier.
func (t *Team) ID() TeamID { return t.id }

// MatchID returns the identifier of the match this team belongs to.
func (t *Team) MatchID() MatchID { return t.matchID }

// Label returns the team's display label (e.g., "Team A", "Reds").
func (t *Team) Label() string { return t.label }

// Players returns a copy of the player IDs in this team.
//
// The returned slice is a copy: callers can freely mutate it without
// affecting the team's internal state.
func (t *Team) Players() []PlayerID {
	playersCopy := make([]PlayerID, len(t.players))
	copy(playersCopy, t.players)
	return playersCopy
}

// Size returns the number of players in this team.
func (t *Team) Size() int { return len(t.players) }
