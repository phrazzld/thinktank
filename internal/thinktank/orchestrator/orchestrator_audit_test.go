package orchestrator

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// This test file uses the MockAuditLogger defined in mocks_test.go

// TestLogAuditEvent tests the logAuditEvent helper method in various scenarios
func TestLogAuditEvent(t *testing.T) {
	tests := []struct {
		name            string
		op              string
		status          string
		inputs          map[string]interface{}
		outputs         map[string]interface{}
		err             error
		withCorrelation bool
		mockLogError    error
		expectWarnLog   bool
	}{
		{
			name:            "basic successful case",
			op:              "TestOperation",
			status:          "Success",
			inputs:          map[string]interface{}{"param1": "value1"},
			outputs:         map[string]interface{}{"result1": "output1"},
			err:             nil,
			withCorrelation: false,
			mockLogError:    nil,
			expectWarnLog:   false,
		},
		{
			name:            "with error parameter",
			op:              "ErrorOperation",
			status:          "Failure",
			inputs:          map[string]interface{}{"param1": "value1"},
			outputs:         map[string]interface{}{"result1": "output1"},
			err:             errors.New("test error"),
			withCorrelation: false,
			mockLogError:    nil,
			expectWarnLog:   false,
		},
		{
			name:            "with nil inputs",
			op:              "NilInputsOperation",
			status:          "Success",
			inputs:          nil,
			outputs:         map[string]interface{}{"result1": "output1"},
			err:             nil,
			withCorrelation: false,
			mockLogError:    nil,
			expectWarnLog:   false,
		},
		{
			name:            "with nil outputs",
			op:              "NilOutputsOperation",
			status:          "Success",
			inputs:          map[string]interface{}{"param1": "value1"},
			outputs:         nil,
			err:             nil,
			withCorrelation: false,
			mockLogError:    nil,
			expectWarnLog:   false,
		},
		{
			name:            "with correlation ID in context",
			op:              "CorrelationOperation",
			status:          "Success",
			inputs:          map[string]interface{}{"param1": "value1"},
			outputs:         map[string]interface{}{"result1": "output1"},
			err:             nil,
			withCorrelation: true,
			mockLogError:    nil,
			expectWarnLog:   false,
		},
		{
			name:            "with LogOp returning error",
			op:              "LogErrorOperation",
			status:          "Success",
			inputs:          map[string]interface{}{"param1": "value1"},
			outputs:         map[string]interface{}{"result1": "output1"},
			err:             nil,
			withCorrelation: false,
			mockLogError:    errors.New("log write error"),
			expectWarnLog:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test dependencies
			mockLogger := fileutil.NewMockLogger()
			mockAuditLogger := NewMockAuditLogger()
			mockAuditLogger.LogError = tt.mockLogError

			// Create minimal orchestrator instance with just the required dependencies
			o := &Orchestrator{
				logger:      mockLogger,
				auditLogger: mockAuditLogger,
				config:      &config.CliConfig{},
			}

			// Create context with correlation ID if needed
			var ctx context.Context
			if tt.withCorrelation {
				ctx = logutil.WithCustomCorrelationID(context.Background(), "test-correlation-id")
			} else {
				ctx = context.Background()
			}

			// Call the method under test
			o.logAuditEvent(ctx, tt.op, tt.status, tt.inputs, tt.outputs, tt.err)

			// Verify the audit logger was called correctly
			if len(mockAuditLogger.LogCalls) != 1 {
				t.Errorf("Expected 1 call to LogOp, got %d", len(mockAuditLogger.LogCalls))
				return
			}

			// Get the recorded call
			call := mockAuditLogger.LogCalls[0]

			// Verify operation and status were passed correctly
			if call.Operation != tt.op {
				t.Errorf("Expected operation %q, got %q", tt.op, call.Operation)
			}
			if call.Status != tt.status {
				t.Errorf("Expected status %q, got %q", tt.status, call.Status)
			}

			// Verify error was passed correctly
			if (call.Error != nil && tt.err == nil) || (call.Error == nil && tt.err != nil) ||
				(call.Error != nil && tt.err != nil && call.Error.Error() != tt.err.Error()) {
				t.Errorf("Expected error %v, got %v", tt.err, call.Error)
			}

			// Verify inputs map is never nil
			if call.Inputs == nil {
				t.Error("Expected inputs map to never be nil")
			}

			// Verify outputs map is never nil
			if call.Outputs == nil {
				t.Error("Expected outputs map to never be nil")
			}

			// Verify correlation ID was added if present in context
			if tt.withCorrelation {
				correlationID, found := call.Inputs["correlation_id"]
				if !found {
					t.Error("Expected correlation_id in inputs, but was not found")
				} else if correlationID != "test-correlation-id" {
					t.Errorf("Expected correlation_id %q, got %q", "test-correlation-id", correlationID)
				}
			}

			// For nil inputs/outputs cases, verify they were replaced with empty maps
			if tt.inputs == nil {
				// Ensure inputs is an empty map (except possibly correlation ID)
				if len(call.Inputs) > 1 || (len(call.Inputs) == 1 && !tt.withCorrelation) {
					t.Errorf("Expected empty inputs map, got %v", call.Inputs)
				}
			} else {
				// For non-nil inputs, verify all values were passed correctly
				// Make a copy of the input map to compare
				expectedInputs := make(map[string]interface{})
				for k, v := range tt.inputs {
					expectedInputs[k] = v
				}

				// If we expect correlation ID, add it to expected inputs
				if tt.withCorrelation {
					expectedInputs["correlation_id"] = "test-correlation-id"
				}

				if !reflect.DeepEqual(call.Inputs, expectedInputs) {
					t.Errorf("Expected inputs %v, got %v", expectedInputs, call.Inputs)
				}
			}

			if tt.outputs == nil {
				// Ensure outputs is an empty map
				if len(call.Outputs) > 0 {
					t.Errorf("Expected empty outputs map, got %v", call.Outputs)
				}
			} else {
				// For non-nil outputs, verify all values were passed correctly
				if !reflect.DeepEqual(call.Outputs, tt.outputs) {
					t.Errorf("Expected outputs %v, got %v", tt.outputs, call.Outputs)
				}
			}

			// Verify warn log was emitted if LogOp returned an error
			if tt.expectWarnLog {
				warnMsgs := mockLogger.GetWarnMessages()
				if len(warnMsgs) == 0 {
					t.Error("Expected warning log when LogOp returns error, but none found")
				} else {
					foundLogError := false
					for _, msg := range warnMsgs {
						if (tt.mockLogError != nil) && (msg != "") && contain(msg, "Failed to write audit log") {
							foundLogError = true
							break
						}
					}
					if !foundLogError {
						t.Error("Expected warning log about audit log failure, but none found")
					}
				}
			}
		})
	}
}

// copyMapExceptCorrelationID is not needed anymore

// Helper function to check if a string contains a substring
func contain(s, substr string) bool {
	return s != "" && substr != "" && strings.Contains(s, substr)
}
