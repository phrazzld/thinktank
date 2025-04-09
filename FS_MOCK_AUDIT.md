# Filesystem Mocking Audit Results

This document catalogs test files using deprecated filesystem mocking patterns.

## Migration Targets

* Replace `jest.mock('fs')` with helpers from `test/setup/fs.ts` like `setupBasicFs`
* Replace imports from `mockFsUtils.ts` with the new virtual filesystem approach

## Files Needing Migration

| File Path | Category | Complexity | Notes | Priority |
|-----------|----------|------------|-------|----------|
| src/cli/__tests__/cli.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/cli/__tests__/run-command-xdg.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/cli/__tests__/run-command.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/core/__tests__/configManager.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/utils/__tests__/binaryFileDetection.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/utils/__tests__/fileReader.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/utils/__tests__/fileSizeLimit.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/utils/__tests__/formatCombinedInput.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/utils/__tests__/gitignoreComplexPatterns.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/utils/__tests__/gitignoreFiltering.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/utils/__tests__/gitignoreUtils.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/utils/__tests__/readContextFile.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/utils/__tests__/readContextPaths.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/workflow/__tests__/inputHandler.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/workflow/__tests__/outputHandler.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/__tests__/utils/__tests__/fsTestSetup.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |
| src/__tests__/utils/__tests__/virtualFsUtils.test.ts | Mixed | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |


## Categories:
* **Direct Mock:** Uses `jest.mock('fs')` or `jest.mock('fs/promises')`
* **Legacy Util:** Imports/uses `mockFsUtils.ts`
* **Mixed:** Uses both deprecated patterns and potentially some new helpers

## Complexity Guidelines (Fill in manually):
* **Low:** Few mock interactions, simple test logic that can be easily migrated
* **Medium:** Moderate number of mock interactions, some complexity in test logic
* **High:** Many mock interactions, complex setup, deep integration with the mocking logic

## Priority Guidelines (Fill in manually):
* 1 (Highest): Start with these files (Low complexity, high impact)
* 2: Second wave of migration
* 3: Lower priority files

