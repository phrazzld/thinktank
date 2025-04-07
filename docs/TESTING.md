# Thinktank Testing Guide

## Best Practices

### Error Handling in Tests

When testing error scenarios in filesystem operations, prefer these approaches (in order of preference):

1. **Use virtualFsUtils for in-memory filesystem tests** (Recommended)
   ```typescript
   // Import the virtual filesystem utilities
   import { createVirtualFs, resetVirtualFs, mockFsModules } from '../../__tests__/utils/virtualFsUtils';
   
   // Mock fs modules
   jest.mock('fs', () => mockFsModules().fs);
   jest.mock('fs/promises', () => mockFsModules().fsPromises);
   
   // Import fs after mocking
   import fs from 'fs';
   import fsPromises from 'fs/promises';
   
   describe('My test suite', () => {
     beforeEach(() => {
       resetVirtualFs();
     });
     
     it('should handle file not found errors', async () => {
       // Create empty filesystem without the file we'll try to access
       createVirtualFs({});
       
       // No special mocking needed - file doesn't exist in virtual FS
       await expect(readFileContent('/nonexistent.txt'))
         .rejects.toThrow(/File not found/);
     });
   });
   ```

2. **Using mockFsUtils for legacy tests**
   ```typescript
   import { mockedFs, createFsError } from '../../__tests__/utils/mockFsUtils';
   
   describe('My legacy test suite', () => {
     beforeEach(() => {
       mockedFs.access.mockReset();
       mockedFs.readFile.mockReset();
     });
     
     it('should handle file not found errors', async () => {
       // Create a proper ENOENT error
       const error = createFsError('ENOENT', 'File not found', 'access', '/path/to/file.txt');
       
       // Mock fs.access to throw the error
       mockedFs.access.mockRejectedValue(error);
       
       await expect(readFileContent('/path/to/file.txt'))
         .rejects.toThrow(/File not found/);
     });
   });
   ```

### Production Code Guidelines

- **Keep test code separate from production code.** Production code should never contain test-specific behavior.
- **Never add test-specific flags** to error objects or other data structures in production code.
- **Use standard Node.js error patterns** in both production and test code for consistency.
- **Use dependency injection** where appropriate to make code more testable without requiring test-specific logic.

## File System Testing

The project uses two approaches for filesystem testing:

1. **In-memory filesystem using memfs (recommended)**
   - Uses `virtualFsUtils.ts` to provide a real filesystem-like interface
   - Fully implements the Node.js fs API
   - Avoids worker crashes and type issues

2. **Legacy mocking approach using mockFsUtils**
   - Uses Jest mocks to intercept fs module calls
   - Being phased out due to reliability issues

## Error Creation

When simulating filesystem errors in tests:

1. **For new tests**, use `virtualFsUtils.createFsError()` to create standardized filesystem errors.
2. **For legacy tests**, use `mockFsUtils.createFsError()` or the re-exported version from `test-helpers.ts`.
3. **Avoid custom error flags** that modify error behavior in production code.

# Test Migration Guide

See the test utilities README.md file (`src/__tests__/utils/README.md`) for a comprehensive guide on migrating from the old mockFsUtils approach to the new virtualFsUtils approach. It includes step-by-step instructions with code examples for all common filesystem operations.
