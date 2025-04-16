# TODO

## LLM Provider Configuration Refactor

- [x] **T001:** Define Core Configuration Structs
    - **Action:** Define `ProviderDefinition`, `ModelDefinition`, and `ParameterDefinition` Go structs in a suitable package (e.g., `internal/config` or `internal/registry`). Ensure correct YAML tags (`yaml:"..."`) are added for parsing the `models.yaml` file. Define types clearly (e.g., for parameters).
    - **Depends On:** None
    - **AC Ref:** AC1, AC5, AC6

- [x] **T002:** Implement YAML Configuration Loading
    - **Action:** Add `gopkg.in/yaml.v3` dependency. Implement logic to read the models.yaml file from a single, consistent location (`~/.config/architect/models.yaml`). Implement robust error handling for file not found, invalid YAML syntax, and missing required fields. Store the parsed configuration data.
    - **Depends On:** T001
    - **AC Ref:** AC1

- [ ] **T003:** Create Registry Package and Core Logic
    - **Action:** Create the `internal/registry` package. Implement the `Registry` struct to hold loaded `ProviderDefinition` and `ModelDefinition` maps. Implement methods: `LoadConfig()`, `GetModel(name string) (*ModelDefinition, error)`, `GetProvider(name string) (*ProviderDefinition, error)`, and `RegisterProviderImplementation(name string, impl providers.Provider)`. Implement `GetProviderImplementation(name string) (providers.Provider, error)`.
    - **Depends On:** T001, T002
    - **AC Ref:** AC1, AC2, AC3, AC4, AC5, AC9

- [ ] **T004:** Define Provider Interface
    - **Action:** Create `internal/providers/provider.go` (or a similar path). Define the `Provider` interface with the method `CreateClient(ctx context.Context, apiKey string, modelID string, apiEndpoint string) (llm.LLMClient, error)`. Ensure the `llm.LLMClient` interface from `internal/llm/client.go` is used as the return type.
    - **Depends On:** None
    - **AC Ref:** AC3

- [ ] **T005:** Refactor Gemini Implementation for Registry
    - **Action:** Create `internal/providers/gemini/provider.go`. Implement the `providers.Provider` interface. Ensure the existing `gemini.NewLLMClient` (or a new factory function) returns the `llm.LLMClient` interface. Modify `geminiClient.GenerateContent` to accept and utilize a `map[string]interface{}` for parameters based on `ModelDefinition`. Update `geminiClient.GetModelInfo` to use configuration data from models.yaml. Remove any Gemini-specific provider detection logic from `internal/architect/api.go` or other central places.
    - **Depends On:** T004
    - **AC Ref:** AC3, AC4, AC6, AC9

- [ ] **T006:** Refactor OpenAI Implementation for Registry
    - **Action:** Create `internal/providers/openai/provider.go`. Implement the `providers.Provider` interface. Ensure the existing `openai.NewClient` (or equivalent) returns the `llm.LLMClient` interface. Modify `openaiClient.GenerateContent` to accept and utilize a `map[string]interface{}` for parameters based on `ModelDefinition`. Update `openaiClient.GetModelInfo` to use configuration data (token limits) from models.yaml. Remove any OpenAI-specific provider detection logic from `internal/architect/api.go` or other central places.
    - **Depends On:** T004
    - **AC Ref:** AC3, AC4, AC6, AC9

- [ ] **T007:** Integrate Registry into Core Application Logic
    - **Action:** In the application startup sequence (`internal/architect/app.go` or `main.go`), initialize and load the registry using `registry.LoadConfig`. Register the Gemini (T005) and OpenAI (T006) provider implementations using `registry.RegisterProviderImplementation`. Refactor the client creation logic (likely within the orchestrator or main execution flow): remove calls to `DetectProviderFromModel`, use `registry.GetModel` to get `ModelDefinition`, use `registry.GetProvider` to get `ProviderDefinition`, retrieve the API key using the `api_key_sources` map from the config and the `ProviderDefinition`, get the provider implementation using `registry.GetProviderImplementation`, and call `provider.CreateClient` to get the `llm.LLMClient`. Assess and refactor/remove the existing `APIService` interface/implementation if it becomes redundant.
    - **Depends On:** T003, T005, T006
    - **AC Ref:** AC2, AC3, AC4, AC5, AC9

- [ ] **T008:** Implement Model Parameter Handling
    - **Action:** In the client interaction logic (likely orchestrator or model processor), retrieve parameters from the `ModelDefinition.Parameters`. Use these parameters directly without CLI overrides. Update the `llm.LLMClient.GenerateContent` interface (if necessary) and implementations (T005, T006) to accept a `map[string]interface{}` parameter map and pass relevant parameters to the underlying API calls.
    - **Depends On:** T001, T007
    - **AC Ref:** AC6

- [ ] **T009:** Update Token Limit Logic
    - **Action:** Identify all code locations that check or use token limits (e.g., `ContextGatherer`, `ModelProcessor`/`Orchestrator`). Modify this logic to retrieve `ContextWindow` and `MaxOutputTokens` from the `ModelDefinition` obtained via the Registry (T003, T007). Remove any hardcoded token limits related to specific models.
    - **Depends On:** T001, T007
    - **AC Ref:** AC5

- [ ] **T010:** Create Default `models.yaml` File
    - **Action:** Create a well-structured default `models.yaml` file for the `~/.config/architect/` directory. Include common Gemini models (e.g., `gemini-1.5-pro`) and OpenAI models (e.g., `gpt-4-turbo`, `gpt-3.5-turbo`, `o3-mini`). Populate fields like `provider`, `api_model_id`, `context_window`, `max_output_tokens`, and define common `parameters` (like `temperature`) with types and defaults. Add comments explaining the structure and how to customize it. Include this file in the repository with installation instructions to copy it to the user's config directory.
    - **Depends On:** T001
    - **AC Ref:** AC1

- [ ] **T011:** Update Documentation
    - **Action:** Update `README.md` to explain the new `models.yaml` configuration system: its purpose, its location (`~/.config/architect/models.yaml`), structure (referencing T010), and how to add/modify providers and models. Include installation instructions to create the config directory and install the default models.yaml file. Remove documentation related to the old provider detection logic and any deprecated flags.
    - **Depends On:** T002, T010
    - **AC Ref:** AC12

- [ ] **T012:** Add/Update Tests
    - **Action:** Write comprehensive unit tests for the `internal/registry` package, covering config loading, definition lookups, and implementation registration/retrieval. Write unit tests for the YAML parsing logic (T002), including error handling for invalid formats or missing files. Update existing unit/integration tests for components like `ContextGatherer` and `ModelProcessor`/`Orchestrator` to use the new registry/config-based approach for client creation and limit checking. Ensure integration tests cover scenarios using both Gemini and OpenAI models defined in a test `models.yaml`. Verify all existing tests pass after the refactoring.
    - **Depends On:** T003, T005, T006, T007, T008, T009
    - **AC Ref:** AC9, AC10, AC11

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Issue/Assumption:** `APIService` Refactoring Scope
    - **Context:** Task T7 mentions potentially refactoring or removing the existing `APIService` interface/implementation (`internal/architect/api.go`). The extent of this refactoring needs clarification.
    - **Assumption:** The primary goal is to replace the client creation and provider detection logic within `APIService` with the new registry/provider pattern. The interface might be kept for response processing helpers, or those helpers might be moved elsewhere.

- [ ] **Issue/Assumption:** Parameter Constraints
    - **Context:** The `ParameterDefinition` struct (T001) should handle parameter types. Should it also support defining and validating constraints (e.g., min/max values for numbers, allowed enum values for strings)?
    - **Assumption:** Initially, only type validation (float, int, string) will be implemented. Constraint validation can be added later if required.
