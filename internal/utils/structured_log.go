package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

func (l Level) String() string {
	if name, ok := levelNames[l]; ok {
		return name
	}
	return fmt.Sprintf("Level(%d)", l)
}

type Logger struct {
	handler slog.Handler
	mu      sync.RWMutex
	level   Level
}

var (
	globalLogger     *Logger
	globalLoggerOnce sync.Once
)

func getGlobalLogger() *Logger {
	globalLoggerOnce.Do(func() {
		if globalLogger == nil {
			globalLogger = &Logger{
				handler: slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
					Level: slog.LevelInfo,
				}),
				level: LevelInfo,
			}
		}
	})
	return globalLogger
}

type contextKey string

const loggerContextKey contextKey = "slog_logger"

func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

// FromContext retrieves the logger from context or returns the global logger.
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerContextKey).(*Logger); ok {
		return logger
	}
	return getGlobalLogger()
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	return FromContext(ctx)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.log(context.Background(), LevelDebug, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.log(context.Background(), LevelInfo, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.log(context.Background(), LevelWarn, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.log(context.Background(), LevelError, msg, args...)
}

func (l *Logger) DebugCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelDebug, msg, args...)
}

func (l *Logger) InfoCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelInfo, msg, args...)
}

func (l *Logger) WarnCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelWarn, msg, args...)
}

func (l *Logger) ErrorCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelError, msg, args...)
}

func (l *Logger) log(ctx context.Context, level Level, msg string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if level < l.level {
		return
	}

	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := l.handler.WithAttrs([]slog.Attr{
		slog.String("level", level.String()),
	})

	record := slog.Record{
		Message: msg,
		Time:    time.Now(),
		Level:   slogLevel,
	}
	record.AddAttrs(toSlogAttrs(args...)...)

	_ = handler.Handle(ctx, record)
}

func toSlogAttrs(args ...any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		attrs = append(attrs, slog.Any(key, args[i+1]))
	}
	return attrs
}

type loggerCtx struct {
	logger *Logger
}

type handlerOptions struct {
	level      Level
	jsonOutput bool
	addSource  bool
}

type HandlerOption func(*handlerOptions)

func WithLevel(level Level) HandlerOption {
	return func(o *handlerOptions) {
		o.level = level
	}
}

func WithJSONOutput(json bool) HandlerOption {
	return func(o *handlerOptions) {
		o.jsonOutput = json
	}
}

func WithSource(addSource bool) HandlerOption {
	return func(o *handlerOptions) {
		o.addSource = addSource
	}
}

func NewHandler(opts ...HandlerOption) slog.Handler {
	options := &handlerOptions{
		level:      LevelInfo,
		jsonOutput: false,
		addSource:  false,
	}
	for _, opt := range opts {
		opt(options)
	}

	handlerOpts := &slog.HandlerOptions{
		AddSource: options.addSource,
	}
	switch options.level {
	case LevelDebug:
		handlerOpts.Level = slog.LevelDebug
	case LevelInfo:
		handlerOpts.Level = slog.LevelInfo
	case LevelWarn:
		handlerOpts.Level = slog.LevelWarn
	case LevelError:
		handlerOpts.Level = slog.LevelError
	}

	if options.jsonOutput {
		return slog.NewJSONHandler(os.Stderr, handlerOpts)
	}
	return slog.NewTextHandler(os.Stderr, handlerOpts)
}

// NewLogger creates a new Logger with the given handler.
func NewLogger(handler slog.Handler) *Logger {
	return &Logger{
		handler: handler,
		level:   LevelInfo,
	}
}

// GlobalLogger returns the global logger instance.
func GlobalLogger() *Logger {
	return getGlobalLogger()
}

// SetLevel sets the logging level for the logger.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) With(args ...any) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return NewLogger(l.handler.WithAttrs(toSlogAttrs(args...)))
}

func SetGlobalLogger(logger *Logger) {
	getGlobalLogger().mu.Lock()
	defer getGlobalLogger().mu.Unlock()
	globalLogger = logger
}

func SetGlobalLevel(level Level) {
	getGlobalLogger().SetLevel(level)
}

func SetGlobalJSONOutput(json bool) {
	getGlobalLogger().mu.Lock()
	defer getGlobalLogger().mu.Unlock()
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if json {
		getGlobalLogger().handler = slog.NewJSONHandler(os.Stderr, handlerOpts)
	} else {
		getGlobalLogger().handler = slog.NewTextHandler(os.Stderr, handlerOpts)
	}
}

func Debug(msg string, args ...any) {
	FromContext(context.Background()).Debug(msg, args...)
}

func Info(msg string, args ...any) {
	FromContext(context.Background()).Info(msg, args...)
}

func Warn(msg string, args ...any) {
	FromContext(context.Background()).Warn(msg, args...)
}

func Error(msg string, args ...any) {
	FromContext(context.Background()).Error(msg, args...)
}
