# Audit Report: Logger Calls for Secret Leaks in OpenRouter Provider

## Summary
After reviewing all logger calls in the `internal/providers/openrouter/*` files, I found no direct leaks of API keys or other sensitive information through logging. The OpenRouter implementation demonstrates a security-conscious approach to logging, where API keys and sensitive information are properly handled. However, there is one area that could benefit from improved precautions around logging URLs with potential embedded credentials.

## Methodology
1. Identified all files in the `internal/providers/openrouter/` directory
2. Used grep to search for all logger calls (Debug, Info, Warn, Error)
3. Examined each logger call to identify what data is being logged
4. Checked for any variables that might contain sensitive data
5. Verified that no API keys, tokens, or credentials are logged

## Findings

### Files Reviewed
1. `/Users/phaedrus/Development/architect/internal/providers/openrouter/client.go`
2. `/Users/phaedrus/Development/architect/internal/providers/openrouter/provider.go`
3. `/Users/phaedrus/Development/architect/internal/providers/openrouter/errors.go`

### Logger Calls Analysis

#### client.go
- Line 264: Logs `"Sending request to OpenRouter API: %s", apiURL` - **Potential Risk**: While the standard OpenRouter API URL does not contain sensitive information, custom API endpoints might include embedded credentials in the URL.
- Line 287: Logs `"Failed to close response body: %v", closeErr` - No sensitive data
- Line 411-412: Logs `"Counted %d tokens for model %s", count, c.modelID` - No sensitive data
- Line 443-444: Logs warning about model ID format - No sensitive data
- Line 577-578: Logs unknown provider warning - No sensitive data
- Line 582-583: Logs model info debug message - No sensitive data

#### provider.go
- Line 39: Logs `"Creating OpenRouter client for model: %s", modelID` - No sensitive data
- Line 44: Logs `"Using provided API key"` - **Good Practice**: Only logs the presence of an API key, not the key itself
- Line 51: Logs `"Using API key from OPENROUTER_API_KEY environment variable"` - **Good Practice**: Only logs the source of the API key, not the key itself
- Line 58-60: Logs API endpoint info - **Potential Risk**: Similar to the issue in client.go, custom API endpoints might include embedded credentials
- Line 70: Logs warning about model ID format - No sensitive data

### Safety Practices Observed
1. **No direct logging of API keys** - The code logs the presence/source of API keys but never the key values
2. **API key handling** - API keys are properly validated without logging their values
3. **Error handling** - Error responses are carefully formatted to avoid including any potential sensitive information
4. **Request/response handling** - The code properly avoids logging full request/response payloads that might contain sensitive data

### Areas of Improvement
1. **API Endpoint URL Logging** - The code currently logs API endpoints directly. While standard OpenRouter endpoints are safe to log, custom endpoints might contain embedded credentials in the URL string. The code should sanitize any custom API endpoint URLs before logging them, similar to how it handles API keys.

## Recommendations

1. **Sanitize API Endpoint URLs** - Modify the logging code to sanitize API endpoint URLs before logging them. For example:
   ```go
   // Before logging an API endpoint URL, sanitize it to remove any potential credentials
   func sanitizeURL(url string) string {
       // Parse the URL
       parsedURL, err := url.Parse(url)
       if err != nil {
           // If parsing fails, return a generic representation
           return "[unparsable-url]"
       }

       // If there's user info, remove it
       if parsedURL.User != nil {
           parsedURL.User = nil
       }

       return parsedURL.String()
   }
   ```

2. **Add Logging Guidelines** - Add documentation to the OpenRouter package that clearly outlines logging best practices, particularly emphasizing never to log API keys, tokens, or URLs with embedded credentials.

3. **Consider Log Level Review** - Some of the debug-level logs might reveal more information than necessary. Consider if some debug logs with endpoint information should be moved to trace level or made more generic.

## Conclusion
The OpenRouter provider implementation generally follows good security practices for logging. No direct leaks of API keys or sensitive credentials were found. However, there is a minor potential risk with logging custom API endpoints that might contain embedded credentials. This can be addressed by sanitizing URLs before logging them.

This audit provides assurance that the OpenRouter provider doesn't leak API keys or other obvious secrets through logging. With the recommended improvements, the security posture of the package's logging can be further strengthened.
