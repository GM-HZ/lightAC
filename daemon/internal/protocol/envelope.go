package protocol

import "time"

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
