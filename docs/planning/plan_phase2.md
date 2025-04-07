# Thinktank Test Suite Refactoring - Phase 2

## Phase 2: Simplify Gitignore Testing

**Objective:** Test the actual `gitignoreUtils` implementation against virtual `.gitignore` files instead of mocking the utility itself.

### Steps:

1. **Enhance Filesystem Utilities:**
   - Ensure `virtualFsUtils.ts` can create hidden files like `.gitignore`
   - Verify `addVirtualGitignoreFile` or integrate its logic into `virtualFsUtils.ts`

2. **Refactor Gitignore-Related Tests:**
   - Remove mocks (`jest.mock('../gitignoreUtils')`) and imports from `mockGitignoreUtils`
   - Import actual functions from `src/utils/gitignoreUtils`
   - Setup virtual `.gitignore` files:
     ```typescript
     beforeEach(() => {
       resetVirtualFs();
       createVirtualFs({
         '/project/': '', // Create directory
         '/project/.gitignore': '*.log\n/dist/\n', // Add gitignore
         '/project/src/': '',
         '/project/src/app.ts': 'content',
         '/project/app.log': 'log content',
         '/project/dist/bundle.js': 'bundle content'
       });
       // Clear ignore cache if needed
       gitignoreUtils.clearIgnoreCache();
     });
     ```
   - Test the actual implementation:
     ```typescript
     it('should ignore files based on virtual .gitignore', async () => {
       // Get the filter for the virtual directory
       const filter = await gitignoreUtils.createIgnoreFilter('/project');
       expect(filter.ignores('app.log')).toBe(true);
       expect(filter.ignores('dist/bundle.js')).toBe(true);
       expect(filter.ignores('src/app.ts')).toBe(false);

       // Test shouldIgnorePath directly
       expect(await gitignoreUtils.shouldIgnorePath('/project', 'app.log')).toBe(true);
       expect(await gitignoreUtils.shouldIgnorePath('/project', 'src/app.ts')).toBe(false);
     });
     ```

3. **Remove `mockGitignoreUtils.ts`:**
   - Once all dependent tests are refactored, delete this file
