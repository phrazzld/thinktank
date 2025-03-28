# Architect

A powerful code-base context analysis and planning tool that leverages Google's Gemini AI to generate detailed, actionable technical plans for software projects.

## Overview

Architect analyzes your codebase and uses Gemini AI to create comprehensive technical plans for new features, refactoring, bug fixes, or any software development task. By understanding your existing code structure and patterns, Architect provides contextually relevant guidance tailored to your specific project.

## Features

- **Contextual Analysis**: Analyzes your codebase to understand its structure, patterns, and dependencies
- **Smart Filtering**: Include/exclude specific file types or directories from analysis
- **Gemini AI Integration**: Leverages Google's powerful Gemini models for intelligent planning
- **Customizable Output**: Configure the format of the generated plan
- **Git-Aware**: Respects .gitignore patterns when scanning your codebase

## Installation

```bash
# Build from source
git clone https://github.com/yourusername/architect.git
cd architect
go build
```

## Usage

```bash
# Basic usage
architect --task "Your task description" path/to/your/project

# Example: Create a plan for implementing user authentication
architect --task "Implement JWT-based user authentication and authorization" ./

# Specify output file (default is PLAN.md)
architect --task "..." --output auth_plan.md ./

# Include only specific file extensions
architect --task "..." --include .go,.md ./

# Use a different Gemini model
architect --task "..." --model gemini-1.5-flash ./
```

### Required Environment Variable

```bash
# Set your Gemini API key
export GEMINI_API_KEY="your-api-key-here"
```

## Configuration Options

| Flag | Description | Default |
|------|-------------|---------|
| `--task` | Description of the task or goal for the plan | (Required) |
| `--output` | Output file path for the generated plan | `PLAN.md` |
| `--model` | Gemini model to use for generation | `gemini-2.5-pro-exp-03-25` |
| `--verbose` | Enable verbose logging output | `false` |
| `--include` | Comma-separated list of file extensions to include | (All files) |
| `--exclude` | Comma-separated list of file extensions to exclude | (Common binary and media files) |
| `--exclude-names` | Comma-separated list of file/dir names to exclude | (Common directories like .git, node_modules) |
| `--format` | Format string for each file. Use {path} and {content} | `<{path}>\n```\n{content}\n```\n</{path}>\n\n` |

## Generated Plan Structure

The generated PLAN.md includes:

1. **Overview**: Brief explanation of the task and changes involved
2. **Task Breakdown**: Detailed list of specific tasks with effort estimates
3. **Implementation Details**: Guidance for complex tasks with code snippets
4. **Potential Challenges**: Identified risks, edge cases, and dependencies
5. **Testing Strategy**: Approach for verifying the implementation
6. **Open Questions**: Ambiguities needing clarification

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)