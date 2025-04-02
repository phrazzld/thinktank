# TODO

## Dependencies
- [x] Add XDG Library Dependency
  - Description: Add GitHub's adrg/xdg library to handle XDG Base Directory Specification
  - Dependencies: None
  - Priority: High

## Configuration Structure
- [x] Define Configuration Data Structure
  - Description: Create AppConfig struct to hold configuration values from files/flags
  - Dependencies: None
  - Priority: High
- [x] Create Configuration Logic
  - Description: Implement loading logic with precedence (user → system → defaults)
  - Dependencies: XDG Library, Configuration Structure
  - Priority: High

## Template Management
- [x] Implement Embedded Templates
  - Description: Use Go embed package to bundle default templates within binary
  - Dependencies: None
  - Priority: High
- [x] Refactor PromptManager
  - Description: Update to search templates in user, system, and embedded locations
  - Dependencies: XDG Library, Embedded Templates
  - Priority: High

## Integration
- [ ] Update Main Package
  - Description: Load configuration early in startup and pass to components
  - Dependencies: Configuration Logic, Refactored PromptManager
  - Priority: Medium
- [ ] Ensure Backwards Compatibility
  - Description: Verify command-line flags override config file values correctly
  - Dependencies: Configuration Logic, Updated Main Package
  - Priority: Medium

## Testing
- [ ] Write Unit Tests
  - Description: Test config loading, template finding, and merging logic
  - Dependencies: All implementation tasks
  - Priority: Medium
- [ ] Add Integration Tests
  - Description: Test full application with various config locations and scenarios
  - Dependencies: All implementation tasks
  - Priority: Medium
- [ ] Manual Testing
  - Description: Test on different OSes and confirm expected behavior
  - Dependencies: All implementation tasks
  - Priority: Low

## Documentation
- [ ] Update Project Documentation
  - Description: Add configuration file format and locations to README/docs
  - Dependencies: Implementation tasks
  - Priority: Low

## Assumptions
- We'll use Viper or similar library for config file handling
- TOML is the preferred config format
- Command-line flags will override config file values
- User-specific configs will override system-wide configs