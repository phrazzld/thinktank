# TODO

## Provider Implementation
- [x] Create OpenRouterProvider Error Class
  - Description: Define an OpenRouterProviderError class extending Error
  - Dependencies: None
  - Priority: High

- [x] Create Base OpenRouterProvider Class
  - Description: Define the OpenRouterProvider class implementing LLMProvider interface
  - Dependencies: None
  - Priority: High
  
- [x] Implement Provider Constructor
  - Description: Create constructor with API key parameter and auto-registration
  - Dependencies: Base OpenRouterProvider class
  - Priority: High

- [ ] Configure OpenAI Client
  - Description: Set up OpenAI client with OpenRouter's baseURL and headers
  - Dependencies: Provider constructor
  - Priority: High

- [ ] Implement Generate Method
  - Description: Implement generate method using OpenAI SDK with OpenRouter endpoints
  - Dependencies: OpenAI client configuration
  - Priority: High

- [ ] Implement ListModels Method
  - Description: Add listModels method to fetch available models from OpenRouter
  - Dependencies: OpenAI client configuration
  - Priority: High

- [ ] Add Error Handling
  - Description: Implement comprehensive error handling for all API operations
  - Dependencies: Generate and ListModels methods
  - Priority: Medium

## Configuration Updates
- [ ] Update Default Configuration
  - Description: Add OpenRouter example model to thinktank.config.default.json
  - Dependencies: None
  - Priority: Medium

- [ ] Update Environment Example
  - Description: Add OPENROUTER_API_KEY to .env.example
  - Dependencies: None
  - Priority: Medium

## Integration
- [ ] Register Provider
  - Description: Import OpenRouter provider in runThinktank.ts
  - Dependencies: Working OpenRouterProvider implementation
  - Priority: High

- [ ] Verify ListModels Integration
  - Description: Ensure OpenRouter works with listModelsWorkflow.ts
  - Dependencies: Working OpenRouterProvider with listModels
  - Priority: Medium

## Testing
- [ ] Create Provider Unit Tests
  - Description: Create unit tests for OpenRouter provider functionality
  - Dependencies: OpenRouterProvider implementation
  - Priority: Medium

- [ ] Add RunThinktank Tests
  - Description: Add test cases for OpenRouter in runThinktank.test.ts
  - Dependencies: Provider registration
  - Priority: Medium

- [ ] Add ListModels Tests
  - Description: Add test cases for OpenRouter in listModelsWorkflow.test.ts
  - Dependencies: Provider registration
  - Priority: Medium

- [ ] Perform Manual Testing
  - Description: Test with real API key to verify end-to-end functionality
  - Dependencies: All implementation tasks
  - Priority: Low

## Documentation
- [ ] Update README
  - Description: Add OpenRouter configuration instructions to README.md
  - Dependencies: Working implementation
  - Priority: Medium

- [ ] Document Extension Example
  - Description: Add OpenRouter as an example in extending thinktank section
  - Dependencies: Working implementation
  - Priority: Low

## Assumptions and Questions

1. OpenRouter API is fully compatible with OpenAI SDK as stated in PLAN.md.
2. OpenAI SDK is already installed and properly configured in the project.
3. The OpenRouter API key follows the same pattern as other providers in the system.
4. The OpenRouter models follow the format `provider/model-id` (e.g., `openai/gpt-4o`).
5. The listModels endpoint returns data in a format similar to OpenAI's model listing, with possible minor differences.
6. No schema migrations or database changes are needed for this feature.
7. The implementation follows the Atomic Design pattern already established in the codebase.