# Context Formatting Strategy in Thinktank

When using Thinktank with context files and directories, the tool organizes the content in a structured format that helps Large Language Models understand and work with multiple sources effectively.

## How Context Files Are Formatted

The formatting strategy follows a clear approach that:

1. Separates context documents from the main prompt
2. Uses markdown headers to create visual and semantic divisions
3. Preserves file paths and applies appropriate syntax highlighting
4. Creates a clean, readable structure for both machines and humans

### Example 

Here's how a prompt with multiple context files would be formatted:

```markdown
# CONTEXT DOCUMENTS

## File: /path/to/code/example.js
```javascript
function calculateTotal(items) {
  return items.reduce((sum, item) => sum + item.price, 0);
}

module.exports = { calculateTotal };
```

## File: /path/to/data/config.json
```json
{
  "maxItems": 100,
  "taxation": {
    "rate": 0.07,
    "exempt": ["food", "medicine"]
  }
}
```

# USER PROMPT

Can you explain how the calculation function works and how it would interact with the tax exemption rules in the config file?
```

## Technical Implementation Details

The context formatting is performed by the `formatCombinedInput` function in `src/utils/fileReader.ts`. This function:

1. Filters out any context files with errors or null content
2. Creates a "CONTEXT DOCUMENTS" section header
3. For each context file:
   - Adds a subheading with the normalized file path
   - Determines the appropriate language for syntax highlighting based on file extension
   - Wraps the file content in markdown code blocks with language identifiers
4. Adds a "USER PROMPT" section header
5. Appends the original prompt text

### Language Detection

The system automatically determines the appropriate language for syntax highlighting based on the file extension. It supports a wide range of common programming languages and file formats including:

- JavaScript/TypeScript (js, jsx, ts, tsx)
- HTML/XML (html, xml)
- JSON/YAML (json, yaml, yml)
- CSS/SCSS (css, scss)
- Python (py)
- Java/Kotlin (java, kt)
- C/C++/C# (c, cpp, cs)
- Shell scripts (sh, bash)
- Markdown (md)
- Many other common formats

If the file extension is not recognized, it defaults to plain text formatting.

## Benefits of This Approach

1. **Clear Separation**: Distinct sections make it easy for LLMs to distinguish between reference material and the user's request
2. **Proper Code Formatting**: Syntax highlighting ensures code maintains its structure and readability
3. **Context Awareness**: File paths help the model understand the relationship between different files
4. **Error Handling**: Files with errors are excluded to avoid confusing the model
5. **Consistent Structure**: The uniform format ensures LLMs can reliably parse the input regardless of the number or type of context files

## Best Practices for Using Context

1. **Include Only Relevant Files**: Too many files can exceed model context limits and dilute focus
2. **Use Directories Wisely**: Include directories when the full codebase context is needed
3. **Prioritize Key Files**: Place the most important context files first in your command
4. **Check File Sizes**: Very large files might be truncated - consider breaking them up if needed
5. **Be Specific in Your Prompt**: Reference the included files by name in your prompt to direct the model's attention

This formatting approach ensures that Thinktank can effectively provide LLMs with the context they need to answer complex queries accurately, whether you're analyzing code, working with configuration files, or explaining technical documentation.