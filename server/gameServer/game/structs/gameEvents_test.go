package structs

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGameEventJSONUsesCamelCase(t *testing.T) {
	event := GameEvent{
		PlayerID:  "p1",
		Type:      EventCardPlayedAssassin,
		ExpiresAt: time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC),
		Payload:   map[string]any{"targetPlayer": "p2", "targetCardIndex": 1},
	}

	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("erro ao serializar GameEvent: %v", err)
	}

	want := `{"playerId":"p1","type":"CARD_PLAYED_ASSASSIN","expiresAt":"2026-11-05T12:00:00Z","payload":{"targetCardIndex":1,"targetPlayer":"p2"}}`
	if string(raw) != want {
		t.Fatalf("GameEvent = %s, quer %s", raw, want)
	}
}
