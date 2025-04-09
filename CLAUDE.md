# thinktank Development Guidelines

## Commands
- Build: `npm run build`
- Dev: `npm run dev`
- Start: `npm start`
- Lint: `npm run lint` (never disable linting or type checking)
- Fix newline issues: `node scripts/add-missing-newlines.js` (fixes common EOL linting errors)
- Tests: `npm test` (write tests first - TDD approach)
- Single test: `npm test -- -t "test name"`
- Test specific files: `npm test -- path/to/file.test.ts`

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
- **Commits**: Use conventional commit labels (feat, fix, refactor); keep atomic

## Architecture
- **Atomic Design**: atoms→molecules→organisms→templates→runtime
- **Core Types**: Define in `atoms/types.ts` (ModelConfig, LLMResponse, etc.)
- **Providers**: Implement `LLMProvider` interface with `providerId` and `generate()`
- **Functional Style**: Favor pure functions with minimal side effects
- **Immutability**: Prefer immutable data to simplify debugging
- **Abstractions**: Avoid premature abstraction; abstract only after patterns emerge
- **Simplicity**: Prioritize readability and maintainability (KISS)
- **Dependencies**: Limit external dependencies to those adding clear value
- **Performance**: Write clear, efficient code; optimize only after profiling
