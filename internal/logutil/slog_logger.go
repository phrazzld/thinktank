package logutil

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// SlogLogger implements LoggerInterface using the standard log/slog package.
// It provides structured JSON logging with context-aware methods.
type SlogLogger struct {
	logger      *slog.Logger
	infoLogger  *slog.Logger
	errorLogger *slog.Logger
	ctx         context.Context
	streamSplit bool
}

// Ensure SlogLogger implements LoggerInterface
var _ LoggerInterface = (*SlogLogger)(nil)

// MultiLevelHandler is a custom slog.Handler that routes logs to different
// output streams based on their level
type MultiLevelHandler struct {
	infoHandler  slog.Handler // For DEBUG and INFO logs (to STDOUT)
	errorHandler slog.Handler // For WARN and ERROR logs (to STDERR)
}

// Enabled implements slog.Handler.Enabled
func (h *MultiLevelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

// Handle implements slog.Handler.Handle
// Routes log records to the appropriate handler based on level
func (h *MultiLevelHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level >= slog.LevelWarn {
		return h.errorHandler.Handle(ctx, record)
	}
	return h.infoHandler.Handle(ctx, record)
}

// WithAttrs implements slog.Handler.WithAttrs
func (h *MultiLevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MultiLevelHandler{
		infoHandler:  h.infoHandler.WithAttrs(attrs),
		errorHandler: h.errorHandler.WithAttrs(attrs),
	}
}

// WithGroup implements slog.Handler.WithGroup
func (h *MultiLevelHandler) WithGroup(name string) slog.Handler {
	return &MultiLevelHandler{
		infoHandler:  h.infoHandler.WithGroup(name),
		errorHandler: h.errorHandler.WithGroup(name),
	}
}

// NewSlogLogger creates a new SlogLogger with JSON formatting
func NewSlogLogger(writer io.Writer, level slog.Level) *SlogLogger {
	if writer == nil {
		writer = os.Stderr
	}

	// Create a JSON handler with the specified level
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(writer, opts)
	logger := slog.New(handler)

	return &SlogLogger{
		logger:      logger,
		ctx:         context.Background(),
		streamSplit: false,
	}
}

// NewSlogLoggerWithStreamSeparation creates a new SlogLogger that routes logs to different
// streams based on severity level:
// - DEBUG and INFO logs go to stdout (or specified infoWriter)
// - WARN and ERROR logs go to stderr (or specified errorWriter)
func NewSlogLoggerWithStreamSeparation(
	infoWriter io.Writer,
	errorWriter io.Writer,
	level slog.Level,
) *SlogLogger {
	if infoWriter == nil {
		infoWriter = os.Stdout
	}
	if errorWriter == nil {
		errorWriter = os.Stderr
	}

	// Create handler options with the specified level
	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Create separate handlers for info and error streams
	infoHandler := slog.NewJSONHandler(infoWriter, opts)
	errorHandler := slog.NewJSONHandler(errorWriter, opts)

	// Create a multiLevelHandler that routes logs based on level
	multiHandler := &MultiLevelHandler{
		infoHandler:  infoHandler,
		errorHandler: errorHandler,
	}

	// Create loggers for different purposes
	logger := slog.New(multiHandler)
	infoLogger := slog.New(infoHandler)
	errorLogger := slog.New(errorHandler)

	return &SlogLogger{
		logger:      logger,
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		ctx:         context.Background(),
		streamSplit: true,
	}
}

// WithContext returns a new logger with the given context
func (s *SlogLogger) WithContext(ctx context.Context) LoggerInterface {
	if ctx == nil {
		ctx = context.Background()
	}
	return &SlogLogger{
		logger:      s.logger,
		infoLogger:  s.infoLogger,
		errorLogger: s.errorLogger,
		ctx:         ctx,
		streamSplit: s.streamSplit,
	}
}

// Debug logs a message at DEBUG level
func (s *SlogLogger) Debug(format string, args ...interface{}) {
	// Use DebugContext with the logger's context to ensure correlation ID is included
	s.DebugContext(s.ctx, format, args...)
}

// Info logs a message at INFO level
func (s *SlogLogger) Info(format string, args ...interface{}) {
	// Use InfoContext with the logger's context to ensure correlation ID is included
	s.InfoContext(s.ctx, format, args...)
}

// Warn logs a message at WARN level
func (s *SlogLogger) Warn(format string, args ...interface{}) {
	// Use WarnContext with the logger's context to ensure correlation ID is included
	s.WarnContext(s.ctx, format, args...)
}

// Error logs a message at ERROR level
func (s *SlogLogger) Error(format string, args ...interface{}) {
	// Use ErrorContext with the logger's context to ensure correlation ID is included
	s.ErrorContext(s.ctx, format, args...)
}

// Fatal logs a message at ERROR level and then exits
func (s *SlogLogger) Fatal(format string, args ...interface{}) {
	// Use FatalContext with the logger's context to ensure correlation ID is included
	s.FatalContext(s.ctx, format, args...)
}

// DebugContext logs a message at DEBUG level with correlation ID from context
// and optional key-value pairs for structured logging
func (s *SlogLogger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	// If empty context is provided, use the logger's internal context
	if ctx == context.TODO() || ctx == context.Background() || ctx == nil {
		ctx = s.ctx
	}

	// Handle structured logging differently than format string
	var message string
	var kvPairs []interface{}

	// Check if first argument is a slog.Attr, which indicates structured logging
	isStructured := len(args) > 0 && isAttr(args[0])

	if isStructured {
		// For structured logging, use message as is and pass all args as key-value pairs
		message = msg
		kvPairs = args
	} else {
		// For format string, format the message
		message = fmt.Sprintf(msg, args...)
	}

	// Add correlation ID if present
	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		kvPairs = append(kvPairs, slog.String("correlation_id", correlationID))
	}

	// Log with appropriate logger based on stream separation
	if s.streamSplit {
		s.infoLogger.DebugContext(ctx, message, kvPairs...)
	} else {
		s.logger.DebugContext(ctx, message, kvPairs...)
	}
}

// InfoContext logs a message at INFO level with correlation ID from context
// and optional key-value pairs for structured logging
func (s *SlogLogger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	// If empty context is provided, use the logger's internal context
	if ctx == context.TODO() || ctx == context.Background() || ctx == nil {
		ctx = s.ctx
	}

	// Handle structured logging differently than format string
	var message string
	var kvPairs []interface{}

	// Check if first argument is a slog.Attr, which indicates structured logging
	isStructured := len(args) > 0 && isAttr(args[0])

	if isStructured {
		// For structured logging, use message as is and pass all args as key-value pairs
		message = msg
		kvPairs = args
	} else {
		// For format string, format the message
		message = fmt.Sprintf(msg, args...)
	}

	// Add correlation ID if present
	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		kvPairs = append(kvPairs, slog.String("correlation_id", correlationID))
	}

	// Log with appropriate logger based on stream separation
	if s.streamSplit {
		s.infoLogger.InfoContext(ctx, message, kvPairs...)
	} else {
		s.logger.InfoContext(ctx, message, kvPairs...)
	}
}

// WarnContext logs a message at WARN level with correlation ID from context
// and optional key-value pairs for structured logging
func (s *SlogLogger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	// If empty context is provided, use the logger's internal context
	if ctx == context.TODO() || ctx == context.Background() || ctx == nil {
		ctx = s.ctx
	}

	// Handle structured logging differently than format string
	var message string
	var kvPairs []interface{}

	// Check if first argument is a slog.Attr, which indicates structured logging
	isStructured := len(args) > 0 && isAttr(args[0])

	if isStructured {
		// For structured logging, use message as is and pass all args as key-value pairs
		message = msg
		kvPairs = args
	} else {
		// For format string, format the message
		message = fmt.Sprintf(msg, args...)
	}

	// Add correlation ID if present
	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		kvPairs = append(kvPairs, slog.String("correlation_id", correlationID))
	}

	// Log with appropriate logger based on stream separation
	if s.streamSplit {
		s.errorLogger.WarnContext(ctx, message, kvPairs...)
	} else {
		s.logger.WarnContext(ctx, message, kvPairs...)
	}
}

// ErrorContext logs a message at ERROR level with correlation ID from context
// and optional key-value pairs for structured logging
func (s *SlogLogger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	// If empty context is provided, use the logger's internal context
	if ctx == context.TODO() || ctx == context.Background() || ctx == nil {
		ctx = s.ctx
	}

	// Handle structured logging differently than format string
	var message string
	var kvPairs []interface{}

	// Check if first argument is a slog.Attr, which indicates structured logging
	isStructured := len(args) > 0 && isAttr(args[0])

	if isStructured {
		// For structured logging, use message as is and pass all args as key-value pairs
		message = msg
		kvPairs = args
	} else {
		// For format string, format the message
		message = fmt.Sprintf(msg, args...)
	}

	// Add correlation ID if present
	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		kvPairs = append(kvPairs, slog.String("correlation_id", correlationID))
	}

	// Log with appropriate logger based on stream separation
	if s.streamSplit {
		s.errorLogger.ErrorContext(ctx, message, kvPairs...)
	} else {
		s.logger.ErrorContext(ctx, message, kvPairs...)
	}
}

// FatalContext logs a message at ERROR level with correlation ID from context
// and optional key-value pairs for structured logging, then exits
func (s *SlogLogger) FatalContext(ctx context.Context, msg string, args ...interface{}) {
	// If empty context is provided, use the logger's internal context
	if ctx == context.TODO() || ctx == context.Background() || ctx == nil {
		ctx = s.ctx
	}

	// Handle structured logging differently than format string
	var message string
	var kvPairs []interface{}

	// Check if first argument is a slog.Attr, which indicates structured logging
	isStructured := len(args) > 0 && isAttr(args[0])

	if isStructured {
		// For structured logging, use message as is and pass all args as key-value pairs
		message = msg
		kvPairs = args
	} else {
		// For format string, format the message
		message = fmt.Sprintf(msg, args...)
	}

	// Add correlation ID if present
	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		kvPairs = append(kvPairs, slog.String("correlation_id", correlationID))
	}

	// Log with appropriate logger based on stream separation
	if s.streamSplit {
		s.errorLogger.ErrorContext(ctx, message, kvPairs...)
	} else {
		s.logger.ErrorContext(ctx, message, kvPairs...)
	}
	osExit(1)
}

// isAttr checks if the interface is a slog.Attr or can be used as a structured logging attribute
func isAttr(arg interface{}) bool {
	switch arg.(type) {
	case slog.Attr, *slog.Attr, slog.Value:
		return true
	default:
		// Check if it's a slog helper function result like slog.String()
		return fmt.Sprintf("%T", arg) == "slog.Attr"
	}
}

// Println implements the standard logger interface, logs at INFO level
func (s *SlogLogger) Println(v ...interface{}) {
	message := fmt.Sprintln(v...)
	// Use InfoContext with the logger's context to ensure correlation ID is included
	s.InfoContext(s.ctx, message)
}

// Printf implements the standard logger interface, logs at INFO level
func (s *SlogLogger) Printf(format string, v ...interface{}) {
	// Use InfoContext with the logger's context to ensure correlation ID is included
	s.InfoContext(s.ctx, format, v...)
}

// ConvertLogLevelToSlog converts our LogLevel to slog.Level
func ConvertLogLevelToSlog(level LogLevel) slog.Level {
	switch level {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// NewSlogLoggerFromLogLevel creates a new SlogLogger with a level from our LogLevel enum
func NewSlogLoggerFromLogLevel(writer io.Writer, level LogLevel) *SlogLogger {
	slogLevel := ConvertLogLevelToSlog(level)
	return NewSlogLogger(writer, slogLevel)
}

// NewSlogLoggerWithStreamSeparationFromLogLevel creates a new SlogLogger with stream separation
// using a level from our LogLevel enum
func NewSlogLoggerWithStreamSeparationFromLogLevel(
	infoWriter io.Writer,
	errorWriter io.Writer,
	level LogLevel,
) *SlogLogger {
	slogLevel := ConvertLogLevelToSlog(level)
	return NewSlogLoggerWithStreamSeparation(infoWriter, errorWriter, slogLevel)
}

// EnableStreamSeparation is a helper function to create a new logger with stream
// separation, using stdout for info/debug and stderr for warn/error
func EnableStreamSeparation(logger *SlogLogger) *SlogLogger {
	// If logger is already using stream separation, return it as is
	if logger.streamSplit {
		return logger
	}

	// Determine what level to use from the existing logger
	var level slog.Level
	if logger.logger.Handler().Enabled(context.Background(), slog.LevelDebug) {
		level = slog.LevelDebug
	} else if logger.logger.Handler().Enabled(context.Background(), slog.LevelInfo) {
		level = slog.LevelInfo
	} else if logger.logger.Handler().Enabled(context.Background(), slog.LevelWarn) {
		level = slog.LevelWarn
	} else {
		level = slog.LevelError
	}

	// Create a new logger with stream separation and copy the context
	newLogger := NewSlogLoggerWithStreamSeparation(os.Stdout, os.Stderr, level)
	newLogger.ctx = logger.ctx

	return newLogger
}
