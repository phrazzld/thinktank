# Architect Developer Guide

This guide provides detailed information for developers working on the Architect project. It covers the project structure, development workflow, testing approach, and other technical details.

## Project Structure

The project follows a standard Go package structure:

```
/architect/
├── CLAUDE.md          # Basic style guide and commands
├── DEVELOPER.md       # This file
├── LICENSE            # Project license
├── README.md          # User-facing documentation
├── go.mod             # Go module definition
├── go.sum             # Go module dependencies
├── main.go            # Application entry point
├── main_test.go       # Tests for main package
└── internal/          # Internal packages not exportable
    ├── fileutil/      # File handling utilities
    ├── gemini/        # Gemini API client
    ├── logutil/       # Logging infrastructure
    ├── prompt/        # Prompt template management
    │   └── templates/ # Prompt template files
    └── spinner/       # CLI progress indicators
```

### Key Packages

1. **fileutil**
   - Handles file scanning, filtering, and content processing
   - Responsible for determining which files to include in context
   - Processes file content for token counting and formatting

2. **gemini**
   - Provides a clean interface to Google's Gemini generative AI API
   - Manages token counting, content generation, and error handling
   - Uses an interface-based design for easy testing via mocks

3. **logutil**
   - Provides structured logging with different severity levels
   - Supports colored output for better readability
   - Implements a common interface for consistent logging

4. **prompt**
   - Manages prompt templates used for Gemini API requests
   - Loads and processes templates with context variables
   - Supports custom templates for different use cases

5. **spinner**
   - Provides visual progress indicators during long operations
   - Integrates with the logging system for consistent output

## Development Workflow

### Development Environment Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/architect.git
   cd architect
   ```

2. Install required dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   go build
   ```

### Common Tasks

- **Building**: `go build`
- **Running**: `go run main.go --task "TASK DESCRIPTION" PATH/TO/FILES/OR/DIRS`
- **Testing**: `go test ./...`
- **Testing a specific file**: `go test ./PACKAGE_PATH/FILE_test.go`
- **Linting/Formatting**: `go fmt ./...`
- **Static Analysis**: `go vet ./...`
- **Dependency Management**: `go mod tidy`

## Logging System

The application uses a custom logging system implemented in the `logutil` package. This provides structured logging with different severity levels.

### Log Levels

- **Debug**: Detailed information useful for debugging (most verbose)
- **Info**: General information about application operation
- **Warn**: Warning messages that don't prevent normal operation
- **Error**: Error messages that may affect functionality
- **Fatal**: Critical errors that terminate the application

### Usage in Code

```go
// Create a logger
logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[component] ", true)

// Log at different levels
logger.Debug("Detailed debug information: %s", debugInfo)
logger.Info("General information: %s", generalInfo)
logger.Warn("Warning message: %s", warningMessage)
logger.Error("Error occurred: %v", err)
logger.Fatal("Critical error, exiting: %v", criticalErr) // Will call os.Exit(1)
```

### Command-Line Options

Users can control logging behavior with these flags:

- `--verbose`: Set log level to Debug (shorthand for `--log-level=debug`)
- `--log-level`: Set specific log level (debug, info, warn, error)
- `--color`: Enable/disable colored log output

## Testing Approach

The project uses Go's standard testing package with a table-driven approach for most tests. Mock objects are used extensively to isolate units of code during testing.

### Testing Patterns

1. **Table-Driven Tests**: Most tests use table-driven approaches to test multiple scenarios:

   ```go
   tests := []struct {
       name     string
       input    string
       expected int
   }{
       {
           name:     "Empty string",
           input:    "",
           expected: 0,
       },
       // More test cases...
   }
   
   for _, test := range tests {
       t.Run(test.name, func(t *testing.T) {
           result := functionUnderTest(test.input)
           if result != test.expected {
               t.Errorf("Expected %d, got %d", test.expected, result)
           }
       })
   }
   ```

2. **Mock Objects**: Interfaces are used to enable mocking of dependencies:

   ```go
   // Using a mock client
   mockClient := &gemini.MockClient{
       CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
           return &gemini.TokenCount{Total: 15}, nil
       },
   }
   
   // Test with the mock
   result := functionUnderTest(mockClient, "test input")
   ```

3. **Test Helpers**: Common testing functionality is extracted into helper functions:

   ```go
   // Example test helper
   func setupTestEnvironment(t *testing.T) (string, func()) {
       // Create test resources
       
       // Return cleanup function
       return testPath, func() {
           // Clean up resources
       }
   }
   ```

### Testing Files

Each package contains its own test files:

- `fileutil_test.go` - Tests for the fileutil package
- `gemini_test.go` - Tests for the Gemini client
- `logutil_test.go` - Tests for the logging system
- `main_test.go` - Tests for the application's core logic

## Prompt Template System

The application uses a template system for generating prompts for the Gemini API. Templates are written in Go's text/template format and stored in the `internal/prompt/templates` directory.

### Default Template Location

The default template is located at `internal/prompt/templates/default.tmpl`.

### Template Variables

Templates can use the following variables:

- `{{.Task}}`: The task description provided by the user
- `{{.Context}}`: The codebase context gathered from files

### Creating Custom Templates

1. Create a new `.tmpl` file following the Go text/template syntax
2. Reference the template variables as needed
3. Use the `--prompt-template` flag to specify your custom template

Example custom template:
```
You are a senior software developer specializing in {{.Language}}.

**Task:**
{{.Task}}

**Project Context:**
{{.Context}}

Please analyze the code and provide detailed feedback on:
1. Code quality and best practices
2. Potential bugs and edge cases
3. Performance considerations
4. Improvement suggestions
```

## Gemini API Integration

The application integrates with Google's Gemini AI API through a client wrapper in the `internal/gemini` package.

### API Client Structure

1. **Client Interface**: Defines methods for interacting with the API:
   - `GenerateContent`: Generate AI content from a prompt
   - `CountTokens`: Count tokens in a prompt
   - `GetModelInfo`: Get model capabilities and limits
   - `Close`: Release resources

2. **Implementation**: The actual implementation handles API communication details
3. **Mock Implementation**: Used for testing without calling the actual API

### Error Handling

The Gemini client provides enhanced error handling with:
- Detailed error messages from the API
- Categorization of error types (authentication, quota, invalid input, etc.)
- Suggestions for fixing common issues

## File Handling System

The `fileutil` package provides functionality for gathering context from project files.

### Key Components

1. **Config**: Configuration for file processing with filters:
   - `IncludeExts`: File extensions to include
   - `ExcludeExts`: File extensions to exclude
   - `ExcludeNames`: Directories or file names to exclude
   - `Format`: Format string for file content

2. **File Processing Pipeline**:
   - Traverse directories and files
   - Apply filters (extension, name, binary detection, Git ignore)
   - Process and format content
   - Calculate statistics (characters, lines, tokens)

### Customization

File processing can be customized with CLI flags:
- `--include`: Specify which file extensions to include
- `--exclude`: Specify which file extensions to exclude
- `--exclude-names`: Specify directory or file names to exclude
- `--format`: Customize the output format for file content

## Common Development Tasks

### Adding a New Feature

1. Update interfaces if needed (in respective packages)
2. Implement the feature with appropriate tests
3. Update documentation (README.md and DEVELOPER.md)
4. Run tests and linting: `go test ./... && go vet ./...`

### Modifying Prompt Templates

1. Edit existing templates in `internal/prompt/templates/` or add new ones
2. Test the template with example inputs
3. Update documentation if adding new template variables

### Adding Command-Line Flags

1. Add the flag definition in the `parseFlags()` function in `main.go`
2. Update the `Configuration` struct if needed
3. Implement the logic for the new flag
4. Update documentation (README.md)
5. Add tests for the new functionality

## Style Guidelines

See CLAUDE.md for detailed style guidelines, but key points include:

- **Imports**: Group standard library imports first, followed by third-party
- **Formatting**: Use `gofmt` standards (4-space indentation)
- **Error Handling**: Always check errors, log with context
- **Naming**:
  - Functions: camelCase for unexported, PascalCase for exported
  - Variables: descriptive, self-documenting names
  - Packages: short, lowercase, no underscores
- **Types**: Use strong typing, avoid empty interfaces when possible
- **Comments**: Document exported functions and types with godoc style
- **Documentation**: Update README.md when adding new features
- **Testing**: Write tests for new functionality, maintain test coverage

## Debugging Tips

1. **Enable Debug Logging**: Use `--verbose` or `--log-level=debug`
2. **Dry Run Mode**: Use `--dry-run` to see which files would be included without making API calls
3. **Check Token Counts**: Monitor token usage in debug logs for large projects
4. **API Errors**: Check for detailed error messages and suggestions when API calls fail

## Advanced Topics

### Adding a New Package

1. Create a new directory under `internal/`
2. Define interfaces first to enable mock-based testing
3. Implement the package functionality
4. Add tests with high coverage
5. Update this guide with information about the new package

### Modifying Token Calculation

The application uses two token counting approaches:
1. **Estimation**: A simple word-based approach for quick estimates
2. **Accurate Counting**: Uses the Gemini API for precise counts

If you need to modify token calculation:
1. Adjust the `estimateTokenCount` function in `fileutil.go` for the estimation approach
2. For API-based counting, modify the handling in `CalculateStatisticsWithTokenCounting`