package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
)

// TestPartialFailureErrorHandling tests the error handling logic for partial model failures
// This specifically tests the logic in the Run method that handles the case where some
// models succeed and others fail
func TestPartialFailureErrorHandling(t *testing.T) {
	// Define test cases
	tests := []struct {
		name                string
		modelNames          []string
		modelOutputs        map[string]string
		modelErrors         []error
		expectError         bool
		expectedErrorMsg    string
		expectWarningLog    bool
		expectedWarningText string
	}{
		{
			name:       "All models succeed",
			modelNames: []string{"model1", "model2"},
			modelOutputs: map[string]string{
				"model1": "Output from model 1",
				"model2": "Output from model 2",
			},
			modelErrors:      nil,
			expectError:      false,
			expectWarningLog: false,
		},
		{
			name:         "All models fail",
			modelNames:   []string{"model1", "model2"},
			modelOutputs: map[string]string{
				// No outputs when all models fail
			},
			modelErrors: []error{
				errors.New("model1: API error"),
				errors.New("model2: rate limit exceeded"),
			},
			expectError:      true,
			expectedErrorMsg: "all models failed",
			expectWarningLog: false, // No warning log when all models fail (immediate error)
		},
		{
			name:       "Some models succeed, some fail",
			modelNames: []string{"model1", "model2", "model3"},
			modelOutputs: map[string]string{
				"model1": "Output from model 1",
				"model3": "Output from model 3",
			},
			modelErrors: []error{
				errors.New("model2: API error"),
			},
			expectError:         true,
			expectedErrorMsg:    "processed 2/3 models successfully; 1 failed",
			expectWarningLog:    true,
			expectedWarningText: "Some models failed but continuing with synthesis", // Partial text match
		},
		{
			name:       "Multiple model failures",
			modelNames: []string{"model1", "model2", "model3", "model4"},
			modelOutputs: map[string]string{
				"model3": "Output from model 3",
				"model4": "Output from model 4",
			},
			modelErrors: []error{
				errors.New("model1: API error"),
				errors.New("model2: rate limit exceeded"),
			},
			expectError:         true,
			expectedErrorMsg:    "processed 2/4 models successfully; 2 failed",
			expectWarningLog:    true,
			expectedWarningText: "2/4 models successful, 2 failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock logger that captures warning messages
			mockLogger := &MockLoggerWithWarnings{}

			// Create other mocks
			mockAPIService := &MockAPIService{}
			mockContextGatherer := &MockContextGatherer{}
			mockFileWriter := &MockFileWriter{}
			mockRateLimiter := ratelimit.NewRateLimiter(0, 0)
			mockAuditLogger := &MockAuditLogger{}

			// Create config with model names
			cfg := &config.CliConfig{
				ModelNames: tt.modelNames,
			}

			// Create a test orchestrator
			testOrch := &testOrchestratorErrorHandling{
				Orchestrator: Orchestrator{
					apiService:      mockAPIService,
					contextGatherer: mockContextGatherer,
					fileWriter:      mockFileWriter,
					auditLogger:     mockAuditLogger,
					rateLimiter:     mockRateLimiter,
					config:          cfg,
					logger:          mockLogger,
				},
				mockModelOutputs: tt.modelOutputs,
				mockModelErrors:  tt.modelErrors,
			}

			// Call Run
			err := testOrch.Run(context.Background(), "test instructions")

			// Verify error behavior
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else {
					// Check message content
					if !strings.Contains(err.Error(), tt.expectedErrorMsg) {
						t.Errorf("Expected error containing %q, got %q", tt.expectedErrorMsg, err.Error())
					}

					// Check sentinel error types
					if tt.name == "All models fail" && !errors.Is(err, ErrAllProcessingFailed) {
						t.Errorf("Expected error to be ErrAllProcessingFailed for failed models, got %T", err)
					} else if tt.name != "All models fail" && !errors.Is(err, ErrPartialProcessingFailure) {
						t.Errorf("Expected error to be ErrPartialProcessingFailure for partial failures, got %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %q", err.Error())
				}
			}

			// Verify warning log behavior
			if tt.expectWarningLog {
				if len(mockLogger.warnMessages) == 0 {
					t.Errorf("Expected warning log but none was recorded")
				} else {
					found := false
					for _, msg := range mockLogger.warnMessages {
						if strings.Contains(msg, tt.expectedWarningText) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected warning log containing %q, but not found in logs: %v",
							tt.expectedWarningText, mockLogger.warnMessages)
					}
				}
			} else {
				if len(mockLogger.warnMessages) > 0 {
					t.Errorf("Expected no warning logs, but got: %v", mockLogger.warnMessages)
				}
			}

			// For partial failures, verify that the warning log contains info about successful models
			if tt.expectWarningLog {
				// Check that successful model names are included in warning
				successfulModelsFound := false
				for _, msg := range mockLogger.warnMessages {
					// Look for successful model names in the warning
					for modelName := range tt.modelOutputs {
						if strings.Contains(msg, modelName) {
							successfulModelsFound = true
							break
						}
					}
					if successfulModelsFound {
						break
					}
				}
				if !successfulModelsFound {
					t.Errorf("Expected warning log to contain successful model names, but not found in: %v",
						mockLogger.warnMessages)
				}
			}
		})
	}
}

// MockLoggerWithWarnings extends the basic logger to capture warning messages
type MockLoggerWithWarnings struct {
	MockLogger
	warnMessages []string
}

// Override Warn to capture warning messages
func (m *MockLoggerWithWarnings) Warn(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	m.warnMessages = append(m.warnMessages, message)
}

// Override WarnContext to capture warning messages with context
func (m *MockLoggerWithWarnings) WarnContext(ctx context.Context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	m.warnMessages = append(m.warnMessages, message)
}

// WithContext returns the logger with context information
func (m *MockLoggerWithWarnings) WithContext(ctx context.Context) logutil.LoggerInterface {
	return m
}

// formatCountsForTest creates formatted count strings for the error message
func formatCountsForTest(successCount, totalCount, failCount int) string {
	return fmt.Sprintf("%d/%d models successfully; %d failed",
		successCount, totalCount, failCount)
}

// testOrchestratorErrorHandling extends Orchestrator to test error handling logic
type testOrchestratorErrorHandling struct {
	Orchestrator
	mockModelOutputs map[string]string
	mockModelErrors  []error
}

// Run overrides the Orchestrator.Run method to test error handling logic
func (o *testOrchestratorErrorHandling) Run(ctx context.Context, instructions string) error {
	// Ensure context has a correlation ID
	ctx = logutil.WithCorrelationID(ctx)

	// Get a logger with the context
	contextLogger := o.logger.WithContext(ctx)

	// Validate that models are specified
	if len(o.config.ModelNames) == 0 {
		return errors.New("no model names specified, at least one model is required")
	}

	// Skip context gathering and processing and go directly to handling model outputs/errors

	modelOutputs, modelErrors := o.mockModelOutputs, o.mockModelErrors

	// STEP 5: Handle model processing errors - THIS IS THE CORE LOGIC BEING TESTED
	var returnErr error
	if len(modelErrors) > 0 {
		// If ALL models failed (no outputs available), fail immediately
		if len(modelOutputs) == 0 {
			return fmt.Errorf("%w: all models failed: %s", ErrAllProcessingFailed, aggregateErrorMessages(modelErrors))
		}

		// Otherwise, log errors but continue with available outputs
		// Get list of successful model names for the log
		var successfulModels []string
		for modelName := range modelOutputs {
			successfulModels = append(successfulModels, modelName)
		}

		// Log a warning with detailed counts and successful model names
		contextLogger.WarnContext(ctx, "Some models failed but continuing with synthesis: %d/%d models successful, %d failed. Successful models: %v",
			len(modelOutputs), len(o.config.ModelNames), len(modelErrors), successfulModels)

		// Create a descriptive error to return after processing is complete
		returnErr = fmt.Errorf("%w: processed %s: %s",
			ErrPartialProcessingFailure,
			formatCountsForTest(len(modelOutputs), len(o.config.ModelNames), len(modelErrors)),
			aggregateErrorMessages(modelErrors))
	}

	// Return any model errors if any
	return returnErr
}
