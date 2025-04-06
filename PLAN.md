```markdown
# PLAN.md

## 1. Overview

This plan outlines the technical steps required to add a feature allowing users to provide an arbitrary number of files and directories as context when running a prompt with `thinktank`. The main prompt will still be provided via a single file, but users will be able to specify additional paths (files or directories) whose contents will be combined with the prompt and sent to the LLMs.

## 2. Task Breakdown

| Task                                                                 | Effort | Affected Files/Modules                                                                                                |
| :------------------------------------------------------------------- | :----- | :-------------------------------------------------------------------------------------------------------------------- |
| **1. Modify CLI `run` Command**                                      | S      | `src/cli/commands/run.ts`, `src/cli/index.ts`                                                                         |
|    - Update `commander` definition to accept variadic context paths. |        |                                                                                                                       |
|    - Pass context paths to the `runThinktank` workflow function.     |        |                                                                                                                       |
| **2. Enhance Input Handling**                                        | M      | `src/workflow/inputHandler.ts`, `src/utils/fileReader.ts` (potentially new file for context reading)                |
|    - Create function(s) to read content from files and directories recursively. |        |                                                                                                                       |
|    - Implement logic to ignore specific files/directories (e.g., `.git`, `node_modules`). |        |                                                                                                                       |
|    - Define and implement a formatting strategy for combining prompt and context content. |        |                                                                                                                       |
|    - Modify `_processInput` helper or create a new one to handle prompt + context. |        | `src/workflow/runThinktankHelpers.ts`                                                                                 |
| **3. Update Workflow Orchestration**                                 | S      | `src/workflow/runThinktank.ts`, `src/workflow/runThinktankHelpers.ts`, `src/workflow/queryExecutor.ts`                |
|    - Adapt `runThinktank` to receive context paths from CLI options. |        |                                                                                                                       |
|    - Modify `_processInput` call/logic to handle context paths.      |        |                                                                                                                       |
|    - Pass the combined prompt+context string to `_executeQueries`.   |        |                                                                                                                       |
|    - Update `executeQueries` options and implementation to accept combined content. |        |                                                                                                                       |
| **4. Adapt LLM Providers**                                           | S      | `src/providers/*.ts`, `src/core/types.ts`                                                                             |
|    - Ensure `generate` methods in providers can handle larger input strings (prompt + context). |        |                                                                                                                       |
|    - (Optional) Check and potentially log warnings for provider-specific context limits. |        |                                                                                                                       |
| **5. Implement Unit Tests**                                          | M      | `src/workflow/__tests__/inputHandler.test.ts`, `src/utils/__tests__/fileReader.test.ts` (or new context reader test) |
|    - Test recursive directory reading logic.                         |        |                                                                                                                       |
|    - Test file/directory filtering logic.                            |        |                                                                                                                       |
|    - Test context formatting logic.                                  |        |                                                                                                                       |
|    - Test modifications in `_processInput` helper.                   |        | `src/workflow/__tests__/processInputHelper.test.ts`                                                                   |
| **6. Implement Integration Tests**                                   | M      | `src/cli/__tests__/run-command.test.ts`, `src/workflow/__tests__/runThinktank.test.ts`                                |
|    - Test `run` command with various context path arguments (files, dirs, mixed). |        |                                                                                                                       |
|    - Test `runThinktank` workflow correctly passes combined context. |        |                                                                                                                       |
| **7. Implement E2E Tests**                                           | M      | `src/cli/__tests__/cli.e2e.test.ts`                                                                                   |
|    - Add E2E tests simulating CLI usage with context files/dirs.     |        |                                                                                                                       |
| **8. Update Documentation**                                          | S      | `README.md`, CLI help text (`src/cli/commands/run.ts`)                                                                |
|    - Document the new `[contextPaths...]` argument for the `run` command. |        |                                                                                                                       |
|    - Provide usage examples.                                         |        |                                                                                                                       |
|    - Explain the context formatting strategy.                        |        |                                                                                                                       |

## 3. Implementation Details

### 3.1. CLI Modifications (`src/cli/commands/run.ts`)

-   Modify the `commander` definition for the `run` command:
    ```typescript
    // Before:
    // .argument('<promptFile>', 'Path to the file containing the prompt')

    // After:
    .argument('<promptFile>', 'Path to the file containing the main prompt')
    .argument('[contextPaths...]', 'Optional paths to files or directories to include as context') // Variadic argument
    ```
-   Update the action handler to receive `contextPaths`:
    ```typescript
    .action(async (promptFile: string, contextPaths: string[], options: { /* existing options */ }) => {
        // ... existing validation ...

        // Pass contextPaths to runThinktank
        await runThinktank({
            input: promptFile,
            contextPaths, // Pass the new argument
            // ... other options ...
        });
    });
    ```
-   Update `RunOptions` interface in `src/workflow/runThinktank.ts` to include `contextPaths?: string[]`.

### 3.2. Input Handling (`src/workflow/inputHandler.ts`, new context reader utility)

-   **Create `readContextPaths` function:**
    -   Input: `paths: string[]` (list of file/directory paths)
    -   Output: `Promise<Array<{ path: string; content: string }>>`
    -   Logic:
        -   Iterate through each path.
        -   Use `fs.stat` to determine if it's a file or directory.
        -   If file: Read content using `fs.readFile`.
        -   If directory:
            -   Recursively read directory contents using `fs.readdir(path, { withFileTypes: true })`.
            -   Ignore specified directories/files (always ignore whatever is enumerated in .gitignore).
            -   For each file within the directory, read its content.
            -   Handle potential errors (permissions, not found).
        -   Consider adding limits (e.g., max depth, max files, max size) - *Defer strict limits for now, maybe log warnings.*
        -   Handle binary files: Skip or log a warning. *Start by skipping.*
    -   Example recursive read function (simplified):
        ```typescript
        import fs from 'fs/promises';
        import path from 'path';

        const IGNORED_DIRS = new Set(['.git', 'node_modules', 'dist', 'coverage']);
        const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB limit per file

        async function readDirectoryRecursive(dirPath: string, basePath: string = dirPath): Promise<Array<{ path: string; content: string }>> {
            let results: Array<{ path: string; content: string }> = [];
            try {
                const entries = await fs.readdir(dirPath, { withFileTypes: true });
                for (const entry of entries) {
                    const fullPath = path.join(dirPath, entry.name);
                    const relativePath = path.relative(basePath, fullPath);

                    if (entry.isDirectory()) {
                        if (!IGNORED_DIRS.has(entry.name)) {
                            results = results.concat(await readDirectoryRecursive(fullPath, basePath));
                        }
                    } else if (entry.isFile()) {
                        try {
                            const stats = await fs.stat(fullPath);
                            if (stats.size > MAX_FILE_SIZE) {
                                console.warn(`Skipping large file: ${relativePath} (${(stats.size / (1024*1024)).toFixed(2)}MB)`);
                                continue;
                            }
                            // Basic check for binary - improve later if needed
                            const content = await fs.readFile(fullPath, 'utf-8');
                            if (content.includes('\uFFFD')) { // Replacement character often indicates binary
                                console.warn(`Skipping potentially binary file: ${relativePath}`);
                                continue;
                            }
                            results.push({ path: relativePath, content });
                        } catch (readError) {
                            console.warn(`Could not read file ${relativePath}: ${readError instanceof Error ? readError.message : String(readError)}`);
                        }
                    }
                }
            } catch (dirError) {
                 console.warn(`Could not read directory ${dirPath}: ${dirError instanceof Error ? dirError.message : String(dirError)}`);
            }
            return results;
        }
        ```

-   **Create `formatCombinedInput` function:**
    -   Input: `promptContent: string`, `contextFiles: Array<{ path: string; content: string }>`
    -   Output: `string` (combined content)
    -   Logic: Implement the chosen formatting strategy.
        ```typescript
        function formatCombinedInput(promptContent: string, contextFiles: Array<{ path: string; content: string }>): string {
            let combined = `PROMPT:\n${promptContent}\n\n`;
            if (contextFiles.length > 0) {
                combined += "CONTEXT FILES:\n\n";
                contextFiles.forEach(file => {
                    combined += `--- START FILE: ${file.path} ---\n`;
                    combined += `${file.content}\n`;
                    combined += `--- END FILE: ${file.path} ---\n\n`;
                });
            }
            return combined.trim();
        }
        ```

-   **Modify `_processInput` helper (`src/workflow/runThinktankHelpers.ts`):**
    -   Accept `contextPaths: string[]` as input.
    -   Call `processInput` for the main prompt file.
    -   If `contextPaths` exist, call `readContextPaths`.
    -   Call `formatCombinedInput` to merge prompt and context.
    -   Return the combined content in the `InputResult`. Update `ProcessInputResult` interface.

### 3.3. Workflow Updates (`src/workflow/runThinktank.ts`, `src/workflow/runThinktankHelpers.ts`)

-   Update `RunOptions` interface to include `contextPaths?: string[]`.
-   Pass `contextPaths` from `runThinktank` options to the `_processInput` helper call.
-   Update `ProcessInputResult` interface in `runThinktankTypes.ts` to reflect the combined content.
-   Update `_executeQueries` parameters (`ExecuteQueriesParams`) to accept the combined `promptContent: string` instead of just `prompt: string`.
-   Modify the call to `_executeQueries` in `runThinktank` to pass `inputResult.content` (which now contains the combined prompt+context).

### 3.4. Provider Adaptation (`src/providers/*.ts`)

-   No explicit code changes are likely needed in the `generate` methods *initially*, as they already accept a string prompt.
-   **Consideration:** LLMs have context window limits. If the combined prompt+context exceeds these limits, the API calls will fail.
    -   *Initial approach:* Let the API calls fail naturally. The existing error handling should catch this (e.g., token limit errors).
    -   *Future enhancement:* Add logic (perhaps in `queryExecutor` or before) to check the size of the combined content against known model limits and either truncate or warn the user.

## 4. Potential Challenges & Considerations

-   **Context Window Limits:** The combined size of the prompt and context files might exceed the LLM's context window limit. This needs clear error handling or potentially automatic truncation/warning.
-   **Performance:** Reading many files or large directories can be slow and memory-intensive. Implement efficient reading and consider streaming or chunking for very large files (defer if not immediately needed).
-   **Context Formatting:** The chosen formatting (`--- START FILE ---`, etc.) needs to be clear and effective for the LLM to understand the separation between the prompt and different context files. Experimentation might be needed.
-   **File Types:** Handling binary files, very large files, or specific file types (e.g., images, videos) needs a defined strategy (skip, error, attempt text extraction). Skipping non-text files is a safe starting point.
-   **Ignoring Files/Dirs:** Parse .gitignore, use that
-   **Error Handling:** Need robust error handling for file not found, permission denied, read errors during context processing.
-   **User Experience:** Providing many paths on the CLI can be cumbersome. Consider alternative ways to specify context in the future (e.g., via config file).

## 5. Testing Strategy

-   **Unit Tests:**
    -   `inputHandler` / context reader utility:
        -   Test `readDirectoryRecursive` with nested directories, various file types, ignored directories, permission errors (mocked), large file skipping.
        -   Test `formatCombinedInput` with zero, one, and multiple context files.
    -   `_processInput` helper: Test integration of prompt reading and context reading/formatting.
    -   `run.ts` (CLI command): Mock `runThinktank` and verify `contextPaths` are parsed correctly from `process.argv`.
-   **Integration Tests:**
    -   `runThinktank`: Test the full workflow with mocked providers, ensuring the combined context string is correctly passed through `_processInput` to `_executeQueries`. Test scenarios with context paths provided and omitted.
-   **E2E Tests (`cli.e2e.test.ts`):**
    -   Create temporary files and directories.
    -   Run the compiled `thinktank` binary using `execa`.
    -   Test `thinktank run <prompt> <file>`
    -   Test `thinktank run <prompt> <dir>`
    -   Test `thinktank run <prompt> <file1> <dir1> <file2>`
    -   Test with paths containing spaces or special characters.
    -   Test with non-existent context paths (should likely warn and proceed).
    -   Verify output files (if `-o` is used) contain expected content reflecting the context (mocked response).
-   **Manual Testing:**
    -   Test with real code repositories as context.
    -   Test with large text files.
    -   Test edge cases like empty files, empty directories.

## 6. Open Questions

1.  **Final Context Formatting:** Is the proposed `--- START FILE --- ... --- END FILE ---` format optimal? Should relative paths be used within the context block? (Decision: Start with the proposed format using relative paths.)
2.  **Context Size Limits:** Should we implement proactive size checking/truncation, or rely on API errors initially? (Decision: Rely on API errors initially, add warnings/limits later if needed.)
3.  **File Encoding:** Assume UTF-8 for now? How to handle other encodings? (Decision: Assume UTF-8. Log warnings for files that fail UTF-8 decoding.)
```
