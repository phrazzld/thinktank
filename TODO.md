# TODO

## XDG Directory Support
- [x] Create XDG path resolution utility function
  - Description: Implement a reusable function to determine the correct configuration directory path based on platform
  - Dependencies: None
  - Priority: High

- [x] Update configuration path constants
  - Description: Remove CONFIG_SEARCH_PATHS and replace with the new XDG-based approach
  - Dependencies: XDG path resolution function
  - Priority: High

## Configuration Loading
- [x] Refactor loadConfig function
  - Description: Simplify to use a single path with proper error handling and default creation
  - Dependencies: XDG path resolution function
  - Priority: High

- [x] Remove configuration merging logic
  - Description: Remove mergeConfigs and related functionality from configManager.ts
  - Dependencies: Refactored loadConfig
  - Priority: Medium

- [x] Create minimal default configuration template
  - Description: Define the initial configuration structure created when no file exists
  - Dependencies: None
  - Priority: Medium

## Configuration Saving
- [x] Update saveConfig function
  - Description: Ensure it uses the canonical path and properly creates directories
  - Dependencies: XDG path resolution function
  - Priority: High

- [x] Implement improved error handling
  - Description: Add detailed error reporting for file system operations and validation failures
  - Dependencies: Updated saveConfig
  - Priority: Medium

## CLI Command Updates
- [x] Update the 'config path' command
  - Description: Modify to display the single canonical configuration path
  - Dependencies: XDG path resolution function
  - Priority: Medium

- [x] Update the 'config show' command
  - Description: Remove mergeWithDefaults option and update to use new loadConfig
  - Dependencies: Refactored loadConfig
  - Priority: Medium

- [x] Update model management commands
  - Description: Modify add/remove/enable/disable to use the new configuration approach
  - Dependencies: Updated loadConfig and saveConfig
  - Priority: Medium

- [x] Update group management commands
  - Description: Modify group creation/editing commands to use the new configuration approach
  - Dependencies: Updated loadConfig and saveConfig
  - Priority: Medium

## Testing
- [x] Write tests for XDG path resolution
  - Description: Create comprehensive tests for path determination across platforms
  - Dependencies: XDG path resolution function
  - Priority: High

- [x] Update configManager tests
  - Description: Modify existing tests to work with the new simplified approach
  - Dependencies: Refactored loadConfig and saveConfig
  - Priority: High

- [x] Test CLI commands with new configuration system
  - Description: Ensure all CLI commands work correctly with the new path resolution
  - Dependencies: Updated CLI commands
  - Priority: Medium

## Documentation
- [ ] Update README with new configuration location
  - Description: Document the canonical configuration location for users
  - Dependencies: Implemented XDG path resolution
  - Priority: Low

- [ ] Add migration notes
  - Description: Provide guidance for users with existing configurations
  - Dependencies: Complete implementation
  - Priority: Low

## Platform-Specific Considerations
- [x] Ensure Windows support is robust
  - Description: Test directory creation and permissions on Windows platform
  - Dependencies: XDG path resolution function
  - Priority: Medium

- [x] Ensure macOS support is robust
  - Description: Test with both standard macOS and XDG paths
  - Dependencies: XDG path resolution function
  - Priority: Medium