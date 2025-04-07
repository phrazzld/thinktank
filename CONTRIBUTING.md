# Contributing to thinktank

Thank you for considering contributing to thinktank! This document provides guidelines and best practices to help you contribute effectively.

## Getting Started

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Development Philosophy

Our development philosophy centers on maintainability, reliability, and scalability. Following these principles will help ensure high-quality contributions.

### Test-Driven Development (TDD)

- Write failing tests first to clearly define expected behavior
- Keep tests small, focused, and fast
- Aim for meaningful, high-quality test coverage
- Run tests using `npm test` or `npm test -- -t "test name"` for a specific test

### Type Safety

- Use TypeScript's strictest settings; avoid `any` or vague types
- Define clear, expressive type aliases and interfaces
- Never disable type checking; resolve root causes rather than suppressing warnings

### Code Style and Quality

- Use automated formatters for consistent style
- Never disable linting; run `npm run lint` to ensure code quality
- Choose meaningful, self-descriptive names for clarity
- Prioritize readability and straightforwardness over cleverness
- Ensure all files end with a newline character (use `npm run fix:newlines` to check and fix)
- Follow EditorConfig settings for consistent file formatting

### Architectural Guidelines

- Write clean, modular, functional code
- Favor pure functions with minimal side effects
- Prefer immutability to simplify state management and debugging
- Organize code into small, composable modules
- Follow the domain-oriented architecture established in the project

### Abstraction Principles

- Avoid premature abstraction; abstract only after clear patterns emerge
- Frequently reassess abstractions—complexity signals a flawed design
- If a reusable component accumulates too many special-case parameters:
  - Replace its usages with simpler, copied versions
  - Later, abstract only genuinely shared functionality

### Commit Guidelines

- Follow the Conventional Commits specification for commit messages
- Keep commits small, self-contained, and logically grouped
- Provide concise commit messages explaining why changes were made
- Format: `type(scope): subject` (e.g., `feat(cli): add new option for model selection`)

### Documentation

- Update documentation when adding or changing functionality
- Strive for self-documenting code through clear naming and structure
- Comment only to clarify rationale or context, not functionality
- Add JSDoc comments for public APIs and complex functions

### Dependencies

- Limit external dependencies to those adding clear value
- Regularly upgrade dependencies to maintain security and stability

### Performance

- Avoid premature optimization; profile first to identify bottlenecks
- Write clear, efficient code by default, optimizing only as needed

## Pull Request Process

1. Ensure your code passes all tests and lint checks
2. Update the documentation with details of changes to the interface
3. You may merge the Pull Request once it has been approved by a maintainer

## Code of Conduct

Please be respectful and inclusive in all interactions related to this project. We are committed to providing a welcoming and inspiring community for all.

## Questions?

If you have any questions or need clarification, please open an issue or reach out to the maintainers.
