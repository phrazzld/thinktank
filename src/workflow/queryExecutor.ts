/**
 * Query Executor module for parallel API calls to LLM providers
 * 
 * Handles concurrent execution of LLM API calls with proper error handling
 * and status tracking.
 */
import { 
  LLMResponse, 
  ModelConfig, 
  SystemPrompt,
  ModelOptions
} from '../core/types';
import { getProvider } from '../core/llmRegistry';
import { findModelGroup } from '../core/configManager';
import { AppConfig } from '../core/types';
import { 
  categorizeError,
  getTroubleshootingTip,
  errorCategories
} from '../utils/consoleUtils';

/**
 * Error thrown by the QueryExecutor module
 */
export class QueryExecutorError extends Error {
  /**
   * The category of error (e.g., "API", "Provider", etc.)
   */
  category?: string;
  
  /**
   * List of suggestions to help resolve the error
   */
  suggestions?: string[];
  
  /**
   * Examples of valid commands related to this error context
   */
  examples?: string[];
  
  constructor(message: string, public readonly cause?: Error) {
    super(message);
    this.name = 'QueryExecutorError';
    this.category = errorCategories.API;
  }
}

/**
 * Status of a model query
 */
export type QueryStatus = 'pending' | 'running' | 'success' | 'error';

/**
 * Status information for a model query
 */
export interface ModelQueryStatus {
  /**
   * The status of the query
   */
  status: QueryStatus;
  
  /**
   * Error message, if the status is 'error'
   */
  message?: string;
  
  /**
   * Detailed error object, if available
   */
  detailedError?: Error;
  
  /**
   * Start time of the query in milliseconds
   */
  startTime?: number;
  
  /**
   * End time of the query in milliseconds
   */
  endTime?: number;
  
  /**
   * Duration of the query in milliseconds
   */
  durationMs?: number;
}

/**
 * Options for executing queries
 */
export interface QueryExecutionOptions {
  /**
   * The prompt to send to the models
   */
  prompt: string;
  
  /**
   * System prompt override for all models
   */
  systemPrompt?: string;
  
  /**
   * Whether to enable thinking capability for Claude models
   */
  enableThinking?: boolean;
  
  /**
   * Timeout in milliseconds for each query (default: 120000)
   */
  timeoutMs?: number;
  
  /**
   * Status update callback
   * 
   * Called when a model's status changes
   */
  onStatusUpdate?: (
    modelKey: string, 
    status: ModelQueryStatus, 
    allStatuses: Record<string, ModelQueryStatus>
  ) => void;
}

/**
 * Result of query execution
 */
export interface QueryExecutionResult {
  /**
   * Array of model responses
   */
  responses: Array<LLMResponse & { configKey: string }>;
  
  /**
   * Status information for each model
   */
  statuses: Record<string, ModelQueryStatus>;
  
  /**
   * Timing information
   */
  timing: {
    /**
     * Start time of all queries in milliseconds
     */
    startTime: number;
    
    /**
     * End time of all queries in milliseconds
     */
    endTime: number;
    
    /**
     * Duration of all queries in milliseconds
     */
    durationMs: number;
  };
}

/**
 * Gets a key string for identifying a model
 * 
 * @param model - The model config
 * @returns A string in the format "provider:modelId"
 */
function getModelKey(model: ModelConfig): string {
  return `${model.provider}:${model.modelId}`;
}

/**
 * Execute queries in parallel to multiple LLM models
 * 
 * @param config - The application configuration
 * @param models - Array of models to query
 * @param options - Query execution options
 * @returns Promise that resolves to an object with all responses and statuses
 */
export async function executeQueries(
  config: AppConfig,
  models: ModelConfig[],
  options: QueryExecutionOptions
): Promise<QueryExecutionResult> {
  // Initialize timing
  const startTime = Date.now();
  
  // Initialize statuses for all models
  const statuses: Record<string, ModelQueryStatus> = {};
  models.forEach(model => {
    const modelKey = getModelKey(model);
    statuses[modelKey] = { status: 'pending' };
  });
  
  // Create an array to hold the promises for each model query
  const queryPromises: Array<Promise<LLMResponse & { configKey: string }>> = [];
  
  // Process each model
  for (const model of models) {
    const modelKey = getModelKey(model);
    const provider = getProvider(model.provider);
    
    // Skip if provider not found and add error entry
    if (!provider) {
      // Create provider not found error response
      const errorResponse: LLMResponse & { configKey: string } = {
        provider: model.provider,
        modelId: model.modelId,
        text: '',
        error: `Provider '${model.provider}' not found for model ${modelKey}`,
        configKey: modelKey
      };
      
      // Update status with error
      statuses[modelKey] = {
        status: 'error',
        message: errorResponse.error,
        detailedError: new QueryExecutorError(errorResponse.error || 'Provider not found')
      };
      
      // Call status update callback if provided
      if (options.onStatusUpdate) {
        options.onStatusUpdate(modelKey, statuses[modelKey], statuses);
      }
      
      // Add a resolved promise with the error response
      queryPromises.push(Promise.resolve(errorResponse));
      continue;
    }
    
    // Determine which system prompt to use
    let systemPrompt: SystemPrompt | undefined;
    let modelGroupName: string | undefined;
    
    if (options.systemPrompt) {
      // Use CLI override
      systemPrompt = {
        text: options.systemPrompt,
        metadata: { source: 'cli-override' }
      };
    } else if (model.systemPrompt) {
      // Use model-specific system prompt
      systemPrompt = model.systemPrompt;
    } else {
      // For regular cases, find the group
      const groupInfo = findModelGroup(config, model);
      if (groupInfo) {
        modelGroupName = groupInfo.groupName;
        systemPrompt = groupInfo.systemPrompt;
      }
    }
    
    // If no system prompt was found, use a default
    if (!systemPrompt) {
      systemPrompt = {
        text: 'You are a helpful, accurate, and intelligent assistant. Provide clear, concise, and correct information.',
        metadata: { source: 'default-fallback' }
      };
    }
    
    // Prepare model options with thinking capability if applicable
    const modelOptions: ModelOptions = { ...model.options };
    
    // Enable thinking capability for Claude models if requested
    if (options.enableThinking && model.provider === 'anthropic' && model.modelId.includes('claude-3')) {
      modelOptions.thinking = {
        type: 'enabled',
        budget_tokens: 16000 // Default budget
      };
    }
    
    // Update status to running
    statuses[modelKey] = { 
      status: 'running',
      startTime: Date.now()
    };
    
    // Call status update callback if provided
    if (options.onStatusUpdate) {
      options.onStatusUpdate(modelKey, statuses[modelKey], statuses);
    }
    
    // Create a controller for the fetch abort
    const controller = new AbortController();
    
    // Create the promise for this model
    const queryPromise = Promise.race([
      provider.generate(options.prompt, model.modelId, modelOptions, systemPrompt),
      new Promise<never>((_, reject) => {
        // Set a timeout to prevent getting stuck on a model that's taking too long
        const timeoutId = setTimeout(() => {
          // Abort any ongoing fetch requests
          controller.abort();
          reject(new Error(`Model ${modelKey} timed out after ${options.timeoutMs || 120000}ms. The API might be unresponsive.`));
        }, options.timeoutMs || 120000); // Default to 2 minute timeout if not specified
        
        // Clean up the timeout if the main promise resolves or rejects
        setTimeout(() => clearTimeout(timeoutId), options.timeoutMs || 120000);
      })
    ])
      .then(response => {
        // Calculate duration
        const endTime = Date.now();
        const durationMs = endTime - (statuses[modelKey].startTime || endTime);
        
        // Update status with success
        statuses[modelKey] = { 
          status: 'success',
          startTime: statuses[modelKey].startTime,
          endTime,
          durationMs
        };
        
        // Call status update callback if provided
        if (options.onStatusUpdate) {
          options.onStatusUpdate(modelKey, statuses[modelKey], statuses);
        }
        
        // Add group information to the response if applicable
        const responseWithKey: LLMResponse & { configKey: string } = {
          ...response,
          configKey: modelKey,
        };
        
        if (modelGroupName && systemPrompt) {
          responseWithKey.groupInfo = {
            name: modelGroupName,
            systemPrompt
          };
        }
        
        return responseWithKey;
      })
      .catch(error => {
        // Calculate duration
        const endTime = Date.now();
        const durationMs = endTime - (statuses[modelKey].startTime || endTime);
        
        // Get error message and categorize it
        const errorMessage = error instanceof Error ? error.message : String(error);
        const errorObj = error instanceof Error ? error : new Error(String(error));
        const category = categorizeError(errorObj);
        const tip = getTroubleshootingTip(errorObj, category);
        
        // Update status with error
        statuses[modelKey] = { 
          status: 'error',
          message: errorMessage,
          detailedError: errorObj,
          startTime: statuses[modelKey].startTime,
          endTime,
          durationMs
        };
        
        // Call status update callback if provided
        if (options.onStatusUpdate) {
          options.onStatusUpdate(modelKey, statuses[modelKey], statuses);
        }
        
        // Return error response
        return {
          provider: model.provider,
          modelId: model.modelId,
          text: '',
          error: errorMessage,
          errorCategory: category,
          errorTip: tip,
          configKey: modelKey,
        };
      });
    
    queryPromises.push(queryPromise);
  }
  
  // Execute all queries in parallel
  const responses = await Promise.all(queryPromises);
  
  // Calculate overall timing
  const endTime = Date.now();
  const durationMs = endTime - startTime;
  
  // Return the results
  return {
    responses,
    statuses,
    timing: {
      startTime,
      endTime,
      durationMs
    }
  };
}