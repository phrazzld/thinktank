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

	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"

	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// HTTPClient is an interface for an HTTP client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// geminiClient implements the llm.LLMClient interface using Google's genai SDK
type geminiClient struct {
	client      *genai.Client
	model       *genai.GenerativeModel
	modelName   string
	apiKey      string
	apiEndpoint string
	logger      logutil.LoggerInterface

	// Model info caching
	modelInfoMutex sync.RWMutex
	httpClient     HTTPClient
}

// geminiClientOption defines a function type for applying options to geminiClient
type geminiClientOption func(*geminiClient)

// newGeminiClient creates a new Gemini client with Google's genai SDK
func newGeminiClient(ctx context.Context, apiKey, modelName, apiEndpoint string, opts ...geminiClientOption) (llm.LLMClient, error) {
	if apiKey == "" {
		return nil, errors.New("API key cannot be empty")
	}

	if modelName == "" {
		return nil, errors.New("model name cannot be empty")
	}

	// Create standard logger for internal client use
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[gemini] ")

	// Prepare client options
	var clientOpts []option.ClientOption

	if apiEndpoint != "" {
		// Custom endpoint (likely for testing)
		logger.Debug("Using custom Gemini API endpoint: %s", apiEndpoint)
		clientOpts = append(clientOpts,
			option.WithEndpoint(apiEndpoint),
			option.WithoutAuthentication()) // Skip auth for mock server
	} else {
		// Default endpoint with API key
		clientOpts = append(clientOpts, option.WithAPIKey(apiKey))
	}

	// Initialize the Google genai client
	client, err := genai.NewClient(ctx, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Initialize the model with default settings
	model := client.GenerativeModel(modelName)
	config := DefaultModelConfig()
	model.SetMaxOutputTokens(config.MaxOutputTokens)
	model.SetTemperature(config.Temperature)
	model.SetTopP(config.TopP)

	// Create the client with default values
	gc := &geminiClient{
		client:         client,
		model:          model,
		modelName:      modelName,
		apiKey:         apiKey,
		apiEndpoint:    apiEndpoint,
		logger:         logger,
		modelInfoMutex: sync.RWMutex{},
		httpClient:     &http.Client{Timeout: 10 * time.Second}, // Default HTTP client
	}

	// Apply any custom options
	for _, opt := range opts {
		opt(gc)
	}

	return gc, nil
}

// GenerateContent implements the llm.LLMClient interface
func (c *geminiClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if prompt == "" {
		return nil, CreateAPIError(
			llm.CategoryInvalidRequest,
			"Cannot generate content with an empty prompt",
			errors.New("prompt cannot be empty"),
			"Provide a task description using the --instructions flag",
		)
	}

	// Apply parameters if provided
	if params != nil {
		// Temperature
		if temp, ok := params["temperature"]; ok {
			switch v := temp.(type) {
			case float64:
				c.model.SetTemperature(float32(v))
			case float32:
				c.model.SetTemperature(v)
			case int:
				c.model.SetTemperature(float32(v))
			}
		}

		// TopP
		if topP, ok := params["top_p"]; ok {
			switch v := topP.(type) {
			case float64:
				c.model.SetTopP(float32(v))
			case float32:
				c.model.SetTopP(v)
			case int:
				c.model.SetTopP(float32(v))
			}
		}

		// TopK
		if topK, ok := params["top_k"]; ok {
			switch v := topK.(type) {
			case int:
				c.model.SetTopK(int32(v))
			case int32:
				c.model.SetTopK(v)
			case int64:
				c.model.SetTopK(int32(v))
			case float64:
				c.model.SetTopK(int32(v))
			}
		}

		// MaxOutputTokens
		if maxTokens, ok := params["max_output_tokens"]; ok {
			switch v := maxTokens.(type) {
			case int:
				c.model.SetMaxOutputTokens(int32(v))
			case int32:
				c.model.SetMaxOutputTokens(v)
			case int64:
				c.model.SetMaxOutputTokens(int32(v))
			case float64:
				c.model.SetMaxOutputTokens(int32(v))
			}
		}
	}

	// Generate content
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		apiErr := FormatAPIError(err, 0)
		// Log detailed info for debugging
		c.logger.Debug("Gemini API Error: %s", apiErr.DebugInfo())
		return nil, apiErr
	}

	// Check for empty response
	if resp == nil {
		return nil, CreateAPIError(
			llm.CategoryUnknown,
			"Received an empty response from the Gemini API",
			errors.New("received nil response from Gemini API"),
			"This is likely a temporary issue. Please try again in a few moments.",
		)
	}

	// Check for empty candidates
	if len(resp.Candidates) == 0 {
		return nil, CreateAPIError(
			llm.CategoryUnknown,
			"The Gemini API returned no generation candidates",
			errors.New("received empty candidates from Gemini API"),
			"This could be due to content filtering. Try modifying your prompt or task description.",
		)
	}

	candidate := resp.Candidates[0]

	// Check for empty content
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return &llm.ProviderResult{
			Content:      "",
			FinishReason: string(candidate.FinishReason),
			SafetyInfo:   toProviderSafety(mapSafetyRatings(candidate.SafetyRatings)),
			Truncated:    candidate.FinishReason == genai.FinishReasonMaxTokens,
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
		// In newer versions of the genai Go SDK, the field may be named differently
		// So we'll use a conservative default count if we can't extract it directly
		tokenCount = 0
	}

	// Build provider-agnostic result
	result := &llm.ProviderResult{
		Content:      contentBuilder.String(),
		FinishReason: string(candidate.FinishReason),
		TokenCount:   tokenCount,
		Truncated:    candidate.FinishReason == genai.FinishReasonMaxTokens,
		SafetyInfo:   toProviderSafety(mapSafetyRatings(candidate.SafetyRatings)),
	}

	return result, nil
}

// CountTokens implements the llm.LLMClient interface
func (c *geminiClient) CountTokens(ctx context.Context, text string) (int32, error) {
	if text == "" {
		return 0, nil
	}

	// Create request to the tokenization endpoint
	endpoint := "https://generativelanguage.googleapis.com/v1beta/models/" + c.modelName + ":countTokens"
	if c.apiEndpoint != "" {
		endpoint = strings.TrimSuffix(c.apiEndpoint, "/") + "/v1beta/models/" + c.modelName + ":countTokens"
	}

	// Prepare the request body
	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": text,
					},
				},
			},
		},
	}

	// Convert request body to JSON
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, CreateAPIError(
			llm.CategoryInvalidRequest,
			"Failed to marshal token counting request",
			err,
			"This is an internal error. Please report it to the developers.",
		)
	}

	// Create the HTTP request
	url := endpoint
	if c.apiEndpoint == "" {
		// Add API key if using public endpoint
		url += "?key=" + c.apiKey
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(requestData)))
	if err != nil {
		return 0, CreateAPIError(
			llm.CategoryNetwork,
			"Failed to create HTTP request for token counting",
			err,
			"This is likely a temporary issue with network connectivity. Check your internet connection and try again.",
		)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, CreateAPIError(
			llm.CategoryNetwork,
			"Failed to connect to Gemini API for token counting",
			err,
			"Check your internet connection and try again. If the issue persists, the API might be experiencing downtime.",
		)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, CreateAPIError(
			llm.CategoryNetwork,
			"Failed to read token counting response",
			err,
			"This is likely a temporary network issue. Please try again.",
		)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		errorMsg := string(body)
		apiErr := FormatAPIError(
			fmt.Errorf("API returned error: %s", errorMsg),
			resp.StatusCode,
		)
		return 0, apiErr
	}

	// Parse the response
	var tokenResponse struct {
		TotalTokens int32 `json:"totalTokens"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return 0, CreateAPIError(
			llm.CategoryInvalidRequest,
			"Failed to parse token counting response",
			err,
			"This is likely an issue with the API response format. Please try again later.",
		)
	}

	return tokenResponse.TotalTokens, nil
}

// GetModelLimits implements the llm.LLMClient interface
func (c *geminiClient) GetModelLimits(ctx context.Context) (*llm.ModelLimits, error) {
	// Create URL for model info endpoint
	modelName := c.modelName
	var url string

	if c.apiEndpoint != "" {
		// Use custom endpoint with no authentication
		url = fmt.Sprintf("%s/v1beta/models/%s",
			strings.TrimSuffix(c.apiEndpoint, "/"), modelName)
	} else {
		// Use default endpoint with API key
		url = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s?key=%s",
			modelName, c.apiKey)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		apiErr := CreateAPIError(
			llm.CategoryNetwork,
			"Failed to create HTTP request for model information",
			err,
			"This is likely a temporary issue with network connectivity. Check your internet connection and try again.",
		)
		return nil, apiErr
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		apiErr := CreateAPIError(
			llm.CategoryNetwork,
			"Failed to connect to Gemini API to fetch model information",
			err,
			"Check your internet connection and try again. If the issue persists, the API might be experiencing downtime.",
		)
		apiErr.Details = err.Error()
		return nil, apiErr
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		apiErr := FormatAPIError(
			fmt.Errorf("API returned error: %s", bodyStr),
			resp.StatusCode,
		)

		// Add model-specific details
		apiErr.Details = fmt.Sprintf("Model: %s, Status: %d, Response: %s",
			modelName, resp.StatusCode, bodyStr)

		// For 404 errors, provide more context about the model
		if resp.StatusCode == http.StatusNotFound {
			apiErr.Message = fmt.Sprintf("Model '%s' not found or not accessible", modelName)
			apiErr.Suggestion = "Check that the model name is correct and that you have access to it. Consider using a different model."
		}

		// Log detailed error for debugging
		c.logger.Debug("Model info error: %s", apiErr.DebugInfo())

		return nil, apiErr
	}

	// Parse response
	var modelDetails struct {
		Name             string `json:"name"`
		InputTokenLimit  int32  `json:"inputTokenLimit"`
		OutputTokenLimit int32  `json:"outputTokenLimit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelDetails); err != nil {
		apiErr := CreateAPIError(
			llm.CategoryInvalidRequest,
			"Failed to parse model information response from Gemini API",
			err,
			"This is likely a temporary API issue or a change in the API response format. Try again later.",
		)
		apiErr.Details = err.Error()
		return nil, apiErr
	}

	// If limits are zero, use reasonable defaults
	if modelDetails.InputTokenLimit <= 0 {
		modelDetails.InputTokenLimit = 8192
	}
	if modelDetails.OutputTokenLimit <= 0 {
		modelDetails.OutputTokenLimit = 2048
	}

	return &llm.ModelLimits{
		InputTokenLimit:  modelDetails.InputTokenLimit,
		OutputTokenLimit: modelDetails.OutputTokenLimit,
	}, nil
}

// Close implements the llm.LLMClient interface by releasing resources
func (c *geminiClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// GetModelName implements the llm.LLMClient interface by returning the model name
func (c *geminiClient) GetModelName() string {
	return c.modelName
}

// GetTemperature returns the temperature setting (for backward compatibility)
func (c *geminiClient) GetTemperature() float32 {
	if c.model != nil && c.model.Temperature != nil {
		return *c.model.Temperature
	}
	return DefaultModelConfig().Temperature
}

// GetMaxOutputTokens returns the max output tokens setting (for backward compatibility)
func (c *geminiClient) GetMaxOutputTokens() int32 {
	if c.model != nil && c.model.MaxOutputTokens != nil {
		return int32(*c.model.MaxOutputTokens)
	}
	return DefaultModelConfig().MaxOutputTokens
}

// GetTopP returns the topP setting (for backward compatibility)
func (c *geminiClient) GetTopP() float32 {
	if c.model != nil && c.model.TopP != nil {
		return *c.model.TopP
	}
	return DefaultModelConfig().TopP
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

// toProviderSafety converts Gemini safety ratings to provider-agnostic safety info
func toProviderSafety(ratings []SafetyRating) []llm.Safety {
	if ratings == nil {
		return nil
	}

	safetyInfo := make([]llm.Safety, len(ratings))
	for i, rating := range ratings {
		safetyInfo[i] = llm.Safety{
			Category: rating.Category,
			Blocked:  rating.Blocked,
			Score:    rating.Score,
		}
	}
	return safetyInfo
}

// The following methods are for implementing the Client interface
// while the geminiClient primarily implements llm.LLMClient interface

// For backward compatibility, create a ClientAdapter type that implements the Client interface
// by wrapping a geminiClient (which now implements llm.LLMClient)

// ClientAdapter wraps a geminiClient to implement the legacy Client interface
type ClientAdapter struct {
	llmClient  llm.LLMClient
	geminiImpl *geminiClient // Optional direct access for Gemini-specific methods
}

// NewClientAdapter creates a ClientAdapter from an llm.LLMClient
func NewClientAdapter(llmClient llm.LLMClient) *ClientAdapter {
	adapter := &ClientAdapter{
		llmClient: llmClient,
	}

	// If it's a geminiClient, store direct reference for Gemini-specific methods
	if gc, ok := llmClient.(*geminiClient); ok {
		adapter.geminiImpl = gc
	}

	return adapter
}

// GenerateContent implements the Client interface
func (a *ClientAdapter) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*GenerationResult, error) {
	// Call the provider-agnostic method
	result, err := a.llmClient.GenerateContent(ctx, prompt, params)
	if err != nil {
		return nil, err
	}

	// Convert to Gemini-specific format
	return &GenerationResult{
		Content:       result.Content,
		FinishReason:  result.FinishReason,
		TokenCount:    result.TokenCount,
		Truncated:     result.Truncated,
		SafetyRatings: fromProviderSafety(result.SafetyInfo),
	}, nil
}

// CountTokens implements the Client interface
func (a *ClientAdapter) CountTokens(ctx context.Context, prompt string) (*TokenCount, error) {
	// Call the provider-agnostic method
	total, err := a.llmClient.CountTokens(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Convert to Gemini-specific format
	return &TokenCount{
		Total: total,
	}, nil
}

// GetModelInfo implements the Client interface
func (a *ClientAdapter) GetModelInfo(ctx context.Context) (*ModelInfo, error) {
	// Call the provider-agnostic method
	limits, err := a.llmClient.GetModelLimits(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to Gemini-specific format
	return &ModelInfo{
		Name:             a.llmClient.GetModelName(),
		InputTokenLimit:  limits.InputTokenLimit,
		OutputTokenLimit: limits.OutputTokenLimit,
	}, nil
}

// GetModelName implements the Client interface
func (a *ClientAdapter) GetModelName() string {
	return a.llmClient.GetModelName()
}

// For Gemini-specific methods, try to use direct geminiClient reference if available,
// otherwise return reasonable defaults

// GetTemperature implements the Client interface
func (a *ClientAdapter) GetTemperature() float32 {
	if a.geminiImpl != nil {
		return a.geminiImpl.GetTemperature()
	}
	return DefaultModelConfig().Temperature
}

// GetMaxOutputTokens implements the Client interface
func (a *ClientAdapter) GetMaxOutputTokens() int32 {
	if a.geminiImpl != nil {
		return a.geminiImpl.GetMaxOutputTokens()
	}
	return DefaultModelConfig().MaxOutputTokens
}

// GetTopP implements the Client interface
func (a *ClientAdapter) GetTopP() float32 {
	if a.geminiImpl != nil {
		return a.geminiImpl.GetTopP()
	}
	return DefaultModelConfig().TopP
}

// Close implements the Client interface
func (a *ClientAdapter) Close() error {
	return a.llmClient.Close()
}

// fromProviderSafety converts llm.Safety to SafetyRating
func fromProviderSafety(ratings []llm.Safety) []SafetyRating {
	if ratings == nil {
		return nil
	}

	safetyRatings := make([]SafetyRating, len(ratings))
	for i, rating := range ratings {
		safetyRatings[i] = SafetyRating{
			Category: rating.Category,
			Blocked:  rating.Blocked,
			Score:    rating.Score,
		}
	}
	return safetyRatings
}

// AsLLMClient is maintained for backward compatibility
// This function adapts the Gemini Client interface to the provider-agnostic LLMClient interface
func AsLLMClient(client Client) llm.LLMClient {
	// If it's a ClientAdapter, extract the wrapped LLMClient
	if adapter, ok := client.(*ClientAdapter); ok {
		return adapter.llmClient
	}

	// Otherwise create a new adapter from scratch
	return &geminiLLMAdapter{client: client}
}

// geminiLLMAdapter adapts a legacy Client to implement the provider-agnostic llm.LLMClient interface
type geminiLLMAdapter struct {
	client Client
}

// GenerateContent implements llm.LLMClient.GenerateContent
func (a *geminiLLMAdapter) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	// This is backward compatibility code to handle params with legacy client interface
	if params != nil {
		// Try to apply parameters using known interfaces
		// Temperature
		if setter, ok := a.client.(interface{ SetTemperature(float32) }); ok {
			if temp, ok := getFloatParam(params, "temperature"); ok {
				setter.SetTemperature(temp)
			}
		}

		// TopP
		if setter, ok := a.client.(interface{ SetTopP(float32) }); ok {
			if topP, ok := getFloatParam(params, "top_p"); ok {
				setter.SetTopP(topP)
			}
		}

		// TopK
		if setter, ok := a.client.(interface{ SetTopK(int32) }); ok {
			if topK, ok := getIntParam(params, "top_k"); ok {
				setter.SetTopK(topK)
			}
		}

		// MaxOutputTokens
		if setter, ok := a.client.(interface{ SetMaxOutputTokens(int32) }); ok {
			if maxTokens, ok := getIntParam(params, "max_output_tokens"); ok {
				setter.SetMaxOutputTokens(maxTokens)
			}
		}
	}

	// Call method with params
	result, err := a.client.GenerateContent(ctx, prompt, params)
	if err != nil {
		return nil, err
	}
	return &llm.ProviderResult{
		Content:      result.Content,
		FinishReason: result.FinishReason,
		TokenCount:   result.TokenCount,
		Truncated:    result.Truncated,
		SafetyInfo:   toProviderSafety(result.SafetyRatings),
	}, nil
}

// CountTokens implements llm.LLMClient.CountTokens
func (a *geminiLLMAdapter) CountTokens(ctx context.Context, text string) (int32, error) {
	result, err := a.client.CountTokens(ctx, text)
	if err != nil {
		return 0, err
	}
	return result.Total, nil
}

// GetModelLimits implements llm.LLMClient.GetModelLimits
func (a *geminiLLMAdapter) GetModelLimits(ctx context.Context) (*llm.ModelLimits, error) {
	info, err := a.client.GetModelInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &llm.ModelLimits{
		InputTokenLimit:  info.InputTokenLimit,
		OutputTokenLimit: info.OutputTokenLimit,
	}, nil
}

// GetModelName implements llm.LLMClient.GetModelName
func (a *geminiLLMAdapter) GetModelName() string {
	return a.client.GetModelName()
}

// Close implements llm.LLMClient.Close
func (a *geminiLLMAdapter) Close() error {
	return a.client.Close()
}

// Helper to extract float parameter
func getFloatParam(params map[string]interface{}, name string) (float32, bool) {
	if params == nil {
		return 0, false
	}

	if val, ok := params[name]; ok {
		switch v := val.(type) {
		case float64:
			return float32(v), true
		case float32:
			return v, true
		case int:
			return float32(v), true
		case int32:
			return float32(v), true
		case int64:
			return float32(v), true
		}
	}
	return 0, false
}

// Helper to extract int parameter
func getIntParam(params map[string]interface{}, name string) (int32, bool) {
	if params == nil {
		return 0, false
	}

	if val, ok := params[name]; ok {
		switch v := val.(type) {
		case int:
			return int32(v), true
		case int32:
			return v, true
		case int64:
			return int32(v), true
		case float64:
			return int32(v), true
		case float32:
			return int32(v), true
		}
	}
	return 0, false
}
