package fileutil

import (
	"fmt"
	"strings"
	"sync"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// MockLogger implements logutil.LoggerInterface for testing
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
}

// NewMockLogger creates a new mock logger for testing
func NewMockLogger() *MockLogger {
	return &MockLogger{
		messages:    make([]string, 0),
		debugMsgs:   make([]string, 0),
		infoMsgs:    make([]string, 0),
		warnMsgs:    make([]string, 0),
		errorMsgs:   make([]string, 0),
		fatalMsgs:   make([]string, 0),
		logLevel:    logutil.DebugLevel, // Default to debug for tests
		verboseMode: true,
	}
}

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

// SetLevel sets the log level
func (m *MockLogger) SetLevel(level logutil.LogLevel) {
	m.logLevel = level
}

// GetLevel returns the current log level
func (m *MockLogger) GetLevel() logutil.LogLevel {
	return m.logLevel
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

// SetVerbose sets the verbose mode for testing
func (m *MockLogger) SetVerbose(verbose bool) {
	m.verboseMode = verbose
}
