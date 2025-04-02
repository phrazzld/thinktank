# TODO: Implement Fun Output Directory Naming with Gemini

## Create Name Generator

- [x] Create nameGenerator.ts file
  - Description: Create new file with utility functions for generating fun run names
  - Dependencies: @google/generative-ai package
  - Priority: High

- [x] Implement Gemini structured output schema
  - Description: Define the JSON schema for structured output responses
  - Dependencies: None
  - Priority: High

- [x] Implement generateFunName() function
  - Description: Create function that prompts Gemini for adjective-noun names
  - Dependencies: Gemini JSON schema
  - Priority: High

- [x] Add API error handling to generateFunName()
  - Description: Handle potential API errors, authentication issues, rate limiting
  - Dependencies: generateFunName() implementation
  - Priority: High

- [x] Add response validation to generateFunName()
  - Description: Validate response format with regex (/^[a-z]+-[a-z]+$/)
  - Dependencies: generateFunName() implementation
  - Priority: High

- [x] Add fallback mechanism to handle invalid responses
  - Description: Add code to extract valid names from otherwise invalid responses
  - Dependencies: generateFunName() validation
  - Priority: Medium

- [x] Implement generateFallbackName() function
  - Description: Create timestamp-based fallback name function (run-YYYYMMDD-HHmmss)
  - Dependencies: None
  - Priority: High

## Workflow Integration

- [x] Import name generator in runThinktank.ts
  - Description: Add imports for nameGenerator functions
  - Dependencies: nameGenerator.ts implementation
  - Priority: High

- [x] Add run name generation to runThinktank workflow
  - Description: Call generateFunName() after config loading but before directory creation
  - Dependencies: nameGenerator.ts implementation
  - Priority: High

- [x] Implement fallback to timestamp if name generation fails
  - Description: Call generateFallbackName() when generateFunName() returns null
  - Dependencies: nameGenerator.ts implementation
  - Priority: High

- [x] Store friendly name for later use in console output
  - Description: Store generated name in variable accessible throughout function scope
  - Dependencies: Name generation implementation
  - Priority: High

## Console Output Updates

- [x] Update output directory creation message
  - Description: Include run name in output directory creation message
  - Dependencies: Run name integration in workflow
  - Priority: Medium

- [x] Update model completion success message
  - Description: Include run name in completion message
  - Dependencies: Run name integration in workflow
  - Priority: Medium

- [x] Update final output directory display message
  - Description: Add separate line for run name in final output 
  - Dependencies: Run name integration in workflow
  - Priority: Medium

## Testing

- [ ] Create nameGenerator.test.ts
  - Description: Create test file for name generator functions
  - Dependencies: nameGenerator.ts implementation
  - Priority: Medium

- [ ] Add successful name generation test
  - Description: Test generateFunName() with mocked successful API response
  - Dependencies: nameGenerator.test.ts setup
  - Priority: Medium

- [ ] Add API error tests
  - Description: Test handling of API errors, network issues, auth failures
  - Dependencies: nameGenerator.test.ts setup
  - Priority: Medium

- [ ] Add invalid response format tests
  - Description: Test handling of responses that don't match expected format
  - Dependencies: nameGenerator.test.ts setup
  - Priority: Medium

- [ ] Add fallback name generation test
  - Description: Test generateFallbackName() timestamp format
  - Dependencies: nameGenerator.test.ts setup
  - Priority: Medium

- [ ] Update runThinktank tests
  - Description: Update existing tests to handle new run name functionality
  - Dependencies: Run name integration in workflow
  - Priority: Medium

## Documentation (Optional)

- [ ] Update README.md
  - Description: Document the new run naming feature and GEMINI_API_KEY requirement
  - Dependencies: Complete implementation
  - Priority: Low

## Assumptions and Questions

1. We're assuming the Gemini-2.0-flash model reliably supports structured output/JSON schema mode. If not, we'll need to handle raw text responses with more robust validation.

2. We're assuming the fallback name format (run-YYYYMMDD-HHmmss) is acceptable. This should be confirmed.

3. The current implementation plan only displays the friendly name in console output and doesn't store it elsewhere (like in a metadata file). This should be confirmed as sufficient.

4. We're assuming the small latency added by the API call (~1-3 seconds) is acceptable for the improved user experience.