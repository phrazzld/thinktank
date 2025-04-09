/**
 * This file was created to deliberately introduce a lint error for testing the CI workflow.
 * The violation of 'no-var' rule is intentional.
 */

export function testFunction(): string {
  // Using 'var' instead of 'const' or 'let' (violates no-var rule, line 33 in .eslintrc.js)
  var message = "This is a deliberate lint error";
  
  return message;
}