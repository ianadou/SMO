// Package ranking provides the algorithm that updates player rankings
// based on the votes received during a match.
//
// The Calculator takes a list of players and a list of votes and returns
// the updated rankings as a map. It does not mutate the input players,
// keeping the domain entities immutable.
//
// The current algorithm is a weighted average: each player's new ranking
// is a blend of their current ranking and the average of the votes they
// received in the current match. The blending strength is controlled by
// the learning rate (alpha) supplied at construction time.
package ranking
