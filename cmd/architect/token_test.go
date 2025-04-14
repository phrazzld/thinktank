package architect

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

type mockLogger struct {
	logutil.LoggerInterface
	debugCalled bool
	debugArgs   []interface{}
}

func (m *mockLogger) Debug(format string, args ...interface{}) {
	m.debugCalled = true
	m.debugArgs = args
}

func (m *mockLogger) Info(format string, args ...interface{}) {
	// No-op for testing
}

func (m *mockLogger) Error(format string, args ...interface{}) {
	// No-op for testing
}

// mockInfoLogger extends mockLogger to track Info and Error calls
type mockInfoLogger struct {
	mockLogger
	infoCalled  bool
	infoMsg     string
	errorCalled bool
	errorMsg    string
}

func (m *mockInfoLogger) Info(format string, args ...interface{}) {
	m.infoCalled = true
	m.infoMsg = format
}

func (m *mockInfoLogger) Error(format string, args ...interface{}) {
	m.errorCalled = true
	m.errorMsg = format
}

// Reset clears the state of the logger for a new test
func (m *mockInfoLogger) Reset() {
	m.debugCalled = false
	m.debugArgs = nil
	m.infoCalled = false
	m.infoMsg = ""
	m.errorCalled = false
	m.errorMsg = ""
}

func TestGetTokenInfo(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	// Test normal operation
	t.Run("NormalOperation", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: 1000}, nil
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		tokenManager, err := NewTokenManager(logger, mockClient)
		if err != nil {
			t.Fatalf("Failed to create token manager: %v", err)
		}

		info, err := tokenManager.GetTokenInfo(ctx, "test prompt")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if info == nil {
			t.Fatal("Expected token info, got nil")
		}
		if info.TokenCount != 1000 {
			t.Errorf("Expected TokenCount=1000, got %d", info.TokenCount)
		}
		if info.InputLimit != 2000 {
			t.Errorf("Expected InputLimit=2000, got %d", info.InputLimit)
		}
		if info.ExceedsLimit {
			t.Error("Expected ExceedsLimit=false, got true")
		}
		if !logger.debugCalled {
			t.Error("Expected logger.Debug to be called, but it wasn't")
		}
	})

	// Test exceeding limit
	t.Run("ExceedsLimit", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: 3000}, nil
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		tokenManager, err := NewTokenManager(logger, mockClient)
		if err != nil {
			t.Fatalf("Failed to create token manager: %v", err)
		}

		info, err := tokenManager.GetTokenInfo(ctx, "test prompt")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if info == nil {
			t.Fatal("Expected token info, got nil")
		}
		if info.TokenCount != 3000 {
			t.Errorf("Expected TokenCount=3000, got %d", info.TokenCount)
		}
		if info.InputLimit != 2000 {
			t.Errorf("Expected InputLimit=2000, got %d", info.InputLimit)
		}
		if !info.ExceedsLimit {
			t.Error("Expected ExceedsLimit=true, got false")
		}
		if info.LimitError == "" {
			t.Error("Expected non-empty LimitError")
		}
	})

	// Test error in GetModelInfo
	t.Run("ErrorInGetModelInfo", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return nil, errors.New("model info error")
			},
		}

		tokenManager, err := NewTokenManager(logger, mockClient)
		if err != nil {
			t.Fatalf("Failed to create token manager: %v", err)
		}

		_, err = tokenManager.GetTokenInfo(ctx, "test prompt")
		if err == nil {
			t.Error("Expected error when GetModelInfo fails, got nil")
		}
	})

	// Test error in CountTokens
	t.Run("ErrorInCountTokens", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return nil, errors.New("count tokens error")
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		tokenManager, err := NewTokenManager(logger, mockClient)
		if err != nil {
			t.Fatalf("Failed to create token manager: %v", err)
		}

		_, err = tokenManager.GetTokenInfo(ctx, "test prompt")
		if err == nil {
			t.Error("Expected error when CountTokens fails, got nil")
		}
	})

	// Test API error in GetModelInfo
	t.Run("APIErrorInGetModelInfo", func(t *testing.T) {
		apiErr := &gemini.APIError{
			StatusCode: 400,
			Message:    "API error",
		}
		mockClient := &gemini.MockClient{
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return nil, apiErr
			},
		}

		tokenManager, err := NewTokenManager(logger, mockClient)
		if err != nil {
			t.Fatalf("Failed to create token manager: %v", err)
		}

		_, err = tokenManager.GetTokenInfo(ctx, "test prompt")
		if err != apiErr {
			t.Errorf("Expected API error to be passed through, got: %v", err)
		}
	})

	// Test nil client
	t.Run("NilClient", func(t *testing.T) {
		_, err := NewTokenManager(logger, nil)
		if err == nil {
			t.Error("Expected error when client is nil, got nil")
		}
	})
}

func TestCheckTokenLimit(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	// Test token count within limits
	t.Run("TokenCountWithinLimits", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: 1000}, nil
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		tokenManager, err := NewTokenManager(logger, mockClient)
		if err != nil {
			t.Fatalf("Failed to create token manager: %v", err)
		}

		err = tokenManager.CheckTokenLimit(ctx, "test prompt")
		if err != nil {
			t.Errorf("Expected no error for token count within limits, got: %v", err)
		}
	})

	// Test token count exceeds limits
	t.Run("TokenCountExceedsLimits", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: 3000}, nil
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		tokenManager, err := NewTokenManager(logger, mockClient)
		if err != nil {
			t.Fatalf("Failed to create token manager: %v", err)
		}

		err = tokenManager.CheckTokenLimit(ctx, "test prompt")
		if err == nil {
			t.Error("Expected error for token count exceeding limits, got nil")
		}
	})

	// Test error getting model info
	t.Run("ErrorGettingModelInfo", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return nil, errors.New("model info error")
			},
		}

		tokenManager, err := NewTokenManager(logger, mockClient)
		if err != nil {
			t.Fatalf("Failed to create token manager: %v", err)
		}

		err = tokenManager.CheckTokenLimit(ctx, "test prompt")
		if err == nil {
			t.Error("Expected error when getting model info fails, got nil")
		}
	})

	// Test error counting tokens
	t.Run("ErrorCountingTokens", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return nil, errors.New("token counting error")
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		tokenManager, err := NewTokenManager(logger, mockClient)
		if err != nil {
			t.Fatalf("Failed to create token manager: %v", err)
		}

		err = tokenManager.CheckTokenLimit(ctx, "test prompt")
		if err == nil {
			t.Error("Expected error when counting tokens fails, got nil")
		}
	})
}

func TestPromptForConfirmation(t *testing.T) {
	// Create an enhanced mock logger that captures info messages
	logger := &mockInfoLogger{}
	mockClient := &gemini.MockClient{}

	tokenManager, err := NewTokenManager(logger, mockClient)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Test threshold = 0 (disabled)
	t.Run("ThresholdDisabled", func(t *testing.T) {
		logger.Reset()
		result := tokenManager.PromptForConfirmation(5000, 0)
		if !result {
			t.Error("Expected true when threshold is disabled (0), got false")
		}
		// Verify no info logs were made since no prompt was needed
		if logger.infoCalled {
			t.Error("Expected no info logs for disabled threshold")
		}
	})

	// Test threshold > tokenCount (no confirmation needed)
	t.Run("ThresholdNotExceeded", func(t *testing.T) {
		logger.Reset()
		result := tokenManager.PromptForConfirmation(5000, 6000)
		if !result {
			t.Error("Expected true when token count is below threshold, got false")
		}
		// Verify no info logs were made since no prompt was needed
		if logger.infoCalled {
			t.Error("Expected no info logs when below threshold")
		}
	})

	// Test confirmation prompt with "yes" response
	t.Run("ConfirmationPromptWithYes", func(t *testing.T) {
		// Save and restore stdin
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		// Create a pipe to simulate user input
		r, w, _ := os.Pipe()
		os.Stdin = r

		// Write "yes" to the input
		go func() {
			_, _ = w.Write([]byte("yes\n"))
			_ = w.Close()
		}()

		logger.Reset()
		result := tokenManager.PromptForConfirmation(5000, 4000)

		// Should return true for "yes"
		if !result {
			t.Error("Expected true for 'yes' response, got false")
		}

		// Verify the info logs were called for the prompt
		if !logger.infoCalled {
			t.Error("Expected info logs for confirmation prompt")
		}
	})

	// Test confirmation prompt with "y" response
	t.Run("ConfirmationPromptWithY", func(t *testing.T) {
		// Save and restore stdin
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		// Create a pipe to simulate user input
		r, w, _ := os.Pipe()
		os.Stdin = r

		// Write "y" to the input
		go func() {
			_, _ = w.Write([]byte("y\n"))
			_ = w.Close()
		}()

		logger.Reset()
		result := tokenManager.PromptForConfirmation(5000, 4000)

		// Should return true for "y"
		if !result {
			t.Error("Expected true for 'y' response, got false")
		}
	})

	// Test confirmation prompt with "no" response
	t.Run("ConfirmationPromptWithNo", func(t *testing.T) {
		// Save and restore stdin
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		// Create a pipe to simulate user input
		r, w, _ := os.Pipe()
		os.Stdin = r

		// Write "no" to the input
		go func() {
			_, _ = w.Write([]byte("no\n"))
			_ = w.Close()
		}()

		logger.Reset()
		result := tokenManager.PromptForConfirmation(5000, 4000)

		// Should return false for "no"
		if result {
			t.Error("Expected false for 'no' response, got true")
		}
	})

	// Test confirmation prompt with empty response (default is "no")
	t.Run("ConfirmationPromptWithEmptyResponse", func(t *testing.T) {
		// Save and restore stdin
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		// Create a pipe to simulate user input
		r, w, _ := os.Pipe()
		os.Stdin = r

		// Write just a newline (empty response)
		go func() {
			_, _ = w.Write([]byte("\n"))
			_ = w.Close()
		}()

		logger.Reset()
		result := tokenManager.PromptForConfirmation(5000, 4000)

		// Should return false for empty response (default no)
		if result {
			t.Error("Expected false for empty response, got true")
		}
	})

	// Test error reading from stdin
	t.Run("ErrorReadingFromStdin", func(t *testing.T) {
		// Save and restore stdin
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		// Create a closed pipe to simulate read error
		r, w, _ := os.Pipe()
		_ = w.Close() // Close the write end immediately to cause read error
		os.Stdin = r

		logger.Reset()
		result := tokenManager.PromptForConfirmation(5000, 4000)

		// Should return false on read error
		if result {
			t.Error("Expected false on stdin read error, got true")
		}

		// Verify the error was logged
		if !logger.errorCalled {
			t.Error("Expected error to be logged on stdin read error")
		}
	})
}
