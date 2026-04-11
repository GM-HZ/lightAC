package store

import (
	"context"
	"testing"
)

func TestEventStoreAppendsEventsWithMonotonicCursorAndListsFromCursor(t *testing.T) {
	ctx := context.Background()
	store, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	first, err := store.AppendEvent(ctx, AppendEventInput{
		SessionID:  "sess_1",
		EventType:  "run_state.changed",
		Title:      "Running",
		Body:       "Agent is running",
		Severity:   "info",
		Source:     "daemon_synthetic",
		Confidence: "high",
	})
	if err != nil {
		t.Fatalf("append first event: %v", err)
	}
	second, err := store.AppendEvent(ctx, AppendEventInput{
		SessionID:  "sess_1",
		EventType:  "summary.generated",
		Title:      "Summary",
		Body:       "Done",
		Severity:   "success",
		Source:     "daemon_synthetic",
		Confidence: "high",
	})
	if err != nil {
		t.Fatalf("append second event: %v", err)
	}

	if first.Cursor != "00000001" {
		t.Fatalf("first cursor = %q, want 00000001", first.Cursor)
	}
	if second.Cursor != "00000002" {
		t.Fatalf("second cursor = %q, want 00000002", second.Cursor)
	}

	events, err := store.ListEventsAfter(ctx, "sess_1", "00000001", 10)
	if err != nil {
		t.Fatalf("list events after cursor: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].Cursor != "00000002" {
		t.Fatalf("events[0].Cursor = %q, want 00000002", events[0].Cursor)
	}
	if events[0].EventType != "summary.generated" {
		t.Fatalf("events[0].EventType = %q, want summary.generated", events[0].EventType)
	}
}
