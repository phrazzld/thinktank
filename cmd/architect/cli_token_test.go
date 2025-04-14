// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// mockTokenManager is a simplified mock for testing CLI token management integration
type mockTokenManager struct {
	getTokenInfoCalled bool
	checkLimitCalled   bool
	promptCalled       bool
	confirmThreshold   int
	returnedPrompt     bool
}

func (m *mockTokenManager) GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error) {
	m.getTokenInfoCalled = true
	return &TokenResult{
		TokenCount:   1000,
		InputLimit:   2000,
		ExceedsLimit: false,
	}, nil
}

func (m *mockTokenManager) CheckTokenLimit(ctx context.Context, prompt string) error {
	m.checkLimitCalled = true
	return nil
}

func (m *mockTokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	m.promptCalled = true
	m.confirmThreshold = threshold
	return m.returnedPrompt
}

// mockClientFactory creates a fake client for testing
type mockClientFactory struct {
	createTokenManagerCalled bool
	createdTokenManager      TokenManager
}

func (f *mockClientFactory) CreateTokenManager(config *config.CliConfig) (TokenManager, error) {
	f.createTokenManagerCalled = true
	if f.createdTokenManager != nil {
		return f.createdTokenManager, nil
	}
	return nil, fmt.Errorf("no mock token manager configured")
}

// TestCLITokenManagement tests the integration between CLI and TokenManager
func TestCLITokenManagement(t *testing.T) {
	// Create a test function that simulates running the CLI with different confirm-tokens values
	testTokenConfirmation := func(confirmTokens int, mockPromptResponse bool) (*mockTokenManager, error) {
		// Create a mock token manager
		tokenManager := &mockTokenManager{
			returnedPrompt: mockPromptResponse,
		}

		// Create a mock client factory
		factory := &mockClientFactory{
			createdTokenManager: tokenManager,
		}

		// Create a CLI config with the specified confirm-tokens value
		cfg := &config.CliConfig{
			ConfirmTokens: confirmTokens,
			ModelNames:    []string{"test-model"},
			APIKey:        "test-key",
		}

		// Create a context
		ctx := context.Background()

		// Simulate running a CLI operation that would use the token manager
		// This is meant to be similar to how the real CLI would use these components
		_, err := runWithTokenManager(ctx, cfg, factory, "test prompt")

		return tokenManager, err
	}

	// Test with confirm-tokens = 0 (disabled)
	t.Run("ConfirmTokensDisabled", func(t *testing.T) {
		tokenManager, err := testTokenConfirmation(0, true)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify the token manager was used correctly
		if !tokenManager.checkLimitCalled {
			t.Error("Expected CheckTokenLimit to be called")
		}

		// PromptForConfirmation should be called even with threshold=0,
		// but it should return true without prompting
		if !tokenManager.promptCalled {
			t.Error("Expected PromptForConfirmation to be called")
		}

		// Verify the correct threshold was passed
		if tokenManager.confirmThreshold != 0 {
			t.Errorf("Expected threshold=0, got %d", tokenManager.confirmThreshold)
		}
	})

	// Test with confirm-tokens = 500 (enabled) and user confirms
	t.Run("ConfirmTokensEnabledUserConfirms", func(t *testing.T) {
		tokenManager, err := testTokenConfirmation(500, true)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify the token manager was used correctly
		if !tokenManager.checkLimitCalled {
			t.Error("Expected CheckTokenLimit to be called")
		}

		if !tokenManager.promptCalled {
			t.Error("Expected PromptForConfirmation to be called")
		}

		// Verify the correct threshold was passed
		if tokenManager.confirmThreshold != 500 {
			t.Errorf("Expected threshold=500, got %d", tokenManager.confirmThreshold)
		}
	})

	// Test with confirm-tokens = 500 (enabled) and user declines
	t.Run("ConfirmTokensEnabledUserDeclines", func(t *testing.T) {
		tokenManager, err := testTokenConfirmation(500, false)
		if err == nil {
			t.Fatal("Expected error when user declines confirmation, got nil")
		}

		// Verify error message
		if !strings.Contains(err.Error(), "aborted") {
			t.Errorf("Expected error to contain 'aborted', got: %v", err)
		}

		// Verify the token manager was used correctly
		if !tokenManager.checkLimitCalled {
			t.Error("Expected CheckTokenLimit to be called")
		}

		if !tokenManager.promptCalled {
			t.Error("Expected PromptForConfirmation to be called")
		}

		// Verify the correct threshold was passed
		if tokenManager.confirmThreshold != 500 {
			t.Errorf("Expected threshold=500, got %d", tokenManager.confirmThreshold)
		}
	})
}

// Test ParseFlagsWithEnv specifically for the confirm-tokens flag
func TestConfirmTokensFlag(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedValue    int
		expectedInConfig bool
	}{
		{
			name:             "Default value (0)",
			args:             []string{"--instructions", "instructions.md", "path1"},
			expectedValue:    0,
			expectedInConfig: true,
		},
		{
			name:             "Custom value",
			args:             []string{"--instructions", "instructions.md", "--confirm-tokens", "500", "path1"},
			expectedValue:    500,
			expectedInConfig: true,
		},
		{
			name:             "Zero value explicitly set",
			args:             []string{"--instructions", "instructions.md", "--confirm-tokens", "0", "path1"},
			expectedValue:    0,
			expectedInConfig: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new FlagSet for each test
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			// Disable output to avoid cluttering test output
			fs.SetOutput(io.Discard)

			// Mock environment getter
			getenv := func(key string) string {
				if key == apiKeyEnvVar {
					return "test-api-key"
				}
				return ""
			}

			// Parse flags
			cfg, err := ParseFlagsWithEnv(fs, tt.args, getenv)
			if err != nil {
				t.Fatalf("ParseFlagsWithEnv() error = %v", err)
			}

			// Verify confirm-tokens value was properly set in config
			if cfg.ConfirmTokens != tt.expectedValue {
				t.Errorf("ConfirmTokens = %d, want %d", cfg.ConfirmTokens, tt.expectedValue)
			}
		})
	}
}

// runWithTokenManager simulates a CLI operation that uses the token manager
// This is a simplified version of what might happen in the real CLI
func runWithTokenManager(ctx context.Context, cfg *config.CliConfig, factory *mockClientFactory, prompt string) (string, error) {
	// Create token manager
	tokenManager, err := factory.CreateTokenManager(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create token manager: %w", err)
	}

	// Check token limit
	if err := tokenManager.CheckTokenLimit(ctx, prompt); err != nil {
		return "", fmt.Errorf("token limit check failed: %w", err)
	}

	// Get token info for confirmation prompt
	tokenInfo, err := tokenManager.GetTokenInfo(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get token info: %w", err)
	}

	// Prompt for confirmation if needed
	confirmed := tokenManager.PromptForConfirmation(tokenInfo.TokenCount, cfg.ConfirmTokens)
	if !confirmed {
		return "", fmt.Errorf("operation aborted by user")
	}

	// Continue with operation (simplified for test)
	return "success", nil
}

// ClientFactory interface mocks how CLI would create token manager
type ClientFactory interface {
	CreateTokenManager(config *config.CliConfig) (TokenManager, error)
}

// realClientFactory would be the actual implementation in production code
type realClientFactory struct {
	logger logutil.LoggerInterface
}

func (f *realClientFactory) CreateTokenManager(cfg *config.CliConfig) (TokenManager, error) {
	// Create a context for client creation
	ctx := context.Background()

	// Create a Gemini client
	apiEndpoint := cfg.APIEndpoint
	if apiEndpoint == "" {
		apiEndpoint = "https://generativelanguage.googleapis.com" // Default endpoint
	}

	client, err := gemini.NewClient(ctx, cfg.APIKey, cfg.ModelNames[0], apiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Create token manager
	tokenManager, err := NewTokenManager(f.logger, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create token manager: %w", err)
	}

	return tokenManager, nil
}
