package ws

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"lightac/daemon/internal/protocol"
	"lightac/daemon/internal/store"

	"nhooyr.io/websocket"
)

func TestServerHandlesHelloHeartbeatAndEventSubscribe(t *testing.T) {
	ctx := context.Background()
	eventStore, err := store.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open event store: %v", err)
	}
	defer eventStore.Close()

	if _, err := eventStore.AppendEvent(ctx, store.AppendEventInput{
		SessionID:  "sess_mock",
		EventType:  "run_state.changed",
		Title:      "Running",
		Body:       "Mock session is running",
		Severity:   "info",
		Source:     "daemon_synthetic",
		Confidence: "high",
	}); err != nil {
		t.Fatalf("append mock event: %v", err)
	}

	server := httptest.NewServer(NewServer(eventStore))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	writeEnvelope(t, ctx, conn, protocol.Envelope{
		Version:   "1",
		Type:      "hello",
		ID:        "msg_hello",
		RequestID: "req_hello",
		CreatedAt: time.Now().UTC(),
		Payload: map[string]any{
			"client_type": "m0_test_client",
		},
	})
	gotHello := readEnvelope(t, ctx, conn)
	if gotHello.Type != "hello.ack" {
		t.Fatalf("hello response type = %q, want hello.ack", gotHello.Type)
	}
	if gotHello.RequestID != "req_hello" {
		t.Fatalf("hello response request_id = %q, want req_hello", gotHello.RequestID)
	}

	writeEnvelope(t, ctx, conn, protocol.Envelope{
		Version:   "1",
		Type:      "heartbeat",
		ID:        "msg_heartbeat",
		RequestID: "req_heartbeat",
		CreatedAt: time.Now().UTC(),
	})
	gotHeartbeat := readEnvelope(t, ctx, conn)
	if gotHeartbeat.Type != "heartbeat.ack" {
		t.Fatalf("heartbeat response type = %q, want heartbeat.ack", gotHeartbeat.Type)
	}

	writeEnvelope(t, ctx, conn, protocol.Envelope{
		Version:   "1",
		Type:      "session.list",
		ID:        "msg_session_list",
		RequestID: "req_session_list",
		CreatedAt: time.Now().UTC(),
	})
	gotSessions := readEnvelope(t, ctx, conn)
	if gotSessions.Type != "session.list.result" {
		t.Fatalf("session list response type = %q, want session.list.result", gotSessions.Type)
	}
	sessions, ok := gotSessions.Payload["sessions"].([]any)
	if !ok {
		t.Fatalf("sessions payload has type %T, want []any", gotSessions.Payload["sessions"])
	}
	if len(sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(sessions))
	}

	writeEnvelope(t, ctx, conn, protocol.Envelope{
		Version:   "1",
		Type:      "event.subscribe",
		ID:        "msg_subscribe",
		RequestID: "req_subscribe",
		SessionID: "sess_mock",
		CreatedAt: time.Now().UTC(),
		Payload: map[string]any{
			"from_cursor": "",
		},
	})
	gotEvent := readEnvelope(t, ctx, conn)
	if gotEvent.Type != "event.appended" {
		t.Fatalf("event response type = %q, want event.appended", gotEvent.Type)
	}
	if gotEvent.SessionID != "sess_mock" {
		t.Fatalf("event response session_id = %q, want sess_mock", gotEvent.SessionID)
	}
	if gotEvent.Cursor != "00000001" {
		t.Fatalf("event response cursor = %q, want 00000001", gotEvent.Cursor)
	}
}

func writeEnvelope(t *testing.T, ctx context.Context, conn *websocket.Conn, envelope protocol.Envelope) {
	t.Helper()
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
		t.Fatalf("write envelope: %v", err)
	}
}

func readEnvelope(t *testing.T, ctx context.Context, conn *websocket.Conn) protocol.Envelope {
	t.Helper()
	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read envelope: %v", err)
	}
	var envelope protocol.Envelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	return envelope
}
