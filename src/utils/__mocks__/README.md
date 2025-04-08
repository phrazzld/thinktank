# Manual Mocks for Utils Module (DEPRECATED)

> ⚠️ **DEPRECATED**: This manual mocks approach is deprecated. Please use the centralized mock setup from `jest/setupFiles/` for all new tests and when refactoring existing tests. See `jest/README.md` for details on the preferred approach.

This directory contains manual mocks for Jest testing. These mocks replace the actual implementations
when you use `jest.mock('../moduleName')` in your test files.

## Migration Guide

Convert from manual mocks to the centralized approach:

**Before** (using manual mocks - deprecated):
```typescript
// Using manual mocks
jest.mock('../../utils/fileReader');
import { readContextFile } from '../../utils/fileReader';

// Later in test
readContextFile.mockResolvedValueOnce({ 
  path: '/test.txt', 
  content: 'mocked content',
  error: null 
});
```

**After** (using centralized approach - preferred):
```typescript
// Using centralized setup
import { setupBasicFs } from '../../../jest/setupFiles/fs';

beforeEach(() => {
  setupBasicFs({
    '/test.txt': 'mocked content'
  });
});

// Later in test
const { readContextFile } = await import('../../utils/fileReader');
const result = await readContextFile('/test.txt');
```

## Why Migrate?

The centralized approach:
- Aligns with our testing philosophy of minimizing mocking
- Tests real behavior through the virtual filesystem
- Reduces brittle implementation-specific mocks
- Simplifies test setup and maintenance
- Provides more realistic test behavior

## Available Mocks (For Legacy Tests Only)

### fileReader.ts

Provides mocks for file reading utilities:

- `fileExists`: Mock for checking if a file exists
- `readContextFile`: Mock for reading individual files
- `readDirectoryContents`: Mock for reading directory contents

## Conflicts Between Approaches

If you encounter conflicts between manual mocks and other approaches:

1. Preferably, migrate to the centralized approach in `jest/setupFiles/`

2. Or disable the manual mock for your test file as a temporary solution:

```typescript
// At the top of your test file
jest.unmock('../fileReader');

// Then use your own mocks
const readContextFile = jest.fn().mockResolvedValue({ ... });
```
