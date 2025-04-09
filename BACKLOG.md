# BACKLOG

This backlog tracks planned work for the `thinktank` project, categorized by focus area.

## PRIORITY
* **[Infra]** Set up GitHub Actions for CI/CD (linting, testing, building).
* **[DX]** Tidy up generated file structure / format. Don't include group name, for instance.
* **[Feature]** Allow users to add a final *synthesize* step, sending all model outputs to a chosen model for summarization.

## 🧪 Testing & Testability Refactoring

* **[Testing]** Fix skipped tests using Hybrid Virtual FS + Targeted Spies approach:
    * `readDirectoryContents.test.ts`
    * `virtualFsUtils.test.ts`
    * `gitignoreFiltering.test.ts`
    * `gitignoreFilterIntegration.test.ts`
* **[Testing]** Complete `gitignore` integration tests using `memfs` and `addVirtualGitignoreFile`, removing mocks from `readContextPaths.test.ts`.
* **[Testing]** Address complex `gitignore` pattern limitations (investigate, fix, or document).
* **[Testing]** Standardize path handling in all tests using `pathUtils.ts`. Ensure consistency across OS platforms.
* **[Testing]** Eliminate remaining complex mocking patterns (`Object.defineProperty`, etc.), favoring `memfs` or simple stubs/spies.
* **[Testing]** Ensure consistent test isolation (cache clearing, `resetVirtualFs`).
* **[Testing]** Remove `console.log` and other debug code from tests.
* **[Testing]** Develop/expand shared test setup helpers (`fsTestSetup.ts`) based on `TESTING_PHILOSOPHY.md`.
* **[Testing]** Remove all HTTP mocking and replace with proper dependency injections.
* **[Testing]** Update tests to use Dependency Injection (DI) and mock interfaces instead of concrete implementations.
* **[Testing]** Review test coverage post-refactoring and add high-value tests focusing on behavior.

## 🏗️ Architectural Refactoring

* **[Refactor]** Define core interfaces (`LLMClient`, `FileSystem`, `ConfigManager`) in `src/core/`.
* **[Refactor]** Implement Dependency Injection in `workflow` modules and `cli`.
* **[Refactor]** Isolate I/O side effects from core logic (e.g., `_processOutput`, `_logCompletionSummary`).
* **[Refactor]** Refactor CLI command handlers out of `commander` action callbacks.
* **[Refactor]** Simplify cascading configuration, fully centralize to user config.
* **[Refactor]** Decouple dependencies with interfaces.
* **[Refactor]** Centralize constants (API endpoints, default options, etc.).
* **[Refactor]** Increase TypeScript type coverage and reduce usage of `any`.

## ✨ Features & Enhancements

* **[Feature]** Allow users to save task prompts in their config file.
* **[Feature]** Allow users to define an arbitrary number of steps in their task (e.g., Plan -> Critique -> Revise Plan).
* **[Feature]** Support ad-hoc task/prompt definitions (passing a string instead of a filepath).
* **[Feature]** Make it easy to write output to a logfile.
* **[Feature]** Estimate cost per LLM request and log it. Integrate with provider cost APIs if possible.
* **[Feature]** Implement better token count handling (e.g., set max tokens per model, warn if exceeded).
* **[Feature] [DX]** Add interactive mode for running prompts without files.
* **[Feature]** Add output comparison/diffing feature for model responses.
* **[Feature]** Implement streaming support to display responses as they generate.
* **[Feature]** Create a plugin system for custom providers or output formatters.
* **[Feature]** Allow users to define reusable prompt templates in config.
* **[Feature]** Add configurable retry logic for transient API errors.
* **[Feature]** Add optional context summarization step before sending prompt.
* **[Feature]** Fetch and display model capabilities (context length, modalities) in `models list`.

## 🐛 Bug Fixes & DX Improvements

* **[Bug]** Fix program hanging for a period after completing a run.
* **[Bug]** Fix error: `Error from openai:o3-mini: (0 , errors_1.isProviderRateLimitError) is not a function`.
* **[Bug]** Debug intermittent failures/interruptions when invoking `thinktank` with Claude models.
* **[Bug]** Ensure consistent path normalization (`/` vs `\`) across all OS platforms, especially for context paths and config locations.
* **[Bug]** Investigate potential spinner throttling issues (`disable-spinner-throttling` option).
* **[DX]** Improve default configuration initialization (`thinktank config init` or similar).
* **[DX]** Running without a specified group should use the `default` group, not all enabled models.
* **[DX]** Improve CLI UI/UX:
    * Show progress indicators (spinners) for each running model, changing to completion indicators (checkmarks).
    * Fix "double i" icon issue in informational messages.
    * Show progress bar when processing many context files/directories.
* **[DX]** Add `thinktank config validate` command.
* **[DX]** Review and improve clarity and actionability of error messages across the application.
* **[DX]** Enhance `--dry-run` output to show exactly what would be sent to models.
* **[DX]** Add shell autocompletion scripts (bash, zsh, fish).
* **[DX]** Enhance configuration schema validation (Zod) for better error reporting.

## 📚 Documentation & Process

* **[Docs]** Fix broken documentation links in `README.md`.
* **[Docs]** Create/update `TESTING.md` based on `TESTING_PHILOSOPHY.md` and new testing approach (`memfs`/DI).
* **[Docs]** Update `CONTRIBUTING.md` with new testing standards.
* **[Docs]** Add detailed API documentation for extending the tool (new providers, etc.).
* **[Docs]** Create a configuration deep-dive document with advanced examples.
* **[Docs]** Expand the troubleshooting section in `README.md`.

## ⚙️ Infrastructure & Maintenance

* **[Maintenance]** Perform a dependency audit for security and outdated packages.
