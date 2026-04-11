package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"lightac/daemon/internal/protocol"
	"lightac/daemon/internal/store"

	"nhooyr.io/websocket"
)

type Server struct {
	store *store.EventStore
}

func NewServer(store *store.EventStore) http.Handler {
	return &Server{store: store}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ws" {
		http.NotFound(w, r)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx := r.Context()
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
		}

		var envelope protocol.Envelope
		if err := json.Unmarshal(data, &envelope); err != nil {
			_ = sendEnvelope(ctx, conn, protocol.Envelope{
				Version:   "1",
				Type:      "error",
				ID:        "msg_error",
				RequestID: envelope.RequestID,
				CreatedAt: time.Now().UTC(),
				Payload: map[string]any{
					"code":    "invalid_json",
					"message": err.Error(),
				},
			})
			continue
		}

		if err := s.handleEnvelope(ctx, conn, envelope); err != nil {
			_ = sendEnvelope(ctx, conn, protocol.Envelope{
				Version:   "1",
				Type:      "error",
				ID:        "msg_error",
				RequestID: envelope.RequestID,
				SessionID: envelope.SessionID,
				CreatedAt: time.Now().UTC(),
				Payload: map[string]any{
					"code":    "handler_error",
					"message": err.Error(),
				},
			})
		}
	}
}

func (s *Server) handleEnvelope(ctx context.Context, conn *websocket.Conn, envelope protocol.Envelope) error {
	switch envelope.Type {
	case "hello":
		return sendEnvelope(ctx, conn, protocol.Envelope{
			Version:   "1",
			Type:      "hello.ack",
			ID:        "msg_hello_ack",
			RequestID: envelope.RequestID,
			CreatedAt: time.Now().UTC(),
			Payload: map[string]any{
				"protocol_version": "1",
			},
		})
	case "heartbeat":
		return sendEnvelope(ctx, conn, protocol.Envelope{
			Version:   "1",
			Type:      "heartbeat.ack",
			ID:        "msg_heartbeat_ack",
			RequestID: envelope.RequestID,
			CreatedAt: time.Now().UTC(),
		})
	case "session.list":
		return sendEnvelope(ctx, conn, protocol.Envelope{
			Version:   "1",
			Type:      "session.list.result",
			ID:        "msg_session_list_result",
			RequestID: envelope.RequestID,
			CreatedAt: time.Now().UTC(),
			Payload: map[string]any{
				"sessions": []map[string]any{
					{
						"session_id":    "sess_mock",
						"provider_type": "codex",
						"backend_type":  "mock",
						"title":         "Mock Session",
						"status":        "running",
						"is_active":     true,
						"has_attention": false,
					},
				},
			},
		})
	case "event.subscribe":
		fromCursor, _ := envelope.Payload["from_cursor"].(string)
		events, err := s.store.ListEventsAfter(ctx, envelope.SessionID, fromCursor, 100)
		if err != nil {
			return err
		}
		for _, event := range events {
			if err := sendEnvelope(ctx, conn, protocol.Envelope{
				Version:   "1",
				Type:      "event.appended",
				ID:        "msg_event_" + event.Cursor,
				RequestID: envelope.RequestID,
				SessionID: event.SessionID,
				Cursor:    event.Cursor,
				CreatedAt: event.CreatedAt,
				Payload: map[string]any{
					"event_id":   event.ID,
					"event_type": event.EventType,
					"title":      event.Title,
					"body":       event.Body,
					"severity":   event.Severity,
					"source":     event.Source,
					"confidence": event.Confidence,
				},
			}); err != nil {
				return err
			}
		}
		return nil
	default:
		return sendEnvelope(ctx, conn, protocol.Envelope{
			Version:   "1",
			Type:      "error",
			ID:        "msg_error",
			RequestID: envelope.RequestID,
			SessionID: envelope.SessionID,
			CreatedAt: time.Now().UTC(),
			Payload: map[string]any{
				"code": "message_type_unsupported",
			},
		})
	}
}

func sendEnvelope(ctx context.Context, conn *websocket.Conn, envelope protocol.Envelope) error {
	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, data)
}
