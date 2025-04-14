package architect

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/gemini"
)

// TestTokenManagerPromptForConfirmation tests the user confirmation behavior directly
// without running the full application flow through architect.Execute.
func TestTokenManagerPromptForConfirmation(t *testing.T) {
	type confirmTestCase struct {
		name           string
		tokenCount     int32
		threshold      int
		userInput      string
		expected       bool
		expectedPrompt string
	}

	tests := []confirmTestCase{
		{
			name:           "Below Threshold - No Confirmation Needed",
			tokenCount:     500,
			threshold:      1000,
			userInput:      "", // No input needed
			expected:       true,
			expectedPrompt: "", // No prompt should be shown
		},
		{
			name:           "Threshold Disabled - No Confirmation Needed",
			tokenCount:     5000,
			threshold:      0,  // Disabled
			userInput:      "", // No input needed
			expected:       true,
			expectedPrompt: "", // No prompt should be shown
		},
		{
			name:           "Above Threshold - User Confirms with 'y'",
			tokenCount:     5000,
			threshold:      1000,
			userInput:      "y\n",
			expected:       true,
			expectedPrompt: "Token count",
		},
		{
			name:           "Above Threshold - User Rejects with 'n'",
			tokenCount:     5000,
			threshold:      1000,
			userInput:      "n\n",
			expected:       false,
			expectedPrompt: "Token count",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Save original stdin and restore it after the test
			origStdin := os.Stdin
			defer func() { os.Stdin = origStdin }()

			// Create a pipe to simulate user input
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			defer r.Close()
			defer w.Close()

			// Set stdin to the read end of the pipe
			os.Stdin = r

			// Write the test case's user input to the write end of the pipe
			if tc.userInput != "" {
				_, err = w.Write([]byte(tc.userInput))
				if err != nil {
					t.Fatalf("Failed to write to pipe: %v", err)
				}
			}

			// Create a mock logger to capture the logs
			logger := &mockContextLogger{}

			// Create a mock audit logger
			auditLogger := &mockAuditLogger{
				entries: make([]auditlog.AuditEntry, 0),
			}

			// Create a mock client
			client := &mockGeminiClient{
				getModelNameFunc: func() string {
					return "test-model"
				},
			}

			// Create the token manager
			tokenManager, err := NewTokenManager(logger, auditLogger, client)
			if err != nil {
				t.Fatalf("Failed to create token manager: %v", err)
			}

			// Call the method being tested directly
			result := tokenManager.PromptForConfirmation(tc.tokenCount, tc.threshold)

			// Verify the result
			if result != tc.expected {
				t.Errorf("PromptForConfirmation() = %v, want %v", result, tc.expected)
			}

			// If we expect a prompt, verify that it was shown
			if tc.expectedPrompt != "" {
				var foundPrompt bool
				for _, msg := range logger.infoMessages {
					if strings.Contains(msg, tc.expectedPrompt) {
						foundPrompt = true
						break
					}
				}
				if !foundPrompt {
					t.Errorf("Expected prompt containing %q, but it wasn't shown. Messages: %v", tc.expectedPrompt, logger.infoMessages)
				}
			} else {
				// If we don't expect a prompt, verify no token count messages were shown
				for _, msg := range logger.infoMessages {
					if strings.Contains(msg, "Token count") && strings.Contains(msg, "threshold") {
						t.Errorf("Unexpected prompt shown: %q", msg)
						break
					}
				}
			}
		})
	}
}

// TestTokenManagerGetTokenInfo tests the token counting and limit checking behavior
func TestTokenManagerGetTokenInfo(t *testing.T) {
	type tokenInfoTestCase struct {
		name              string
		prompt            string
		inputTokenLimit   int32
		responseTokens    int32
		expectedExceeds   bool
		expectedError     bool
		expectedErrorType string
	}

	tests := []tokenInfoTestCase{
		{
			name:            "Within Limit",
			prompt:          "This is a test prompt with a reasonable length.",
			inputTokenLimit: 100,
			responseTokens:  10,
			expectedExceeds: false,
			expectedError:   false,
		},
		{
			name:              "Exceeds Limit",
			prompt:            "This is a test prompt with a reasonable length that will exceed our artificial limit.",
			inputTokenLimit:   5,
			responseTokens:    20,
			expectedExceeds:   true,
			expectedError:     false,
			expectedErrorType: "TokenLimitExceededError",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock logger and audit logger
			logger := &mockContextLogger{}
			auditLogger := &mockAuditLogger{
				entries: make([]auditlog.AuditEntry, 0),
			}

			// Create a mock client
			client := &mockGeminiClient{
				getModelNameFunc: func() string {
					return "test-model"
				},
				countTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
					return &gemini.TokenCount{Total: tc.responseTokens}, nil
				},
				getModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
					return &gemini.ModelInfo{
						Name:             "test-model",
						InputTokenLimit:  tc.inputTokenLimit,
						OutputTokenLimit: tc.inputTokenLimit,
					}, nil
				},
			}

			// Create the token manager
			tokenManager, err := NewTokenManager(logger, auditLogger, client)
			if err != nil {
				t.Fatalf("Failed to create token manager: %v", err)
			}

			// Call the method being tested
			ctx := context.Background()
			tokenInfo, err := tokenManager.GetTokenInfo(ctx, tc.prompt)

			// Verify errors
			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
				return
			}
			if err != nil && !tc.expectedError {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify token info
			if tokenInfo.ExceedsLimit != tc.expectedExceeds {
				t.Errorf("ExceedsLimit = %v, want %v", tokenInfo.ExceedsLimit, tc.expectedExceeds)
			}

			// Verify token count
			if tokenInfo.TokenCount != tc.responseTokens {
				t.Errorf("TokenCount = %d, want %d", tokenInfo.TokenCount, tc.responseTokens)
			}

			// Verify input limit
			if tokenInfo.InputLimit != tc.inputTokenLimit {
				t.Errorf("InputLimit = %d, want %d", tokenInfo.InputLimit, tc.inputTokenLimit)
			}

			// If exceeding limit, verify error message
			if tc.expectedExceeds {
				if tokenInfo.LimitError == "" {
					t.Errorf("Expected a limit error message, but got empty string")
				}

				// Check audit log for the appropriate error entry
				if tc.expectedErrorType != "" {
					foundError := false
					for _, entry := range auditLogger.entries {
						if entry.Error != nil && entry.Error.Type == tc.expectedErrorType {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("Expected audit log entry with error type %s, but didn't find one", tc.expectedErrorType)
					}
				}
			}
		})
	}
}

// TestTokenManagerWithNilClient tests that a TokenManager cannot be created with a nil client
func TestTokenManagerWithNilClient(t *testing.T) {
	// Create a mock logger and audit logger
	logger := &mockContextLogger{}
	auditLogger := &mockAuditLogger{
		entries: make([]auditlog.AuditEntry, 0),
	}

	// Try to create the token manager with nil client
	tokenManager, err := NewTokenManager(logger, auditLogger, nil)

	// Verify error is returned
	if err == nil {
		t.Errorf("Expected error when creating TokenManager with nil client, got nil")
	}
	if tokenManager != nil {
		t.Errorf("Expected nil TokenManager, got %v", tokenManager)
	}
	if err != nil && !strings.Contains(err.Error(), "nil") {
		t.Errorf("Expected error to mention 'nil', got: %v", err)
	}
}

// TestTokenManagerCheckTokenLimit tests the CheckTokenLimit method of the TokenManager
func TestTokenManagerCheckTokenLimit(t *testing.T) {
	// Create a context for testing
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name            string
		prompt          string
		inputTokenLimit int32
		responseTokens  int32
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Under Limit",
			prompt:          "This is a test prompt.",
			inputTokenLimit: 100,
			responseTokens:  10,
			expectError:     false,
		},
		{
			name:            "Exceeds Limit",
			prompt:          "This is a test prompt that exceeds the limit.",
			inputTokenLimit: 5,
			responseTokens:  20,
			expectError:     true,
			errorContains:   "exceeds token limit",
		},
		{
			name:            "Error Getting Token Info",
			prompt:          "This is a test prompt.",
			inputTokenLimit: 0, // Will cause an error in the mock client
			responseTokens:  0, // Doesn't matter as we'll error first
			expectError:     true,
			errorContains:   "test error getting model info",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			logger := &mockContextLogger{}
			auditLogger := &mockAuditLogger{
				entries: make([]auditlog.AuditEntry, 0),
			}

			// Create a mock client
			mockClient := &mockGeminiClient{
				getModelNameFunc: func() string {
					return "test-model"
				},
				countTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
					return &gemini.TokenCount{Total: tc.responseTokens}, nil
				},
				getModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
					if tc.inputTokenLimit == 0 {
						// Simulate an error getting model info
						return nil, errors.New("test error getting model info")
					}
					return &gemini.ModelInfo{
						Name:             "test-model",
						InputTokenLimit:  tc.inputTokenLimit,
						OutputTokenLimit: tc.inputTokenLimit,
					}, nil
				},
			}

			// Create the token manager
			tokenManager, err := NewTokenManager(logger, auditLogger, mockClient)
			if err != nil {
				t.Fatalf("Failed to create token manager: %v", err)
			}

			// Call the CheckTokenLimit method
			err = tokenManager.CheckTokenLimit(ctx, tc.prompt)

			// Verify error expectations
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain %q, but got: %v", tc.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}
		})
	}
}
