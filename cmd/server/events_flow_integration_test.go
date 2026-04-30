//go:build integration

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	matchusecase "github.com/ianadou/smo/application/usecases/match"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/events"
	"github.com/ianadou/smo/infrastructure/clock"
	"github.com/ianadou/smo/infrastructure/events/inmemory"
	"github.com/ianadou/smo/infrastructure/notifications/discord"
	persistinmemory "github.com/ianadou/smo/infrastructure/persistence/inmemory"
)

// TestEventsFlow_E2E_PostsToDiscord_OnMarkTeamsReady is the end-to-end
// guard for the Domain Events foundation (PRs #54, #55, ADR 0004).
//
// It wires the production pieces — inmemory.Publisher, the real
// discord.Subscriber, the real discord.HTTPNotifier — and then drives
// a full MarkTeamsReadyUseCase.Execute call. A bug in any of:
//
//   - the EventName constant,
//   - the publisher's Subscribe key,
//   - the subscriber's event-type assertion,
//   - the use case's Publish call,
//   - the notifier's HTTP request shape,
//
// shows up here as a missing or malformed POST on the fake Discord
// server. None of these are covered by the per-unit tests.
//
// The fake Discord server uses TLS because validateWebhookURL in the
// Group entity rejects non-HTTPS URLs. The notifier's http.Client
// skips verification because the test certificate is self-signed —
// InsecureSkipVerify is acceptable HERE because the target address
// is the test process itself (httptest.NewTLSServer).
//
// Tracked in issue #56.
func TestEventsFlow_E2E_PostsToDiscord_OnMarkTeamsReady(t *testing.T) {
	var (
		mu   sync.Mutex
		got  []byte
		hits atomic.Int32
	)
	discordSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		got = body
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer discordSrv.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	publisher := inmemory.NewPublisher(logger)

	groupRepo := persistinmemory.NewGroupRepository()
	matchRepo := newInmemoryMatchRepoForEventsTest()

	group := mustNewGroup(t, "g-1", discordSrv.URL)
	_ = groupRepo.Save(context.Background(), group)

	match := mustRehydrateOpenMatch(t, "m-1", "g-1")
	_ = matchRepo.Save(context.Background(), match)

	// http.Client that trusts the httptest TLS cert. ONLY safe in
	// tests — a real notifier never skips verification.
	tlsClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test target only
		},
	}
	notifier := discord.NewHTTPNotifier(tlsClient)
	subscriber := discord.NewSubscriber(notifier, groupRepo, matchRepo, logger)
	publisher.Subscribe(events.MatchTeamsReadyEventName, subscriber)

	useCase := matchusecase.NewMarkTeamsReadyUseCase(matchRepo, publisher, clock.New())

	if _, err := useCase.Execute(context.Background(), "m-1"); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if hits.Load() != 1 {
		t.Fatalf("expected 1 POST to Discord, got %d", hits.Load())
	}

	mu.Lock()
	body := append([]byte(nil), got...)
	mu.Unlock()

	var parsed struct {
		Embeds []struct {
			Title string `json:"title"`
		} `json:"embeds"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("Discord payload is not JSON: %v. body=%s", err, string(body))
	}
	if len(parsed.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d. body=%s", len(parsed.Embeds), string(body))
	}
	if parsed.Embeds[0].Title == "" {
		t.Errorf("expected embed title to be set, got empty. body=%s", string(body))
	}
}

func mustNewGroup(t *testing.T, id entities.GroupID, webhookURL string) *entities.Group {
	t.Helper()
	g, err := entities.RehydrateGroup(entities.GroupSnapshot{
		ID:          id,
		Name:        "Events Test Group",
		OrganizerID: "org-1",
		WebhookURL:  webhookURL,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("RehydrateGroup: %v", err)
	}
	return g
}

func mustRehydrateOpenMatch(t *testing.T, id entities.MatchID, groupID entities.GroupID) *entities.Match {
	t.Helper()
	m, err := entities.RehydrateMatch(entities.MatchSnapshot{
		ID:          id,
		GroupID:     groupID,
		Title:       "Events E2E Match",
		Venue:       "Stadium",
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      entities.MatchStatusOpen,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("RehydrateMatch: %v", err)
	}
	return m
}

// inmemoryMatchRepoForEventsTest is a minimal in-memory match repo
// that satisfies the small subset of ports.MatchRepository the test
// actually exercises (FindByID, UpdateStatus, Save). The domain
// package does not ship a real in-memory match adapter, and pulling
// in the postgres repo would force a Postgres testcontainer for a
// test that does not need persistence.
type inmemoryMatchRepoForEventsTest struct {
	mu      sync.Mutex
	matches map[entities.MatchID]*entities.Match
}

func newInmemoryMatchRepoForEventsTest() *inmemoryMatchRepoForEventsTest {
	return &inmemoryMatchRepoForEventsTest{matches: make(map[entities.MatchID]*entities.Match)}
}

func (r *inmemoryMatchRepoForEventsTest) Save(_ context.Context, m *entities.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matches[m.ID()] = m
	return nil
}

func (r *inmemoryMatchRepoForEventsTest) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.matches[id]
	if !ok {
		return nil, domainerrors.ErrMatchNotFound
	}
	return m, nil
}

func (r *inmemoryMatchRepoForEventsTest) ListByGroup(context.Context, entities.GroupID) ([]*entities.Match, error) {
	return nil, nil
}

func (r *inmemoryMatchRepoForEventsTest) UpdateStatus(_ context.Context, m *entities.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matches[m.ID()] = m
	return nil
}

func (r *inmemoryMatchRepoForEventsTest) Finalize(context.Context, *entities.Match) error {
	return nil
}

func (r *inmemoryMatchRepoForEventsTest) Delete(context.Context, entities.MatchID) error {
	return nil
}
