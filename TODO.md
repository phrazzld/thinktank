# TODO

## CLI Parser Restructuring

- [x] Simplify CLI Argument Parsing
  - Description: Rewrite cli.ts to use positional arguments instead of flags/options
  - Dependencies: None
  - Priority: High

- [x] Implement Smart Command Detection
  - Description: Add logic to detect and handle the three command patterns (models, prompt+group, prompt+model)
  - Dependencies: Simplified CLI Argument Parsing
  - Priority: High
  
- [x] Add Group Validation
  - Description: Check if specified group exists in config and show available groups if not
  - Dependencies: Simplified CLI Argument Parsing
  - Priority: High

- [x] Add Specific Model Parsing
  - Description: Add logic to detect provider:model format and extract provider/model values
  - Dependencies: Simplified CLI Argument Parsing
  - Priority: High

- [ ] Improve Error Messages
  - Description: Create clear, actionable error messages for common CLI errors
  - Dependencies: Smart Command Detection
  - Priority: Medium

## Configuration Updates

- [x] Add Default Group to Config Schema
  - Description: Update config schema to include defaultGroup property
  - Dependencies: None
  - Priority: High

- [x] Create Default Configuration
  - Description: Update the default config template with sensible groups and default group
  - Dependencies: Config Schema Update
  - Priority: Medium

- [x] Add Default Group Selection Logic
  - Description: Implement logic to select default group if none is specified
  - Dependencies: Config Schema Update
  - Priority: High

## runThinktank Simplification

- [x] Update RunOptions Interface
  - Description: Simplify the options interface to support both group and specific model use cases
  - Dependencies: None
  - Priority: High

- [x] Implement Model Selection Logic
  - Description: Add logic to select models based on group name or specific model identifier
  - Dependencies: RunOptions Interface Update
  - Priority: High

- [x] Implement System Prompt Selection Logic
  - Description: Select appropriate system prompt based on group or find the group a specific model belongs to
  - Dependencies: Model Selection Logic
  - Priority: Medium

- [x] Improve Progress Display
  - Description: Update spinner and progress messages to show per-model status
  - Dependencies: Model Selection Logic
  - Priority: Medium

- [x] Create Model List Display
  - Description: Display a numbered list of models being processed
  - Dependencies: Model Selection Logic
  - Priority: Medium

- [x] Simplify Output Directory Structure
  - Description: Create a more intuitive output directory structure based on group or model name
  - Dependencies: Model Selection Logic
  - Priority: Medium

## Error Handling Improvements

- [x] Enhance File Not Found Errors
  - Description: Provide clear error messages when prompt file is not found
  - Dependencies: None
  - Priority: Medium

- [x] Enhance Provider/Model Errors
  - Description: Provide helpful error messages when provider or model is invalid
  - Dependencies: Specific Model Parsing
  - Priority: Medium

- [ ] Enhance Missing API Key Errors
  - Description: Improve error messages for missing API keys with instructions on how to set them
  - Dependencies: None
  - Priority: Medium

## Testing

- [x] Test Group Use Case
  - Description: Create tests for running prompts through model groups
  - Dependencies: RunOptions Interface Update, Model Selection Logic
  - Priority: High

- [x] Test Specific Model Use Case
  - Description: Create tests for running prompts through specific models
  - Dependencies: Specific Model Parsing, Model Selection Logic
  - Priority: High

- [ ] Test Models Listing
  - Description: Create tests for the models listing command
  - Dependencies: Smart Command Detection
  - Priority: Medium

- [ ] Test Error Handling
  - Description: Create tests for various error scenarios
  - Dependencies: Error Handling Improvements
  - Priority: Medium

## Documentation

- [ ] Update Usage Documentation
  - Description: Update README with new command structure and examples
  - Dependencies: All implementation tasks
  - Priority: Low

- [ ] Add Configuration Examples
  - Description: Document the new configuration options with examples
  - Dependencies: Configuration Updates
  - Priority: Low