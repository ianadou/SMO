package entities

import (
	"fmt"
	"slices"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

// TeamLockLeadTime is how long before kickoff the team composition
// freezes. Inside this window the rosters can no longer be edited, so
// players can rely on a stable line-up as the match begins.
const TeamLockLeadTime = 10 * time.Minute

// AssignTeams sets the two-team composition of the match.
//
// Teams may be (re)assigned while the match is open or teams_ready, up
// until TeamLockLeadTime before the scheduled kickoff; from in_progress
// onward, or once that lead time is reached, the rosters are frozen
// (ErrTeamsNotEditable / ErrTeamsLocked). Both teams must be non-empty,
// no player may appear in both, and the two teams must differ in size by
// at most one (ErrInvalidAssignment).
//
// now is injected rather than read from a clock so the rule stays pure
// and deterministic under test.
func (m *Match) AssignTeams(teamA, teamB []PlayerID, now time.Time) error {
	if m.status != MatchStatusOpen && m.status != MatchStatusTeamsReady {
		return fmt.Errorf("%w: cannot edit teams in status %q",
			domainerrors.ErrTeamsNotEditable, m.status)
	}

	if !now.Before(m.scheduledAt.Add(-TeamLockLeadTime)) {
		return fmt.Errorf("%w: teams freeze %s before kickoff (scheduled for %s)",
			domainerrors.ErrTeamsLocked, TeamLockLeadTime, m.scheduledAt.Format(time.RFC3339))
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

// TeamOf reports which side the given player is assigned to. The second
// return value is false when the player is on neither roster (or teams
// have not been assigned yet).
func (m *Match) TeamOf(playerID PlayerID) (TeamSide, bool) {
	if slices.Contains(m.teamA, playerID) {
		return TeamSideA, true
	}
	if slices.Contains(m.teamB, playerID) {
		return TeamSideB, true
	}
	return "", false
}

// TeammatesOf returns the other members of the given player's team,
// preserving roster order. It returns nil when the player is on neither
// roster — callers that need to distinguish "no teammates" from "not in
// the match" should use TeamOf.
func (m *Match) TeammatesOf(playerID PlayerID) []PlayerID {
	side, ok := m.TeamOf(playerID)
	if !ok {
		return nil
	}

	roster := m.teamA
	if side == TeamSideB {
		roster = m.teamB
	}

	teammates := make([]PlayerID, 0, len(roster)-1)
	for _, id := range roster {
		if id != playerID {
			teammates = append(teammates, id)
		}
	}
	return teammates
}
