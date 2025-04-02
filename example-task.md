# Implement XDG-compliant Configuration Management

## Problem Statement
The architect tool currently uses hardcoded relative paths for template files and other configurations. This approach fails when the tool is run from directories other than the project root, as evidenced by errors like:

```
Failed to build prompt: failed to load prompt template: template file not found: stat default.tmpl: no such file or directory
```

## Task Requirements
Refactor architect to use XDG Base Directory Specification for configuration management:

1. Implement a robust configuration system following XDG Base Directory Specification
2. Store default templates and configuration files in standardized locations
3. Support the following lookup precedence:
   - User-specified paths via command line arguments
   - User-specific configuration directory (~/.config/architect/)
   - System-wide configuration directory (/etc/architect/)
   - Built-in defaults (embedded in binary)
4. Ensure backward compatibility with existing command-line arguments
5. Add documentation describing configuration file locations and customization options

## Expected Results
- The tool should work correctly when run from any directory
- Users should be able to customize templates by placing them in their ~/.config/architect/ directory
- The tool should gracefully fall back to system-wide configs or built-in defaults when custom configs are unavailable
- Implementation should follow Go best practices for configuration management
