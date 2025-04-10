package config

import (
	"fmt"
	"github.com/phrazzld/architect/internal/logutil"
)

// mockLogger implements logutil.LoggerInterface for testing
type mockLogger struct {
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		debugMessages: []string{},
		infoMessages:  []string{},
		warnMessages:  []string{},
		errorMessages: []string{},
	}
}

func (m *mockLogger) Debug(format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) Warn(format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) Fatal(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf("FATAL: "+format, args...))
}

func (m *mockLogger) Printf(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) Println(args ...interface{}) {
	m.infoMessages = append(m.infoMessages, fmt.Sprint(args...))
}

func (m *mockLogger) SetLevel(level logutil.LogLevel) {}