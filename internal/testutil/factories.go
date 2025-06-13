package testutil

import (
	"fmt"
	"strings"
	"time"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/registry"
)

// Test Data Builder/Factory pattern implementation for thinktank test structures.
// This provides fluent APIs for creating test data with sensible defaults
// and easy customization, including invalid variations for error testing.

// ProviderDefinitionBuilder provides a fluent API for building ProviderDefinition test data
type ProviderDefinitionBuilder struct {
	name    string
	baseURL string
}

// NewProviderDefinition creates a new ProviderDefinitionBuilder with sensible defaults
func NewProviderDefinition() *ProviderDefinitionBuilder {
	return &ProviderDefinitionBuilder{
		name:    "test-provider",
		baseURL: "https://api.test-provider.example.com/v1",
	}
}

// WithName sets the provider name
func (b *ProviderDefinitionBuilder) WithName(name string) *ProviderDefinitionBuilder {
	b.name = name
	return b
}

// WithBaseURL sets the provider base URL
func (b *ProviderDefinitionBuilder) WithBaseURL(baseURL string) *ProviderDefinitionBuilder {
	b.baseURL = baseURL
	return b
}

// WithoutBaseURL removes the base URL (uses provider default)
func (b *ProviderDefinitionBuilder) WithoutBaseURL() *ProviderDefinitionBuilder {
	b.baseURL = ""
	return b
}

// InvalidName creates a provider with an invalid name for testing
func (b *ProviderDefinitionBuilder) InvalidName() *ProviderDefinitionBuilder {
	b.name = ""
	return b
}

// InvalidBaseURL creates a provider with an invalid base URL for testing
func (b *ProviderDefinitionBuilder) InvalidBaseURL() *ProviderDefinitionBuilder {
	b.baseURL = "not-a-valid-url"
	return b
}

// Build creates the final ProviderDefinition
func (b *ProviderDefinitionBuilder) Build() registry.ProviderDefinition {
	return registry.ProviderDefinition{
		Name:    b.name,
		BaseURL: b.baseURL,
	}
}

// ParameterDefinitionBuilder provides a fluent API for building ParameterDefinition test data
type ParameterDefinitionBuilder struct {
	paramType  string
	defaultVal interface{}
	min        interface{}
	max        interface{}
	enumValues []string
}

// NewParameterDefinition creates a new ParameterDefinitionBuilder with float type defaults
func NewParameterDefinition() *ParameterDefinitionBuilder {
	return &ParameterDefinitionBuilder{
		paramType:  "float",
		defaultVal: 0.7,
		min:        0.0,
		max:        1.0,
	}
}

// NewFloatParameter creates a float parameter with typical ranges
func NewFloatParameter() *ParameterDefinitionBuilder {
	return NewParameterDefinition() // Already defaults to float
}

// NewIntParameter creates an integer parameter with typical ranges
func NewIntParameter() *ParameterDefinitionBuilder {
	return &ParameterDefinitionBuilder{
		paramType:  "int",
		defaultVal: 1024,
		min:        1,
		max:        8192,
	}
}

// NewStringParameter creates a string parameter with enum values
func NewStringParameter() *ParameterDefinitionBuilder {
	return &ParameterDefinitionBuilder{
		paramType:  "string",
		defaultVal: "default",
		enumValues: []string{"default", "option1", "option2"},
	}
}

// WithType sets the parameter type
func (b *ParameterDefinitionBuilder) WithType(paramType string) *ParameterDefinitionBuilder {
	b.paramType = paramType
	return b
}

// WithDefault sets the default value
func (b *ParameterDefinitionBuilder) WithDefault(defaultVal interface{}) *ParameterDefinitionBuilder {
	b.defaultVal = defaultVal
	return b
}

// WithRange sets min and max values for numeric types
func (b *ParameterDefinitionBuilder) WithRange(min, max interface{}) *ParameterDefinitionBuilder {
	b.min = min
	b.max = max
	return b
}

// WithEnumValues sets allowed values for string parameters
func (b *ParameterDefinitionBuilder) WithEnumValues(values []string) *ParameterDefinitionBuilder {
	b.enumValues = values
	return b
}

// InvalidType creates a parameter with an invalid type for testing
func (b *ParameterDefinitionBuilder) InvalidType() *ParameterDefinitionBuilder {
	b.paramType = "invalid-type"
	return b
}

// InvalidRange creates a parameter with invalid range (min > max) for testing
func (b *ParameterDefinitionBuilder) InvalidRange() *ParameterDefinitionBuilder {
	b.min = 1.0
	b.max = 0.0 // min > max
	return b
}

// Build creates the final ParameterDefinition
func (b *ParameterDefinitionBuilder) Build() registry.ParameterDefinition {
	return registry.ParameterDefinition{
		Type:       b.paramType,
		Default:    b.defaultVal,
		Min:        b.min,
		Max:        b.max,
		EnumValues: b.enumValues,
	}
}

// ModelDefinitionBuilder provides a fluent API for building ModelDefinition test data
type ModelDefinitionBuilder struct {
	name            string
	provider        string
	apiModelID      string
	contextWindow   int32
	maxOutputTokens int32
	parameters      map[string]registry.ParameterDefinition
}

// NewModelDefinition creates a new ModelDefinitionBuilder with sensible defaults
func NewModelDefinition() *ModelDefinitionBuilder {
	return &ModelDefinitionBuilder{
		name:            "test-model",
		provider:        "test-provider",
		apiModelID:      "test-model-api-id",
		contextWindow:   4096,
		maxOutputTokens: 2048,
		parameters: map[string]registry.ParameterDefinition{
			"temperature": NewFloatParameter().WithRange(0.0, 1.0).WithDefault(0.7).Build(),
			"top_p":       NewFloatParameter().WithRange(0.0, 1.0).WithDefault(0.9).Build(),
			"max_tokens":  NewIntParameter().WithRange(1, 4096).WithDefault(1024).Build(),
		},
	}
}

// WithName sets the model name
func (b *ModelDefinitionBuilder) WithName(name string) *ModelDefinitionBuilder {
	b.name = name
	return b
}

// WithProvider sets the provider
func (b *ModelDefinitionBuilder) WithProvider(provider string) *ModelDefinitionBuilder {
	b.provider = provider
	return b
}

// WithAPIModelID sets the API model ID
func (b *ModelDefinitionBuilder) WithAPIModelID(apiModelID string) *ModelDefinitionBuilder {
	b.apiModelID = apiModelID
	return b
}

// WithContextWindow sets the context window size
func (b *ModelDefinitionBuilder) WithContextWindow(contextWindow int32) *ModelDefinitionBuilder {
	b.contextWindow = contextWindow
	return b
}

// WithMaxOutputTokens sets the max output tokens
func (b *ModelDefinitionBuilder) WithMaxOutputTokens(maxOutputTokens int32) *ModelDefinitionBuilder {
	b.maxOutputTokens = maxOutputTokens
	return b
}

// WithParameters sets the parameters map
func (b *ModelDefinitionBuilder) WithParameters(parameters map[string]registry.ParameterDefinition) *ModelDefinitionBuilder {
	b.parameters = parameters
	return b
}

// WithGeminiParameters uses Gemini-specific parameters
func (b *ModelDefinitionBuilder) WithGeminiParameters() *ModelDefinitionBuilder {
	b.parameters = map[string]registry.ParameterDefinition{
		"temperature":      NewFloatParameter().WithRange(0.0, 1.0).WithDefault(0.7).Build(),
		"top_p":            NewFloatParameter().WithRange(0.0, 1.0).WithDefault(0.9).Build(),
		"max_tokens":       NewIntParameter().WithRange(1, 8192).WithDefault(1024).Build(),
		"reasoning_effort": NewFloatParameter().WithRange(0.0, 1.0).WithDefault(0.5).Build(),
	}
	return b
}

// InvalidName creates a model with an invalid name for testing
func (b *ModelDefinitionBuilder) InvalidName() *ModelDefinitionBuilder {
	b.name = ""
	return b
}

// InvalidProvider creates a model with an invalid provider for testing
func (b *ModelDefinitionBuilder) InvalidProvider() *ModelDefinitionBuilder {
	b.provider = ""
	return b
}

// InvalidAPIModelID creates a model with an invalid API model ID for testing
func (b *ModelDefinitionBuilder) InvalidAPIModelID() *ModelDefinitionBuilder {
	b.apiModelID = ""
	return b
}

// InvalidContextWindow creates a model with an invalid context window for testing
func (b *ModelDefinitionBuilder) InvalidContextWindow() *ModelDefinitionBuilder {
	b.contextWindow = -1
	return b
}

// InvalidMaxOutputTokens creates a model with invalid max output tokens for testing
func (b *ModelDefinitionBuilder) InvalidMaxOutputTokens() *ModelDefinitionBuilder {
	b.maxOutputTokens = -1
	return b
}

// Build creates the final ModelDefinition
func (b *ModelDefinitionBuilder) Build() registry.ModelDefinition {
	return registry.ModelDefinition{
		Name:            b.name,
		Provider:        b.provider,
		APIModelID:      b.apiModelID,
		ContextWindow:   b.contextWindow,
		MaxOutputTokens: b.maxOutputTokens,
		Parameters:      b.parameters,
	}
}

// ModelsConfigBuilder provides a fluent API for building ModelsConfig test data
type ModelsConfigBuilder struct {
	apiKeySources map[string]string
	providers     []registry.ProviderDefinition
	models        []registry.ModelDefinition
}

// NewModelsConfig creates a new ModelsConfigBuilder with sensible defaults
func NewModelsConfig() *ModelsConfigBuilder {
	return &ModelsConfigBuilder{
		apiKeySources: map[string]string{
			"openai":        "OPENAI_API_KEY",
			"gemini":        "GEMINI_API_KEY",
			"openrouter":    "OPENROUTER_API_KEY",
			"test-provider": "TEST_PROVIDER_API_KEY",
		},
		providers: []registry.ProviderDefinition{
			NewProviderDefinition().WithName("test-provider").Build(),
		},
		models: []registry.ModelDefinition{
			NewModelDefinition().WithName("test-model").WithProvider("test-provider").Build(),
		},
	}
}

// WithAPIKeySources sets the API key sources map
func (b *ModelsConfigBuilder) WithAPIKeySources(apiKeySources map[string]string) *ModelsConfigBuilder {
	b.apiKeySources = apiKeySources
	return b
}

// WithProviders sets the providers list
func (b *ModelsConfigBuilder) WithProviders(providers []registry.ProviderDefinition) *ModelsConfigBuilder {
	b.providers = providers
	return b
}

// WithModels sets the models list
func (b *ModelsConfigBuilder) WithModels(models []registry.ModelDefinition) *ModelsConfigBuilder {
	b.models = models
	return b
}

// AddProvider adds a provider to the configuration
func (b *ModelsConfigBuilder) AddProvider(provider registry.ProviderDefinition) *ModelsConfigBuilder {
	b.providers = append(b.providers, provider)
	return b
}

// AddModel adds a model to the configuration
func (b *ModelsConfigBuilder) AddModel(model registry.ModelDefinition) *ModelsConfigBuilder {
	b.models = append(b.models, model)
	return b
}

// InvalidAPIKeySources creates config with invalid API key sources for testing
func (b *ModelsConfigBuilder) InvalidAPIKeySources() *ModelsConfigBuilder {
	b.apiKeySources = map[string]string{
		"": "", // Invalid empty provider name and key
	}
	return b
}

// Build creates the final ModelsConfig
func (b *ModelsConfigBuilder) Build() registry.ModelsConfig {
	return registry.ModelsConfig{
		APIKeySources: b.apiKeySources,
		Providers:     b.providers,
		Models:        b.models,
	}
}

// SafetyBuilder provides a fluent API for building Safety test data
type SafetyBuilder struct {
	category string
	blocked  bool
	score    float32
}

// NewSafety creates a new SafetyBuilder with sensible defaults
func NewSafety() *SafetyBuilder {
	return &SafetyBuilder{
		category: "harassment",
		blocked:  false,
		score:    0.1,
	}
}

// WithCategory sets the safety category
func (b *SafetyBuilder) WithCategory(category string) *SafetyBuilder {
	b.category = category
	return b
}

// WithBlocked sets whether content was blocked
func (b *SafetyBuilder) WithBlocked(blocked bool) *SafetyBuilder {
	b.blocked = blocked
	return b
}

// WithScore sets the safety score
func (b *SafetyBuilder) WithScore(score float32) *SafetyBuilder {
	b.score = score
	return b
}

// Blocked creates a blocked safety result for testing
func (b *SafetyBuilder) Blocked() *SafetyBuilder {
	b.blocked = true
	b.score = 0.9
	return b
}

// HighRisk creates a high-risk but not blocked safety result for testing
func (b *SafetyBuilder) HighRisk() *SafetyBuilder {
	b.blocked = false
	b.score = 0.8
	return b
}

// LowRisk creates a low-risk safety result for testing
func (b *SafetyBuilder) LowRisk() *SafetyBuilder {
	b.blocked = false
	b.score = 0.1
	return b
}

// Build creates the final Safety
func (b *SafetyBuilder) Build() llm.Safety {
	return llm.Safety{
		Category: b.category,
		Blocked:  b.blocked,
		Score:    b.score,
	}
}

// ProviderResultBuilder provides a fluent API for building ProviderResult test data
type ProviderResultBuilder struct {
	content      string
	finishReason string
	truncated    bool
	safetyInfo   []llm.Safety
}

// NewProviderResult creates a new ProviderResultBuilder with sensible defaults
func NewProviderResult() *ProviderResultBuilder {
	return &ProviderResultBuilder{
		content:      "This is a test response from the LLM provider.",
		finishReason: "stop",
		truncated:    false,
		safetyInfo:   []llm.Safety{},
	}
}

// WithContent sets the response content
func (b *ProviderResultBuilder) WithContent(content string) *ProviderResultBuilder {
	b.content = content
	return b
}

// WithFinishReason sets the finish reason
func (b *ProviderResultBuilder) WithFinishReason(finishReason string) *ProviderResultBuilder {
	b.finishReason = finishReason
	return b
}

// WithTruncated sets whether the response was truncated
func (b *ProviderResultBuilder) WithTruncated(truncated bool) *ProviderResultBuilder {
	b.truncated = truncated
	return b
}

// WithSafetyInfo sets the safety information
func (b *ProviderResultBuilder) WithSafetyInfo(safetyInfo []llm.Safety) *ProviderResultBuilder {
	b.safetyInfo = safetyInfo
	return b
}

// AddSafety adds a safety entry to the result
func (b *ProviderResultBuilder) AddSafety(safety llm.Safety) *ProviderResultBuilder {
	b.safetyInfo = append(b.safetyInfo, safety)
	return b
}

// Truncated creates a truncated response for testing
func (b *ProviderResultBuilder) Truncated() *ProviderResultBuilder {
	b.finishReason = "length"
	b.truncated = true
	b.content = "This response was truncated due to length limitations..."
	return b
}

// SafetyBlocked creates a safety-blocked response for testing
func (b *ProviderResultBuilder) SafetyBlocked() *ProviderResultBuilder {
	b.finishReason = "safety"
	b.truncated = false
	b.content = ""
	b.safetyInfo = []llm.Safety{
		NewSafety().WithCategory("harmful").Blocked().Build(),
	}
	return b
}

// EmptyContent creates a response with no content for testing
func (b *ProviderResultBuilder) EmptyContent() *ProviderResultBuilder {
	b.content = ""
	return b
}

// JSONContent creates a response with JSON content for testing
func (b *ProviderResultBuilder) JSONContent() *ProviderResultBuilder {
	b.content = `{"name": "Test", "value": 123, "items": ["a", "b", "c"]}`
	return b
}

// CodeContent creates a response with code content for testing
func (b *ProviderResultBuilder) CodeContent() *ProviderResultBuilder {
	b.content = `// Here's an example function
function hello() {
    console.log("Hello, world!");
    return 42;
}`
	return b
}

// Build creates the final ProviderResult
func (b *ProviderResultBuilder) Build() *llm.ProviderResult {
	return &llm.ProviderResult{
		Content:      b.content,
		FinishReason: b.finishReason,
		Truncated:    b.truncated,
		SafetyInfo:   b.safetyInfo,
	}
}

// ChatCompletionMessageBuilder provides a fluent API for building ChatCompletionMessage test data
type ChatCompletionMessageBuilder struct {
	role    string
	content string
}

// NewChatCompletionMessage creates a new ChatCompletionMessageBuilder with sensible defaults
func NewChatCompletionMessage() *ChatCompletionMessageBuilder {
	return &ChatCompletionMessageBuilder{
		role:    "user",
		content: "Hello, how are you?",
	}
}

// WithRole sets the message role
func (b *ChatCompletionMessageBuilder) WithRole(role string) *ChatCompletionMessageBuilder {
	b.role = role
	return b
}

// WithContent sets the message content
func (b *ChatCompletionMessageBuilder) WithContent(content string) *ChatCompletionMessageBuilder {
	b.content = content
	return b
}

// AsUser creates a user message
func (b *ChatCompletionMessageBuilder) AsUser() *ChatCompletionMessageBuilder {
	b.role = "user"
	return b
}

// AsAssistant creates an assistant message
func (b *ChatCompletionMessageBuilder) AsAssistant() *ChatCompletionMessageBuilder {
	b.role = "assistant"
	return b
}

// AsSystem creates a system message
func (b *ChatCompletionMessageBuilder) AsSystem() *ChatCompletionMessageBuilder {
	b.role = "system"
	return b
}

// InvalidRole creates a message with invalid role for testing
func (b *ChatCompletionMessageBuilder) InvalidRole() *ChatCompletionMessageBuilder {
	b.role = "invalid-role"
	return b
}

// EmptyContent creates a message with empty content for testing
func (b *ChatCompletionMessageBuilder) EmptyContent() *ChatCompletionMessageBuilder {
	b.content = ""
	return b
}

// LongContent creates a message with very long content for testing
func (b *ChatCompletionMessageBuilder) LongContent() *ChatCompletionMessageBuilder {
	b.content = fmt.Sprintf("This is a very long message content that repeats: %s",
		strings.Repeat("Long content ", 100))
	return b
}

// Build creates a generic map structure (since we don't have a specific type imported)
func (b *ChatCompletionMessageBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"role":    b.role,
		"content": b.content,
	}
}

// ChatCompletionRequestBuilder provides a fluent API for building ChatCompletionRequest test data
type ChatCompletionRequestBuilder struct {
	model            string
	messages         []map[string]interface{}
	temperature      *float32
	topP             *float32
	frequencyPenalty *float32
	presencePenalty  *float32
	maxTokens        *int32
	stream           bool
}

// NewChatCompletionRequest creates a new ChatCompletionRequestBuilder with sensible defaults
func NewChatCompletionRequest() *ChatCompletionRequestBuilder {
	temp := float32(0.7)
	topP := float32(0.9)
	maxTokens := int32(1024)

	return &ChatCompletionRequestBuilder{
		model: "test-model",
		messages: []map[string]interface{}{
			NewChatCompletionMessage().AsUser().WithContent("Test prompt").Build(),
		},
		temperature: &temp,
		topP:        &topP,
		maxTokens:   &maxTokens,
		stream:      false,
	}
}

// WithModel sets the model
func (b *ChatCompletionRequestBuilder) WithModel(model string) *ChatCompletionRequestBuilder {
	b.model = model
	return b
}

// WithMessages sets the messages
func (b *ChatCompletionRequestBuilder) WithMessages(messages []map[string]interface{}) *ChatCompletionRequestBuilder {
	b.messages = messages
	return b
}

// AddMessage adds a message to the request
func (b *ChatCompletionRequestBuilder) AddMessage(message map[string]interface{}) *ChatCompletionRequestBuilder {
	b.messages = append(b.messages, message)
	return b
}

// WithTemperature sets the temperature
func (b *ChatCompletionRequestBuilder) WithTemperature(temperature float32) *ChatCompletionRequestBuilder {
	b.temperature = &temperature
	return b
}

// WithTopP sets the top_p parameter
func (b *ChatCompletionRequestBuilder) WithTopP(topP float32) *ChatCompletionRequestBuilder {
	b.topP = &topP
	return b
}

// WithFrequencyPenalty sets the frequency penalty
func (b *ChatCompletionRequestBuilder) WithFrequencyPenalty(penalty float32) *ChatCompletionRequestBuilder {
	b.frequencyPenalty = &penalty
	return b
}

// WithPresencePenalty sets the presence penalty
func (b *ChatCompletionRequestBuilder) WithPresencePenalty(penalty float32) *ChatCompletionRequestBuilder {
	b.presencePenalty = &penalty
	return b
}

// WithMaxTokens sets the max tokens
func (b *ChatCompletionRequestBuilder) WithMaxTokens(maxTokens int32) *ChatCompletionRequestBuilder {
	b.maxTokens = &maxTokens
	return b
}

// WithStream sets the stream parameter
func (b *ChatCompletionRequestBuilder) WithStream(stream bool) *ChatCompletionRequestBuilder {
	b.stream = stream
	return b
}

// InvalidModel creates a request with invalid model for testing
func (b *ChatCompletionRequestBuilder) InvalidModel() *ChatCompletionRequestBuilder {
	b.model = ""
	return b
}

// InvalidTemperature creates a request with invalid temperature for testing
func (b *ChatCompletionRequestBuilder) InvalidTemperature() *ChatCompletionRequestBuilder {
	invalidTemp := float32(-1.5) // Outside valid range
	b.temperature = &invalidTemp
	return b
}

// InvalidTopP creates a request with invalid top_p for testing
func (b *ChatCompletionRequestBuilder) InvalidTopP() *ChatCompletionRequestBuilder {
	invalidTopP := float32(1.5) // Outside valid range
	b.topP = &invalidTopP
	return b
}

// InvalidMaxTokens creates a request with invalid max tokens for testing
func (b *ChatCompletionRequestBuilder) InvalidMaxTokens() *ChatCompletionRequestBuilder {
	invalidMaxTokens := int32(-1)
	b.maxTokens = &invalidMaxTokens
	return b
}

// EmptyMessages creates a request with no messages for testing
func (b *ChatCompletionRequestBuilder) EmptyMessages() *ChatCompletionRequestBuilder {
	b.messages = []map[string]interface{}{}
	return b
}

// Build creates a generic map structure representing the request
func (b *ChatCompletionRequestBuilder) Build() map[string]interface{} {
	request := map[string]interface{}{
		"model":    b.model,
		"messages": b.messages,
		"stream":   b.stream,
	}

	if b.temperature != nil {
		request["temperature"] = *b.temperature
	}
	if b.topP != nil {
		request["top_p"] = *b.topP
	}
	if b.frequencyPenalty != nil {
		request["frequency_penalty"] = *b.frequencyPenalty
	}
	if b.presencePenalty != nil {
		request["presence_penalty"] = *b.presencePenalty
	}
	if b.maxTokens != nil {
		request["max_tokens"] = *b.maxTokens
	}

	return request
}

// ChatCompletionResponseBuilder provides a fluent API for building ChatCompletionResponse test data
type ChatCompletionResponseBuilder struct {
	id      string
	object  string
	created int64
	model   string
	choices []map[string]interface{}
	usage   map[string]interface{}
}

// NewChatCompletionResponse creates a new ChatCompletionResponseBuilder with sensible defaults
func NewChatCompletionResponse() *ChatCompletionResponseBuilder {
	return &ChatCompletionResponseBuilder{
		id:      "chatcmpl-test123",
		object:  "chat.completion",
		created: time.Now().Unix(),
		model:   "test-model",
		choices: []map[string]interface{}{
			{
				"index": 0,
				"message": NewChatCompletionMessage().AsAssistant().
					WithContent("This is a test response.").Build(),
				"finish_reason": "stop",
			},
		},
		usage: map[string]interface{}{
			"prompt_tokens":     10,
			"completion_tokens": 20,
			"total_tokens":      30,
		},
	}
}

// WithID sets the response ID
func (b *ChatCompletionResponseBuilder) WithID(id string) *ChatCompletionResponseBuilder {
	b.id = id
	return b
}

// WithModel sets the model
func (b *ChatCompletionResponseBuilder) WithModel(model string) *ChatCompletionResponseBuilder {
	b.model = model
	return b
}

// WithChoices sets the choices
func (b *ChatCompletionResponseBuilder) WithChoices(choices []map[string]interface{}) *ChatCompletionResponseBuilder {
	b.choices = choices
	return b
}

// WithUsage sets the usage information
func (b *ChatCompletionResponseBuilder) WithUsage(usage map[string]interface{}) *ChatCompletionResponseBuilder {
	b.usage = usage
	return b
}

// TruncatedResponse creates a response that was truncated for testing
func (b *ChatCompletionResponseBuilder) TruncatedResponse() *ChatCompletionResponseBuilder {
	b.choices = []map[string]interface{}{
		{
			"index": 0,
			"message": NewChatCompletionMessage().AsAssistant().
				WithContent("This response was truncated...").Build(),
			"finish_reason": "length",
		},
	}
	return b
}

// EmptyResponse creates a response with no content for testing
func (b *ChatCompletionResponseBuilder) EmptyResponse() *ChatCompletionResponseBuilder {
	b.choices = []map[string]interface{}{
		{
			"index": 0,
			"message": NewChatCompletionMessage().AsAssistant().
				WithContent("").Build(),
			"finish_reason": "stop",
		},
	}
	return b
}

// MultipleChoices creates a response with multiple choices for testing
func (b *ChatCompletionResponseBuilder) MultipleChoices() *ChatCompletionResponseBuilder {
	b.choices = []map[string]interface{}{
		{
			"index": 0,
			"message": NewChatCompletionMessage().AsAssistant().
				WithContent("First response option.").Build(),
			"finish_reason": "stop",
		},
		{
			"index": 1,
			"message": NewChatCompletionMessage().AsAssistant().
				WithContent("Second response option.").Build(),
			"finish_reason": "stop",
		},
	}
	return b
}

// Build creates a generic map structure representing the response
func (b *ChatCompletionResponseBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"id":      b.id,
		"object":  b.object,
		"created": b.created,
		"model":   b.model,
		"choices": b.choices,
		"usage":   b.usage,
	}
}
