# BACKLOG

## Phase 1: Stabilize Current Testing Refactor
- Fix skipped tests in `readDirectoryContents.test.ts`, `virtualFsUtils.test.ts`, `gitignoreFiltering.test.ts`, `gitignoreFilterIntegration.test.ts`. Use Hybrid Virtual FS + Targeted Spies approach.
- Complete gitignore integration tests using `memfs` and `addVirtualGitignoreFile`, removing mocks from `readContextPaths.test.ts`.
- Address complex gitignore pattern limitations (investigate, fix, or document).

## Phase 2: Standardize and Clean Up Tests
- Standardize path handling in all tests using `pathUtils.ts`.
- Eliminate remaining complex mocking patterns (`Object.defineProperty`, etc.), favoring `memfs` or simple stubs/spies.
- Ensure consistent test isolation (cache clearing, `resetVirtualFs`).
- Remove `console.log` and other debug code from tests.
- Develop/expand shared test setup helpers (`fsTestSetup.ts`) based on `TESTING_PHILOSOPHY.md`.

## Phase 3: Architectural Refactoring for Testability
- Define core interfaces (`LLMClient`, `FileSystem`, `ConfigManager`) in `src/core/`.
- Implement Dependency Injection in `workflow` modules and `cli`.
- Isolate I/O side effects from core logic (`_processOutput`, `_logCompletionSummary`).
- Refactor CLI command handlers out of `commander` action callbacks.
- Update tests to use DI and mock interfaces instead of concrete implementations.

## Phase 4: Continuous Improvement & Documentation
- Integrate `TESTING_PHILOSOPHY.md` into AI agent instructions (`plan.md`, `execute.md`, `review.md`).
- Fix broken documentation links in `README.md`.
- Create/update `TESTING.md` based on `TESTING_PHILOSOPHY.md` and new testing approach (`memfs`/DI).
- Update `CONTRIBUTING.md` with new testing standards.
- Review test coverage post-refactoring and add high-value tests focusing on behavior.

- simplify cascading config, fully centralize to user config [cite: 1248]
- decouple dependencies with interfaces. remove all http mocking and replace with proper dependency injections [cite: 1248]
- isolate side effects. separate i/o from logic. [cite: 1248]
- set up github actions [cite: 1249]
- program still hangs for a while after completing a run [cite: 1249]
- better default config init [cite: 1249]
- running w/o specifying a group should run the default group models -- not every fucking enabled model in the config lol [cite: 1249]
- user should be able to save task prompts in their config [cite: 1249]
- user should be able to define an arbitrary number of steps in their task [cite: 1249]
	* ie task is 1) generate a plan file, 2) critique the plan file, 3) generate a second draft plan file [cite: 1249]
- user should be able to add a _synthesize_ step at the end [cite: 1249]
	* pick a model to send all of the model outputs to for synthesis [cite: 1249]
- fix error: `Error from openai:o3-mini: (0 , errors_1.isProviderRateLimitError) is not a function` [cite: 1249]
- improve cli ui/ux [cite: 1249]
	* show progress indicator for each running model, spinners when in progress -> checkmarks or green circles when completed [cite: 1249]
	* fix "double i" problem (two i icons showing for a lot of the info messages) [cite: 1249]
- support ad-hoc task/prompt definitions (ie passing a string instead of a filepath) [cite: 1249]
- make it easy to write output to logfile [cite: 1249]
- debug why claude code keeps encountering failures / interrupted by user issues when invoking thinktank (but not always??) [cite: 1249]
- better token count handling (maybe set max allowable tokens for each configured model and prevent a request and log a warning if token count exceeds threshold)
- estimate cost per request, log it

