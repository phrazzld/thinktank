# Conventional Commits Guide

## Overview

This project uses the [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages. This standardized format enables automated version determination, changelog generation, and improves readability and navigation of git history.

## Baseline Policy: Important Note

**Important:** Commit message validation only applies to commits made **after** our baseline commit:

```
Baseline Commit: 1300e4d675ac087783199f1e608409e6853e589f (May 18, 2025)
```

This approach allows us to:
1. Preserve git history for commits made before the standard was adopted
2. Enforce the standard for all new development
3. Maintain a clean transition between historical code and new contributions

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

### Breaking Changes

For breaking changes, add a `!` after the type/scope or include `BREAKING CHANGE:` in the footer:

```
feat!: add new required parameter to API endpoint

BREAKING CHANGE: API clients will need to provide the new parameter
```

## Examples

```
feat: add user authentication
fix(parser): handle edge case in JSON parsing
docs: update installation instructions
refactor(api): simplify request handling logic
test: add integration tests for order processing
chore: update dependency versions
```

## Validation Tools

The following tools enforce the conventional commit standard:

1. **Pre-Commit Hook**: Validates commit messages during local development
2. **Pre-Push Hook**: Validates commits before they are pushed to the remote repository
3. **CI Pipeline**: Validates all commits in pull requests against the standard

All three validation systems are configured to only check commits made after the baseline commit.

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

## Why Use Conventional Commits?

1. **Automated Versioning**: The commit type determines semantic version bumps
   - `fix` = PATCH version
   - `feat` = MINOR version  
   - `BREAKING CHANGE` = MAJOR version
2. **Automated Changelog**: Commit messages are grouped by type in the generated changelog
3. **Intent Clarity**: The type prefix immediately conveys the purpose of each change
4. **Improved Collaboration**: Consistent format makes code review and history navigation easier

## Additional Resources

- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Our CI/CD Pipeline Documentation](ci-troubleshooting.md)
