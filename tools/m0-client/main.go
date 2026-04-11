package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"nhooyr.io/websocket"
)

type Envelope struct {
	Version   string         `json:"version"`
	Type      string         `json:"type"`
	ID        string         `json:"id"`
	RequestID string         `json:"request_id,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	Cursor    string         `json:"cursor,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	Payload   map[string]any `json:"payload,omitempty"`
}

func main() {
	url := flag.String("url", "ws://127.0.0.1:8765/ws", "daemon websocket URL")
	flag.Parse()

	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, *url, nil)
	if err != nil {
		log.Fatalf("dial daemon: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	exchange(ctx, conn, Envelope{
		Version:   "1",
		Type:      "hello",
		ID:        "msg_hello",
		RequestID: "req_hello",
		CreatedAt: time.Now().UTC(),
		Payload: map[string]any{
			"client_type": "m0_client",
		},
	})
	exchange(ctx, conn, Envelope{
		Version:   "1",
		Type:      "heartbeat",
		ID:        "msg_heartbeat",
		RequestID: "req_heartbeat",
		CreatedAt: time.Now().UTC(),
	})
	exchange(ctx, conn, Envelope{
		Version:   "1",
		Type:      "session.list",
		ID:        "msg_session_list",
		RequestID: "req_session_list",
		CreatedAt: time.Now().UTC(),
	})
	exchange(ctx, conn, Envelope{
		Version:   "1",
		Type:      "event.subscribe",
		ID:        "msg_event_subscribe",
		RequestID: "req_event_subscribe",
		SessionID: "sess_mock",
		CreatedAt: time.Now().UTC(),
		Payload: map[string]any{
			"from_cursor": "",
		},
	})
}

func exchange(ctx context.Context, conn *websocket.Conn, envelope Envelope) {
	data, err := json.Marshal(envelope)
	if err != nil {
		log.Fatalf("marshal %s: %v", envelope.Type, err)
	}
	if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
		log.Fatalf("write %s: %v", envelope.Type, err)
	}
	_, response, err := conn.Read(ctx)
	if err != nil {
		log.Fatalf("read response for %s: %v", envelope.Type, err)
	}
	fmt.Println(string(response))
}
