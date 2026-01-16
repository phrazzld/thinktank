// Package thinktank contains the core application logic for the thinktank tool.
// This file contains test implementations for adapter functionality.
package thinktank

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/models"
)

// setupAPIServiceAdapterTest creates test fixtures for APIServiceAdapter testing
func setupAPIServiceAdapterTest() (*APIServiceAdapter, *MockAPIServiceForAdapter) {
	mockAPIService := &MockAPIServiceForAdapter{}
	adapter := &APIServiceAdapter{
		APIService: mockAPIService,
	}
	return adapter, mockAPIService
}

// Note: setupContextGathererAdapterTest was removed as part of issue #121
// ContextGatherer now directly implements interfaces.ContextGatherer

// setupFileWriterAdapterTest creates test fixtures for FileWriterAdapter testing
func setupFileWriterAdapterTest() (*FileWriterAdapter, *MockFileWriter) {
	mockFileWriter := NewMockFileWriter()
	adapter := &FileWriterAdapter{
		FileWriter: mockFileWriter,
	}
	return adapter, mockFileWriter
}

// Basic Test Cases for APIServiceAdapter

// TestAPIServiceAdapter_InitLLMClient verifies that InitLLMClient calls are properly delegated
func TestAPIServiceAdapter_InitLLMClient(t *testing.T) {
	adapter, mock := setupAPIServiceAdapterTest()

	// Set up expected return values
	expectedClient := &llm.MockLLMClient{}
	expectedErr := errors.New("test error")
	mock.InitLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		return expectedClient, expectedErr
	}

	// Call the adapter method
	ctx := context.Background()
	client, err := adapter.InitLLMClient(ctx, "test-key", "test-model", "test-endpoint")

	// Verify that the adapter delegated the call correctly
	if client != expectedClient {
		t.Errorf("Expected client %v, got %v", expectedClient, client)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Verify that the mock was called with the expected parameters
	if len(mock.InitLLMClientCalls) != 1 {
		t.Fatalf("Expected 1 call to InitLLMClient, got %d", len(mock.InitLLMClientCalls))
	}
	call := mock.InitLLMClientCalls[0]
	if call.APIKey != "test-key" || call.ModelName != "test-model" || call.APIEndpoint != "test-endpoint" {
		t.Errorf("Unexpected parameters: API Key = %s, Model Name = %s, API Endpoint = %s",
			call.APIKey, call.ModelName, call.APIEndpoint)
	}
}

// TestAPIServiceAdapter_ProcessLLMResponse verifies that ProcessLLMResponse calls are properly delegated
func TestAPIServiceAdapter_ProcessLLMResponse(t *testing.T) {
	adapter, mock := setupAPIServiceAdapterTest()

	// Set up expected return values
	expectedResult := "processed content"
	expectedErr := errors.New("test error")
	mock.ProcessLLMResponseFunc = func(result *llm.ProviderResult) (string, error) {
		return expectedResult, expectedErr
	}

	// Create a test input
	providerResult := &llm.ProviderResult{Content: "test content"}

	// Call the adapter method
	result, err := adapter.ProcessLLMResponse(providerResult)

	// Verify that the adapter delegated the call correctly
	if result != expectedResult {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Verify that the mock was called with the expected parameters
	if len(mock.ProcessLLMResponseCalls) != 1 {
		t.Fatalf("Expected 1 call to ProcessLLMResponse, got %d", len(mock.ProcessLLMResponseCalls))
	}
	call := mock.ProcessLLMResponseCalls[0]
	if call.Result != providerResult {
		t.Errorf("Unexpected provider result: %v", call.Result)
	}
}

// TestAPIServiceAdapter_GetErrorDetails verifies that GetErrorDetails calls are properly delegated
func TestAPIServiceAdapter_GetErrorDetails(t *testing.T) {
	adapter, mock := setupAPIServiceAdapterTest()

	// Set up expected return value
	expectedDetails := "detailed error message"
	mock.GetErrorDetailsFunc = func(err error) string {
		return expectedDetails
	}

	// Create a test input
	testErr := errors.New("test error")

	// Call the adapter method
	details := adapter.GetErrorDetails(testErr)

	// Verify that the adapter delegated the call correctly
	if details != expectedDetails {
		t.Errorf("Expected details %v, got %v", expectedDetails, details)
	}

	// Verify that the mock was called with the expected parameters
	if len(mock.GetErrorDetailsCalls) != 1 {
		t.Fatalf("Expected 1 call to GetErrorDetails, got %d", len(mock.GetErrorDetailsCalls))
	}
	call := mock.GetErrorDetailsCalls[0]
	if call.Err != testErr {
		t.Errorf("Unexpected error: %v", call.Err)
	}
}

// TestAPIServiceAdapter_IsEmptyResponseError verifies that IsEmptyResponseError calls are properly delegated
func TestAPIServiceAdapter_IsEmptyResponseError(t *testing.T) {
	adapter, mock := setupAPIServiceAdapterTest()

	// Set up expected return value
	expectedResult := true
	mock.IsEmptyResponseErrorFunc = func(err error) bool {
		return expectedResult
	}

	// Create a test input
	testErr := errors.New("test error")

	// Call the adapter method
	result := adapter.IsEmptyResponseError(testErr)

	// Verify that the adapter delegated the call correctly
	if result != expectedResult {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}

	// Verify that the mock was called with the expected parameters
	if len(mock.IsEmptyResponseErrorCalls) != 1 {
		t.Fatalf("Expected 1 call to IsEmptyResponseError, got %d", len(mock.IsEmptyResponseErrorCalls))
	}
	call := mock.IsEmptyResponseErrorCalls[0]
	if call.Err != testErr {
		t.Errorf("Unexpected error: %v", call.Err)
	}
}

// TestAPIServiceAdapter_IsSafetyBlockedError verifies that IsSafetyBlockedError calls are properly delegated
func TestAPIServiceAdapter_IsSafetyBlockedError(t *testing.T) {
	adapter, mock := setupAPIServiceAdapterTest()

	// Set up expected return value
	expectedResult := true
	mock.IsSafetyBlockedErrorFunc = func(err error) bool {
		return expectedResult
	}

	// Create a test input
	testErr := errors.New("test error")

	// Call the adapter method
	result := adapter.IsSafetyBlockedError(testErr)

	// Verify that the adapter delegated the call correctly
	if result != expectedResult {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}

	// Verify that the mock was called with the expected parameters
	if len(mock.IsSafetyBlockedErrorCalls) != 1 {
		t.Fatalf("Expected 1 call to IsSafetyBlockedError, got %d", len(mock.IsSafetyBlockedErrorCalls))
	}
	call := mock.IsSafetyBlockedErrorCalls[0]
	if call.Err != testErr {
		t.Errorf("Unexpected error: %v", call.Err)
	}
}

// Tests for methods with fallback logic

// TestAPIServiceAdapter_GetModelParameters_WithImplementation verifies GetModelParameters with an implementation
func TestAPIServiceAdapter_GetModelParameters_WithImplementation(t *testing.T) {
	adapter, mock := setupAPIServiceAdapterTest()

	// Set up expected return values
	expectedParams := map[string]interface{}{"temperature": 0.7}
	expectedErr := errors.New("test error")
	mock.GetModelParametersFunc = func(ctx context.Context, modelName string) (map[string]interface{}, error) {
		return expectedParams, expectedErr
	}

	// Call the adapter method
	ctx := context.Background()
	params, err := adapter.GetModelParameters(ctx, "test-model")

	// Verify that the adapter delegated the call correctly
	if !reflect.DeepEqual(params, expectedParams) {
		t.Errorf("Expected params %v, got %v", expectedParams, params)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Verify that the mock was called with the expected parameters
	if len(mock.GetModelParametersCalls) != 1 {
		t.Fatalf("Expected 1 call to GetModelParameters, got %d", len(mock.GetModelParametersCalls))
	}
	call := mock.GetModelParametersCalls[0]
	if call.ModelName != "test-model" {
		t.Errorf("Unexpected model name: %s", call.ModelName)
	}
}

// TestAPIServiceAdapter_GetModelParameters_WithoutImplementation verifies GetModelParameters fallback
func TestAPIServiceAdapter_GetModelParameters_WithoutImplementation(t *testing.T) {
	// Create a minimal API service that only implements the required interfaces.APIService methods
	mockAPIService := NewMockAPIServiceWithoutExtensions()
	adapter := &APIServiceAdapter{
		APIService: mockAPIService,
	}

	// Call the adapter method with adapter that uses interface assertion for extension methods
	ctx := context.Background()
	params, err := adapter.GetModelParameters(ctx, "test-model")

	// Verify that the adapter returned the fallback values
	if len(params) != 0 {
		t.Errorf("Expected empty params, got %v", params)
	}
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

// TestAPIServiceAdapter_ValidateModelParameter_WithImplementation verifies ValidateModelParameter with an implementation
func TestAPIServiceAdapter_ValidateModelParameter_WithImplementation(t *testing.T) {
	adapter, mock := setupAPIServiceAdapterTest()

	// Set up expected return values
	expectedResult := true
	expectedErr := errors.New("test error")
	mock.ValidateModelParameterFunc = func(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
		return expectedResult, expectedErr
	}

	// Call the adapter method
	ctx := context.Background()
	result, err := adapter.ValidateModelParameter(ctx, "test-model", "test-param", 0.7)

	// Verify that the adapter delegated the call correctly
	if result != expectedResult {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Verify that the mock was called with the expected parameters
	if len(mock.ValidateModelParameterCalls) != 1 {
		t.Fatalf("Expected 1 call to ValidateModelParameter, got %d", len(mock.ValidateModelParameterCalls))
	}
	call := mock.ValidateModelParameterCalls[0]
	if call.ModelName != "test-model" || call.ParamName != "test-param" || call.Value != 0.7 {
		t.Errorf("Unexpected parameters: Model Name = %s, Param Name = %s, Value = %v",
			call.ModelName, call.ParamName, call.Value)
	}
}

// TestAPIServiceAdapter_ValidateModelParameter_WithoutImplementation verifies ValidateModelParameter fallback
func TestAPIServiceAdapter_ValidateModelParameter_WithoutImplementation(t *testing.T) {
	// Create a minimal API service that only implements the required interfaces.APIService methods
	mockAPIService := NewMockAPIServiceWithoutExtensions()
	adapter := &APIServiceAdapter{
		APIService: mockAPIService,
	}

	// Call the adapter method with adapter that uses interface assertion for extension methods
	ctx := context.Background()
	result, err := adapter.ValidateModelParameter(ctx, "test-model", "test-param", 0.7)

	// Verify that the adapter returned the fallback values
	if !result {
		t.Errorf("Expected true result, got %v", result)
	}
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

// TestAPIServiceAdapter_GetModelDefinition_WithImplementation verifies GetModelDefinition with an implementation
func TestAPIServiceAdapter_GetModelDefinition_WithImplementation(t *testing.T) {
	adapter, mock := setupAPIServiceAdapterTest()

	// Set up expected return values
	expectedDef := &models.ModelInfo{Provider: "test-provider", APIModelID: "test-model"}
	expectedErr := errors.New("test error")
	mock.GetModelDefinitionFunc = func(ctx context.Context, modelName string) (*models.ModelInfo, error) {
		return expectedDef, expectedErr
	}

	// Call the adapter method
	ctx := context.Background()
	def, err := adapter.GetModelDefinition(ctx, "test-model")

	// Verify that the adapter delegated the call correctly
	if def != expectedDef {
		t.Errorf("Expected definition %v, got %v", expectedDef, def)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Verify that the mock was called with the expected parameters
	if len(mock.GetModelDefinitionCalls) != 1 {
		t.Fatalf("Expected 1 call to GetModelDefinition, got %d", len(mock.GetModelDefinitionCalls))
	}
	call := mock.GetModelDefinitionCalls[0]
	if call.ModelName != "test-model" {
		t.Errorf("Unexpected model name: %s", call.ModelName)
	}
}

// TestAPIServiceAdapter_GetModelDefinition_WithoutImplementation verifies GetModelDefinition fallback
func TestAPIServiceAdapter_GetModelDefinition_WithoutImplementation(t *testing.T) {
	// Create a minimal API service that only implements the required interfaces.APIService methods
	mockAPIService := NewMockAPIServiceWithoutExtensions()
	adapter := &APIServiceAdapter{
		APIService: mockAPIService,
	}

	// Call the adapter method with adapter that uses interface assertion for extension methods
	ctx := context.Background()
	def, err := adapter.GetModelDefinition(ctx, "test-model")

	// Verify that the adapter returned the fallback values
	if def != nil {
		t.Errorf("Expected nil definition, got %v", def)
	}
	if err == nil || err.Error() != "model definition not available" {
		t.Errorf("Expected 'model definition not available' error, got %v", err)
	}
}

// TestAPIServiceAdapter_GetModelTokenLimits_WithImplementation verifies GetModelTokenLimits with an implementation
func TestAPIServiceAdapter_GetModelTokenLimits_WithImplementation(t *testing.T) {
	adapter, mock := setupAPIServiceAdapterTest()

	// Set up expected return values
	expectedContextWindow := int32(8192)
	expectedMaxTokens := int32(2048)
	expectedErr := errors.New("test error")
	mock.GetModelTokenLimitsFunc = func(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
		return expectedContextWindow, expectedMaxTokens, expectedErr
	}

	// Call the adapter method
	ctx := context.Background()
	contextWindow, maxTokens, err := adapter.GetModelTokenLimits(ctx, "test-model")

	// Verify that the adapter delegated the call correctly
	if contextWindow != expectedContextWindow {
		t.Errorf("Expected context window %v, got %v", expectedContextWindow, contextWindow)
	}
	if maxTokens != expectedMaxTokens {
		t.Errorf("Expected max tokens %v, got %v", expectedMaxTokens, maxTokens)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Verify that the mock was called with the expected parameters
	if len(mock.GetModelTokenLimitsCalls) != 1 {
		t.Fatalf("Expected 1 call to GetModelTokenLimits, got %d", len(mock.GetModelTokenLimitsCalls))
	}
	call := mock.GetModelTokenLimitsCalls[0]
	if call.ModelName != "test-model" {
		t.Errorf("Unexpected model name: %s", call.ModelName)
	}
}

// We're removing TestAPIServiceAdapter_GetModelTokenLimits_WithoutImplementation_KnownModels
// Since we've refactored the adapter to remove the special type assertions for MockAPIServiceWithoutExtensions,
// this test is no longer relevant, and the adapter falls back to the implementation in the mock itself

// Note: ContextGathererAdapter tests were removed as part of issue #121
// ContextGatherer now directly implements interfaces.ContextGatherer

// Tests for FileWriterAdapter

// TestFileWriterAdapter_SaveToFile verifies that SaveToFile calls are properly delegated
func TestFileWriterAdapter_SaveToFile(t *testing.T) {
	adapter, mock := setupFileWriterAdapterTest()

	// Set up expected return value
	expectedErr := errors.New("test error")
	mock.SaveToFileFunc = func(ctx context.Context, content, outputFile string) error {
		return expectedErr
	}

	// Create test inputs
	ctx := context.Background()
	content := "test content"
	outputFile := "test-output.txt"

	// Call the adapter method
	err := adapter.SaveToFile(ctx, content, outputFile)

	// Verify that the adapter delegated the call correctly
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Verify that the mock was called with the expected parameters
	if len(mock.SaveToFileCalls) != 1 {
		t.Fatalf("Expected 1 call to SaveToFile, got %d", len(mock.SaveToFileCalls))
	}
	call := mock.SaveToFileCalls[0]
	if call.Content != content || call.OutputFile != outputFile {
		t.Errorf("Unexpected parameters: Content = %s, OutputFile = %s", call.Content, call.OutputFile)
	}
}

// Note: Tests for conversion functions (TestInternalToInterfacesGatherConfig,
// TestInternalToInterfacesContextStats, TestInterfacesToInternalContextStats) were removed
// as part of issue #121. These conversion functions are no longer needed since
// ContextStats, GatherConfig, and ContextGatherer are now defined once in interfaces/.
