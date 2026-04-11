package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEnvelopeRoundTripPreservesCoreFields(t *testing.T) {
	createdAt := time.Date(2026, 4, 10, 8, 12, 0, 0, time.UTC)
	input := Envelope{
		Version:   "1",
		Type:      "event.appended",
		ID:        "msg_1",
		RequestID: "req_1",
		SessionID: "sess_1",
		Cursor:    "00000001",
		CreatedAt: createdAt,
		Payload: map[string]any{
			"event_type": "run_state.changed",
			"title":      "Running",
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	var got Envelope
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}

	if got.Version != input.Version {
		t.Fatalf("Version = %q, want %q", got.Version, input.Version)
	}
	if got.Type != input.Type {
		t.Fatalf("Type = %q, want %q", got.Type, input.Type)
	}
	if got.ID != input.ID {
		t.Fatalf("ID = %q, want %q", got.ID, input.ID)
	}
	if got.RequestID != input.RequestID {
		t.Fatalf("RequestID = %q, want %q", got.RequestID, input.RequestID)
	}
	if got.SessionID != input.SessionID {
		t.Fatalf("SessionID = %q, want %q", got.SessionID, input.SessionID)
	}
	if got.Cursor != input.Cursor {
		t.Fatalf("Cursor = %q, want %q", got.Cursor, input.Cursor)
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Fatalf("CreatedAt = %s, want %s", got.CreatedAt, createdAt)
	}
}
