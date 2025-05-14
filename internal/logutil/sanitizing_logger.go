package logutil

import (
	"context"
	"fmt"
	"strings"
)

// SanitizingLogger extends SecretDetectingLogger to not only detect secrets
// but also sanitize/redact them from log output
type SanitizingLogger struct {
	*SecretDetectingLogger
	redactionString string
}

// Ensure SanitizingLogger implements LoggerInterface
var _ LoggerInterface = (*SanitizingLogger)(nil)

// NewSanitizingLogger creates a new SanitizingLogger that sanitizes secrets in log messages
func NewSanitizingLogger(delegate LoggerInterface) *SanitizingLogger {
	return &SanitizingLogger{
		SecretDetectingLogger: NewSecretDetectingLogger(delegate),
		redactionString:       "[REDACTED]",
	}
}

// SanitizeMessage replaces any detected secrets in the message with [REDACTED]
func SanitizeMessage(message string) string {
	if message == "" {
		return message
	}

	sanitized := message
	// Apply each pattern and replace any matches with the redaction string
	for _, pattern := range DefaultSecretPatterns {
		// Different handling based on the pattern type
		switch pattern.Name {
		case "API Key Format", "Generic Secret", "OpenAI API Key", "Google API Key":
			// For API keys and generic secrets, we want to replace the entire match
			sanitized = pattern.Regex.ReplaceAllString(sanitized, "[REDACTED]")

		case "Bearer Token", "Basic Auth":
			// For auth headers, keep the "Authorization: " part but redact the token
			sanitized = pattern.Regex.ReplaceAllString(sanitized, "Authorization: [REDACTED]")

		case "URL with Credentials":
			// For URLs with credentials, replace username:password@ with [REDACTED]
			sanitized = pattern.Regex.ReplaceAllString(sanitized, "https://[REDACTED]@")

		default:
			// Default case: replace the entire match
			sanitized = pattern.Regex.ReplaceAllString(sanitized, "[REDACTED]")
		}
	}

	return sanitized
}

// SanitizeError sanitizes error messages to remove sensitive information
func SanitizeError(err error) string {
	if err == nil {
		return ""
	}
	return SanitizeMessage(err.Error())
}

// SanitizeArgs sanitizes any errors or strings in argument lists
func SanitizeArgs(args []interface{}) []interface{} {
	sanitized := make([]interface{}, len(args))
	for i, arg := range args {
		if err, ok := arg.(error); ok {
			// For errors, sanitize the error message but maintain it as a string
			// Use %% to escape % characters in the sanitized string to avoid format issues
			sanitized[i] = strings.ReplaceAll(SanitizeError(err), "%", "%%")
		} else if str, ok := arg.(string); ok {
			// For strings, apply sanitization
			sanitized[i] = SanitizeMessage(str)
		} else {
			// For other types, keep as is
			sanitized[i] = arg
		}
	}
	return sanitized
}

// AddSanitizationPattern adds a custom pattern for sanitization
func (s *SanitizingLogger) AddSanitizationPattern(pattern SecretPattern) {
	s.AddPattern(pattern)
}

// SetRedactionString allows customizing the redaction placeholder
func (s *SanitizingLogger) SetRedactionString(redaction string) {
	s.redactionString = redaction
}

// WithContext returns a new logger with the given context
func (s *SanitizingLogger) WithContext(ctx context.Context) LoggerInterface {
	newLogger := &SanitizingLogger{
		SecretDetectingLogger: s.SecretDetectingLogger.WithContext(ctx).(*SecretDetectingLogger),
		redactionString:       s.redactionString,
	}
	return newLogger
}

// Debug logs a sanitized message at DEBUG level
func (s *SanitizingLogger) Debug(format string, args ...interface{}) {
	// Format the message first, then sanitize it
	msg := fmt.Sprintf(format, args...)
	sanitizedMsg := SanitizeMessage(msg)

	// We need to pass the sanitized message as is (not as a format string)
	s.delegate.Debug("%s", sanitizedMsg)
}

// Info logs a sanitized message at INFO level
func (s *SanitizingLogger) Info(format string, args ...interface{}) {
	// Format the message first, then sanitize it
	msg := fmt.Sprintf(format, args...)
	sanitizedMsg := SanitizeMessage(msg)

	// We need to pass the sanitized message as is (not as a format string)
	s.delegate.Info("%s", sanitizedMsg)
}

// Warn logs a sanitized message at WARN level
func (s *SanitizingLogger) Warn(format string, args ...interface{}) {
	// Format the message first, then sanitize it
	msg := fmt.Sprintf(format, args...)
	sanitizedMsg := SanitizeMessage(msg)

	// We need to pass the sanitized message as is (not as a format string)
	s.delegate.Warn("%s", sanitizedMsg)
}

// Error logs a sanitized message at ERROR level
func (s *SanitizingLogger) Error(format string, args ...interface{}) {
	// Format the message first, then sanitize it
	msg := fmt.Sprintf(format, args...)
	sanitizedMsg := SanitizeMessage(msg)

	// We need to pass the sanitized message as is (not as a format string)
	s.delegate.Error("%s", sanitizedMsg)
}

// Fatal logs a sanitized message at FATAL level then calls os.Exit(1)
func (s *SanitizingLogger) Fatal(format string, args ...interface{}) {
	// Format the message first, then sanitize it
	msg := fmt.Sprintf(format, args...)
	sanitizedMsg := SanitizeMessage(msg)

	// We need to pass the sanitized message as is (not as a format string)
	s.delegate.Fatal("%s", sanitizedMsg)
}

// Println implements the standard logger interface with sanitization
func (s *SanitizingLogger) Println(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	sanitizedMsg := SanitizeMessage(msg)
	s.delegate.Println(sanitizedMsg)
}

// Printf implements the standard logger interface with sanitization
func (s *SanitizingLogger) Printf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	sanitizedMsg := SanitizeMessage(msg)
	s.delegate.Printf("%s", sanitizedMsg)
}

// DebugContext logs a sanitized message at DEBUG level with context information
func (s *SanitizingLogger) DebugContext(ctx context.Context, format string, args ...interface{}) {
	// Format the message first, then sanitize it
	msg := fmt.Sprintf(format, args...)
	sanitizedMsg := SanitizeMessage(msg)

	// We need to pass the sanitized message as is (not as a format string)
	s.delegate.DebugContext(ctx, "%s", sanitizedMsg)
}

// InfoContext logs a sanitized message at INFO level with context information
func (s *SanitizingLogger) InfoContext(ctx context.Context, format string, args ...interface{}) {
	// Format the message first, then sanitize it
	msg := fmt.Sprintf(format, args...)
	sanitizedMsg := SanitizeMessage(msg)

	// We need to pass the sanitized message as is (not as a format string)
	s.delegate.InfoContext(ctx, "%s", sanitizedMsg)
}

// WarnContext logs a sanitized message at WARN level with context information
func (s *SanitizingLogger) WarnContext(ctx context.Context, format string, args ...interface{}) {
	// Format the message first, then sanitize it
	msg := fmt.Sprintf(format, args...)
	sanitizedMsg := SanitizeMessage(msg)

	// We need to pass the sanitized message as is (not as a format string)
	s.delegate.WarnContext(ctx, "%s", sanitizedMsg)
}

// ErrorContext logs a sanitized message at ERROR level with context information
func (s *SanitizingLogger) ErrorContext(ctx context.Context, format string, args ...interface{}) {
	// Format the message first, then sanitize it
	msg := fmt.Sprintf(format, args...)
	sanitizedMsg := SanitizeMessage(msg)

	// We need to pass the sanitized message as is (not as a format string)
	s.delegate.ErrorContext(ctx, "%s", sanitizedMsg)
}

// FatalContext logs a sanitized message at FATAL level with context information
// then calls os.Exit(1)
func (s *SanitizingLogger) FatalContext(ctx context.Context, format string, args ...interface{}) {
	// Format the message first, then sanitize it
	msg := fmt.Sprintf(format, args...)
	sanitizedMsg := SanitizeMessage(msg)

	// We need to pass the sanitized message as is (not as a format string)
	s.delegate.FatalContext(ctx, "%s", sanitizedMsg)
}

// WithSecretSanitization wraps a logger with secret sanitization capabilities
func WithSecretSanitization(logger LoggerInterface) *SanitizingLogger {
	return NewSanitizingLogger(logger)
}
