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
thinktank --instructions task.txt --model gemini-2.5-pro-exp-03-25 --model gpt-4-turbo ./
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
| `--model` | Model to use (repeatable) | `gemini-2.5-pro-exp-03-25` |
| `--synthesis-model` | Model to synthesize results from multiple models | None |
| `--output-dir` | Output directory | Auto-generated timestamp-based name |
| `--include` | File extensions to include (.go,.md) | All files |
| `--dry-run` | Preview without API calls | `false` |
| `--log-level` | Logging level (debug,info,warn,error) | `info` |

## Models Setup

1. Create config directory: `mkdir -p ~/.config/thinktank`
2. Copy default config: `cp config/models.yaml ~/.config/thinktank/`
3. Customize as needed for different models or custom endpoints

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
thinktank --instructions complex-task.txt --model gemini-2.5-pro-exp-03-25 --model gpt-4-turbo --synthesis-model gpt-4-turbo ./src
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

## Troubleshooting

- **Context Length Errors**: Reduce scope with `--include` or use a model with larger context
- **API Key Issues**: Ensure correct environment variables are set for each provider
- **No Files Processed**: Check paths and filters with `--dry-run`
- **Rate Limiting**: Adjust `--max-concurrent` (default: 5) and `--rate-limit` (default: 60)

## Learn More

- [OpenRouter Integration](docs/openrouter-integration.md)
- Detailed configuration options: `thinktank --help`

## License

[MIT](LICENSE)
