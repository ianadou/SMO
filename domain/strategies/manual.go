package strategies

import (
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// ManualAssignmentStrategy returns a pre-defined team composition supplied
// by the organizer at construction time.
//
// This strategy is used when the organizer wants full control over team
// composition, typically as an override after a previous assignment was
// produced by another strategy.
//
// The strategy validates at construction time that:
//   - Both teams are non-empty.
//   - No player ID appears in both teams.
//
// The strategy validates at Assign time that:
//   - Every player in the input is exactly in one of the two teams.
//   - No team contains a player ID that is not in the input.
type ManualAssignmentStrategy struct {
	teamA []entities.PlayerID
	teamB []entities.PlayerID
}

// NewManualAssignmentStrategy builds a ManualAssignmentStrategy with the
// given pre-defined team compositions.
//
// Returns ErrInvalidAssignment if either team is empty or if a player ID
// appears in both teams.
func NewManualAssignmentStrategy(teamA, teamB []entities.PlayerID) (*ManualAssignmentStrategy, error) {
	if len(teamA) == 0 || len(teamB) == 0 {
		return nil, fmt.Errorf("%w: both teams must contain at least one player", domainerrors.ErrInvalidAssignment)
	}

	// Build a set of team A members to detect duplicates in team B.
	inTeamA := make(map[entities.PlayerID]struct{}, len(teamA))
	for _, id := range teamA {
		inTeamA[id] = struct{}{}
	}
	for _, id := range teamB {
		if _, exists := inTeamA[id]; exists {
			return nil, fmt.Errorf("%w: player %q appears in both teams", domainerrors.ErrInvalidAssignment, id)
		}
	}

	// Defensive copies so the strategy is independent of caller mutations.
	teamACopy := make([]entities.PlayerID, len(teamA))
	copy(teamACopy, teamA)
	teamBCopy := make([]entities.PlayerID, len(teamB))
	copy(teamBCopy, teamB)

	return &ManualAssignmentStrategy{
		teamA: teamACopy,
		teamB: teamBCopy,
	}, nil
}

// Assign returns the pre-defined teams after verifying that they exactly
// match the input players (no missing player, no extra player).
func (s *ManualAssignmentStrategy) Assign(players []*entities.Player) ([]entities.PlayerID, []entities.PlayerID, error) {
	expectedSet := make(map[entities.PlayerID]struct{}, len(players))
	for _, player := range players {
		expectedSet[player.ID()] = struct{}{}
	}

	assignedSet := make(map[entities.PlayerID]struct{}, len(s.teamA)+len(s.teamB))
	for _, id := range s.teamA {
		assignedSet[id] = struct{}{}
	}
	for _, id := range s.teamB {
		assignedSet[id] = struct{}{}
	}

	if len(expectedSet) != len(assignedSet) {
		return nil, nil, fmt.Errorf("%w: input has %d players but assignment covers %d", domainerrors.ErrInvalidAssignment, len(expectedSet), len(assignedSet))
	}

	for id := range expectedSet {
		if _, ok := assignedSet[id]; !ok {
			return nil, nil, fmt.Errorf("%w: player %q is in the input but not in the assignment", domainerrors.ErrInvalidAssignment, id)
		}
	}

	// Return defensive copies so the caller cannot mutate the strategy state.
	teamACopy := make([]entities.PlayerID, len(s.teamA))
	copy(teamACopy, s.teamA)
	teamBCopy := make([]entities.PlayerID, len(s.teamB))
	copy(teamBCopy, s.teamB)

	return teamACopy, teamBCopy, nil
}
