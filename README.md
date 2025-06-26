# thinktank

A context-aware LLM tool that analyzes codebases and generates responses to your instructions using Gemini, OpenAI, or OpenRouter models.

## Quick Start

```bash
# Installation
git clone https://github.com/phrazzld/thinktank.git
cd thinktank
go install

# Set API key for your preferred model
export GEMINI_API_KEY="your-key"     # For Gemini models (recommended)
export OPENAI_API_KEY="your-key"     # For OpenAI models
export OPENROUTER_API_KEY="your-key" # For OpenRouter models

# Create instructions file
echo "Analyze this codebase and suggest improvements" > instructions.txt

# Basic usage - Simplified Interface (Recommended)
thinktank instructions.txt ./my-project

# Preview without API calls
thinktank instructions.txt ./my-project --dry-run

# Verbose output
thinktank instructions.txt ./my-project --verbose

# With custom model (using environment variable)
export THINKTANK_MODEL="gpt-4o"
thinktank instructions.txt ./my-project
```

## Key Features

- **Multiple LLM Providers**: Supports Gemini, OpenAI, and OpenRouter models
- **Smart Filtering**: Include/exclude specific files or directories
- **Concurrent Processing**: Compare responses from multiple models in parallel
- **Result Synthesis**: Combine outputs from multiple models using a synthesis model
- **Git-Aware**: Respects .gitignore patterns
- **Structured Output**: Formats responses based on your specific instructions

## Interface Options

thinktank supports two interfaces for different user needs:

### Simplified Interface (Recommended)
Perfect for most users. Uses positional arguments for a clean, intuitive experience:

```bash
# Pattern: thinktank instructions.txt target_path [options]
thinktank task.md ./src                    # Basic analysis with smart defaults
thinktank task.md ./src --dry-run          # Preview files and token count
thinktank task.md ./src --verbose          # With detailed output
```

### Complex Interface (Advanced/Legacy)
For advanced users who need extensive configuration. Uses traditional flags:

```bash
# Pattern: thinktank --instructions file.txt [options] target_path
thinktank --instructions task.md --model gpt-4o ./src
```

**💡 Migration Tip**: If you're using the complex interface, simply move the instructions file to the first position to use the simplified interface.

## Migration Guide

### From Complex to Simplified Interface

If you're currently using the complex interface, here's how to migrate to the cleaner simplified interface:

#### Basic Pattern
```bash
# Old (Complex Interface)
thinktank --instructions task.txt ./src

# New (Simplified Interface)
thinktank task.txt ./src
```

#### Common Migrations
| Old Complex Command | New Simplified Command |
|-------------------|---------------------|
| `thinktank --instructions task.txt --model gpt-4o ./src` | `export THINKTANK_MODEL="gpt-4o" && thinktank task.txt ./src` |
| `thinktank --instructions task.txt --dry-run ./src` | `thinktank task.txt ./src --dry-run` |
| `thinktank --instructions task.txt --verbose ./src` | `thinktank task.txt ./src --verbose` |
| `thinktank --instructions task.txt --output-dir out ./src` | `export THINKTANK_OUTPUT_DIR="out" && thinktank task.txt ./src` |

#### Environment Variables for Advanced Configuration

For advanced settings, use environment variables instead of complex flags:

```bash
# Rate limiting
export THINKTANK_OPENAI_RATE_LIMIT=100
export THINKTANK_GEMINI_RATE_LIMIT=1000

# Concurrency
export THINKTANK_MAX_CONCURRENT=5

# Behavior
export THINKTANK_SUPPRESS_DEPRECATION_WARNINGS=1

# Then use simplified interface
thinktank task.txt ./src
```

#### Automatic Detection

thinktank automatically detects which interface you're using:
- **Simplified**: First argument doesn't start with `--` and has `.txt` or `.md` extension
- **Complex**: First argument starts with `--` (like `--instructions`)

Both interfaces work simultaneously, so you can migrate gradually.

## Configuration

### Required
- **Instructions File**: Text file with your analysis request (first argument in simplified interface)
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

## Rate Limiting & Performance Optimization

thinktank provides intelligent rate limiting with provider-specific optimizations to help you get the best performance while staying within API limits.

### Rate Limiting Hierarchy

Rate limits are applied in this priority order:
1. **Model-specific overrides** (for special models like DeepSeek R1)
2. **Provider-specific settings** (`--openai-rate-limit`, `--gemini-rate-limit`, `--openrouter-rate-limit`)
3. **Global rate limit** (`--rate-limit`)
4. **Provider defaults** (automatic based on typical API capabilities)

### Provider-Specific Defaults & Recommendations

#### OpenAI
- **Default**: 3000 RPM (optimized for paid accounts)
- **Free tier**: Use `--openai-rate-limit 20` for Tier 1 accounts
- **Production**: Most paid accounts can handle the 3000 RPM default
- **High-volume**: Tier 4+ accounts can use `--openai-rate-limit 10000` or higher

#### Google Gemini
- **Default**: 60 RPM (balanced for free and paid tiers)
- **Free tier**: Use `--gemini-rate-limit 15` (aligns with Gemini 1.5 Flash free limit)
- **Paid tier**: Use `--gemini-rate-limit 1000` for significantly faster processing

#### OpenRouter
- **Default**: 20 RPM (conservative for mixed model usage)
- **Free models**: Automatic limit of 20 RPM (cannot be increased)
- **With $10+ balance**: Most models support higher rates; try `--openrouter-rate-limit 100`
- **High-volume**: With sufficient balance, use `--openrouter-rate-limit 500`

### Special Model Considerations

Some models have specific rate limits regardless of provider settings:
- `openrouter/deepseek/deepseek-r1-0528`: Limited to 5 RPM (reasoning model)
- `openrouter/deepseek/deepseek-r1-0528:free`: Limited to 3 RPM (free tier)

### Common Rate Limiting Scenarios

#### Multi-Provider Usage
```bash
# Optimize for mixed providers with different API tiers
thinktank task.txt ./src \
  --model gpt-4.1 --model gemini-2.5-pro --model openrouter/meta-llama/llama-3.3-70b-instruct \
  --openai-rate-limit 100 \
  --gemini-rate-limit 1000 \
  --openrouter-rate-limit 50
```

#### High-Performance Processing
```bash
# For production workloads with paid API tiers
thinktank analyze.txt ./large-codebase \
  --model gpt-4.1 --model gemini-2.5-pro \
  --max-concurrent 10 \
  --openai-rate-limit 5000 \
  --gemini-rate-limit 1000
```

#### Conservative/Budget Mode
```bash
# For free tiers or budget-conscious usage
thinktank task.txt ./project \
  --model gemini-2.5-flash \
  --max-concurrent 2 \
  --gemini-rate-limit 15
```

#### Single Provider Optimization
```bash
# OpenAI-only with high-tier account
thinktank task.txt ./src \
  --model gpt-4.1 --model o4-mini \
  --openai-rate-limit 8000 \
  --max-concurrent 15

# Gemini-only with paid account
thinktank task.txt ./src \
  --model gemini-2.5-pro --model gemini-2.5-flash \
  --gemini-rate-limit 1000 \
  --max-concurrent 8
```

### Troubleshooting Rate Limits

#### Symptoms of Rate Limiting Issues
- Models showing "rate limited" in output
- Long delays between model processing
- 429 errors in verbose logs

#### Solutions by Provider

**OpenAI 429 Errors:**
- Reduce `--openai-rate-limit` (try 100-500 for lower tiers)
- Check your usage tier at [OpenAI Platform](https://platform.openai.com/usage)
- Consider upgrading your account tier for higher limits

**Gemini Rate Limiting:**
- Free tier: Reduce `--gemini-rate-limit` to 10-15
- Paid tier: Check quota at [Google AI Studio](https://makersuite.google.com/)
- For persistent issues, try `--max-concurrent 3`

**OpenRouter Issues:**
- Ensure account balance > $10 for higher daily limits
- Use `--openrouter-rate-limit 10` for conservative usage
- Check model-specific limits at [OpenRouter Models](https://openrouter.ai/models)

#### General Optimization Tips
- Start conservative and increase gradually
- Use `--dry-run` to estimate request volume before processing
- Monitor actual usage patterns with `--verbose` logging
- Consider time-based processing for large codebases

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
thinktank feature-plan.txt ./src

# Code review with specific model
export THINKTANK_MODEL="gpt-4o"
thinktank code-review.txt ./pull-request

# Architecture analysis with file filtering
export THINKTANK_INCLUDE=".go,.md,.yaml"
thinktank arch-questions.txt .

# General code questions with custom output
export THINKTANK_OUTPUT_DIR="answers"
thinktank questions.txt ./src

# Multiple model comparison (using complex interface)
thinktank --instructions complex-task.txt --model gemini-2.5-pro --model gpt-4o ./src

# Synthesis (combine outputs from multiple models)
thinktank --instructions complex-task.txt --model gemini-2.5-pro --model gpt-4o --synthesis-model gpt-4o ./src

# Partial success mode (continue if some models fail)
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

**Quick Fixes for the Simplified Interface:**

```bash
# Authentication Error
echo $GEMINI_API_KEY  # Check if API key is set
export GEMINI_API_KEY="your-key-here"  # Set it if missing

# Instructions file not found
ls instructions.txt  # Check file exists
echo "Analyze this code" > instructions.txt  # Create if missing

# Target path not found
ls ./my-project  # Check target exists
thinktank instructions.txt . --dry-run  # Use current directory

# Rate limiting issues
thinktank instructions.txt ./src --gemini-rate-limit 15  # Reduce rate for free tier

# Input too large
thinktank instructions.txt ./src --include .go,.md  # Filter file types
thinktank instructions.txt ./src/specific-folder  # Target specific folder
```

**Complex Interface Issues:**

For comprehensive troubleshooting guidance, see **[Troubleshooting Guide](docs/TROUBLESHOOTING.md)**.

**Traditional flag-based solutions:**
- **Authentication Errors**: Check API key environment variables (`OPENAI_API_KEY`, `GEMINI_API_KEY`, `OPENROUTER_API_KEY`)
- **Rate Limiting**: Use `--openai-rate-limit`, `--gemini-rate-limit`, `--openrouter-rate-limit` flags to match your account tier
- **Input Too Large**: Use `--include` or `--exclude` flags to filter files, or target specific directories
- **Network Issues**: Retry the command - most network errors are temporary
- **Model Failures**: Use `--partial-success-ok` for redundancy when using multiple models
- **CI/Automation**: Set `THINKTANK_SUPPRESS_DEPRECATION_WARNINGS=1` to hide deprecation warnings in automated environments

For detailed diagnosis steps, error code references, and provider-specific solutions, see the [Troubleshooting Guide](docs/TROUBLESHOOTING.md).

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

- [Troubleshooting Guide](docs/TROUBLESHOOTING.md) - Comprehensive problem diagnosis and solutions
- [Modern CLI Output Format & Rollback Guide](docs/MODERN_CLI_OUTPUT.md)
- [OpenRouter Integration](docs/openrouter-integration.md)
- [Development Philosophy](docs/DEVELOPMENT_PHILOSOPHY.md)
- [Error Handling and Logging Standards](docs/ERROR_HANDLING_AND_LOGGING.md)
- Detailed configuration options: `thinktank --help`

## License

[MIT](LICENSE)
