package entities

import "time"

// MatchParticipant is a value-object projection used by the
// ListMatchParticipants use case. It is NOT an aggregate: it is the
// flattened result of "for this match, give me each confirmed
// invitation's player together with the timestamp of confirmation".
//
// The struct lives in domain/entities so the InvitationRepository port
// can return it without leaking infrastructure types. There is no
// constructor: the projection is built by the persistence layer from
// a JOIN, never by domain code on its own.
type MatchParticipant struct {
	PlayerID    PlayerID
	PlayerName  string
	ConfirmedAt time.Time
}
