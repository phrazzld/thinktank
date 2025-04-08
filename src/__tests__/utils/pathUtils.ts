/**
 * Path normalization utilities for test files.
 * 
 * @deprecated Use src/utils/pathUtils.ts utilities instead. This module is
 * maintained for backward compatibility with existing tests.
 */
import { normalizePathGeneral } from '../../utils/pathUtils';

/**
 * Normalizes a path to use forward slashes consistently and handles leading slashes 
 * according to project requirements.
 * 
 * @deprecated Use normalizePathGeneral from src/utils/pathUtils.ts instead
 * 
 * @param inputPath - The path to normalize
 * @param keepLeadingSlash - Whether to preserve leading slash (default: false)
 * @returns The normalized path
 */
export function normalizePath(inputPath: string, keepLeadingSlash = false): string {
  return normalizePathGeneral(inputPath, keepLeadingSlash);
}
