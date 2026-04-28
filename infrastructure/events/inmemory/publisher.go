// Package inmemory provides the in-process synchronous implementation
// of the domain.ports.EventPublisher port. It is the default and only
// publisher for SMO; introducing an async or out-of-process variant
// would warrant a dedicated ADR.
package inmemory

import (
	"context"
	"log/slog"
	"sync"

	"github.com/ianadou/smo/domain/events"
	"github.com/ianadou/smo/domain/ports"
)

// Publisher dispatches domain events to subscribers registered by
// event name. Dispatch is synchronous: Publish returns only after
// every matching subscriber has been invoked.
//
// A subscriber that returns an error is logged at WARN and the
// dispatch continues with the next subscriber. A slow subscriber
// stalls the caller — subscribers must therefore be cheap and
// non-blocking. See ADR 0004.
type Publisher struct {
	logger      *slog.Logger
	mu          sync.RWMutex
	subscribers map[string][]ports.EventSubscriber
}

// NewPublisher returns a Publisher ready to accept Subscribe and
// Publish calls. The logger is used to record subscriber errors and
// must not be nil.
func NewPublisher(logger *slog.Logger) *Publisher {
	return &Publisher{
		logger:      logger,
		subscribers: make(map[string][]ports.EventSubscriber),
	}
}

// Subscribe registers sub to receive every event whose EventName
// matches eventName. A subscriber may register for multiple events by
// calling Subscribe several times.
func (p *Publisher) Subscribe(eventName string, sub ports.EventSubscriber) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers[eventName] = append(p.subscribers[eventName], sub)
}

// Publish dispatches event to every subscriber registered for
// event.EventName(). It is safe to call concurrently with Subscribe.
func (p *Publisher) Publish(ctx context.Context, event events.Event) {
	p.mu.RLock()
	subs := p.subscribers[event.EventName()]
	p.mu.RUnlock()

	for _, sub := range subs {
		if err := sub.Handle(ctx, event); err != nil {
			p.logger.WarnContext(ctx, "event subscriber returned error",
				"event_name", event.EventName(),
				"error", err,
			)
		}
	}
}
