# thinktank

A context-aware LLM tool that analyzes codebases and generates responses to your instructions using Gemini, OpenAI, or OpenRouter models.

## Quick Start

```bash
# Installation
git clone https://github.com/phrazzld/thinktank.git
cd thinktank
go install

# Set required API key(s) for the model(s) you want to use
export GEMINI_API_KEY="your-key"  # For Gemini models
export OPENAI_API_KEY="your-key"  # For OpenAI models
export OPENROUTER_API_KEY="your-key"  # For OpenRouter models

# Basic usage
thinktank --instructions task.txt ./my-project

# Multiple models
thinktank --instructions task.txt --model gemini-2.5-pro --model gpt-4-turbo ./
```

## Key Features

- **Multiple LLM Providers**: Supports Gemini, OpenAI, and OpenRouter models
- **Smart Filtering**: Include/exclude specific files or directories
- **Concurrent Processing**: Compare responses from multiple models in parallel
- **Result Synthesis**: Combine outputs from multiple models using a synthesis model
- **Git-Aware**: Respects .gitignore patterns
- **Structured Output**: Formats responses based on your specific instructions

## Configuration

### Required
- **Instructions File**: `--instructions task.txt` (Required except for dry runs)
- **API Keys**: Environment variables for each model provider you use

### Common Options

| Flag | Description | Default |
|------|-------------|---------|
| `--model` | Model to use (repeatable) | `gemini-2.5-pro` |
| `--synthesis-model` | Model to synthesize results from multiple models | None |
| `--output-dir` | Output directory | Auto-generated timestamp-based name |
| `--include` | File extensions to include (.go,.md) | All files |
| `--dry-run` | Preview without API calls | `false` |
| `--partial-success-ok` | Return success code if any model succeeds | `false` |
| `--log-level` | Logging level (debug,info,warn,error) | `info` |

### Output and Logging Options

| Flag | Description | Default |
|------|-------------|---------|
| `--quiet`, `-q` | Suppress console output (errors only) | `false` |
| `--json-logs` | Show JSON logs on stderr (preserves old behavior) | `false` |
| `--no-progress` | Disable progress indicators (show only start/complete) | `false` |
| `--verbose` | Enable both console output AND JSON logs to stderr | `false` |

## Supported Models

thinktank supports the following LLM models out of the box:

**OpenAI Models:**
  - gpt-4.1
  - o4-mini

**Gemini Models:**
  - gemini-2.5-flash
  - gemini-2.5-pro

**OpenRouter Models:**
  - openrouter/deepseek/deepseek-chat-v3-0324
  - openrouter/deepseek/deepseek-r1
  - openrouter/x-ai/grok-3-beta

No additional configuration is needed - simply set the appropriate API key environment variable and use any supported model name with the `--model` flag.

### Adding New Models

To add a new model, edit `internal/models/models.go` directly:

1. Add a new entry to the `ModelDefinitions` map with the model name as key
2. Provide the required `ModelInfo` struct with:
   - `Provider`: The provider name (openai, gemini, or openrouter)
   - `APIModelID`: The actual model ID used in API calls
   - `ContextWindow`: Maximum input + output tokens
   - `MaxOutputTokens`: Maximum output tokens
   - `DefaultParams`: Provider-specific parameters (temperature, top_p, etc.)
3. Run tests: `go test ./internal/models`
4. Submit a pull request with your changes

Example:
```go
"new-model-name": {
    Provider:        "openai",
    APIModelID:      "gpt-5",
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
thinktank --instructions feature-plan.txt ./src

# Code review
thinktank --instructions code-review.txt --model gpt-4-turbo ./pull-request

# Architecture analysis
thinktank --instructions arch-questions.txt --include .go,.md,.yaml ./

# General code questions
thinktank --instructions questions.txt --output-dir answers ./src

# Using synthesis to combine multiple model outputs
thinktank --instructions complex-task.txt --model gemini-2.5-pro --model gpt-4-turbo --synthesis-model gpt-4-turbo ./src

# Allow partial success (return success code even if some models fail)
thinktank --instructions task.txt --model model1 --model model2 --partial-success-ok ./src
```

### Synthesis Feature

The synthesis feature allows you to combine outputs from multiple models into a single coherent response. When you specify multiple models with `--model` and set a synthesis model with `--synthesis-model`, thinktank will:

1. Process your instructions with each specified primary model
2. Save individual model outputs as usual
3. Send all model outputs to the synthesis model
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

- **Context Length Errors**: Reduce scope with `--include` or use a model with larger context
- **API Key Issues**: Ensure correct environment variables are set for each provider
- **No Files Processed**: Check paths and filters with `--dry-run`
- **Rate Limiting**: Adjust `--max-concurrent` (default: 5) and `--rate-limit` (default: 60)
- **Model Availability**: If one model is unreliable, use `--partial-success-ok` to allow other models to succeed
- **Terminal Compatibility**: Output automatically adapts to your environment. Use `--no-progress` for minimal output or `--json-logs` for structured logging.

## Development & Contributing

### Code Coverage Requirements

The project maintains high test coverage standards to ensure reliability and maintainability:

- **Target Coverage**: 90% overall code coverage
- **Minimum Threshold**: 75% overall and per-package (enforced in CI)
- **Core APIs**: Special focus on complete coverage for core API components

#### Coverage Tools

Several scripts are available to check and validate test coverage:

| Script | Description | Usage |
|--------|-------------|-------|
| `check-coverage.sh` | Checks overall coverage against threshold | `./scripts/check-coverage.sh [threshold]` |
| `check-package-coverage.sh` | Validates per-package coverage | `./scripts/check-package-coverage.sh [threshold]` |
| `check-registry-coverage.sh` | Reports coverage for models package components | `./scripts/check-registry-coverage.sh [threshold]` |
| `pre-submit-coverage.sh` | Comprehensive pre-submission check | `./scripts/pre-submit-coverage.sh [options]` |

#### Pre-Submission Coverage Validation

Before submitting code, run the pre-submission coverage check script to ensure your changes maintain or improve coverage:

```bash
# Basic coverage check with default threshold (75%)
./scripts/pre-submit-coverage.sh

# With custom threshold and verbose output
./scripts/pre-submit-coverage.sh --threshold 80 --verbose

# Including models package specific checks
./scripts/pre-submit-coverage.sh --registry --verbose
```

See `./scripts/pre-submit-coverage.sh --help` for additional options.

### Testing Practices

- Write tests before implementing features (TDD approach)
- Focus on integration tests that verify workflows across components
- Mock only true external dependencies, not internal collaborators
- Maintain high coverage, particularly for critical components
- Run the full test suite before submitting changes

## Learn More

- [Modern CLI Output Format & Rollback Guide](docs/MODERN_CLI_OUTPUT.md)
- [OpenRouter Integration](docs/openrouter-integration.md)
- [Development Philosophy](docs/DEVELOPMENT_PHILOSOPHY.md)
- [Error Handling and Logging Standards](docs/ERROR_HANDLING_AND_LOGGING.md)
- Detailed configuration options: `thinktank --help`

## License

[MIT](LICENSE)
