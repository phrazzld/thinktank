package config

import (
	"reflect"
	"testing"

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
	m.debugMessages = append(m.debugMessages, format)
}

func (m *mockLogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *mockLogger) Warn(format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, format)
}

func (m *mockLogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, format)
}

func (m *mockLogger) Fatal(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, "FATAL: "+format)
}

func (m *mockLogger) Printf(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *mockLogger) Println(args ...interface{}) {
	m.infoMessages = append(m.infoMessages, "println")
}

func (m *mockLogger) SetLevel(level logutil.LogLevel) {}

func TestAppConfigStructHasNoFieldClarifyTask(t *testing.T) {
	// Get the type of AppConfig struct
	configType := reflect.TypeOf(AppConfig{})

	// Check if the ClarifyTask field exists
	_, exists := configType.FieldByName("ClarifyTask")
	if exists {
		t.Error("ClarifyTask field still exists in AppConfig struct but should have been removed")
	}
}