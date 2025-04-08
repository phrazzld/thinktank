/**
 * Tests for the filesystem mock setup utilities
 */
const path = require('path');
const fs = require('fs');
const fsPromises = require('fs/promises');

// Import the helpers to test
const {
  setupBasicFs,
  resetFs,
  createFsError,
  getFs,
  normalizePath,
  normalizeTestPath,
  inspectVirtualFs,
  setupPlatformEnv
} = require('../fs');

describe('Filesystem Mock Setup', () => {
  // Reset before each test
  beforeEach(() => {
    resetFs();
  });

  describe('setupBasicFs', () => {
    it('should create a basic filesystem structure', () => {
      setupBasicFs({
        '/test/file.txt': 'Test content',
        '/test/dir/nested.txt': 'Nested content'
      });

      // Verify files were created
      expect(fs.existsSync('/test/file.txt')).toBe(true);
      expect(fs.existsSync('/test/dir/nested.txt')).toBe(true);
      expect(fs.readFileSync('/test/file.txt', 'utf8')).toBe('Test content');
      expect(fs.readFileSync('/test/dir/nested.txt', 'utf8')).toBe('Nested content');
    });

    it('should automatically normalize paths', () => {
      setupBasicFs({
        'relative/path.txt': 'Relative content',
        'C:\\windows\\path.txt': 'Windows content'
      });

      // Normalized paths should exist in the filesystem
      const normalizedRelative = normalizePath('relative/path.txt');
      const normalizedWindows = normalizePath('C:\\windows\\path.txt');

      expect(fs.existsSync(normalizedRelative)).toBe(true);
      expect(fs.existsSync(normalizedWindows)).toBe(true);
    });

    it('should not reset the filesystem if reset option is false', () => {
      // First setup
      setupBasicFs({
        '/test/file1.txt': 'File 1 content'
      });

      // Second setup with reset: false
      setupBasicFs({
        '/test/file2.txt': 'File 2 content'
      }, { reset: false });

      // Both files should exist
      expect(fs.existsSync('/test/file1.txt')).toBe(true);
      expect(fs.existsSync('/test/file2.txt')).toBe(true);
    });
  });

  describe('createFsError', () => {
    it('should create a properly formatted fs error', () => {
      const error = createFsError('ENOENT', 'No such file or directory', 'open', '/test/missing.txt');

      expect(error.code).toBe('ENOENT');
      expect(error.message).toContain('No such file or directory');
      expect(error.syscall).toBe('open');
      expect(error.path).toBe('/test/missing.txt');
      expect(error instanceof Error).toBe(true);
    });
  });

  describe('normalizeTestPath', () => {
    it('should normalize relative paths with a leading slash', () => {
      expect(normalizeTestPath('project/file.txt')).toBe('/project/file.txt');
      expect(normalizeTestPath('dir/subdir')).toBe('/dir/subdir');
    });

    it('should handle paths that already have a leading slash', () => {
      expect(normalizeTestPath('/already/normalized')).toBe('/already/normalized');
    });

    it('should normalize Windows paths', () => {
      expect(normalizeTestPath('C:\\Windows\\Path')).toBe('/C:/Windows/Path');
    });
  });

  describe('inspectVirtualFs', () => {
    it('should return the filesystem state as JSON', () => {
      setupBasicFs({
        '/test/file1.txt': 'Content 1',
        '/test/file2.txt': 'Content 2',
        '/test/subdir/file3.txt': 'Content 3'
      });

      const state = inspectVirtualFs();
      expect(state['/test/file1.txt']).toBe('Content 1');
      expect(state['/test/file2.txt']).toBe('Content 2');
      expect(state['/test/subdir/file3.txt']).toBe('Content 3');
    });

    it('should return a specific directory when provided', () => {
      setupBasicFs({
        '/test/file1.txt': 'Content 1',
        '/other/file2.txt': 'Content 2'
      });

      const testDirState = inspectVirtualFs('/test');
      expect(testDirState['/test/file1.txt']).toBe('Content 1');
      expect(testDirState['/other/file2.txt']).toBeUndefined();
    });
  });

  describe('setupPlatformEnv', () => {
    it('should mock process.platform and environment variables', () => {
      const originalPlatform = process.platform;
      const originalHome = process.env.HOME;

      const restore = setupPlatformEnv('win32', { 
        HOME: 'C:\\Users\\Test',
        CUSTOM_VAR: 'test-value'
      });

      expect(process.platform).toBe('win32');
      expect(process.env.HOME).toBe('C:\\Users\\Test');
      expect(process.env.CUSTOM_VAR).toBe('test-value');

      restore();

      expect(process.platform).toBe(originalPlatform);
      expect(process.env.HOME).toBe(originalHome);
      expect(process.env.CUSTOM_VAR).toBeUndefined();
    });

    it('should handle undefined environment variables', () => {
      const originalAppData = process.env.APPDATA;
      delete process.env.APPDATA; // Ensure it's undefined

      const restore = setupPlatformEnv('win32', { 
        APPDATA: 'C:\\Users\\Test\\AppData\\Roaming'
      });

      expect(process.env.APPDATA).toBe('C:\\Users\\Test\\AppData\\Roaming');

      restore();

      expect(process.env.APPDATA).toBeUndefined();
    });
  });
});