package entities

import (
	"fmt"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

// AssignTeams sets the two-team composition of the match.
//
// Teams may only be (re)assigned while the match is open; from
// teams_ready onward the rosters are frozen (ErrTeamsNotEditable).
// Both teams must be non-empty, no player may appear in both, and the
// two teams must differ in size by at most one (ErrInvalidAssignment).
func (m *Match) AssignTeams(teamA, teamB []PlayerID) error {
	if m.status != MatchStatusOpen {
		return fmt.Errorf("%w: cannot edit teams in status %q",
			domainerrors.ErrTeamsNotEditable, m.status)
	}

	if len(teamA) == 0 || len(teamB) == 0 {
		return fmt.Errorf("%w: both teams must be non-empty",
			domainerrors.ErrInvalidAssignment)
	}

	diff := len(teamA) - len(teamB)
	if diff < -1 || diff > 1 {
		return fmt.Errorf("%w: teams must differ in size by at most one (got %d vs %d)",
			domainerrors.ErrInvalidAssignment, len(teamA), len(teamB))
	}

	seen := make(map[PlayerID]struct{}, len(teamA)+len(teamB))
	for _, id := range append(append([]PlayerID{}, teamA...), teamB...) {
		if _, dup := seen[id]; dup {
			return fmt.Errorf("%w: player %q appears more than once",
				domainerrors.ErrInvalidAssignment, id)
		}
		seen[id] = struct{}{}
	}

	m.teamA = clonePlayerIDs(teamA)
	m.teamB = clonePlayerIDs(teamB)
	return nil
}
