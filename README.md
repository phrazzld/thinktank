# thinktank

A context-aware LLM tool that analyzes codebases and generates responses to your instructions using AI models through OpenRouter's unified API.

## Quick Start

```bash
# Installation
git clone https://github.com/phrazzld/thinktank.git
cd thinktank
go install

# Set API key (all models now use OpenRouter)
export OPENROUTER_API_KEY="your-key" # Get your key at https://openrouter.ai/keys

# Create instructions file
echo "Analyze this codebase and suggest improvements" > instructions.txt

# Basic usage - Simplified Interface (Recommended)
thinktank instructions.txt ./my-project

# Preview without API calls
thinktank instructions.txt ./my-project --dry-run

# Verbose output
thinktank instructions.txt ./my-project --verbose

# Force synthesis mode (multiple models + synthesis)
thinktank instructions.txt ./my-project --synthesis
```

## Key Features

- **Multiple AI Models**: Access to various AI models through OpenRouter's unified API
- **Smart Filtering**: Include/exclude specific files or directories
- **Concurrent Processing**: Compare responses from multiple models in parallel
- **Result Synthesis**: Combine outputs from multiple models using a synthesis model
- **Git-Aware**: Respects .gitignore patterns
- **Structured Output**: Formats responses based on your specific instructions

## Interface

thinktank uses a simplified interface designed for clarity and ease of use:

```bash
# Pattern: thinktank instructions.txt target_path... [options]
thinktank task.md ./src                    # Basic analysis with smart defaults
thinktank task.md ./src --dry-run          # Preview files and token count

# NEW: Multiple target paths
thinktank task.md file1.go file2.go       # Analyze specific files
thinktank task.md src/ tests/ docs/       # Analyze multiple directories
thinktank task.md main.go src/ tests/     # Mix files and directories
thinktank task.md ./src --verbose          # With detailed output
thinktank task.md ./src --synthesis        # Force multi-model analysis
```

## Available Options

The simplified interface supports these flags:

| Flag | Description | Example |
|------|-------------|---------|
| `--dry-run` | Preview files and token count without API calls | `thinktank task.txt ./src --dry-run` |
| `--verbose` | Enable detailed output and logging | `thinktank task.txt ./src --verbose` |
| `--synthesis` | Force multi-model analysis with synthesis | `thinktank task.txt ./src --synthesis` |
| `--debug` | Enable debug-level logging | `thinktank task.txt ./src --debug` |
| `--quiet` | Suppress console output (errors only) | `thinktank task.txt ./src --quiet` |
| `--json-logs` | Show JSON logs on stderr | `thinktank task.txt ./src --json-logs` |
| `--no-progress` | Disable progress indicators | `thinktank task.txt ./src --no-progress` |

## Configuration

### Required
- **Instructions File**: Text file with your analysis request (first argument in simplified interface)
- **API Key**: Single environment variable `OPENROUTER_API_KEY` for all models

### Model Selection

thinktank uses intelligent model selection based on your input size and available API keys:

- **Small inputs**: Single fast model (e.g., `gemini-2.5-flash`)
- **Large inputs**: Multiple high-capacity models with automatic synthesis
- **With `--synthesis` flag**: Always uses multiple models with synthesis

### Output Directory

Output files are automatically saved to timestamped directories:
```
thinktank_YYYYMMDD_HHMMSS_NNNN
```

Example: `thinktank_20250627_143022_7841`

## Rate Limiting & Performance Optimization

thinktank provides intelligent rate limiting with provider-specific optimizations to help you get the best performance while staying within API limits.

### Rate Limiting Hierarchy

Rate limits are applied in this priority order:
1. **Model-specific overrides** (for special models like DeepSeek R1)
2. **OpenRouter rate limiting** (`--openrouter-rate-limit`)
3. **Global rate limit** (`--rate-limit`)
4. **Default settings** (20 RPM for conservative mixed model usage)

### OpenRouter Rate Limiting

#### Default Settings
- **Default**: 20 RPM (conservative for mixed model usage)
- **Free models**: Automatic limit of 20 RPM (cannot be increased)
- **With $10+ balance**: Most models support higher rates; try `--openrouter-rate-limit 100`
- **High-volume**: With sufficient balance, use `--openrouter-rate-limit 500`

### Special Model Considerations

Some models have specific rate limits regardless of provider settings:
- `openrouter/deepseek/deepseek-r1-0528`: Limited to 5 RPM (reasoning model)
- `openrouter/deepseek/deepseek-r1-0528:free`: Limited to 3 RPM (free tier)

### Common Rate Limiting Scenarios

#### Basic Usage Examples
```bash
# Default intelligent selection (automatic rate limiting)
thinktank task.txt ./src

# Verbose output for troubleshooting
thinktank task.txt ./src --verbose

# Force synthesis mode for complex analysis
thinktank task.txt ./src --synthesis

# Preview what would be processed
thinktank task.txt ./src --dry-run
```

### Troubleshooting Rate Limits

#### Symptoms of Rate Limiting Issues
- Models showing "rate limited" in output
- Long delays between model processing
- 429 errors in verbose logs

#### Solutions by Provider

**OpenRouter Rate Limiting:**
- Ensure account balance > $10 for higher daily limits
- Use `--openrouter-rate-limit 10` for conservative usage
- Check model-specific limits at [OpenRouter Models](https://openrouter.ai/models)
- Free tier models are automatically limited to 20 RPM

#### General Optimization Tips
- Start conservative and increase gradually
- Use `--dry-run` to estimate request volume before processing
- Monitor actual usage patterns with `--verbose` logging
- Consider time-based processing for large codebases

## Supported Models

thinktank supports the following LLM models out of the box:

All models are now accessed through OpenRouter for unified API management:

**OpenAI Models (via OpenRouter):**
  - gpt-4.1 (openai/gpt-4.1)
  - o4-mini (openai/o4-mini)
  - o3 (openai/o3)

**Google Models (via OpenRouter):**
  - gemini-2.5-flash (google/gemini-2.5-flash)
  - gemini-2.5-pro (google/gemini-2.5-pro)

**Native OpenRouter Models:**
  - openrouter/deepseek/deepseek-chat-v3-0324
  - openrouter/deepseek/deepseek-r1
  - openrouter/x-ai/grok-3-beta

No additional configuration is needed - simply set the appropriate API key environment variable and use any supported model name with the `--model` flag.

### Adding New Models

To add a new model, edit `internal/models/models.go` directly:

1. Add a new entry to the `ModelDefinitions` map with the model name as key
2. Provide the required `ModelInfo` struct with:
   - `Provider`: Always "openrouter" (unified provider)
   - `APIModelID`: The OpenRouter model ID (e.g., "openai/gpt-5", "google/new-model")
   - `ContextWindow`: Maximum input + output tokens
   - `MaxOutputTokens`: Maximum output tokens
   - `DefaultParams`: Model-specific parameters (temperature, top_p, etc.)
3. Run tests: `go test ./internal/models`
4. Submit a pull request with your changes

Example:
```go
"new-model-name": {
    Provider:        "openrouter",
    APIModelID:      "openai/gpt-5",
    ContextWindow:   200000,
    MaxOutputTokens: 50000,
    DefaultParams: map[string]interface{}{
        "temperature": 0.7,
        "top_p":       1.0,
    },
},
```

## Common Use Cases

```bash
# Technical planning
thinktank feature-plan.txt ./src

# Code review (uses intelligent model selection)
thinktank code-review.txt ./pull-request

# Architecture analysis
thinktank arch-questions.txt .

# Complex analysis with synthesis
thinktank complex-task.txt ./src --synthesis

# Debugging with verbose output
thinktank debug-task.txt ./problematic-code --verbose --debug
```

### Synthesis Feature

The synthesis feature automatically combines outputs from multiple models into a single coherent response. When you use the `--synthesis` flag or have large inputs that trigger multi-model analysis, thinktank will:

1. Process your instructions with multiple appropriate models
2. Save individual model outputs as usual
3. Send all model outputs to a synthesis model
4. Generate a final synthesized output that combines insights from all models

This is particularly useful for complex tasks where different models might have complementary strengths, or when you want to obtain a consensus view across multiple AI systems.

## Output

The output depends entirely on your instructions, but common use cases include:
- Technical implementation plans
- Code reviews and analysis
- Architecture recommendations
- Bug fixes and debugging assistance
- Documentation generation

Output files are saved in the specified directory (or auto-generated directory) with one file per model. If a synthesis model is specified, an additional file containing the synthesized output will be created with the naming format `<synthesis-model-name>-synthesis.md`.

### Modern CLI Output Format

thinktank features a modern, clean CLI output design inspired by tools like ripgrep, eza, and bat. The output automatically adapts to your environment (interactive terminals vs CI/automation) and provides clear, scannable results.

#### Interactive Terminal Output

In interactive terminals, thinktank displays Unicode symbols and semantic colors:

```
Processing 3 models...
[1/3] gemini-2.5-pro: ✓ completed (2.3s)
[2/3] gpt-4.1: ✓ completed (1.8s)
[3/3] o4-mini: ✗ rate limited

SUMMARY
───────
● 3 models processed
● 2 successful, 1 failed
● Synthesis: ✓ completed
● Output directory: ./thinktank_20250619_143022_7841

⚠ Partial success - some models failed
  ● Success rate: 67% (2/3 models)
  ● Check failed model details above for specific issues

OUTPUT FILES
────────────
  gemini-2.5-pro.md                                                2.4K
  gpt-4.1.md                                                       3.1K
  synthesis.md                                                     5.2K

FAILED MODELS
─────────────
  o4-mini                                                    rate limited
```

#### CI/Automation Output

In CI environments or when `CI=true`, output uses ASCII alternatives for maximum compatibility:

```
Processing 3 models...
Completed model 1/3: gemini-2.5-pro (2.3s)
Completed model 2/3: gpt-4.1 (1.8s)
Failed model 3/3: o4-mini (rate limited)

SUMMARY
-------
* 3 models processed
* 2 successful, 1 failed
* Synthesis: [OK] completed
* Output directory: ./thinktank_20250619_143022_7841

[!] Partial success - some models failed
  * Success rate: 67% (2/3 models)
  * Check failed model details above for specific issues

OUTPUT FILES
------------
  gemini-2.5-pro.md                                                2.4K
  gpt-4.1.md                                                       3.1K
  synthesis.md                                                     5.2K

FAILED MODELS
-------------
  o4-mini                                                    rate limited
```

#### Key Features

- **Environment Adaptation**: Automatically detects interactive vs CI environments
- **Unicode Fallback**: Graceful ASCII alternatives when Unicode isn't supported
- **Semantic Colors**: Green for success, red for errors, yellow for warnings (interactive only)
- **Responsive Layout**: Adapts to terminal width for optimal readability
- **Human-Readable Sizes**: File sizes displayed as "2.4K", "1.5M", etc.
- **Professional Aesthetics**: Clean, scannable output without emoji clutter

#### Output Control Flags

| Flag | Description | Use Case |
|------|-------------|----------|
| `--quiet`, `-q` | Suppress console output (errors only) | Scripting, when only caring about exit codes |
| `--json-logs` | Show JSON logs on stderr | Legacy behavior, structured logging |
| `--no-progress` | Disable progress indicators | Cleaner output for logs/CI |
| `--verbose` | Enable detailed logging | Debugging, troubleshooting |

### Output Directory Naming

When no `--output-dir` is specified, thinktank automatically generates a timestamped directory name using the format:

```
thinktank_YYYYMMDD_HHMMSS_NNNN
```

Where:
- `YYYYMMDD` is the current date (year, month, day)
- `HHMMSS` is the current time (hour, minute, second)
- `NNNN` is a 4-digit random number to ensure uniqueness

Examples:
- `thinktank_20250424_152230_3721` (April 24, 2025 at 15:22:30, random number 3721)
- `thinktank_20250424_152231_0498` (Same date, one second later, different random number)

This naming convention ensures that each run has a unique, sortable, and identifiable output directory.

## Error Handling and Troubleshooting

### Exit Codes

By default, thinktank returns:
- **0 (success)**: All models completed successfully
- **1 (failure)**: One or more models failed

When using `--partial-success-ok` (tolerant mode):
- **0 (success)**: At least one model succeeded and generated usable output (even if others failed)
- **1 (failure)**: All models failed or other critical error occurred

This tolerant mode is particularly useful when using multiple models for redundancy, allowing the process to succeed if at least one model delivers a valid result.

### Common Issues

**Quick Fixes:**

```bash
# Authentication Error
echo $OPENROUTER_API_KEY  # Check if API key is set
export OPENROUTER_API_KEY="your-key-here"  # Set it if missing

# Instructions file not found
ls instructions.txt  # Check file exists
echo "Analyze this code" > instructions.txt  # Create if missing

# Target path not found
ls ./my-project  # Check target exists
thinktank instructions.txt . --dry-run  # Use current directory

# Too much output
thinktank instructions.txt ./src --quiet  # Suppress console output

# Need more detail
thinktank instructions.txt ./src --verbose --debug  # Enable detailed logging
```

For comprehensive troubleshooting guidance, see **[Troubleshooting Guide](docs/TROUBLESHOOTING.md)**.

## Development & Contributing

### Code Coverage Requirements

The project maintains high test coverage standards to ensure reliability and maintainability:

Coverage is measured locally, not enforced in CI. Focus on writing good tests, not hitting arbitrary numbers.

#### Running Tests

```bash
# Run full test suite locally
./scripts/test-local.sh

# Run specific package tests
go test -race ./internal/models/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### CI Philosophy

Following Carmack's approach: fast feedback, no flaky tests, simple configuration. CI runs in <2 minutes with:
- Format verification (`go fmt`)
- Linting (`go vet`)
- Tests with race detection
- Binary build

See [CI Migration Guide](docs/carmack-ci-migration.md) for details.

### Testing Practices

- Write tests before implementing features (TDD approach)
- Focus on integration tests that verify workflows across components
- Mock only true external dependencies, not internal collaborators
- Maintain high coverage, particularly for critical components
- Run the full test suite before submitting changes

## Learn More

### User Documentation
- [Troubleshooting Guide](docs/TROUBLESHOOTING.md) - Comprehensive problem diagnosis and solutions
- [OpenRouter Integration](docs/openrouter-integration.md) - Using OpenRouter models

### Developer Documentation
- [Documentation Overview](docs/README.md) - Documentation organization and structure
- [Development Guide](docs/DEVELOPMENT.md) - Setup and development guidelines
- [Modern CLI Output Format](docs/MODERN_CLI_OUTPUT.md) - CLI output design
- [Error Handling and Logging Standards](docs/ERROR_HANDLING_AND_LOGGING.md) - Error handling patterns
- [Simple Parser Design](docs/simple-parser-design.md) - CLI parser architecture
- [Testing Documentation](docs/testing/) - Testing strategies and methodologies
- [Coverage Analysis](docs/coverage/) - Test coverage analysis and tools

### Operations & Quality
- [Quality Dashboard](docs/quality-dashboard/) - Quality metrics and monitoring
- [Security Documentation](docs/security/) - Security roadmap and scanning
- [Operations Docs](docs/operations/) - Performance, quality gates, and operations

For detailed configuration options, run: `thinktank --help`

## License

[MIT](LICENSE)
