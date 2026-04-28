package idgen

import (
	"testing"

	"github.com/google/uuid"
)

func TestUUIDGenerator_Generate_ReturnsValidUUIDv4(t *testing.T) {
	t.Parallel()

	g := New()
	got := g.Generate()

	parsed, err := uuid.Parse(got)
	if err != nil {
		t.Fatalf("Generate() returned %q which is not a valid UUID: %v", got, err)
	}
	if parsed.Version() != 4 {
		t.Errorf("expected UUID v4, got version %d (%q)", parsed.Version(), got)
	}
}

func TestUUIDGenerator_Generate_ReturnsDistinctValues(t *testing.T) {
	t.Parallel()

	g := New()
	seen := make(map[string]struct{}, 100)
	for range 100 {
		id := g.Generate()
		if _, dup := seen[id]; dup {
			t.Fatalf("Generate() returned a duplicate within 100 calls: %q", id)
		}
		seen[id] = struct{}{}
	}
}
