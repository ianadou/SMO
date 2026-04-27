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

	// ErrInvalidParameter is returned when a configuration parameter is
	// outside its accepted range (e.g., a learning rate outside (0, 1]).
	ErrInvalidParameter = errors.New("invalid parameter")

	// ErrInvalidEmail is returned when an email string fails the format
	// check at organizer creation.
	ErrInvalidEmail = errors.New("invalid email")

	// ErrInvalidPassword is returned when a password fails the policy
	// (currently: minimum 12 characters). Distinct from ErrInvalidCredentials
	// so the registration endpoint can return a useful 400 while login keeps
	// a uniform 401.
	ErrInvalidPassword = errors.New("invalid password")
)

// Business rule errors — returned when a domain operation is not allowed
// in the current state of the entity.
var (
	// ErrInvalidTransition is returned when a state machine transition is
	// not allowed from the current state.
	ErrInvalidTransition = errors.New("invalid transition")

	// ErrInvalidAssignment is returned when an assignment strategy cannot
	// produce a valid team distribution from the given input.
	ErrInvalidAssignment = errors.New("invalid assignment")

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

	// ErrMatchNotCompleted is returned when an operation that requires
	// the match to be in the completed state (e.g., casting a vote) is
	// attempted before the match has been marked completed.
	ErrMatchNotCompleted = errors.New("match is not completed")
)

// Repository errors — returned when a persistence operation fails for a
// reason that has a meaningful business interpretation (not a generic
// "database is down" failure).
var (
	// ErrGroupNotFound is returned when a group lookup by ID has no match.
	ErrGroupNotFound = errors.New("group not found")

	// ErrMatchNotFound is returned when a match lookup by ID has no match.
	ErrMatchNotFound = errors.New("match not found")

	// ErrPlayerNotFound is returned when a player lookup by ID has no match.
	ErrPlayerNotFound = errors.New("player not found")

	// ErrInvitationNotFound is returned when an invitation lookup by ID
	// or token hash has no match.
	ErrInvitationNotFound = errors.New("invitation not found")

	// ErrInvitationExpired is returned when attempting to accept an
	// invitation whose expires_at is in the past.
	ErrInvitationExpired = errors.New("invitation expired")

	// ErrInvitationAlreadyUsed is returned when attempting to accept an
	// invitation that has already been consumed.
	ErrInvitationAlreadyUsed = errors.New("invitation already used")

	// ErrVoteNotFound is returned when a vote lookup by ID has no match.
	ErrVoteNotFound = errors.New("vote not found")

	// ErrAlreadyVoted is returned when a voter attempts to cast a second
	// vote for the same (match, voted_player) pair. Triggered by the
	// unique constraint (match_id, voter_id, voted_id).
	ErrAlreadyVoted = errors.New("already voted for this player in this match")

	// ErrReferencedEntityNotFound is returned when a persistence operation
	// fails because it references another entity that does not exist
	// (e.g., creating a group with an organizer_id that is not in the
	// organizers table). This corresponds to a foreign-key violation at
	// the SQL level, but the concept is storage-agnostic: any backend
	// that enforces referential integrity can produce this error.
	ErrReferencedEntityNotFound = errors.New("referenced entity not found")

	// ErrEmailAlreadyExists is returned when registration would create
	// an organizer with an email already taken. Triggered by the UNIQUE
	// constraint on organizers.email.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrOrganizerNotFound is returned when an organizer lookup by ID
	// or email has no match.
	ErrOrganizerNotFound = errors.New("organizer not found")

	// ErrInvalidCredentials is returned by the login flow when the
	// email/password combination does not match. Deliberately distinct
	// from ErrOrganizerNotFound and ErrInvalidPassword: the login error
	// must not reveal whether the email exists, to prevent enumeration.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidToken is returned when a JWT cannot be parsed, has an
	// invalid signature, is expired, or carries unexpected claims.
	ErrInvalidToken = errors.New("invalid token")
)
