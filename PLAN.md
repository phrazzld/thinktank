```markdown
# PLAN: Improve UI/UX of Console Output

## 1. Overview

This plan details the steps to enhance the console output of the `thinktank` CLI tool. The goal is to improve user experience by making the output more informative, structured, readable, and visually appealing. This involves refining progress indicators, structuring results display (especially with groups), improving error message clarity, and integrating dependencies effectively, aligning with the existing architecture and atomic design principles. This plan supersedes the previous `PLAN.md` and integrates tasks from `TODO.md`.

## 2. Task Breakdown

| Task                                                 | Description                                                                                                | Effort | Primary Files/Modules Affected                                        | Dependencies        |
| :--------------------------------------------------- | :--------------------------------------------------------------------------------------------------------- | :----- | :-------------------------------------------------------------------- | :------------------ |
| **Setup & Core Primitives**                          |                                                                                                            |        |                                                                       |                     |
| Install Dependencies                                 | Add `cli-table3` and `figures` to `package.json`. Verify/update `chalk` and `ora` versions.                | S      | `package.json`, `package-lock.json`                                   | -                   |
| Create Console Utility Module                        | Create `src/atoms/consoleUtils.ts` to centralize `chalk`, `figures`, and common formatting helpers.        | M      | `src/atoms/consoleUtils.ts`                                           | Installed deps      |
| **Enhanced Progress Indicators**                     |                                                                                                            |        |                                                                       |                     |
| Implement Stage Timing                               | Add start/end timestamps for major stages (config load, model processing, file writing) in `runThinktank`. | M      | `src/templates/runThinktank.ts`                                       | -                   |
| Refine Spinner Messages                              | Update `ora` spinner text in `runThinktank` to show counts (current/total), percentages, and stage timing. | M      | `src/templates/runThinktank.ts`                                       | Stage Timing        |
| **Group-Based Output Organization**                  |                                                                                                            |        |                                                                       |                     |
| Add In-Progress Group Headers                        | In `runThinktank`, before processing models of a group, print a formatted header (name, count, description). | M      | `src/templates/runThinktank.ts`, `src/atoms/consoleUtils.ts`          | Console Utils       |
| **Structured Results Display**                       |                                                                                                            |        |                                                                       |                     |
| Design Final Summary Table Structure                 | Define columns (Model, Group, Status, Time, Tokens, etc.) and formatting rules (colors, symbols).          | S      | Design Document / This Plan                                           | -                   |
| Implement Tabular Results Formatter                  | Create/update function (e.g., in `outputFormatter` or new molecule) using `cli-table3` for the final summary. | L      | `src/molecules/outputFormatter.ts` (or new), `src/atoms/consoleUtils.ts` | Console Utils, Deps |
| Integrate Performance Metrics into Table             | Ensure response time and token counts (if available in `LLMResponse.metadata`) are included in the table.    | S      | Tabular Formatter function                                            | LLMResponse changes |
| Calculate and Display Group Summary Stats            | Compute success/error counts and avg. response time per group for display (potentially below table/headers). | M      | `src/templates/runThinktank.ts` (data aggregation), Formatter function | Group Headers, Data |
| **Better Error Handling Display**                    |                                                                                                            |        |                                                                       |                     |
| Create Error Formatting Helper                       | Develop a function in `consoleUtils` or a dedicated error molecule to format errors consistently.          | M      | `src/atoms/consoleUtils.ts` (or new), `src/templates/runThinktank.ts` | Console Utils       |
| Categorize & Color-Code Errors                       | Enhance error messages logged during `runThinktank` (API, config, file write) using the helper.            | M      | `src/templates/runThinktank.ts`, Error Formatter                      | Error Formatter     |
| Add Contextual Troubleshooting Tips                  | Implement logic to detect common error patterns (e.g., API key issues) and suggest fixes via the helper.   | M      | Error Formatter, `src/templates/runThinktank.ts`                      | Error Formatter     |
| **Integration & Refinement**                         |                                                                                                            |        |                                                                       |                     |
| Integrate Formatting into `runThinktank`             | Replace/augment existing `console.log` and `spinner` calls with new utilities and formatters.              | L      | `src/templates/runThinktank.ts`                                       | All previous tasks  |
| Refactor `outputFormatter.ts`                        | Adapt or replace existing functions to use the new table formatter and console utils. Remove redundancy.   | M      | `src/molecules/outputFormatter.ts`                                    | Tabular Formatter   |
| Implement `--verbose` Flag                           | Add flag handling in `cli.ts` and conditional detailed logging (e.g., full metadata) in `runThinktank`.    | M      | `src/runtime/cli.ts`, `src/templates/runThinktank.ts`                 | -                   |
| **Documentation & Polish (Lower Priority)**          |                                                                                                            |        |                                                                       |                     |
| Add First-Run Usage Hints                            | Implement simple mechanism (e.g., check for a state file) to display tips on first execution.              | S      | `src/runtime/cli.ts`                                                  | -                   |
| Add Documentation Links                              | Include relevant documentation links in error messages or help output where appropriate.                   | S      | Error Formatter, `src/runtime/cli.ts`                                 | -                   |
| Implement Fallback Formatting (Optional)             | Add detection for limited terminals and provide simpler, non-color/non-Unicode output.                     | M      | `src/atoms/consoleUtils.ts`, `src/templates/runThinktank.ts`          | Console Utils       |

*Effort Estimation: S = Small (<= 2 hours), M = Medium (2-6 hours), L = Large (6+ hours)*

## 3. Implementation Details

### 3.1. Console Utility Module (`src/atoms/consoleUtils.ts`)

*   **Purpose:** Centralize terminal styling and symbol usage. Avoid direct `chalk` and `figures` calls elsewhere.
*   **Exports:**
    *   Re-export `chalk` instance (configured based on terminal support/flags).
    *   Export common symbols from `figures` (e.g., `tick`, `cross`, `warning`, `info`).
    *   Helper functions:
        *   `styleSuccess(text: string)`: Applies green color, maybe prepends `figures.tick`.
        *   `styleError(text: string)`: Applies red color, maybe prepends `figures.cross`.
        *   `styleWarning(text: string)`: Applies yellow color, maybe prepends `figures.warning`.
        *   `styleInfo(text: string)`: Applies blue/cyan color, maybe prepends `figures.info`.
        *   `styleHeader(text: string)`: Applies bold, maybe blue color.
        *   `styleDim(text: string)`: Applies dim styling.
        *   `formatError(error: Error | string, category?: string, tip?: string)`: Standardized error display (see Error Handling).

```typescript
// src/atoms/consoleUtils.ts (Example Snippet)
import chalk from 'chalk'; // Use compatible version (v4 based on package.json)
import figures from 'figures'; // Use compatible version (v6 based on package.json)

// TODO: Add logic to disable colors based on terminal support or flags
const enabledChalk = new chalk.Instance({ level: 2 }); // Example: Force color level

export const colors = enabledChalk;
export const symbols = figures;

export function styleSuccess(text: string): string {
  return `${colors.green(symbols.tick)} ${text}`;
}

export function styleError(text: string): string {
  return `${colors.red(symbols.cross)} ${text}`;
}

// ... other style helpers

export function formatError(
    error: Error | string,
    category?: string,
    tip?: string
): string {
    const errorMsg = error instanceof Error ? error.message : error;
    let output = `${colors.red.bold('Error')}${category ? ` (${colors.yellow(category)})` : ''}: ${errorMsg}`;
    if (tip) {
        output += `\n  ${colors.cyan(symbols.info)} Tip: ${tip}`;
    }
    // Add stack trace in verbose mode?
    return output;
}
```

### 3.2. Enhanced Progress Indicators (`runThinktank.ts`)

*   Measure time using `performance.now()` or `Date.now()` at the start/end of logical stages (config load, model prep, API calls, file writing).
*   Update `ora` spinner text dynamically:
    *   During model processing: `spinner.text = \`Processing models [${completed}/${total}] (${percent}%): ${currentModelKey} - ${elapsedTime}s\`;`
    *   During file writing: `spinner.text = \`Writing files [${written}/${total}] (${percent}%): ${currentFilename} - ${elapsedTime}s\`;`
*   Use `spinner.info()`, `spinner.succeed()`, `spinner.warn()`, `spinner.fail()` with messages formatted using `consoleUtils`.

### 3.3. Group-Based Output Organization (`runThinktank.ts`, Formatter)

*   **In-Progress Headers:** Before the loop processing models within a specific group in `runThinktank.ts`:
    ```typescript
    // Inside runThinktank, when iterating through groups or models mapped to groups
    import { styleHeader, styleDim } from '../atoms/consoleUtils';
    // ...
    console.log(styleHeader(`\n📋 Group: ${groupName} (${groupModels.length} models)`));
    if (groupDescription) {
        console.log(styleDim(groupDescription));
    }
    console.log(styleDim('─'.repeat(80)));
    // ... start processing models in this group
    ```
*   **Final Summary Stats:** After `Promise.all(callPromises)` and `Promise.all(fileWritePromises)`, aggregate results by group. Calculate success/error counts and average response times per group. Display this information either before or after the main results table, or potentially integrated if the table library supports group summaries.

### 3.4. Structured Results Display (Formatter Module)

*   Use `cli-table3`. Define columns precisely.
*   Create a dedicated function `formatResultsTable(results: Array<LLMResponse & { configKey: string }>): string`.
*   Map `LLMResponse` data to table rows. Use `consoleUtils` for styling status, group names, etc.
*   Handle missing data gracefully (e.g., '-' for missing tokens/time).

```typescript
// src/molecules/outputFormatter.ts (or new module) - Example Snippet
import Table from 'cli-table3';
import { LLMResponse } from '../atoms/types';
import { colors, symbols, styleSuccess, styleError, styleWarning, styleDim } from '../atoms/consoleUtils';

export function formatResultsTable(results: Array<LLMResponse & { configKey: string }>): string {
  const table = new Table({
    head: [
      colors.bold('Model'),
      colors.bold('Group'),
      colors.bold('Status'),
      colors.bold('Time (ms)'),
      colors.bold('Tokens'),
    ],
    colAligns: ['left', 'left', 'center', 'right', 'right'],
    style: { head: ['blue'] } // Example styling
  });

  // Sort results perhaps by group then model?
  results.sort((a, b) => {
      const groupA = a.groupInfo?.name || 'default';
      const groupB = b.groupInfo?.name || 'default';
      if (groupA !== groupB) return groupA.localeCompare(groupB);
      return a.configKey.localeCompare(b.configKey);
  });

  results.forEach(result => {
    let statusText: string;
    if (result.error) {
      // Consider differentiating API errors vs other errors if possible
      statusText = styleError('Error');
    } else {
      statusText = styleSuccess('Success');
      // Add warning symbol if metadata contains warnings?
    }

    table.push([
      result.configKey,
      result.groupInfo?.name || styleDim('default'),
      statusText,
      result.metadata?.responseTime ?? styleDim('-'),
      result.metadata?.usage?.total_tokens ?? styleDim('-'),
    ]);
  });

  return table.toString();
}

// Update formatResults in outputFormatter.ts to call formatResultsTable
// Or replace formatResults entirely if the table is the primary output now.
```

### 3.5. Better Error Handling Display (`runThinktank.ts`, `consoleUtils.ts`)

*   Wrap error-prone operations (API calls, file writes) in `try...catch`.
*   Inside `catch` blocks, use the `formatError` helper from `consoleUtils`.
*   Attempt to categorize errors based on context or error type/message:
    *   API Key Error: Check `error.message` for keywords like 'API key', 'authentication', '401', '403'. Suggest checking `.env` or config.
    *   Network Error: Check for 'ECONNREFUSED', 'ETIMEDOUT', 'ENOTFOUND'. Suggest checking internet connection or service status.
    *   Configuration Error: Catch errors during `loadConfig` or model validation. Suggest checking `thinktank.config.js`.
    *   File System Error: Catch errors during `fs.mkdir` or `fs.writeFile`. Suggest checking permissions or disk space.
*   Pass category and relevant tip to `formatError`.
*   Log formatted errors using `console.error` or `spinner.fail()`.

### 3.6. Dependency Integration

*   Ensure `package.json` reflects the correct, compatible versions:
    *   `chalk`: Currently `^4.1.2`. Plan mentioned `^5.0.0`. **Decision:** Stick with v4 for now unless v5 features are essential, as v5 is ESM-only and the project is CJS (`"type": "commonjs"`). If upgrading, requires build/runtime changes. *Assume v4 for now.*
    *   `cli-table3`: Plan `^0.6.3`, actual `^0.6.5`. Use `^0.6.5`.
    *   `figures`: Plan `^5.0.0`, actual `^6.1.0`. Use `^6.1.0`.
    *   `ora`: Plan "existing", actual `^5.4.1`. Use `^5.4.1`.
*   Import dependencies correctly (e.g., `import chalk from 'chalk';` for v4).

## 4. Potential Challenges & Considerations

*   **Refactoring `runThinktank.ts`:** This file has significant orchestration logic. Adding detailed progress, group headers, and improved error logging requires careful integration without breaking existing functionality.
*   **Terminal Compatibility:** Ensuring consistent rendering of colors, symbols, and table borders across different terminals (macOS Terminal, iTerm2, Windows Terminal, VS Code integrated, basic Linux TTY) can be tricky. The `chalk` library helps, but testing is crucial. Fallback formatting adds complexity.
*   **Information Overload:** Balancing detailed information (timing, tokens, verbose logs) with clarity. The table format helps, but too many columns or rows could become unwieldy. Consider truncation or alternative displays for very large result sets.
*   **Performance:** While likely negligible, complex formatting and frequent spinner updates could theoretically add minor overhead. Profile if any slowdown is perceived.
*   **State Management for Progress:** Accurately tracking completed/pending/failed counts across concurrent operations needs careful state management within `runThinktank.ts`.
*   **Atomic Design Adherence:** Ensure the `consoleUtils` remains truly atomic (basic reusable elements) and doesn't creep into molecule-level concerns. Decide if the table formatter is a complex atom or a molecule.
*   **Error Categorization Reliability:** Reliably categorizing errors based on messages can be brittle. Rely on error types or codes where possible.

## 5. Testing Strategy

*   **Unit Tests:**
    *   `src/atoms/consoleUtils.ts`: Test all helper functions (`styleSuccess`, `formatError`, etc.) with various inputs. Mock `chalk` and `figures` to assert correct calls and output structure, independent of actual terminal rendering.
    *   `src/molecules/outputFormatter.ts` (or new table formatter module): Test `formatResultsTable` with different `results` data (empty, single, multiple, errors, missing metadata, different groups). Mock `cli-table3` to verify constructor options and `push` calls. Assert the structure of the returned string (or table object).
*   **Integration Tests:**
    *   `src/templates/runThinktank.ts`: Mock file system operations (`fs`), API calls (`provider.generate`), and configuration loading. Verify that:
        *   Correct spinner messages (including counts, timing) are generated at different stages.
        *   Group headers are logged correctly based on mocked config/results.
        *   `formatResultsTable` (or equivalent) is called with the correctly aggregated results data.
        *   `formatError` is called with appropriate arguments for simulated errors.
    *   `src/runtime/cli.ts`: Use `execa` or similar to run the CLI command with different arguments (`--group`, `--model`, `--verbose`, invalid inputs). Assert on the structure and key elements of the console output (presence of table, headers, error messages). Mock `runThinktank` to control its output/errors.
*   **Manual Testing:**
    *   Run `thinktank` commands in various real terminal emulators (as listed in Challenges).
    *   Test with different configurations: no groups, multiple groups, models inside/outside groups.
    *   Test with varying numbers of models (0, 1, 10, 50+).
    *   Trigger real errors: invalid API key, non-existent model, network offline (if possible), invalid config file. Verify error display and tips.
    *   Test with `--verbose` flag.
    *   Test with long group names/descriptions or model IDs.
    *   Verify output redirection (`thinktank ... > output.log`) produces clean text without ANSI codes (or test color disabling flags if implemented).

## 6. Open Questions

1.  **Dependency Versions:** Confirm sticking with `chalk` v4 due to CJS compatibility, or if migrating the project build to handle ESM v5 is desired/planned.
2.  **Table Width:** How should the results table handle very long model names or potentially many columns if more metadata is added? Truncate? Wrap? Allow horizontal scrolling (less ideal for CLI)?
3.  **Verbose Mode Content:** What specific extra information should `--verbose` enable? Full API request/response? Detailed timing breakdowns? Debug logs from dependencies? Stack traces for errors?
4.  **Fallback Formatting:** Is the "Fallback formatting for limited terminals" task (low priority) a firm requirement, or can we assume modern terminal capabilities for now?
5.  **First-Run Hint Mechanism:** Is storing a simple state file (e.g., `~/.thinktank_state`) acceptable for detecting the first run, or is another mechanism preferred?
6.  **Error Categorization Specificity:** How granular should error categorization be (e.g., differentiate between 401/403/429 API errors)?

```