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

describe('mockGitignoreUtils', () => {
  // This test just verifies that the interfaces and functions are exported correctly
  describe('interfaces', () => {
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

  describe('resetMockGitignore', () => {
    beforeEach(() => {
      // Clear mocks before each test
      jest.clearAllMocks();
      
      // Reset any mock state
      resetMockGitignore();
    });
    
    it('should reset all mock functions to their initial state', () => {
      // Setup mock implementations
      mockedGitignoreUtils.shouldIgnorePath.mockResolvedValue(true);
      mockedGitignoreUtils.createIgnoreFilter.mockResolvedValue({
        ignores: jest.fn().mockReturnValue(true)
      } as any);
      
      // Verify mocks are set up
      expect(mockedGitignoreUtils.shouldIgnorePath).toHaveBeenCalledTimes(0);
      expect(mockedGitignoreUtils.createIgnoreFilter).toHaveBeenCalledTimes(0);
      
      // Reset mocks
      resetMockGitignore();
      
      // Verify mocks have been reset
      expect(mockedGitignoreUtils.shouldIgnorePath.mock.calls).toHaveLength(0);
      expect(mockedGitignoreUtils.shouldIgnorePath.mock.instances).toHaveLength(0);
      expect(mockedGitignoreUtils.createIgnoreFilter.mock.calls).toHaveLength(0);
      expect(mockedGitignoreUtils.createIgnoreFilter.mock.instances).toHaveLength(0);
    });
  });
  
  describe('setupMockGitignore', () => {
    beforeEach(() => {
      // Reset mocks before each test
      resetMockGitignore();
    });
    
    it('should configure default shouldIgnorePath behavior', async () => {
      // Setup with default behavior (include paths by default)
      setupMockGitignore({ defaultIgnoreBehavior: false });
      
      // Test default behavior - paths should not be ignored
      const result = await mockedGitignoreUtils.shouldIgnorePath('/any/path', 'file.txt');
      expect(result).toBe(false);
    });
    
    it('should configure default createIgnoreFilter behavior', async () => {
      // Setup with default ignore patterns
      setupMockGitignore({
        defaultIgnorePatterns: ['node_modules', '.git']
      });
      
      // Get the filter
      const filter = await mockedGitignoreUtils.createIgnoreFilter('/any/path');
      
      // Test that default patterns are ignored
      expect(filter.ignores('node_modules/package.json')).toBe(true);
      expect(filter.ignores('.git/config')).toBe(true);
      expect(filter.ignores('src/index.ts')).toBe(false);
    });
    
    it('should support custom default ignore patterns', async () => {
      // Setup with custom ignore patterns
      setupMockGitignore({
        defaultIgnorePatterns: ['dist', 'coverage']
      });
      
      // Get the filter
      const filter = await mockedGitignoreUtils.createIgnoreFilter('/any/path');
      
      // Test that custom patterns are ignored
      expect(filter.ignores('dist/main.js')).toBe(true);
      expect(filter.ignores('coverage/report.html')).toBe(true);
      expect(filter.ignores('src/index.ts')).toBe(false);
    });
    
    it('should support custom include patterns that override ignore patterns', async () => {
      // Setup with both ignore and include patterns
      setupMockGitignore({
        defaultIgnorePatterns: ['*.log', 'tmp'],
        defaultIncludePatterns: ['important.log']
      });
      
      // Get the filter
      const filter = await mockedGitignoreUtils.createIgnoreFilter('/any/path');
      
      // Test that ignore patterns are respected except for include patterns
      expect(filter.ignores('error.log')).toBe(true);
      expect(filter.ignores('tmp/file.txt')).toBe(true);
      expect(filter.ignores('important.log')).toBe(false); // Should be included despite *.log pattern
    });
  });
});