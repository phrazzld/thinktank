# Add max file size limit checks

## Task Goal
Implement file size limit checks to prevent reading overly large files during context gathering, which could cause memory issues or exceed token limits for LLMs.

## Chosen Implementation Approach
1. Define a configurable `MAX_FILE_SIZE` constant in the fileReader module (defaulting to 10MB)
2. Add a check in the `readContextFile` function that:
   - Checks the file size before reading the content
   - If the file exceeds the size limit, returns a specific error code and message
   - Logs a warning about the skipped file
3. Update relevant tests to verify this behavior

The implementation will:
- Use the existing `stats` object (already retrieved in `readContextFile`) to get the file size
- Place the size check after verifying the path is a file but before reading content
- Follow the established error handling pattern with specific error codes
- Provide a clear, informative error message including the file size and limit

## Reasoning for Approach
I considered several approaches:

1. **Pre-read size check (selected)**: Checking file size before reading content
   - Pros: Prevents loading large files into memory, efficient, predictable
   - Cons: Requires an additional file system stat operation (though this is already being done in our implementation)

2. **Post-read truncation**: Reading the file but truncating content if it exceeds a limit
   - Pros: Would allow partial content from large files
   - Cons: Still loads the entire file into memory, which could cause issues with very large files

3. **Streaming approach**: Reading files in chunks and stopping if a limit is reached
   - Pros: Efficient memory usage even with large files
   - Cons: More complex implementation, doesn't align with the current non-streaming file reading approach

I selected the pre-read size check approach because:
- It's the most protective against memory issues
- It aligns with the project's emphasis on reliability and error handling
- It's consistent with how we handle other file errors (early detection and reporting)
- It's simple and maintainable
- It doesn't require changing the fundamental approach to file reading
- It follows the pattern established with binary file detection
- It prevents wasting resources by trying to read files that would ultimately be too large for the LLM context window anyway

This approach will effectively filter out excessively large files, providing clear feedback to users while maintaining system stability.