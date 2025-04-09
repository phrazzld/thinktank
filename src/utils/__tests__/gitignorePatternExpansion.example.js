/**
 * Example of using the expandBracePattern helper to handle complex patterns
 * 
 * This file demonstrates how to use the expandBracePattern helper to handle
 * brace expansion patterns with the ignore library.
 */
const ignore = require('ignore');
const { expandBracePattern } = require('../gitignoreUtils');

/**
 * Creates an ignore filter from gitignore content with support for brace expansion
 * 
 * @param {string} gitignoreContent - Content of a .gitignore file, potentially with brace patterns
 * @returns {object} An ignore filter that handles brace expansion patterns
 */
function createIgnoreFilterWithBraceSupport(gitignoreContent) {
  const ignoreFilter = ignore();
  
  // Split content into lines
  const lines = gitignoreContent.split('\n');
  
  // Process each line
  for (const line of lines) {
    // Skip empty lines and comments
    if (!line || line.startsWith('#')) {
      continue;
    }
    
    // Check if the line contains a brace pattern
    if (line.includes('{') && line.includes('}')) {
      // Use our helper to expand the brace pattern
      const expandedPatterns = expandBracePattern(line);
      
      // Add all expanded patterns to the filter
      ignoreFilter.add(expandedPatterns);
    } else {
      // Regular pattern, add directly
      ignoreFilter.add(line);
    }
  }
  
  return ignoreFilter;
}

// Example usage
console.log('=== Example: Using expandBracePattern for .gitignore processing ===');

// Sample gitignore content with brace expansion patterns
const gitignoreContent = `
# Ignore common build directories
build*/
dist/

# Ignore common image formats
*.{jpg,png,gif}

# Ignore logs but keep important ones
*.log
!important/*.log

# Ignore numbered temp files
temp-[0-9].txt
`;

console.log('Processing gitignore content:');
console.log(gitignoreContent);

// Create an ignore filter with brace expansion support
const filter = createIgnoreFilterWithBraceSupport(gitignoreContent);

// Test some paths
const testPaths = [
  'document.txt',
  'image.jpg',
  'image.png',
  'image.gif',
  'image.svg',
  'build/output.js',
  'build-debug/output.js',
  'dist/bundle.js',
  'app.log',
  'important/critical.log',
  'temp-1.txt',
  'temp-a.txt'
];

console.log('\nResults:');
for (const path of testPaths) {
  const isIgnored = filter.ignores(path);
  console.log(`${path}: ${isIgnored ? 'IGNORED' : 'NOT IGNORED'}`);
}

// Explain the key takeaway
console.log('\nKey takeaway:');
console.log('By using the expandBracePattern helper, we can support brace expansion patterns');
console.log('like "*.{jpg,png,gif}" that would otherwise not work with the ignore library.');
