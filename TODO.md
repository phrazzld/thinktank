# TODO: Registry Elimination - Merge Blockers and Follow-up Work

## 🚨 MERGE BLOCKERS (Required before merge)

### [x] Restore Comprehensive Parameter Validation in ValidateModelParameter
**Context**: During registry elimination, the `ValidateModelParameter` function in `internal/thinktank/registry_api.go` was simplified to only validate the `temperature` parameter. The previous comprehensive validation for `top_p`, `max_tokens`, `top_k`, `frequency_penalty`, `presence_penalty`, and enum string values was completely removed.

**Problem**: This creates a functional regression where invalid parameters that were previously caught by application validation will now be sent to LLM provider APIs, causing runtime failures and poor user experience.

**Requirements**:
1. **Extend ModelInfo struct** in `internal/models/models.go` to include parameter constraints:
   ```go
   type ParameterConstraint struct {
       Type      string      // "int", "float", "string"
       MinValue  *float64    // for numeric types
       MaxValue  *float64    // for numeric types
       EnumValues []string   // for string enums
   }

   type ModelInfo struct {
       // ... existing fields
       ParameterConstraints map[string]ParameterConstraint
   }
   ```

2. **Update model definitions** to include parameter constraints for each model:
   - `temperature`: float, 0.0-2.0
   - `top_p`: float, 0.0-1.0
   - `max_tokens`: int, 1-context_window
   - `top_k`: int, 1-100 (for applicable models)
   - `frequency_penalty`: float, -2.0-2.0 (OpenAI models)
   - `presence_penalty`: float, -2.0-2.0 (OpenAI models)

3. **Implement validation logic** in `ValidateModelParameter` function:
   - Look up parameter constraints from `ModelInfo.ParameterConstraints`
   - Validate parameter type matches expected type
   - For numeric types: check min/max bounds
   - For string types: validate against enum values if specified
   - Return descriptive error messages for validation failures

4. **Add comprehensive tests** in `internal/models/models_test.go`:
   - Test valid parameter values pass validation
   - Test invalid parameter values are rejected with appropriate errors
   - Test edge cases (boundary values, type mismatches)
   - Test all supported parameters for each model type

**Success Criteria**:
- All parameters that were validated in the old registry system are validated in the new system
- Invalid parameter values are caught before being sent to LLM provider APIs
- Clear, actionable error messages are returned for validation failures
- No functional regression in parameter validation behavior

**Estimated Effort**: 4-6 hours
**Priority**: CRITICAL - Must be completed before merge

---

## ✅ POST-MERGE IMPROVEMENTS (Not blocking, can be addressed incrementally)

### [ ] Enhance API Key Format Validation
**Context**: Add provider-specific API key format validation to catch malformed keys early.
**Why not blocking**: The system worked without format validation before. This is a quality improvement, not fixing a regression.

### [ ] Audit Error Handling Consistency
**Context**: Review `handleProcessingOutcome` and `handleError` functions for consistent error propagation patterns.
**Why not blocking**: Error handling still works. This is about improving consistency, not fixing broken functionality.

### [ ] Verify Context Propagation in New Code Paths
**Context**: Ensure all new goroutines and external calls properly receive and handle context.
**Why not blocking**: No evidence that context propagation is broken. This is defensive verification, not fixing a confirmed problem.

### [ ] Standardize Log Message Formats
**Context**: Ensure consistent log message formatting and adequate context across the refactored codebase.
**Why not blocking**: Logging works. This is about improving consistency and debuggability.

---

## ❌ NON-ISSUES (Already resolved or not applicable)

### ~~Nil Pointer Dereference in ModelProcessor~~
**Status**: Already fixed with nil check according to code review
**Action**: No further action needed

### ~~Security Issue in Error Handling~~
**Status**: Speculative concern with no concrete vulnerability identified
**Action**: No action needed unless specific security issue is demonstrated

---

## Notes

**Merge Philosophy**: Only functional regressions that break existing behavior should block merges. Quality improvements, defensive programming, and speculative concerns can be addressed incrementally to maintain development velocity while ensuring system reliability.

**Validation Regression Impact**: This is the only identified functional regression where behavior that previously worked (comprehensive parameter validation) no longer works as expected. All other identified issues are either already fixed, speculative, or quality improvements rather than regressions.
