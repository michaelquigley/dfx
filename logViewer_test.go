package dfx

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"
)

func parseFields(t *testing.T, fields string) map[string]interface{} {
	t.Helper()
	if fields == "" {
		return map[string]interface{}{}
	}

	parsed := map[string]interface{}{}
	if err := json.Unmarshal([]byte(fields), &parsed); err != nil {
		t.Fatalf("failed to parse fields '%s': %v", fields, err)
	}
	return parsed
}

func TestSlogHandler_WithAttrsReturnsIndependentHandlers(t *testing.T) {
	buffer := NewLogBuffer(16)
	base := NewSlogHandler(buffer, &SlogHandlerOptions{
		MinLevel:  slog.LevelInfo,
		StartTime: time.Now(),
	})

	derivedA := base.WithAttrs([]slog.Attr{slog.String("scope", "a")})
	derivedB := base.WithAttrs([]slog.Attr{slog.String("scope", "b")})

	recordA := slog.NewRecord(time.Now(), slog.LevelInfo, "a", 0)
	recordB := slog.NewRecord(time.Now(), slog.LevelInfo, "b", 0)
	recordBase := slog.NewRecord(time.Now(), slog.LevelInfo, "base", 0)

	if err := derivedA.Handle(context.Background(), recordA); err != nil {
		t.Fatalf("unexpected error handling record a: %v", err)
	}
	if err := derivedB.Handle(context.Background(), recordB); err != nil {
		t.Fatalf("unexpected error handling record b: %v", err)
	}
	if err := base.Handle(context.Background(), recordBase); err != nil {
		t.Fatalf("unexpected error handling base record: %v", err)
	}

	messages := buffer.Messages()
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}

	fieldsA := parseFields(t, messages[0].Fields)
	fieldsB := parseFields(t, messages[1].Fields)
	fieldsBase := parseFields(t, messages[2].Fields)

	if fieldsA["scope"] != "a" {
		t.Fatalf("expected first record scope 'a', got '%v'", fieldsA["scope"])
	}
	if fieldsB["scope"] != "b" {
		t.Fatalf("expected second record scope 'b', got '%v'", fieldsB["scope"])
	}
	if _, ok := fieldsBase["scope"]; ok {
		t.Fatalf("expected base handler attrs to remain empty, got '%v'", fieldsBase["scope"])
	}
}

func TestNewLogViewer_DefaultDisabledMessageConfig(t *testing.T) {
	lv := NewLogViewer(nil)

	if !lv.ShowDisabledMessage {
		t.Fatalf("expected ShowDisabledMessage to default to 'true'")
	}
	if lv.DisabledMessage != "logging capture disabled" {
		t.Fatalf("expected disabled message 'logging capture disabled', got '%s'", lv.DisabledMessage)
	}
}

func TestLogViewer_ShouldRenderDisabledMessage(t *testing.T) {
	lv := NewLogViewer(nil)
	lv.Visible = true

	if !lv.shouldRenderDisabledMessage() {
		t.Fatalf("expected disabled message to render when visible and buffer is nil")
	}

	lv.ShowDisabledMessage = false
	if lv.shouldRenderDisabledMessage() {
		t.Fatalf("expected disabled message rendering to be disabled")
	}

	lv.ShowDisabledMessage = true
	lv.DisabledMessage = ""
	if lv.shouldRenderDisabledMessage() {
		t.Fatalf("expected empty disabled message to suppress rendering")
	}

	lv.DisabledMessage = "logging capture disabled"
	lv.Buffer = NewLogBuffer(1)
	if lv.shouldRenderDisabledMessage() {
		t.Fatalf("expected non-nil buffer to suppress disabled rendering")
	}

	lv.Buffer = nil
	lv.Visible = false
	if lv.shouldRenderDisabledMessage() {
		t.Fatalf("expected invisible log viewer to suppress disabled rendering")
	}
}
