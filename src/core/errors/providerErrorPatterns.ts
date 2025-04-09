/**
 * Direct pattern matching for provider error detection
 * 
 * This module provides direct pattern matching functions for detecting different types of provider errors.
 * It duplicates the patterns from categorization.ts but provides a more direct approach to avoid any
 * module resolution issues that might occur with the regular error utilities.
 */

/**
 * Rate limit patterns to detect rate limiting errors
 */
const RATE_LIMIT_PATTERNS = [
  /rate\s+limit/i,
  /429/,
  /too\s+many\s+requests/i,
  /quota\s+exceeded/i,
  /exceeded\s+your\s+current\s+quota/i,
  /exceeded\s+your\s+quota/i,
];

/**
 * Token limit patterns to detect token/context length errors
 */
const TOKEN_LIMIT_PATTERNS = [
  /token\s+limit/i,
  /maximum\s+context\s+length/i,
  /maximum\s+token/i,
  /context\s+window/i,
  /context\s+length/i,
];

/**
 * Content policy patterns to detect content policy violations
 */
const CONTENT_POLICY_PATTERNS = [
  /content\s+policy/i,
  /content\s+filter/i,
  /safety\s+system/i,
  /violates/i,
  /harmful/i,
  /inappropriate/i,
  /prohibited/i,
];

/**
 * Authentication error patterns to detect auth issues
 */
const AUTH_ERROR_PATTERNS = [
  /authentication/i,
  /auth/i,
  /api\s+key/i,
  /apikey/i,
  /unauthorized/i,
  /invalid\s+key/i,
  /401/,
];

/**
 * Network error patterns to detect network issues
 */
const NETWORK_ERROR_PATTERNS = [
  /network/i,
  /connection/i,
  /timeout/i,
  /econnrefused/i,
  /socket/i,
  /dns/i,
  /etimedout/i,
  /connect/i,
];

/**
 * Tests if a message matches any pattern in an array of regex patterns
 * 
 * @param message The error message to test
 * @param patterns Array of regex patterns to test against
 * @returns Whether the message matches any pattern
 */
function testPatterns(message: string, patterns: RegExp[]): boolean {
  return patterns.some(pattern => pattern.test(message));
}

/**
 * Direct implementation of rate limit error detection
 * 
 * @param errorMessage Error message to check
 * @returns Whether the message indicates a rate limit error
 */
export function detectRateLimitError(errorMessage: string): boolean {
  return testPatterns(errorMessage, RATE_LIMIT_PATTERNS);
}

/**
 * Direct implementation of token limit error detection
 * 
 * @param errorMessage Error message to check
 * @returns Whether the message indicates a token limit error
 */
export function detectTokenLimitError(errorMessage: string): boolean {
  return testPatterns(errorMessage, TOKEN_LIMIT_PATTERNS);
}

/**
 * Direct implementation of content policy error detection
 * 
 * @param errorMessage Error message to check
 * @returns Whether the message indicates a content policy violation
 */
export function detectContentPolicyError(errorMessage: string): boolean {
  return testPatterns(errorMessage, CONTENT_POLICY_PATTERNS);
}

/**
 * Direct implementation of auth error detection
 * 
 * @param errorMessage Error message to check
 * @returns Whether the message indicates an authentication error
 */
export function detectAuthError(errorMessage: string): boolean {
  return testPatterns(errorMessage, AUTH_ERROR_PATTERNS);
}

/**
 * Direct implementation of network error detection
 * 
 * @param errorMessage Error message to check
 * @returns Whether the message indicates a network error
 */
export function detectNetworkError(errorMessage: string): boolean {
  return testPatterns(errorMessage, NETWORK_ERROR_PATTERNS);
}
