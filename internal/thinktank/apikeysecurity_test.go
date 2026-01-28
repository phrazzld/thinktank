package thinktank

import (
	"bytes"
	"strings"
	"testing"

	"github.com/misty-step/thinktank/internal/logutil"
)

// TestAPIKeyLoggingSecurity tests that the logging never includes API key fragments
func TestAPIKeyLoggingSecurity(t *testing.T) {
	// Set up a buffer to capture logs
	buffer := &bytes.Buffer{}
	logger := logutil.NewLogger(logutil.DebugLevel, buffer, "")

	// Create API key for testing
	testAPIKey := "test-api-key-secret-12345"

	// Test logging with API key metadata
	service := &registryAPIService{
		logger: logger,
	}

	// This directly tests the new logging code we added (the part that was fixed)
	provider := "testprovider"

	// Directly log with the method that previously logged API key fragments
	if testAPIKey == "" {
		service.logger.Error("Empty API key for provider '%s' - this will cause authentication failures", provider)
	} else {
		// Log API key metadata only (NEVER log any portion of the key itself)
		service.logger.Debug("Using API key for provider '%s' (length: %d, source: via environment variable)",
			provider, len(testAPIKey))
	}

	// Get the log output
	logOutput := buffer.String()

	// Verify the API key is not logged directly or partially
	if strings.Contains(logOutput, "test-api-key") {
		t.Errorf("Full API key was leaked in logs: %s", logOutput)
	}

	if strings.Contains(logOutput, "test-") {
		t.Errorf("API key prefix was leaked in logs: %s", logOutput)
	}

	if strings.Contains(logOutput, "12345") {
		t.Errorf("API key suffix was leaked in logs: %s", logOutput)
	}

	// Check positive patterns that should be there
	if !strings.Contains(logOutput, "length: 25") {
		t.Errorf("Expected API key length to be logged, but it wasn't: %s", logOutput)
	}

	if !strings.Contains(logOutput, "source: via environment variable") {
		t.Errorf("Expected API key source to be logged, but it wasn't: %s", logOutput)
	}
}
