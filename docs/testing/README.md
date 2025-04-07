# Thinktank Testing Guide

This directory contains comprehensive documentation on testing practices for the Thinktank project.

## Overview

Thinktank follows test-driven development (TDD) principles:

1. **Write tests first**: Before implementing a feature or fixing a bug, write tests that define the expected behavior.
2. **Focus on behavior, not implementation**: Test what the code does, not how it does it.
3. **Test edge cases and error conditions**: Don't just test the happy path; test error handling and edge cases.
4. **Isolation**: Tests should be independent of each other and the external environment.

## Test Suite Organization

Tests are organized parallel to the source code structure:

```
src/
├── core/
│   ├── __tests__/       # Tests for core functionality
│   ├── errors/
│   │   ├── __tests__/   # Tests for error handling
├── utils/
│   ├── __tests__/       # Tests for utility functions
├── __tests__/           # Shared test utilities
    ├── utils/
        ├── virtualFsUtils.ts    # Virtual filesystem utilities
```

## Running Tests

The following npm scripts are available for running tests:

```bash
# Run all tests
npm test

# Run tests with coverage report
npm run test:cov

# Run tests with specific debug options
npm run test:debug

# Run a specific test file or test name
npm test -- -t "test name"
npm test -- path/to/file.test.ts
```

## Testing Guidelines

For specific testing scenarios, refer to these guides:

- [Filesystem Testing](./filesystem-testing.md) - How to test code that interacts with the filesystem
- [Error Handling](./error-handling.md) - Patterns for testing error conditions
- [E2E Testing](./e2e-testing.md) - End-to-end testing approaches

## Migration Status

The project is undergoing a test infrastructure migration:

- ✅ Phase 1: Migration to `memfs` for most filesystem tests
- 🔄 Phase 2: Simplifying gitignore testing
- 🔄 Phase 3-10: Additional improvements (see [Master Plan](../planning/master-plan.md))

All new tests should use the current recommended patterns in these guides rather than legacy approaches.