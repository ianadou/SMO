package loginlockout

import (
	"context"
	"testing"
)

func TestNoopTracker_IsLocked_AlwaysFalse(t *testing.T) {
	t.Parallel()

	tracker := NewNoopTracker()

	locked, err := tracker.IsLocked(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if locked {
		t.Errorf("NoopTracker must report not-locked, got locked=true")
	}
}

func TestNoopTracker_RecordFailureAndSuccess_NeverError(t *testing.T) {
	t.Parallel()

	tracker := NewNoopTracker()
	ctx := context.Background()

	if err := tracker.RecordFailure(ctx, "alice@example.com"); err != nil {
		t.Errorf("RecordFailure must be a no-op, got error: %v", err)
	}
	if err := tracker.RecordSuccess(ctx, "alice@example.com"); err != nil {
		t.Errorf("RecordSuccess must be a no-op, got error: %v", err)
	}
}

func TestHashEmail_NormalizesCaseAndTruncates(t *testing.T) {
	t.Parallel()

	lower := hashEmail("alice@example.com")
	upper := hashEmail("Alice@Example.COM")

	if lower != upper {
		t.Errorf("hashEmail must be case-insensitive, got %q vs %q", lower, upper)
	}
	if len(lower) != emailHashLen {
		t.Errorf("hashEmail must truncate to %d chars, got %d (%q)", emailHashLen, len(lower), lower)
	}
}

func TestHashEmail_DistinctEmailsHaveDistinctHashes(t *testing.T) {
	t.Parallel()

	a := hashEmail("alice@example.com")
	b := hashEmail("bob@example.com")

	if a == b {
		t.Errorf("hashEmail must not collide on distinct emails, both produced %q", a)
	}
}
