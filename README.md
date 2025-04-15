# architect

A powerful code-base context analysis and planning tool that leverages Google's Gemini and OpenAI models to generate detailed, actionable technical plans for software projects.

## Overview

architect analyzes your codebase and uses Gemini or OpenAI models to create comprehensive technical plans for new features, refactoring, bug fixes, or any software development task. By understanding your existing code structure and patterns, architect provides contextually relevant guidance tailored to your specific project.

## Important Updates

### Instruction Input Method

**Please note:** The way you provide instructions to architect has changed:

* The `--instructions` flag is now **required** for generating plans. You must provide the instructions in a file. This allows for more complex and structured inputs.
* The `--task-file` flag has been **removed** and replaced with the `--instructions` flag.
* For `--dry-run` operations, the `--instructions` flag is optional. You can run a simple dry run with just `architect --dry-run ./` to see which files would be included.

Please update your workflows accordingly. See the Usage examples and Configuration Options below for details.

### Upcoming API Changes (v0.8.0)

**Important:** The following internal APIs are deprecated and will be removed in v0.8.0 (Q4 2024):

* `InitClient` method - Use `InitLLMClient` instead
* `ProcessResponse` method - Use `ProcessLLMResponse` instead
* `llmToGeminiClientAdapter` - Use the provider-agnostic `llm.LLMClient` interface directly
* The entire `internal/architect/compat` package

If you're developing with or extending Architect, please migrate to the provider-agnostic methods. See the [Migration Guide](MIGRATION-GUIDE.md) and [Deprecation Plan](DEPRECATED-API-REMOVAL-PLAN.md) for details.

## Features

- **Contextual Analysis**: Analyzes your codebase to understand its structure, patterns, and dependencies
- **Smart Filtering**: Include/exclude specific file types or directories from analysis
- **Multiple AI Providers**: Support for both Gemini and OpenAI models
- **Customizable Output**: Configure the format of the generated plan
- **Git-Aware**: Respects .gitignore patterns when scanning your codebase
- **Token Management**: Smart token counting with limit checking to prevent API errors
- **Dry Run Mode**: Preview which files would be included and see token statistics before API calls
- **XML-Structured Approach**: Uses a simple XML structure for instructions and context
- **Instructions File Input**: Load instructions from external files
- **Structured Logging**: Clear, structured logs with configurable verbosity levels and optional audit logging
- **User Confirmation**: Optional confirmation for large token counts

## Installation

```bash
# Build from source
git clone https://github.com/yourusername/architect.git
cd architect

# Recommended: Run the setup script to check dependencies and install pre-commit hooks
./scripts/setup.sh

# Or build manually
go build

# Install globally
go install
```

The setup script checks for required dependencies (like Go and pre-commit), offers to install them if missing, and configures the development environment automatically.

## Usage

```bash
# Basic usage (Instructions in instructions.txt)
architect --instructions instructions.txt path/to/your/project

# Example: Create a plan using an instructions file
# Contents of auth_instructions.txt: "Implement JWT-based user authentication and authorization"
architect --instructions auth_instructions.txt ./

# Specify output directory (default generates a unique run name directory)
architect --instructions instructions.txt --output-dir custom-dir ./

# Generate plans with multiple models (repeatable flag)
architect --instructions instructions.txt --model gemini-1.5-pro --model gemini-2.5-pro-exp-03-25 ./

# Include only specific file extensions
architect --instructions instructions.txt --include .go,.md ./

# Use a different Gemini model
architect --instructions instructions.txt --model gemini-2.5-pro-exp-03-25 ./

# Use an OpenAI model
architect --instructions instructions.txt --model gpt-4-turbo ./

# Use both Gemini and OpenAI models
architect --instructions instructions.txt --model gemini-2.5-pro-exp-03-25 --model gpt-4-turbo ./

# Enable structured audit logging (JSON Lines format)
architect --instructions instructions.txt --audit-log-file audit.jsonl ./

# Dry run to see which files would be included (without generating a plan)
architect --dry-run ./

# Dry run with instructions file
architect --dry-run --instructions instructions.txt ./

# Request confirmation before proceeding if token count exceeds threshold
architect --instructions instructions.txt --confirm-tokens 25000 ./

# Control concurrency and rate limiting for multiple models
architect --instructions task.txt --model gemini-1.5-pro --model gpt-4-turbo --max-concurrent 3 --rate-limit 30 ./
```

### Required Environment Variables

```bash
# Set your Gemini API key (required for Gemini models)
export GEMINI_API_KEY="your-gemini-api-key-here"

# Set your OpenAI API key (required for OpenAI models)
export OPENAI_API_KEY="your-openai-api-key-here"
```

> **Note**: You only need to set the API key for the provider(s) you're using. If you're exclusively using Gemini models, only `GEMINI_API_KEY` is required. Similarly, if you're only using OpenAI models, only `OPENAI_API_KEY` is required. If you're using both providers, both environment variables need to be set.

## Configuration Options

| Flag | Description | Default |
|------|-------------|---------|
| `--instructions` | Path to a file containing the instructions (Required unless --dry-run is used) | `(Required)` |
| `--output-dir` | Directory path to store generated plans (one per model) | `(Auto-generated run name)` |
| `--model` | Model to use for generation (repeatable for multiple models, e.g., `gemini-2.5-pro-exp-03-25`, `gpt-4-turbo`) | `gemini-2.5-pro-exp-03-25` |
| `--verbose` | Enable verbose logging output (shorthand for --log-level=debug) | `false` |
| `--log-level` | Set logging level (debug, info, warn, error) | `info` |
| `--include` | Comma-separated list of file extensions to include | (All files) |
| `--exclude` | Comma-separated list of file extensions to exclude | (Common binary and media files) |
| `--exclude-names` | Comma-separated list of file/dir names to exclude | (Common directories like .git, node_modules) |
| `--format` | Format string for each file. Use {path} and {content} | `<{path}>\n```\n{content}\n```\n</{path}>\n\n` |
| `--dry-run` | Show files that would be included and token count, but don't call the API | `false` |
| `--confirm-tokens` | Prompt for confirmation if token count exceeds this value (0 = never prompt) | `0` |
| `--audit-log-file` | Path to write structured audit logs (JSON Lines format) | `(Disabled)` |
| `--max-concurrent` | Maximum number of concurrent API requests (0 = no limit) | `5` |
| `--rate-limit` | Maximum requests per minute per model (0 = no limit) | `60` |

## Configuration

architect is configured entirely through command-line flags and environment variables (`GEMINI_API_KEY` and/or `OPENAI_API_KEY`). There are no configuration files to manage, which simplifies usage and deployment.

### Default Values

architect comes with sensible defaults for most options:

- Output directory: Auto-generated run name directory (e.g., `eloquent-rabbit`)
- Output files: One file per model with name format `modelname.md`
- Model: `gemini-2.5-pro-exp-03-25` (can specify multiple models with repeatable `--model` flag)
- Supported models:
  - Gemini models: `gemini-1.5-pro`, `gemini-2.5-pro-exp-03-25`, etc.
  - OpenAI models: `gpt-4-turbo`, `gpt-4`, `gpt-3.5-turbo`, etc.
- Log level: `info`
- File formatting: XML-style format for context building
- Default exclusions for common binary, media files, and directories

You can override any of these defaults using the appropriate command-line flags.


## Token Management

architect implements intelligent token management to prevent API errors and optimize context:

- **Accurate Token Counting**: Uses provider-specific APIs to get precise token counts for your content
- **Pre-API Validation**: Checks token count against model limits before making API calls
- **Fail-Fast Strategy**: Provides clear error messages when token limits would be exceeded
- **Token Statistics**: Shows token usage as a percentage of the model's limit
- **Optional Confirmation**: Can prompt for confirmation when token count exceeds a threshold

When token limits are exceeded, try:
1. Limiting file scope with `--include` or additional filtering flags
2. Using a model with a higher token limit
3. Splitting your task into smaller, more focused requests

## Multi-Provider and Multi-Model Support

architect supports generating plans with multiple AI models from multiple providers simultaneously:

- **Multiple Models**: Specify multiple models with the repeatable `--model` flag
- **Multiple Providers**: Seamlessly use both Gemini and OpenAI models in the same run
- **Provider Detection**: Automatically detects the appropriate provider based on model name
- **Organized Output**: Each model's plan is saved as a separate file in the output directory
- **Run Name Directories**: Automatically creates a uniquely named directory for each run
- **Concurrent Processing**: Processes multiple models in parallel using Go's concurrency primitives
- **Rate Limiting**: Prevents overwhelming API endpoints with configurable concurrency and rate limits
- **Error Isolation**: If one model fails, others will still complete successfully
- **Output Files**: Each model's output is saved in its own file within the output directory

Example:
```bash
# Generate plans with both Gemini and OpenAI models
architect --instructions task.md --model gemini-1.5-pro --model gpt-4-turbo ./src
```

This will generate:
```
/current-dir/random-runname/
  ├─ gemini-1.5-pro.md
  └─ gpt-4-turbo.md
```

### Concurrency Control

architect processes multiple models concurrently for improved performance, with built-in safeguards to prevent overwhelming API endpoints:

- **Concurrent Execution**: Multiple model requests run in parallel using goroutines
- **Per-Model Rate Limiting**: Each model has its own rate limit bucket to prevent API throttling
- **Provider-Aware Processing**: Handles different API requirements for Gemini and OpenAI automatically
- **Configurable Limits**: Adjust concurrency and rate limits to match your API quota

## Generated Plan Structure

Each generated plan file includes:

1. **Overview**: Brief explanation of the task and changes involved
2. **Task Breakdown**: Detailed list of specific tasks with effort estimates
3. **Implementation Details**: Guidance for complex tasks with code snippets
4. **Potential Challenges**: Identified risks, edge cases, and dependencies
5. **Testing Strategy**: Approach for verifying the implementation
6. **Open Questions**: Ambiguities needing clarification


## XML-Structured Approach

architect uses a simple XML structure to organize instructions and context:

1. Instructions are wrapped in `<instructions>...</instructions>` tags
2. Context files are wrapped in `<context>...</context>` tags
3. Each file in the context is formatted with its path and content
4. All XML special characters in file content are properly escaped

This approach provides a clear separation between the instructions and the codebase context, making it easier for the LLM to understand and process the input.

## Audit Logging

The structured audit logging feature records detailed information about each operation:

- **Format**: JSON Lines format for easy machine parsing
- **Content**: Records timestamps, operations, status, duration, token counts, and errors
- **Usage**: Enable with `--audit-log-file audit.jsonl` (replace with your desired path)
- **Purpose**: Provides auditability and data for programmatic analysis

Each log entry contains structured information that can be processed by tools, with fields including:
- Operation name
- Status (success/failure)
- Execution duration
- Input/output details
- Token counts
- Error information (if applicable)

## Troubleshooting

### API Key Issues
- For Gemini models, ensure `GEMINI_API_KEY` is set correctly in your environment
- For OpenAI models, ensure `OPENAI_API_KEY` is set correctly in your environment
- Check that your API keys have appropriate permissions for the models you're using

### Token Limit Errors
- Use `--dry-run` to see token statistics without making API calls
- Reduce the number of files analyzed with `--include` or other filtering flags
- Try a model with a higher token limit

### Performance Tips
- Start with small, focused parts of your codebase and gradually include more
- Use `--include` to target only relevant file types for your task
- Set `--log-level=debug` for detailed information about the analysis process
- Use multiple models with the `--model` flag (processed concurrently)
- Adjust `--max-concurrent` based on your system's resources and API limits

### Rate Limiting Issues
- If you encounter API rate limit errors, reduce the `--max-concurrent` value (e.g., from 5 to 3)
- Adjust the `--rate-limit` value to match your API quota (requests per minute per model)
- For large projects with multiple models, consider more conservative limits (e.g., `--max-concurrent 2 --rate-limit 30`)
- Remember that rate limits are applied per model, so using multiple models with high request rates may still encounter API throttling

### Common Issues
- **No files processed**: Check paths and filters; use `--dry-run` to see what would be included
- **Missing API key**: Ensure the appropriate API key environment variable (`GEMINI_API_KEY` or `OPENAI_API_KEY`) is set correctly for the models you're using
- **Path issues**: When running commands, use absolute or correct relative paths to your project files
- **Flag precedence**: Remember that CLI flags always take precedence over default values
- **Model name errors**: Ensure you're using valid model names with the `--model` flag (prefix with `gemini-` for Gemini models or `gpt-` for OpenAI models)
- **Output directory permissions**: Check you have write access to the output directory when using `--output-dir`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)
