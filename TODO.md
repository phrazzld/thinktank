# TODO

## Setup & Structure
- [x] Create Feature Branch
  - Description: Create `refactor/simplify-architecture` branch (using simplify-cli)
  - Dependencies: None
  - Priority: High

- [x] Establish New Directory Structure
  - Description: Create src/core, src/providers, src/cli, src/utils, src/workflow directories
  - Dependencies: Feature branch created
  - Priority: High

- [x] Move Existing Files to New Structure
  - Description: Relocate files from atomic design structure to new locations
  - Dependencies: New directory structure created
  - Priority: High

- [x] Update Import Paths
  - Description: Fix all import paths throughout the project to match new structure
  - Dependencies: Files moved to new structure
  - Priority: High
  - Note: Main code files updated and build is successful; some test files still have failures

- [ ] Remove Old Atomic Design Structure
  - Description: Delete the old atoms/molecules/organisms/templates directories after ensuring all functionality is migrated
  - Dependencies: All components migrated to new structure, all tests passing with new structure
  - Priority: Medium

## Core Components
- [x] Refine ConfigManager
  - Description: Implement/update ConfigManager with load, save, add, remove, update methods for models/groups
  - Dependencies: New directory structure
  - Priority: High

- [x] Implement Cascading Configuration System
  - Description: Create resolveModelOptions function with proper hierarchy (base defaults → provider defaults → model defaults → user config → group overrides → CLI overrides)
  - Dependencies: ConfigManager implementation
  - Priority: High

- [x] Update LLMRegistry
  - Description: Adapt LLMRegistry to work with the new structure and ConfigManager
  - Dependencies: ConfigManager implementation
  - Priority: Medium

## CLI Commands
- [x] Update Commander.js Setup
  - Description: Install/update commander.js and set up basic CLI structure
  - Dependencies: New directory structure
  - Priority: High
  - Note: Implemented main CLI structure with commander.js, created commands for run, models, and config with subcommands

- [x] Implement Config Commands
  - Description: Create thinktank config command suite (path, show, models/groups management)
  - Dependencies: ConfigManager, Commander.js setup
  - Priority: High
  - Note: Implemented full config command suite with path, show, and models/groups management commands

- [x] Implement Run Command with Model Selection
  - Description: Add --models flag for direct model specification in run command
  - Dependencies: ConfigManager, Commander.js setup
  - Priority: High
  - Note: Flag implemented with robust validation, error handling, and compatibility with group filtering

## Workflow Refactoring
- [x] Create InputHandler Module
  - Description: Implement module to load prompt from file/stdin and config
  - Dependencies: New directory structure
  - Priority: Medium
  - Note: Implemented robust input handling with support for files, stdin, and direct text

- [x] Create ModelSelector Module
  - Description: Implement logic to determine models to query based on CLI flags and config
  - Dependencies: ConfigManager
  - Priority: Medium
  - Note: Implemented robust model selection with support for multiple models, groups, API key validation, and detailed error handling

- [x] Create QueryExecutor Module
  - Description: Implement parallel API call handling with proper error management
  - Dependencies: LLMRegistry updates
  - Priority: Medium
  - Note: Implemented robust parallel API call processing with status tracking, timeout handling, and detailed error management

- [x] Create OutputHandler Module
  - Description: Implement result formatting and writing to files/console
  - Dependencies: QueryExecutor module
  - Priority: Medium
  - Note: Implemented robust file output and console formatting with comprehensive error handling and detailed status tracking

## Provider Updates
- [ ] Update Provider Implementations
  - Description: Modify providers to accept resolved ModelOptions and remove internal defaulting
  - Dependencies: Cascading configuration system
  - Priority: Medium

## User Experience
- [ ] Implement Preset Feature
  - Description: Add --preset flag for predefined model sets
  - Dependencies: ModelSelector module
  - Priority: Low

- [ ] Refine Logging System
  - Description: Create logger utility with verbosity control
  - Dependencies: None
  - Priority: Medium

- [ ] Enhance Error Handling
  - Description: Improve error messages throughout the workflow
  - Dependencies: Workflow modules
  - Priority: Medium

## Testing
- [ ] Update Unit Tests
  - Description: Write/update unit tests for ConfigManager, resolveModelOptions, modelSelector, and utility functions
  - Dependencies: Implementation of respective components
  - Priority: High

- [ ] Add Integration Tests
  - Description: Write tests for config commands and main run command workflow
  - Dependencies: Implementation of CLI commands
  - Priority: Medium

## Documentation
- [ ] Update README.md
  - Description: Reflect new structure, CLI commands, and configuration approach
  - Dependencies: All implementations complete
  - Priority: Medium

- [ ] Create User Guides/Examples
  - Description: Document the cascading configuration hierarchy and usage examples
  - Dependencies: All implementations complete
  - Priority: Low

## Final Steps
- [ ] Perform Code Review
  - Description: Review all components for quality, consistency, and completeness
  - Dependencies: All implementations complete
  - Priority: High

- [ ] Merge to Main
  - Description: Merge the feature branch after successful testing and review
  - Dependencies: All tasks completed, tests passing
  - Priority: High