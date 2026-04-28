package ports

import (
	"context"

	"github.com/ianadou/smo/domain/events"
)

// EventPublisher delivers a domain event to every subscriber registered
// for that event's name. Implementations are expected to be in-process
// and synchronous: a Publish call returns once all subscribers have
// been invoked.
//
// Use cases depend on this port to emit events without knowing which
// subscribers exist. Adapters such as Discord notifications or
// Prometheus counters register themselves at wiring time in
// cmd/server/main.go.
//
// See docs/adr/0004-domain-events-pattern.md for the rationale.
type EventPublisher interface {
	// Publish dispatches the event to every subscriber that registered
	// for events.Event.EventName(). A subscriber error must not abort
	// the dispatch of the same event to other subscribers — see ADR.
	Publish(ctx context.Context, event events.Event)
}

// EventSubscriber reacts to a domain event. The Handle method receives
// the event by value; concrete subscribers type-assert to the specific
// event struct they registered for.
type EventSubscriber interface {
	// Handle processes the event. A returned error is logged by the
	// publisher but never propagated: a subscriber that fails must not
	// take down the calling use case.
	Handle(ctx context.Context, event events.Event) error
}
