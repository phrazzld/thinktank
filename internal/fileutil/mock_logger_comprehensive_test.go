package fileutil

import (
	"context"
	"strings"
	"testing"
)

func TestMockLogger_BasicLogging(t *testing.T) {
	logger := NewMockLogger()

	// Test Debug
	logger.Debug("debug message")
	messages := logger.GetDebugMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 debug message, got %d", len(messages))
	}
	if !strings.Contains(messages[0], "debug message") {
		t.Errorf("Expected debug message to contain 'debug message', got: %s", messages[0])
	}

	// Test Info
	logger.Info("info message")
	infoMessages := logger.GetInfoMessages()
	if len(infoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(infoMessages))
	}

	// Test Warn
	logger.Warn("warn message")
	warnMessages := logger.GetWarnMessages()
	if len(warnMessages) != 1 {
		t.Errorf("Expected 1 warn message, got %d", len(warnMessages))
	}

	// Test Error
	logger.Error("error message")
	errorMessages := logger.GetErrorMessages()
	if len(errorMessages) != 1 {
		t.Errorf("Expected 1 error message, got %d", len(errorMessages))
	}

	// Test Fatal
	logger.Fatal("fatal message")
	fatalMessages := logger.GetFatalMessages()
	if len(fatalMessages) != 1 {
		t.Errorf("Expected 1 fatal message, got %d", len(fatalMessages))
	}
}

func TestMockLogger_PrintMethods(t *testing.T) {
	logger := NewMockLogger()

	// Test Println
	logger.Println("println message")
	messages := logger.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message after Println, got %d", len(messages))
	}

	// Test Printf
	logger.Printf("printf message %d", 42)
	messages = logger.GetMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages after Printf, got %d", len(messages))
	}
	if !strings.Contains(messages[1], "42") {
		t.Errorf("Expected printf message to contain '42', got: %s", messages[1])
	}
}

func TestMockLogger_LevelManagement(t *testing.T) {
	logger := NewMockLogger()

	// Test SetLevel
	logger.SetLevel(2) // Arbitrary level

	// Test GetLevel
	level := logger.GetLevel()
	if level != 2 {
		t.Errorf("Expected level 2, got %d", level)
	}
}

func TestMockLogger_MessageRetrieval(t *testing.T) {
	logger := NewMockLogger()

	// Add various types of messages
	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")
	logger.Fatal("fatal msg")

	// Test GetMessages (all messages)
	allMessages := logger.GetMessages()
	if len(allMessages) != 5 {
		t.Errorf("Expected 5 total messages, got %d", len(allMessages))
	}

	// Test individual message type retrieval
	debugMsgs := logger.GetDebugMessages()
	if len(debugMsgs) != 1 {
		t.Errorf("Expected 1 debug message, got %d", len(debugMsgs))
	}

	infoMsgs := logger.GetInfoMessages()
	if len(infoMsgs) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(infoMsgs))
	}

	warnMsgs := logger.GetWarnMessages()
	if len(warnMsgs) != 1 {
		t.Errorf("Expected 1 warn message, got %d", len(warnMsgs))
	}

	errorMsgs := logger.GetErrorMessages()
	if len(errorMsgs) != 1 {
		t.Errorf("Expected 1 error message, got %d", len(errorMsgs))
	}

	fatalMsgs := logger.GetFatalMessages()
	if len(fatalMsgs) != 1 {
		t.Errorf("Expected 1 fatal message, got %d", len(fatalMsgs))
	}
}

func TestMockLogger_ClearMessages(t *testing.T) {
	logger := NewMockLogger()

	// Add some messages
	logger.Info("message 1")
	logger.Error("message 2")

	messages := logger.GetMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages before clear, got %d", len(messages))
	}

	// Clear messages
	logger.ClearMessages()

	// Verify all message types are cleared
	if len(logger.GetMessages()) != 0 {
		t.Error("Expected no messages after clear")
	}
	if len(logger.GetDebugMessages()) != 0 {
		t.Error("Expected no debug messages after clear")
	}
	if len(logger.GetInfoMessages()) != 0 {
		t.Error("Expected no info messages after clear")
	}
	if len(logger.GetWarnMessages()) != 0 {
		t.Error("Expected no warn messages after clear")
	}
	if len(logger.GetErrorMessages()) != 0 {
		t.Error("Expected no error messages after clear")
	}
	if len(logger.GetFatalMessages()) != 0 {
		t.Error("Expected no fatal messages after clear")
	}
}

func TestMockLogger_ContainsMessage(t *testing.T) {
	logger := NewMockLogger()

	logger.Info("test message")
	logger.Error("error occurred")

	// Test ContainsMessage
	if !logger.ContainsMessage("test message") {
		t.Error("Expected ContainsMessage to find 'test message'")
	}

	if !logger.ContainsMessage("error occurred") {
		t.Error("Expected ContainsMessage to find 'error occurred'")
	}

	if logger.ContainsMessage("nonexistent message") {
		t.Error("Expected ContainsMessage to not find 'nonexistent message'")
	}
}

func TestMockLogger_VerboseMode(t *testing.T) {
	logger := NewMockLogger()

	// Test SetVerbose
	logger.SetVerbose(true)
	logger.Info("verbose message")

	logger.SetVerbose(false)
	logger.Info("non-verbose message")

	// Should capture messages regardless of verbose setting
	messages := logger.GetMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages regardless of verbose setting, got %d", len(messages))
	}
}

func TestMockLogger_ContextMethods(t *testing.T) {
	logger := NewMockLogger()
	ctx := context.Background()

	// Test DebugContext
	logger.DebugContext(ctx, "debug context message")
	debugMsgs := logger.GetDebugMessages()
	if len(debugMsgs) != 1 {
		t.Errorf("Expected 1 debug context message, got %d", len(debugMsgs))
	}

	// Test InfoContext
	logger.InfoContext(ctx, "info context message")
	infoMsgs := logger.GetInfoMessages()
	if len(infoMsgs) != 1 {
		t.Errorf("Expected 1 info context message, got %d", len(infoMsgs))
	}

	// Test WarnContext
	logger.WarnContext(ctx, "warn context message")
	warnMsgs := logger.GetWarnMessages()
	if len(warnMsgs) != 1 {
		t.Errorf("Expected 1 warn context message, got %d", len(warnMsgs))
	}

	// Test ErrorContext
	logger.ErrorContext(ctx, "error context message")
	errorMsgs := logger.GetErrorMessages()
	if len(errorMsgs) != 1 {
		t.Errorf("Expected 1 error context message, got %d", len(errorMsgs))
	}

	// Test FatalContext
	logger.FatalContext(ctx, "fatal context message")
	fatalMsgs := logger.GetFatalMessages()
	if len(fatalMsgs) != 1 {
		t.Errorf("Expected 1 fatal context message, got %d", len(fatalMsgs))
	}
}

func TestMockLogger_WithContext(t *testing.T) {
	logger := NewMockLogger()
	ctx := context.Background()

	// Test WithContext
	contextLogger := logger.WithContext(ctx)
	if contextLogger == nil {
		t.Error("Expected non-nil context logger")
	}

	// Should return the same logger instance
	if contextLogger != logger {
		t.Error("Expected WithContext to return the same logger instance")
	}
}

func TestMockLogger_FormattedMessages(t *testing.T) {
	logger := NewMockLogger()

	// Test formatted messages
	logger.Info("formatted message with %s and %d", "string", 123)
	infoMsgs := logger.GetInfoMessages()
	if len(infoMsgs) != 1 {
		t.Errorf("Expected 1 formatted info message, got %d", len(infoMsgs))
	}

	if !strings.Contains(infoMsgs[0], "string") || !strings.Contains(infoMsgs[0], "123") {
		t.Errorf("Expected formatted message to contain 'string' and '123', got: %s", infoMsgs[0])
	}

	// Test context formatted messages
	ctx := context.Background()
	logger.ErrorContext(ctx, "context error with %v", []int{1, 2, 3})
	errorMsgs := logger.GetErrorMessages()
	if len(errorMsgs) != 1 {
		t.Errorf("Expected 1 formatted error context message, got %d", len(errorMsgs))
	}
}

func TestMockLogger_EmptyState(t *testing.T) {
	logger := NewMockLogger()

	// Test empty state
	if len(logger.GetMessages()) != 0 {
		t.Error("Expected empty messages initially")
	}

	if logger.ContainsMessage("any message") {
		t.Error("Expected ContainsMessage to return false for empty logger")
	}

	// Test clearing empty logger
	logger.ClearMessages()
	if len(logger.GetMessages()) != 0 {
		t.Error("Expected messages to remain empty after clearing empty logger")
	}
}

func TestMockLogger_MessageOrdering(t *testing.T) {
	logger := NewMockLogger()

	// Add messages in specific order
	logger.Info("first")
	logger.Debug("second")
	logger.Error("third")

	messages := logger.GetMessages()
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}

	// Verify order is preserved
	if !strings.Contains(messages[0], "first") {
		t.Error("Expected first message to contain 'first'")
	}
	if !strings.Contains(messages[1], "second") {
		t.Error("Expected second message to contain 'second'")
	}
	if !strings.Contains(messages[2], "third") {
		t.Error("Expected third message to contain 'third'")
	}
}
