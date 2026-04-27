package entities

import (
	"errors"
	"strings"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestNewOrganizer_BuildsEntity_WhenInputsAreValid(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC)
	o, err := NewOrganizer("org-1", "  Alice@Example.COM  ", "  Alice Doe  ", "hash-not-validated-here", createdAt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if o.ID() != "org-1" {
		t.Errorf("expected ID 'org-1', got %q", o.ID())
	}
	if o.Email() != "alice@example.com" {
		t.Errorf("expected lowercased email, got %q", o.Email())
	}
	if o.DisplayName() != "Alice Doe" {
		t.Errorf("expected trimmed display name, got %q", o.DisplayName())
	}
}

func TestNewOrganizer_RejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	validTime := time.Now()
	cases := []struct {
		name    string
		id      OrganizerID
		email   string
		display string
		hash    string
		ts      time.Time
		want    error
	}{
		{name: "empty id", id: "", email: "a@b.c", display: "X", hash: "h", ts: validTime, want: domainerrors.ErrInvalidID},
		{name: "missing at sign", id: "o", email: "not-an-email", display: "X", hash: "h", ts: validTime, want: domainerrors.ErrInvalidEmail},
		{name: "empty display", id: "o", email: "a@b.c", display: "", hash: "h", ts: validTime, want: domainerrors.ErrInvalidName},
		{name: "display too long", id: "o", email: "a@b.c", display: strings.Repeat("a", 51), hash: "h", ts: validTime, want: domainerrors.ErrInvalidName},
		{name: "empty hash", id: "o", email: "a@b.c", display: "X", hash: "", ts: validTime, want: domainerrors.ErrInvalidPassword},
		{name: "zero timestamp", id: "o", email: "a@b.c", display: "X", hash: "h", ts: time.Time{}, want: domainerrors.ErrInvalidDate},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			o, err := NewOrganizer(tc.id, tc.email, tc.display, tc.hash, tc.ts)
			if o != nil {
				t.Errorf("expected nil organizer, got %+v", o)
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("expected %v, got %v", tc.want, err)
			}
		})
	}
}
