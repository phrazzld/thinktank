// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"fmt"
	"io"
	"os"
)

// PrintHelp displays comprehensive usage information to the provided writer.
// Following Unix tradition with Go clarity, this provides everything a user
// needs to effectively use the thinktank CLI.
func PrintHelp(w io.Writer) {
	_, _ = fmt.Fprint(w, helpText)
}

// PrintHelpToStdout is a convenience function that prints help to stdout
func PrintHelpToStdout() {
	PrintHelp(os.Stdout)
}

// helpText contains the comprehensive help documentation.
// This is a constant to ensure consistent help output and zero runtime cost.
const helpText = `thinktank - AI-powered code analysis and synthesis tool

USAGE:
    thinktank instructions.txt target_path... [flags]

DESCRIPTION:
    Thinktank analyzes codebases and generates responses based on your
    instructions using AI models. It can process multiple files and
    directories, apply various analysis strategies, and synthesize
    responses from multiple models.

ARGUMENTS:
    instructions.txt    Path to instructions file (.txt or .md format)
                       Contains the task or questions for the AI models

    target_path...     One or more files or directories to analyze
                       Can specify multiple paths separated by spaces

FLAGS:
    --help, -h         Show this help message and exit

    --dry-run          Preview what would be processed without making API calls
                       Shows file list, token count, and model selection

    --verbose          Enable detailed output and debug logging
                       Includes API responses and processing details

    --synthesis        Force synthesis mode with multiple models
                       Combines responses from different models for better results

    --model MODEL      Select specific AI model (default: gemini-2.5-pro)
                       Available: gemini-2.5-pro, gpt-4.1, o4-mini, and more

    --output-dir DIR   Set output directory (default: auto-generated timestamp)
                       Created if it doesn't exist

    --quiet            Suppress non-essential console output
                       Only shows errors and final results

    --json-logs        Output structured JSON logs to stderr
                       Useful for debugging and integration

    --no-progress      Disable progress indicators
                       Helpful for CI environments or log capture

    --debug            Enable debug-level logging
                       Maximum verbosity for troubleshooting

EXAMPLES:
    # Basic usage - analyze a single directory
    thinktank instructions.md ./src

    # Analyze multiple paths
    thinktank task.txt ./src ./tests ./docs

    # Preview without making API calls
    thinktank instructions.md ./project --dry-run

    # Use a specific model with verbose output
    thinktank guide.md main.go --model gpt-4.1 --verbose

    # Force synthesis mode for comprehensive analysis
    thinktank analysis.txt ./complex-code --synthesis

    # Quiet mode for scripts
    thinktank task.md ./src --quiet --output-dir ./results

ENVIRONMENT VARIABLES:
    GEMINI_API_KEY         API key for Google Gemini models
    OPENAI_API_KEY         API key for OpenAI models (GPT-4, etc.)
    OPENROUTER_API_KEY     API key for OpenRouter models

    At least one API key must be set. The tool will automatically
    select compatible models based on available API keys.

FILE FORMATS:
    Supports most text-based file formats including:
    - Source code: .go, .py, .js, .ts, .java, .cpp, etc.
    - Documentation: .md, .txt, .rst
    - Configuration: .json, .yaml, .toml, .xml
    - Web: .html, .css

    Binary files and common build artifacts are automatically excluded.

TROUBLESHOOTING:
    "API key not set" error:
        Set the appropriate environment variable for your chosen provider.
        Example: export GEMINI_API_KEY="your-key-here"

    "Target path not found" error:
        Verify the file or directory exists and you have read permissions.
        Use 'ls' to confirm the path is correct.

    "Model not available" error:
        The selected model requires an API key that isn't configured.
        Use --dry-run to see which models are available.

    High token count warnings:
        Large codebases may exceed model context limits.
        Consider analyzing specific subdirectories or using a model
        with a larger context window.

MORE INFORMATION:
    Documentation: https://github.com/phrazzld/thinktank
    Report issues: https://github.com/phrazzld/thinktank/issues
`
