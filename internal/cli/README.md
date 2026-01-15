# internal/cli

Command-line interface layer.

## Overview

Handles argument parsing, configuration, error presentation, and exit codes. Delegates all business logic to `internal/thinktank`.

## Key Components

| File | Purpose |
|------|---------|
| `main.go` | Entry point: Main() function |
| `simple_parser.go` | Argument parsing (simplified interface) |
| `simple_config.go` | Configuration building from args |
| `errors.go` | Error formatting and exit code mapping |
| `help.go` | Help text generation |
| `output_manager.go` | Console output coordination |

## Design

The CLI uses a "simplified interface" pattern:
```bash
thinktank instructions.txt target_path... [flags]
```

All flags are optional with smart defaults. Model selection is automatic based on input size.

## Entry Point

```go
// cmd/thinktank/main.go calls this
func Main() {
    // Parse args → validate → execute → handle errors → exit
}
```

## Testing

- `simple_parser_test.go` - Argument parsing edge cases
- `main_test.go` - Integration tests with mocked execution
