package inmemory

import (
	"context"
	"log/slog"

	"github.com/ianadou/smo/domain/events"
)

// LoggingSubscriber writes one INFO log line per event it receives.
// It serves as a permanent audit trail of every domain event emitted
// by the system: useful for production debugging, compliance, and as
// a smoke test that the publisher is wired correctly.
//
// The logger is injected (not slog.Default()) so tests can capture
// output without monkey-patching the global logger. See ADR 0004.
type LoggingSubscriber struct {
	logger *slog.Logger
}

// NewLoggingSubscriber returns a subscriber that logs every received
// event to the given logger.
func NewLoggingSubscriber(logger *slog.Logger) *LoggingSubscriber {
	return &LoggingSubscriber{logger: logger}
}

// Handle logs the event at INFO level with structured fields. The
// returned error is always nil: a logging failure should not block
// other subscribers, and slog itself does not surface I/O errors
// through its API.
func (s *LoggingSubscriber) Handle(ctx context.Context, event events.Event) error {
	switch ev := event.(type) {
	case events.MatchTeamsReady:
		s.logger.InfoContext(ctx, "domain event",
			"event_name", ev.EventName(),
			"group_id", string(ev.GroupID),
			"match_id", string(ev.MatchID),
		)
	default:
		s.logger.InfoContext(ctx, "domain event", "event_name", event.EventName())
	}
	return nil
}
