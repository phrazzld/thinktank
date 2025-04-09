/**
 * Formatter for completion summary data
 */
import { CompletionSummaryData, PureCompletionSummaryResult } from '../workflow/types';
import { colors, styleSuccess, styleError, styleDim } from './consoleUtils';

export interface FormatOptions {
  useColors?: boolean;
  includeMetadata?: boolean;
}

const DEFAULT_FORMAT_OPTIONS: FormatOptions = {
  useColors: true,
  includeMetadata: false,
};

/**
 * Formats completion summary data into a user-friendly display format
 *
 * @param data - The raw completion summary data
 * @param options - Format options
 * @returns The formatted completion summary result containing summary text and error details
 */
export function formatCompletionSummary(
  data: CompletionSummaryData,
  options: FormatOptions = {}
): PureCompletionSummaryResult {
  // Merge with default options
  const opts = { ...DEFAULT_FORMAT_OPTIONS, ...options };
  const { useColors } = opts;

  const {
    totalModels,
    successCount,
    failureCount,
    errors,
    runName,
    outputDirectoryPath,
    totalExecutionTimeMs,
  } = data;

  const lines: string[] = [];
  const errorDetails: string[] = [];

  // Build summary message
  const completionMessage = runName
    ? `'${runName}' (${totalModels} models)`
    : `${totalModels} models`;
  const timeText = totalExecutionTimeMs ? ` in ${(totalExecutionTimeMs / 1000).toFixed(2)}s` : '';

  // Status message based on success/failure ratio
  if (failureCount === 0 && totalModels > 0) {
    lines.push(
      useColors
        ? styleSuccess(`✓ Successfully completed ${completionMessage}${timeText}`)
        : `Successfully completed ${completionMessage}${timeText}`
    );
  } else if (successCount > 0) {
    const percentage = Math.round((successCount / totalModels) * 100);
    lines.push(
      useColors
        ? styleError(
            `⚠ Partially completed ${completionMessage}${timeText} - ${percentage}% success (${successCount}/${totalModels})`
          )
        : `Partially completed ${completionMessage}${timeText} - ${percentage}% success (${successCount}/${totalModels})`
    );
  } else if (totalModels > 0) {
    lines.push(
      useColors
        ? styleError(`✗ All models failed for ${completionMessage}${timeText}`)
        : `All models failed for ${completionMessage}${timeText}`
    );
  } else {
    lines.push(
      useColors
        ? styleDim(`! No models were queried for ${completionMessage}${timeText}`)
        : `No models were queried for ${completionMessage}${timeText}`
    );
  }

  // Add output directory info
  lines.push(
    useColors
      ? styleDim(`+ Output saved to: ${outputDirectoryPath}`)
      : `Output saved to: ${outputDirectoryPath}`
  );

  // Add detailed errors if any
  if (failureCount > 0) {
    // Add a heading for error details
    errorDetails.push(useColors ? colors.red.bold('\nFailed Models:') : '\nFailed Models:');

    // Group errors by category if available
    const errorsByCategory: Record<string, Array<{ modelKey: string; message: string }>> = {};

    // Add to errorsByCategory
    errors.forEach(err => {
      const category = err.category || 'Unknown';
      if (!errorsByCategory[category]) {
        errorsByCategory[category] = [];
      }
      errorsByCategory[category].push({
        modelKey: err.modelKey,
        message: err.message,
      });
    });

    // Format grouped errors
    Object.entries(errorsByCategory).forEach(([category, categoryErrors]) => {
      // Add category header if there are multiple categories and more than one error
      if (Object.keys(errorsByCategory).length > 1) {
        errorDetails.push(
          useColors ? colors.red(`  ${category} errors:`) : `  ${category} errors:`
        );
      }

      // Add each error under this category
      categoryErrors.forEach(err => {
        const errorLine = `  - ${err.modelKey}: ${err.message}`;
        errorDetails.push(useColors ? colors.red(errorLine) : errorLine);
      });
    });
  }

  return {
    summaryText: lines.join('\n'),
    errorDetails: errorDetails.length > 0 ? errorDetails : undefined,
  };
}
