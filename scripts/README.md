# Thinktank Scripts

This directory contains utility scripts for the thinktank project.

## Scripts

### audit-fs-mocks.js

A utility script that scans the codebase for test files using deprecated filesystem mocking approaches. It helps identify test files that need to be migrated to the new virtual filesystem approach using `memfs`.

**Usage:**

```bash
node scripts/audit-fs-mocks.js
```

**Features:**

- Scans test files for different fs mocking patterns:
  - Direct mocking (`jest.mock('fs')`, `jest.mock('fs/promises')`)
  - Legacy utilities (imports/usage from `mockFsUtils.ts`)
  - Virtual filesystem approach (imports/usage from `test/setup/fs.ts` or `virtualFsUtils.ts`)
- Categorizes files based on the patterns found
- Generates a markdown report (`FS_MOCK_AUDIT.md`) with:
  - A table of files needing migration
  - Categories and complexity guidelines
  - Migration targets and priorities

**Implementation:**

- Uses regular expressions to identify different mocking patterns
- Implements dependency injection for better testability
- Includes comprehensive tests for all functionality
