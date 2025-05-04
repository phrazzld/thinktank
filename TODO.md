# Test Coverage Remediation Tasks

## Registry API Test Coverage
- [x] Create registry_api_test.go file with test harness
- [x] Implement test for NewRegistryAPIService constructor
- [x] Implement test for InitLLMClient, verifying API key resolution
- [x] Implement test for GetModelParameters, covering request parameter creation
- [x] Implement test for ValidateModelParameter, including validation edge cases
  - ✓ Refactored test files to address file size limit
- [x] Implement test for GetModelDefinition, verifying definition retrieval logic
- [x] Implement test for GetModelTokenLimits, testing standard and custom token limits
- [x] Implement test for ProcessLLMResponse, testing response transformation
- [x] Implement test for IsEmptyResponseError, covering error classification
- [x] Implement test for IsSafetyBlockedError, testing safety error detection
- [x] Implement test for getEnvVarNameForProvider, covering provider-specific envvar patterns
- [ ] Implement test for GetErrorDetails, verifying error detail extraction

## Adapter Test Coverage
- [ ] Create adapters_test.go file with test harness
- [ ] Implement test for InitLLMClient adapter method
- [ ] Implement test for ProcessLLMResponse adapter method
- [ ] Implement test for GetErrorDetails adapter method
- [ ] Implement test for IsEmptyResponseError adapter method
- [ ] Implement test for IsSafetyBlockedError adapter method
- [ ] Implement test for GetModelParameters adapter method
- [ ] Implement test for ValidateModelParameter adapter method
- [ ] Implement test for GetModelDefinition adapter method
- [ ] Implement test for GetModelTokenLimits adapter method
- [ ] Implement test for interfacesToInternalContextStats data conversion
- [ ] Implement test for internalToInterfacesContextStats data conversion
- [ ] Implement test for internalToInterfacesGatherConfig data conversion
- [ ] Implement test for DisplayDryRunInfo adapter method
- [ ] Implement test for SaveToFile adapter method
- [ ] Implement test for GatherContext adapter method

## Integration Tests
- [ ] Create registry_api_boundary_test.go for integration testing
- [ ] Implement test for adapter-registry full interactions
- [ ] Implement test for error propagation across boundaries
- [ ] Implement test for real-world workflows with mock providers

## Test Infrastructure
- [ ] Create mock implementation of Registry interface for testing
- [ ] Create test fixtures for model definitions and responses
- [ ] Update test helpers to support registry API testing
- [ ] Create shared test scenarios that can be used across test files

## Coverage Validation
- [ ] Update coverage scripts to properly report registry API coverage
- [ ] Create pre-submission coverage check script
- [ ] Document coverage expectations in README.md

## Optional Temporary Measures
- [ ] Temporarily lower coverage threshold to 64% in CI config
- [ ] Add .nocover comments to adapter methods that are pure wrappers
- [ ] Configure coverage exclusion patterns for test helpers
