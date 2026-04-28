package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

func TestGroupResponseFromEntity_BuildsResponseFromGroup(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 4, 10, 14, 0, 0, 0, time.UTC)
	group, err := entities.NewGroup("group-1", "Foot du jeudi", "org-1", "", createdAt)
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}

	response := GroupResponseFromEntity(group)

	if response.ID != "group-1" {
		t.Errorf("expected ID 'group-1', got %q", response.ID)
	}
	if response.Name != "Foot du jeudi" {
		t.Errorf("expected name 'Foot du jeudi', got %q", response.Name)
	}
	if response.OrganizerID != "org-1" {
		t.Errorf("expected OrganizerID 'org-1', got %q", response.OrganizerID)
	}
	if response.HasWebhook {
		t.Errorf("expected HasWebhook=false for empty webhook URL, got true")
	}
	if !response.CreatedAt.Equal(createdAt) {
		t.Errorf("expected CreatedAt %v, got %v", createdAt, response.CreatedAt)
	}
}

func TestGroupResponseFromEntity_HasWebhookIsTrue_WhenWebhookConfigured(t *testing.T) {
	t.Parallel()

	const webhookURL = "https://discord.com/api/webhooks/123/abc"
	group, err := entities.NewGroup("group-1", "Group", "org-1", webhookURL, time.Now())
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}

	response := GroupResponseFromEntity(group)

	if !response.HasWebhook {
		t.Errorf("expected HasWebhook=true when webhook configured, got false")
	}
}

// TestGroupResponseFromEntity_NeverIncludesWebhookURL_InJSON is a
// guard: any future change that adds the URL field to the response
// (intentionally or by reflex) will fail this test. The webhook URL
// is a secret and must never travel back to the client in clear.
func TestGroupResponseFromEntity_NeverIncludesWebhookURL_InJSON(t *testing.T) {
	t.Parallel()

	const secretURL = "https://discord.com/api/webhooks/SECRET-1234567890/SECRET-token-abcdef"
	group, _ := entities.NewGroup("group-1", "Group", "org-1", secretURL, time.Now())

	response := GroupResponseFromEntity(group)
	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	jsonStr := string(raw)
	for _, snippet := range []string{
		secretURL,
		"SECRET-1234567890",
		"SECRET-token-abcdef",
		"discord.com/api/webhooks",
	} {
		if strings.Contains(jsonStr, snippet) {
			t.Errorf("JSON response leaks webhook secret %q: %s", snippet, jsonStr)
		}
	}
}
