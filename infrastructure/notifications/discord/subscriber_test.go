package discord_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/events"
	"github.com/ianadou/smo/infrastructure/notifications/discord"
)

const sampleWebhookURL = "https://discord.com/api/webhooks/1234567890/test-token"

// fakeNotifier records calls without making any HTTP request.
type fakeNotifier struct {
	mu        sync.Mutex
	calls     int
	lastURL   string
	lastTitle string
	err       error
}

func (n *fakeNotifier) Send(_ context.Context, webhookURL string, payload discord.Payload) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls++
	n.lastURL = webhookURL
	n.lastTitle = payload.Title
	return n.err
}

type fakeGroupRepo struct {
	groups map[entities.GroupID]*entities.Group
	err    error
}

func (r *fakeGroupRepo) FindByID(_ context.Context, id entities.GroupID) (*entities.Group, error) {
	if r.err != nil {
		return nil, r.err
	}
	g, ok := r.groups[id]
	if !ok {
		return nil, domainerrors.ErrGroupNotFound
	}
	return g, nil
}

func (r *fakeGroupRepo) Save(context.Context, *entities.Group) error { return nil }
func (r *fakeGroupRepo) ListByOrganizer(context.Context, entities.OrganizerID) ([]*entities.Group, error) {
	return nil, nil
}
func (r *fakeGroupRepo) Update(context.Context, *entities.Group) error  { return nil }
func (r *fakeGroupRepo) Delete(context.Context, entities.GroupID) error { return nil }

type fakeMatchRepo struct {
	matches map[entities.MatchID]*entities.Match
	err     error
}

func (r *fakeMatchRepo) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	if r.err != nil {
		return nil, r.err
	}
	m, ok := r.matches[id]
	if !ok {
		return nil, domainerrors.ErrMatchNotFound
	}
	return m, nil
}

func (r *fakeMatchRepo) Save(context.Context, *entities.Match) error { return nil }
func (r *fakeMatchRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Match, error) {
	return nil, nil
}
func (r *fakeMatchRepo) UpdateStatus(context.Context, *entities.Match) error { return nil }
func (r *fakeMatchRepo) Finalize(context.Context, *entities.Match) error     { return nil }
func (r *fakeMatchRepo) Delete(context.Context, entities.MatchID) error      { return nil }

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newGroupWithWebhook(t *testing.T, id entities.GroupID, webhookURL string) *entities.Group {
	t.Helper()
	g, err := entities.NewGroup(id, "Test Group", "org-1", webhookURL, time.Now())
	if err != nil {
		t.Fatalf("NewGroup: %v", err)
	}
	return g
}

func newMatch(t *testing.T, id entities.MatchID, groupID entities.GroupID) *entities.Match {
	t.Helper()
	m, err := entities.RehydrateMatch(entities.MatchSnapshot{
		ID:          id,
		GroupID:     groupID,
		Title:       "Foot du jeudi",
		Venue:       "Stade A",
		ScheduledAt: time.Date(2026, 5, 1, 19, 0, 0, 0, time.UTC),
		Status:      entities.MatchStatusTeamsReady,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("RehydrateMatch: %v", err)
	}
	return m
}

func TestSubscriber_Handle_SendsNotification_WhenWebhookConfigured(t *testing.T) {
	t.Parallel()

	group := newGroupWithWebhook(t, "g-1", sampleWebhookURL)
	match := newMatch(t, "m-1", "g-1")
	notifier := &fakeNotifier{}
	groupRepo := &fakeGroupRepo{groups: map[entities.GroupID]*entities.Group{"g-1": group}}
	matchRepo := &fakeMatchRepo{matches: map[entities.MatchID]*entities.Match{"m-1": match}}
	sub := discord.NewSubscriber(notifier, groupRepo, matchRepo, discardLogger())

	err := sub.Handle(context.Background(), events.MatchTeamsReady{GroupID: "g-1", MatchID: "m-1"})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if notifier.calls != 1 {
		t.Errorf("expected notifier called once, got %d", notifier.calls)
	}
	if notifier.lastURL != sampleWebhookURL {
		t.Errorf("notifier received wrong URL: %q", notifier.lastURL)
	}
	if notifier.lastTitle != "Teams ready — Foot du jeudi" {
		t.Errorf("unexpected title: %q", notifier.lastTitle)
	}
}

func TestSubscriber_Handle_NoOp_WhenGroupHasNoWebhook(t *testing.T) {
	t.Parallel()

	group := newGroupWithWebhook(t, "g-1", "")
	match := newMatch(t, "m-1", "g-1")
	notifier := &fakeNotifier{}
	groupRepo := &fakeGroupRepo{groups: map[entities.GroupID]*entities.Group{"g-1": group}}
	matchRepo := &fakeMatchRepo{matches: map[entities.MatchID]*entities.Match{"m-1": match}}
	sub := discord.NewSubscriber(notifier, groupRepo, matchRepo, discardLogger())

	err := sub.Handle(context.Background(), events.MatchTeamsReady{GroupID: "g-1", MatchID: "m-1"})
	if err != nil {
		t.Errorf("expected nil error for no-webhook group, got %v", err)
	}
	if notifier.calls != 0 {
		t.Errorf("notifier must not be called when webhook is empty, got %d calls", notifier.calls)
	}
}

func TestSubscriber_Handle_IgnoresUnknownEventType(t *testing.T) {
	t.Parallel()

	notifier := &fakeNotifier{}
	sub := discord.NewSubscriber(notifier, &fakeGroupRepo{}, &fakeMatchRepo{}, discardLogger())

	err := sub.Handle(context.Background(), unrelatedEvent{})
	if err != nil {
		t.Errorf("Handle on unrelated event must return nil, got %v", err)
	}
	if notifier.calls != 0 {
		t.Errorf("notifier called for unrelated event: %d calls", notifier.calls)
	}
}

func TestSubscriber_Handle_ReturnsError_WhenGroupRepoFails(t *testing.T) {
	t.Parallel()

	notifier := &fakeNotifier{}
	groupRepo := &fakeGroupRepo{err: errors.New("db down")}
	sub := discord.NewSubscriber(notifier, groupRepo, &fakeMatchRepo{}, discardLogger())

	err := sub.Handle(context.Background(), events.MatchTeamsReady{GroupID: "g-1", MatchID: "m-1"})

	if err == nil {
		t.Fatalf("expected error from group repo failure, got nil")
	}
	if notifier.calls != 0 {
		t.Errorf("notifier must not be called when group lookup fails")
	}
}

func TestSubscriber_Handle_NoOp_WhenMatchDisappeared(t *testing.T) {
	t.Parallel()

	group := newGroupWithWebhook(t, "g-1", sampleWebhookURL)
	notifier := &fakeNotifier{}
	groupRepo := &fakeGroupRepo{groups: map[entities.GroupID]*entities.Group{"g-1": group}}
	matchRepo := &fakeMatchRepo{matches: map[entities.MatchID]*entities.Match{}}
	sub := discord.NewSubscriber(notifier, groupRepo, matchRepo, discardLogger())

	err := sub.Handle(context.Background(), events.MatchTeamsReady{GroupID: "g-1", MatchID: "m-1"})
	if err != nil {
		t.Errorf("expected nil for missing match (data-consistency anomaly is logged not returned), got %v", err)
	}
	if notifier.calls != 0 {
		t.Errorf("notifier must not be called when match is missing")
	}
}

type unrelatedEvent struct{}

func (unrelatedEvent) EventName() string { return "unrelated.event" }
