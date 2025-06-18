# Registry Elimination Implementation Tasks

## Phase 1: Foundation - Create Models Package

### Model Definition Tasks
- [x] Extract current model metadata from `config/models.yaml` and document all 7 models with their exact configuration values including provider, API model ID, context window, max output tokens, and default parameters
- [x] Create directory structure `internal/models/` with files `models.go`, `models_test.go`, and `doc.go` following Go package conventions
- [x] Define `ModelInfo` struct in `internal/models/models.go` with fields: Provider (string), APIModelID (string), ContextWindow (int), MaxOutputTokens (int), DefaultParams (map[string]interface{})
- [x] Implement hardcoded `modelDefinitions` map[string]ModelInfo with entries for all 7 models: gpt-4.1, o4-mini, gemini-2.5-pro, gemini-2.5-flash, and 3 openrouter models extracted from current registry

### Core Function Implementation
- [x] Implement `GetModelInfo(name string) (ModelInfo, error)` that returns model metadata from modelDefinitions map with proper error handling for unknown models
- [x] Implement `GetProviderForModel(name string) (string, error)` that extracts provider from model info with validation that model exists
- [x] Implement `ListAllModels() []string` that returns sorted slice of all model names from modelDefinitions map
- [x] Implement `ListModelsForProvider(provider string) []string` that filters and returns models matching given provider
- [x] Implement `GetAPIKeyEnvVar(provider string) string` with switch statement mapping providers (openai, gemini, openrouter) to environment variables
- [x] Implement `IsModelSupported(name string) bool` that checks if model exists in modelDefinitions map

### Unit Testing Tasks
- [ ] Write table-driven test for `GetModelInfo` covering all 7 valid models plus 3 invalid model names with appropriate error assertions
- [ ] Write test for `GetProviderForModel` verifying correct provider extraction for each model and error handling for unknown models
- [ ] Write test for `ListAllModels` verifying all 7 models are returned in alphabetical order
- [ ] Write test for `ListModelsForProvider` testing each provider (openai, gemini, openrouter) returns correct model subsets
- [ ] Write test for `GetAPIKeyEnvVar` verifying correct environment variable names for all providers plus empty string for unknown provider
- [ ] Write test for `IsModelSupported` checking true for all 7 models and false for non-existent models
- [ ] Verify test coverage reaches 90%+ using `go test -cover ./internal/models` and document coverage percentage

## Phase 2: Service Integration - Update Registry API

### Registry API Refactoring
- [ ] Create backup of current `internal/thinktank/registry_api.go` as `registry_api.go.backup` for reference during refactoring
- [ ] Update imports in `registry_api.go` to use `internal/models` package instead of `internal/registry` package
- [ ] Replace `registryManager` field in `RegistryAPIService` struct with direct usage of models package functions
- [ ] Update `NewRegistryAPIService` constructor to remove registry manager parameter and initialization logic

### InitLLMClient Simplification
- [ ] Refactor `InitLLMClient` method to call `models.GetModelInfo(modelName)` instead of registry lookup with proper error propagation
- [ ] Simplify provider detection logic to use model info's Provider field directly without registry indirection
- [ ] Update API key resolution to call `models.GetAPIKeyEnvVar(provider)` and validate key exists in environment
- [ ] Replace provider registry lookup with direct client instantiation based on provider name using switch statement
- [ ] Update OpenRouter client creation to hardcode base URL "https://openrouter.ai/api/v1" when provider is "openrouter"

### Method Updates
- [ ] Update `GetModelParameters` to use `models.GetModelInfo` and return DefaultParams field with proper error handling
- [ ] Simplify `ValidateModelParameter` to perform basic range validation for temperature (0.0-2.0) without complex registry rules
- [ ] Update `GetModelTokenLimits` to return ContextWindow and MaxOutputTokens from model info structure
- [ ] Update `ListModels` to directly call `models.ListAllModels()` without registry wrapper
- [ ] Update `ListProviderModels` to call `models.ListModelsForProvider(provider)` directly

### Integration Testing
- [ ] Write integration test creating LLM clients for all 7 models with mock API keys verifying correct client type returned
- [ ] Test API key resolution for each provider verifying correct environment variable is checked and error on missing key
- [ ] Test parameter validation focusing on temperature limits (0.0-2.0) with edge cases
- [ ] Test token limit retrieval for each model verifying correct context window and max output values
- [ ] Run existing registry API integration tests to ensure no behavioral regression

## Phase 3: CLI Migration - Update Application Layer

### CLI Command Updates
- [ ] Update `cmd/thinktank/cli.go` imports to remove registry package and add models package import
- [ ] Remove all references to `registry.GetGlobalManager()` and related initialization code from CLI setup
- [ ] Update model validation in CLI to use `models.IsModelSupported(modelName)` for --model flag validation
- [ ] Update synthesis model validation to use `models.IsModelSupported(synthesisModel)` when --synthesis flag is provided
- [ ] Refactor provider detection for API key validation to use `models.GetProviderForModel(modelName)` directly

### App Layer Updates
- [ ] Update `internal/thinktank/app.go` to remove registry manager field from App struct
- [ ] Remove registry manager parameter from NewApp constructor and update all call sites
- [ ] Update app initialization to create RegistryAPIService without registry manager dependency
- [ ] Verify all model lookups in app layer now use models package functions directly
- [ ] Remove any remaining registry-related initialization or lifecycle code

### CLI Validation Testing
- [ ] Test CLI with valid model names for all 7 supported models verifying acceptance
- [ ] Test CLI with invalid model names verifying appropriate error messages
- [ ] Test synthesis flag with valid and invalid model names verifying validation works
- [ ] Test API key validation ensures correct environment variables are checked for each model's provider
- [ ] Run manual CLI commands for each model type with --dry-run flag to verify no behavioral changes

## Phase 4: System Cleanup - Remove Registry Package

### File Deletion Tasks
- [ ] Delete entire `internal/registry/` directory after verifying no active imports remain using `grep -r "internal/registry" .`
- [ ] Delete `config/models.yaml` file and any other YAML model configuration files
- [ ] Remove registry configuration loading from `config/install.sh` if present
- [ ] Delete any registry-specific test fixtures or test data files
- [ ] Remove registry package from go.mod if listed as internal dependency

### Import Cleanup
- [ ] Run `grep -r "registry" --include="*.go" .` to find all registry import statements and remove them
- [ ] Use `goimports -w .` to clean up and organize all Go imports after registry removal
- [ ] Search for and remove any registry-related constants or type aliases in other packages
- [ ] Remove any registry-specific error types or error handling code
- [ ] Update any comments or documentation that reference the registry system

### Build Verification
- [ ] Run `go build ./...` to ensure no compilation errors after registry removal
- [ ] Run `go mod tidy` to clean up any unused dependencies
- [ ] Verify no references to registry remain using `grep -r "Registry" --include="*.go" . | grep -v "RegistryAPI"`
- [ ] Check for dead code using `staticcheck ./...` and remove any functions that only served registry

## Phase 5: Validation & Documentation

### Comprehensive Testing
- [ ] Run full test suite with `go test ./...` and verify all tests pass with no registry-related failures
- [ ] Run tests with race detection `go test -race ./...` to ensure no concurrency issues introduced
- [ ] Execute E2E test suite `./internal/e2e/run_e2e_tests.sh` verifying all scenarios pass
- [ ] Run coverage check `./scripts/check-coverage.sh` confirming 90%+ threshold maintained
- [ ] Perform manual testing of each model type with actual API calls (not dry-run) to verify functionality

### Documentation Updates
- [ ] Update `README.md` section on adding new models to explain editing `internal/models/models.go` directly
- [ ] Create `internal/models/README.md` documenting the models package structure and how to add new models
- [ ] Update `DEVELOPMENT.md` to remove references to registry configuration and YAML files
- [ ] Document the new model addition process: 1) Add to modelDefinitions map, 2) Run tests, 3) Submit PR
- [ ] Update any architecture diagrams or documentation that show registry as a component

### Performance Validation
- [ ] Measure application startup time before and after registry removal using `time go run cmd/thinktank/main.go --help`
- [ ] Compare binary size before and after using `go build -o before cmd/thinktank/main.go` and check with `ls -lh`
- [ ] Run informal benchmark of model lookup operations to verify O(1) performance maintained
- [ ] Document performance improvements in PR description with specific metrics

### Final Verification Checklist
- [ ] Verify line count reduction of ~10,800 lines using `git diff --stat` after all changes
- [ ] Confirm all 7 production models work identically to previous registry-based system
- [ ] Validate no user-facing behavior changes by comparing output of common commands
- [ ] Ensure no configuration files need deployment or management
- [ ] Create git tag `pre-registry-removal` on current branch for potential rollback reference
- [ ] Schedule team review session and document any concerns or questions raised
