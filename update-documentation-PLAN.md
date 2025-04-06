# Update Documentation

## Goal
Update documentation to reflect the XDG-compliant configuration approach and remove any references to thinktank.config.json in the project root.

## Implementation Approach
The project has moved to using XDG-compliant configuration file paths but documentation may still reference the old approach of having thinktank.config.json in the project root. I'll implement this task with the following steps:

1. Scan all markdown documentation files (README.md, CONTRIBUTING.md, etc.) to identify references to configuration file paths
2. Update references to use the new XDG-compliant paths, which store configuration in:
   - Linux: ~/.config/thinktank/thinktank.config.json
   - macOS: ~/Library/Application Support/thinktank/thinktank.config.json
   - Windows: %APPDATA%\thinktank\thinktank.config.json
3. Add clear documentation about the configuration system, including:
   - How to locate the configuration file
   - How to edit configuration
   - The precedence order (CLI args > env vars > config file > defaults)
4. Ensure all code examples in documentation use the correct path references

## Reasoning
This approach is preferred because:

1. **Accuracy**: Documentation must accurately reflect the actual implementation to prevent user confusion.
2. **Discoverability**: Clear documentation on where to find and how to modify configuration files improves user experience.
3. **Consistency**: All documentation should consistently reference the same configuration approach.
4. **Standards Compliance**: Following the XDG Base Directory Specification is a best practice for configuration management.

Alternative approaches considered:
- Implementing a compatibility layer that would allow both old and new configuration paths (rejected due to maintenance overhead and potential confusion)
- Creating symbolic links from the old location to the new one (rejected as it goes against the purpose of using XDG standard)