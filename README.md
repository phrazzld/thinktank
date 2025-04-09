# thinktank

A CLI tool for querying multiple Large Language Models with the same prompt and comparing their responses.

## Overview

thinktank sends the same text prompt to multiple LLM providers simultaneously and displays their responses side-by-side.

- Compare how different models interpret the same prompt
- Find the best model for specific types of queries
- Test prompts across different providers
- Understand model differences

## Installation

### Prerequisites

- Node.js 18.x or higher
- pnpm ([installation instructions](https://pnpm.io/installation))

### Install from source

```bash
# Clone the repository
git clone https://github.com/phrazzld/thinktank.git
cd thinktank

# Install dependencies
pnpm install

# Build the project
pnpm run build

# Make CLI executable
chmod +x dist/src/cli/index.js

# Install globally
pnpm link --global
```

### Environment Variables

Create a `.env` file with API keys:

```
OPENAI_API_KEY=your_openai_api_key_here
ANTHROPIC_API_KEY=your_anthropic_api_key_here
GEMINI_API_KEY=your_gemini_api_key_here
OPENROUTER_API_KEY=your_openrouter_api_key_here
```

## Quick Start

```bash
# Send a prompt to default models
thinktank run prompt.txt

# Use a specific group of models
thinktank run prompt.txt --group coding

# Use a specific model
thinktank run prompt.txt --model openai:gpt-4o

# Include context files/directories
thinktank run prompt.txt file1.js directory/

# List available models
thinktank models list

# View configuration
thinktank config show
```

## Documentation

For detailed information, check these guides:

### User Guides
- [Usage Guide](docs/guides/usage.md)
- [Configuration Guide](docs/guides/configuration.md)
- [Troubleshooting](docs/reference/troubleshooting.md)

### Development
- [Architecture](docs/reference/architecture.md)
- [Contributing](docs/development/CONTRIBUTING.md)
- [Testing Philosophy](docs/development/TESTING_PHILOSOPHY.md)
- [Best Practices](docs/development/BEST_PRACTICES.md)
- [GitHub Actions](docs/development/GITHUB_ACTIONS.md)

### Project Management
- [Backlog](docs/project/BACKLOG.md)
- [Plan](docs/project/PLAN.md)
- [Todo](docs/project/TODO.md)
- [Done](docs/project/DONE.md)

## Features

- **Multiple Providers**: OpenAI, Anthropic, Google, OpenRouter
- **Context Files**: Include files and directories for context-aware responses
- **Configuration System**: Cascading configuration with sensible defaults
- **Claude Thinking**: Access Claude's thinking/reasoning process
- **Run Organization**: Automatic naming and organization of runs
- **Output Files**: Responses saved to files for later reference

## License

[MIT](LICENSE)

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](docs/development/CONTRIBUTING.md) for details.
