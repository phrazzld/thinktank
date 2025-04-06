```markdown
# PLAN: Refactor runThinktank Workflow (T6)

## 1. Overview

This plan details the refactoring of the `runThinktank` function located in `src/workflow/runThinktank.ts`. The goal is to improve its readability and maintainability by breaking it down into smaller, more focused logical steps, as outlined in `REFACTOR_PLAN.md` task T6. This refactoring will specifically focus on:

*   Decomposing the main function into distinct phases (setup, input processing, model selection, query execution, output handling).
*   Leveraging the new centralized error handling system (`src/core/errors.ts`) for consistent error reporting and propagation.
*   Ensuring proper resource management, particularly for the `ora` spinner, to prevent hangs and provide clear user feedback during execution, including failures.

## 2. Task Breakdown

| Task                                              | Description                                                                                                                                                              | Affected Files/Modules                                                                                                                                                              | Effort |
| :------------------------------------------------ | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :----- |
| **T6.1: Analyze Current Flow**                    | Map out the existing sequence of operations within `runThinktank`, identifying distinct logical phases and dependencies.                                                 | `src/workflow/runThinktank.ts`                                                                                                                                                      | S      |
| **T6.2: Define Helper Functions**                 | Create private helper functions within `runThinktank.ts` corresponding to the identified phases (e.g., `_setupWorkflow`, `_processInput`, `_selectModels`, `_executeQueries`, `_processOutput`). | `src/workflow/runThinktank.ts`                                                                                                                                                      | M      |
| **T6.3: Refactor Main `runThinktank` Function**   | Rewrite the main `runThinktank` function to orchestrate calls to the new helper functions. Implement a top-level `try...catch` block for workflow-level error handling. | `src/workflow/runThinktank.ts`                                                                                                                                                      | M      |
| **T6.4: Implement Error Propagation**             | Ensure each helper function catches errors from the modules it calls (e.g., `configManager`, `inputHandler`) and throws specific `ThinktankError` subtypes. The main `catch` block should handle these errors appropriately, potentially re-throwing them for the CLI handler. | `src/workflow/runThinktank.ts`, `src/core/errors.ts`                                                                                                                                | L      |
| **T6.5: Manage Spinner Lifecycle**                | Refactor `ora` spinner usage. Ensure the spinner is started, its text updated for each phase, and reliably stopped (`succeed`, `fail`, `warn`, `info`) in both success and error paths within the main function and helper functions. | `src/workflow/runThinktank.ts`, `ora`                                                                                                                                             | M      |
| **T6.6: Resource Cleanup Review**                 | Review if any resources (like SDK clients, though unlikely with current providers) need explicit cleanup. Ensure the refactored flow facilitates graceful termination and addresses the "hanging" issue noted in `BACKLOG.md`. | `src/workflow/runThinktank.ts`, `src/providers/*`                                                                                                                                   | S      |
| **T6.7: Logging Integration**                     | Ensure `logger` calls (`info`, `warn`, `debug`) are placed appropriately within the new structure for clarity.                                                           | `src/workflow/runThinktank.ts`, `src/utils/logger.ts`                                                                                                                               | S      |
| **T6.8: Update Unit & Integration Tests**         | Refactor existing tests for `runThinktank.ts` to reflect the new structure. Add unit tests for the new helper functions, mocking their dependencies. Enhance integration tests to verify the orchestration flow and error handling. | `src/workflow/__tests__/runThinktank.test.ts`, `src/workflow/__tests__/runThinktank-error-handling.test.ts`, `src/workflow/__tests__/error-propagation.test.ts` | L      |

**Effort Estimation:** S = Small (<1 day), M = Medium (1-3 days), L = Large (3-5 days)

## 3. Implementation Details

### 3.1. Refactored `runThinktank` Structure

The main `runThinktank` function will be simplified to orchestrate calls to private helper functions, each responsible for a specific phase of the workflow.

```typescript
// src/workflow/runThinktank.ts (Pseudocode)
import ora from 'ora';
import { logger } from '../utils/logger';
import { ThinktankError, ConfigError, FileSystemError, ApiError } from '../core/errors';
// ... other imports

export async function runThinktank(options: RunOptions): Promise<string> {
  const spinner = ora('Starting thinktank...').start();
  let outputDirectoryPath: string | undefined; // Keep track for cleanup/logging on error

  try {
    // Phase 1: Setup (Config, Run Name, Output Dir)
    const { config, friendlyRunName, outputDir } = await _setupWorkflow(options, spinner);
    outputDirectoryPath = outputDir; // Store path

    // Phase 2: Input Processing
    const inputResult = await _processInput(options, spinner);

    // Phase 3: Model Selection
    const selectionResult = await _selectModels(config, options, spinner);
    if (selectionResult.models.length === 0) {
      // Handle case where no models are available/selected gracefully
      spinner.warn('No models selected or available for querying.');
      return 'No models were selected for querying.';
    }

    // Phase 4: Query Execution
    spinner.text = `Querying ${selectionResult.models.length} model(s)...`;
    const queryResults = await _executeQueries(config, selectionResult, inputResult, options, spinner);

    // Phase 5: Output Processing (File writing & Console formatting)
    spinner.text = 'Processing outputs...';
    const { consoleOutput, fileOutputResult } = await _processOutput(queryResults, outputDirectoryPath, options, friendlyRunName, spinner);

    // Final success message (handled within _processOutput or here)
    spinner.succeed(`Run '${friendlyRunName}' completed successfully.`);
    // Log final summary info (e.g., file paths)
    _logCompletionSummary(fileOutputResult, friendlyRunName);

    return consoleOutput;

  } catch (error) {
    // Centralized error handling for the entire workflow
    spinner.fail('Workflow failed.'); // Ensure spinner stops on error
    return _handleWorkflowError(error, outputDirectoryPath); // Re-throw or format error
  }
}

// --- Helper Function Definitions ---

async function _setupWorkflow(options: RunOptions, spinner: ora.Ora) {
  spinner.text = 'Loading configuration...';
  const config = await loadConfig({ configPath: options.configPath });
  spinner.succeed('Configuration loaded.');

  spinner.start('Generating run identifier...');
  const friendlyRunName = generateFunName(); // Now synchronous
  spinner.succeed(`Run identifier generated: ${friendlyRunName}`);

  spinner.start('Setting up output directory...');
  const outputDir = await createOutputDirectory({
    outputDirectory: options.output,
    directoryIdentifier: options.specificModel || options.groupName,
    friendlyRunName
  });
  spinner.succeed(`Output directory ready: ${outputDir}`);

  return { config, friendlyRunName, outputDir };
}

async function _processInput(options: RunOptions, spinner: ora.Ora) {
  spinner.start('Processing input...');
  try {
    const inputResult = await processInput({ input: options.input });
    spinner.succeed(`Input processed from ${inputResult.sourceType} (${inputResult.content.length} chars).`);
    return inputResult;
  } catch (error) {
    // Wrap input errors if they are not already ThinktankErrors
    if (!(error instanceof ThinktankError)) {
      throw new FileSystemError(`Failed to process input: ${error instanceof Error ? error.message : String(error)}`, error instanceof Error ? error : undefined);
    }
    throw error; // Re-throw ThinktankError
  }
}

async function _selectModels(config: AppConfig, options: RunOptions, spinner: ora.Ora) {
  spinner.start('Selecting models...');
  try {
    const selectionResult = selectModels(config, {
      models: options.models,
      specificModel: options.specificModel,
      groupName: options.groupName,
      groups: options.groups,
      includeDisabled: true, // Let selection handle disabled logic based on context
      validateApiKeys: true, // Perform API key validation
      throwOnError: true      // Ensure errors are thrown for handling
    });

    // Log warnings from selection
    selectionResult.warnings.forEach(warning => spinner.warn(warning));

    if (selectionResult.models.length === 0) {
      spinner.warn('No models available for querying after selection.');
    } else {
      spinner.succeed(`Selected ${selectionResult.models.length} model(s).`);
      // Log selected models for clarity
      const modelList = selectionResult.models.map(m => `${m.provider}:${m.modelId}`).join(', ');
      logger.info(`Models to be queried: ${modelList}`);
    }
    return selectionResult;
  } catch (error) {
    // Wrap selection errors if they are not already ThinktankErrors
    if (!(error instanceof ThinktankError)) {
      throw new ConfigError(`Failed during model selection: ${error instanceof Error ? error.message : String(error)}`, error instanceof Error ? error : undefined);
    }
    throw error; // Re-throw ThinktankError
  }
}

async function _executeQueries(config: AppConfig, selectionResult: ModelSelectionResult, inputResult: InputResult, options: RunOptions, spinner: ora.Ora) {
  spinner.start(`Querying ${selectionResult.models.length} model(s)...`);
  try {
    const queryResults = await executeQueries(config, selectionResult.models, {
      prompt: inputResult.content,
      systemPrompt: options.systemPrompt,
      enableThinking: options.enableThinking,
      timeoutMs: 660000, // 11 min timeout
      onStatusUpdate: (modelKey, status) => {
        // Update spinner text based on model status
        if (status.status === 'running') spinner.text = `Querying ${modelKey}...`;
        // Success/Fail messages handled by executeQueries logging/spinner updates
      }
    });
    // Log summary within executeQueries or here
    spinner.succeed(`Finished querying ${selectionResult.models.length} model(s).`);
    return queryResults;
  } catch (error) {
     // Wrap execution errors if they are not already ThinktankErrors
    if (!(error instanceof ThinktankError)) {
      throw new ApiError(`Error during query execution: ${error instanceof Error ? error.message : String(error)}`, error instanceof Error ? error : undefined);
    }
    throw error; // Re-throw ThinktankError
  }
}

async function _processOutput(queryResults: QueryExecutionResult, outputDirectoryPath: string, options: RunOptions, friendlyRunName: string, spinner: ora.Ora) {
  spinner.start('Writing output files...');
  try {
    const fileOutputResult = await writeResponsesToFiles(
      queryResults.responses,
      outputDirectoryPath,
      {
        includeMetadata: options.includeMetadata,
        throwOnError: false, // Log errors but don't stop the whole process
        onStatusUpdate: (fileDetail) => {
           if (fileDetail.status === 'success') spinner.succeed(`Wrote file: ${fileDetail.filename}`);
           if (fileDetail.status === 'error') spinner.fail(`Failed to write file: ${fileDetail.filename} - ${fileDetail.error}`);
           spinner.start('Writing output files...'); // Keep spinner going
        }
      }
    );
    spinner.succeed('Finished writing files.');

    const consoleOutput = formatForConsole(queryResults.responses, {
      includeMetadata: options.includeMetadata,
      useColors: options.useColors !== false,
      includeThinking: options.includeThinking,
      useTable: process.env.NODE_ENV !== 'test'
    });

    return { consoleOutput, fileOutputResult };
  } catch (error) {
     // Wrap output errors if they are not already ThinktankErrors
    if (!(error instanceof ThinktankError)) {
      throw new FileSystemError(`Error processing output: ${error instanceof Error ? error.message : String(error)}`, error instanceof Error ? error : undefined);
    }
    throw error; // Re-throw ThinktankError
  }
}

function _logCompletionSummary(fileOutputResult: FileOutputResult, friendlyRunName: string) {
    if (fileOutputResult.failedWrites === 0) {
      logger.success(
        `Run '${friendlyRunName}' completed. ${fileOutputResult.succeededWrites} responses saved to ${fileOutputResult.outputDirectory}`
      );
    } else {
      logger.warn(
        `Run '${friendlyRunName}' completed with issues: ${fileOutputResult.succeededWrites} successful, ${fileOutputResult.failedWrites} failed writes`
      );
      const failedFiles = fileOutputResult.files.filter(file => file.status === 'error');
      logger.error('Files with errors:');
      failedFiles.forEach(file => {
        logger.error(`  - ${file.filename}: ${file.error || 'Unknown error'}`);
      });
       logger.info(`Successful writes saved to ${fileOutputResult.outputDirectory}`);
    }
}


function _handleWorkflowError(error: unknown, outputDirectoryPath?: string): never {
  // Ensure error is a ThinktankError
  let thinktankError: ThinktankError;
  if (error instanceof ThinktankError) {
    thinktankError = error;
  } else if (error instanceof Error) {
    // Attempt to categorize standard errors
    thinktankError = new ThinktankError(`Workflow error: ${error.message}`, { cause: error });
    // TODO: Add categorization logic here if needed, or rely on the CLI handler's categorization
  } else {
    thinktankError = new ThinktankError('Unknown workflow error occurred.');
  }

  // Log the specific error context if available
  if (outputDirectoryPath) {
    logger.error(`Error occurred. Partial results might be in: ${outputDirectoryPath}`);
  }

  // Re-throw the structured error for the CLI handler
  throw thinktankError;
}
```

### 3.2. Error Handling Strategy

*   Each helper function (`_setupWorkflow`, `_processInput`, etc.) will be responsible for catching errors from the modules it calls (e.g., `loadConfig`, `processInput`).
*   If the caught error is *not* already a `ThinktankError` subtype, the helper function will wrap it in an appropriate `ThinktankError` (e.g., `ConfigError`, `FileSystemError`, `ApiError`) providing context.
*   The helper function will then re-throw the `ThinktankError`.
*   The main `runThinktank` function's `try...catch` block will catch these propagated `ThinktankError`s.
*   The `_handleWorkflowError` function will be called from the main `catch` block. It ensures the error is a `ThinktankError` and re-throws it. The actual user-facing formatting and exit logic resides in the `handleError` function in `src/cli/index.ts`, which will catch this final thrown error.

Example `catch` block within a helper:

```typescript
async function _processInput(options: RunOptions, spinner: ora.Ora): Promise<InputResult> {
  spinner.start('Processing input...');
  try {
    const inputResult = await processInput({ input: options.input });
    spinner.succeed(`Input processed...`);
    return inputResult;
  } catch (error) {
    if (error instanceof ThinktankError) { // Includes InputError
      throw error; // Re-throw if already specific
    } else if (error instanceof Error) {
      // Wrap generic error in a specific ThinktankError subtype
      throw new FileSystemError(`Failed to process input: ${error.message}`, error);
    } else {
      throw new ThinktankError('Unknown error processing input.');
    }
  }
}
```

### 3.3. Spinner Management

*   A single `ora` instance will be created at the beginning of `runThinktank`.
*   Each helper function will receive the `spinner` instance as an argument.
*   Helper functions will update `spinner.text` at the start of their operation.
*   Helper functions will call `spinner.succeed()` upon successful completion *before* returning.
*   The main `catch` block in `runThinktank` *must* call `spinner.fail()` to ensure the spinner stops if any helper throws an error.
*   Callbacks like `onStatusUpdate` in `executeQueries` and `writeResponsesToFiles` should update the spinner text or use `spinner.info/warn/succeed/fail` for intermediate steps, always restarting the spinner (`spinner.start()`) if necessary after a non-terminating status update.

## 4. Potential Challenges & Considerations

*   **State Management:** Passing necessary data (config, input results, selected models) between helper functions needs careful handling to avoid prop drilling or overly complex function signatures. Consider returning a context object from setup.
*   **Error Granularity:** Ensuring the *correct* `ThinktankError` subtype is thrown from each stage is crucial for the CLI's `handleError` function to provide relevant suggestions.
*   **Spinner Flicker:** Frequent starting/stopping/updating of the spinner text might cause flickering in the terminal. Optimize updates where possible.
*   **Resource Cleanup:** While this refactor focuses on `runThinktank`, the underlying "hanging" issue might stem from provider SDKs not closing connections properly. This refactor might make the issue easier to debug but may not fix it directly. Ensure no unhandled promise rejections are left hanging.
*   **Testing Complexity:** Mocking dependencies for the new helper functions and testing the error propagation flow requires careful test setup.

## 5. Testing Strategy

*   **Unit Tests:**
    *   Create new unit tests for each helper function (`_setupWorkflow`, `_processInput`, etc.).
    *   Mock the modules/functions called by each helper (e.g., mock `loadConfig` for `_setupWorkflow`).
    *   Verify that helpers correctly process inputs, return expected outputs, and update the spinner text.
    *   Test error handling within each helper: ensure they catch errors from dependencies and throw the correct `ThinktankError` subtype.
*   **Integration Tests (`runThinktank.test.ts`):**
    *   Refactor existing tests to work with the new structure.
    *   Mock the *helper* functions (`_setupWorkflow`, etc.) to test the main orchestration logic in `runThinktank`.
    *   Verify the sequence of calls to helper functions.
    *   Test the main `try...catch` block: simulate errors thrown by helpers and verify the correct final error is thrown and the spinner is stopped.
    *   Test the "no models selected" path.
*   **E2E Tests:**
    *   Run the existing E2E test suite (`cli.e2e.test.ts`) to ensure no regressions in overall CLI behavior, output formatting, or error reporting.
*   **Manual Testing:**
    *   Run commands with various options (`--models`, `--group`, `--output`, error conditions like invalid files/models) and observe spinner behavior and error messages.
    *   Verify the "hanging" issue is improved or unchanged.

## 6. Open Questions

1.  **Error Handling in `runThinktank`:** Should the main `catch` block in `runThinktank` attempt any specific error handling/logging, or should it *purely* rely on `_handleWorkflowError` to re-throw for the CLI's `handleError`? (Current plan: Re-throw via `_handleWorkflowError`).
2.  **Resource Cleanup Deep Dive:** Does this refactor provide enough insight to pinpoint the "hanging" issue, or is a separate investigation needed, potentially involving profiling or deeper SDK analysis? (Assume separate investigation might be needed if refactor doesn't solve it).
```