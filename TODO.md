# TODO

## Core Interface Updates
- [x] Update LLMProvider Interface
  - Description: Add optional listModels method to LLMProvider in atoms/types.ts
  - Dependencies: None
  - Priority: High

- [x] Create LLMAvailableModel Interface
  - Description: Define interface for model information returned by providers
  - Dependencies: None
  - Priority: High

## Model Listing Output Formatter
- [x] Implement formatModelList Function
  - Description: Add function to format model lists from providers in outputFormatter.ts
  - Dependencies: LLMAvailableModel interface
  - Priority: High

## Anthropic Provider Integration
- [x] Install Anthropic SDK
  - Description: Add @anthropic-ai/sdk as a dependency with npm install
  - Dependencies: None
  - Priority: High

- [x] Create Anthropic Provider File
  - Description: Create anthropic.ts in molecules/llmProviders directory
  - Dependencies: Anthropic SDK
  - Priority: High

- [x] Implement generate Method
  - Description: Implement generate method in anthropic provider using Anthropic SDK
  - Dependencies: Anthropic Provider File
  - Priority: High

- [x] Implement listModels Method
  - Description: Implement listModels method to fetch available models from Anthropic API
  - Dependencies: Anthropic Provider File, LLMAvailableModel Interface
  - Priority: High

- [x] Register Anthropic Provider
  - Description: Ensure anthropic provider is registered in llmRegistry
  - Dependencies: Anthropic Provider Implementation
  - Priority: High

## Model Listing Workflow
- [x] Create listModelsWorkflow Template
  - Description: Create new template file for listing models functionality
  - Dependencies: LLMProvider Interface Update
  - Priority: Medium

- [x] Implement listAvailableModels Function
  - Description: Create main function to list models across providers
  - Dependencies: Updated LLMProvider Interface, formatModelList Function
  - Priority: Medium

## CLI Update
- [ ] Add list-models Command
  - Description: Update CLI to add the new list-models command with provider flag
  - Dependencies: listModelsWorkflow template
  - Priority: Medium

## Configuration Updates
- [ ] Update Default Config
  - Description: Add Anthropic examples to templates/thinktank.config.default.json
  - Dependencies: None
  - Priority: Medium

- [ ] Update Environment Example
  - Description: Add ANTHROPIC_API_KEY to .env.example if it exists
  - Dependencies: None
  - Priority: Medium

## Testing
- [ ] Unit Test Anthropic Provider
  - Description: Create tests for both generate and listModels methods 
  - Dependencies: Anthropic Provider Implementation
  - Priority: High
  - TDD: Implement these tests first before implementation

- [ ] Unit Test formatModelList
  - Description: Test function with various inputs (successful, errors, empty)
  - Dependencies: formatModelList Implementation
  - Priority: High
  - TDD: Implement these tests first before implementation

- [ ] Integration Test listModelsWorkflow
  - Description: Test the listing workflow with mocked components
  - Dependencies: listModelsWorkflow Implementation
  - Priority: Medium
  - TDD: Implement these tests first before implementation

- [ ] Integration Test CLI list-models Command
  - Description: Test the CLI command parsing and execution
  - Dependencies: CLI Update
  - Priority: Medium
  - TDD: Implement these tests first before implementation

- [ ] Manual E2E Testing
  - Description: Test all features with real API calls once implemented
  - Dependencies: All implementation complete
  - Priority: Low

## Documentation
- [ ] Update README.md
  - Description: Document new list-models command, Anthropic provider, and configuration
  - Dependencies: All implementation complete
  - Priority: Medium

## Optional Follow-up
- [ ] Implement listModels for OpenAI
  - Description: Add listModels method to OpenAI provider
  - Dependencies: Core Interface Updates
  - Priority: Low

- [ ] Refine Error Handling
  - Description: Improve handling of missing API keys and unsupported methods
  - Dependencies: listModelsWorkflow Implementation
  - Priority: Low