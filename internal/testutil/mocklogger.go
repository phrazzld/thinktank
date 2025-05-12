// Package testutil provides testing utilities for the entire codebase
package testutil

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// MockLogger implements both logutil.LoggerInterface and auditlog.AuditLogger
// for testing purposes. It records all log calls so they can be asserted in tests.
type MockLogger struct {
	mutex       sync.Mutex
	messages    []string
	debugMsgs   []string
	infoMsgs    []string
	warnMsgs    []string
	errorMsgs   []string
	fatalMsgs   []string
	logLevel    logutil.LogLevel
	verboseMode bool

	// Audit logging support
	auditEntries []auditlog.AuditEntry
	logOpCalls   []LogOpCall
	logError     error // For simulating errors in audit logging
}

// LogOpCall represents a single call to the LogOp method
type LogOpCall struct {
	Operation     string
	Status        string
	Inputs        map[string]interface{}
	Outputs       map[string]interface{}
	Error         error
	CorrelationID string // Track correlation ID from context
}

// NewMockLogger creates a new mock logger for testing
func NewMockLogger() *MockLogger {
	return &MockLogger{
		messages:     make([]string, 0),
		debugMsgs:    make([]string, 0),
		infoMsgs:     make([]string, 0),
		warnMsgs:     make([]string, 0),
		errorMsgs:    make([]string, 0),
		fatalMsgs:    make([]string, 0),
		logLevel:     logutil.DebugLevel, // Default to debug for tests
		verboseMode:  true,
		auditEntries: make([]auditlog.AuditEntry, 0),
		logOpCalls:   make([]LogOpCall, 0),
		logError:     nil,
	}
}

// SetLogError sets the error to be returned by audit logging methods
func (m *MockLogger) SetLogError(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.logError = err
}

// ClearLogError clears any configured error for audit logging methods
func (m *MockLogger) ClearLogError() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.logError = nil
}

//
// logutil.LoggerInterface implementation
//

// Println implements LoggerInterface.Println
func (m *MockLogger) Println(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = append(m.messages, msg)
	m.infoMsgs = append(m.infoMsgs, msg)
}

// Printf logs a formatted message at the default level
func (m *MockLogger) Printf(format string, args ...interface{}) {
	// Only log if we're in verbose mode and the format starts with "Verbose:"
	if m.verboseMode || !strings.HasPrefix(format, "Verbose:") {
		msg := fmt.Sprintf(format, args...)
		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.messages = append(m.messages, msg)
	}
}

// Debug logs a formatted message at debug level
func (m *MockLogger) Debug(format string, args ...interface{}) {
	if m.logLevel <= logutil.DebugLevel {
		msg := fmt.Sprintf(format, args...)
		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.messages = append(m.messages, msg)
		m.debugMsgs = append(m.debugMsgs, msg)
	}
}

// Info logs a formatted message at info level
func (m *MockLogger) Info(format string, args ...interface{}) {
	if m.logLevel <= logutil.InfoLevel {
		msg := fmt.Sprintf(format, args...)
		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.messages = append(m.messages, msg)
		m.infoMsgs = append(m.infoMsgs, msg)
	}
}

// Warn logs a formatted message at warn level
func (m *MockLogger) Warn(format string, args ...interface{}) {
	if m.logLevel <= logutil.WarnLevel {
		msg := fmt.Sprintf(format, args...)
		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.messages = append(m.messages, msg)
		m.warnMsgs = append(m.warnMsgs, msg)
	}
}

// Error logs a formatted message at error level
func (m *MockLogger) Error(format string, args ...interface{}) {
	if m.logLevel <= logutil.ErrorLevel {
		msg := fmt.Sprintf(format, args...)
		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.messages = append(m.messages, msg)
		m.errorMsgs = append(m.errorMsgs, msg)
	}
}

// Fatal logs a formatted message at fatal level
func (m *MockLogger) Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = append(m.messages, msg)
	m.fatalMsgs = append(m.fatalMsgs, msg)
	// Note: We don't exit in tests
}

// DebugContext logs a formatted message at debug level with context
func (m *MockLogger) DebugContext(ctx context.Context, format string, args ...interface{}) {
	if m.logLevel <= logutil.DebugLevel {
		msg := fmt.Sprintf(format, args...)
		// We get the correlation ID but don't format it directly in the message
		// to avoid linter warnings about manual correlation_id= formatting
		_ = logutil.GetCorrelationID(ctx)

		// Just store the message for testing purposes
		recordedMsg := msg

		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.messages = append(m.messages, recordedMsg)
		m.debugMsgs = append(m.debugMsgs, recordedMsg)
	}
}

// InfoContext logs a formatted message at info level with context
func (m *MockLogger) InfoContext(ctx context.Context, format string, args ...interface{}) {
	if m.logLevel <= logutil.InfoLevel {
		msg := fmt.Sprintf(format, args...)
		// We get the correlation ID but don't format it directly in the message
		// to avoid linter warnings about manual correlation_id= formatting
		_ = logutil.GetCorrelationID(ctx)

		// Just store the message for testing purposes
		recordedMsg := msg

		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.messages = append(m.messages, recordedMsg)
		m.infoMsgs = append(m.infoMsgs, recordedMsg)
	}
}

// WarnContext logs a formatted message at warn level with context
func (m *MockLogger) WarnContext(ctx context.Context, format string, args ...interface{}) {
	if m.logLevel <= logutil.WarnLevel {
		msg := fmt.Sprintf(format, args...)
		// We get the correlation ID but don't format it directly in the message
		// to avoid linter warnings about manual correlation_id= formatting
		_ = logutil.GetCorrelationID(ctx)

		// Just store the message for testing purposes
		recordedMsg := msg

		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.messages = append(m.messages, recordedMsg)
		m.warnMsgs = append(m.warnMsgs, recordedMsg)
	}
}

// ErrorContext logs a formatted message at error level with context
func (m *MockLogger) ErrorContext(ctx context.Context, format string, args ...interface{}) {
	if m.logLevel <= logutil.ErrorLevel {
		msg := fmt.Sprintf(format, args...)
		// We get the correlation ID but don't format it directly in the message
		// to avoid linter warnings about manual correlation_id= formatting
		_ = logutil.GetCorrelationID(ctx)

		// Just store the message for testing purposes
		recordedMsg := msg

		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.messages = append(m.messages, recordedMsg)
		m.errorMsgs = append(m.errorMsgs, recordedMsg)
	}
}

// FatalContext logs a formatted message at fatal level with context
func (m *MockLogger) FatalContext(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// We get the correlation ID but don't format it directly in the message
	// to avoid linter warnings about manual correlation_id= formatting
	_ = logutil.GetCorrelationID(ctx)

	// Just store the message for testing purposes
	recordedMsg := msg

	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = append(m.messages, recordedMsg)
	m.fatalMsgs = append(m.fatalMsgs, recordedMsg)
	// Note: We don't exit in tests
}

// WithContext returns a logger with context information
func (m *MockLogger) WithContext(ctx context.Context) logutil.LoggerInterface {
	// For mock logger, we just return the same logger
	return m
}

// SetLevel sets the log level
func (m *MockLogger) SetLevel(level logutil.LogLevel) {
	m.logLevel = level
}

// GetLevel returns the current log level
func (m *MockLogger) GetLevel() logutil.LogLevel {
	return m.logLevel
}

// SetVerbose sets the verbose mode for testing
func (m *MockLogger) SetVerbose(verbose bool) {
	m.verboseMode = verbose
}

//
// auditlog.AuditLogger implementation
//

// Log implements the AuditLogger.Log method with context
func (m *MockLogger) Log(ctx context.Context, entry auditlog.AuditEntry) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If error is configured, return it
	if m.logError != nil {
		return m.logError
	}

	// Ensure timestamp is set
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	// Add correlation ID from context if not already present
	correlationID := logutil.GetCorrelationID(ctx)
	if correlationID != "" {
		if entry.Inputs == nil {
			entry.Inputs = make(map[string]interface{})
		}
		// Only add if not already present
		if _, exists := entry.Inputs["correlation_id"]; !exists {
			entry.Inputs["correlation_id"] = correlationID
		}
	}

	// Record the entry
	m.auditEntries = append(m.auditEntries, entry)
	return nil
}

// LogLegacy implements the backward-compatible AuditLogger.LogLegacy method
func (m *MockLogger) LogLegacy(entry auditlog.AuditEntry) error {
	return m.Log(context.Background(), entry)
}

// LogOp implements the AuditLogger.LogOp method with context
func (m *MockLogger) LogOp(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Extract correlation ID from context
	ctxCorrelationID := logutil.GetCorrelationID(ctx)

	// Record the call
	m.logOpCalls = append(m.logOpCalls, LogOpCall{
		Operation:     operation,
		Status:        status,
		Inputs:        inputs,
		Outputs:       outputs,
		Error:         err,
		CorrelationID: ctxCorrelationID,
	})

	// If error is configured, return it
	if m.logError != nil {
		return m.logError
	}

	// Make a copy of inputs to avoid modifying the original map
	inputsCopy := make(map[string]interface{})
	for k, v := range inputs {
		inputsCopy[k] = v
	}

	// Add correlation ID from context if not already present
	correlationID := logutil.GetCorrelationID(ctx)
	if correlationID != "" {
		// Only add if not already present
		if _, exists := inputsCopy["correlation_id"]; !exists {
			inputsCopy["correlation_id"] = correlationID
		}
	}

	// Create and record the entry
	entry := auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: operation,
		Status:    status,
		Inputs:    inputsCopy,
		Outputs:   outputs,
	}

	// Add message based on status and operation
	var message string
	switch status {
	case "Success":
		message = fmt.Sprintf("%s completed successfully", operation)
	case "InProgress":
		message = fmt.Sprintf("%s started", operation)
	case "Failure":
		message = fmt.Sprintf("%s failed", operation)
	default:
		message = fmt.Sprintf("%s - %s", operation, status)
	}
	entry.Message = message

	// Add error information if provided
	if err != nil {
		entry.Error = &auditlog.ErrorInfo{
			Message: err.Error(),
			Type:    "TestError", // Simple error type for testing
		}
	}

	m.auditEntries = append(m.auditEntries, entry)
	return nil
}

// LogOpLegacy implements the backward-compatible AuditLogger.LogOpLegacy method
func (m *MockLogger) LogOpLegacy(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	return m.LogOp(context.Background(), operation, status, inputs, outputs, err)
}

// Close implements the AuditLogger.Close method
func (m *MockLogger) Close() error {
	// Nothing to close in the mock
	return nil
}

//
// Query methods for assertions in tests
//

// GetAuditEntries returns all recorded audit entries
func (m *MockLogger) GetAuditEntries() []auditlog.AuditEntry {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make([]auditlog.AuditEntry, len(m.auditEntries))
	copy(result, m.auditEntries)
	return result
}

// GetLogOpCalls returns all recorded LogOp calls
func (m *MockLogger) GetLogOpCalls() []LogOpCall {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make([]LogOpCall, len(m.logOpCalls))
	copy(result, m.logOpCalls)
	return result
}

// ClearAuditRecords clears all recorded audit entries and LogOp calls
func (m *MockLogger) ClearAuditRecords() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.auditEntries = make([]auditlog.AuditEntry, 0)
	m.logOpCalls = make([]LogOpCall, 0)
}

// GetMessages returns all logged messages
func (m *MockLogger) GetMessages() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make([]string, len(m.messages))
	copy(result, m.messages)
	return result
}

// GetDebugMessages returns debug level messages
func (m *MockLogger) GetDebugMessages() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make([]string, len(m.debugMsgs))
	copy(result, m.debugMsgs)
	return result
}

// GetInfoMessages returns info level messages
func (m *MockLogger) GetInfoMessages() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make([]string, len(m.infoMsgs))
	copy(result, m.infoMsgs)
	return result
}

// GetWarnMessages returns warn level messages
func (m *MockLogger) GetWarnMessages() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make([]string, len(m.warnMsgs))
	copy(result, m.warnMsgs)
	return result
}

// GetErrorMessages returns error level messages
func (m *MockLogger) GetErrorMessages() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make([]string, len(m.errorMsgs))
	copy(result, m.errorMsgs)
	return result
}

// GetFatalMessages returns fatal level messages
func (m *MockLogger) GetFatalMessages() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	result := make([]string, len(m.fatalMsgs))
	copy(result, m.fatalMsgs)
	return result
}

// ClearMessages clears all logged messages
func (m *MockLogger) ClearMessages() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = make([]string, 0)
	m.debugMsgs = make([]string, 0)
	m.infoMsgs = make([]string, 0)
	m.warnMsgs = make([]string, 0)
	m.errorMsgs = make([]string, 0)
	m.fatalMsgs = make([]string, 0)
}

// ContainsMessage checks if a message was logged (substring match)
func (m *MockLogger) ContainsMessage(substr string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, msg := range m.messages {
		if strings.Contains(msg, substr) {
			return true
		}
	}
	return false
}

// Compile-time checks to ensure interface implementation
var _ logutil.LoggerInterface = (*MockLogger)(nil)
var _ auditlog.AuditLogger = (*MockLogger)(nil)
