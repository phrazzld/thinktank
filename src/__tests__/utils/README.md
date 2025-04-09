# [DEPRECATED] Legacy Test Utilities

⚠️ **WARNING: DEPRECATED PATTERNS** ⚠️

The utilities and patterns in this directory are **deprecated** and should **not** be used for new tests. They do not align with the project's current testing philosophy and infrastructure.

## Preferred Approach

Please refer to the following resources for the recommended testing approach:

- **[Jest Testing Documentation](../../jest/README.md)**: For an overview of the testing strategy and Jest configuration
- **[Test Setup Helpers](../../test/setup/README.md)**: For detailed documentation and examples of the standard test setup helpers
- **[Test Data Factories](../../test/factories/README.md)**: For utilities creating standard test data objects

## Migration Guide

If you're using utilities from this directory, please migrate to the new approach:

### From `mockFsUtils.ts` to `virtualFsUtils.ts` and `test/setup/fs.ts`

```typescript
// OLD (DEPRECATED)
import { setupMockFs, mockReadFile } from '../../../__tests__/utils/mockFsUtils';
setupMockFs();
mockReadFile('/path/to/file.txt', 'content');

// NEW (RECOMMENDED)
import { setupTestHooks, setupBasicFs } from '../../test/setup';
setupTestHooks();
setupBasicFs({
  '/path/to/file.txt': 'content'
});
```

### From `mockGitignoreUtils.ts` to `test/setup/gitignore.ts`

```typescript
// OLD (DEPRECATED)
import * as mockGitignoreUtils from '../../../__tests__/utils/mockGitignoreUtils';
mockGitignoreUtils.setupMockGitignore();
mockGitignoreUtils.mockShouldIgnorePath('/dir', 'file.log', true);

// NEW (RECOMMENDED)
import { setupWithGitignore } from '../../test/setup/gitignore';
await setupWithGitignore('/dir', '*.log', {
  'file.txt': 'content',
  'file.log': 'should be ignored'
});
```

### From `mockFactories.ts` to `test/factories/*` and `test/setup/*`

```typescript
// OLD (DEPRECATED)
import { mockConsoleLogger, mockFileSystem } from '../../../__tests__/utils/mockFactories';
const mockLogger = mockConsoleLogger();
const mockFs = mockFileSystem();

// NEW (RECOMMENDED)
import { createMockConsoleLogger, createMockFileSystem } from '../../test/setup';
const mockLogger = createMockConsoleLogger();
const mockFs = createMockFileSystem();
```

## Remaining Applicable Utilities

### Virtual File System (`virtualFsUtils.ts`)

The `virtualFsUtils.ts` module remains relevant but should be used through the higher-level helpers in `test/setup/fs.ts` when possible:

```typescript
// Prefer this approach when possible
import { setupBasicFs } from '../../test/setup';
setupBasicFs({
  '/path/to/file.txt': 'File content'
});

// Only use virtualFsUtils directly for low-level virtual FS operations
import { resetVirtualFs, getVirtualFs } from '../../../__tests__/utils/virtualFsUtils';
resetVirtualFs();
const virtualFs = getVirtualFs();
virtualFs.writeFileSync('/manual/path.txt', 'content');
```
