/**
 * Error logging utility functions
 * 
 * This module provides utilities for enhanced error logging and debugging.
 */

/**
 * Log a detailed error with context, stack trace, and cause chain
 * 
 * @param error The error to log
 * @param context Optional context identifier for the error
 */
export function logDetailedError(error: unknown, context?: string): void {
  console.error(`ERROR ${context ? `[${context}]: ` : ''}`);
  
  if (error instanceof Error) {
    console.error(`- Message: ${error.message}`);
    console.error(`- Stack: ${error.stack || 'No stack trace available'}`);
    
    // Log additional properties
    const additionalProps: Record<string, unknown> = {};
    for (const key in error) {
      if (key !== 'message' && key !== 'stack' && key !== 'cause') {
              const anyError = error as unknown as Record<string, unknown>;
        additionalProps[key] = anyError[key];
      }
    }
    
    if (Object.keys(additionalProps).length > 0) {
      console.error('- Additional properties:');
      for (const [key, value] of Object.entries(additionalProps)) {
        console.error(`  - ${key}: ${formatValue(value)}`);
      }
    }
    
    // Log cause chain recursively
    if ('cause' in error && error.cause) {
      console.error('- Caused by:');
      logDetailedError(error.cause, 'cause');
    }
  } else {
    console.error(`- Non-Error object: ${formatValue(error)}`);
  }
}

/**
 * Format a value for logging
 * 
 * @param value The value to format
 * @returns Formatted string representation of the value
 */
function formatValue(value: unknown): string {
  if (value === null) {
    return 'null';
  }
  
  if (value === undefined) {
    return 'undefined';
  }
  
  if (typeof value === 'function') {
    return `[Function: ${value.name || 'anonymous'}]`;
  }
  
  if (typeof value === 'object') {
    try {
      return JSON.stringify(value);
    } catch (e) {
      return `[${Object.prototype.toString.call(value)}]`;
    }
  }
  
  return String(value);
}

/**
 * Test if an object is potentially a CommonJS module
 * 
 * This helps debug issues with module imports/exports by checking
 * if an object looks like a default or named export structure from CommonJS.
 * 
 * @param obj The object to test
 * @returns A string describing the object's structure
 */
export function inspectModuleObject(obj: unknown): string {
  if (obj === null || obj === undefined) {
    return `Not a module object, value is ${obj === null ? 'null' : 'undefined'}`;
  }
  
  if (typeof obj !== 'object') {
    return `Not a module object, value is of type ${typeof obj}`;
  }
  
  const output: string[] = ['Module inspection:'];
  
  // Check for potential default export
  if ('default' in (obj as Record<string, unknown>)) {
    const defaultExport = (obj as Record<string, unknown>).default;
    output.push(`- Has default export: ${typeof defaultExport}`);
    
    if (typeof defaultExport === 'function') {
      const funcName = typeof defaultExport === 'function' && 'name' in defaultExport 
        ? (defaultExport as {name?: string}).name || 'anonymous' 
        : 'anonymous';
      output.push(`  - Function name: ${funcName}`);
    } else if (defaultExport !== null && typeof defaultExport === 'object') {
      output.push(`  - Object type: ${Object.prototype.toString.call(defaultExport)}`);
      output.push(`  - Properties: ${Object.keys(defaultExport).join(', ')}`);
    }
  }
  
  // Check for likely named exports
  const entries = Object.entries(obj as Record<string, unknown>)
    .filter(([key]) => key !== 'default' && key !== '__esModule');
  
  if (entries.length > 0) {
    output.push(`- Named exports (${entries.length}):`);
    entries.forEach(([key, value]) => {
      const type = typeof value;
      let extra = '';
      if (type === 'function') {
        const valueObj = value;
        if (valueObj && typeof valueObj === 'object' && 'name' in valueObj && typeof valueObj.name === 'string') {
          extra = `function${valueObj.name ? ` "${valueObj.name}"` : ''}`;
        } else {
          extra = 'function (anonymous)';
        }
      } else if (type === 'object' && value !== null) {
        extra = `object with ${Object.keys(value as object).length} properties`;
      } else {
        extra = String(value);
      }
      
      output.push(`  - ${key}: ${type} (${extra})`);
    });
  }
  
  // Check for __esModule flag which indicates ES module transpiled to CommonJS
  if ('__esModule' in (obj as Record<string, unknown>)) {
    output.push(`- Has __esModule flag: ${Boolean((obj as Record<string, unknown>).__esModule)}`);
  }
  
  return output.join('\n');
}
