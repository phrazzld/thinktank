# Implementation Plan: Update README.md with new usage instructions

## Task
Add clear documentation in README.md about the requirement for --task-file and deprecation of --task.

## Chosen Approach
I'll implement **Approach 2: Prominent Notice + Inline Update** as recommended in the analysis.

This approach adds a dedicated section to highlight the change, in addition to updating the relevant examples and configuration table. It provides high visibility for the change while ensuring that both new and existing users understand the new requirements clearly.

### Implementation Steps:
1. Add a new section titled "Important Update: Task Input Method" after the Overview section
2. Update the Usage section examples to use --task-file instead of --task
3. Modify the Configuration Options table to reflect the changes to --task and --task-file flags

### Code Changes:

**1. Add New Section: "Important Update: Task Input Method"**
```markdown
## Important Update: Task Input Method

**Please note:** The way you provide the task description to Architect has changed:

* The `--task-file` flag is now **required** for generating plans. You must provide the task description in a file. This allows for more complex and structured task inputs.
* The `--task` flag is now **deprecated** and will be removed in a future release. It should no longer be used for generating plans.
* For `--dry-run` operations, neither `--task-file` nor `--task` is strictly required, but the deprecated `--task` flag may still be used during the transition period if needed for context scoping.

Please update your workflows accordingly. See the Usage examples and Configuration Options below for details.
```

**2. Update Usage Section Examples**
```markdown
## Usage

```bash
# Basic usage (Task description in task.txt)
architect --task-file task.txt path/to/your/project

# Example: Create a plan using a task file (e.g., auth_task.txt)
# Contents of auth_task.txt: "Implement JWT-based user authentication and authorization"
architect --task-file auth_task.txt ./

# Specify output file (default is PLAN.md)
architect --task-file task.txt --output auth_plan.md ./

# Include only specific file extensions
architect --task-file task.txt --include .go,.md ./

# Use a different Gemini model
architect --task-file task.txt --model gemini-1.5-pro ./

# Dry run to see which files would be included
architect --dry-run ./
# Or with deprecated --task:
architect --dry-run --task "Task description for context" ./
# Or with --task-file:
architect --dry-run --task-file task.txt ./

# Request confirmation before proceeding if token count exceeds threshold
architect --task-file task.txt --confirm-tokens 25000 ./

# Use a custom prompt template
architect --task-file task.txt --prompt-template custom_template.tmpl ./

# Enable interactive task clarification
architect --task-file task.txt --clarify ./
```
```

**3. Update Configuration Options Table**
```markdown
## Configuration Options

| Flag | Description | Default |
|------|-------------|---------|
| `--task` | (Deprecated) Description of the task. Use --task-file instead. Will be removed in a future version. | `""` |
| `--task-file` | Path to a file containing the task description (Required unless --dry-run is used). | `(Required)` |
| `--output` | Output file path for the generated plan | `PLAN.md` |
...
```

## Reasoning for Choice

I selected Approach 2 (Prominent Notice + Inline Update) for the following reasons:

1. **Clarity for All Users**: By adding a dedicated section for the change, the approach ensures high visibility for both new and existing users. This is critical for a breaking change that affects the primary way the tool is used.

2. **Effective Communication**: The dedicated section clearly explains the new requirement for --task-file, the deprecation of --task, and the exceptions for dry-run mode, leaving no ambiguity about the changes.

3. **Comprehensive Updates**: In addition to the new section, this approach also updates all examples and the configuration table, ensuring consistency throughout the documentation.

4. **Support Reduction**: Clear, upfront communication about the breaking change will minimize confusion, errors, and potential support requests from users.

5. **Industry Best Practice**: Adding prominent notices for significant breaking changes is a standard practice in well-maintained documentation.

While Approach 1 (minimal changes) would be easier to implement, it risks users missing the important change. Approach 3 is better but still might not be prominent enough. Approach 2 provides the right balance of visibility, completeness, and clarity for all users.