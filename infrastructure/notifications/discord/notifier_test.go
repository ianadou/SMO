package discord_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ianadou/smo/infrastructure/notifications/discord"
)

// fakeWebhook is a minimal stand-in for the Discord webhook endpoint.
// It records the payload it received so tests can assert on it.
type fakeWebhook struct {
	server      *httptest.Server
	statusCode  int
	lastBody    []byte
	receivedReq bool
}

func newFakeWebhook(t *testing.T, statusCode int) *fakeWebhook {
	t.Helper()
	w := &fakeWebhook{statusCode: statusCode}
	w.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		w.receivedReq = true
		body, _ := io.ReadAll(r.Body)
		w.lastBody = body
		rw.WriteHeader(w.statusCode)
	}))
	t.Cleanup(w.server.Close)
	return w
}

func newNotifier() *discord.HTTPNotifier {
	return discord.NewHTTPNotifier(&http.Client{Timeout: 5 * time.Second})
}

func TestHTTPNotifier_Send_PostsPayloadAsJSON(t *testing.T) {
	t.Parallel()
	hook := newFakeWebhook(t, http.StatusNoContent)
	notifier := newNotifier()

	err := notifier.Send(context.Background(), hook.server.URL, discord.Payload{
		Title:       "Teams ready",
		Description: "Foot du jeudi",
		Fields:      []discord.Field{{Name: "Venue", Value: "Stade A", Inline: true}},
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if !hook.receivedReq {
		t.Fatalf("webhook never received the request")
	}

	var got struct {
		Embeds []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Fields      []struct {
				Name   string `json:"name"`
				Value  string `json:"value"`
				Inline bool   `json:"inline"`
			} `json:"fields"`
		} `json:"embeds"`
	}
	if err := json.Unmarshal(hook.lastBody, &got); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if len(got.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(got.Embeds))
	}
	if got.Embeds[0].Title != "Teams ready" {
		t.Errorf("title mismatch: got %q", got.Embeds[0].Title)
	}
	if got.Embeds[0].Description != "Foot du jeudi" {
		t.Errorf("description mismatch: got %q", got.Embeds[0].Description)
	}
	if len(got.Embeds[0].Fields) != 1 || got.Embeds[0].Fields[0].Name != "Venue" {
		t.Errorf("fields mismatch: got %+v", got.Embeds[0].Fields)
	}
}

func TestHTTPNotifier_Send_TruncatesOverlengthTitle(t *testing.T) {
	t.Parallel()
	hook := newFakeWebhook(t, http.StatusNoContent)
	notifier := newNotifier()

	tooLong := strings.Repeat("a", 300)
	err := notifier.Send(context.Background(), hook.server.URL, discord.Payload{Title: tooLong})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	var got struct {
		Embeds []struct {
			Title string `json:"title"`
		} `json:"embeds"`
	}
	_ = json.Unmarshal(hook.lastBody, &got)
	if length := len(got.Embeds[0].Title); length != 256 {
		t.Errorf("expected title truncated to 256, got %d chars", length)
	}
}

func TestHTTPNotifier_Send_TruncatesOverlengthFieldValue(t *testing.T) {
	t.Parallel()
	hook := newFakeWebhook(t, http.StatusNoContent)
	notifier := newNotifier()

	tooLong := strings.Repeat("x", 2000)
	err := notifier.Send(context.Background(), hook.server.URL, discord.Payload{
		Fields: []discord.Field{{Name: "Long", Value: tooLong}},
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	var got struct {
		Embeds []struct {
			Fields []struct {
				Value string `json:"value"`
			} `json:"fields"`
		} `json:"embeds"`
	}
	_ = json.Unmarshal(hook.lastBody, &got)
	if length := len(got.Embeds[0].Fields[0].Value); length != 1024 {
		t.Errorf("expected field value truncated to 1024, got %d chars", length)
	}
}

func TestHTTPNotifier_Send_ReturnsErrNotifierFailed_OnNon2xx(t *testing.T) {
	t.Parallel()
	hook := newFakeWebhook(t, http.StatusBadRequest)
	notifier := newNotifier()

	err := notifier.Send(context.Background(), hook.server.URL, discord.Payload{Title: "x"})

	if !errors.Is(err, discord.ErrNotifierFailed) {
		t.Errorf("expected ErrNotifierFailed, got %v", err)
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to mention status 400, got %q", err.Error())
	}
}

// TestHTTPNotifier_Send_NeverLeaksWebhookURL_InTransportError is the
// key security regression guard: if Discord (or anything else) is
// unreachable, the underlying net/http error embeds the request URL
// verbatim. The notifier must drop it so it never reaches logs or
// upstream subscribers.
func TestHTTPNotifier_Send_NeverLeaksWebhookURL_InTransportError(t *testing.T) {
	t.Parallel()
	notifier := discord.NewHTTPNotifier(&http.Client{Timeout: 100 * time.Millisecond})

	const secretToken = "S3CR3T-TOKEN-must-not-leak"
	unreachable := "http://127.0.0.1:1/api/webhooks/12345/" + secretToken

	err := notifier.Send(context.Background(), unreachable, discord.Payload{Title: "x"})

	if err == nil {
		t.Fatalf("expected error from unreachable URL, got nil")
	}
	if !errors.Is(err, discord.ErrNotifierFailed) {
		t.Errorf("expected ErrNotifierFailed, got %v", err)
	}
	if strings.Contains(err.Error(), secretToken) {
		t.Fatalf("SECURITY: error message leaks webhook token: %q", err.Error())
	}
	if strings.Contains(err.Error(), unreachable) {
		t.Fatalf("SECURITY: error message leaks full webhook URL: %q", err.Error())
	}
}

func TestHTTPNotifier_Send_NeverLeaksWebhookURL_OnMalformedURL(t *testing.T) {
	t.Parallel()
	notifier := newNotifier()

	const secretToken = "OTHER-S3CR3T-token"
	// Newline in URL forces NewRequestWithContext to fail; that error
	// path also embeds the URL by default.
	malformed := "http://discord.com/\n" + secretToken

	err := notifier.Send(context.Background(), malformed, discord.Payload{Title: "x"})

	if err == nil {
		t.Fatalf("expected error from malformed URL, got nil")
	}
	if strings.Contains(err.Error(), secretToken) {
		t.Fatalf("SECURITY: error message leaks token from malformed URL: %q", err.Error())
	}
}
