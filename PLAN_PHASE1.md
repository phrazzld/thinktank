# Thinktank Test Suite Refactoring - Phase 1

## Phase 1: Complete `memfs` Migration for Filesystem Tests

**Objective:** Eliminate the legacy `mockFsUtils.ts` and ensure all filesystem-related unit/integration tests use the `virtualFsUtils.ts` (`memfs`) approach for speed, isolation, and consistency.

### Steps:

1. **Identify Target Test Files:**
   - Review all files under `src/**/__tests__/*.test.ts`
   - Identify tests currently importing or relying on `mockFsUtils.ts`
   - Cross-reference with skipped tests in `jest.config.js` (`testPathIgnorePatterns`)
   
2. **Refactor Each Target Test File:**
   - Remove legacy imports from `../../__tests__/utils/mockFsUtils`
   - Add `memfs` setup:
     ```typescript
     import { mockFsModules } from '../../__tests__/utils/virtualFsUtils';
     jest.mock('fs', () => mockFsModules().fs);
     jest.mock('fs/promises', () => mockFsModules().fsPromises);
     // Now import fs, fs/promises, and the module under test
     import fsPromises from 'fs/promises';
     import { yourFunctionUnderTest } from '../yourModule';
     ```
   - Replace mock setups with `createVirtualFs`:
     ```typescript
     // Before (Legacy)
     // beforeEach(() => {
     //   resetMockFs();
     //   setupMockFs();
     //   mockReadFile('/path/to/file.txt', 'content');
     //   mockStat('/path/to/dir', { isDirectory: () => true });
     //   mockMkdir('/output/dir', true);
     // });

     // After (memfs)
     beforeEach(() => {
       resetVirtualFs();
       createVirtualFs({
         '/path/to/file.txt': 'content',
         '/path/to/dir/': '', // Creates a directory
         // '/output/dir/' will be created by the function under test
       });
     });
     ```
   - Update assertions to verify filesystem state:
     ```typescript
     // Before (Legacy)
     // await createDirectory('/output/dir');
     // expect(mockedFs.mkdir).toHaveBeenCalledWith('/output/dir', { recursive: true });

     // After (memfs)
     await createDirectory('/output/dir');
     const virtualFs = getVirtualFs();
     expect(virtualFs.existsSync('/output/dir')).toBe(true);
     expect(virtualFs.statSync('/output/dir').isDirectory()).toBe(true);
     ```
   - Refactor error testing:
     ```typescript
     it('should handle write permission errors', async () => {
       createVirtualFs({ '/path/to/': '' }); // Ensure parent dir exists
       const writeFileSpy = jest.spyOn(fsPromises, 'writeFile');
       writeFileSpy.mockRejectedValueOnce(
         createFsError('EACCES', 'Permission denied', 'writeFile', '/path/to/file.txt')
       );

       await expect(writeFileFunction('/path/to/file.txt', 'content'))
         .rejects.toThrow(/Permission denied/);

       writeFileSpy.mockRestore(); // Clean up the spy
     });
     ```
   - Update `jest.config.js` by removing its path from `testPathIgnorePatterns`

3. **Remove Legacy Utilities:**
   - Once all test files are migrated, delete `src/__tests__/utils/mockFsUtils.ts`
   - Remove any helper functions in `test-helpers.ts` related to the old mocking strategy
