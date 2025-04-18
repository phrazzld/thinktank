# TODO

## Implement OpenRouter Provider Support

- [x] **T001:** Update `config/models.yaml` for OpenRouter
    - **Action:** Add `"openrouter"` to the `providers` list. Add `"openrouter": "OPENROUTER_API_KEY"` to `api_key_sources`. Add example OpenRouter model definitions (e.g., `openrouter/deepseek/deepseek-chat-v3-0324`, `openrouter/deepseek/deepseek-r1`, `openrouter/x-ai/grok-3-beta`) under the `models` list, ensuring `provider: openrouter` and correct `api_model_id` format are used. Include necessary parameters like `context_window`, `max_output_tokens`, and `temperature`.
    - **Depends On:** None
    - **AC Ref:** PLAN.md Step 1

- [ ] **T002:** Update Configuration Documentation and Scripts
    - **Action:** Modify `config/README.md` to mention the new "openrouter" provider and the `OPENROUTER_API_KEY` source in `models.yaml`. Update `config/install.sh` to check for and mention the `OPENROUTER_API_KEY` environment variable during setup.
    - **Depends On:** [T001]
    - **AC Ref:** PLAN.md Step 1

- [ ] **T003:** Create OpenRouter Provider Package Structure
    - **Action:** Create the directory `internal/providers/openrouter`. Create the file `internal/providers/openrouter/provider.go`. Create the file `internal/providers/openrouter/client.go`.
    - **Depends On:** None
    - **AC Ref:** PLAN.md Step 2

- [ ] **T004:** Implement `openrouter.Provider` Struct and `CreateClient` Method
    - **Action:** In `provider.go`, define the `OpenRouterProvider` struct. Implement the `providers.Provider` interface for this struct, specifically the `CreateClient(ctx context.Context, apiKey string, modelID string, apiEndpoint string) (llm.LLMClient, error)` method. This method should instantiate and return an `openrouterClient` instance, handling the API key and using the default OpenRouter API endpoint (`https://openrouter.ai/api/v1`) if `apiEndpoint` is empty.
    - **Depends On:** [T003]
    - **AC Ref:** PLAN.md Step 3

- [ ] **T005:** Define `openrouter.Client` Struct and Constructor
    - **Action:** In `client.go`, define the `openrouterClient` struct that will implement the `llm.LLMClient` interface. Create a `NewClient(apiKey string, modelID string, apiEndpoint string) *openrouterClient` constructor function within the `openrouter` package. The constructor should initialize the client with necessary fields like the API key, model ID, API endpoint, and an HTTP client.
    - **Depends On:** [T003]
    - **AC Ref:** PLAN.md Step 4

- [ ] **T006:** Implement `GenerateContent` Success Path in `openrouter.Client`
    - **Action:** Implement the `GenerateContent` method on `openrouterClient`. Use `net/http` to make a POST request to the `/chat/completions` endpoint. Set required headers (`Authorization: Bearer <API_KEY>`, `Content-Type: application/json`). Marshal the request body based on the OpenAI-compatible format using the client's `modelID` and the method's `prompt` and `params`. Unmarshal the JSON response and map the primary content to `llm.ProviderResult.Content`. Handle the basic success case (HTTP 200 OK).
    - **Depends On:** [T005]
    - **AC Ref:** PLAN.md Step 4

- [ ] **T007:** Implement `FormatAPIError` Function for OpenRouter
    - **Action:** Create a `FormatAPIError` function within the `openrouter` package, similar to existing error handling functions. This function should take an error and potentially an HTTP status code, analyze the error (e.g., type assertion, string matching for common OpenRouter/OpenAI error formats), and return an `llm.CategorizedError`.
    - **Depends On:** [T006]
    - **AC Ref:** PLAN.md Step 4

- [ ] **T008:** Implement Error Handling in `GenerateContent`
    - **Action:** Enhance the `GenerateContent` method to handle API errors and non-200 HTTP status codes. Use the `FormatAPIError` function (T007) to categorize errors and return them appropriately wrapped. Map response fields (finish reason, token counts if available in error response) to `llm.ProviderResult`.
    - **Depends On:** [T006, T007]
    - **AC Ref:** PLAN.md Step 4

- [ ] **T009:** Implement `CountTokens` in `openrouter.Client`
    - **Action:** Implement the `CountTokens` method. Since OpenRouter lacks a dedicated token counting endpoint, use the `tiktoken-go` library. Add logic to parse the `modelID` (e.g., "anthropic/claude-3.5-sonnet") to determine the underlying model family and select the appropriate encoding (e.g., `cl100k_base`). Return the token count in the `llm.ProviderTokenCount` struct.
    - **Depends On:** [T005]
    - **AC Ref:** PLAN.md Step 4

- [ ] **T010:** Implement `GetModelInfo` in `openrouter.Client`
    - **Action:** Implement the `GetModelInfo` method. Fetch model limits (context window, max output tokens) from the `Registry` using the `modelID` stored in the client. If registry information is missing, fall back to reasonable default values (e.g., a large context window like 128k and output limit like 4k). Return the information in the `llm.ProviderModelInfo` struct.
    - **Depends On:** [T001, T005]
    - **AC Ref:** PLAN.md Step 4

- [ ] **T011:** Implement `GetModelName` in `openrouter.Client`
    - **Action:** Implement the `GetModelName` method. Return the `modelID` that was provided during the client's creation.
    - **Depends On:** [T005]
    - **AC Ref:** PLAN.md Step 4

- [ ] **T012:** Implement `Close` Method in `openrouter.Client`
    - **Action:** Implement the `Close` method. If using the standard `net/http` client without custom configurations requiring cleanup, this can be a no-op.
    - **Depends On:** [T005]
    - **AC Ref:** PLAN.md Step 4

- [ ] **T013:** Register OpenRouter Provider Implementation
    - **Action:** Modify the `registerProviders` function in `internal/registry/manager.go`. Import the `internal/providers/openrouter` package. Instantiate the `OpenRouterProvider` using its constructor (e.g., `openrouter.NewProvider(m.logger)`). Register the instance with the registry using `m.registry.RegisterProviderImplementation("openrouter", openRouterProvider)`. Add appropriate logging.
    - **Depends On:** [T004]
    - **AC Ref:** PLAN.md Step 5

- [ ] **T014:** Write Unit Tests for `openrouter.Client`
    - **Action:** Create `internal/providers/openrouter/client_test.go`. Use `net/http/httptest` to mock the OpenRouter API. Write tests to verify: correct request formation (URL, method, headers, body) for `GenerateContent`; correct response parsing for success cases; correct error handling for various API errors/status codes; correct token counting logic in `CountTokens`; correct model info retrieval in `GetModelInfo`.
    - **Depends On:** [T006, T008, T009, T010, T011, T012]
    - **AC Ref:** PLAN.md Step 6

- [ ] **T015:** Write Unit Tests for `openrouter.Provider`
    - **Action:** Create `internal/providers/openrouter/provider_test.go`. Write tests to verify that `CreateClient` returns a non-nil client implementing `llm.LLMClient` and handles API keys and model IDs correctly.
    - **Depends On:** [T004]
    - **AC Ref:** PLAN.md Step 6

- [ ] **T016:** Add Integration Tests for OpenRouter
    - **Action:** Add or modify tests in `internal/integration` to include scenarios using OpenRouter models specified via CLI flags (e.g., `--model openrouter/google/gemini-1.5-pro`). Ensure these tests use the existing mock server infrastructure but validate that requests are correctly routed through the new OpenRouter provider implementation.
    - **Depends On:** [T001, T013, T014, T015]
    - **AC Ref:** PLAN.md Step 6

- [ ] **T017:** Update Main `README.md` Documentation
    - **Action:** Modify the main `README.md`. Add OpenRouter to the list of supported providers. Explain the requirement for the `OPENROUTER_API_KEY` environment variable. Provide examples of using OpenRouter models via the `--model` flag.
    - **Depends On:** [T016]
    - **AC Ref:** PLAN.md Step 7

- [ ] **T018:** Update Configuration `README.md` Documentation
    - **Action:** Modify `config/README.md` to explain the new `openrouter` provider section and the `OPENROUTER_API_KEY` source added to `models.yaml`.
    - **Depends On:** [T001, T016]
    - **AC Ref:** PLAN.md Step 7

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
- [ ] **Assumption:** The `tiktoken-go` library is already available as a project dependency, likely via `openai-go`.
    - **Context:** PLAN.md Step 4 mentions using `tiktoken-go` for `CountTokens`.
- [ ] **Assumption:** The existing error handling pattern (`FormatAPIError`) used for other providers is suitable for adaptation to OpenRouter's OpenAI-compatible error responses.
    - **Context:** PLAN.md Step 4 suggests creating a new `FormatAPIError` function similar to existing ones.
- [ ] **Assumption:** Optional OpenRouter headers (`HTTP-Referer`, `X-Title`) are not required for the initial implementation and will be added later if needed.
    - **Context:** PLAN.md Step 4 mentions these headers and suggests making them configurable later. Task T006 omits them for the initial implementation.
- [ ] **Assumption:** The `GetModelInfo` implementation should prioritize fetching limits from the registry config (`models.yaml`) and use hardcoded defaults only as a fallback.
    - **Context:** PLAN.md Step 4 mentions "registry-derived limits" and "fallback to reasonable defaults". Task T010 implements this logic.
- [ ] **Assumption:** The project structure and interfaces (`internal/providers`, `internal/registry`, `providers.Provider`, `llm.LLMClient`) exist and are stable as described in the PLAN.md.
    - **Context:** The entire plan relies on the existing registry pattern and interfaces.
