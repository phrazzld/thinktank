# CONTRIBUTING

**Objective:** Contribute maintainable, reliable, and testable code.

## Core Principles

1.  **Testability First:** Design code for testability from the start. Refer to `TESTING_PHILOSOPHY.MD`. If code is hard to test simply, refactor the code *before* writing complex tests.
2.  **Modularity & Decoupling:** Write clean, modular code. Prefer dependency injection, especially for external concerns (Filesystem, APIs), to facilitate testing. Follow the established domain-oriented architecture.
3.  **Type Safety:** Use TypeScript's strictest settings. Avoid `any`. Define clear types/interfaces. Do not disable type checking.
4.  **Code Quality:** Use formatters (`pnpm run format`) and linters (`pnpm run lint`). Do not disable linting. Use clear, self-descriptive names. Prioritize readability. Ensure files end with a newline (`pnpm run fix:newlines`). Follow `.editorconfig`.
5.  **Simplicity (KISS):** Avoid premature or overly complex abstractions. Refactor complexity when identified.
6.  **Commit Hygiene:** Use Conventional Commits (`type(scope): subject`). Keep commits atomic and logically grouped. Explain the 'why' in messages.

## Testing Mandates

* **Philosophy:** Strictly adhere to `TESTING_PHILOSOPHY.MD`.
* **TDD:** Write failing tests first for new features/bug fixes to define expected behavior.
* **Behavior Focus:** Test public interfaces and behavior, not internal implementation details.
* **Simplicity:** Tests must be simple, readable, and maintainable.
* **Minimize Mocking:** **AVOID COMPLEX MOCKS.** Mock *only* at external system boundaries as defined in `TESTING_PHILOSOPHY.MD`. Use the project's virtual filesystem utilities (`virtualFsUtils.ts`, `fsTestSetup.ts`) for filesystem tests; do not mock the `fs` module directly.
* **Execution:** Run tests via `pnpm test`. Ensure all tests pass before committing.

## Workflow

1.  **Branch:** Create a feature branch (`git checkout -b type/short-description`).
2.  **Code & Test:** Implement changes following TDD and all principles above.
3.  **Lint & Format:** Run `pnpm run lint` and `pnpm run format`. Fix all issues.
4.  **Commit:** Commit changes using Conventional Commits format.
5.  **Document:** Update relevant documentation (`README.md`, `docs/`, JSDoc) if APIs or user-facing functionality changes.
6.  **Pull Request:** Push branch and open a Pull Request. Ensure all checks pass.

## Dependencies

* Limit external dependencies. Justify additions.
* Keep dependencies updated (`pnpm update`).

## Performance

* Write efficient code. Profile before optimizing complex sections.
