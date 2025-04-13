package architect

import (
	"context"
	"errors"
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
	logger := &mockLogger{}
	mockClient := &gemini.MockClient{}

	tokenManager, err := NewTokenManager(logger, mockClient)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Test threshold = 0 (disabled)
	t.Run("ThresholdDisabled", func(t *testing.T) {
		result := tokenManager.PromptForConfirmation(5000, 0)
		if !result {
			t.Error("Expected true when threshold is disabled (0), got false")
		}
	})

	// Test threshold > tokenCount (no confirmation needed)
	t.Run("ThresholdNotExceeded", func(t *testing.T) {
		result := tokenManager.PromptForConfirmation(5000, 6000)
		if !result {
			t.Error("Expected true when token count is below threshold, got false")
		}
	})

	// Note: We can't easily test the interactive prompt behavior in a unit test
	// since it depends on terminal input. In a real-world scenario, this would be
	// better tested with a mock reader that can be injected for testing.
}
