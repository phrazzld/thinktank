import { ThrottledSpinner, FileOutputStatus } from '../throttledSpinner';
import { ModelQueryStatus } from '../../workflow/queryExecutor';
import ora from 'ora';

// Mock ora
jest.mock('ora', () => {
  const mockSpinner = {
    start: jest.fn().mockReturnThis(),
    stop: jest.fn().mockReturnThis(),
    succeed: jest.fn().mockReturnThis(),
    fail: jest.fn().mockReturnThis(),
    warn: jest.fn().mockReturnThis(),
    info: jest.fn().mockReturnThis(),
    text: '',
    isSpinning: true
  };
  
  return jest.fn().mockImplementation(() => mockSpinner);
});

describe('Enhanced ThrottledSpinner Methods', () => {
  let throttledSpinner: ThrottledSpinner;
  const mockSpinner = ora('');
  
  beforeEach(() => {
    jest.useFakeTimers();
    jest.clearAllMocks();
    throttledSpinner = new ThrottledSpinner({ initialText: 'Initial text' });
    // Reset spinner text
    mockSpinner.text = '';
  });
  
  afterEach(() => {
    jest.useRealTimers();
  });
  
  test('should have updateForModelStatus method', () => {
    // Test fails because the method does not exist yet
    expect(typeof throttledSpinner.updateForModelStatus).toBe('function');
  });
  
  test('should update text appropriately for running model status', () => {
    // Test will fail until we implement the method
    const modelKey = 'openai:gpt-4';
    const status: ModelQueryStatus = { status: 'running' as any };
    
    throttledSpinner.updateForModelStatus(modelKey, status);
    expect(throttledSpinner.getCurrentText()).toContain('Querying');
    expect(throttledSpinner.getCurrentText()).toContain(modelKey);
  });
  
  test('should update text appropriately for success model status', () => {
    // Test will fail until we implement the method
    const modelKey = 'anthropic:claude-3';
    const status: ModelQueryStatus = { 
      status: 'success' as any,
      durationMs: 1500
    };
    
    throttledSpinner.updateForModelStatus(modelKey, status);
    expect(throttledSpinner.getCurrentText()).toContain('Received response from');
    expect(throttledSpinner.getCurrentText()).toContain(modelKey);
    expect(throttledSpinner.getCurrentText()).toContain('1500ms');
  });
  
  test('should update text appropriately for error model status', () => {
    // Test will fail until we implement the method
    const modelKey = 'google:gemini';
    const status: ModelQueryStatus = { 
      status: 'error' as any,
      message: 'API key invalid'
    };
    
    throttledSpinner.updateForModelStatus(modelKey, status);
    expect(throttledSpinner.getCurrentText()).toContain('Error from');
    expect(throttledSpinner.getCurrentText()).toContain(modelKey);
    expect(throttledSpinner.getCurrentText()).toContain('API key invalid');
  });
  
  test('should have updateForFileStatus method', () => {
    // Test fails because the method does not exist yet
    expect(typeof throttledSpinner.updateForFileStatus).toBe('function');
  });
  
  test('should update text appropriately for pending file status', () => {
    // Test will fail until we implement the method
    const fileDetail: FileOutputStatus = { 
      modelKey: 'openai:gpt-4',
      filename: 'output.txt',
      status: 'pending'
    };
    
    throttledSpinner.updateForFileStatus(fileDetail);
    expect(throttledSpinner.getCurrentText()).toContain('Writing file for');
    expect(throttledSpinner.getCurrentText()).toContain(fileDetail.modelKey);
  });
  
  test('should update text appropriately for success file status', () => {
    // Test will fail until we implement the method
    const fileDetail: FileOutputStatus = { 
      modelKey: 'anthropic:claude-3',
      filename: 'output.txt',
      status: 'success'
    };
    
    throttledSpinner.updateForFileStatus(fileDetail);
    expect(throttledSpinner.getCurrentText()).toContain('Wrote results for');
    expect(throttledSpinner.getCurrentText()).toContain(fileDetail.modelKey);
  });
  
  test('should update text appropriately for error file status', () => {
    // Test will fail until we implement the method
    const fileDetail: FileOutputStatus = { 
      modelKey: 'google:gemini',
      filename: 'output.txt',
      status: 'error',
      error: 'Permission denied'
    };
    
    throttledSpinner.updateForFileStatus(fileDetail);
    expect(throttledSpinner.getCurrentText()).toContain('Error writing file for');
    expect(throttledSpinner.getCurrentText()).toContain(fileDetail.modelKey);
    expect(throttledSpinner.getCurrentText()).toContain('Permission denied');
  });
  
  test('should have updateForModelSummary method', () => {
    // Test fails because the method does not exist yet
    expect(typeof throttledSpinner.updateForModelSummary).toBe('function');
  });
  
  test('should update text appropriately for complete success model summary', () => {
    // Test will fail until we implement the method
    const successCount = 3;
    const failureCount = 0;
    
    throttledSpinner.updateForModelSummary(successCount, failureCount);
    expect(throttledSpinner.getCurrentText()).toContain('Query execution complete');
    expect(throttledSpinner.getCurrentText()).toContain('3 succeeded');
    expect(throttledSpinner.getCurrentText()).toContain('0 failed');
  });
  
  test('should update text appropriately for partial success model summary', () => {
    // Test will fail until we implement the method
    const successCount = 2;
    const failureCount = 1;
    
    throttledSpinner.updateForModelSummary(successCount, failureCount);
    expect(throttledSpinner.getCurrentText()).toContain('Query execution complete');
    expect(throttledSpinner.getCurrentText()).toContain('2 succeeded');
    expect(throttledSpinner.getCurrentText()).toContain('1 failed');
  });
  
  test('should have updateForFileSummary method', () => {
    // Test fails because the method does not exist yet
    expect(typeof throttledSpinner.updateForFileSummary).toBe('function');
  });
  
  test('should update text appropriately for complete success file summary', () => {
    // Test will fail until we implement the method
    const succeededWrites = 3;
    const failedWrites = 0;
    
    throttledSpinner.updateForFileSummary(succeededWrites, failedWrites);
    expect(throttledSpinner.getCurrentText()).toContain('Output processing complete');
    expect(throttledSpinner.getCurrentText()).toContain('3 files written');
    expect(throttledSpinner.getCurrentText()).toContain('0 failed');
  });
  
  test('should update text appropriately for partial success file summary', () => {
    // Test will fail until we implement the method
    const succeededWrites = 2;
    const failedWrites = 1;
    
    throttledSpinner.updateForFileSummary(succeededWrites, failedWrites);
    expect(throttledSpinner.getCurrentText()).toContain('Output processing complete');
    expect(throttledSpinner.getCurrentText()).toContain('2 files written');
    expect(throttledSpinner.getCurrentText()).toContain('1 failed');
  });
});
