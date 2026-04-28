package events

import (
	"testing"
	"time"
)

func TestMatchTeamsReady_EventName_IsStable(t *testing.T) {
	t.Parallel()

	ev := MatchTeamsReady{GroupID: "g-1", MatchID: "m-1", OccurredAt: time.Now()}

	if ev.EventName() != MatchTeamsReadyEventName {
		t.Errorf("expected EventName %q, got %q", MatchTeamsReadyEventName, ev.EventName())
	}
	if MatchTeamsReadyEventName != "match.teams_ready" {
		t.Errorf("MatchTeamsReadyEventName changed: subscribers will silently stop receiving — see ADR 0004")
	}
}
