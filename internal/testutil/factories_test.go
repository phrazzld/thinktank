package testutil

import (
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/registry"
)

func TestProviderDefinitionBuilder(t *testing.T) {
	t.Run("creates provider with defaults", func(t *testing.T) {
		provider := NewProviderDefinition().Build()

		if provider.Name != "test-provider" {
			t.Errorf("Expected default name 'test-provider', got %s", provider.Name)
		}
		if provider.BaseURL != "https://api.test-provider.example.com/v1" {
			t.Errorf("Expected default base URL, got %s", provider.BaseURL)
		}
	})

	t.Run("allows customization", func(t *testing.T) {
		provider := NewProviderDefinition().
			WithName("custom-provider").
			WithBaseURL("https://custom.example.com").
			Build()

		if provider.Name != "custom-provider" {
			t.Errorf("Expected custom name, got %s", provider.Name)
		}
		if provider.BaseURL != "https://custom.example.com" {
			t.Errorf("Expected custom base URL, got %s", provider.BaseURL)
		}
	})

	t.Run("supports provider without base URL", func(t *testing.T) {
		provider := NewProviderDefinition().
			WithName("default-url-provider").
			WithoutBaseURL().
			Build()

		if provider.BaseURL != "" {
			t.Errorf("Expected empty base URL, got %s", provider.BaseURL)
		}
	})

	t.Run("creates invalid provider for testing", func(t *testing.T) {
		provider := NewProviderDefinition().
			InvalidName().
			Build()

		if provider.Name != "" {
			t.Errorf("Expected empty name for invalid provider, got %s", provider.Name)
		}
	})

	t.Run("creates provider with invalid base URL", func(t *testing.T) {
		provider := NewProviderDefinition().
			InvalidBaseURL().
			Build()

		if provider.BaseURL != "not-a-valid-url" {
			t.Errorf("Expected invalid base URL, got %s", provider.BaseURL)
		}
	})
}

func TestParameterDefinitionBuilder(t *testing.T) {
	t.Run("creates float parameter with defaults", func(t *testing.T) {
		param := NewParameterDefinition().Build()

		if param.Type != "float" {
			t.Errorf("Expected type 'float', got %s", param.Type)
		}
		if param.Default != 0.7 {
			t.Errorf("Expected default 0.7, got %v", param.Default)
		}
		if param.Min != 0.0 {
			t.Errorf("Expected min 0.0, got %v", param.Min)
		}
		if param.Max != 1.0 {
			t.Errorf("Expected max 1.0, got %v", param.Max)
		}
	})

	t.Run("creates int parameter", func(t *testing.T) {
		param := NewIntParameter().Build()

		if param.Type != "int" {
			t.Errorf("Expected type 'int', got %s", param.Type)
		}
		if param.Default != 1024 {
			t.Errorf("Expected default 1024, got %v", param.Default)
		}
	})

	t.Run("creates string parameter with enum", func(t *testing.T) {
		param := NewStringParameter().Build()

		if param.Type != "string" {
			t.Errorf("Expected type 'string', got %s", param.Type)
		}
		if len(param.EnumValues) != 3 {
			t.Errorf("Expected 3 enum values, got %d", len(param.EnumValues))
		}
	})

	t.Run("allows customization", func(t *testing.T) {
		param := NewParameterDefinition().
			WithType("custom").
			WithDefault("custom-default").
			WithRange(0.5, 2.0).
			Build()

		if param.Type != "custom" {
			t.Errorf("Expected custom type, got %s", param.Type)
		}
		if param.Default != "custom-default" {
			t.Errorf("Expected custom default, got %v", param.Default)
		}
	})

	t.Run("creates invalid parameter for testing", func(t *testing.T) {
		param := NewParameterDefinition().
			InvalidType().
			Build()

		if param.Type != "invalid-type" {
			t.Errorf("Expected invalid type, got %s", param.Type)
		}
	})

	t.Run("creates parameter with invalid range", func(t *testing.T) {
		param := NewParameterDefinition().
			InvalidRange().
			Build()

		// min should be > max (1.0 > 0.0)
		if param.Min != 1.0 || param.Max != 0.0 {
			t.Errorf("Expected invalid range (min > max), got min=%v, max=%v", param.Min, param.Max)
		}
	})
}

func TestModelDefinitionBuilder(t *testing.T) {
	t.Run("creates model with defaults", func(t *testing.T) {
		model := NewModelDefinition().Build()

		if model.Name != "test-model" {
			t.Errorf("Expected default name 'test-model', got %s", model.Name)
		}
		if model.Provider != "test-provider" {
			t.Errorf("Expected default provider 'test-provider', got %s", model.Provider)
		}
		if model.ContextWindow != 4096 {
			t.Errorf("Expected default context window 4096, got %d", model.ContextWindow)
		}
		if model.MaxOutputTokens != 2048 {
			t.Errorf("Expected default max output tokens 2048, got %d", model.MaxOutputTokens)
		}
		if model.Parameters == nil {
			t.Error("Expected parameters to be set")
		}
	})

	t.Run("allows full customization", func(t *testing.T) {
		customParams := map[string]registry.ParameterDefinition{
			"custom_param": NewParameterDefinition().WithType("string").Build(),
		}

		model := NewModelDefinition().
			WithName("custom-model").
			WithProvider("custom-provider").
			WithAPIModelID("custom-api-id").
			WithContextWindow(8192).
			WithMaxOutputTokens(4096).
			WithParameters(customParams).
			Build()

		if model.Name != "custom-model" {
			t.Errorf("Expected custom name, got %s", model.Name)
		}
		if model.ContextWindow != 8192 {
			t.Errorf("Expected custom context window, got %d", model.ContextWindow)
		}
		if len(model.Parameters) != 1 {
			t.Errorf("Expected 1 custom parameter, got %d", len(model.Parameters))
		}
	})

	t.Run("supports Gemini parameters", func(t *testing.T) {
		model := NewModelDefinition().
			WithGeminiParameters().
			Build()

		// Check if it includes reasoning_effort parameter (specific to Gemini)
		if _, hasReasoningEffort := model.Parameters["reasoning_effort"]; !hasReasoningEffort {
			t.Error("Expected Gemini parameters to include reasoning_effort")
		}
	})

	t.Run("creates invalid model variations", func(t *testing.T) {
		tests := []struct {
			name      string
			builder   func() *ModelDefinitionBuilder
			validator func(registry.ModelDefinition) bool
		}{
			{
				name:      "invalid name",
				builder:   func() *ModelDefinitionBuilder { return NewModelDefinition().InvalidName() },
				validator: func(m registry.ModelDefinition) bool { return m.Name == "" },
			},
			{
				name:      "invalid provider",
				builder:   func() *ModelDefinitionBuilder { return NewModelDefinition().InvalidProvider() },
				validator: func(m registry.ModelDefinition) bool { return m.Provider == "" },
			},
			{
				name:      "invalid API model ID",
				builder:   func() *ModelDefinitionBuilder { return NewModelDefinition().InvalidAPIModelID() },
				validator: func(m registry.ModelDefinition) bool { return m.APIModelID == "" },
			},
			{
				name:      "invalid context window",
				builder:   func() *ModelDefinitionBuilder { return NewModelDefinition().InvalidContextWindow() },
				validator: func(m registry.ModelDefinition) bool { return m.ContextWindow < 0 },
			},
			{
				name:      "invalid max output tokens",
				builder:   func() *ModelDefinitionBuilder { return NewModelDefinition().InvalidMaxOutputTokens() },
				validator: func(m registry.ModelDefinition) bool { return m.MaxOutputTokens < 0 },
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				model := tt.builder().Build()
				if !tt.validator(model) {
					t.Errorf("Invalid model variation %s did not produce expected invalid state", tt.name)
				}
			})
		}
	})
}

func TestModelsConfigBuilder(t *testing.T) {
	t.Run("creates config with defaults", func(t *testing.T) {
		config := NewModelsConfig().Build()

		if len(config.APIKeySources) == 0 {
			t.Error("Expected API key sources to be populated")
		}
		if len(config.Providers) == 0 {
			t.Error("Expected providers to be populated")
		}
		if len(config.Models) == 0 {
			t.Error("Expected models to be populated")
		}
	})

	t.Run("allows adding providers and models", func(t *testing.T) {
		provider := NewProviderDefinition().WithName("new-provider").Build()
		model := NewModelDefinition().WithName("new-model").Build()

		config := NewModelsConfig().
			AddProvider(provider).
			AddModel(model).
			Build()

		if len(config.Providers) != 2 { // Default + added
			t.Errorf("Expected 2 providers, got %d", len(config.Providers))
		}
		if len(config.Models) != 2 { // Default + added
			t.Errorf("Expected 2 models, got %d", len(config.Models))
		}
	})

	t.Run("creates invalid config for testing", func(t *testing.T) {
		config := NewModelsConfig().
			InvalidAPIKeySources().
			Build()

		// Should have an empty key for empty provider name
		if _, hasEmptyKey := config.APIKeySources[""]; !hasEmptyKey {
			t.Error("Expected invalid API key sources with empty provider name")
		}
	})
}

func TestSafetyBuilder(t *testing.T) {
	t.Run("creates safety with defaults", func(t *testing.T) {
		safety := NewSafety().Build()

		if safety.Category != "harassment" {
			t.Errorf("Expected default category 'harassment', got %s", safety.Category)
		}
		if safety.Blocked != false {
			t.Errorf("Expected default blocked false, got %t", safety.Blocked)
		}
		if safety.Score != 0.1 {
			t.Errorf("Expected default score 0.1, got %f", safety.Score)
		}
	})

	t.Run("allows customization", func(t *testing.T) {
		safety := NewSafety().
			WithCategory("violence").
			WithBlocked(true).
			WithScore(0.8).
			Build()

		if safety.Category != "violence" {
			t.Errorf("Expected category 'violence', got %s", safety.Category)
		}
		if safety.Blocked != true {
			t.Errorf("Expected blocked true, got %t", safety.Blocked)
		}
		if safety.Score != 0.8 {
			t.Errorf("Expected score 0.8, got %f", safety.Score)
		}
	})

	t.Run("creates safety variations", func(t *testing.T) {
		tests := []struct {
			name      string
			builder   func() *SafetyBuilder
			validator func(llm.Safety) bool
		}{
			{
				name:      "blocked",
				builder:   func() *SafetyBuilder { return NewSafety().Blocked() },
				validator: func(s llm.Safety) bool { return s.Blocked && s.Score == 0.9 },
			},
			{
				name:      "high risk",
				builder:   func() *SafetyBuilder { return NewSafety().HighRisk() },
				validator: func(s llm.Safety) bool { return !s.Blocked && s.Score == 0.8 },
			},
			{
				name:      "low risk",
				builder:   func() *SafetyBuilder { return NewSafety().LowRisk() },
				validator: func(s llm.Safety) bool { return !s.Blocked && s.Score == 0.1 },
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				safety := tt.builder().Build()
				if !tt.validator(safety) {
					t.Errorf("Safety variation %s did not produce expected state", tt.name)
				}
			})
		}
	})
}

func TestProviderResultBuilder(t *testing.T) {
	t.Run("creates result with defaults", func(t *testing.T) {
		result := NewProviderResult().Build()

		if result.Content == "" {
			t.Error("Expected default content to be non-empty")
		}
		if result.FinishReason != "stop" {
			t.Errorf("Expected default finish reason 'stop', got %s", result.FinishReason)
		}
		if result.Truncated != false {
			t.Errorf("Expected default truncated false, got %t", result.Truncated)
		}
		if result.SafetyInfo == nil {
			t.Error("Expected safety info to be initialized")
		}
	})

	t.Run("allows customization", func(t *testing.T) {
		customSafety := NewSafety().WithCategory("custom").Build()
		result := NewProviderResult().
			WithContent("Custom content").
			WithFinishReason("custom_reason").
			WithTruncated(true).
			AddSafety(customSafety).
			Build()

		if result.Content != "Custom content" {
			t.Errorf("Expected custom content, got %s", result.Content)
		}
		if result.FinishReason != "custom_reason" {
			t.Errorf("Expected custom finish reason, got %s", result.FinishReason)
		}
		if result.Truncated != true {
			t.Errorf("Expected truncated true, got %t", result.Truncated)
		}
		if len(result.SafetyInfo) != 1 {
			t.Errorf("Expected 1 safety info entry, got %d", len(result.SafetyInfo))
		}
	})

	t.Run("creates result variations", func(t *testing.T) {
		tests := []struct {
			name      string
			builder   func() *ProviderResultBuilder
			validator func(*llm.ProviderResult) bool
		}{
			{
				name:      "truncated",
				builder:   func() *ProviderResultBuilder { return NewProviderResult().Truncated() },
				validator: func(r *llm.ProviderResult) bool { return r.Truncated && r.FinishReason == "length" },
			},
			{
				name:      "safety blocked",
				builder:   func() *ProviderResultBuilder { return NewProviderResult().SafetyBlocked() },
				validator: func(r *llm.ProviderResult) bool { return r.FinishReason == "safety" && len(r.SafetyInfo) > 0 },
			},
			{
				name:      "empty content",
				builder:   func() *ProviderResultBuilder { return NewProviderResult().EmptyContent() },
				validator: func(r *llm.ProviderResult) bool { return r.Content == "" },
			},
			{
				name:      "JSON content",
				builder:   func() *ProviderResultBuilder { return NewProviderResult().JSONContent() },
				validator: func(r *llm.ProviderResult) bool { return strings.Contains(r.Content, "{") },
			},
			{
				name:      "code content",
				builder:   func() *ProviderResultBuilder { return NewProviderResult().CodeContent() },
				validator: func(r *llm.ProviderResult) bool { return strings.Contains(r.Content, "function") },
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.builder().Build()
				if !tt.validator(result) {
					t.Errorf("Result variation %s did not produce expected state", tt.name)
				}
			})
		}
	})
}

func TestChatCompletionMessageBuilder(t *testing.T) {
	t.Run("creates message with defaults", func(t *testing.T) {
		message := NewChatCompletionMessage().Build()

		if message["role"] != "user" {
			t.Errorf("Expected default role 'user', got %v", message["role"])
		}
		if message["content"] == "" {
			t.Error("Expected default content to be non-empty")
		}
	})

	t.Run("allows role customization", func(t *testing.T) {
		tests := []struct {
			name     string
			builder  func() *ChatCompletionMessageBuilder
			expected string
		}{
			{
				name:     "user message",
				builder:  func() *ChatCompletionMessageBuilder { return NewChatCompletionMessage().AsUser() },
				expected: "user",
			},
			{
				name:     "assistant message",
				builder:  func() *ChatCompletionMessageBuilder { return NewChatCompletionMessage().AsAssistant() },
				expected: "assistant",
			},
			{
				name:     "system message",
				builder:  func() *ChatCompletionMessageBuilder { return NewChatCompletionMessage().AsSystem() },
				expected: "system",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				message := tt.builder().Build()
				if message["role"] != tt.expected {
					t.Errorf("Expected role %s, got %v", tt.expected, message["role"])
				}
			})
		}
	})

	t.Run("creates invalid message variations", func(t *testing.T) {
		tests := []struct {
			name      string
			builder   func() *ChatCompletionMessageBuilder
			validator func(map[string]interface{}) bool
		}{
			{
				name:      "invalid role",
				builder:   func() *ChatCompletionMessageBuilder { return NewChatCompletionMessage().InvalidRole() },
				validator: func(m map[string]interface{}) bool { return m["role"] == "invalid-role" },
			},
			{
				name:      "empty content",
				builder:   func() *ChatCompletionMessageBuilder { return NewChatCompletionMessage().EmptyContent() },
				validator: func(m map[string]interface{}) bool { return m["content"] == "" },
			},
			{
				name:      "long content",
				builder:   func() *ChatCompletionMessageBuilder { return NewChatCompletionMessage().LongContent() },
				validator: func(m map[string]interface{}) bool { return len(m["content"].(string)) > 1000 },
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				message := tt.builder().Build()
				if !tt.validator(message) {
					t.Errorf("Message variation %s did not produce expected state", tt.name)
				}
			})
		}
	})
}

func TestChatCompletionRequestBuilder(t *testing.T) {
	t.Run("creates request with defaults", func(t *testing.T) {
		request := NewChatCompletionRequest().Build()

		if request["model"] != "test-model" {
			t.Errorf("Expected default model 'test-model', got %v", request["model"])
		}

		messages, ok := request["messages"].([]map[string]interface{})
		if !ok || len(messages) == 0 {
			t.Error("Expected messages to be populated")
		}

		if request["temperature"] == nil {
			t.Error("Expected temperature to be set")
		}

		if request["stream"] != false {
			t.Errorf("Expected stream false, got %v", request["stream"])
		}
	})

	t.Run("allows parameter customization", func(t *testing.T) {
		request := NewChatCompletionRequest().
			WithModel("custom-model").
			WithTemperature(0.5).
			WithTopP(0.8).
			WithMaxTokens(2048).
			WithStream(true).
			Build()

		if request["model"] != "custom-model" {
			t.Errorf("Expected custom model, got %v", request["model"])
		}
		if request["temperature"] != float32(0.5) {
			t.Errorf("Expected temperature 0.5, got %v", request["temperature"])
		}
		if request["stream"] != true {
			t.Errorf("Expected stream true, got %v", request["stream"])
		}
	})

	t.Run("allows message customization", func(t *testing.T) {
		customMessage := NewChatCompletionMessage().AsSystem().WithContent("System prompt").Build()
		request := NewChatCompletionRequest().
			WithMessages([]map[string]interface{}{customMessage}).
			Build()

		messages := request["messages"].([]map[string]interface{})
		if len(messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(messages))
		}
		if messages[0]["role"] != "system" {
			t.Errorf("Expected system message, got %v", messages[0]["role"])
		}
	})

	t.Run("creates invalid request variations", func(t *testing.T) {
		tests := []struct {
			name      string
			builder   func() *ChatCompletionRequestBuilder
			validator func(map[string]interface{}) bool
		}{
			{
				name:      "invalid model",
				builder:   func() *ChatCompletionRequestBuilder { return NewChatCompletionRequest().InvalidModel() },
				validator: func(r map[string]interface{}) bool { return r["model"] == "" },
			},
			{
				name:      "invalid temperature",
				builder:   func() *ChatCompletionRequestBuilder { return NewChatCompletionRequest().InvalidTemperature() },
				validator: func(r map[string]interface{}) bool { return r["temperature"].(float32) < 0 },
			},
			{
				name:      "invalid top_p",
				builder:   func() *ChatCompletionRequestBuilder { return NewChatCompletionRequest().InvalidTopP() },
				validator: func(r map[string]interface{}) bool { return r["top_p"].(float32) > 1.0 },
			},
			{
				name:      "invalid max tokens",
				builder:   func() *ChatCompletionRequestBuilder { return NewChatCompletionRequest().InvalidMaxTokens() },
				validator: func(r map[string]interface{}) bool { return r["max_tokens"].(int32) < 0 },
			},
			{
				name:      "empty messages",
				builder:   func() *ChatCompletionRequestBuilder { return NewChatCompletionRequest().EmptyMessages() },
				validator: func(r map[string]interface{}) bool { return len(r["messages"].([]map[string]interface{})) == 0 },
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				request := tt.builder().Build()
				if !tt.validator(request) {
					t.Errorf("Request variation %s did not produce expected state", tt.name)
				}
			})
		}
	})
}

func TestChatCompletionResponseBuilder(t *testing.T) {
	t.Run("creates response with defaults", func(t *testing.T) {
		response := NewChatCompletionResponse().Build()

		if response["id"] == "" {
			t.Error("Expected ID to be populated")
		}
		if response["object"] != "chat.completion" {
			t.Errorf("Expected object 'chat.completion', got %v", response["object"])
		}
		if response["model"] != "test-model" {
			t.Errorf("Expected model 'test-model', got %v", response["model"])
		}

		choices, ok := response["choices"].([]map[string]interface{})
		if !ok || len(choices) == 0 {
			t.Error("Expected choices to be populated")
		}

		usage, ok := response["usage"].(map[string]interface{})
		if !ok {
			t.Error("Expected usage to be populated")
		} else {
			if usage["total_tokens"] != 30 {
				t.Errorf("Expected total tokens 30, got %v", usage["total_tokens"])
			}
		}
	})

	t.Run("creates response variations", func(t *testing.T) {
		tests := []struct {
			name      string
			builder   func() *ChatCompletionResponseBuilder
			validator func(map[string]interface{}) bool
		}{
			{
				name:    "truncated response",
				builder: func() *ChatCompletionResponseBuilder { return NewChatCompletionResponse().TruncatedResponse() },
				validator: func(r map[string]interface{}) bool {
					choices := r["choices"].([]map[string]interface{})
					return choices[0]["finish_reason"] == "length"
				},
			},
			{
				name:    "empty response",
				builder: func() *ChatCompletionResponseBuilder { return NewChatCompletionResponse().EmptyResponse() },
				validator: func(r map[string]interface{}) bool {
					choices := r["choices"].([]map[string]interface{})
					message := choices[0]["message"].(map[string]interface{})
					return message["content"] == ""
				},
			},
			{
				name:    "multiple choices",
				builder: func() *ChatCompletionResponseBuilder { return NewChatCompletionResponse().MultipleChoices() },
				validator: func(r map[string]interface{}) bool {
					choices := r["choices"].([]map[string]interface{})
					return len(choices) == 2
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				response := tt.builder().Build()
				if !tt.validator(response) {
					t.Errorf("Response variation %s did not produce expected state", tt.name)
				}
			})
		}
	})
}

// TestFactoryIntegration demonstrates realistic usage patterns combining multiple factories
func TestFactoryIntegration(t *testing.T) {
	t.Run("complete provider setup", func(t *testing.T) {
		// Create a complete provider setup for testing
		provider := NewProviderDefinition().
			WithName("integration-provider").
			WithBaseURL("https://api.integration.test").
			Build()

		model := NewModelDefinition().
			WithName("integration-model").
			WithProvider("integration-provider").
			WithAPIModelID("integration-v1").
			WithContextWindow(8192).
			WithMaxOutputTokens(4096).
			Build()

		config := NewModelsConfig().
			AddProvider(provider).
			AddModel(model).
			Build()

		// Verify the integration works correctly
		if len(config.Providers) < 2 { // Default + added
			t.Error("Expected provider to be added to config")
		}
		if len(config.Models) < 2 { // Default + added
			t.Error("Expected model to be added to config")
		}

		// Find our added provider and model
		var foundProvider *registry.ProviderDefinition
		var foundModel *registry.ModelDefinition

		for _, p := range config.Providers {
			if p.Name == "integration-provider" {
				foundProvider = &p
				break
			}
		}

		for _, m := range config.Models {
			if m.Name == "integration-model" {
				foundModel = &m
				break
			}
		}

		if foundProvider == nil {
			t.Error("Added provider not found in config")
		}
		if foundModel == nil {
			t.Error("Added model not found in config")
		}
	})

	t.Run("complete API conversation flow", func(t *testing.T) {
		// Create a request with explicit messages (not using AddMessage to avoid default message)
		systemMessage := NewChatCompletionMessage().AsSystem().WithContent("You are a helpful assistant.").Build()
		userMessage := NewChatCompletionMessage().AsUser().WithContent("Hello!").Build()

		request := NewChatCompletionRequest().
			WithModel("conversation-model").
			WithMessages([]map[string]interface{}{systemMessage, userMessage}).
			Build()

		// Create a response
		response := NewChatCompletionResponse().
			WithModel("conversation-model").
			Build()

		// Create a provider result
		result := NewProviderResult().
			WithContent("Hello! How can I help you today?").
			WithFinishReason("stop").
			Build()

		// Verify the conversation flow makes sense
		messages := request["messages"].([]map[string]interface{})
		if len(messages) != 2 {
			t.Errorf("Expected 2 messages in conversation, got %d", len(messages))
		}

		if messages[0]["role"] != "system" {
			t.Error("Expected first message to be system role")
		}
		if messages[1]["role"] != "user" {
			t.Error("Expected second message to be user role")
		}

		if result.FinishReason != "stop" {
			t.Error("Expected successful completion")
		}

		// Verify response structure
		choices := response["choices"].([]map[string]interface{})
		if len(choices) == 0 {
			t.Error("Expected response to have choices")
		}
	})

	t.Run("error scenario testing", func(t *testing.T) {
		// Create various error scenarios for comprehensive testing

		// Invalid request
		invalidRequest := NewChatCompletionRequest().
			InvalidModel().
			InvalidTemperature().
			EmptyMessages().
			Build()

		// Safety-blocked result
		blockedResult := NewProviderResult().
			SafetyBlocked().
			Build()

		// Truncated result
		truncatedResult := NewProviderResult().
			Truncated().
			Build()

		// Verify error conditions
		if invalidRequest["model"] != "" {
			t.Error("Expected invalid model to be empty")
		}
		if blockedResult.FinishReason != "safety" {
			t.Error("Expected safety-blocked result")
		}
		if !truncatedResult.Truncated {
			t.Error("Expected truncated result")
		}
	})
}
