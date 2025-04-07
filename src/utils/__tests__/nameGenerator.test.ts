/**
 * Tests for the name generator module
 */
import { generateFunName, generateFallbackName, ADJECTIVES, NOUNS } from '../nameGenerator';

// This will be used after refactoring
// No need to mock Google Generative AI anymore

// Mock process.env
const originalEnv = process.env;

describe('nameGenerator', () => {
  beforeEach(() => {
    // Reset all mocks
    jest.resetAllMocks();
    
    // Reset process.env
    process.env = { ...originalEnv };
    
    // Mock the logger to avoid console output during tests
    jest.spyOn(console, 'debug').mockImplementation(() => {});
    jest.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterAll(() => {
    // Restore process.env
    process.env = originalEnv;
  });

  describe('generateFunName', () => {
    it('should return a name in the format "adjective-noun"', () => {
      const name = generateFunName();
      expect(name).toMatch(/^[a-z]+-[a-z]+$/);
    });

    it('should select words from the predefined lists', () => {
      const name = generateFunName();
      const [adjective, noun] = name.split('-');
      
      expect(ADJECTIVES).toContain(adjective);
      expect(NOUNS).toContain(noun);
    });

    it('should generate different names on multiple calls (usually)', () => {
      // We'll generate multiple names and check that we get at least 2 unique results
      // This is probabilistic but with 50+ words in each list, it's incredibly unlikely
      // to get the same name three times in a row by chance
      const names = new Set([
        generateFunName(),
        generateFunName(),
        generateFunName(),
        generateFunName(),
        generateFunName()
      ]);
      
      // We should have at least 2 unique names out of 5 attempts
      expect(names.size).toBeGreaterThan(1);
    });

    it('should always return a valid string', () => {
      // Function should never return null or undefined
      const name = generateFunName();
      expect(typeof name).toBe('string');
      expect(name.length).toBeGreaterThan(0);
    });
  });

  describe('generateFallbackName', () => {
    it('should generate a timestamp-based name in the correct format', () => {
      const result = generateFallbackName();
      
      // Check that it matches the expected pattern: run-YYYYMMDD-HHmmss
      expect(result).toMatch(/^run-\d{8}-\d{6}$/);
    });
  });

  describe('word lists', () => {
    it('should have at least 50 adjectives and 50 nouns', () => {
      expect(ADJECTIVES).toBeDefined();
      expect(NOUNS).toBeDefined();
      expect(ADJECTIVES.length).toBeGreaterThanOrEqual(50);
      expect(NOUNS.length).toBeGreaterThanOrEqual(50);
    });

    it('should contain only lowercase strings with no special characters', () => {
      // All adjectives should be lowercase and contain only letters
      ADJECTIVES.forEach((adj: string) => {
        expect(adj).toMatch(/^[a-z]+$/);
      });

      // All nouns should be lowercase and contain only letters
      NOUNS.forEach((noun: string) => {
        expect(noun).toMatch(/^[a-z]+$/);
      });
    });

    it('should not have duplicate words in each list', () => {
      // Check for duplicates in ADJECTIVES
      const uniqueAdjectives = new Set(ADJECTIVES);
      expect(uniqueAdjectives.size).toBe(ADJECTIVES.length);

      // Check for duplicates in NOUNS
      const uniqueNouns = new Set(NOUNS);
      expect(uniqueNouns.size).toBe(NOUNS.length);
    });
  });
});
