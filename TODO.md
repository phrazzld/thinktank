# TODO

## Dependencies & Setup
- [x] Install Google Generative AI SDK
  - Description: Add the official Google Generative AI Node.js SDK package
  - Dependencies: None
  - Priority: High

- [x] Verify HTTP Client
  - Description: Ensure axios is available for the listModels HTTP request
  - Dependencies: None
  - Priority: High

## Configuration Updates
- [ ] Update Default Config File
  - Description: Add Google Gemini models to templates/thinktank.config.default.json
  - Dependencies: None
  - Priority: High

- [ ] Update Default Constants
  - Description: Add Google Gemini models to DEFAULT_CONFIG in src/atoms/constants.ts
  - Dependencies: None
  - Priority: High

- [ ] Update Environment Example
  - Description: Add GEMINI_API_KEY to .env.example
  - Dependencies: None
  - Priority: Medium

- [ ] Verify API Key Helper
  - Description: Ensure getDefaultApiKeyEnvVar('google') correctly returns "GEMINI_API_KEY"
  - Dependencies: None
  - Priority: Medium

## Provider Implementation
- [ ] Create Error Class
  - Description: Create GoogleProviderError class for error handling
  - Dependencies: None
  - Priority: High

- [ ] Create Provider Class
  - Description: Create GoogleProvider class implementing LLMProvider interface
  - Dependencies: Error class
  - Priority: High

- [ ] Implement Generate Method
  - Description: Add generate method using the Google GenAI SDK
  - Dependencies: Provider class
  - Priority: High

- [ ] Map Model Options
  - Description: Create mapOptions method to convert generic options to Gemini-specific parameters
  - Dependencies: Provider class, Generate method
  - Priority: High

- [ ] Implement ListModels Method
  - Description: Add listModels method using axios to call Gemini models API
  - Dependencies: Provider class, HTTP client
  - Priority: High

- [ ] Add Error Handling
  - Description: Implement comprehensive error handling for API failures
  - Dependencies: Provider class, all methods
  - Priority: Medium

## Provider Registration
- [ ] Create Default Provider Instance
  - Description: Export default googleProvider instance for auto-registration
  - Dependencies: Complete provider implementation
  - Priority: High

- [ ] Update Import in Templates
  - Description: Import Google provider in src/templates/runThinktank.ts
  - Dependencies: Provider implementation
  - Priority: High

- [ ] Update Import in ListModels
  - Description: Import Google provider in src/templates/listModelsWorkflow.ts
  - Dependencies: Provider implementation
  - Priority: High

## Testing
- [ ] Create Provider Test File
  - Description: Create src/molecules/llmProviders/__tests__/google.test.ts
  - Dependencies: None (test-first approach)
  - Priority: High

- [ ] Test Constructor & Registration
  - Description: Test proper registration and initialization
  - Dependencies: Test file
  - Priority: High

- [ ] Test Generate Method
  - Description: Test successful text generation and parameter mapping
  - Dependencies: Test file
  - Priority: High

- [ ] Test Error Handling
  - Description: Test various error scenarios like missing API key, API errors
  - Dependencies: Test file
  - Priority: Medium

- [ ] Test ListModels Method
  - Description: Test successful model listing and response parsing
  - Dependencies: Test file
  - Priority: High

- [ ] Update Integration Tests
  - Description: Update existing integration tests to cover Google provider
  - Dependencies: Complete implementation
  - Priority: Medium

- [ ] Manual E2E Testing
  - Description: Test with real API key, verify outputs and error handling
  - Dependencies: Complete implementation
  - Priority: Low

## Documentation
- [ ] Update README.md
  - Description: Document Google Gemini provider integration and usage
  - Dependencies: Complete implementation
  - Priority: Medium

## Assumptions and Questions

1. Provider ID: Using 'google' as the provider ID seems most appropriate for clarity.
2. Environment Variable Name: Using 'GEMINI_API_KEY' follows the pattern of other providers.
3. API Structure: The PLAN.md document provides sample code that includes specific paths and parameters, but the actual API might have variations or updates that require adjustment.
4. SDK Coverage: The Google GenAI SDK might not have a built-in method for listing models, requiring direct HTTP calls via axios.
5. Parameter Mapping: The mapping between thinktank's generic ModelOptions and Gemini's specific parameters might need refinement based on API documentation.
6. Token Counting: The Google API response structure for token usage metrics might differ from the example provided in PLAN.md.
7. Authentication: The API key is assumed to be used directly in API calls rather than through a more complex authentication flow.