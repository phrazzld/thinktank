# Implementation Plan: Eliminate Registry System

## Executive Summary

This plan details the elimination of the 10,803-line registry system in favor of a simple 100-line solution using hardcoded maps. This aligns with the GORDIAN simplification initiative and leyline principles of simplicity, YAGNI, and explicit over implicit design.

## Technical Approaches Analysis

### Approach 1: Direct Hardcoded Maps (SELECTED) ✅

**Philosophy Alignment**: Maximum simplicity, explicit behavior, zero configuration overhead

**Implementation**:
```go
// internal/models/models.go - Complete replacement for registry
package models

type ModelInfo struct {
    Provider        string
    APIModelID      string
    ContextWindow   int
    MaxOutputTokens int
    DefaultParams   map[string]interface{}
}

var modelDefinitions = map[string]ModelInfo{
    "gpt-4.1": {
        Provider:        "openai",
        APIModelID:      "gpt-4-turbo-2024-12-17",
        ContextWindow:   1000000,
        MaxOutputTokens: 200000,
        DefaultParams:   map[string]interface{}{"temperature": 0.7},
    },
    // ... other models
}

func GetModelInfo(name string) (ModelInfo, error) {
    if info, ok := modelDefinitions[name]; ok {
        return info, nil
    }
    return ModelInfo{}, fmt.Errorf("unknown model: %s", name)
}
```

**Pros**:
- Dead simple - anyone can understand in 30 seconds
- Zero initialization overhead
- No configuration files to manage
- Compile-time validation of model data
- Trivial to test

**Cons**:
- Requires code change to add models (but we rarely add models)
- No runtime configuration (but this wasn't actually used)

### Approach 2: Embedded JSON Configuration

**Philosophy Alignment**: Moderate simplicity, keeps some configuration separation

**Implementation**:
```go
//go:embed models.json
var modelsJSON []byte

var models map[string]ModelInfo

func init() {
    json.Unmarshal(modelsJSON, &models)
}
```

**Pros**:
- Separates data from code
- Still compile-time embedded
- JSON is simpler than YAML

**Cons**:
- Adds unmarshaling complexity
- Error handling in init()
- Still configuration to maintain
- Violates YAGNI - we don't need this flexibility

### Approach 3: Interface-Based Registry Lite

**Philosophy Alignment**: Over-engineered for the problem

**Implementation**:
```go
type ModelRegistry interface {
    GetModel(name string) (ModelInfo, error)
    ListModels() []string
}

type staticRegistry struct {
    models map[string]ModelInfo
}
```

**Pros**:
- Testable via interface
- Could swap implementations

**Cons**:
- Unnecessary abstraction
- Violates YAGNI principle
- Adds complexity without benefit

## Selected Approach: Direct Hardcoded Maps

Based on leyline principles and the actual usage patterns, Approach 1 is the clear winner. It embodies:
- **Simplicity**: 100 lines replaces 10,800 lines
- **Explicitness**: Model data is right there in the code
- **YAGNI**: No speculative flexibility
- **Maintainability**: Anyone can modify model data

## Architecture Blueprint

### New Structure
```
internal/
├── models/
│   ├── models.go      # All model definitions and lookup functions
│   └── models_test.go # Simple tests
├── apikey/
│   └── resolver.go    # API key resolution logic (simplified)
└── thinktank/
    └── registry_api.go # Updated to use models package
```

### Key Interfaces
```go
// internal/models/models.go
func GetModelInfo(name string) (ModelInfo, error)
func GetProviderForModel(name string) (string, error)
func ListAllModels() []string
func ListModelsForProvider(provider string) []string
func GetAPIKeyEnvVar(provider string) string
```

### Data Flow
1. CLI/API receives model name
2. Call `models.GetModelInfo(name)` for metadata
3. Use provider info to resolve API key
4. Create appropriate client directly

## Implementation Steps

### Phase 1: Create Models Package (2 hours)
1. Create `internal/models/models.go` with all model definitions
2. Implement lookup functions matching registry interface
3. Add comprehensive tests for all functions
4. Verify all 7 models are correctly defined

### Phase 2: Update Registry API Service (3 hours)
1. Modify `internal/thinktank/registry_api.go`:
   - Import models package instead of registry
   - Replace registry manager calls with direct model lookups
   - Simplify InitLLMClient to use models directly
2. Update API key resolution to use simple switch statement
3. Remove provider registration logic

### Phase 3: Update CLI Integration (2 hours)
1. Update `cmd/thinktank/cli.go`:
   - Remove registry initialization
   - Use models package for validation
   - Simplify provider detection logic
2. Update `internal/thinktank/app.go`:
   - Remove registry manager creation
   - Pass models directly where needed

### Phase 4: Remove Registry Package (1 hour)
1. Delete `internal/registry/` directory
2. Delete `config/models.yaml`
3. Remove registry-related configuration from install script
4. Update any remaining imports

### Phase 5: Cleanup and Validation (2 hours)
1. Run all tests to ensure functionality preserved
2. Run E2E tests with all supported models
3. Update documentation
4. Remove any dead code

## Testing Strategy

### Unit Tests
- Test each lookup function with valid/invalid inputs
- Verify all model metadata is correct
- Test provider grouping functions

### Integration Tests
- Verify client creation works for all models
- Test API key resolution for each provider
- Ensure error handling is preserved

### E2E Tests
- Run existing E2E suite to verify no regression
- Test actual API calls with each model type

### Coverage Requirements
- Maintain 90%+ coverage on new models package
- No reduction in overall coverage

## Risk Analysis & Mitigation

### Risk 1: Missing Model Metadata
- **Severity**: Low
- **Mitigation**: Carefully extract all model data from current YAML
- **Validation**: Unit tests verify each model's metadata

### Risk 2: Breaking API Compatibility
- **Severity**: Medium
- **Mitigation**: Keep same public interfaces in registry_api.go
- **Validation**: E2E tests catch any breaking changes

### Risk 3: Configuration Flexibility Loss
- **Severity**: Very Low
- **Mitigation**: None needed - we haven't added models in months
- **Validation**: Document how to add new models in code

### Risk 4: API Key Resolution Changes
- **Severity**: Medium
- **Mitigation**: Carefully preserve existing environment variable mappings
- **Validation**: Test each provider's API key resolution

## Security Considerations

### API Key Handling
- Continue using environment variables only
- No hardcoded secrets in model definitions
- Maintain existing sanitization in logs

### Configuration Security
- No external configuration files to compromise
- All model data is compile-time constant
- No runtime modification possible

## Observability & Logging

### Logging Changes
- Add debug log when model is looked up
- Log error with correlation ID when model not found
- Preserve existing structured logging

### Metrics
- No metrics changes needed
- Model lookup is now O(1) constant time

## Open Questions

1. **Model Aliases**: Current registry supports model aliases. Do we need this?
   - Analysis: No usage found in codebase
   - Decision: Drop this feature

2. **Parameter Validation**: Registry has complex parameter validation. Keep it?
   - Analysis: Only temperature is ever customized
   - Decision: Simplify to basic range validation

3. **Provider Base URLs**: Some providers have custom base URLs. How to handle?
   - Analysis: Only OpenRouter uses custom URL
   - Decision: Hardcode in client creation

## Success Criteria

1. **Code Reduction**: 10,000+ lines removed
2. **Functionality**: All E2E tests pass
3. **Performance**: No degradation (likely improvement)
4. **Simplicity**: New developer can understand in 5 minutes
5. **Maintainability**: Adding a model is a 5-line change

## Timeline

- **Total Estimate**: 10-12 hours
- **Phase 1**: 2 hours - Create models package
- **Phase 2**: 3 hours - Update registry API
- **Phase 3**: 2 hours - Update CLI
- **Phase 4**: 1 hour - Remove registry
- **Phase 5**: 2 hours - Cleanup and validation
- **Buffer**: 2 hours - Unknown unknowns

## Post-Implementation

1. Update documentation to explain model addition
2. Create PR with clear explanation of changes
3. Coordinate with team on any deployment considerations
4. Monitor for any issues in first few days

## Conclusion

This plan eliminates unnecessary complexity while preserving all required functionality. By following leyline principles and embracing radical simplification, we'll have a cleaner, more maintainable codebase that's easier to understand and modify.
