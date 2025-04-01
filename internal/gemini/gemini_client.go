// internal/gemini/gemini_client.go
package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/phrazzld/architect/internal/logutil"

	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// geminiClient implements the Client interface using Google's genai SDK
type geminiClient struct {
	client    *genai.Client
	model     *genai.GenerativeModel
	modelName string
	apiKey    string
	logger    logutil.LoggerInterface

	// Model info caching
	modelInfoCache map[string]*ModelInfo
	modelInfoMutex sync.RWMutex
	httpClient     *http.Client
}

// newGeminiClient creates a new Gemini client with Google's genai SDK
func newGeminiClient(ctx context.Context, apiKey, modelName string) (Client, error) {
	if apiKey == "" {
		return nil, errors.New("API key cannot be empty")
	}

	if modelName == "" {
		return nil, errors.New("model name cannot be empty")
	}

	// Create standard logger for internal client use
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[gemini] ", true)

	// Initialize the Google genai client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Initialize the model with default settings
	model := client.GenerativeModel(modelName)
	config := DefaultModelConfig()
	model.SetMaxOutputTokens(config.MaxOutputTokens)
	model.SetTemperature(config.Temperature)
	model.SetTopP(config.TopP)

	return &geminiClient{
		client:         client,
		model:          model,
		modelName:      modelName,
		apiKey:         apiKey,
		logger:         logger,
		modelInfoCache: make(map[string]*ModelInfo),
		modelInfoMutex: sync.RWMutex{},
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// GenerateContent sends a text prompt to Gemini and returns the generated content
func (c *geminiClient) GenerateContent(ctx context.Context, prompt string) (*GenerationResult, error) {
	if prompt == "" {
		return nil, errors.New("prompt cannot be empty")
	}

	// Generate content
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Check for empty response
	if resp == nil {
		return nil, errors.New("received nil response from Gemini API")
	}

	// Check for empty candidates
	if len(resp.Candidates) == 0 {
		return nil, errors.New("received empty candidates from Gemini API")
	}

	candidate := resp.Candidates[0]

	// Check for empty content
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return &GenerationResult{
			Content:       "",
			FinishReason:  string(candidate.FinishReason),
			SafetyRatings: mapSafetyRatings(candidate.SafetyRatings),
			Truncated:     candidate.FinishReason == genai.FinishReasonMaxTokens,
		}, nil
	}

	// Extract text content
	var contentBuilder strings.Builder
	for _, part := range candidate.Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			contentBuilder.WriteString(string(textPart))
		}
	}

	// Get token usage if available
	var tokenCount int32
	if resp.UsageMetadata != nil {
		tokenCount = resp.UsageMetadata.TotalTokenCount
	}

	// Build result
	result := &GenerationResult{
		Content:       contentBuilder.String(),
		FinishReason:  string(candidate.FinishReason),
		SafetyRatings: mapSafetyRatings(candidate.SafetyRatings),
		TokenCount:    tokenCount,
		Truncated:     candidate.FinishReason == genai.FinishReasonMaxTokens,
	}

	return result, nil
}

// CountTokens counts the tokens in a given prompt
func (c *geminiClient) CountTokens(ctx context.Context, prompt string) (*TokenCount, error) {
	if prompt == "" {
		return &TokenCount{Total: 0}, nil
	}

	resp, err := c.model.CountTokens(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to count tokens: %w", err)
	}

	return &TokenCount{
		Total: resp.TotalTokens,
	}, nil
}

// ModelDetailsResponse represents the API response for model details
type ModelDetailsResponse struct {
	Name                       string   `json:"name"`
	BaseModelID                string   `json:"baseModelId"`
	Version                    string   `json:"version"`
	DisplayName                string   `json:"displayName"`
	Description                string   `json:"description"`
	InputTokenLimit            int32    `json:"inputTokenLimit"`
	OutputTokenLimit           int32    `json:"outputTokenLimit"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	Temperature                float32  `json:"temperature"`
	TopP                       float32  `json:"topP"`
	TopK                       int32    `json:"topK"`
}

// GetModelInfo retrieves information about the current model
func (c *geminiClient) GetModelInfo(ctx context.Context) (*ModelInfo, error) {
	// Check cache first
	c.modelInfoMutex.RLock()
	if info, ok := c.modelInfoCache[c.modelName]; ok {
		c.modelInfoMutex.RUnlock()
		return info, nil
	}
	c.modelInfoMutex.RUnlock()

	// Not in cache, fetch from API
	info, err := c.fetchModelInfo(ctx, c.modelName)
	if err != nil {
		// If API fetch fails, use conservative defaults
		c.logger.Warn("Failed to fetch model info for %s: %v. Using default values.", c.modelName, err)

		info = &ModelInfo{
			Name:             c.modelName,
			InputTokenLimit:  30720, // Conservative default
			OutputTokenLimit: 8192,  // Conservative default
		}
	}

	// Cache the result (even default values to avoid repeated failures)
	c.modelInfoMutex.Lock()
	c.modelInfoCache[c.modelName] = info
	c.modelInfoMutex.Unlock()

	return info, nil
}

// fetchModelInfo calls the Generative Language API to get model details
func (c *geminiClient) fetchModelInfo(ctx context.Context, modelName string) (*ModelInfo, error) {
	// Construct API URL
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s?key=%s",
		modelName, c.apiKey)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch model info: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var modelDetails ModelDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelDetails); err != nil {
		return nil, fmt.Errorf("failed to parse model info response: %w", err)
	}

	// Convert to our internal model
	info := &ModelInfo{
		Name:             modelDetails.Name,
		InputTokenLimit:  modelDetails.InputTokenLimit,
		OutputTokenLimit: modelDetails.OutputTokenLimit,
	}

	c.logger.Debug("Fetched model info for %s: input limit=%d, output limit=%d",
		modelName, info.InputTokenLimit, info.OutputTokenLimit)

	return info, nil
}

// Close releases resources used by the client
func (c *geminiClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// mapSafetyRatings converts genai safety ratings to our internal format
func mapSafetyRatings(ratings []*genai.SafetyRating) []SafetyRating {
	if ratings == nil {
		return nil
	}

	result := make([]SafetyRating, len(ratings))
	for i, rating := range ratings {
		result[i] = SafetyRating{
			Category: string(rating.Category),
			Blocked:  rating.Blocked,
			Score:    float32(rating.Probability),
		}
	}

	return result
}
