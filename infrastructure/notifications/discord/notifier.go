// Package discord implements the Discord webhook notification adapter.
//
// The package exposes a Notifier interface (so subscribers in the same
// package can be tested with a fake) and an HTTPNotifier
// implementation that POSTs to a Discord webhook URL.
//
// Two security-sensitive behaviors live here:
//
//   - The webhook URL is a secret (it embeds a token Discord uses to
//     authenticate the caller). The notifier therefore NEVER includes
//     the URL in returned errors or log lines. A leaked URL gives
//     anyone permission to post to the channel.
//
//   - User-controlled fields are truncated to Discord's documented
//     limits before being marshalled. An over-length field would
//     cause Discord to return 400 Bad Request and lose the
//     notification entirely.
package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Discord webhook field length limits per the official documentation
// (https://discord.com/developers/docs/resources/webhook). Going over
// any of these returns 400.
const (
	maxEmbedTitleLength       = 256
	maxEmbedDescriptionLength = 4096
	maxFieldNameLength        = 256
	maxFieldValueLength       = 1024
)

// ErrNotifierFailed is the only error type the notifier returns to
// callers. It carries an HTTP status (when applicable) but never the
// webhook URL — see the package doc for the rationale.
var ErrNotifierFailed = errors.New("discord notifier: send failed")

// Notifier abstracts the act of sending a Payload to a Discord
// webhook. The interface exists so subscribers can be unit-tested
// against a fake notifier without spinning up an HTTP server.
type Notifier interface {
	Send(ctx context.Context, webhookURL string, payload Payload) error
}

// Payload is the in-process representation of a Discord embed message.
// It is converted to the Discord wire format inside Send.
type Payload struct {
	Title       string
	Description string
	Fields      []Field
}

// Field is one labelled value in an embed. Discord allows up to 25
// fields per embed; the notifier does not enforce that limit because
// SMO never builds payloads with more than a handful of fields.
type Field struct {
	Name   string
	Value  string
	Inline bool
}

// HTTPNotifier sends Payloads via HTTP POST to a Discord webhook URL.
type HTTPNotifier struct {
	client *http.Client
}

// NewHTTPNotifier returns an HTTPNotifier using the given client. The
// caller is responsible for setting a sensible per-request timeout on
// the client; subscribers that call Send must also pass a context with
// a deadline so a slow Discord response cannot stall a use case.
func NewHTTPNotifier(client *http.Client) *HTTPNotifier {
	return &HTTPNotifier{client: client}
}

// Send POSTs the payload to webhookURL. The URL is treated as a
// secret: it never appears in any returned error.
func (n *HTTPNotifier) Send(ctx context.Context, webhookURL string, payload Payload) error {
	body, err := json.Marshal(toWireFormat(payload))
	if err != nil {
		return fmt.Errorf("%w: marshal payload: %w", ErrNotifierFailed, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		// http.NewRequestWithContext can include the URL in its error
		// message when the URL is malformed. We deliberately drop the
		// underlying error and return a generic one to avoid leaking.
		return fmt.Errorf("%w: build request", ErrNotifierFailed)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		// net/http error messages embed the request URL verbatim.
		// Drop the underlying error to prevent the secret from
		// reaching logs or upstream subscribers.
		return fmt.Errorf("%w: transport error", ErrNotifierFailed)
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w: discord returned status %d", ErrNotifierFailed, resp.StatusCode)
	}
	return nil
}

// wireEmbed mirrors the Discord embed JSON shape. Kept private to this
// package: callers build a Payload, never a wireEmbed.
type wireEmbed struct {
	Title       string      `json:"title,omitempty"`
	Description string      `json:"description,omitempty"`
	Fields      []wireField `json:"fields,omitempty"`
}

type wireField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type wirePayload struct {
	Embeds []wireEmbed `json:"embeds"`
}

func toWireFormat(p Payload) wirePayload {
	embed := wireEmbed{
		Title:       truncate(p.Title, maxEmbedTitleLength),
		Description: truncate(p.Description, maxEmbedDescriptionLength),
	}
	if len(p.Fields) > 0 {
		embed.Fields = make([]wireField, len(p.Fields))
		for i, f := range p.Fields {
			embed.Fields[i] = wireField{
				Name:   truncate(f.Name, maxFieldNameLength),
				Value:  truncate(f.Value, maxFieldValueLength),
				Inline: f.Inline,
			}
		}
	}
	return wirePayload{Embeds: []wireEmbed{embed}}
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit]
}
