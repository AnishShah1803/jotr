package utils

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestLoggerLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		expected string
	}{
		{"Debug level", LevelDebug, "DEBUG"},
		{"Info level", LevelInfo, "INFO"},
		{"Warn level", LevelWarn, "WARN"},
		{"Error level", LevelError, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	handler := slog.NewTextHandler(&bytes.Buffer{}, nil)
	logger := NewLogger(handler)
	if logger == nil {
		t.Error("NewLogger returned nil")
	}
}

func TestLoggerSetLevel(t *testing.T) {
	handler := slog.NewTextHandler(&bytes.Buffer{}, nil)
	logger := NewLogger(handler)

	logger.SetLevel(LevelDebug)
	if logger.level != LevelDebug {
		t.Errorf("SetLevel did not set level correctly")
	}

	logger.SetLevel(LevelError)
	if logger.level != LevelError {
		t.Errorf("SetLevel did not set level correctly")
	}
}

func TestLoggerWith(t *testing.T) {
	handler := slog.NewTextHandler(&bytes.Buffer{}, nil)
	logger := NewLogger(handler)

	loggerWith := logger.With("key", "value")
	if loggerWith == nil {
		t.Error("With returned nil")
	}
}

func TestLoggerDebug(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := NewLogger(handler)
	logger.SetLevel(LevelDebug)

	logger.Debug("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Debug output did not contain message, got: %s", output)
	}
}

func TestLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := NewLogger(handler)

	logger.Info("info message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Errorf("Info output did not contain message, got: %s", output)
	}
}

func TestLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := NewLogger(handler)

	logger.Warn("warn message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "warn message") {
		t.Errorf("Warn output did not contain message, got: %s", output)
	}
}

func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := NewLogger(handler)

	logger.Error("error message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "error message") {
		t.Errorf("Error output did not contain message, got: %s", output)
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})
	logger := NewLogger(handler)
	logger.SetLevel(LevelError)

	logger.Info("this should not appear")
	logger.Debug("this should not appear either")

	output := buf.String()
	if output != "" {
		t.Errorf("Expected no output for filtered levels, got: %s", output)
	}
}

func TestGlobalLogger(t *testing.T) {
	logger := GlobalLogger()
	if logger == nil {
		t.Error("GlobalLogger returned nil")
	}
}

func TestFromContext(t *testing.T) {
	ctx := context.Background()
	logger := FromContext(ctx)
	if logger == nil {
		t.Error("FromContext returned nil")
	}
}

func TestWithLogger(t *testing.T) {
	ctx := context.Background()
	handler := slog.NewTextHandler(&bytes.Buffer{}, nil)
	logger := NewLogger(handler)

	ctx = WithLogger(ctx, logger)
	retrieved := FromContext(ctx)

	if retrieved != logger {
		t.Error("WithLogger did not correctly store logger in context")
	}
}

func TestNewHandler(t *testing.T) {
	handler := NewHandler()
	if handler == nil {
		t.Error("NewHandler returned nil")
	}

	handlerJSON := NewHandler(WithJSONOutput(true))
	if handlerJSON == nil {
		t.Error("NewHandler with JSON output returned nil")
	}

	handlerDebug := NewHandler(WithLevel(LevelDebug))
	if handlerDebug == nil {
		t.Error("NewHandler with debug level returned nil")
	}

	handlerSource := NewHandler(WithSource(true))
	if handlerSource == nil {
		t.Error("NewHandler with source returned nil")
	}
}

func TestSetGlobalLevel(t *testing.T) {
	SetGlobalLevel(LevelDebug)
	logger := GlobalLogger()
	if logger.level != LevelDebug {
		t.Errorf("SetGlobalLevel did not set global logger level correctly")
	}
}

func TestSetGlobalJSONOutput(t *testing.T) {
	SetGlobalJSONOutput(true)
}

func TestLoggerWithContext(t *testing.T) {
	handler := slog.NewTextHandler(&bytes.Buffer{}, nil)
	logger := NewLogger(handler)

	ctx := context.Background()
	loggerWithCtx := logger.WithContext(ctx)

	if loggerWithCtx == nil {
		t.Error("WithContext returned nil")
	}
}

func TestLoggerCtxMethods(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := NewLogger(handler)
	logger.SetLevel(LevelDebug)

	ctx := context.Background()

	logger.DebugCtx(ctx, "debug ctx message")
	logger.InfoCtx(ctx, "info ctx message")
	logger.WarnCtx(ctx, "warn ctx message")
	logger.ErrorCtx(ctx, "error ctx message")

	output := buf.String()
	if !strings.Contains(output, "debug ctx message") {
		t.Errorf("DebugCtx output did not contain message, got: %s", output)
	}
	if !strings.Contains(output, "info ctx message") {
		t.Errorf("InfoCtx output did not contain message, got: %s", output)
	}
	if !strings.Contains(output, "warn ctx message") {
		t.Errorf("WarnCtx output did not contain message, got: %s", output)
	}
	if !strings.Contains(output, "error ctx message") {
		t.Errorf("ErrorCtx output did not contain message, got: %s", output)
	}
}

func TestToSlogAttrs(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		expected int
	}{
		{
			name:     "empty args",
			args:     []any{},
			expected: 0,
		},
		{
			name:     "single pair",
			args:     []any{"key", "value"},
			expected: 1,
		},
		{
			name:     "multiple pairs",
			args:     []any{"key1", "value1", "key2", 42, "key3", true},
			expected: 3,
		},
		{
			name:     "unmatched pair",
			args:     []any{"key"},
			expected: 0,
		},
		{
			name:     "non-string key",
			args:     []any{123, "value"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := toSlogAttrs(tt.args...)
			if len(attrs) != tt.expected {
				t.Errorf("toSlogAttrs returned %d attrs, want %d", len(attrs), tt.expected)
			}
		})
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	var buf bytes.Buffer
	SetGlobalLogger(NewLogger(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	SetGlobalLevel(LevelDebug)

	Debug("debug msg")
	Info("info msg")
	Warn("warn msg")
	Error("error msg")

	output := buf.String()
	if !strings.Contains(output, "debug msg") {
		t.Error("Debug function did not log")
	}
	if !strings.Contains(output, "info msg") {
		t.Error("Info function did not log")
	}
	if !strings.Contains(output, "warn msg") {
		t.Error("Warn function did not log")
	}
	if !strings.Contains(output, "error msg") {
		t.Error("Error function did not log")
	}
}

func TestJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	handler := NewLogger(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler.Info("json message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "json message") {
		t.Errorf("JSON output did not contain message, got: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("JSON output did not contain expected key-value pair, got: %s", output)
	}
}
