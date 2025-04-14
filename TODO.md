# OpenAI Provider Support Task Breakdown

This document outlines the tasks needed to implement OpenAI provider support in the Architect tool.

## Tasks

### Task ID: TASK-001 [x]

**Title**: Define Generic LLM Interface

**Type**: Development

**Description**:
Create `internal/llm/client.go` with the `LLMClient` interface, `ProviderResult`, `ProviderTokenCount`, and `ProviderModelInfo` structs. The interface should include methods for `GenerateContent`, `CountTokens`, `GetModelInfo`, `Close`, and `GetModelName`.

**Acceptance Criteria**:
- `internal/llm/client.go` exists
- `LLMClient` interface is defined with the specified methods
- `ProviderResult`, `ProviderTokenCount`, and `ProviderModelInfo` structs are defined
- The code compiles without errors

**Estimated Effort**: Low

**Depends On**: None

**Priority**: Must-Have

### Task ID: TASK-002 [x]

**Title**: Refactor Gemini Client to Implement LLMClient Interface

**Type**: Development

**Description**:
Modify `internal/gemini/gemini_client.go` to implement the `llm.LLMClient` interface. Map internal Gemini types to the new generic types defined in `internal/llm/client.go`.

**Acceptance Criteria**:
- `internal/gemini/gemini_client.go` implements the `llm.LLMClient` interface
- Internal Gemini types are correctly mapped to generic types
- Existing Gemini functionality remains unaffected
- The code compiles without errors

**Estimated Effort**: Medium

**Depends On**: TASK-001

**Priority**: Must-Have

### Task ID: TASK-003 [x]

**Title**: Add Official OpenAI Go Package Dependency

**Type**: Infrastructure

**Description**:
Run `go get github.com/openai/openai-go` to add the official OpenAI Go package as a dependency.

**Acceptance Criteria**:
- The `go.mod` file is updated with the official OpenAI package
- The `go.sum` file is updated
- The command executes successfully

**Estimated Effort**: Low

**Depends On**: None

**Priority**: Must-Have

### Task ID: TASK-004 [x]

**Title**: Implement OpenAI Client

**Type**: Development

**Description**:
Create `internal/openai/openai_client.go` implementing the `llm.LLMClient` interface. Use the official OpenAI Go client (`github.com/openai/openai-go`) for initialization and API calls. Implement `GenerateContent` using OpenAI's chat completion API. Implement `CountTokens` using the official client's token counting capabilities. Implement `GetModelInfo` with hardcoded limits for common models. Implement remaining interface methods.

**Acceptance Criteria**:
- `internal/openai/openai_client.go` exists
- The code implements the `llm.LLMClient` interface
- `GenerateContent`, `CountTokens`, `GetModelInfo`, `Close`, and `GetModelName` methods are implemented
- The code compiles without errors
- OpenAI API key is retrieved from the `OPENAI_API_KEY` environment variable

**Estimated Effort**: High

**Depends On**: TASK-001, TASK-003

**Priority**: Must-Have

### Task ID: TASK-005 [~]

**Title**: Create OpenAI Client Tests

**Type**: Testing

**Description**:
Create unit tests for the OpenAI client in `internal/openai/openai_client_test.go`. Test client initialization and all interface methods. Test error handling specific to OpenAI.

**Acceptance Criteria**:
- `internal/openai/openai_client_test.go` exists
- Unit tests cover client initialization and all interface methods
- Error handling specific to OpenAI is tested
- Tests pass successfully

**Estimated Effort**: Medium

**Depends On**: TASK-004

**Priority**: Must-Have

### Task ID: TASK-006

**Title**: Update API Service for Provider Detection

**Type**: Development

**Description**:
Modify `internal/architect/api.go` to detect the provider based on the model name. Change `InitClient` to return `llm.LLMClient` interface. Update error handling for both providers.

**Acceptance Criteria**:
- `internal/architect/api.go` is modified to detect the provider based on the model name
- `InitClient` returns `llm.LLMClient` interface
- Error handling is updated for both Gemini and OpenAI providers
- The code compiles without errors

**Estimated Effort**: Medium

**Depends On**: TASK-002, TASK-004

**Priority**: Must-Have

### Task ID: TASK-007 [x]

**Title**: Update Configuration for OpenAI API Key

**Type**: Development

**Description**:
Add `OpenAIAPIKeyEnvVar` constant to `internal/config/config.go`. Update CLI validation to check for required API keys based on the model.

**Acceptance Criteria**:
- `internal/config/config.go` includes the `OpenAIAPIKeyEnvVar` constant
- CLI validation checks for required API keys based on the model
- The code compiles without errors

**Estimated Effort**: Low

**Depends On**: None

**Priority**: Must-Have

### Task ID: TASK-008

**Title**: Update Orchestrator and ModelProcessor

**Type**: Development

**Description**:
Modify `internal/architect/orchestrator/orchestrator.go` and `internal/architect/modelproc/processor.go` to use the `llm.LLMClient` interface. Update method calls and type handling.

**Acceptance Criteria**:
- Both components use the `llm.LLMClient` interface
- Method calls and type handling are updated accordingly
- The code compiles without errors

**Estimated Effort**: Medium

**Depends On**: TASK-006

**Priority**: Must-Have

### Task ID: TASK-009

**Title**: Update Documentation

**Type**: Documentation

**Description**:
Add OpenAI support details to `README.md`. Update model flag description to include OpenAI examples.

**Acceptance Criteria**:
- `README.md` includes details about OpenAI support
- Model flag description includes OpenAI examples
- Documentation is clear and accurate

**Estimated Effort**: Low

**Depends On**: TASK-008

**Priority**: Should-Have

### Task ID: TASK-010

**Title**: Update Existing Tests for Interface Changes

**Type**: Testing

**Description**:
Update existing tests to accommodate interface changes. Modify test mocks and fixtures to work with the new `llm.LLMClient` interface.

**Acceptance Criteria**:
- Existing tests pass after interface changes
- Tests cover the updated functionality
- Test mocks are updated to implement the new interface

**Estimated Effort**: Medium

**Depends On**: TASK-008

**Priority**: Must-Have

### Task ID: TASK-011

**Title**: Add Integration Tests for Multi-Provider Support

**Type**: Testing

**Description**:
Create integration tests that verify the multi-provider functionality. Test specifying OpenAI models in configuration. Test concurrent processing of both Gemini and OpenAI models. Test output file handling for different providers.

**Acceptance Criteria**:
- Integration tests cover specifying OpenAI models
- Integration tests verify Gemini functionality remains intact
- Tests for concurrent processing with multiple providers
- Tests for API key validation based on provider
- All tests pass successfully

**Estimated Effort**: Medium

**Depends On**: TASK-010

**Priority**: Must-Have

## Task Summary

| Task Type      | Count |
|----------------|-------|
| Development    | 6     |
| Testing        | 3     |
| Documentation  | 1     |
| Infrastructure | 1     |
| **Total**      | **11**|

| Priority       | Count |
|----------------|-------|
| Must-Have      | 10    |
| Should-Have    | 1     |
| Nice-to-Have   | 0     |
| **Total**      | **11**|

**Critical Path**: TASK-001 → TASK-002 → TASK-006 → TASK-008 → TASK-010 → TASK-011

## Clarifications Needed

1. **Token Counting Approach**: Need to investigate how the official OpenAI Go package (`github.com/openai/openai-go`) handles token counting and ensure it meets our needs.

2. **Model Information**: Token limits for OpenAI models may need to be hardcoded initially. Need to decide on a long-term approach for retrieving model information dynamically.

3. **Error Mapping**: Need to decide how to map OpenAI-specific errors from the official client to consistent error types that can be handled uniformly across providers.

4. **Testing Strategy**: Determine the best approach for mocking the official OpenAI Go client in tests, particularly for integration tests that verify multi-provider functionality.

5. **API Compatibility**: Verify that the official OpenAI Go package provides all the functionality we need, particularly for token counting and model information retrieval.
