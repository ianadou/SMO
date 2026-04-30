package dto

import (
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

func TestInvitationResponseFromEntity_OmitsPlainToken(t *testing.T) {
	t.Parallel()
	createdAt := time.Now()
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	inv, _ := entities.NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, nil, createdAt)

	resp := InvitationResponseFromEntity(inv)
	if resp.ID != "inv-1" {
		t.Errorf("expected 'inv-1', got %q", resp.ID)
	}
	if resp.UsedAt != nil {
		t.Errorf("expected nil UsedAt, got %v", resp.UsedAt)
	}
}
