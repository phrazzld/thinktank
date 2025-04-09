# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# thinktank Development Guidelines

## Commands
- Build: `pnpm run build`
- Dev: `pnpm run dev`
- Start: `pnpm start`
- Lint: `pnpm run lint` (never disable linting or type checking)
- Fix lint: `pnpm run lint:fix`
- Format: `pnpm run format`
- Fix newline issues: `pnpm run fix:newlines` (fixes common EOL linting errors)
- Tests: `pnpm test` (write tests first - TDD approach)
- Single test: `pnpm test -- -t "test name"`
- Test specific files: `pnpm test -- path/to/file.test.ts`
- Debug tests: `pnpm run test:debug`
- Debug specific test file: `pnpm run test:debug:file path/to/file.test.ts`

## Using thinktank
When facing challenging problems or planning complex features, get help from thinktank. Write your task or request to TASK.md. Then run:
```
thinktank run --group faves TASK.md ./
```
- Identify the output directory. Read all the files in there. Synthesize their outputs into an answer or plan.
- If you encounter any errors running thinktank, append those errors to thinktank-errors.log.
- Note: `GEMINI_API_KEY` environment variable is already set.

## Code Style
- **TypeScript**: Use strictest settings; avoid `any` or vague types
- **Imports**: Group by (1)core, (2)libraries, (3)local with blank lines between
- **Formatting**: Use automated formatters for consistent style
- **Naming**: Choose meaningful, self-descriptive names for clarity
- **Models**: Use `provider:modelId` convention (e.g., `openai:gpt-4o`)
- **Error Handling**: Try/catch for async; populate `error` field in `LLMResponse`
- **Testing**: Focus on behavior over implementation; minimize mocking
- **Commits**: Use conventional commit labels (feat, fix, refactor); keep atomic

## Architecture
- **Functional Style**: Favor pure functions with minimal side effects
- **Immutability**: Prefer immutable data to simplify debugging
- **Abstractions**: Avoid premature abstraction; abstract only after patterns emerge
- **Simplicity**: Prioritize readability and maintainability (KISS)
- **Providers**: Implement `LLMProvider` interface with `providerId` and `generate()`
- **Dependencies**: Limit external dependencies to those adding clear value
- **Performance**: Write clear, efficient code; optimize only after profiling
