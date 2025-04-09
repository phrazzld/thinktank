/**
 * Script to audit test files for deprecated fs mocking approaches
 * Generates a table of files needing migration to virtualFs
 */
const glob = require('fast-glob');
const fs = require('fs').promises;
const path = require('path');

// Define search patterns
const patterns = {
  directMock: {
    fs: /jest\.mock\(['"]fs['"]/,
    fsPromises: /jest\.mock\(['"]fs\/promises['"]/
  },
  legacyUtil: {
    import: /from ['"].*mockFsUtils['"]/,
    usage: /\b(mockFs|setupMockFs|mockReadFile|mockWriteFile)\b/
  },
  virtualFs: {
    import: /from ['"].*(?:test\/setup\/fs|virtualFsUtils)['"]/,
    usage: /\b(setupBasicFs|createVirtualFs|addVirtualGitignoreFile|normalizePathForMemfs)\b/
  }
};

// Export patterns for testing
module.exports = { patterns };

/**
 * Analyzes a single file for filesystem mocking patterns
 * 
 * @param {string} filePath - The path to the file
 * @param {string} content - The content of the file
 * @returns {object} - Analysis result object
 */
function analyzeFile(filePath, content) {
  // Check for various patterns
  const hasDirectFsMock = patterns.directMock.fs.test(content) || 
                           patterns.directMock.fsPromises.test(content);
                            
  const hasLegacyUtil = patterns.legacyUtil.import.test(content) || 
                        patterns.legacyUtil.usage.test(content);
                         
  const hasVirtualFs = patterns.virtualFs.import.test(content) || 
                        patterns.virtualFs.usage.test(content);
    
  // Categorize based on found patterns
  let category = 'None';
    
  if (hasDirectFsMock && !hasLegacyUtil && !hasVirtualFs) {
    category = 'Direct Mock';
  } else if (hasLegacyUtil && !hasVirtualFs) {
    category = 'Legacy Util';
  } else if ((hasDirectFsMock || hasLegacyUtil) && hasVirtualFs) {
    category = 'Mixed';
  } else if (hasVirtualFs) {
    category = 'Virtual FS';
  }
    
  return {
    filePath,
    category,
    hasDirectFsMock,
    hasLegacyUtil,
    hasVirtualFs
  };
}

/**
 * Generates markdown report from analysis results
 * 
 * @param {Array} results - The analysis results
 * @returns {string} - Markdown content
 */
function generateMarkdownReport(results) {
  // Generate markdown table
  let markdown = '# Filesystem Mocking Audit Results\n\n';
  markdown += 'This document catalogs test files using deprecated filesystem mocking patterns.\n\n';
  markdown += '## Migration Targets\n\n';
  markdown += '* Replace `jest.mock(\'fs\')` with helpers from `test/setup/fs.ts` like `setupBasicFs`\n';
  markdown += '* Replace imports from `mockFsUtils.ts` with the new virtual filesystem approach\n\n';
  markdown += '## Files Needing Migration\n\n';
  markdown += '| File Path | Category | Complexity | Notes | Priority |\n';
  markdown += '|-----------|----------|------------|-------|----------|\n';
  
  // Add each result to the table
  results.forEach(result => {
    markdown += `| ${result.filePath} | ${result.category} | <!-- TODO: ASSESS --> | <!-- TODO: ADD NOTES --> | <!-- TODO: ASSIGN --> |\n`;
  });
  
  markdown += '\n\n## Categories:\n';
  markdown += '* **Direct Mock:** Uses `jest.mock(\'fs\')` or `jest.mock(\'fs/promises\')`\n';
  markdown += '* **Legacy Util:** Imports/uses `mockFsUtils.ts`\n';
  markdown += '* **Mixed:** Uses both deprecated patterns and potentially some new helpers\n\n';
  markdown += '## Complexity Guidelines (Fill in manually):\n';
  markdown += '* **Low:** Few mock interactions, simple test logic that can be easily migrated\n';
  markdown += '* **Medium:** Moderate number of mock interactions, some complexity in test logic\n';
  markdown += '* **High:** Many mock interactions, complex setup, deep integration with the mocking logic\n\n';
  markdown += '## Priority Guidelines (Fill in manually):\n';
  markdown += '* 1 (Highest): Start with these files (Low complexity, high impact)\n';
  markdown += '* 2: Second wave of migration\n';
  markdown += '* 3: Lower priority files\n\n';
  
  return markdown;
}

/**
 * Main function to audit fs mocks in test files
 * 
 * @param {object} options - Options for the audit
 * @param {function} logger - Logger function (defaults to console.log)
 * @param {object} dependencies - Optional dependencies for testing
 * @returns {object} - Audit results and markdown report
 */
async function auditFsMocks(options = {}, logger = console.log, dependencies = {}) {
  try {
    logger('Scanning for test files...');
    
    // Set default options
    const defaultOptions = {
      patterns: ['src/**/*.test.ts', 'test/**/*.test.ts'],
      ignore: ['**/node_modules/**', '**/dist/**', '**/examples/**'],
      outputPath: 'FS_MOCK_AUDIT.md'
    };
    
    const config = { ...defaultOptions, ...options };
    
    // Use injected dependencies or defaults
    const fileFinder = dependencies.glob || glob;
    const fileReader = dependencies.readFile || fs.readFile;
    const fileWriter = dependencies.writeFile || fs.writeFile;
    
    // Find all test files
    const files = await fileFinder(config.patterns, { ignore: config.ignore });
    
    logger(`Found ${files.length} test files. Analyzing...`);
    
    // Records of categorized files
    const results = [];
    
    // Process each file
    for (const file of files) {
      // Skip actually reading the file during testing if content is provided
      let content;
      if (dependencies.getContent) {
        content = dependencies.getContent(file);
      } else {
        const filePath = path.resolve(file);
        content = await fileReader(filePath, 'utf8');
      }
      
      const analysis = analyzeFile(file, content);
      
      // Only add to results if it uses deprecated patterns
      if (analysis.category !== 'None' && analysis.category !== 'Virtual FS') {
        results.push(analysis);
      }
    }
    
    // Sort results by category priority (Legacy Util first, Mixed second, Direct Mock third)
    results.sort((a, b) => {
      const priority = { 'Legacy Util': 0, 'Mixed': 1, 'Direct Mock': 2 };
      return priority[a.category] - priority[b.category];
    });
    
    // Generate markdown report
    const markdown = generateMarkdownReport(results);
    
    // Write the results to file if outputPath is provided
    if (config.outputPath) {
      await fileWriter(config.outputPath, markdown);
      logger(`Audit complete! Found ${results.length} files using deprecated fs mocking.`);
      logger(`Results written to ${config.outputPath}`);
    }
    
    return { results, markdown };
  } catch (err) {
    console.error('Error during audit:', err);
    throw err;
  }
}

// Export for testing
module.exports = { 
  patterns,
  analyzeFile,
  generateMarkdownReport,
  auditFsMocks
};

// Run the audit if this script is executed directly
if (require.main === module) {
  auditFsMocks();
}
