# Error Handling Testing Guide

This guide focuses on testing error handling patterns in the Thinktank project, with particular emphasis on our custom error system.

## Error System Overview

Thinktank uses a structured error handling system with:

- **Error categories**: API, Configuration, FileSystem, etc.
- **Contextual information**: Additional details about the error (file paths, model names, etc.)
- **Recovery suggestions**: Helpful messages for users to resolve the issue
- **Error propagation**: Preserving error chains from low-level to high-level code

## Testing Custom Error Types

When testing custom error types, verify all aspects of the error:

```typescript
it('should create a proper FileSystemError', () => {
  const error = new FileSystemError('File not found', {
    path: '/path/to/file.txt',
    operation: 'read',
    suggestions: ['Check if the file exists']
  });
  
  // Verify error properties
  expect(error.name).toBe('FileSystemError');
  expect(error.message).toBe('File not found');
  expect(error.path).toBe('/path/to/file.txt');
  expect(error.operation).toBe('read');
  expect(error.suggestions).toContain('Check if the file exists');
  expect(error.category).toBe('FILE_SYSTEM');
});
```

## Testing Error Propagation

When testing error propagation, verify that the original error is preserved as the cause:

```typescript
it('should preserve original error as cause', async () => {
  // Create a low-level error
  const fsError = createFsError('ENOENT', 'No such file or directory', 'readFile', 'config.json');
  
  // Spy on fs.readFile to throw this error
  const readFileSpy = jest.spyOn(fs, 'readFile');
  readFileSpy.mockRejectedValueOnce(fsError);
  
  try {
    // Call a higher-level function that wraps the error
    await loadConfiguration('config.json');
    fail('Expected error was not thrown');
  } catch (error) {
    // Verify higher-level error properties
    expect(error).toBeInstanceOf(ConfigError);
    expect(error.message).toContain('Failed to load configuration');
    
    // Verify error chain
    expect(error.cause).toBeDefined();
    expect(error.cause).toBe(fsError);
    expect(error.cause.code).toBe('ENOENT');
  }
  
  readFileSpy.mockRestore();
});
```

## Testing Error Translation

Test how your application translates low-level errors to application-specific errors:

```typescript
it('should translate ENOENT to ResourceNotFoundError', async () => {
  const noentError = createFsError('ENOENT', 'No such file or directory', 'readFile', '/config.json');
  const readFileSpy = jest.spyOn(fs, 'readFile');
  readFileSpy.mockRejectedValueOnce(noentError);
  
  try {
    await yourFunction('/config.json');
    fail('Expected error to be thrown');
  } catch (error) {
    expect(error).toBeInstanceOf(ResourceNotFoundError);
    expect(error.resourceType).toBe('file');
    expect(error.resourceId).toBe('/config.json');
    expect(error.cause).toBe(noentError);
  }
  
  readFileSpy.mockRestore();
});

it('should translate EACCES to PermissionError', async () => {
  const accessError = createFsError('EACCES', 'Permission denied', 'writeFile', '/config.json');
  const writeFileSpy = jest.spyOn(fs, 'writeFile');
  writeFileSpy.mockRejectedValueOnce(accessError);
  
  try {
    await yourFunction('/config.json', 'content');
    fail('Expected error to be thrown');
  } catch (error) {
    expect(error).toBeInstanceOf(PermissionError);
    expect(error.operation).toBe('write');
    expect(error.resourcePath).toBe('/config.json');
    expect(error.cause).toBe(accessError);
  }
  
  writeFileSpy.mockRestore();
});
```

## Testing Error Recovery

Test that higher-level code correctly handles and recovers from expected errors:

```typescript
it('should use default configuration when config file is not found', async () => {
  // Simulate file not found
  const noentError = createFsError('ENOENT', 'No such file or directory', 'readFile', '/config.json');
  const readFileSpy = jest.spyOn(fs, 'readFile');
  readFileSpy.mockRejectedValueOnce(noentError);
  
  // Verify recovery behavior
  const config = await loadConfigWithFallback('/config.json');
  expect(config).toEqual(DEFAULT_CONFIG);
  expect(console.warn).toHaveBeenCalledWith(expect.stringContaining('Using default configuration'));
  
  readFileSpy.mockRestore();
});
```

## Testing Error Handling in Async Code

For async code, test both synchronous and asynchronous error paths:

```typescript
// Synchronous error in async function
it('should handle synchronous errors in async function', async () => {
  await expect(async () => {
    await processFn(null); // Null input causes synchronous error
  }).rejects.toThrow('Invalid input');
});

// Asynchronous error
it('should handle asynchronous errors', async () => {
  const apiSpy = jest.spyOn(api, 'fetchData');
  apiSpy.mockRejectedValueOnce(new Error('Network error'));
  
  await expect(fetchAndProcessData('resource-id'))
    .rejects.toThrow('Failed to fetch data: Network error');
    
  apiSpy.mockRestore();
});
```

## Testing API Error Handling

For API-related errors, test various types of failures:

```typescript
it('should handle rate limit errors', async () => {
  const apiSpy = jest.spyOn(openaiProvider, 'generate');
  apiSpy.mockRejectedValueOnce({
    status: 429,
    message: 'Rate limit exceeded'
  });
  
  try {
    await generateText('prompt', 'openai:gpt-4');
    fail('Expected error to be thrown');
  } catch (error) {
    expect(error).toBeInstanceOf(ApiRateLimitError);
    expect(error.provider).toBe('openai');
    expect(error.retryAfter).toBeGreaterThan(0);
    expect(error.suggestions).toContain('Try again later');
  }
  
  apiSpy.mockRestore();
});

it('should handle authentication errors', async () => {
  const apiSpy = jest.spyOn(anthropicProvider, 'generate');
  apiSpy.mockRejectedValueOnce({
    status: 401,
    message: 'Invalid API key'
  });
  
  try {
    await generateText('prompt', 'anthropic:claude-3-opus');
    fail('Expected error to be thrown');
  } catch (error) {
    expect(error).toBeInstanceOf(ApiAuthenticationError);
    expect(error.provider).toBe('anthropic');
    expect(error.suggestions).toContain('Check your API key');
  }
  
  apiSpy.mockRestore();
});
```

## Testing CLI Error Output

For CLI commands, test the formatted error output:

```typescript
it('should format file not found errors nicely in CLI', async () => {
  // Mock console to capture output
  const consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation();
  
  // Create a file not found error
  const fsError = createFsError('ENOENT', 'No such file or directory', 'readFile', 'input.txt');
  const readFileSpy = jest.spyOn(fs, 'readFile');
  readFileSpy.mockRejectedValueOnce(fsError);
  
  // Run the command (assumes process.exit is also mocked)
  await runCommand(['run', 'input.txt']);
  
  // Verify error output format
  expect(consoleErrorSpy).toHaveBeenCalledWith(
    expect.stringContaining('File not found: input.txt')
  );
  expect(consoleErrorSpy).toHaveBeenCalledWith(
    expect.stringContaining('Suggestions:')
  );
  
  // Clean up
  consoleErrorSpy.mockRestore();
  readFileSpy.mockRestore();
});
```

## Best Practices

1. **Test specific error types**: Verify your application creates the correct error types.
2. **Verify error properties**: Check that all error properties (message, cause, suggestions, etc.) are correct.
3. **Test error categories**: Ensure errors are assigned to the right categories.
4. **Test error chains**: Verify cause/source errors are preserved in the error chain.
5. **Test error recovery**: Ensure the application can recover from expected errors.
6. **Test helpful error messages**: Verify that error messages are clear and helpful.
7. **Test suggestions**: Check that appropriate recovery suggestions are provided.

## Common Pitfalls

1. **Swallowing errors**: Make sure errors aren't caught and ignored without testing.
2. **Missing error properties**: Ensure all expected error properties are tested.
3. **Ignoring async errors**: Remember to use `async`/`await` with `expect().rejects` for async code.
4. **Not preserving error chains**: Verify that error causes are preserved.