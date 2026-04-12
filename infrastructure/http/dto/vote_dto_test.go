package dto

import (
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

func TestVoteResponseFromEntity_MapsAllFields(t *testing.T) {
	t.Parallel()
	v, _ := entities.NewVote("v-1", "m-1", "p-1", "p-2", 4, time.Now())
	resp := VoteResponseFromEntity(v)
	if resp.ID != "v-1" || resp.Score != 4 {
		t.Errorf("unexpected: %+v", resp)
	}
}
