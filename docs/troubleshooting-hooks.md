# Troubleshooting Pre-commit Hooks

This guide helps resolve common issues with pre-commit hook installation and operation.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Hook Execution Issues](#hook-execution-issues)
- [Platform-Specific Issues](#platform-specific-issues)
- [Common Error Messages](#common-error-messages)
- [Advanced Troubleshooting](#advanced-troubleshooting)

## Installation Issues

### pre-commit command not found

**Symptom:**
```bash
$ pre-commit --version
bash: pre-commit: command not found
```

**Solutions:**

1. **Install pre-commit using pip (recommended):**
   ```bash
   pip install pre-commit
   # or if you have pip3
   pip3 install pre-commit
   ```

2. **Install using pipx (for isolated environments):**
   ```bash
   pipx install pre-commit
   ```

3. **Install using Homebrew (macOS/Linux):**
   ```bash
   brew install pre-commit
   ```

4. **Install using conda:**
   ```bash
   conda install -c conda-forge pre-commit
   ```

### pre-commit version too old

**Symptom:**
```
ERROR: pre-commit version 2.x.x is older than required 3.0.0
```

**Solution:**
```bash
# Upgrade pre-commit
pip install --upgrade pre-commit
# or
pipx upgrade pre-commit
# or
brew upgrade pre-commit
```

### Permission denied during installation

**Symptom:**
```
ERROR: Could not install packages due to an EnvironmentError: [Errno 13] Permission denied
```

**Solutions:**

1. **Use user installation (recommended):**
   ```bash
   pip install --user pre-commit
   ```

2. **Use a virtual environment:**
   ```bash
   python -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   pip install pre-commit
   ```

## Hook Execution Issues

### Hooks not running on commit

**Symptom:**
Commits are created without running validation checks.

**Diagnosis:**
```bash
# Check hook status
make hooks-status

# Look for missing hooks
ls -la .git/hooks/
```

**Solutions:**

1. **Reinstall hooks:**
   ```bash
   make hooks-clean
   make hooks
   ```

2. **Check for custom hooks path:**
   ```bash
   # See if git is looking elsewhere for hooks
   git config core.hooksPath

   # Remove custom path if set
   git config --unset-all core.hooksPath
   ```

### Hook fails with "command not found"

**Symptom:**
```
[ERROR] go fmt...........................................................Failed
- hook id: go-fmt
- exit code: 127
```

**Solution:**
Ensure required tools are in your PATH:
```bash
# Add Go binaries to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Make permanent by adding to ~/.bashrc or ~/.zshrc
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
```

### Conventional commit validation fails

**Symptom:**
```
ERROR: Commit message does not follow conventional commits format
```

**Solution:**
Review the conventional commits format:
- Format: `<type>[optional scope]: <description>`
- Valid types: feat, fix, docs, style, refactor, test, chore
- Example: `feat(api): add user authentication endpoint`

See `docs/conventional-commits.md` for detailed guidelines.

## Platform-Specific Issues

### Windows (WSL)

**Line ending issues:**
```bash
# Configure git to handle line endings
git config --global core.autocrlf true
```

**Path issues:**
Ensure WSL can access Windows Python/pip:
```bash
# In WSL, use Windows Python if needed
/mnt/c/Python39/python.exe -m pip install pre-commit
```

### macOS

**brew vs pip conflicts:**
If you have both brew and pip versions:
```bash
# Check which is being used
which pre-commit

# Uninstall one version
brew uninstall pre-commit  # or pip uninstall pre-commit
```

### Linux

**Python version issues:**
Some distributions have old Python versions:
```bash
# Use python3 explicitly
python3 -m pip install pre-commit

# Or install newer Python
sudo apt update
sudo apt install python3.9 python3.9-pip
```

## Common Error Messages

### "No .git directory found"

**Cause:** Running setup outside a git repository.

**Solution:**
```bash
# Initialize git if needed
git init

# Or ensure you're in the project root
cd /path/to/thinktank
```

### "yaml.scanner.ScannerError"

**Cause:** Invalid YAML in .pre-commit-config.yaml

**Solution:**
```bash
# Validate YAML syntax
python3 -c "import yaml; yaml.safe_load(open('.pre-commit-config.yaml'))"

# Common issues:
# - Incorrect indentation (use spaces, not tabs)
# - Missing quotes around special characters
# - Invalid repo URLs
```

### "Repository not found"

**Cause:** Network issues or invalid repository URL in config.

**Solution:**
```bash
# Check network connectivity
ping github.com

# Verify repo URLs in .pre-commit-config.yaml
# Ensure URLs are accessible
```

## Advanced Troubleshooting

### Debug hook execution

```bash
# Run hooks with verbose output
pre-commit run --verbose --all-files

# Run specific hook
pre-commit run go-fmt --all-files
```

### Reset everything

```bash
# Complete reset
make hooks-clean
rm -rf ~/.cache/pre-commit
make hooks
```

### Check pre-commit cache

```bash
# View cached hooks
ls -la ~/.cache/pre-commit/

# Clear cache if corrupted
rm -rf ~/.cache/pre-commit/
```

### Manual hook inspection

```bash
# View hook content
cat .git/hooks/pre-commit

# Check if hook is executable
ls -l .git/hooks/pre-commit
```

## Getting Help

If you're still experiencing issues:

1. Run diagnostic command:
   ```bash
   make hooks-status
   pre-commit --version
   git --version
   go version
   ```

2. Check existing issues: [GitHub Issues](https://github.com/phrazzld/thinktank/issues)

3. Create a new issue with:
   - Output from diagnostic commands
   - Error messages
   - Steps to reproduce
   - Operating system and version

## Quick Reference

| Command | Purpose |
|---------|---------|
| `./scripts/setup.sh` | Complete setup including pre-commit |
| `make hooks` | Install all hooks |
| `make hooks-status` | Check current hook status |
| `make hooks-clean` | Remove all hooks |
| `make hooks-validate` | Validate configuration |
| `pre-commit run --all-files` | Test all hooks |
