/**
 * Tests for the name generator module
 */
import { GoogleGenerativeAI } from '@google/generative-ai';
import { generateFunName, generateFallbackName } from '../nameGenerator';

// Mock Google Generative AI client
jest.mock('@google/generative-ai');

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
    it('should return null when no API key is available', async () => {
      // Ensure no API key is set
      delete process.env.GEMINI_API_KEY;
      delete process.env.GOOGLE_API_KEY;
      
      const result = await generateFunName();
      
      expect(result).toBeNull();
    });

    it('should return a valid adjective-noun name when API call succeeds with JSON', async () => {
      // Set mock API key
      process.env.GEMINI_API_KEY = 'test-api-key';
      
      // Mock the response
      const mockResponse = {
        text: jest.fn().mockReturnValue('{"name":"clever-meadow"}'),
      };
      
      const mockGenerateContent = jest.fn().mockResolvedValue({
        response: mockResponse
      });
      
      const mockGetGenerativeModel = jest.fn().mockReturnValue({
        generateContent: mockGenerateContent
      });
      
      (GoogleGenerativeAI as jest.Mock).mockImplementation(() => ({
        getGenerativeModel: mockGetGenerativeModel
      }));
      
      const result = await generateFunName();
      
      expect(result).toBe('clever-meadow');
      expect(GoogleGenerativeAI).toHaveBeenCalledWith('test-api-key');
      expect(mockGetGenerativeModel).toHaveBeenCalled();
      expect(mockGenerateContent).toHaveBeenCalled();
    });

    it('should return a valid adjective-noun name when API call succeeds with plain text', async () => {
      // Set mock API key
      process.env.GEMINI_API_KEY = 'test-api-key';
      
      // Mock the response for text generation
      const mockResponse = {
        text: jest.fn().mockReturnValue('swift-stream'),
      };
      
      const mockGenerateContent = jest.fn().mockResolvedValue({
        response: mockResponse
      });
      
      const mockGetGenerativeModel = jest.fn().mockReturnValue({
        generateContent: mockGenerateContent
      });
      
      (GoogleGenerativeAI as jest.Mock).mockImplementation(() => ({
        getGenerativeModel: mockGetGenerativeModel
      }));
      
      const result = await generateFunName();
      
      expect(result).toBe('swift-stream');
    });

    it('should extract a valid name from invalid response if possible', async () => {
      // Set mock API key
      process.env.GEMINI_API_KEY = 'test-api-key';
      
      // Mock the response with extra text
      const mockResponse = {
        text: jest.fn().mockReturnValue('Here is a good name: happy-valley. Hope you like it!'),
      };
      
      const mockGenerateContent = jest.fn().mockResolvedValue({
        response: mockResponse
      });
      
      const mockGetGenerativeModel = jest.fn().mockReturnValue({
        generateContent: mockGenerateContent
      });
      
      (GoogleGenerativeAI as jest.Mock).mockImplementation(() => ({
        getGenerativeModel: mockGetGenerativeModel
      }));
      
      const result = await generateFunName();
      
      expect(result).toBe('happy-valley');
    });

    it('should return null when API call fails', async () => {
      // Set mock API key
      process.env.GEMINI_API_KEY = 'test-api-key';
      
      // Mock the API call to throw an error
      const mockGenerateContent = jest.fn().mockRejectedValue(new Error('API error'));
      
      const mockGetGenerativeModel = jest.fn().mockReturnValue({
        generateContent: mockGenerateContent
      });
      
      (GoogleGenerativeAI as jest.Mock).mockImplementation(() => ({
        getGenerativeModel: mockGetGenerativeModel
      }));
      
      const result = await generateFunName();
      
      expect(result).toBeNull();
    });

    it('should return null when response format is invalid', async () => {
      // Set mock API key
      process.env.GEMINI_API_KEY = 'test-api-key';
      
      // Mock the response with invalid format
      const mockResponse = {
        text: jest.fn().mockReturnValue('This is not a valid name format'),
      };
      
      const mockGenerateContent = jest.fn().mockResolvedValue({
        response: mockResponse
      });
      
      const mockGetGenerativeModel = jest.fn().mockReturnValue({
        generateContent: mockGenerateContent
      });
      
      (GoogleGenerativeAI as jest.Mock).mockImplementation(() => ({
        getGenerativeModel: mockGetGenerativeModel
      }));
      
      const result = await generateFunName();
      
      expect(result).toBeNull();
    });
  });

  describe('generateFallbackName', () => {
    it('should generate a timestamp-based name in the correct format', () => {
      const result = generateFallbackName();
      
      // Check that it matches the expected pattern: run-YYYYMMDD-HHmmss
      expect(result).toMatch(/^run-\d{8}-\d{6}$/);
    });
  });
});