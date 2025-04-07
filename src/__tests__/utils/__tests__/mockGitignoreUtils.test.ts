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

  describe('mockShouldIgnorePath', () => {
    beforeEach(() => {
      // Reset mocks before each test
      resetMockGitignore();
      
      // Setup default behavior
      setupMockGitignore();
    });
    
    it('should configure shouldIgnorePath to return true for specific paths', async () => {
      // Configure specific path to be ignored
      mockShouldIgnorePath('/src/ignored-file.txt', true);
      
      // Verify path is ignored
      const result = await mockedGitignoreUtils.shouldIgnorePath('/src', 'ignored-file.txt');
      expect(result).toBe(true);
    });
    
    it('should configure shouldIgnorePath to return false for specific paths', async () => {
      // Setup default behavior to ignore all paths
      setupMockGitignore({ defaultIgnoreBehavior: true });
      
      // Configure specific path to not be ignored (override default)
      mockShouldIgnorePath('/src/important.txt', false);
      
      // Verify path is not ignored
      const result = await mockedGitignoreUtils.shouldIgnorePath('/src', 'important.txt');
      expect(result).toBe(false);
    });
    
    it('should support RegExp patterns for path matching', async () => {
      // Configure regex pattern to ignore all .log files
      mockShouldIgnorePath(/\.log$/, true);
      
      // Verify .log files are ignored
      const result1 = await mockedGitignoreUtils.shouldIgnorePath('/logs', 'app.log');
      expect(result1).toBe(true);
      
      // Verify other files are not ignored
      const result2 = await mockedGitignoreUtils.shouldIgnorePath('/src', 'app.js');
      expect(result2).toBe(false);
    });
    
    it('should give precedence to more recently added rules', async () => {
      // First rule - ignore all .js files
      mockShouldIgnorePath(/\.js$/, true);
      
      // Second rule - don't ignore specific .js file (should take precedence)
      mockShouldIgnorePath('/src/special.js', false);
      
      // Verify the specific rule takes precedence
      const result = await mockedGitignoreUtils.shouldIgnorePath('/src', 'special.js');
      expect(result).toBe(false);
    });
    
    it('should match against combined base path and file path', async () => {
      // Configure rule for combined path
      mockShouldIgnorePath('/base/path/file.txt', true);
      
      // Verify it matches when basePath and filePath combine to the pattern
      const result = await mockedGitignoreUtils.shouldIgnorePath('/base/path', 'file.txt');
      expect(result).toBe(true);
    });
    
    it('should match against just the file path component', async () => {
      // Configure rule for just the file name
      mockShouldIgnorePath('secret.txt', true);
      
      // Verify it matches regardless of base path
      const result1 = await mockedGitignoreUtils.shouldIgnorePath('/some/path', 'secret.txt');
      expect(result1).toBe(true);
      
      const result2 = await mockedGitignoreUtils.shouldIgnorePath('/different/path', 'secret.txt');
      expect(result2).toBe(true);
    });
  });
  
  describe('mockCreateIgnoreFilter', () => {
    beforeEach(() => {
      // Reset mocks before each test
      resetMockGitignore();
      
      // Setup default behavior
      setupMockGitignore();
    });
    
    it('should configure createIgnoreFilter to use custom string array patterns', async () => {
      // Configure specific directory to use custom ignore patterns
      mockCreateIgnoreFilter('/custom/path', ['*.txt', 'temp']);
      
      // Get the filter
      const filter = await mockedGitignoreUtils.createIgnoreFilter('/custom/path');
      
      // Verify custom patterns are applied
      expect(filter.ignores('file.txt')).toBe(true);
      expect(filter.ignores('temp/file.js')).toBe(true);
      expect(filter.ignores('src/main.js')).toBe(false);
    });
    
    it('should configure createIgnoreFilter to use custom function-based patterns', async () => {
      // Configure specific directory to use a custom ignore function
      mockCreateIgnoreFilter('/custom/path', (path) => {
        return path.includes('secret') || path.endsWith('.config');
      });
      
      // Get the filter
      const filter = await mockedGitignoreUtils.createIgnoreFilter('/custom/path');
      
      // Verify custom function is applied
      expect(filter.ignores('secret-file.js')).toBe(true);
      expect(filter.ignores('app.config')).toBe(true);
      expect(filter.ignores('src/main.js')).toBe(false);
    });
    
    it('should override default patterns for specific directories', async () => {
      // Setup default configuration with common ignore patterns
      setupMockGitignore({
        defaultIgnorePatterns: ['node_modules', 'dist']
      });
      
      // Configure specific directory to use different patterns
      mockCreateIgnoreFilter('/custom/path', ['*.txt']);
      
      // Get the filter for the custom path
      const customFilter = await mockedGitignoreUtils.createIgnoreFilter('/custom/path');
      
      // Get the filter for a different path (should use defaults)
      const defaultFilter = await mockedGitignoreUtils.createIgnoreFilter('/different/path');
      
      // Verify custom path uses custom patterns
      expect(customFilter.ignores('file.txt')).toBe(true);
      expect(customFilter.ignores('node_modules/package.json')).toBe(false); // Doesn't use default patterns
      
      // Verify different path uses default patterns
      expect(defaultFilter.ignores('node_modules/package.json')).toBe(true);
      expect(defaultFilter.ignores('file.txt')).toBe(false);
    });
    
    it('should give precedence to more recently added rules', async () => {
      // First rule
      mockCreateIgnoreFilter('/project', ['*.js']);
      
      // Second rule for same directory (should take precedence)
      mockCreateIgnoreFilter('/project', ['*.css']);
      
      // Get the filter
      const filter = await mockedGitignoreUtils.createIgnoreFilter('/project');
      
      // Verify the second rule takes precedence
      expect(filter.ignores('styles.css')).toBe(true);
      expect(filter.ignores('script.js')).toBe(false); // First rule is overridden
    });
    
    it('should allow other ignore filter methods to work with string array patterns', async () => {
      // Configure specific directory
      mockCreateIgnoreFilter('/custom/path', ['*.log']);
      
      // Get the filter
      const filter = await mockedGitignoreUtils.createIgnoreFilter('/custom/path');
      
      // Test filter method
      const filtered = filter.filter(['app.log', 'app.js', 'app.css']);
      expect(filtered).toEqual(['app.js', 'app.css']);
      
      // Test createFilter method
      const filterFn = filter.createFilter();
      expect(filterFn('error.log')).toBe(false); // Returns false for ignored paths
      expect(filterFn('main.js')).toBe(true);
      
      // Test test method
      const testResult = filter.test('debug.log');
      expect(testResult.ignored).toBe(true);
      expect(testResult.unignored).toBe(false);
    });
    
    it('should allow other ignore filter methods to work with function-based patterns', async () => {
      // Configure specific directory with function-based pattern
      mockCreateIgnoreFilter('/custom/path', (path) => path.startsWith('_') || path.endsWith('.temp'));
      
      // Get the filter
      const filter = await mockedGitignoreUtils.createIgnoreFilter('/custom/path');
      
      // Test filter method
      const filtered = filter.filter(['_hidden.js', 'visible.js', 'data.temp']);
      expect(filtered).toEqual(['visible.js']);
      
      // Test createFilter method
      const filterFn = filter.createFilter();
      expect(filterFn('_private.css')).toBe(false); // Returns false for ignored paths
      expect(filterFn('public.css')).toBe(true);
    });
  });
});
