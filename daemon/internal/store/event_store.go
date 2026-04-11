package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type EventStore struct {
	db *sql.DB
}

type AppendEventInput struct {
	SessionID  string
	EventType  string
	Title      string
	Body       string
	Severity   string
	Source     string
	Confidence string
}

type AgentEvent struct {
	ID         int64
	SessionID  string
	Cursor     string
	EventType  string
	Title      string
	Body       string
	Severity   string
	Source     string
	Confidence string
	CreatedAt  time.Time
}

func Open(ctx context.Context, path string) (*EventStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	store := &EventStore{db: db}
	if err := store.init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *EventStore) Close() error {
	return s.db.Close()
}

func (s *EventStore) init(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS agent_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id TEXT NOT NULL,
	cursor INTEGER NOT NULL,
	event_type TEXT NOT NULL,
	title TEXT NOT NULL,
	body TEXT NOT NULL,
	severity TEXT NOT NULL,
	source TEXT NOT NULL,
	confidence TEXT NOT NULL,
	created_at TEXT NOT NULL,
	UNIQUE(session_id, cursor)
);
CREATE INDEX IF NOT EXISTS idx_agent_events_session_cursor
	ON agent_events(session_id, cursor);
`)
	if err != nil {
		return fmt.Errorf("initialize event store: %w", err)
	}
	return nil
}

func (s *EventStore) AppendEvent(ctx context.Context, input AppendEventInput) (AgentEvent, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AgentEvent{}, fmt.Errorf("begin append event transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var lastCursor sql.NullInt64
	if err = tx.QueryRowContext(ctx, `
SELECT MAX(cursor)
FROM agent_events
WHERE session_id = ?
`, input.SessionID).Scan(&lastCursor); err != nil {
		return AgentEvent{}, fmt.Errorf("read last cursor: %w", err)
	}

	nextCursor := int64(1)
	if lastCursor.Valid {
		nextCursor = lastCursor.Int64 + 1
	}
	createdAt := time.Now().UTC().Truncate(time.Microsecond)

	result, err := tx.ExecContext(ctx, `
INSERT INTO agent_events (
	session_id, cursor, event_type, title, body, severity, source, confidence, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`, input.SessionID, nextCursor, input.EventType, input.Title, input.Body, input.Severity, input.Source, input.Confidence, createdAt.Format(time.RFC3339Nano))
	if err != nil {
		return AgentEvent{}, fmt.Errorf("insert event: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return AgentEvent{}, fmt.Errorf("read inserted event id: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return AgentEvent{}, fmt.Errorf("commit append event transaction: %w", err)
	}

	return AgentEvent{
		ID:         id,
		SessionID:  input.SessionID,
		Cursor:     formatCursor(nextCursor),
		EventType:  input.EventType,
		Title:      input.Title,
		Body:       input.Body,
		Severity:   input.Severity,
		Source:     input.Source,
		Confidence: input.Confidence,
		CreatedAt:  createdAt,
	}, nil
}

func (s *EventStore) ListEventsAfter(ctx context.Context, sessionID string, cursor string, limit int) ([]AgentEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	cursorNumber, err := parseCursor(cursor)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT id, session_id, cursor, event_type, title, body, severity, source, confidence, created_at
FROM agent_events
WHERE session_id = ? AND cursor > ?
ORDER BY cursor ASC
LIMIT ?
`, sessionID, cursorNumber, limit)
	if err != nil {
		return nil, fmt.Errorf("query events after cursor: %w", err)
	}
	defer rows.Close()

	var events []AgentEvent
	for rows.Next() {
		var event AgentEvent
		var cursorNumber int64
		var createdAt string
		if err := rows.Scan(
			&event.ID,
			&event.SessionID,
			&cursorNumber,
			&event.EventType,
			&event.Title,
			&event.Body,
			&event.Severity,
			&event.Source,
			&event.Confidence,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse event created_at: %w", err)
		}
		event.Cursor = formatCursor(cursorNumber)
		event.CreatedAt = parsedCreatedAt
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}
	return events, nil
}

func formatCursor(cursor int64) string {
	return fmt.Sprintf("%08d", cursor)
}

func parseCursor(cursor string) (int64, error) {
	if cursor == "" {
		return 0, nil
	}
	var parsed int64
	if _, err := fmt.Sscanf(cursor, "%d", &parsed); err != nil {
		return 0, fmt.Errorf("parse cursor %q: %w", cursor, err)
	}
	return parsed, nil
}
