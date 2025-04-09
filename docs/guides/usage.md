# Usage Guide

This guide covers the essential commands and common usage patterns for thinktank.

## Basic Usage

```bash
# Send a prompt to default models
thinktank run prompt.txt

# Send a prompt to a specific group
thinktank run prompt.txt --group coding

# Send a prompt to one specific model
thinktank run prompt.txt --model openai:gpt-4o

# Send a prompt to multiple models
thinktank run prompt.txt --models openai:gpt-4o,anthropic:claude-3-opus

# List all available models
thinktank models list
```

## Including Context Files

You can include additional files or directories as context for your prompt:

```bash
# Include a single file as context
thinktank run prompt.txt path/to/context-file.js

# Include multiple files as context
thinktank run prompt.txt file1.js file2.md file3.txt

# Include a directory (recursively reads all files in the directory)
thinktank run prompt.txt path/to/directory/

# Mix of files and directories
thinktank run prompt.txt file1.js path/to/directory/ file2.md
```

When directories are provided, thinktank will:
- Recursively traverse the directory
- Respect `.gitignore` patterns
- Skip binary files
- Apply size limits

## Working with Model Groups

```bash
# View available groups
thinktank config groups list

# Create a new group
thinktank config groups create coding --models openai:gpt-4o,anthropic:claude-3-opus

# Use a group for a query
thinktank run prompt.txt --group coding
```

## Claude's Thinking Capability

For Claude models, you can enable "thinking" to see the model's reasoning process:

```bash
# Enable thinking for Claude models
thinktank run prompt.txt --thinking

# Display thinking output in results
thinktank run prompt.txt --thinking --show-thinking
```

Note: When using thinking, temperature is automatically set to 1.

## Output Files

thinktank saves individual model responses to separate files in the output directory:

```
./thinktank-output/run-YYYYMMDD-HHMMSS/
├── openai-gpt-4o.md
├── anthropic-claude-3-opus.md
└── ...
```

Each file includes:
- The model's response
- Metadata like token usage
- Timestamp

## Additional Options

| Option | Description |
|--------|-------------|
| `--help`, `-h` | Show help information |
| `--version`, `-v` | Show version number |
| `--no-color` | Disable colored output |
| `--thinking` | Enable Claude's thinking capability |
| `--show-thinking` | Display thinking output |
| `--config` | Specify custom config file |

For more detailed information, see the [Configuration Guide](./configuration.md).