// Package vote contains the use cases for the Vote aggregate:
// casting peer ratings post-match, retrieving individual votes,
// and listing votes by match.
//
// CastVote enforces two cross-aggregate rules:
//  1. the target match must be in the Completed state
//  2. a voter can vote at most once per (match, voted_player) pair
//
// The first rule is validated by loading the match; the second is
// enforced by the UNIQUE constraint on the votes table.
package vote
