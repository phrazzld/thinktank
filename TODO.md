# TODO

## Provider-Agnostic Architecture Migration Tasks

- [x] **Task Title:** Remove deprecated methods from APIService interface
  - **Action:** Update `internal/architect/interfaces/interfaces.go` to remove the deprecated `InitClient` and `ProcessResponse` methods from the APIService interface definition.
  - **Depends On:** None
  - **AC Ref:** N/A

- [x] **Task Title:** Remove deprecated method implementations from API service
  - **Action:** Update `internal/architect/api.go` to remove the deprecated `InitClient` and `ProcessResponse` method implementations and all references to the compatibility package.
  - **Depends On:** Remove deprecated methods from APIService interface
  - **AC Ref:** N/A

- [x] **Task Title:** Delete the compatibility package
  - **Action:** Remove the entire `internal/architect/compat` directory containing `compat.go` and `compat_test.go`.
  - **Depends On:** Remove deprecated method implementations from API service
  - **AC Ref:** N/A

- [x] **Task Title:** Update modelproc package to use provider-agnostic methods
  - **Action:** Refactor `internal/architect/modelproc/processor.go` and `internal/architect/modelproc/processor_llm_test.go` to use only provider-agnostic methods like `InitLLMClient` and `ProcessLLMResponse`.
  - **Depends On:** None
  - **AC Ref:** N/A

- [x] **Task Title:** Update adapter tests to use provider-agnostic methods
  - **Action:** Refactor `internal/architect/api_adapter_test.go` and `internal/architect/adapters_test.go` to use only provider-agnostic methods and remove references to deprecated methods. This is required for the build to pass, as the tests currently refer to methods that have been removed.
  - **Depends On:** None
  - **AC Ref:** N/A

- [x] **Task Title:** Update mocks in test files to remove deprecated methods
  - **Action:** Update all mock implementations in test files (like `internal/architect/modelproc/mocks_test.go`) to remove deprecated methods and use only provider-agnostic interfaces.
  - **Depends On:** Remove deprecated methods from APIService interface
  - **AC Ref:** N/A

- [x] **Task Title:** Update orchestrator tests to use provider-agnostic methods
  - **Action:** Refactor all orchestrator test files in `internal/architect/orchestrator/` that use deprecated methods to use the provider-agnostic alternatives.
  - **Depends On:** None
  - **AC Ref:** N/A

- [x] **Task Title:** Update integration tests to use provider-agnostic methods
  - **Action:** Refactor integration tests in `internal/integration/` directory that use deprecated methods, especially `multi_model_test.go` and `test_runner.go`.
  - **Depends On:** None
  - **AC Ref:** N/A

- [x] **Task Title:** Simplify gemini package to focus on LLM interface
  - **Action:** Update `internal/gemini/gemini_client.go` to focus on the LLM interface implementation and remove redundant methods that were only needed for the compatibility layer.
  - **Depends On:** Remove deprecated methods from APIService interface
  - **AC Ref:** N/A

- [x] **Task Title:** Remove all remaining references to deprecated methods
  - **Action:** Run a project-wide search for any remaining references to `InitClient` and `ProcessResponse` and update them to use provider-agnostic alternatives.
  - **Depends On:** Update all other tasks above
  - **AC Ref:** N/A

- [x] **Task Title:** Update app.Execute to use provider-agnostic error handling
  - **Action:** Refactor error handling in `internal/architect/app.go` to ensure it leverages the provider-agnostic error utilities and doesn't rely on any Gemini-specific error handling.
  - **Depends On:** None
  - **AC Ref:** N/A

- [x] **Task Title:** Clean up imports across the codebase
  - **Action:** Remove unnecessary imports of the gemini package in files that only need the provider-agnostic interfaces, replacing them with imports of the llm package where appropriate.
  - **Depends On:** All other tasks
  - **AC Ref:** N/A
