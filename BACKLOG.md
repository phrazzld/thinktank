# BACKLOG

- [ ] **[Refactor]**: Enhance Error Handling and Logging Consistency (Refined T13, Original T38, T47, T74, T75, T81)
  - **Complexity**: Complex
  - **Rationale**: Critical for reliability, debugging, and operational excellence. Adheres to Development Philosophy on Consistent Error Handling (§8), Logging Strategy (§10), and Go Appendix §8 (Error Handling), §11 (Logging). Ensures separation of concerns, rich self-describing errors, structured JSON logs, context propagation, and sanitized error details.
  - **Expected Outcome**: All errors are rich, self-describing types (e.g., `llm.LLMError`). Consistent structured JSON logging via `log/slog` with mandatory context fields (correlation_id, model, phase). Sanitized error details in logs. Improved user-facing error messages. Custom loggers replaced.
  - **Dependencies**: None.
## High Priority

### Core Features & Value Delivery

- [ ] **[Feature]**: Develop Flexible Workflow Engine (Original T25)
  - **Complexity**: Complex
  - **Rationale**: Enables users to define, compose, and automate multi-step AI workflows (e.g., plan → critique → revise → synthesize) via configuration or CLI flags. Central to product value, extensibility, and delivering a delightful software experience.
  - **Expected Outcome**: Users can define and execute custom processing pipelines. The system supports chaining operations, passing context between steps, and potentially includes built-in common workflows.
  - **Dependencies**: Robust Input/Output Handling, Enhanced Error Handling & Logging.
### Technical Excellence & Foundational Improvements

- [ ] **[Refactor]**: Improve Registry API Service Testability and Dependency Inversion (Refined T51, Original T46, T48, T50, T58 part)
  - **Complexity**: Medium
  - **Rationale**: Enhances code structure, testability, and maintainability by adhering to the Dependency Inversion Principle and Mocking Policy (Philosophy §2, §3, §7.3, Go Appendix §7, §10). Eliminates internal mocking.
  - **Expected Outcome**: `internal/thinktank/registry_api.go` and its callers use interfaces for dependencies (registry, logger, etc.) via DI. Improved unit testability without violations of mocking policies.

- [ ] **[Refactor]**: Audit and Fix `context.Context` Propagation & Enable Race Detection (Original T62)
  - **Complexity**: Medium
  - **Rationale**: Ensures correct handling of cancellation, deadlines, and request-scoped data in concurrent operations, and detects race conditions. Mandatory per Go Appendix §9 and Philosophy §9.2.
  - **Expected Outcome**: `context.Context` correctly used as the first argument and propagated throughout the codebase. `go test -race ./...` added to all CI test commands, and any detected races are fixed.

### CI/CD & Automation

- [ ] **[Automation/Process]**: Implement Automated Semantic Versioning via Conventional Commits (Refined T14, Original T56, T57)
  - **Complexity**: Medium
  - **Rationale**: Provides a clear, automated release process, improves maintainability, and adheres to Philosophy §9. Mandatory for robust software delivery. Enforces consistent commit history.
  - **Expected Outcome**: Conventional Commits specification adopted and enforced (e.g., via `commitlint`). CI pipeline automates version bumping, Git tagging, and `CHANGELOG.md` generation.

- [ ] **[Automation/Process]**: Enhance CI Pipeline with Mandatory Quality Gates (Original T59)
  - **Complexity**: Medium
  - **Rationale**: Guarantees minimum code quality and security standards are met before code is merged or released. Mandatory per Philosophy §9.2.
  - **Expected Outcome**: CI workflow enforces mandatory, blocking steps for: Lint, Format Check, Unit Tests, Integration Tests, Coverage Check, Security Scan, Build.

- [ ] **[Automation/Security]**: Implement Automated Vulnerability Scanning in CI (Original T60)
  - **Complexity**: Simple
  - **Rationale**: Essential security practice to prevent deploying code with known vulnerabilities. Mandatory requirement of Philosophy §9.2.7 and Go Appendix §12.
  - **Expected Outcome**: CI pipeline includes `govulncheck` (Go) and potentially `npm audit` (if applicable), failing the build on discovery of new Critical or High severity vulnerabilities.

### Testing & Reliability

- [ ] **[Fix/Test]**: Restore and Enforce Full E2E Testing in CI (Refined T52, Original T45 part)
  - **Complexity**: Medium
  - **Rationale**: Ensures the deployed binary works correctly end-to-end, preventing regressions in critical user flows. Mandatory for operational reliability and adherence to Philosophy §9.2 (CI Pipeline Mandatory Stages) and §10 (Testing Strategy). Addresses `SKIP_BINARY_EXECUTION=true` workaround.
  - **Expected Outcome**: True end-to-end validation of the compiled binary runs successfully and is a mandatory blocking step in CI, potentially using Dockerized environments.

- [ ] **[Test/Process]**: Restore and Enforce CI Coverage Threshold (Refined T54)
  - **Complexity**: Simple
  - **Rationale**: Maintains code quality and prevents coverage degradation, supporting the "Design for Testability" principle (Philosophy §3, §10.4, Go Appendix §10).
  - **Expected Outcome**: Code coverage (e.g., 75%+) is accurately calculated (excluding test utilities) and enforced by CI, failing builds below the threshold.

## Medium Priority

### Core Features & Value Delivery

- [ ] **[Feature]**: Standardize Output Handling & Add JSON Output Mode (Original T17, T19, T33 part)
  - **Complexity**: Medium
  - **Rationale**: Improves integration with other tools and scripting by using standard streams (stdout for primary results, stderr for logs/errors) and providing a machine-readable JSON output mode (`--output-format json`). Makes file output optional (`--output <path>`), defaulting to stdout. Supports "Automate Everything" (Philosophy §6).
  - **Expected Outcome**: Primary results to stdout, logs/errors to stderr. File output is configurable. `--output-format json` flag provides structured JSON results.

- [ ] **[Feature]**: Enable Custom System Prompts & Model Parameters (Original T20, T21)
  - **Complexity**: Medium
  - **Rationale**: Increases flexibility and allows users to tailor the model's persona/instructions and fine-tune generation parameters (temperature, top-p, max tokens, etc.) for specific tasks via CLI flags or a configuration file.
  - **Expected Outcome**: Users can easily provide custom system prompts and configure all relevant model parameters, which are validated and respected.

- [ ] **[Feature]**: Implement Token Counting & Cost Estimation (Original T23, T24, T22 part)
  - **Complexity**: Medium
  - **Rationale**: Provides essential visibility into model usage (approaching limits) and helps users avoid unexpected costs or truncated outputs. Provides transparency on API costs. Adheres to Observability principles (Philosophy §14).
  - **Expected Outcome**: Token usage accurately tracked using provider APIs. Warnings for approaching limits. Estimated request cost logged. Model metadata (limits, pricing) cached and used.

- [ ] **[Feature]**: Implement Built-in Plan → Critique → Refine Workflow (Original T27)
  - **Complexity**: Medium
  - **Rationale**: Provides immediate value by offering a common, powerful multi-step process without requiring users to define a custom workflow. Builds on the workflow engine concept.
  - **Expected Outcome**: A predefined critique/refine workflow is available via a simple flag (e.g., `--critique-refine`).
  - **Dependencies**: Flexible Workflow Engine.

### Technical Debt Reduction & Foundation

- [ ] **[Refactor]**: Consolidate Test Mock Implementations into `internal/testutil` (Original T70, T46)
  - **Complexity**: Medium
  - **Rationale**: Reduces significant code duplication across test files, improving test maintainability and reducing overall codebase size. Supports robust testing infrastructure.
  - **Expected Outcome**: Centralized mocks for common interfaces (e.g., `APIService`, `LLMClient`, `ConfigLoader`) in `internal/testutil`. Test files use these shared mocks.

- [ ] **[Refactor]**: Centralize API Key Resolution Logic (Original T44, T77)
  - **Complexity**: Simple
  - **Rationale**: Ensures consistent API key handling across all code paths and providers, preventing subtle bugs and adhering to the "Explicit is Better than Implicit" principle (Philosophy §5).
  - **Expected Outcome**: API key loading logic (checking specific env vars, falling back to generic ones) is unified and reliable in a shared utility.

- [ ] **[Refactor/Test]**: Audit Tests Against Strict Mocking Policy (Original T58)
  - **Complexity**: Medium
  - **Rationale**: Enforces Dev Philosophy §7.3 (No Mocking Internal Collaborators), leading to better-designed, more loosely coupled code and more reliable tests verifying behavior through public APIs.
  - **Expected Outcome**: All tests adhere to the mocking policy. Code and tests refactored using DI, consumer interfaces, or pure functions to eliminate internal mocks where identified.

- [ ] **[Automation/Process]**: Enforce Tooling (Formatting, Linting, `go mod tidy`) via Pre-commit Hooks and CI (Original T65, T66)
  - **Complexity**: Simple
  - **Rationale**: Automates code style, basic quality checks, and dependency management, providing fast feedback and ensuring consistency. Mandatory per Philosophy §9.1 and Go Appendix §2, §3, §12.
  - **Expected Outcome**: Pre-commit hooks for `goimports`/`gofmt`, `golangci-lint`, and `go mod tidy` are active. CI includes mandatory checks for formatting, linting, and tidy modules.

### Operational Excellence

- [ ] **[Enhancement]**: Use Distinct Exit Codes for Different Outcomes (Original T18)
  - **Complexity**: Simple
  - **Rationale**: Improves scriptability and integration with other tools by providing clear, machine-readable status signals for success, user errors, API errors, file system errors, etc.
  - **Expected Outcome**: `thinktank` exits with documented, distinct codes for various outcomes.

### Provider & Model Support

- [ ] **[Enhancement]**: Add Gemini Grounding Support (Original T16)
  - **Complexity**: Medium
  - **Rationale**: Expands supported LLM features, enabling more sophisticated use cases by allowing Gemini models to reference specific documents or data.
  - **Expected Outcome**: Users can leverage Gemini's grounding capabilities through `thinktank`.

- [ ] **[Enhancement]**: Add Grok API Support
  - **Complexity**: Medium
  - **Rationale**: Expands the range of supported LLM providers, offering users more choice.
  - **Expected Outcome**: Users can configure and use models from Grok via its native API.

## Low Priority

### Advanced Features & Value Delivery

- [ ] **[Feature]**: Add Context Preprocessing (e.g., Summarization) (Original T28)
  - **Complexity**: Complex
  - **Rationale**: Enables handling larger inputs than a model's context window allows and can potentially improve model focus by providing a concise summary, possibly triggered automatically.
  - **Expected Outcome**: Large contexts can be processed effectively, with optional summarization for models with smaller context windows.

- [ ] **[Research/Feature]**: Auto-Select Appropriate Models Based on Task/Context (Original T29)
  - **Complexity**: Complex
  - **Rationale**: Simplifies usage for new users and helps experienced users leverage the best model for a given task without manual selection.
  - **Expected Outcome**: The program can suggest or automatically choose suitable models based on task description and context size/type.

- [ ] **[Feature]**: Integrate AST Parsing for Supported Languages (Original T30)
  - **Complexity**: Complex
  - **Rationale**: Provides deeper code understanding to the LLM by supplying richer structural context beyond raw text, potentially leading to more accurate analyses or code generation.
  - **Expected Outcome**: Code context sent to LLMs can include structural information from ASTs when applicable.

- [ ] **[Feature]**: Add Git Integration for Context (Original T31)
  - **Complexity**: Medium
  - **Rationale**: Integrates with common developer workflows by allowing `git diff` output or `git blame`/commit history to be used as context, providing valuable historical context for code analysis.
  - **Expected Outcome**: Users can provide git-based context (e.g., `--context-git-diff <ref>`).

- [ ] **[Feature]**: Add Code Generation Mode & Patch File Output (Original T34, T35)
  - **Complexity**: Complex
  - **Rationale**: Provides direct value by translating plans into executable code snippets or files, and outputting changes as `.patch` files for easy application, supporting "Automate Everything" (Philosophy §6).
  - **Expected Outcome**: The program can attempt to generate code based on plans. Option to output suggested code changes as patch files.

- [ ] **[Feature]**: Add Instruction Enhancement ("Modify My Instructions") (Original T40)
  - **Complexity**: Medium
  - **Rationale**: Helps users automatically improve their prompts for better results by rewriting instructions according to best prompt engineering practices before sending the actual request.
  - **Expected Outcome**: A flag (e.g., `--tune-prompt`) enables automatic optimization of user instructions.

### Technical Debt Reduction & Foundation (Less Critical)

- [ ] **[Refactor]**: Consolidate Provider-Specific API Error Formatting & Parameter Application Logic (Original T75, T76)
  - **Complexity**: Simple
  - **Rationale**: Reduces code duplication and centralizes logic in provider implementations, improving maintainability.
  - **Expected Outcome**: Shared utilities for parsing provider API errors and applying model parameters based on `registry.ParameterDefinition`.

- [ ] **[Refactor]**: Simplify Parameter Validation Logic in `registry_api.go` (Original T79)
  - **Complexity**: Medium
  - **Rationale**: Improves the maintainability and extensibility of parameter validation logic by replacing large type switches with more robust approaches (e.g., reflection, map-based lookups).
  - **Expected Outcome**: Parameter validation code is easier to understand and update.

- [ ] **[Refactor]**: Remove Redundant Comments & Simplify Shell Scripts (Original T82, T83)
  - **Complexity**: Simple
  - **Rationale**: Improves code readability, reduces documentation drift (Philosophy §7, §12.3), and enhances maintainability of build/CI scripts.
  - **Expected Outcome**: Code comments are focused on *why*, not *how*. Shell scripts in `scripts/` are simpler and easier to understand.

## Future Considerations

### Platform Capabilities & Extensibility
- **Improve Arbitrary Model Handling** (Original T15): Further enhance robustness and parameter discovery for models not explicitly defined in the registry.
- **Add Context Metadata to LLM Prompts** (Original T39): Include file paths, git status, etc., with context sent to LLMs.
- **Support User-Defined Plugins/Adapters**: Design an interface for community contributions and custom extensions.

### Operational Excellence & Observability
- **Implement Metrics and Tracing**: Add instrumentation for collecting key application metrics (e.g., request duration, error rates) and distributed tracing spans (OpenTelemetry) per Philosophy §14.
- **Advanced Cost Tracking**: Integrate with provider billing APIs for real-time cost and quota management.

### Testing & Quality
- **Restore Remaining Disabled Integration Tests** (Original T45 part): Ensure all integration tests across all providers are updated and re-enabled if not covered by E2E restoration.
- **Standardize Test File Naming** (Original T49): Consider renaming test files to more standard Go patterns (e.g., `<provider>_secrets_test.go`).
- **Remove Unused/Obsolete Test Files** (Original T78): Clean up files that contain only comments or are no longer relevant.

### Documentation & Housekeeping
- **Consolidate/Remove Legacy API Service Implementation** (Original T42): Deprecate and remove old `apiService` once functionality is fully migrated.
- **Improve Model Name Lookup (Case Insensitivity)** (Original T43): Prevent user confusion due to model capitalization variations.
- **Document Token Handling Changes** (Original T55): Clarify history and impact of any refactoring in token management logic.
- **Review README.md Against Philosophy Checklist** (Original T69): Ensure primary documentation is complete, useful, and aligns with Philosophy §12.2.

### Advanced Innovation
- **Model-Based Plan Critique and Self-Tuning**: Research enabling the system to self-evaluate plans and adapt prompts/models.
- **Interactive Web UI / Dashboard**: Explore a graphical interface to broaden user base.
