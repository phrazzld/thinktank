# Create formatCombinedInput function

## Task Goal
Implement a function that combines prompt content with context files in a format that helps LLMs understand the boundaries between different context files and the main prompt, enabling more effective use of the context information.

## Chosen Implementation Approach
I'll create a `formatCombinedInput` function that:

1. Takes two parameters:
   - `promptContent`: The main prompt text (string)
   - `contextFiles`: An array of ContextFileResult objects from the readContextPaths function

2. Formats the combined content using markdown with clear section headers:
   - Wraps each context file in a markdown code block with the file path as the heading
   - Places context files before the main prompt
   - Separates context files with clear boundaries
   - Adds a distinct separator between all context files and the main prompt
   - Only includes successful context files (those without errors)
   - Maintains file paths in relative format (when possible) for clarity

3. The output format will be:
   ```
   # CONTEXT DOCUMENTS
   
   ## File: path/to/file1.js
   ```javascript
   // file1.js content here
   ```
   
   ## File: path/to/file2.md
   ```markdown
   # File2 markdown content
   ```
   
   # USER PROMPT
   
   The original prompt content here...
   ```

## Reasoning for Approach
I considered three potential approaches:

1. **Markdown-based formatting with clear sections (selected)**: Using markdown to structure the content with explicit sections.
   - Pros: Clear boundaries, language-specific code formatting, hierarchical structure, human-readable
   - Cons: Slightly more verbose than minimal approaches

2. **Simple plain text with minimal separators**: Using basic text separators.
   - Pros: Minimal overhead, simple implementation
   - Cons: Less structured, harder for LLMs to distinguish context boundaries, no syntax highlighting cues

3. **JSON/structured data format**: Using a structured format like JSON.
   - Pros: Precise machine-readable boundaries
   - Cons: Less efficient use of token space, less natural for LLMs to process

I selected the markdown-based approach because:
- It provides clear visual and semantic boundaries that help LLMs understand context
- It leverages code blocks with language hints which may help LLMs understand file content better
- It's human-readable, making debugging easier
- It uses token space efficiently while maintaining clarity
- It maintains hierarchical relationships between content
- It's a format that LLMs are very familiar with (trained on large volumes of markdown)
- The structure matches common prompt engineering best practices