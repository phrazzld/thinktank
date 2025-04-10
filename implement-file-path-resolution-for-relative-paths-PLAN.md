# Implement file path resolution for relative paths

## Goal
Add logic to the application to properly resolve relative file paths to appropriate locations using XDG standards, ensuring consistent file handling across different environments and platforms.

## Implementation Approach
I'll create a centralized path resolution system in the application with the following components:

1. Create a new utility function called `resolvePath` that will:
   - Accept a relative or absolute path and a path type (e.g., "log", "config", "output")
   - Check if the path is already absolute
   - If relative, resolve it based on the path type using XDG standards
   - For log files: use XDG_CACHE_HOME
   - For config files: use XDG_CONFIG_HOME
   - For output files: resolve relative to the current working directory
   - Return the fully resolved absolute path

2. Modify the `initAuditLogger` function to use this utility for resolving the audit log file path

3. Update the `saveToFile` function to use this utility instead of its current path resolution logic

4. Apply the same pattern to `readTaskFromFile` and any other functions that handle file paths

This centralized approach will ensure consistent path handling across the application and make it easier to maintain and test path resolution logic.

## Key Design Decisions

1. **Centralized Resolution Logic**: By creating a single function for path resolution, we ensure consistent path handling throughout the application and avoid duplication of logic.

2. **Path Type-Based Resolution**: Different types of files should be stored in different locations according to XDG standards:
   - Log files → XDG_CACHE_HOME (~/.cache/architect/...)
   - Config files → XDG_CONFIG_HOME (~/.config/architect/...)
   - Output files → Current working directory (unless specified as absolute)

3. **Fallback Strategy**: If the XDG directories can't be determined or accessed, we'll fall back to sensible defaults:
   - For logs: ./logs/
   - For config: ./config/
   - For output: ./

4. **Platform Compatibility**: The implementation will work correctly on different operating systems by using Go's platform-independent path handling functions.

This approach aligns with the existing code patterns in the application, such as the `getCacheDir` function, while providing more robust and consistent path handling across the entire application.