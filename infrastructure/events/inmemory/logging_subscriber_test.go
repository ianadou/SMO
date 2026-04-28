package inmemory_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/events"
	"github.com/ianadou/smo/infrastructure/events/inmemory"
)

func TestLoggingSubscriber_Handle_LogsMatchTeamsReady_AtInfoWithFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	sub := inmemory.NewLoggingSubscriber(logger)

	ev := events.MatchTeamsReady{
		GroupID:    "group-42",
		MatchID:    "match-7",
		OccurredAt: time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC),
	}

	if err := sub.Handle(context.Background(), ev); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	output := buf.String()
	expectedFragments := []string{
		"level=INFO",
		"event_name=match.teams_ready",
		"group_id=group-42",
		"match_id=match-7",
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(output, fragment) {
			t.Errorf("expected log to contain %q, got %q", fragment, output)
		}
	}
}

func TestLoggingSubscriber_Handle_ReturnsNil_ForUnknownEventType(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	sub := inmemory.NewLoggingSubscriber(logger)

	if err := sub.Handle(context.Background(), unknownEvent{}); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "event_name=unknown.event") {
		t.Errorf("expected fallback log with event_name field, got %q", buf.String())
	}
}

type unknownEvent struct{}

func (unknownEvent) EventName() string { return "unknown.event" }
