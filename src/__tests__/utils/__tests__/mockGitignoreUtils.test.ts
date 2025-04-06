/**
 * Tests for mockGitignoreUtils interfaces
 * This just verifies that the interfaces are correctly exported
 * Implementation tests will be added in future tasks
 */
import { 
  mockedGitignoreUtils,
  GitignoreMockConfig,
  IgnorePathRule,
  resetMockGitignore,
  setupMockGitignore,
  mockShouldIgnorePath,
  mockCreateIgnoreFilter
} from '../mockGitignoreUtils';

describe('mockGitignoreUtils interfaces', () => {
  // This test just verifies that the interfaces and functions are exported correctly
  it('should export the expected interfaces and functions', () => {
    // Verify that the mocked module is exported
    expect(mockedGitignoreUtils).toBeDefined();
    
    // Verify that the functions are defined
    expect(resetMockGitignore).toBeInstanceOf(Function);
    expect(setupMockGitignore).toBeInstanceOf(Function);
    expect(mockShouldIgnorePath).toBeInstanceOf(Function);
    expect(mockCreateIgnoreFilter).toBeInstanceOf(Function);
    
    // We can't directly test interfaces, but we can create objects that implement them
    const config: GitignoreMockConfig = {
      defaultIgnoreBehavior: false,
      defaultIgnorePatterns: ['node_modules'],
      defaultIncludePatterns: ['src']
    };
    expect(config.defaultIgnoreBehavior).toBe(false);
    
    const rule: IgnorePathRule = {
      pattern: 'node_modules',
      ignored: true
    };
    expect(rule.ignored).toBe(true);
  });
});