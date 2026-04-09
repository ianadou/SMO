// Package errors defines the sentinel errors of the SMO domain.
//
// All domain operations that can fail return one of these errors so that
// callers can distinguish between business rule violations and technical
// failures using errors.Is.
package errors

import "errors"

// Validation errors — returned when input data violates a domain invariant.
var (
	// ErrInvalidID is returned when an entity ID is empty or malformed.
	ErrInvalidID = errors.New("invalid id")

	// ErrInvalidName is returned when a name field is empty or too long.
	ErrInvalidName = errors.New("invalid name")

	// ErrInvalidScore is returned when a vote score is outside [1, 5].
	ErrInvalidScore = errors.New("invalid score")

	// ErrInvalidDate is returned when a date is zero or in the past
	// where a future date is required.
	ErrInvalidDate = errors.New("invalid date")

	// ErrInvalidStatus is returned when a match status string does not
	// match any known status value.
	ErrInvalidStatus = errors.New("invalid status")
)

// Business rule errors — returned when a domain operation is not allowed
// in the current state of the entity.
var (
	// ErrMatchFull is returned when trying to add a player to a match
	// that already has the maximum number of participants.
	ErrMatchFull = errors.New("match is full")

	// ErrTeamFull is returned when trying to add a player to a team
	// that already has the maximum number of participants.
	ErrTeamFull = errors.New("team is full")

	// ErrPlayerNotInMatch is returned when an operation references a player
	// that is not part of the given match.
	ErrPlayerNotInMatch = errors.New("player is not in match")

	// ErrSelfVote is returned when a player tries to vote for themselves.
	ErrSelfVote = errors.New("cannot vote for yourself")
)
