#!/usr/bin/env node
/**
 * Script to ensure all source files end with a newline character
 */
const fs = require('fs');
const path = require('path');
const { promisify } = require('util');
const readFile = promisify(fs.readFile);
const writeFile = promisify(fs.writeFile);
const readdir = promisify(fs.readdir);
const stat = promisify(fs.stat);

// Extensions to process
const EXTENSIONS = ['.ts', '.js', '.json', '.md'];

// Directories to exclude
const EXCLUDE_DIRS = ['node_modules', 'dist', '.git'];

// Tracks changes
let filesFixed = 0;
let filesAlreadyCorrect = 0;
let filesSkipped = 0;

/**
 * Check if file needs a newline at the end
 */
async function needsNewline(filePath) {
  try {
    const content = await readFile(filePath, 'utf8');
    
    // Skip empty files
    if (content.length === 0) {
      return false;
    }
    
    return content.charAt(content.length - 1) !== '\n';
  } catch (error) {
    console.error(`Error reading ${filePath}:`, error.message);
    return false;
  }
}

/**
 * Add a newline to the end of a file if it's missing
 */
async function addNewline(filePath) {
  try {
    const content = await readFile(filePath, 'utf8');
    
    // Skip empty files
    if (content.length === 0) {
      filesSkipped++;
      return;
    }
    
    if (content.charAt(content.length - 1) !== '\n') {
      await writeFile(filePath, content + '\n');
      console.log(`✅ Added newline to ${filePath}`);
      filesFixed++;
    } else {
      filesAlreadyCorrect++;
    }
  } catch (error) {
    console.error(`❌ Error fixing ${filePath}:`, error.message);
    filesSkipped++;
  }
}

/**
 * Process a file if it matches our criteria
 */
async function processFile(filePath) {
  const ext = path.extname(filePath);
  
  if (EXTENSIONS.includes(ext)) {
    if (await needsNewline(filePath)) {
      await addNewline(filePath);
    } else {
      filesAlreadyCorrect++;
    }
  } else {
    filesSkipped++;
  }
}

/**
 * Recursively walk a directory and process files
 */
async function processDirectory(dirPath) {
  try {
    const entries = await readdir(dirPath, { withFileTypes: true });
    
    for (const entry of entries) {
      const fullPath = path.join(dirPath, entry.name);
      
      if (entry.isDirectory()) {
        // Skip excluded directories
        if (!EXCLUDE_DIRS.includes(entry.name)) {
          await processDirectory(fullPath);
        }
      } else if (entry.isFile()) {
        await processFile(fullPath);
      }
    }
  } catch (error) {
    console.error(`Error processing directory ${dirPath}:`, error.message);
  }
}

/**
 * Main execution function
 */
async function main() {
  console.log('🔍 Checking for files missing newlines...');
  
  // Process from project root
  const rootDir = path.resolve(__dirname, '..');
  await processDirectory(rootDir);
  
  // Print summary
  console.log('\n📊 Summary:');
  console.log(`  ✅ ${filesFixed} files fixed`);
  console.log(`  ✓ ${filesAlreadyCorrect} files already had newlines`);
  console.log(`  ⏩ ${filesSkipped} files skipped`);
  
  if (filesFixed > 0) {
    console.log('\n🎉 All files now have trailing newlines!');
  } else if (filesAlreadyCorrect > 0 && filesFixed === 0) {
    console.log('\n🎉 All files already had trailing newlines!');
  }
}

// Execute script
main().catch(err => {
  console.error('Error executing script:', err);
  process.exit(1);
});
