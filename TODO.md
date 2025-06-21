# TODO: Development Tasks

## Multi-Model Reliability (Critical)

### Phase 1: Fix Critical Failures
- [x] Fix `openrouter/deepseek/deepseek-r1-0528:free` batch failure
  - **Issue**: 100% failure in batch runs, 100% success individually
  - **Impact**: Prevents reliable multi-model processing
  - **Effort**: Low (add concurrency limit field)
  - **Priority**: Critical

- [x] Make ModelDefinitions private
  - **Issue**: Currently exported, should be encapsulated
  - **Impact**: Better API design, prevents misuse
  - **Effort**: Low (add accessor functions)
  - **Priority**: Medium
  - **depends-on**: Fix batch failure

### Phase 2: Enhanced Error Reporting
- [x] Add detailed error structure for model failures
  - **Issue**: Generic "model processing failed" errors
  - **Impact**: Better debugging and user experience
  - **Effort**: Medium (new error types, categorization)
  - **Priority**: High

- [x] Implement provider-specific error handling
  - **Issue**: All errors treated identically regardless of provider
  - **Impact**: More actionable error messages
  - **Effort**: Medium (provider-specific error mapping)
  - **Priority**: Medium
  - **depends-on**: Add detailed error structure

### Phase 3: Rate Limiting Optimization
- [x] Implement provider-aware rate limiting
  - **Issue**: One-size-fits-all 60 RPM regardless of provider capabilities
  - **Impact**: Better performance, reduced unnecessary delays
  - **Effort**: Medium (per-provider configuration)
  - **Priority**: Medium

- [x] Add CLI flags for rate limit customization
  - **Issue**: No user control over rate limiting
  - **Impact**: Power users can optimize for their API tiers
  - **Effort**: Low (CLI flag parsing)
  - **Priority**: Low
  - **depends-on**: Implement provider-aware rate limiting

## Code Quality & Architecture

### Immediate Improvements
- [ ] Mock HTTP client in parameter boundary tests
  - **Issue**: TODO comment in openrouter parameter_boundary_test.go:398
  - **Impact**: More reliable, faster tests
  - **Effort**: Low (add mock HTTP transport)
  - **Priority**: Low

- [ ] Resolve OpenAI client configuration TODO
  - **Issue**: TODO comment in openai_client.go:330 about setting configuration
  - **Impact**: Cleaner configuration handling
  - **Effort**: Low (investigate and implement proper setting)
  - **Priority**: Low

### Testing Infrastructure
- [x] Add comprehensive multi-model integration tests
  - **Issue**: No systematic testing of all models together
  - **Impact**: Catch regressions, ensure reliability
  - **Effort**: Medium (test infrastructure, CI integration)
  - **Priority**: Medium

- [ ] Create load testing scripts for model reliability
  - **Issue**: Need to validate performance under various loads
  - **Impact**: Production confidence, performance insights
  - **Effort**: Low (bash scripts using existing CLI)
  - **Priority**: Low
  - **depends-on**: Add comprehensive multi-model integration tests

## Documentation & Usability

### User Experience
- [x] Document rate limiting best practices per provider
  - **Issue**: Users don't know optimal settings for their API tiers
  - **Impact**: Better user experience, fewer rate limit issues
  - **Effort**: Low (documentation update)
  - **Priority**: Low

- [ ] Add troubleshooting guide for model failures
  - **Issue**: Users can't easily diagnose model-specific issues
  - **Impact**: Reduced support burden, better self-service
  - **Effort**: Low (documentation)
  - **Priority**: Low
  - **depends-on**: Add detailed error structure

## Selection Criteria

**For [ ] (unstarted)**: Find first task where all `depends-on:` are `[x]`
**For [~] (in-progress)**: Complete the task currently being worked on
**For [x] (completed)**: Task is done and validated

**Priority Levels**:
- **Critical**: Blocks core functionality
- **High**: Significantly improves user experience
- **Medium**: Good improvements, not blocking
- **Low**: Nice to have, can be deferred

**Effort Levels**:
- **Low**: < 4 hours, single file changes
- **Medium**: 4-16 hours, multi-file changes
- **High**: > 16 hours, architectural changes
