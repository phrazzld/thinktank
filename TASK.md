# Task: Reimagine Configuration Management for Thinktank (COMPLETED)

## Context
Thinktank's current configuration approach uses multiple search locations and prioritization logic that adds unnecessary complexity. As the sole user at this stage, we have the freedom to completely rethink this system without concerns for backward compatibility.

## Goal
Create a clean, simple, and intuitive configuration system based on modern best practices. The configuration should live in a single, predictable location in the user's home directory (following XDG spec for Unix systems), eliminating confusion and simplifying both code and user experience.

## Completed Work

### Phase 1: Investigation and Planning
- Analyzed existing configuration system and identified pain points
- Researched XDG Base Directory Specification and how other tools implement it
- Created a detailed implementation plan with prioritized tasks

### Phase 2: Core Implementation
- Added XDG path resolution utility functions in `fileReader.ts`
  - Implemented platform-specific path detection (Windows, macOS, Linux)
  - Added proper directory creation and error handling
- Updated configuration path constants in `constants.ts`
  - Removed the complex CONFIG_SEARCH_PATHS array
  - Simplified to use a single default template path
- Refactored `loadConfig` function in `configManager.ts`
  - Simplified to use XDG paths by default
  - Improved error handling with specific error messages
  - Maintained support for custom paths via `--config` option
- Enhanced `saveConfig` function
  - Added validation before saving to prevent corruption
  - Improved error handling for file system errors
  - Added detailed logging for better debugging

### Phase 3: CLI Updates
- Updated configuration commands to use the new XDG-based approach
  - Enhanced 'config path' command to show canonical location
  - Updated 'config show' command to use refactored functions
  - Simplified how configuration paths are communicated to users
- Updated model management commands
  - Refactored all model commands (add, remove, enable, disable)
  - Created helper function `displayConfigSavePath` for consistency
  - Ensured all commands properly use the new XDG-based configuration
- Updated group management commands
  - Adapted all group commands to work with the new system
  - Maintained backward compatibility for project-local configurations
  - Improved error handling and user feedback

### Phase 4: Testing and Validation
- Created comprehensive test coverage
  - Unit tests for XDG path resolution
  - Tests for configuration loading and saving
  - Added `empty-config.test.ts` to verify proper initialization
- Verified build and lint processes pass
- Confirmed all commands work as expected

## Benefits Achieved
- Simplified codebase by removing complex config path search logic
- Improved user experience with predictable, standards-compliant config locations
- Enhanced error handling with clear, actionable messages
- Maintained backward compatibility while modernizing the approach
- Added detailed logging for easier troubleshooting
- Standardized how config paths are displayed to users

## Platform Support
- Windows: `%APPDATA%\thinktank\config.json`
- macOS: `~/Library/Preferences/thinktank/config.json`
- Linux/Unix: `~/.config/thinktank/config.json` (or `$XDG_CONFIG_HOME/thinktank/config.json` if set)

The configuration system now follows modern best practices while being simpler to understand and maintain.