import { normalizePath } from './pathUtils';

describe('pathUtils', () => {
  describe('normalizePath', () => {
    test('normalizes paths with different separators to forward slashes', () => {
      expect(normalizePath('path\\to\\file')).toBe('path/to/file');
      expect(normalizePath('path/to/file')).toBe('path/to/file');
    });

    test('removes leading slash by default', () => {
      expect(normalizePath('/path/to/file')).toBe('path/to/file');
    });

    test('keeps leading slash when specified', () => {
      expect(normalizePath('/path/to/file', true)).toBe('/path/to/file');
      expect(normalizePath('path/to/file', true)).toBe('/path/to/file');
    });

    test('handles redundant slashes and dots', () => {
      expect(normalizePath('path//to/../file')).toBe('path/file');
      expect(normalizePath('/path//to/../file')).toBe('path/file');
      expect(normalizePath('/path//to/../file', true)).toBe('/path/file');
    });

    test('handles empty and dot paths', () => {
      expect(normalizePath('')).toBe('.');
      expect(normalizePath('.')).toBe('.');
      expect(normalizePath('./')).toBe('.');
      expect(normalizePath('.', true)).toBe('/.');
    });
  });
});
