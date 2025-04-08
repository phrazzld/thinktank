/**
 * Tests for path normalization utilities
 */
import path from 'path';
import { 
  normalizePathGeneral, 
  normalizePathsForComparison, 
  normalizePathForGitignore 
} from '../pathUtils';

describe('Path Normalization Utilities', () => {
  describe('normalizePathGeneral', () => {
    test('normalizes paths with different separators to forward slashes', () => {
      expect(normalizePathGeneral('path\\to\\file')).toBe('path/to/file');
      expect(normalizePathGeneral('path/to/file')).toBe('path/to/file');
    });

    test('removes leading slash by default', () => {
      expect(normalizePathGeneral('/path/to/file')).toBe('path/to/file');
    });

    test('keeps leading slash when specified', () => {
      expect(normalizePathGeneral('/path/to/file', true)).toBe('/path/to/file');
      expect(normalizePathGeneral('path/to/file', true)).toBe('/path/to/file');
    });

    test('handles redundant slashes and dots', () => {
      expect(normalizePathGeneral('path//to/../file')).toBe('path/file');
      expect(normalizePathGeneral('/path//to/../file')).toBe('path/file');
      expect(normalizePathGeneral('/path//to/../file', true)).toBe('/path/file');
    });

    test('handles empty and dot paths', () => {
      expect(normalizePathGeneral('')).toBe('.');
      expect(normalizePathGeneral('.')).toBe('.');
      expect(normalizePathGeneral('./')).toBe('.');
      expect(normalizePathGeneral('.', true)).toBe('/.');
    });
  });

  describe('normalizePathsForComparison', () => {
    test('normalizes and returns both paths with consistent format', () => {
      const [p1, p2] = normalizePathsForComparison('/path/to/file', 'path\\to\\file');
      expect(p1).toBe('path/to/file');
      expect(p2).toBe('path/to/file');
      expect(p1).toBe(p2);
    });

    test('handles paths with different directory traversal', () => {
      const [p1, p2] = normalizePathsForComparison('path/to/../to/file', 'path/to/file');
      expect(p1).toBe('path/to/file');
      expect(p2).toBe('path/to/file');
      expect(p1).toBe(p2);
    });

    test('normalizes paths with different casing on Windows', () => {
      const [p1, p2] = normalizePathsForComparison('Path/To/File', 'path/to/file');
      // Note: Does not lowercase since that's platform-specific and should be handled separately if needed
      if (process.platform === 'win32') {
        expect(p1.toLowerCase()).toBe(p2.toLowerCase());
      }
    });

    test('handles empty paths', () => {
      const [p1, p2] = normalizePathsForComparison('', '');
      expect(p1).toBe('.');
      expect(p2).toBe('.');
    });
  });

  describe('normalizePathForGitignore', () => {
    test('makes path relative to the base path', () => {
      expect(normalizePathForGitignore('/project/src/file.js', '/project')).toBe('src/file.js');
      expect(normalizePathForGitignore('C:\\project\\src\\file.js', 'C:\\project')).toBe('src/file.js');
    });

    test('handles paths outside of base path', () => {
      expect(normalizePathForGitignore('/other/dir/file.js', '/project')).toBe('../other/dir/file.js');
    });

    test('returns dot for identical paths', () => {
      expect(normalizePathForGitignore('/project', '/project')).toBe('.');
    });

    test('handles nested relative paths correctly', () => {
      expect(normalizePathForGitignore('/project/src/../lib/file.js', '/project')).toBe('lib/file.js');
    });

    test('removes leading ./ from relative paths', () => {
      const result = normalizePathForGitignore(
        path.join('/project', '.', 'src', 'file.js'), 
        '/project'
      );
      expect(result).toBe('src/file.js');
    });
  });
});
