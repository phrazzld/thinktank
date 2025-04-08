# Manual Mocks for Utils Module

This directory contains manual mocks for Jest testing. These mocks replace the actual implementations
when you use `jest.mock('../moduleName')` in your test files.

## Available Mocks

### fileReader.ts

Provides mocks for file reading utilities:

- `fileExists`: Mock for checking if a file exists
- `readContextFile`: Mock for reading individual files
- `readDirectoryContents`: Mock for reading directory contents

### Important Note

**If you're using a different mocking approach in your tests, such as:**

```typescript
// Don't use this approach unless you've also added jest.mock('../fileReader')
const readContextFile = jest.fn().mockResolvedValue({ ... });
```

You might experience conflicts with these manual mocks. In those cases, you should:

1. Either switch to using the manual mock with proper spy:

```typescript
// Import the mock implementation 
import * as fileReader from '../fileReader';

// Use a spy to change implementation for a specific test
jest.spyOn(fileReader, 'readContextFile').mockImplementation(...);
```

2. Or disable the manual mock for your test file:

```typescript
// At the top of your test file
jest.unmock('../fileReader');

// Then use your own mocks
const readContextFile = jest.fn().mockResolvedValue({ ... });
```

## How the Mocks Work

These mocks create Jest mock functions (`jest.fn()`) with default implementations
that can be customized in tests as needed.