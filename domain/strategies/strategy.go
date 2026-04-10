package strategies

import "github.com/ianadou/smo/domain/entities"

// AssignmentStrategy is the contract that all team assignment algorithms
// must implement. Implementations take a list of players and produce two
// teams as ordered slices of PlayerID.
//
// The strategy itself does not generate Team or Match identifiers; it
// only decides which players go into which team. The caller (typically
// a use case) is responsible for building the Team entities around the
// returned PlayerID slices.
//
// Implementations must be deterministic when given the same inputs and
// the same internal state, so that they are testable. Strategies that
// rely on randomness (e.g., RandomAssignmentStrategy) must accept their
// random source as a constructor parameter rather than reading from a
// global one.
type AssignmentStrategy interface {
	// Assign distributes the given players into two teams.
	//
	// Returns ErrInvalidAssignment wrapped with context if the input
	// cannot produce a valid distribution (e.g., fewer than 2 players,
	// duplicate IDs, or strategy-specific constraints not satisfied).
	Assign(players []*entities.Player) (teamA, teamB []entities.PlayerID, err error)
}
