# TODO

## Foundation & Setup
- [x] Implement Structured Logging (F1)
  - Description: Introduce log levels and color support
  - Dependencies: None
  - Priority: High

- [x] Create Gemini Client Wrapper (F3)
  - Description: Create wrapper around genai client to improve organization and testability
  - Dependencies: None
  - Priority: High

- [x] Implement Model Info Fetching (F2)
  - Description: Fetch model details (inputTokenLimit, outputTokenLimit) via API with in-memory caching
  - Dependencies: F3
  - Priority: High

## Core Context & Token Handling
- [x] Implement Accurate Token Counting (C1)
  - Description: Replace estimateTokenCount with genai client's CountTokens method
  - Dependencies: F3
  - Priority: High

- [x] Add Pre-Generation Token Check (C2)
  - Description: Check full prompt against model input limit before API call
  - Dependencies: C1, F2
  - Priority: High

- [x] Implement Context Handling Strategy (Fail Fast) (C3)
  - Description: Fail with clear error message when token limit is exceeded
  - Dependencies: C2
  - Priority: High

- [x] Add User Confirmation Prompt (C4)
  - Description: Add --confirm-tokens flag to prompt user confirmation before proceeding
  - Dependencies: C1
  - Priority: Medium

## Prompt & Task Input
- [x] Externalize Main Prompt (P1)
  - Description: Move prompt template to external file for better maintainability
  - Dependencies: None
  - Priority: Medium

- [x] Add Task File Input Option (P2)
  - Description: Allow task description to be loaded from a file
  - Dependencies: None
  - Priority: Medium

- [ ] Add Task Clarification Option (P3)
  - Description: Implement optional AI-powered task clarification
  - Dependencies: F3, P1
  - Priority: Low

## User Experience (UX)
- [x] Implement Dry Run / List Files (U2)
  - Description: Add flag to show file list and token count without API call
  - Dependencies: C1
  - Priority: High

- [x] Add CLI Spinner (U1)
  - Description: Add visual indicator during API calls
  - Dependencies: None
  - Priority: Medium

- [x] Improve API Error Reporting (U3)
  - Description: Enhance error messages with more user-friendly information
  - Dependencies: F3
  - Priority: Medium

## Testing & Refinement
- [x] Refactor Code for Better Structure (R1)
  - Description: Improve separation of concerns in main.go
  - Dependencies: F3
  - Priority: High

- [x] Add Unit Tests for fileutil (T1)
  - Description: Create comprehensive tests for file filtering logic
  - Dependencies: None
  - Priority: Medium

- [x] Add Unit Tests for Core Logic (T2)
  - Description: Test argument parsing, prompt building, and token counting
  - Dependencies: R1
  - Priority: Medium

- [ ] Create Integration Tests (T3)
  - Description: Add tests simulating CLI runs with mocked Gemini API client
  - Dependencies: F3, T1, T2
  - Priority: Low

- [x] Run Dependency Check (R2)
  - Description: Run go mod tidy and review dependencies
  - Dependencies: None
  - Priority: Low

## Documentation
- [x] Update README.md (D1)
  - Description: Document new flags, features and token handling behavior
  - Dependencies: All implemented features
  - Priority: Medium

- [ ] Update Development Guide (D2)
  - Description: Add testing info, prompt file locations, logging levels
  - Dependencies: All implemented features
  - Priority: Low