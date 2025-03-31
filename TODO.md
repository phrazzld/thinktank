# TODO

## Output Directory Implementation
- [x] Define output directory strategy 
  - Description: Modify existing `--output` option to specify parent directory for run-specific output folder
  - Dependencies: None
  - Priority: High

- [x] Create output directory generator 
  - Description: Implement utility for timestamp-based directory name generation
  - Dependencies: None
  - Priority: High

- [x] Implement directory creation 
  - Description: Add logic in runThinktank.ts to create unique output directory with error handling
  - Dependencies: Output directory strategy
  - Priority: High

- [x] Create filename sanitization utility 
  - Description: Implement function to sanitize provider and model IDs for safe filenames
  - Dependencies: None
  - Priority: High

- [ ] Implement individual file writing 
  - Description: Add logic to write each model response to its own Markdown (.md) file in the output directory
  - Dependencies: Directory creation, filename sanitization
  - Priority: High

- [ ] Enhance console output 
  - Description: Revise console output to show detailed progress info without raw responses
  - Dependencies: Individual file writing
  - Priority: High

- [ ] Add unit tests 
  - Description: Test filename sanitization and directory name generation utilities
  - Dependencies: All implementation tasks
  - Priority: Medium

- [ ] Add integration tests 
  - Description: Test end-to-end functionality of directory creation and file output
  - Dependencies: All implementation tasks
  - Priority: Medium

- [ ] Update documentation 
  - Description: Update README and CLI help with new output directory feature
  - Dependencies: All implementation tasks
  - Priority: Medium

## Assumptions/Questions
- We're modifying the existing `--output` option rather than creating a new `--output-dir` option
- The feature will always be active (no separate flag to enable/disable)
- We will NOT output raw model responses to console, only status/progress information
- Markdown (.md) will be used as the file format for all model outputs