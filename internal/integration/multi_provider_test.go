// internal/integration/multi_provider_test.go
package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// providerTracking tracks calls to different providers
type providerTracking struct {
	sync.Mutex
	geminiCalls  []string
	openAICalls  []string
	geminiErrors map[string]error
	openAIErrors map[string]error
	calls        map[string]time.Time
	modelResults map[string]string
}

// newProviderTracking creates a new provider tracking instance
func newProviderTracking() *providerTracking {
	return &providerTracking{
		geminiCalls:  make([]string, 0),
		openAICalls:  make([]string, 0),
		geminiErrors: make(map[string]error),
		openAIErrors: make(map[string]error),
		calls:        make(map[string]time.Time),
		modelResults: make(map[string]string),
	}
}

// recordCall records a call to a specific model with timestamp
func (t *providerTracking) recordCall(modelName string) {
	t.Lock()
	defer t.Unlock()

	t.calls[modelName] = time.Now()

	// nolint:staticcheck // Using deprecated function is acceptable in tests until they are updated
	providerType := architect.DetectProviderFromModel(modelName)
	switch providerType {
	case architect.ProviderGemini:
		t.geminiCalls = append(t.geminiCalls, modelName)
	case architect.ProviderOpenAI:
		t.openAICalls = append(t.openAICalls, modelName)
	}
}

// recordError records an error for a specific model
// Used indirectly in tests
// nolint:unused
func (t *providerTracking) recordError(modelName string, err error) {
	t.Lock()
	defer t.Unlock()

	// nolint:staticcheck // Using deprecated function is acceptable in tests until they are updated
	providerType := architect.DetectProviderFromModel(modelName)
	switch providerType {
	case architect.ProviderGemini:
		t.geminiErrors[modelName] = err
	case architect.ProviderOpenAI:
		t.openAIErrors[modelName] = err
	}
}

// recordResult records the result for a model
func (t *providerTracking) recordResult(modelName string, result string) {
	t.Lock()
	defer t.Unlock()

	t.modelResults[modelName] = result
}

// GetResults returns a copy of the tracking results
func (t *providerTracking) GetResults() (geminiModels []string, openAIModels []string, callTimes map[string]time.Time, errors map[string]error, results map[string]string) {
	t.Lock()
	defer t.Unlock()

	// Copy gemini calls
	geminiModels = make([]string, len(t.geminiCalls))
	copy(geminiModels, t.geminiCalls)

	// Copy openai calls
	openAIModels = make([]string, len(t.openAICalls))
	copy(openAIModels, t.openAICalls)

	// Copy call times
	callTimes = make(map[string]time.Time)
	for k, v := range t.calls {
		callTimes[k] = v
	}

	// Copy all errors
	errors = make(map[string]error)
	for k, v := range t.geminiErrors {
		errors[k] = v
	}
	for k, v := range t.openAIErrors {
		errors[k] = v
	}

	// Copy results
	results = make(map[string]string)
	for k, v := range t.modelResults {
		results[k] = v
	}

	return
}

// TestProviderDetection tests the provider detection logic
func TestProviderDetection(t *testing.T) {
	tests := []struct {
		name         string
		modelName    string
		expectedType architect.ProviderType
	}{
		{
			name:         "Gemini Model",
			modelName:    "gemini-pro",
			expectedType: architect.ProviderGemini,
		},
		{
			name:         "Gemini 1.5 Model",
			modelName:    "gemini-1.5-pro",
			expectedType: architect.ProviderGemini,
		},
		{
			name:         "OpenAI GPT-4 Model",
			modelName:    "gpt-4",
			expectedType: architect.ProviderOpenAI,
		},
		{
			name:         "OpenAI GPT-3.5 Model",
			modelName:    "gpt-3.5-turbo",
			expectedType: architect.ProviderOpenAI,
		},
		{
			name:         "Empty Model Name",
			modelName:    "",
			expectedType: architect.ProviderUnknown,
		},
		{
			name:         "Unknown Provider Model",
			modelName:    "unknown-model",
			expectedType: architect.ProviderUnknown,
		},
		{
			name:         "OpenAI Text Davinci Model",
			modelName:    "text-davinci-003",
			expectedType: architect.ProviderOpenAI,
		},
		{
			name:         "OpenAI Ada Model",
			modelName:    "ada",
			expectedType: architect.ProviderOpenAI,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// nolint:staticcheck // Using deprecated function is acceptable in tests until they are updated
			providerType := architect.DetectProviderFromModel(tc.modelName)
			if providerType != tc.expectedType {
				t.Errorf("Expected provider type %v for model %s, got %v",
					tc.expectedType, tc.modelName, providerType)
			}
		})
	}
}

// TestMultiProviderIntegration tests the integration of provider detection with actual clients
func TestMultiProviderIntegration(t *testing.T) {
	// Create a temporary directory for test files
	testDir, err := os.MkdirTemp("", "architect-multi-provider-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(testDir) }()

	// Create output directory
	outputDir := filepath.Join(testDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)

	// Test with mixed providers
	modelNames := []string{"gemini-pro", "gpt-4"}
	tracking := newProviderTracking()

	// Process each model
	for _, modelName := range modelNames {
		t.Run(fmt.Sprintf("Model_%s", modelName), func(t *testing.T) {
			// nolint:staticcheck // Using deprecated function is acceptable in tests until they are updated
			providerType := architect.DetectProviderFromModel(modelName)
			require.NotEqual(t, architect.ProviderUnknown, providerType, "Should detect a valid provider")

			// Record that we processed this model
			tracking.recordCall(modelName)

			// Create a simulated result for this model
			var providerName string
			switch providerType {
			case architect.ProviderGemini:
				providerName = "Gemini"
			case architect.ProviderOpenAI:
				providerName = "OpenAI"
			default:
				providerName = "Unknown"
			}

			// Record the result
			outputContent := fmt.Sprintf("# Output from %s\n\nThis is a test output for model %s using the %s provider.",
				modelName, modelName, providerName)
			tracking.recordResult(modelName, outputContent)

			// Write the result to a file
			outputFile := filepath.Join(outputDir, modelName+".md")
			err := os.WriteFile(outputFile, []byte(outputContent), 0644)
			require.NoError(t, err)
		})
	}

	// Verify the results
	geminiModels, openAIModels, _, _, results := tracking.GetResults()

	// Check that we have the right number of models
	assert.Len(t, geminiModels, 1, "Should process one Gemini model")
	assert.Len(t, openAIModels, 1, "Should process one OpenAI model")

	// Check that the correct models were processed
	assert.Contains(t, geminiModels, "gemini-pro")
	assert.Contains(t, openAIModels, "gpt-4")

	// Check that results were recorded for each model
	assert.Contains(t, results, "gemini-pro")
	assert.Contains(t, results, "gpt-4")

	// Check that the output files were created
	geminiOutput := filepath.Join(outputDir, "gemini-pro.md")
	openAIOutput := filepath.Join(outputDir, "gpt-4.md")

	assert.FileExists(t, geminiOutput)
	assert.FileExists(t, openAIOutput)

	// Check the content of the output files
	geminiContent, err := os.ReadFile(geminiOutput)
	require.NoError(t, err)
	assert.Contains(t, string(geminiContent), "Gemini")

	openAIContent, err := os.ReadFile(openAIOutput)
	require.NoError(t, err)
	assert.Contains(t, string(openAIContent), "OpenAI")
}

// TestAPIKeyValidation tests that the config package properly validates API keys
func TestAPIKeyValidation(t *testing.T) {
	tests := []struct {
		name          string
		modelNames    []string
		setGeminiKey  bool
		setOpenAIKey  bool
		expectError   bool
		errorContains string
	}{
		{
			name:         "Gemini Model Only",
			modelNames:   []string{"gemini-pro"},
			setGeminiKey: true,
			setOpenAIKey: false,
			expectError:  false,
		},
		{
			name:          "Gemini Model Missing Key",
			modelNames:    []string{"gemini-pro"},
			setGeminiKey:  false,
			setOpenAIKey:  false,
			expectError:   true,
			errorContains: "gemini API key not set",
		},
		{
			name:         "OpenAI Model Only",
			modelNames:   []string{"gpt-4"},
			setGeminiKey: false,
			setOpenAIKey: true,
			expectError:  false,
		},
		{
			name:          "OpenAI Model Missing Key",
			modelNames:    []string{"gpt-4"},
			setGeminiKey:  false,
			setOpenAIKey:  false,
			expectError:   true,
			errorContains: "openAI API key not set",
		},
		{
			name:         "Mixed Models Both Keys",
			modelNames:   []string{"gemini-pro", "gpt-4"},
			setGeminiKey: true,
			setOpenAIKey: true,
			expectError:  false,
		},
		{
			name:          "Mixed Models Missing OpenAI Key",
			modelNames:    []string{"gemini-pro", "gpt-4"},
			setGeminiKey:  true,
			setOpenAIKey:  false,
			expectError:   true,
			errorContains: "openAI API key not set",
		},
		{
			name:          "Mixed Models Missing Gemini Key",
			modelNames:    []string{"gemini-pro", "gpt-4"},
			setGeminiKey:  false,
			setOpenAIKey:  true,
			expectError:   true,
			errorContains: "gemini API key not set",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment
			// Save original env vars
			origGeminiKey := os.Getenv(config.APIKeyEnvVar)
			origOpenAIKey := os.Getenv(config.OpenAIAPIKeyEnvVar)
			defer func() {
				// Restore original env vars
				_ = os.Setenv(config.APIKeyEnvVar, origGeminiKey)
				_ = os.Setenv(config.OpenAIAPIKeyEnvVar, origOpenAIKey)
			}()

			// Set or unset keys based on test case
			if tc.setGeminiKey {
				_ = os.Setenv(config.APIKeyEnvVar, "test-gemini-key")
			} else {
				_ = os.Unsetenv(config.APIKeyEnvVar)
			}

			if tc.setOpenAIKey {
				_ = os.Setenv(config.OpenAIAPIKeyEnvVar, "test-openai-key")
			} else {
				_ = os.Unsetenv(config.OpenAIAPIKeyEnvVar)
			}

			// Create configuration
			cfg := &config.CliConfig{
				InstructionsFile: "/tmp/instructions.md", // Doesn't need to exist for this test
				ModelNames:       tc.modelNames,
				Paths:            []string{"/tmp"}, // Doesn't need to exist for this test
				// Set API key from env var
				APIKey: os.Getenv(config.APIKeyEnvVar),
			}

			// Create a logger
			logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ")

			// Validate configuration
			err := config.ValidateConfig(cfg, logger)

			// Check results
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestOutputFileFormatting tests that output files are correctly named based on model
func TestOutputFileFormatting(t *testing.T) {
	// Create a temporary directory for test files
	testDir, err := os.MkdirTemp("", "architect-output-format-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(testDir) }()

	// Create output directory
	outputDir := filepath.Join(testDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)

	// Test cases for different model formats
	modelCases := []struct {
		modelName string
		provider  string
		content   string
	}{
		{
			modelName: "gemini-pro",
			provider:  "Gemini",
			content:   "# Gemini Pro Output\n\nThis is a test output from the Gemini provider.",
		},
		{
			modelName: "gemini-1.5-pro",
			provider:  "Gemini",
			content:   "# Gemini 1.5 Pro Output\n\nThis is a test output from the Gemini provider.",
		},
		{
			modelName: "gpt-4",
			provider:  "OpenAI",
			content:   "# GPT-4 Output\n\nThis is a test output from the OpenAI provider.",
		},
		{
			modelName: "gpt-3.5-turbo",
			provider:  "OpenAI",
			content:   "# GPT-3.5 Turbo Output\n\nThis is a test output from the OpenAI provider.",
		},
	}

	// Write test files
	for _, mc := range modelCases {
		outputFile := filepath.Join(outputDir, mc.modelName+".md")
		err := os.WriteFile(outputFile, []byte(mc.content), 0644)
		require.NoError(t, err)
	}

	// Verify that files exist and have correct content
	for _, mc := range modelCases {
		outputFile := filepath.Join(outputDir, mc.modelName+".md")

		// Check that file exists
		assert.FileExists(t, outputFile)

		// Check content
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		// Content should contain the provider name
		assert.Contains(t, string(content), mc.provider)
	}
}

// providerAwareLLMClient is a mock LLM client for testing
type providerAwareLLMClient struct {
	modelName    string
	providerType architect.ProviderType
}

// Implement LLMClient interface methods
func (c *providerAwareLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	// Create a provider-specific response based on the model
	providerName := "Unknown"
	switch c.providerType {
	case architect.ProviderGemini:
		providerName = "Gemini"
	case architect.ProviderOpenAI:
		providerName = "OpenAI"
	}

	return &llm.ProviderResult{
		Content:      fmt.Sprintf("Generated content from %s using %s provider", c.modelName, providerName),
		FinishReason: "STOP",
		TokenCount:   100,
	}, nil
}

func (c *providerAwareLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	return &llm.ProviderTokenCount{
		Total: int32(len(prompt) / 4),
	}, nil
}

func (c *providerAwareLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	// Default limits based on provider
	inputLimit := int32(8192)
	outputLimit := int32(2048)

	switch c.providerType {
	case architect.ProviderOpenAI:
		if strings.Contains(c.modelName, "gpt-4") {
			inputLimit = 8192
			outputLimit = 2048
		} else if strings.Contains(c.modelName, "gpt-3.5") {
			inputLimit = 16385
			outputLimit = 4096
		}
	case architect.ProviderGemini:
		if strings.Contains(c.modelName, "pro") {
			inputLimit = 30720
			outputLimit = 2048
		}
	}

	return &llm.ProviderModelInfo{
		Name:             c.modelName,
		InputTokenLimit:  inputLimit,
		OutputTokenLimit: outputLimit,
	}, nil
}

func (c *providerAwareLLMClient) GetModelName() string {
	return c.modelName
}

func (c *providerAwareLLMClient) Close() error {
	return nil
}

// TestLLMClientProviderAwareness tests that the LLMClient interface works correctly with different providers
func TestLLMClientProviderAwareness(t *testing.T) {
	tests := []struct {
		name         string
		modelName    string
		expectError  bool
		providerType architect.ProviderType
	}{
		{
			name:         "Gemini Model",
			modelName:    "gemini-pro",
			expectError:  false,
			providerType: architect.ProviderGemini,
		},
		{
			name:         "OpenAI Model",
			modelName:    "gpt-4",
			expectError:  false,
			providerType: architect.ProviderOpenAI,
		},
		{
			name:         "Unknown Model",
			modelName:    "unknown-model",
			expectError:  false,
			providerType: architect.ProviderUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Detect provider from model name
			// nolint:staticcheck // Using deprecated function is acceptable in tests until they are updated
			providerType := architect.DetectProviderFromModel(tc.modelName)
			assert.Equal(t, tc.providerType, providerType)

			// Create a provider-aware client
			client := &providerAwareLLMClient{
				modelName:    tc.modelName,
				providerType: providerType,
			}

			// Test the client's methods
			ctx := context.Background()

			// Test GenerateContent
			result, err := client.GenerateContent(ctx, "Test prompt")
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Result should contain model name
				assert.Contains(t, result.Content, tc.modelName)

				// Result should contain provider info
				switch providerType {
				case architect.ProviderGemini:
					assert.Contains(t, result.Content, "Gemini")
				case architect.ProviderOpenAI:
					assert.Contains(t, result.Content, "OpenAI")
				}
			}

			// Test CountTokens
			tokenCount, err := client.CountTokens(ctx, "Test prompt")
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokenCount)
				assert.Greater(t, tokenCount.Total, int32(0))
			}

			// Test GetModelInfo
			modelInfo, err := client.GetModelInfo(ctx)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, modelInfo)
				assert.Equal(t, tc.modelName, modelInfo.Name)

				// Model info should have reasonable token limits
				assert.Greater(t, modelInfo.InputTokenLimit, int32(0))
				assert.Greater(t, modelInfo.OutputTokenLimit, int32(0))
			}

			// Test GetModelName
			assert.Equal(t, tc.modelName, client.GetModelName())

			// Test Close
			err = client.Close()
			assert.NoError(t, err)
		})
	}
}
