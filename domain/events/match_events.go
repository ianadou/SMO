// Package events defines the domain events emitted when significant
// state transitions happen in the business core.
//
// Events are pure value types: they carry the minimal information a
// subscriber needs to react, and nothing about the transport or the
// persistence layer. Subscribers (Discord notification, Prometheus
// counters, audit logs, future account lockout, etc.) live in the
// infrastructure layer and plug into the domain via the Publisher port.
//
// See docs/adr/0004-domain-events-pattern.md for the architectural
// rationale.
package events

import (
	"time"

	"github.com/ianadou/smo/domain/entities"
)

// Event is the marker interface every domain event implements.
//
// The single method EventName returns a stable string identifier used
// by the publisher to dispatch the event to its subscribers. The
// rationale for dispatching by name rather than by Go type is captured
// in ADR 0004.
type Event interface {
	EventName() string
}

// MatchTeamsReadyEventName is the stable identifier under which
// subscribers register for the MatchTeamsReady event. Renaming this
// constant is a breaking change — see ADR 0004.
const MatchTeamsReadyEventName = "match.teams_ready"

// MatchTeamsReady is emitted when a Match transitions to the
// teams_ready status, i.e. the organizer has finished assembling the
// teams and the match is ready to start. Subscribers may use this
// signal to notify players (Discord), increment a counter (Prometheus),
// or write an audit entry.
type MatchTeamsReady struct {
	GroupID    entities.GroupID
	MatchID    entities.MatchID
	OccurredAt time.Time
}

// EventName implements Event.
func (MatchTeamsReady) EventName() string { return MatchTeamsReadyEventName }
