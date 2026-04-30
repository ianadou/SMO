package entities

import (
	"errors"
	"strings"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

const validDiscordWebhook = "https://discord.com/api/webhooks/1234567890/abcdef-XYZ"

func TestNewGroup_ReturnsGroup_WhenInputsAreValid(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	group, err := NewGroup("group-123", "Foot du jeudi", "org-456", "", createdAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if group.ID() != "group-123" {
		t.Errorf("expected ID 'group-123', got %q", group.ID())
	}
	if group.Name() != "Foot du jeudi" {
		t.Errorf("expected name 'Foot du jeudi', got %q", group.Name())
	}
	if group.OrganizerID() != "org-456" {
		t.Errorf("expected organizer ID 'org-456', got %q", group.OrganizerID())
	}
	if group.WebhookURL() != "" {
		t.Errorf("expected empty webhook URL, got %q", group.WebhookURL())
	}
	if !group.CreatedAt().Equal(createdAt) {
		t.Errorf("expected createdAt %v, got %v", createdAt, group.CreatedAt())
	}
}

func TestNewGroup_AcceptsValidDiscordWebhookURL(t *testing.T) {
	t.Parallel()

	group, err := NewGroup("group-1", "Mon Groupe", "org-1", validDiscordWebhook, time.Now())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if group.WebhookURL() != validDiscordWebhook {
		t.Errorf("expected webhook URL stored as-is, got %q", group.WebhookURL())
	}
}

func TestNewGroup_TrimsWhitespaceAroundName(t *testing.T) {
	t.Parallel()

	group, err := NewGroup("group-1", "  Mon Groupe  ", "org-1", "", time.Now())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if group.Name() != "Mon Groupe" {
		t.Errorf("expected trimmed name 'Mon Groupe', got %q", group.Name())
	}
}

func TestNewGroup_ReturnsError_WhenInputsAreInvalid(t *testing.T) {
	t.Parallel()

	validTime := time.Now()
	longName := strings.Repeat("a", 101)

	cases := []struct {
		name        string
		id          GroupID
		groupName   string
		organizerID OrganizerID
		webhookURL  string
		createdAt   time.Time
		wantErr     error
	}{
		{name: "empty id", id: "", groupName: "Valid", organizerID: "org-1", createdAt: validTime, wantErr: domainerrors.ErrInvalidID},
		{name: "empty name", id: "group-1", groupName: "", organizerID: "org-1", createdAt: validTime, wantErr: domainerrors.ErrInvalidName},
		{name: "whitespace-only name", id: "group-1", groupName: "    ", organizerID: "org-1", createdAt: validTime, wantErr: domainerrors.ErrInvalidName},
		{name: "name too long", id: "group-1", groupName: longName, organizerID: "org-1", createdAt: validTime, wantErr: domainerrors.ErrInvalidName},
		{name: "empty organizer id", id: "group-1", groupName: "Valid", organizerID: "", createdAt: validTime, wantErr: domainerrors.ErrInvalidID},
		{name: "zero createdAt", id: "group-1", groupName: "Valid", organizerID: "org-1", createdAt: time.Time{}, wantErr: domainerrors.ErrInvalidDate},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			group, err := NewGroup(testCase.id, testCase.groupName, testCase.organizerID, testCase.webhookURL, testCase.createdAt)

			if group != nil {
				t.Errorf("expected nil group, got %+v", group)
			}
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("expected error %v, got %v", testCase.wantErr, err)
			}
		})
	}
}

// TestNewGroup_RejectsInvalidWebhookURL covers each of the 5 strict
// rules in validateWebhookURL plus a few edge cases that have caught
// implementation bugs in similar URL validators in the wild.
//
// URLs that contain a "user:pass@host" pattern are built at runtime
// from separate string fragments rather than written as a single
// literal: secret scanners (GitGuardian) flag the literal pattern as
// a hardcoded credential leak even in test fixtures meant to assert
// the validator REJECTS such URLs.
func TestNewGroup_RejectsInvalidWebhookURL(t *testing.T) {
	t.Parallel()

	longURL := "https://discord.com/api/webhooks/" + strings.Repeat("a", 2100)

	withUserInfo := func(userInfo, host string) string {
		return "https://" + userInfo + "@" + host + "/api/webhooks/1/abc"
	}

	cases := []struct {
		name string
		url  string
	}{
		{name: "http scheme rejected", url: "http://discord.com/api/webhooks/1/abc"},
		{name: "ftp scheme rejected", url: "ftp://discord.com/api/webhooks/1/abc"},
		{name: "missing scheme rejected", url: "//discord.com/api/webhooks/1/abc"},
		{name: "embedded user only rejected", url: withUserInfo("attacker", "discord.com")},
		{name: "embedded user+password rejected", url: withUserInfo("a:b", "discord.com")},
		{name: "percent-encoded credentials rejected", url: withUserInfo("%75:%70", "discord.com")},
		{name: "empty host rejected", url: "https:///api/webhooks/1/abc"},
		{name: "url too long rejected", url: longURL},
		{name: "control characters rejected", url: "https://discord.com/\x00abc"},
		{name: "raw whitespace rejected", url: "https://discord. com/api/webhooks/1/abc"},
		{name: "non-discord host rejected", url: "https://malicious.com/api/webhooks/1/abc"},
		{name: "discord lookalike host rejected", url: "https://discord.com.attacker.com/api/webhooks/1/abc"},
		{name: "discordapp legacy host rejected", url: "https://discordapp.com/api/webhooks/1/abc"},
		{name: "non-webhook path rejected", url: "https://discord.com/api/users/me"},
		{name: "non-numeric webhook id rejected", url: "https://discord.com/api/webhooks/abc/def"},
		{name: "missing webhook token rejected", url: "https://discord.com/api/webhooks/1"},
		{name: "control char in middle of url rejected", url: "https://discord.com/api/web\nhooks/1/abc"},
		{name: "tab character rejected", url: "https://discord.com/api/webhooks/1/abc\t"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			group, err := NewGroup("group-1", "Valid", "org-1", tc.url, time.Now())

			if group != nil {
				t.Errorf("expected nil group for invalid URL, got %+v", group)
			}
			if !errors.Is(err, domainerrors.ErrInvalidWebhookURL) {
				t.Errorf("expected ErrInvalidWebhookURL, got %v", err)
			}
		})
	}
}

func TestNewGroup_AcceptsEmptyWebhookURL(t *testing.T) {
	t.Parallel()

	// Empty is the explicit "no Discord configured" signal. Must NOT
	// fail validation.
	group, err := NewGroup("group-1", "Valid", "org-1", "", time.Now())
	if err != nil {
		t.Errorf("empty webhook URL must be accepted, got error: %v", err)
	}
	if group == nil || group.WebhookURL() != "" {
		t.Errorf("expected non-nil group with empty WebhookURL, got %+v", group)
	}
}

func TestRehydrateGroup_ReturnsGroup_WhenSnapshotIsValid(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC)
	g, err := RehydrateGroup(GroupSnapshot{
		ID:          "g-1",
		Name:        "  Persisted Group  ",
		OrganizerID: "org-1",
		WebhookURL:  "https://example.test/non-discord/path",
		CreatedAt:   createdAt,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if g.ID() != "g-1" || g.Name() != "Persisted Group" || g.OrganizerID() != "org-1" {
		t.Errorf("unexpected fields after rehydration: %+v", g)
	}
}

func TestRehydrateGroup_AcceptsNonDiscordWebhook_BecauseValidationIsOnWritePath(t *testing.T) {
	t.Parallel()

	// Persistence is trusted: rehydration must not re-run the strict
	// Discord URL rules. NewGroup is the only entry point that does.
	g, err := RehydrateGroup(GroupSnapshot{
		ID:          "g-1",
		Name:        "Legacy",
		OrganizerID: "org-1",
		WebhookURL:  "https://127.0.0.1:8443/api/webhooks/x/y",
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("rehydration must accept non-Discord URLs, got: %v", err)
	}
	if g.WebhookURL() != "https://127.0.0.1:8443/api/webhooks/x/y" {
		t.Errorf("webhook URL not preserved on rehydrate: %q", g.WebhookURL())
	}
}

func TestRehydrateGroup_RejectsInvalidShapes(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cases := []struct {
		name    string
		s       GroupSnapshot
		wantErr error
	}{
		{
			name:    "empty id",
			s:       GroupSnapshot{ID: "", Name: "X", OrganizerID: "org-1", CreatedAt: now},
			wantErr: domainerrors.ErrInvalidID,
		},
		{
			name:    "empty name",
			s:       GroupSnapshot{ID: "g", Name: "", OrganizerID: "org-1", CreatedAt: now},
			wantErr: domainerrors.ErrInvalidName,
		},
		{
			name:    "name too long",
			s:       GroupSnapshot{ID: "g", Name: strings.Repeat("a", 101), OrganizerID: "org-1", CreatedAt: now},
			wantErr: domainerrors.ErrInvalidName,
		},
		{
			name:    "empty organizer",
			s:       GroupSnapshot{ID: "g", Name: "X", OrganizerID: "", CreatedAt: now},
			wantErr: domainerrors.ErrInvalidID,
		},
		{
			name:    "zero createdAt",
			s:       GroupSnapshot{ID: "g", Name: "X", OrganizerID: "org-1"},
			wantErr: domainerrors.ErrInvalidDate,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g, err := RehydrateGroup(tc.s)
			if g != nil {
				t.Errorf("expected nil group, got %+v", g)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}
