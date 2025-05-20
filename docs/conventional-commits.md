# Conventional Commits Guide

## Table of Contents

- [Quick Reference](#-quick-reference)
- [Overview](#overview)
- [Commit Message Format](#commit-message-format)
  - [Types](#types)
  - [Common Scopes](#common-scopes)
  - [Breaking Changes](#breaking-changes)
- [Examples](#examples)
  - [Valid Commit Messages](#valid-commit-messages)
  - [Invalid Commit Messages](#invalid-commit-messages-dont-do-these)
- [Validation and Assistance Tools](#validation-and-assistance-tools)
  - [Using Commitizen](#using-commitizen)
- [Fixing Invalid Commit Messages](#fixing-invalid-commit-messages)
- [Relationship to Semantic Versioning](#relationship-to-semantic-versioning)
- [Benefits of Conventional Commits](#benefits-of-conventional-commits)
- [Best Practices](#best-practices)
- [Additional Resources](#additional-resources)
- [Baseline Validation Policy](#baseline-validation-policy)

## 🔍 Quick Reference

| Type | Description | SemVer Impact | Examples |
| ---- | ----------- | ------------- | -------- |
| **feat** | New feature | MINOR (1.x.0) | `feat: add user authentication` |
| **fix**  | Bug fix | PATCH (1.0.x) | `fix(parser): handle edge case in JSON parsing` |
| **docs** | Documentation | None | `docs: update README with examples` |
| **style** | Code formatting | None | `style: fix indentation in config.go` |
| **refactor** | Code restructuring | None | `refactor: simplify error handling in registry` |
| **perf** | Performance improvement | None | `perf: optimize file scanning algorithm` |
| **test** | Adding/fixing tests | None | `test: add unit tests for config package` |
| **chore** | Maintenance tasks | None | `chore: update dependencies` |
| **ci** | CI/CD changes | None | `ci: add new GitHub action workflow` |
| **build** | Build system changes | None | `build: update Makefile targets` |
| **Breaking Change** | Any type with `!` or BREAKING CHANGE in footer | MAJOR (x.0.0) | `feat!: change API response format` |

> **⚠️ Important:** This project uses a baseline validation policy. Commit messages are only validated **after** our baseline commit:
>
> **Baseline Commit:** `1300e4d675ac087783199f1e608409e6853e589f` (May 18, 2025)
>
> This allows us to preserve git history while enforcing standards for all new development.

## Overview

This project uses the [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages. This standardized format enables automated version determination, changelog generation, and improves readability and navigation of git history.

## Commit Message Format

All commits made after the baseline commit **MUST** follow this format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code (formatting, etc)
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools
- `ci`: Changes to CI configuration files and scripts
- `build`: Changes that affect the build system or external dependencies

### Common Scopes

Scopes are optional but highly recommended as they provide context about which part of the codebase is being modified. Here are common scopes used in this project:

| Scope | Description | Examples |
|-------|-------------|----------|
| `api` | API-related changes | `feat(api): add new endpoint for model listing` |
| `registry` | Model registry functionality | `fix(registry): correct model detection logic` |
| `providers/openai` | OpenAI provider | `fix(providers/openai): handle rate limiting` |
| `providers/gemini` | Gemini provider | `feat(providers/gemini): add support for gemini-1.5 model` |
| `providers/openrouter` | OpenRouter provider | `feat(providers/openrouter): add new provider support` |
| `cli` | Command-line interface | `feat(cli): add --dry-run flag` |
| `config` | Configuration handling | `fix(config): resolve path resolution issue` |
| `fileutil` | File utilities | `perf(fileutil): optimize file scanning` |
| `orchestrator` | Orchestration logic | `refactor(orchestrator): simplify error handling` |
| `e2e` | End-to-end tests | `test(e2e): add test for multi-model synthesis` |
| `docs` | Documentation | `docs(readme): update installation instructions` |

You can use other scopes as needed, especially for specific packages or components. The scope should be lowercase and use kebab-case for multi-word scopes (e.g., `user-auth`).

### Breaking Changes

For breaking changes, add a `!` after the type/scope or include `BREAKING CHANGE:` in the footer:

```
feat!: add new required parameter to API endpoint

BREAKING CHANGE: API clients will need to provide the new parameter
```

## Examples

### Valid Commit Messages

```bash
# Feature additions
feat: add support for OpenRouter models
feat(registry): implement model version detection
feat(api): add streaming response capability

# Bug fixes
fix: resolve panic when config file is missing
fix(providers/openai): handle rate limit errors correctly
fix(registry): fix model validation for custom endpoints

# Documentation updates
docs: add quick start guide to README
docs(contributing): clarify PR process
docs: update file descriptions in glance.md

# Code improvements
refactor: simplify error propagation in orchestrator
style: standardize import order in all files
perf: optimize file scanning for large repositories

# Testing and CI
test: add integration tests for OpenAI provider
test(e2e): cover baseline commit validation scenarios  
ci: update golangci-lint version in GitHub Actions

# Build and dependency updates
build: update Go module dependencies
chore: bump Go version to 1.22
```

### Invalid Commit Messages (Don't Do These)

```bash
# Missing type prefix
add support for OpenRouter models           # ❌ Missing type (should be "feat: add support...")

# Incorrect format
fix - handle rate limit errors              # ❌ Missing colon after type (should use "fix: handle...")
FIX: resolve panic                          # ❌ Type must be lowercase (should be "fix: resolve...")

# Too generic descriptions
chore: fix stuff                            # ❌ Too vague, not descriptive
feat: updates                               # ❌ Not specific enough

# Wrong type usage
feat: fix bug in parser                     # ❌ Wrong type for a bug fix (should be "fix:")
docs: refactor error handling               # ❌ Wrong type for code change (should be "refactor:")

# Capitalization issues
feat: Add streaming capability              # ❌ First letter capitalized (use lowercase)
```

## Validation and Assistance Tools

The following tools help with conventional commit standards:

1. **Pre-Commit Hook**: Validates commit messages during local development
2. **Pre-Push Hook**: Validates commits before they are pushed to the remote repository
3. **CI Pipeline**: Validates all commits in pull requests against the standard
4. **Commitizen**: Interactive tool that guides you through creating conventional commits (optional but recommended)
5. **Git Commit Template**: Pre-filled template to help structure your commit messages

All validation systems are configured to only check commits made after the baseline commit.

### Using Commitizen

[Commitizen](https://github.com/commitizen/cz-cli) is an interactive command-line tool that guides you through creating properly formatted conventional commit messages. It's particularly helpful if you're new to the conventional commits format or want to ensure you don't miss important components.

To use Commitizen:

```bash
# One-time setup: Install dependencies
npm install

# Option 1: Use the script
./scripts/commit.sh

# Option 2: Use Make
make commit

# Option 3: Use npm directly
npm run commit
```

The tool will prompt you to:
- Select a commit type (feat, fix, docs, etc.)
- Enter a scope (optional)
- Write a short description
- Provide a longer description (optional)
- Indicate if there are breaking changes
- Reference issues (optional)

Commitizen ensures your commits follow the conventional format without having to remember all the details.

## Fixing Invalid Commit Messages

If validation fails, you'll need to fix the commit message:

### For the most recent commit

```bash
git commit --amend
```

### For older commits

```bash
git rebase -i <commit-hash>^
```

Then change `pick` to `reword` for the commits you want to modify and save.

### For commits in a PR

1. Create a new branch from before your changes
2. Cherry-pick your commits with corrected messages, or
3. Use interactive rebase to fix the commit messages
4. Force-push to update your PR

## Relationship to Semantic Versioning

This project uses [Semantic Versioning](https://semver.org/) (SemVer) for release versioning. The commit type directly determines how the version number is incremented:

| Commit Type | Version Impact | Example | Description |
|-------------|----------------|---------|-------------|
| `fix` | **PATCH** | 1.0.0 → 1.0.1 | Bug fixes that don't change the API |
| `feat` | **MINOR** | 1.0.0 → 1.1.0 | New features that don't break existing functionality |
| **Breaking Change** | **MAJOR** | 1.0.0 → 2.0.0 | Changes that break backward compatibility |

Breaking changes can be indicated in two ways:
1. Adding `!` after the type: `feat!: remove deprecated endpoint`
2. Including `BREAKING CHANGE:` in the commit footer

Other commit types (`docs`, `style`, `refactor`, etc.) don't trigger version changes unless they include a breaking change marker.

### Automated Process

When a commit is pushed to the repository:
1. The CI pipeline uses `svu` to analyze commit history
2. It calculates the next version number based on commit types
3. For releases, it automatically generates a changelog grouped by commit type
4. Release artifacts are tagged with the appropriate semantic version

## Benefits of Conventional Commits

1. **Automated Versioning**: No manual version decisions needed
2. **Clear Release Notes**: Automated changelog generation with categorized changes
3. **Intent Clarity**: The type prefix immediately conveys the purpose of each change
4. **API Contract**: Users understand the impact of upgrades based on version changes
5. **Improved Collaboration**: Consistent format makes code review and history navigation easier
6. **Better Git History**: More meaningful commit messages and easier filtering

## Best Practices

Follow these guidelines to write effective conventional commit messages:

1. **Be Specific**: Describe what was changed and why, not how
   - ✅ `fix(providers): handle API timeout errors`
   - ❌ `fix: updated code`

2. **Use Imperative Mood**: Write as if giving a command
   - ✅ `feat: add user authentication`
   - ❌ `feat: added user authentication`

3. **Keep Subject Line Short**: Aim for 50-72 characters maximum
   - Most tools truncate subject lines longer than 72 characters

4. **Use Body for Context**: Add detailed explanations in the commit body
   ```
   feat(registry): add model version detection

   Implement automatic detection of model versions from provider APIs.
   This allows users to specify just the model family and have the
   system automatically select the latest compatible version.
   ```

5. **Reference Issues**: Include issue references in the footer
   ```
   fix(cli): resolve path resolution error

   Fixes #123
   ```

6. **One Change Per Commit**: Each commit should represent a single logical change
   - Makes reviews easier
   - Enables clean reverts if needed
   - Produces cleaner history and changelogs

7. **Use Consistent Scopes**: Choose from the common scopes when possible

8. **Document Breaking Changes**: Always clearly mark and explain breaking changes
   ```
   feat(api)!: change response format to JSON

   BREAKING CHANGE: API now returns JSON instead of XML.
   Clients will need to update their parsers accordingly.
   ```

## Additional Resources

- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Our CI/CD Pipeline Documentation](ci-troubleshooting.md)
- [Git Commit Message Guide](https://chris.beams.io/posts/git-commit/)
- [Commitizen CLI Tool](https://github.com/commitizen/cz-cli)

## Baseline Validation Policy

This project implements a baseline validation policy for commit messages. This approach allows us to:

1. **Preserve Git History**: We keep our git history intact, including commits made before adopting the conventional commit standard
2. **Enforce Standards Going Forward**: All new development after the baseline date must follow the standard
3. **Avoid Unnecessary Rebasing**: We don't need to rewrite history which could cause issues for contributors

### How It Works

- **Baseline Commit**: `1300e4d675ac087783199f1e608409e6853e589f` (May 18, 2025)
- **Implementation**: Our CI pipeline and pre-commit hooks are configured to only validate commits made after this baseline commit
- **Script**: We use a custom script (`scripts/ci/validate-baseline-commits.sh`) that filters out commits before the baseline

### For Contributors

- If you're working on code after May 18, 2025, all your commits need to follow the conventional commit format
- Historical commits (before the baseline) won't be validated or trigger CI failures
- You can run the baseline validation script locally to check your commits:
  ```bash
  ./scripts/ci/validate-baseline-commits.sh
  ```

### CI Integration

Our CI pipeline automatically handles this using the baseline validation script. When a pull request is created, only commits made after the baseline date are validated against the conventional commit standard.

This approach ensures we maintain a high quality of commit messages going forward while respecting the project's history.
