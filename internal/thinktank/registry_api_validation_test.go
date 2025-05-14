// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/testutil"
)

// TestValidateModelParameter tests the ValidateModelParameter method
func TestValidateModelParameter(t *testing.T) {
	// Define test cases for ValidateModelParameter
	testCases := []struct {
		name           string
		modelName      string
		paramName      string
		value          interface{}
		modelDef       *registry.ModelDefinition
		getModelErr    error
		registryImpl   interface{}
		expectValid    bool
		expectError    bool
		errorSubstring string
	}{
		// Common failure cases
		{
			name:           "registry does not implement GetModel",
			modelName:      "test-model",
			paramName:      "temperature",
			value:          0.7,
			registryImpl:   "not a registry", // String instead of proper mock registry
			expectValid:    false,
			expectError:    true,
			errorSubstring: "does not implement GetModel method",
		},
		{
			name:           "model not found",
			modelName:      "non-existent-model",
			paramName:      "temperature",
			value:          0.7,
			getModelErr:    llm.Wrap(llm.ErrModelNotFound, "", "model 'non-existent-model' not found in registry", llm.CategoryNotFound),
			expectValid:    false,
			expectError:    true,
			errorSubstring: "model 'non-existent-model' not found",
		},
		{
			name:           "parameter not found",
			modelName:      "test-model",
			paramName:      "non-existent-param",
			value:          0.7,
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'non-existent-param' not defined for model",
		},

		// Float parameter tests
		{
			name:        "float parameter valid",
			modelName:   "test-model",
			paramName:   "temperature",
			value:       0.7,
			expectValid: true,
			expectError: false,
		},
		{
			name:           "float parameter invalid type (string)",
			modelName:      "test-model",
			paramName:      "temperature",
			value:          "not a float",
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'temperature' must be a float",
		},
		{
			name:           "float parameter invalid type (int)",
			modelName:      "test-model",
			paramName:      "temperature",
			value:          5,
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'temperature' must be a float",
		},
		{
			name:           "float parameter below minimum",
			modelName:      "test-model",
			paramName:      "temperature",
			value:          -0.1, // Min is 0.0
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'temperature' value -0.10 is below minimum 0.00",
		},
		{
			name:           "float parameter above maximum",
			modelName:      "test-model",
			paramName:      "temperature",
			value:          1.5, // Max is 1.0
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'temperature' value 1.50 exceeds maximum 1.00",
		},
		{
			name:      "float parameter with nil min/max",
			modelName: "test-model-nil-limits",
			paramName: "temperature_no_limits",
			value:     100.5,
			modelDef: &registry.ModelDefinition{
				Name:       "test-model-nil-limits",
				Provider:   "test-provider",
				APIModelID: "test-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"temperature_no_limits": {
						Type:    "float",
						Default: 0.7,
						Min:     nil,
						Max:     nil,
					},
				},
			},
			expectValid: true,
			expectError: false,
		},

		// Integer parameter tests
		{
			name:        "int parameter valid",
			modelName:   "test-model",
			paramName:   "max_tokens",
			value:       2048,
			expectValid: true,
			expectError: false,
		},
		{
			name:           "int parameter invalid type (string)",
			modelName:      "test-model",
			paramName:      "max_tokens",
			value:          "not an int",
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'max_tokens' must be an integer",
		},
		{
			name:           "int parameter invalid type (float)",
			modelName:      "test-model",
			paramName:      "max_tokens",
			value:          1024.5,
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'max_tokens' must be an integer",
		},
		{
			name:           "int parameter below minimum",
			modelName:      "test-model",
			paramName:      "max_tokens",
			value:          0, // Min is 1
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'max_tokens' value 0 is below minimum 1",
		},
		{
			name:           "int parameter above maximum",
			modelName:      "test-model",
			paramName:      "max_tokens",
			value:          5000, // Max is 4096
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'max_tokens' value 5000 exceeds maximum 4096",
		},
		{
			name:      "int parameter with nil min/max",
			modelName: "test-model-nil-limits",
			paramName: "tokens_no_limits",
			value:     10000,
			modelDef: &registry.ModelDefinition{
				Name:       "test-model-nil-limits",
				Provider:   "test-provider",
				APIModelID: "test-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"tokens_no_limits": {
						Type:    "int",
						Default: 1024,
						Min:     nil,
						Max:     nil,
					},
				},
			},
			expectValid: true,
			expectError: false,
		},

		// String parameter tests
		{
			name:        "string parameter valid enum value",
			modelName:   "test-model",
			paramName:   "model_type",
			value:       "creative",
			expectValid: true,
			expectError: false,
		},
		{
			name:           "string parameter invalid type (int)",
			modelName:      "test-model",
			paramName:      "model_type",
			value:          42,
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'model_type' must be a string",
		},
		{
			name:           "string parameter invalid type (float)",
			modelName:      "test-model",
			paramName:      "model_type",
			value:          3.14,
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'model_type' must be a string",
		},
		{
			name:           "string parameter not in enum",
			modelName:      "test-model",
			paramName:      "model_type",
			value:          "invalid-type",
			expectValid:    false,
			expectError:    true,
			errorSubstring: "parameter 'model_type' value 'invalid-type' is not in allowed values",
		},
		{
			name:      "string parameter with no enum values",
			modelName: "test-model-no-enum",
			paramName: "free_text",
			value:     "any string works",
			modelDef: &registry.ModelDefinition{
				Name:       "test-model-no-enum",
				Provider:   "test-provider",
				APIModelID: "test-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"free_text": {
						Type:       "string",
						Default:    "",
						EnumValues: []string{},
					},
				},
			},
			expectValid: true,
			expectError: false,
		},

		// Edge cases
		{
			name:      "unknown parameter type",
			modelName: "test-model-unknown-type",
			paramName: "weird_param",
			value:     "anything",
			modelDef: &registry.ModelDefinition{
				Name:       "test-model-unknown-type",
				Provider:   "test-provider",
				APIModelID: "test-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"weird_param": {
						Type:    "unknown_type",
						Default: nil,
					},
				},
			},
			expectValid: true, // Unknown types should be accepted (only logged as warning)
			expectError: false,
		},
		{
			name:      "invalid min type for float",
			modelName: "test-model-invalid-min",
			paramName: "float_bad_min",
			value:     0.5,
			modelDef: &registry.ModelDefinition{
				Name:       "test-model-invalid-min",
				Provider:   "test-provider",
				APIModelID: "test-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"float_bad_min": {
						Type:    "float",
						Default: 0.5,
						Min:     "not a number", // Invalid min type
						Max:     1.0,
					},
				},
			},
			expectValid: true, // Should ignore invalid min/max types
			expectError: false,
		},
		{
			name:      "invalid max type for int",
			modelName: "test-model-invalid-max",
			paramName: "int_bad_max",
			value:     100,
			modelDef: &registry.ModelDefinition{
				Name:       "test-model-invalid-max",
				Provider:   "test-provider",
				APIModelID: "test-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"int_bad_max": {
						Type:    "int",
						Default: 100,
						Min:     1,
						Max:     "not a number", // Invalid max type
					},
				},
			},
			expectValid: true, // Should ignore invalid min/max types
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment
			var service *registryAPIService
			var mockRegistry *MockRegistryAPI

			if tc.registryImpl == nil {
				service, mockRegistry, _ = setupTest(t)

				// Add custom model definition if provided
				if tc.modelDef != nil {
					mockRegistry.models[tc.modelName] = tc.modelDef
				}

				// Set get model error if provided
				mockRegistry.getModelErr = tc.getModelErr
			} else {
				// Use the provided registry implementation
				logger := testutil.NewMockLogger()
				service = &registryAPIService{
					registry: tc.registryImpl,
					logger:   logger,
				}
			}

			// Create context for tests
			ctx := context.Background()

			// Call the method being tested
			valid, err := service.ValidateModelParameter(ctx, tc.modelName, tc.paramName, tc.value)

			// Verify expected valid flag
			if valid != tc.expectValid {
				t.Errorf("Expected valid=%v, got %v", tc.expectValid, valid)
			}

			// Verify expected error behavior
			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tc.errorSubstring)
				}
				if !strings.Contains(err.Error(), tc.errorSubstring) {
					t.Errorf("Expected error containing '%s', got '%v'", tc.errorSubstring, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
			}
		})
	}
}
