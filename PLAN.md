# Thinktank Test Suite Refactoring - Phase 4

## Phase 4: Isolate Side Effects

**Objective:** Separate I/O operations from core logic to improve testability.

### Steps:

1. **Identify Functions with Side Effects:**
   - Analyze functions performing I/O (file writing, console logging)
   - Focus on `_processOutput`, `_logCompletionSummary`, etc.

2. **Refactor to Return Data:**
   - Modify functions to return data rather than perform I/O:
     ```typescript
     export async function _processOutput({
       queryResults,
       options,
       friendlyRunName,
     }: ProcessOutputParams): Promise<{ files: FileData[]; consoleOutput: string }> {
       // Return file content, don't write it
       const files = queryResults.responses.map(r => ({
         filename: `${r.provider}-${r.modelId}.md`,
         content: formatResponse(r),
       }));
       return { files, consoleOutput: formatConsoleOutput(files) };
     }
     ```

3. **Move I/O to Higher-Level Functions:**
   - Push actual I/O operations to orchestration functions:
     ```typescript
     // In runThinktank or dedicated I/O module:
     const { files, consoleOutput } = await _processOutput({ ... });
     await writeFiles(outputDirectoryPath, files);
     console.log(consoleOutput);
     ```
