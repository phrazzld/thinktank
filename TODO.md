# Thinktank Issues and Solutions

## ✅ Issue 1: All top-level models are run instead of just models in the specified group (FIXED)
**Problem:** When specifying a group (e.g., "default"), thinktank runs all enabled models from the top-level `models` array instead of just the models in the specified group.

**Root Cause:**  
The issue was in the `getGroup` function in `src/organisms/configManager.ts`. It was returning an empty array for non-default groups that don't exist, and returning top-level models for the default group only when no groups at all were defined. This caused inconsistent behavior.

**Solution:**
Fixed the `getGroup` function to properly return models from the specified group, with a fallback to top-level models:
```typescript
// src/organisms/configManager.ts
export function getGroup(config: AppConfig, groupName: string): ModelConfig[] {
  // If the group exists, return its models
  if (config.groups && config.groups[groupName]) {
    return config.groups[groupName].models;
  }
  
  // If looking for the default group or if the group doesn't exist,
  // return the top-level models array
  return config.models;
}
```

This ensures that:
1. For existing groups, we return the models from that group
2. For the default group or non-existent groups, we return the top-level models array
3. This maintains backward compatibility with existing tests and behavior

## ✅ Issue 2: Missing API key error for Google models despite environment variable being set (FIXED)
**Problem:** Thinktank shows "Missing API keys" error for Google models even when GEMINI_API_KEY is set in the environment.

**Root Cause:**  
The error message was caused by the API key detection logic only checking a single environment variable name pattern based on the provider name. The implementation had three key issues:
1. It didn't check alternate environment variable names (e.g., both GEMINI_API_KEY and GOOGLE_API_KEY)
2. It didn't handle case-insensitive provider matching (e.g., 'Google' vs 'google')
3. The return type was inconsistent (undefined vs null)

**Solution:**
1. Added robust API key mapping in `src/atoms/helpers.ts`:
```typescript
export function getApiKey(config: ModelConfig): string | null {
  // First try the custom environment variable if specified
  if (config.apiKeyEnvVar) {
    const key = process.env[config.apiKeyEnvVar];
    if (key) {
      return key;
    }
  }
  
  // Standard environment variable mappings by provider
  const envVarMappings: Record<string, string[]> = {
    'openai': ['OPENAI_API_KEY'],
    'anthropic': ['ANTHROPIC_API_KEY'],
    'google': ['GEMINI_API_KEY', 'GOOGLE_API_KEY'], // Check multiple possible names
    'openrouter': ['OPENROUTER_API_KEY']
  };
  
  // Handle case-insensitive provider matching
  const provider = config.provider.toLowerCase();
  const possibleVars = envVarMappings[provider] || [`${provider.toUpperCase()}_API_KEY`];
  
  // Try each possible environment variable
  for (const envVar of possibleVars) {
    const key = process.env[envVar];
    if (key) {
      return key;
    }
  }
  
  return null;
}
```

2. Added a diagnostic function to help troubleshoot API key issues:
```typescript
export function debugApiKeyAvailability(): Record<string, boolean> {
  const result: Record<string, boolean> = {};
  
  if (process.env.NODE_ENV === 'development') {
    const keys = [
      'OPENAI_API_KEY', 
      'ANTHROPIC_API_KEY', 
      'GEMINI_API_KEY', 
      'GOOGLE_API_KEY', 
      'OPENROUTER_API_KEY'
    ];
    
    keys.forEach(key => {
      result[key] = !!process.env[key];
    });
  }
  
  return result;
}
```

These changes make the API key detection more robust and user-friendly, while also providing better tools for debugging issues.
```

## Issue 3: Output formatting issues
**Problem:** Output has several formatting issues: 
- Duplicated symbols and emojis
- No progress indicator during processing
- Poor visibility when errors occur

**Root Cause:**  
The issue is in the output formatting logic in `src/molecules/outputFormatter.ts` and possibly in progress reporting in `src/templates/runThinktank.ts`.

**Solution:**
1. Fix the duplicated emoji/symbols issue in `src/molecules/outputFormatter.ts`:
```typescript
// Replace this pattern throughout the file:
console.log(`ℹ ℹ ${message}`);
// With:
console.log(`ℹ ${message}`);

// Similarly for warnings, errors, and other symbols
```

2. Add a progress indicator during processing in `src/templates/runThinktank.ts`:
```typescript
// Before starting model processing:
const spinner = {
  frames: ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'],
  interval: 80,
  currentFrame: 0
};

let spinnerInterval: NodeJS.Timeout | null = null;
let currentModelName: string = '';

// Start the spinner
function startSpinner(modelName: string) {
  currentModelName = modelName;
  spinnerInterval = setInterval(() => {
    process.stdout.write(`\r${spinner.frames[spinner.currentFrame]} Processing ${modelName}...`);
    spinner.currentFrame = (spinner.currentFrame + 1) % spinner.frames.length;
  }, spinner.interval);
}

// Stop the spinner
function stopSpinner(status: 'success' | 'error') {
  if (spinnerInterval) {
    clearInterval(spinnerInterval);
    const icon = status === 'success' ? '✅' : '❌';
    process.stdout.write(`\r${icon} Completed ${currentModelName}    \n`);
    spinnerInterval = null;
  }
}

// Use these functions when processing each model
```

3. Improve error handling and status reporting:
```typescript
// In runThinktank.ts
// When a model fails:
stopSpinner('error');
console.error(`${colors.red('✖')} Error in model ${model.provider}:${model.modelId}: ${formattedError}`);
console.log(`${colors.dim('→')} Continuing with remaining models...\n`);

// When processing completes, show a clear summary:
console.log('\n');
console.log(`${colors.blue('📊')} Results Summary:`);
console.log(`${colors.dim('│')}`);
models.forEach((model, i) => {
  const prefix = `${colors.dim('├')} ${i+1}. `;
  const status = model.error ? `${colors.red('✖')} Failed` : `${colors.green('✓')} Success`;
  console.log(`${prefix}${model.provider}:${model.modelId} - ${status}`);
  if (model.error) {
    console.log(`${colors.dim('│  ')} ${colors.red('→')} ${model.error.message}`);
  }
});
console.log(`${colors.dim('└')} Complete.`);
```

## Issue 4: Temperature validation error with Claude's thinking capability
**Problem:** Claude returned an error about temperature validation when thinking is enabled.

**Root Cause:**  
According to the error message, Claude requires temperature to be set to 1 when thinking is enabled:
```
"temperature` may only be set to 1 when thinking is enabled"
```

**Solution:**
1. Fix the anthropic.ts provider to handle the temperature requirement:
```typescript
// In src/molecules/llmProviders/anthropic.ts
// When thinking is enabled:
if (options?.thinking) {
  const thinkingOpt = options.thinking as ThinkingOptions;
  
  // Force temperature to 1 when thinking is enabled
  const params = {
    ...baseParams,
    temperature: 1, // Override any other temperature value
    thinking: {
      type: 'enabled' as const,
      budget_tokens: thinkingOpt.budget_tokens
    }
  };
  
  response = await client.messages.create(params);
} else {
  // Regular call without thinking
  response = await client.messages.create(baseParams);
}
```

2. Add a warning in the documentation about this limitation:
```markdown
## Claude's Thinking Capability

**Important:** When using Claude's thinking capability, the temperature will automatically be set to 1, regardless of what value you configured. This is a requirement from Anthropic's API.
```

## Implementation Plan

1. Fix the group selection logic first (Issue 1) - highest priority
2. Fix the API key detection for Google models (Issue 2)
3. Implement the fix for Claude's temperature with thinking (Issue 4)
4. Improve the output formatting (Issue 3)

Each fix should be implemented separately and tested before moving to the next one to ensure we don't introduce new issues.