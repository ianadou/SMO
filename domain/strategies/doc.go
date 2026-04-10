// Package strategies provides team assignment algorithms for matches.
//
// All strategies implement the AssignmentStrategy interface, which takes
// a list of players and returns two slices of PlayerID representing the
// two teams. Strategies are pure algorithms with no infrastructure
// dependencies; they can be tested in isolation with deterministic inputs.
//
// Available strategies:
//
//   - RandomAssignmentStrategy: shuffles players and splits them in half.
//   - RankingBasedStrategy: snake draft based on player ranking.
//   - ManualAssignmentStrategy: organizer-defined explicit composition.
package strategies
