// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestCorrelationIDPropagation verifies that correlation IDs are properly propagated
// across all components during a multi-component operation.
//
// This test executes an operation spanning multiple components:
// - Context gathering (ContextGatherer)
// - API service calls (APIService)
// - File writing (FileWriter)
// - Audit logging (AuditLogger)
//
// It injects a specific correlation ID into the initial context and verifies
// that every log entry contains the correct correlation ID.
func TestCorrelationIDPropagation(t *testing.T) {
	// Define a specific correlation ID for testing
	testCorrelationID := "test-correlation-id-123"

	IntegrationTestWithBoundaries(t, func(env *BoundaryTestEnv) {
		// === SETUP TEST SCENARIO ===

		// Simple test with one model to focus on correlation ID propagation
		modelNames := []string{"model1"}
		instructions := "Test correlation ID propagation across components"

		// Setup source files for context gathering
		setupSourceFiles(t, env)

		// Setup mock response
		expectedResponse := "# Test Output\n\nThis is a test response for correlation ID verification."
		env.SetupModelResponse("model1", expectedResponse)

		// Configure test environment
		env.SetupModels(modelNames, "") // No synthesis model to keep it simple
		instructionsPath := env.SetupInstructionsFile(instructions)
		env.Config.InstructionsFile = instructionsPath

		// === INJECT CORRELATION ID ===

		// Create context with specific correlation ID
		ctx := logutil.WithCorrelationID(context.Background(), testCorrelationID)

		// Verify the correlation ID is set correctly
		actualCorrelationID := logutil.GetCorrelationID(ctx)
		if actualCorrelationID != testCorrelationID {
			t.Fatalf("Expected correlation ID %s, got %s", testCorrelationID, actualCorrelationID)
		}

		// === EXECUTE MULTI-COMPONENT OPERATION ===

		err := env.Run(ctx, instructions)
		if err != nil {
			t.Fatalf("Multi-component operation failed: %v", err)
		}

		// === VERIFY CORRELATION ID PROPAGATION ===

		// Get the test logger from the environment
		testLogger, ok := env.Logger.(*logutil.TestLogger)
		if !ok {
			t.Fatalf("Expected TestLogger, got %T", env.Logger)
		}

		// Retrieve all captured log entries
		capturedLogs := testLogger.GetTestLogs()

		// Verify we have log entries (the operation should have generated logs)
		if len(capturedLogs) == 0 {
			t.Fatal("Expected log entries to be captured, but found none")
		}

		// Track missing correlation IDs for detailed reporting
		var logsWithoutCorrelationID []string
		var logsWithWrongCorrelationID []string

		// Verify every log entry contains the correct correlation ID
		for _, logEntry := range capturedLogs {
			expectedPattern := "[correlation_id=" + testCorrelationID + "]"

			if !strings.Contains(logEntry, expectedPattern) {
				// Check if it has any correlation ID
				if strings.Contains(logEntry, "[correlation_id=") {
					logsWithWrongCorrelationID = append(logsWithWrongCorrelationID, logEntry)
				} else {
					logsWithoutCorrelationID = append(logsWithoutCorrelationID, logEntry)
				}
			}
		}

		// Report findings
		totalLogs := len(capturedLogs)
		logsWithCorrectID := totalLogs - len(logsWithoutCorrelationID) - len(logsWithWrongCorrelationID)

		t.Logf("Correlation ID Propagation Summary:")
		t.Logf("  Total log entries: %d", totalLogs)
		t.Logf("  Logs with correct correlation ID: %d", logsWithCorrectID)
		t.Logf("  Logs missing correlation ID: %d", len(logsWithoutCorrelationID))
		t.Logf("  Logs with wrong correlation ID: %d", len(logsWithWrongCorrelationID))

		// Fail the test if any logs are missing the correlation ID
		if len(logsWithoutCorrelationID) > 0 {
			t.Errorf("Found %d log entries without correlation ID:", len(logsWithoutCorrelationID))
			for i, logEntry := range logsWithoutCorrelationID {
				if i < 5 { // Limit output to first 5 for readability
					t.Errorf("  [%d] %s", i+1, logEntry)
				}
			}
			if len(logsWithoutCorrelationID) > 5 {
				t.Errorf("  ... and %d more entries", len(logsWithoutCorrelationID)-5)
			}
		}

		// Fail the test if any logs have wrong correlation ID
		if len(logsWithWrongCorrelationID) > 0 {
			t.Errorf("Found %d log entries with wrong correlation ID:", len(logsWithWrongCorrelationID))
			for i, logEntry := range logsWithWrongCorrelationID {
				if i < 5 { // Limit output to first 5 for readability
					t.Errorf("  [%d] %s", i+1, logEntry)
				}
			}
			if len(logsWithWrongCorrelationID) > 5 {
				t.Errorf("  ... and %d more entries", len(logsWithWrongCorrelationID)-5)
			}
		}

		// Only pass if all logs have the correct correlation ID
		if len(logsWithoutCorrelationID) == 0 && len(logsWithWrongCorrelationID) == 0 {
			t.Logf("✅ Correlation ID propagation verified successfully")
			t.Logf("   All %d log entries contain the correct correlation ID: %s", totalLogs, testCorrelationID)
		}
	})
}

// TestCorrelationIDPropagationWithSynthesis verifies correlation ID propagation
// in a more complex scenario involving synthesis.
func TestCorrelationIDPropagationWithSynthesis(t *testing.T) {
	// Define a specific correlation ID for testing
	testCorrelationID := "test-synthesis-correlation-456"

	IntegrationTestWithBoundaries(t, func(env *BoundaryTestEnv) {
		// === SETUP COMPLEX TEST SCENARIO ===

		// Multiple models and synthesis for more complex workflow
		modelNames := []string{"model1", "model2"}
		synthesisModel := "synthesis-model"
		instructions := "Test correlation ID propagation in synthesis workflow"

		// Setup source files
		setupSourceFiles(t, env)

		// Setup mock responses
		env.SetupModelResponse("model1", "# Output 1\nTest output from model 1.")
		env.SetupModelResponse("model2", "# Output 2\nTest output from model 2.")
		env.SetupModelResponse(synthesisModel, "# Synthesis\nCombined insights from models.")

		// Configure test environment
		env.SetupModels(modelNames, synthesisModel)
		instructionsPath := env.SetupInstructionsFile(instructions)
		env.Config.InstructionsFile = instructionsPath

		// === INJECT CORRELATION ID ===

		ctx := logutil.WithCorrelationID(context.Background(), testCorrelationID)

		// === EXECUTE COMPLEX MULTI-COMPONENT OPERATION ===

		err := env.Run(ctx, instructions)
		if err != nil {
			t.Fatalf("Complex multi-component operation failed: %v", err)
		}

		// === VERIFY CORRELATION ID PROPAGATION ===

		testLogger := env.Logger.(*logutil.TestLogger)
		capturedLogs := testLogger.GetTestLogs()

		if len(capturedLogs) == 0 {
			t.Fatal("Expected log entries to be captured, but found none")
		}

		// Count logs with correct correlation ID
		var correctCorrelationIDCount int
		expectedPattern := "[correlation_id=" + testCorrelationID + "]"

		for _, logEntry := range capturedLogs {
			if strings.Contains(logEntry, expectedPattern) {
				correctCorrelationIDCount++
			}
		}

		// Report results
		totalLogs := len(capturedLogs)
		t.Logf("Synthesis Workflow Correlation ID Summary:")
		t.Logf("  Total log entries: %d", totalLogs)
		t.Logf("  Logs with correct correlation ID: %d", correctCorrelationIDCount)

		// Expect high percentage of logs to have correlation ID
		// Allow some tolerance for logs that may not have context propagation
		expectedMinPercentage := 0.8 // 80% minimum
		actualPercentage := float64(correctCorrelationIDCount) / float64(totalLogs)

		if actualPercentage < expectedMinPercentage {
			t.Errorf("Correlation ID propagation below threshold: %.1f%% < %.1f%%",
				actualPercentage*100, expectedMinPercentage*100)

			// Show some examples of logs without correlation ID
			t.Logf("Sample logs missing correlation ID:")
			count := 0
			for _, logEntry := range capturedLogs {
				if !strings.Contains(logEntry, expectedPattern) && count < 3 {
					t.Logf("  %s", logEntry)
					count++
				}
			}
		} else {
			t.Logf("✅ Synthesis workflow correlation ID propagation verified")
			t.Logf("   %.1f%% of log entries contain the correct correlation ID", actualPercentage*100)
		}
	})
}
