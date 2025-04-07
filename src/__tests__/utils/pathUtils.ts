import path from 'path';

/**
 * Normalizes a path to use forward slashes consistently and handles leading slashes 
 * according to project requirements.
 * 
 * @param inputPath - The path to normalize
 * @param keepLeadingSlash - Whether to preserve leading slash (default: false)
 * @returns The normalized path
 */
export function normalizePath(inputPath: string, keepLeadingSlash = false): string {
  // Replace any backslashes with forward slashes
  let normalized = inputPath.replace(/\\/g, '/');
  
  // Normalize path using path.posix for consistent forward slashes
  normalized = path.posix.normalize(normalized);
  
  // Handle specific case for ./
  if (normalized === './') {
    normalized = '.';
  }

  // Handle leading slash based on parameter
  if (!keepLeadingSlash && normalized.startsWith('/')) {
    normalized = normalized.substring(1);
  } else if (keepLeadingSlash && !normalized.startsWith('/')) {
    normalized = '/' + normalized;
  }

  return normalized;
}
