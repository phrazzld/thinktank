package thinktank

import (
	"context"
	"testing"

	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank/tokenizers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenCountingService_AccuracyComparison(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	// Test with substantial content to see differences
	instructions := "Create a comprehensive analysis of this Go codebase. Focus on architecture patterns, error handling, testing strategies, and performance considerations. Provide detailed recommendations for improvements."

	files := []FileContent{
		{
			Path: "main.go",
			Content: `package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	port   int
	router *http.ServeMux
}

func NewServer(port int) *Server {
	return &Server{
		port:   port,
		router: http.NewServeMux(),
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.setupRoutes()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return server.ListenAndServe()
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/health", s.healthCheck)
	s.router.HandleFunc("/api/data", s.handleData)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleData(w http.ResponseWriter, r *http.Request) {
	// Implementation here
}

func main() {
	server := NewServer(8080)
	if err := server.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}`,
		},
		{
			Path: "handler.go",
			Content: `package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type DataResponse struct {
	Message string ` + "`json:\"message\"`" + `
	Data    []string ` + "`json:\"data\"`" + `
}

func processRequest(r *http.Request) (*DataResponse, error) {
	// Validate request
	if r.Method != http.MethodGet {
		return nil, fmt.Errorf("method not allowed")
	}

	// Process data
	data := []string{"item1", "item2", "item3"}

	return &DataResponse{
		Message: "Success",
		Data:    data,
	}, nil
}`,
		},
	}

	// Test with OpenAI model (should use tiktoken)
	openaiResult, err := service.CountTokensForModel(context.Background(), TokenCountingRequest{
		Instructions: instructions,
		Files:        files,
	}, "gpt-4")

	require.NoError(t, err)
	assert.Greater(t, openaiResult.TotalTokens, 0)

	// Test with Gemini model (should use SentencePiece)
	geminiResult, err := service.CountTokensForModel(context.Background(), TokenCountingRequest{
		Instructions: instructions,
		Files:        files,
	}, "gemini-2.5-pro")

	require.NoError(t, err)
	assert.Greater(t, geminiResult.TotalTokens, 0)

	// The results might be different due to different tokenizers
	// but both should be reasonable estimates
	t.Logf("OpenAI (tiktoken) tokens: %d", openaiResult.TotalTokens)
	t.Logf("Gemini (SentencePiece) tokens: %d", geminiResult.TotalTokens)

	// Both should be in reasonable range for this content
	assert.GreaterOrEqual(t, openaiResult.TotalTokens, 100, "Should count substantial tokens for OpenAI")
	assert.GreaterOrEqual(t, geminiResult.TotalTokens, 100, "Should count substantial tokens for Gemini")
}

func TestTokenCountingService_GetCompatibleModels_EmptyInput(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	models, err := service.GetCompatibleModels(context.Background(), TokenCountingRequest{
		Instructions: "",
		Files:        []FileContent{},
	}, []string{"openai", "gemini"})

	require.NoError(t, err)
	assert.NotEmpty(t, models, "Should return some models even for empty input")
}

func TestTokenCountingService_GetCompatibleModels_SingleCompatibleModel(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	models, err := service.GetCompatibleModels(context.Background(), TokenCountingRequest{
		Instructions: "Short instruction",
		Files: []FileContent{
			{Path: "test.txt", Content: "Small content"},
		},
	}, []string{"openai"})

	require.NoError(t, err)
	assert.NotEmpty(t, models, "Should return compatible models")

	// Check that we have at least one compatible model
	hasCompatible := false
	for _, model := range models {
		if model.IsCompatible {
			hasCompatible = true
			assert.Greater(t, model.ContextWindow, 0, "Compatible model should have context window")
			assert.GreaterOrEqual(t, model.TokenCount, 0, "Should have token count")
			break
		}
	}
	assert.True(t, hasCompatible, "Should have at least one compatible model")
}

func TestTokenCountingService_GetCompatibleModels_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		instructions       string
		files              []FileContent
		availableProviders []string
		expectCompatible   bool
		expectTokenCount   bool
	}{
		{
			name:               "Empty input with OpenAI",
			instructions:       "",
			files:              []FileContent{},
			availableProviders: []string{"openai"},
			expectCompatible:   true,
			expectTokenCount:   false,
		},
		{
			name:         "Small input with multiple providers",
			instructions: "Analyze this code",
			files: []FileContent{
				{Path: "test.go", Content: "package main\nfunc main() {}"},
			},
			availableProviders: []string{"openai", "gemini"},
			expectCompatible:   true,
			expectTokenCount:   true,
		},
		{
			name:         "Large input with Gemini",
			instructions: "Process this large dataset",
			files: func() []FileContent {
				// Create larger content
				content := ""
				for i := 0; i < 100; i++ {
					content += "This is line " + string(rune(i)) + " with some content that adds up.\n"
				}
				return []FileContent{
					{Path: "large.txt", Content: content},
				}
			}(),
			availableProviders: []string{"gemini"},
			expectCompatible:   true,
			expectTokenCount:   true,
		},
		{
			name:         "Multiple files with multiple providers",
			instructions: "Refactor these files for better performance",
			files: []FileContent{
				{Path: "server.go", Content: "package main\n\ntype Server struct {}"},
				{Path: "client.go", Content: "package main\n\ntype Client struct {}"},
				{Path: "utils.go", Content: "package main\n\nfunc Helper() string { return \"help\" }"},
			},
			availableProviders: []string{"openai", "gemini"},
			expectCompatible:   true,
			expectTokenCount:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := NewTokenCountingService()

			models, err := service.GetCompatibleModels(context.Background(), TokenCountingRequest{
				Instructions: tt.instructions,
				Files:        tt.files,
			}, tt.availableProviders)

			require.NoError(t, err)
			assert.NotEmpty(t, models, "Should return model compatibility information")

			if tt.expectCompatible {
				hasCompatible := false
				for _, model := range models {
					if model.IsCompatible {
						hasCompatible = true
						assert.Greater(t, model.ContextWindow, 0, "Compatible model should have context window")

						if tt.expectTokenCount {
							assert.Greater(t, model.TokenCount, 0, "Should count tokens for non-empty input")
						} else {
							assert.GreaterOrEqual(t, model.TokenCount, 0, "Token count should be non-negative")
						}
						break
					}
				}
				assert.True(t, hasCompatible, "Should have at least one compatible model")
			}
		})
	}
}

func TestNewTokenCountingServiceWithManager(t *testing.T) {
	t.Parallel()

	mockManager := &MockAccurateTokenCounterManager{}
	mockLogger := &testutil.MockLogger{}

	service := NewTokenCountingServiceWithManagerAndLogger(mockManager, mockLogger)

	// Test that the service uses the provided manager
	result, err := service.CountTokens(context.Background(), TokenCountingRequest{
		Instructions: "Test",
		Files:        []FileContent{},
	})

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Mock implementations for testing

type MockAccurateTokenCounterManager struct{}

func (m *MockAccurateTokenCounterManager) GetTokenizer(provider string) (tokenizers.AccurateTokenCounter, error) {
	return &MockAccurateTokenCounter{}, nil
}

func (m *MockAccurateTokenCounterManager) SupportsProvider(provider string) bool {
	return true
}

func (m *MockAccurateTokenCounterManager) ClearCache() {
	// No-op for mock
}

// MockAccurateTokenCounter is defined in token_counting_basic_test.go
