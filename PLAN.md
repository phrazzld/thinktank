# Engineering Plan: thinktank CLI Refactoring and Enhancement

## 1. Overview & Goals

The `thinktank` CLI tool provides a valuable capability: querying multiple Large Language Models (LLMs) simultaneously and comparing their outputs. Based on recent critical analyses (Critiques A, B, C), while the tool leverages TypeScript effectively and has a modular concept (especially for LLM providers), its current implementation suffers from several key issues that hinder usability, maintainability, and overall elegance.

Specifically, the strict adherence to atomic design principles feels over-engineered for a CLI context, leading to file fragmentation. The reliance on manual JSON configuration creates a high barrier to entry and is error-prone. The core workflow logic is monolithic, and model selection lacks intuitive ad-hoc capabilities. Handling provider-specific configurations requires significant user effort.

This plan outlines a refactoring and enhancement strategy focused on:

1.  **Simplifying the Code Structure:** Migrating from strict atomic design to a more pragmatic, flatter structure suitable for a CLI application.
2.  **Improving Configuration Management:** Replacing manual JSON editing with a robust, user-friendly CLI command interface and implementing a cascading configuration system for sensible defaults.
3.  **Enhancing Usability:** Streamlining model selection for ad-hoc queries and introducing presets for new users.
4.  **Increasing Maintainability & Readability:** Refactoring the core workflow into smaller, focused modules and improving logging.
5.  **Preserving Core Functionality:** Ensuring all existing capabilities (multi-model querying, groups, provider extensibility) are retained or enhanced.

The end goal is a more elegant, approachable, maintainable, and powerful `thinktank` CLI tool.

## 2. Current State Analysis (Summary of Problems)

Based on the provided critiques, the primary challenges are:

*   **Overly Complex Structure:** Strict atomic design (`atoms`, `molecules`, etc.) leads to excessive file fragmentation and complicates code navigation without significant reuse benefits in this CLI context.
*   **Cumbersome Configuration:** Manual editing of `thinktank.config.json` is required, making it inaccessible for non-technical users, error-prone (syntax/schema errors), and lacking immediate validation.
*   **Monolithic Workflow:** `src/templates/runThinktank.ts` contains overly complex, lengthy logic handling configuration loading, model selection, parallel API calls, and output writing, reducing readability and testability.
*   **Inflexible Model Selection:** Primarily relies on pre-defined groups in the config file, making quick, one-off comparisons across specific models cumbersome.
*   **Difficult Provider-Specific Configuration:** Managing unique options for different models/providers (e.g., Anthropic's `thinking`, OpenAI's modes) requires manual specification within the JSON, lacking sensible defaults or a clear hierarchy.
*   **Code Fragmentation & Verbosity:** Numerous small files and potentially excessive logging clutter the developer experience and runtime output.

## 3. Proposed Architecture & Enhancements

We will implement the following architectural changes and feature enhancements:

### 3.1. Code Structure Simplification

*   **Action:** Abandon the strict atomic design hierarchy. Adopt a flatter, domain-oriented structure:
    *   `src/core/`: Core types (`types.ts`), interfaces (`LLMProvider.ts`), primary logic classes (`ConfigManager.ts`, `LLMRegistry.ts`).
    *   `src/providers/`: Individual LLM provider implementations (e.g., `openai.ts`, `anthropic.ts`, `openrouter.ts`). Provider-specific utilities can be co-located or placed within a `src/providers/utils/` subdirectory if needed, but consolidation is preferred.
    *   `src/cli/`: Command-line interface logic using `commander.js`. Includes the main entry point (`index.ts` or `main.ts`) and command definitions (e.g., `commands/config.ts`, `commands/run.ts`).
    *   `src/utils/`: Common utility functions (e.g., `logger.ts`, `fileUtils.ts`, `helpers.ts`).
    *   `src/workflow/`: Refactored components from the original `runThinktank.ts` (see section 3.5).
*   **Rationale:** Reduces file count, improves navigability, aligns better with typical CLI application structures, and simplifies the mental model for developers.

### 3.2. CLI-Based Configuration Management

*   **Action:** Implement a dedicated `thinktank config` command suite using `commander.js` to manage `thinktank.config.json`.
*   **Commands:**
    *   `thinktank config path`: Display the path to the config file.
    *   `thinktank config show`: Display the current configuration.
    *   `thinktank config models list`: List all configured models in a table format (Provider, ModelId, Enabled, Options).
    *   `thinktank config models add <provider> <modelId> [--options '{"key":"value",...}'] [--enable/--disable]`: Add a new model definition. Use JSON string for options initially, potentially add specific flags like `--temperature`, `--max-tokens` later.
    *   `thinktank config models remove <identifier>`: Remove a model (using index number from `list` or `provider:modelId`).
    *   `thinktank config models enable <identifier>` / `disable <identifier>`: Toggle model availability.
    *   `thinktank config groups list`: List all configured groups.
    *   `thinktank config groups create <groupName> [--prompt "System prompt text"] [--models <modelId1,modelId2...>]`: Create a new group.
    *   `thinktank config groups add-model <groupName> <modelId>`: Add a model to a group.
    *   `thinktank config groups remove-model <groupName> <modelId>`: Remove a model from a group.
    *   `thinktank config groups set-prompt <groupName> --prompt "..."`: Set/update a group's system prompt.
    *   `thinktank config groups remove <groupName>`: Delete a group.
*   **Implementation:**
    *   The `ConfigManager` class (`src/core/ConfigManager.ts`) will handle loading, validating (using Zod or similar for schema validation), modifying, and saving the JSON configuration.
    *   Each command action will load the config, perform the modification via `ConfigManager` methods, validate the result, and save back, providing user feedback.
*   **Benefit:** Eliminates manual JSON editing, provides immediate validation, lowers the barrier to entry, and reduces configuration errors.

### 3.3. Cascading Model Configuration System

*   **Action:** Implement a system to intelligently merge model options from multiple levels, providing sensible defaults and allowing user overrides.
*   **Hierarchy (Lowest to Highest Priority):**
    1.  **Base Defaults:** Universal defaults applicable to all models (e.g., `{ temperature: 0.7, maxTokens: 1000 }`). Hardcoded or loaded from a separate `defaults.json`.
    2.  **Provider Defaults:** Defaults specific to an LLM provider (e.g., Anthropic: `{ thinking: 'enabled' }`).
    3.  **Model-Specific Defaults:** Fine-tuned defaults for a specific model ID (e.g., `claude-3-opus`: `{ temperature: 0.8 }`).
    4.  **User Config Defaults:** Options defined for a model within the `thinktank.config.json` file (via `config models add --options ...`).
    5.  **Group-Specific Overrides:** Options defined within a group definition for a specific model in that group.
    6.  **CLI Invocation Overrides:** Options provided directly via CLI flags during the `run` command (e.g., `thinktank run ... --temperature 0.5`).
*   **Implementation:**
    *   Define a clear `ModelOptions` interface in `src/core/types.ts`.
    *   Create a function `resolveModelOptions(provider: string, modelId: string, userConfigOptions?: ModelOptions, groupOptions?: ModelOptions, cliOptions?: ModelOptions): ModelOptions` within `src/core/ConfigManager.ts` or a dedicated `src/core/optionsResolver.ts`.
    *   This function will fetch defaults (potentially stored in `ConfigManager` or constants) and merge them using object spreading (`{...base, ...provider, ...modelSpecific, ...userConfig, ...group, ...cli}`).
    *   Provider implementations will receive the final, resolved `ModelOptions`.
*   **Benefit:** Simplifies configuration by providing smart defaults while retaining full user control for customization. Handles provider/model nuances gracefully.

### 3.4. Simplified Model Selection

*   **Action:** Introduce a `--models` CLI flag for the main run command to allow direct specification of models for ad-hoc queries.
*   **Syntax:** `thinktank run prompt.txt --models <provider1>:<modelId1>,<provider2>:<modelId2>,...`
*   **Logic:**
    *   If `--models` is provided, use only those models, ignoring enabled status and groups (unless a group name is also provided as the primary target, which might be disallowed for clarity). Resolve their options using the cascading system (Section 3.3).
    *   If `--group <groupName>` is provided, use the models defined in that group.
    *   If neither is provided, use all models marked as `enabled: true` in the configuration.
*   **Implementation:** Update the model selection logic within the refactored workflow (see Section 3.5).
*   **Benefit:** Enables quick, targeted comparisons without needing to pre-configure groups or toggle `enabled` flags.

### 3.5. Workflow Refactoring

*   **Action:** Decompose the monolithic logic in `src/templates/runThinktank.ts` into smaller, single-responsibility functions/modules within the `src/workflow/` directory.
*   **Proposed Modules:**
    *   `inputHandler.ts`: Handles loading the prompt (from file or stdin) and loading the configuration via `ConfigManager`.
    *   `modelSelector.ts`: Implements the logic described in Section 3.4 to determine the final list of models to query based on CLI flags (`--models`, `--group`) and config settings.
    *   `optionsResolver.ts` (or within `ConfigManager`): Implements the cascading configuration logic (Section 3.3).
    *   `queryExecutor.ts`: Takes the list of selected models (with resolved options) and the prompt, uses `LLMRegistry` to get provider instances, and executes API calls in parallel (`Promise.allSettled` for robustness). Includes spinner/progress feedback.
    *   `outputHandler.ts`: Formats the results (successful responses and errors) and writes them to the specified output files/directory or prints to console.
*   **Main Workflow Orchestration:** A simplified function (e.g., `src/cli/commands/run.ts`'s action handler) will call these modules sequentially.
*   **Benefit:** Improves readability, testability, and maintainability by breaking down complexity. Makes future feature additions easier.

### 3.6. Enhanced User Experience

*   **Action:** Implement features to improve ease of use and feedback.
    *   **Presets:** Add a `--preset <presetName>` flag (e.g., `--preset basic`, `--preset coding`) to the `run` command. These presets map to predefined sets of models (potentially including specific options) stored within the application or config defaults. Requires adding preset definitions and updating `modelSelector.ts`.
    *   **Logging Refinement:** Reduce default logging verbosity. Use spinners for active tasks but log only critical steps/summary information by default. Introduce a `--verbose` flag for detailed step-by-step logging useful for debugging. Implement this using a dedicated `logger.ts` utility.
    *   **Error Handling:** Ensure clear, actionable error messages for common issues (invalid config, missing API keys, API errors, file access errors). Leverage `Promise.allSettled` in `queryExecutor.ts` to handle partial failures gracefully.

## 4. Implementation Steps

1.  **Setup:**
    *   Create a new feature branch (e.g., `refactor/simplify-architecture`).
    *   Ensure TypeScript, ESLint, Prettier, and testing frameworks (Jest) are configured correctly.

2.  **Code Structure:**
    *   Create the new directory structure (`src/core`, `src/providers`, `src/cli`, `src/utils`, `src/workflow`).
    *   Move existing files to their new locations (e.g., types to `core`, provider logic to `providers`, entry point logic to `cli`).
    *   Consolidate small utility files where appropriate (e.g., within providers or into `src/utils`).
    *   Update all `import` paths across the project.

3.  **Configuration System:**
    *   Refine/Implement the `ConfigManager` class in `src/core/` with load, save, add, remove, update methods for models and groups. Implement robust validation (e.g., using Zod).
    *   Implement the cascading `resolveModelOptions` function (Section 3.3), including defining base, provider, and model-specific defaults.

4.  **CLI Commands:**
    *   Install/update `commander.js`.
    *   Implement the `thinktank config ...` commands defined in Section 3.2 within `src/cli/commands/config.ts` (or similar), using the `ConfigManager`.

5.  **Model Selection & Workflow:**
    *   Implement the `--models` flag logic in `src/cli/commands/run.ts` and the `modelSelector.ts` module (Section 3.4).
    *   Refactor the existing `runThinktank.ts` logic into the new modules within `src/workflow/` (Section 3.5).
    *   Ensure the main run command orchestrates calls to these workflow modules correctly.

6.  **Provider Updates:**
    *   Update all provider implementations in `src/providers/` to accept the resolved `ModelOptions` object and use it when making API calls. Remove any internal option defaulting that is now handled by the cascading system.

7.  **User Experience Enhancements:**
    *   Implement the `--preset` flag logic.
    *   Refactor logging using a dedicated utility and implement the `--verbose` flag (Section 3.6).
    *   Review and enhance error handling throughout the workflow.

8.  **Testing:**
    *   Write unit tests for `ConfigManager`, `resolveModelOptions`, `modelSelector`, and individual utility functions.
    *   Write integration tests for the `config` commands and the main `run` command workflow (mocking API calls).
    *   Ensure existing tests are updated and passing.

9.  **Documentation:**
    *   Update `README.md` to reflect the new structure, CLI commands (especially `config` and `--models`/`--preset`), and configuration approach.
    *   Add or update user guides/examples.
    *   Document the cascading configuration hierarchy.

10. **Review & Merge:** Conduct code reviews for each major component and merge the feature branch upon successful testing and review.

## 5. Technical Details & Considerations

*   **Dependency Management:** Audit existing dependencies. Remove unused ones. Consider using `npm audit` and potentially `Dependabot` for ongoing maintenance. Keep dependencies minimal where possible.
*   **Configuration File:** The default location for `thinktank.config.json` should be user-configurable (e.g., via environment variable or flag) but default to a standard location (e.g., `~/.config/thinktank/thinktank.config.json`). `ConfigManager` should handle file creation if it doesn't exist.
*   **API Keys:** Continue managing API keys via environment variables or a secure mechanism. Provide clear instructions/errors if keys are missing.
*   **Type Safety:** Leverage TypeScript's strict mode and ensure strong typing throughout, especially around configuration objects and provider interfaces.
*   **Asynchronous Operations:** Use `async/await` consistently. Employ `Promise.allSettled` for parallel API calls to handle individual failures without halting the entire process.
*   **Testing Strategy:** Focus on unit tests for core logic (config management, option resolution) and integration tests for CLI commands and the end-to-end workflow (with mocks).

## 6. Expected Outcomes

*   **Improved Developer Experience:** Easier code navigation, understanding, and maintenance due to a simpler structure and modular workflow.
*   **Enhanced User Experience:** Lower barrier to entry via CLI config management and presets. Increased flexibility via direct model selection (`--models`) and cascading options. Clearer feedback through refined logging and error handling.
*   **Increased Robustness:** Reduced risk of configuration errors due to validation. Better handling of partial failures during execution.
*   **Simplified Maintenance:** Easier to add new providers, commands, or features to the more modular codebase.
*   **Elegant Design:** A more pragmatic and less fragmented architecture better suited to the tool's purpose.

This plan provides a clear roadmap for evolving `thinktank` into a significantly more refined and user-friendly tool while building upon its existing strengths.
