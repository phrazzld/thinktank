# TODO

## Core Types Implementation
- [x] Define `SystemPrompt` interface
  - Description: Create interface for system prompts with text and metadata fields
  - Dependencies: None
  - Priority: High

- [x] Define `ModelGroup` interface
  - Description: Create interface for model groups with system prompts and model lists
  - Dependencies: SystemPrompt interface
  - Priority: High

- [x] Update `AppConfig` interface
  - Description: Expand to support both legacy models array and groups object
  - Dependencies: ModelGroup interface
  - Priority: High

- [x] Update `ModelConfig` interface
  - Description: Add optional systemPrompt property for per-model prompts
  - Dependencies: SystemPrompt interface
  - Priority: High

- [ ] Update `LLMResponse` interface
  - Description: Include group information in response objects
  - Dependencies: ModelGroup interface
  - Priority: Medium

- [ ] Update `ModelOptions` interface
  - Description: Add support for system prompt related options
  - Dependencies: None
  - Priority: Medium

## Configuration Management
- [x] Create Zod schema for `SystemPrompt`
  - Description: Define validation schema for system prompts
  - Dependencies: SystemPrompt interface
  - Priority: High

- [x] Create Zod schema for `ModelGroup`
  - Description: Define validation schema for model groups
  - Dependencies: ModelGroup interface, SystemPrompt schema
  - Priority: High

- [x] Update `appConfigSchema`
  - Description: Modify to support both models array and groups
  - Dependencies: ModelGroup schema
  - Priority: High

- [x] Implement `getGroup` function
  - Description: Create utility to get models from a specific group
  - Dependencies: Updated AppConfig interface
  - Priority: High

- [x] Implement `getEnabledGroupModels` function
  - Description: Create utility to get enabled models from a group
  - Dependencies: getGroup function
  - Priority: High

- [ ] Implement `filterGroupModels` function
  - Description: Create utility to filter models within a group
  - Dependencies: getGroup function
  - Priority: Medium

- [x] Refactor `mergeConfigs` function
  - Description: Update to handle groups alongside models
  - Dependencies: Updated AppConfig interface
  - Priority: High

- [ ] Update `loadConfig` function
  - Description: Normalize legacy configs to include a default group
  - Dependencies: Updated AppConfig schema
  - Priority: High

## Run Workflow Updates
- [ ] Update `RunOptions` interface
  - Description: Add groups parameter and systemPrompt parameter
  - Dependencies: None
  - Priority: High

- [ ] Refactor main workflow to handle multiple groups
  - Description: Modify runThinktank to process models across groups
  - Dependencies: Updated RunOptions interface
  - Priority: High

- [ ] Implement system prompt application
  - Description: Apply appropriate prompts to each query
  - Dependencies: Updated ModelConfig interface
  - Priority: High

- [ ] Add group-based result tracking
  - Description: Track and report results organized by group
  - Dependencies: Updated LLMResponse interface
  - Priority: Medium

- [ ] Implement group-based output organization
  - Description: Ensure output files are organized by group
  - Dependencies: Group-based result tracking
  - Priority: Medium

## CLI Interface Updates
- [ ] Add `--group/-g` parameter
  - Description: Implement CLI option to specify groups to run
  - Dependencies: Updated RunOptions interface
  - Priority: High

- [ ] Add `--system-prompt/-s` parameter
  - Description: Implement CLI option for system prompt override
  - Dependencies: Updated RunOptions interface
  - Priority: High

- [ ] Update CLI help documentation
  - Description: Add information about new parameters and examples
  - Dependencies: New CLI parameters
  - Priority: Medium

- [ ] Update default command handler
  - Description: Modify to handle group-based execution
  - Dependencies: New CLI parameters
  - Priority: High

- [ ] Update list-models command
  - Description: Modify to show group organization
  - Dependencies: Updated configuration manager
  - Priority: Medium

## Output Formatter Updates
- [ ] Add group information to output
  - Description: Update formatter to include group details
  - Dependencies: Updated LLMResponse interface
  - Priority: Medium

- [ ] Implement group-based result aggregation
  - Description: Support grouping results by group in output
  - Dependencies: Group information in output
  - Priority: Medium

- [ ] Ensure backward compatibility
  - Description: Maintain support for non-grouped results
  - Dependencies: None
  - Priority: High

## Provider Integration
- [ ] Update provider interfaces
  - Description: Modify to support system prompts
  - Dependencies: SystemPrompt interface
  - Priority: High

- [ ] Update provider generate methods
  - Description: Modify to accept and apply system prompts
  - Dependencies: Updated provider interfaces
  - Priority: High

## Testing
- [ ] Unit tests for new types and schemas
  - Description: Test validation of new interfaces and schemas
  - Dependencies: Core type implementations
  - Priority: High

- [ ] Unit tests for configuration handling
  - Description: Test group configuration loading and validation
  - Dependencies: Updated configuration manager
  - Priority: High

- [ ] Unit tests for CLI parameters
  - Description: Test handling of new CLI options
  - Dependencies: CLI interface updates
  - Priority: Medium

- [ ] Integration tests for group functionality
  - Description: Test end-to-end workflow with groups
  - Dependencies: All major updates
  - Priority: Medium

- [ ] Backward compatibility tests
  - Description: Verify legacy configs and commands still work
  - Dependencies: All major updates
  - Priority: High

## Documentation
- [ ] Update README with group examples
  - Description: Add documentation for using model groups
  - Dependencies: Implementation completion
  - Priority: Medium

- [ ] Create example configuration files
  - Description: Provide sample configs with groups for reference
  - Dependencies: None
  - Priority: Low