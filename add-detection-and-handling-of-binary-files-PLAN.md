# Add detection and handling of binary files

## Task Goal
Implement logic to detect binary files during context reading and skip them (with appropriate warnings) to prevent inclusion of non-text content in LLM prompts.

## Chosen Implementation Approach
1. Create a utility function `isBinaryFile` in fileReader.ts that will:
   - Examine a small sample (first few KB) of a file's content
   - Use statistical analysis of byte values to determine if the file is likely binary
   - Return a boolean indicating whether the file is binary

2. Modify the `readContextFile` function to:
   - Call the binary detection function after reading the file
   - If binary is detected, return a specific error code and message
   - Skip adding binary file content to the result

3. Add a warning log when binary files are encountered

## Reasoning for Approach
I considered three potential approaches:

1. **Content analysis approach (selected)**: Examining file content to detect binary data by analyzing byte patterns
   - Pros: More accurate than relying solely on extensions, handles edge cases like binary files without extensions
   - Cons: Slightly more expensive operation

2. **File extension filtering**: Maintaining a list of known binary file extensions
   - Pros: Simple, fast, low overhead
   - Cons: Unreliable - file extensions can be misleading or missing

3. **Magic numbers/signatures**: Detecting binary files by looking for specific byte patterns at the start
   - Pros: Effective for many well-known formats
   - Cons: Requires an extensive database of signatures, complex to implement, might miss unusual formats

I selected the content analysis approach because:
- It provides the best balance of accuracy and implementation complexity
- It handles edge cases like binary files without extensions or with misleading extensions
- It's more reliable than extension-based filtering
- It's simpler to implement than a comprehensive magic number/signature system
- It aligns with the project's focus on reliability and robustness

This approach fits well with the existing codebase which already has robust error handling patterns in place to deal with various file reading issues.