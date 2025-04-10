package architect

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// mockLogger for testing
type mockAPILogger struct {
	logutil.LoggerInterface
	debugMessages []string
	infoMessages  []string
	errorMessages []string
}

func (m *mockAPILogger) Debug(format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, format)
}

func (m *mockAPILogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *mockAPILogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, format)
}

// mockAPIService is a test implementation of APIService
type mockAPIService struct {
	logger     logutil.LoggerInterface
	initError  error
	mockClient gemini.Client
}

func newMockAPIService(logger logutil.LoggerInterface, initError error, mockClient gemini.Client) APIService {
	return &mockAPIService{
		logger:     logger,
		initError:  initError,
		mockClient: mockClient,
	}
}

func (m *mockAPIService) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	if m.initError != nil {
		return nil, m.initError
	}
	return m.mockClient, nil
}

func (m *mockAPIService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	return "", errors.New("not implemented in mock")
}

// TestAPIServiceImplementation tests that apiService properly implements APIService
func TestAPIServiceImplementation(t *testing.T) {
	ctx := context.Background()
	logger := &mockAPILogger{}
	apiService := NewAPIService(logger)

	// Test with invalid credentials (this doesn't actually call the real API)
	t.Run("InvalidCredentials", func(t *testing.T) {
		// Using empty strings to simulate invalid credentials
		client, err := apiService.InitClient(ctx, "", "")
		
		// We expect this to fail, but we can't guarantee exactly how it fails
		// because we're not mocking gemini.NewClient
		if err == nil {
			t.Error("Expected error with empty credentials, got nil")
		}
		
		if client != nil {
			t.Error("Expected nil client with error, got non-nil client")
		}
	})
}