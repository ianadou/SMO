// Package player contains the use cases that operate on the Player
// aggregate: creating a player, retrieving one, listing players by
// group, and updating a player's ranking.
//
// New players always start with the default ranking (1000). The
// ranking update use case exists so that match results can adjust
// player rankings after a completed match.
package player
