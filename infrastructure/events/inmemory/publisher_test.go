package inmemory_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/events"
	"github.com/ianadou/smo/infrastructure/events/inmemory"
)

// fakeOtherEvent is a stand-in for a future, unrelated event. It lets
// tests assert that a subscriber registered for MatchTeamsReady does
// NOT receive an event with a different EventName, even though SMO
// only ships one production event today.
type fakeOtherEvent struct{}

func (fakeOtherEvent) EventName() string { return "fake.other" }

// recordingSubscriber implements ports.EventSubscriber by counting
// invocations. Optionally returns a fixed error.
type recordingSubscriber struct {
	calls atomic.Int32
	err   error
}

func (s *recordingSubscriber) Handle(_ context.Context, _ events.Event) error {
	s.calls.Add(1)
	return s.err
}

func newPublisher() *inmemory.Publisher {
	return inmemory.NewPublisher(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func sampleMatchTeamsReady() events.MatchTeamsReady {
	return events.MatchTeamsReady{GroupID: "g-1", MatchID: "m-1", OccurredAt: time.Now()}
}

func TestPublisher_Publish_NoSubscriber_IsNoop(t *testing.T) {
	t.Parallel()
	pub := newPublisher()

	// Must not panic when nothing is subscribed.
	pub.Publish(context.Background(), sampleMatchTeamsReady())
}

func TestPublisher_Publish_SingleSubscriber_ReceivesEvent(t *testing.T) {
	t.Parallel()
	pub := newPublisher()
	sub := &recordingSubscriber{}
	pub.Subscribe(events.MatchTeamsReadyEventName, sub)

	pub.Publish(context.Background(), sampleMatchTeamsReady())

	if got := sub.calls.Load(); got != 1 {
		t.Errorf("expected subscriber called once, got %d", got)
	}
}

func TestPublisher_Publish_MultipleSubscribers_AllReceiveEvent(t *testing.T) {
	t.Parallel()
	pub := newPublisher()
	subA := &recordingSubscriber{}
	subB := &recordingSubscriber{}
	subC := &recordingSubscriber{}
	pub.Subscribe(events.MatchTeamsReadyEventName, subA)
	pub.Subscribe(events.MatchTeamsReadyEventName, subB)
	pub.Subscribe(events.MatchTeamsReadyEventName, subC)

	pub.Publish(context.Background(), sampleMatchTeamsReady())

	for name, sub := range map[string]*recordingSubscriber{"A": subA, "B": subB, "C": subC} {
		if got := sub.calls.Load(); got != 1 {
			t.Errorf("subscriber %s: expected 1 call, got %d", name, got)
		}
	}
}

func TestPublisher_Publish_SubscriberError_DoesNotAbortDispatch(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	pub := inmemory.NewPublisher(logger)

	failing := &recordingSubscriber{err: errors.New("boom")}
	healthy := &recordingSubscriber{}
	pub.Subscribe(events.MatchTeamsReadyEventName, failing)
	pub.Subscribe(events.MatchTeamsReadyEventName, healthy)

	pub.Publish(context.Background(), sampleMatchTeamsReady())

	if got := failing.calls.Load(); got != 1 {
		t.Errorf("failing subscriber: expected 1 call, got %d", got)
	}
	if got := healthy.calls.Load(); got != 1 {
		t.Errorf("healthy subscriber: expected 1 call after a peer failed, got %d", got)
	}
	if !strings.Contains(buf.String(), "event subscriber returned error") {
		t.Errorf("expected WARN log for subscriber error, got %q", buf.String())
	}
}

func TestPublisher_Publish_DispatchesByEventName_Only(t *testing.T) {
	t.Parallel()
	pub := newPublisher()

	teamsReadySub := &recordingSubscriber{}
	otherSub := &recordingSubscriber{}
	pub.Subscribe(events.MatchTeamsReadyEventName, teamsReadySub)
	pub.Subscribe("fake.other", otherSub)

	pub.Publish(context.Background(), fakeOtherEvent{})

	if got := teamsReadySub.calls.Load(); got != 0 {
		t.Errorf("subscriber for match.teams_ready must not receive fake.other, got %d calls", got)
	}
	if got := otherSub.calls.Load(); got != 1 {
		t.Errorf("subscriber for fake.other expected 1 call, got %d", got)
	}
}
