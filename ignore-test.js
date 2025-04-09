/**
 * Simple test of the ignore library's behavior with complex patterns
 */
const ignore = require('ignore');

function testPattern(pattern, paths, description) {
  console.log(`\n=== Testing ${description} ===`);
  console.log(`Pattern: ${pattern}`);

  const ig = ignore();
  ig.add(pattern);

  for (const path of paths) {
    const isIgnored = ig.ignores(path);
    console.log(`${path}: ${isIgnored ? 'IGNORED' : 'NOT IGNORED'}`);
  }
}

// Test 1: Double-asterisk pattern
testPattern('**/*.js', [
  'file.js',
  'src/file.js',
  'src/nested/file.js',
  'src/file.txt'
], 'Double-asterisk pattern');

// Test 2: Brace expansion pattern
testPattern('*.{jpg,png}', [
  'image.jpg',
  'image.png',
  'image.gif',
  'document.txt'
], 'Brace expansion pattern');

// Test 3: Prefix wildcard pattern for directories
testPattern('build-*/', [
  'build-output/file.txt',
  'build-debug/file.txt',
  'building/file.txt',
  'other-dir/file.txt'
], 'Prefix wildcard pattern');

// Test 4: Negated patterns
testPattern('*.log\n!important/*.log', [
  'debug.log',
  'important/critical.log',
  'other/debug.log'
], 'Negated patterns');

// Test 5: Character range pattern
testPattern('[0-9]*.js', [
  '1script.js',
  '2script.js',
  'script.js',
  'ascript.js'
], 'Character range pattern');
