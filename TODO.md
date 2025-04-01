# TODO

## Setup & Core Primitives

- [ ] Install Dependencies
  - Description: Verify cli-table3 and figures are properly installed (already in package.json)
  - Dependencies: None
  - Priority: High

- [x] Create Console Utility Module
  - Description: Create `src/atoms/consoleUtils.ts` to centralize chalk, figures, and common formatting helpers
  - Dependencies: Installed dependencies
  - Priority: High

## Enhanced Progress Indicators

- [ ] Implement Stage Timing
  - Description: Add start/end timestamps for major stages (config load, model processing, file writing) in runThinktank
  - Dependencies: None
  - Priority: Medium

- [ ] Refine Spinner Messages
  - Description: Update ora spinner text in runThinktank to show counts, percentages, and stage timing
  - Dependencies: Stage Timing
  - Priority: High

## Group-Based Output Organization

- [ ] Add In-Progress Group Headers
  - Description: Add formatted headers (name, count, description) for groups in runThinktank
  - Dependencies: Console Utils
  - Priority: Medium

## Structured Results Display

- [ ] Implement Tabular Results Formatter
  - Description: Create function in outputFormatter using cli-table3 for the final summary table
  - Dependencies: Console Utils
  - Priority: High

- [ ] Integrate Performance Metrics
  - Description: Include response time and token counts in the results table
  - Dependencies: Tabular Results Formatter
  - Priority: Medium

- [ ] Calculate Group Summary Stats
  - Description: Compute success/error counts and avg. response time per group for display
  - Dependencies: Tabular Results Formatter, Group Headers
  - Priority: Medium

## Error Handling Display

- [ ] Create Error Formatting Helper
  - Description: Develop a function in consoleUtils to format errors consistently with categories and tips
  - Dependencies: Console Utils
  - Priority: High

- [ ] Categorize & Color-Code Errors
  - Description: Enhance error messages in runThinktank using the error formatter
  - Dependencies: Error Formatter
  - Priority: Medium

- [ ] Add Troubleshooting Tips
  - Description: Implement logic to detect common error patterns and suggest fixes
  - Dependencies: Error Formatter
  - Priority: Medium

## Integration & Refinement

- [ ] Integrate Formatting into runThinktank
  - Description: Replace existing console.log calls with new utilities and formatters
  - Dependencies: All previous tasks
  - Priority: High

- [ ] Refactor outputFormatter.ts
  - Description: Adapt existing functions to use the new table formatter and console utils
  - Dependencies: Tabular Results Formatter
  - Priority: Medium

- [ ] Implement --verbose Flag
  - Description: Add flag handling in cli.ts and conditional detailed logging in runThinktank
  - Dependencies: None
  - Priority: Low

## Documentation & Polish

- [ ] Add First-Run Usage Hints
  - Description: Implement mechanism to display tips on first execution
  - Dependencies: None
  - Priority: Low

- [ ] Add Documentation Links
  - Description: Include relevant documentation links in error messages or help output
  - Dependencies: Error Formatter
  - Priority: Low

- [ ] Implement Terminal Compatibility (Optional)
  - Description: Add detection for limited terminals with non-color/non-Unicode fallback
  - Dependencies: Console Utils
  - Priority: Low

## Testing

- [ ] Unit Test Console Utils
  - Description: Write tests for all helper functions in consoleUtils
  - Dependencies: Console Utils implementation
  - Priority: Medium

- [ ] Unit Test Table Formatter
  - Description: Test the tabular results formatter with various input scenarios
  - Dependencies: Table Formatter implementation
  - Priority: Medium

- [ ] Integration Test runThinktank
  - Description: Test the enhanced console output in the main workflow
  - Dependencies: All core implementations
  - Priority: Medium

- [ ] Manual Terminal Testing
  - Description: Test in various terminal emulators to verify formatting
  - Dependencies: All implementations complete
  - Priority: Medium