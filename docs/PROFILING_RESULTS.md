# Test Profiling Results

This document contains the analysis of performance hotspots identified through profiling the Architect test suite.

## Summary of Key Hotspots

Based on the profiling data collected from unit tests, integration tests, and benchmarks, the following hotspots have been identified:

1. **Token Counting (High CPU Usage)**: The `estimateTokenCount` function consumes over 51% of CPU time in benchmarks
2. **Character Processing (High CPU Usage)**: The `unicode.IsSpace` function consumes 45% of CPU time in benchmarks
3. **Git Operations (High Memory)**: `isGitIgnored` allocates significant memory via `SlicePtrFromStrings` (50% of memory in fileutil tests)
4. **File Operations (I/O Bottleneck)**: `os.ReadFile`, `os.WriteFile`, and `path/filepath.WalkDir` consume significant resources
5. **External Command Execution (High Resource Usage)**: `os/exec.Command` and related functions show up in memory profiles

## Detailed Analysis

### 1. Token Counting Performance

**Function:** `fileutil.estimateTokenCount`  
**Resource Consumption:** 51.75% of CPU in benchmarks  
**Location:** `/internal/fileutil/fileutil.go`

**Context:**
The token counting function is used to estimate the number of tokens in a text string. This is critical for API calls where token limits need to be enforced. The benchmark shows that for large texts (40,000 characters), it takes about 94μs per operation.

**Root Cause:**
Analysis shows that the function is spending most of its time in character processing via `unicode.IsSpace`, which accounts for 45.39% of CPU time. The current implementation likely iterates through each character in the string and checks if it's a whitespace character.

**Recommendation:**
1. Consider a more efficient counting algorithm that scans the string only once
2. Use faster string manipulation functions from the standard library
3. For large texts, consider sampling techniques or more approximate estimations
4. Potential to add caching for repeated calls with the same text

### 2. Git Integration

**Function:** `fileutil.isGitIgnored`  
**Resource Consumption:** 50.55% of memory allocations in fileutil tests  
**Location:** `/internal/fileutil/fileutil.go`

**Context:**
The `isGitIgnored` function checks if a file should be ignored based on Git's ignore rules. This is used during project context gathering to filter out irrelevant files. It executes an external Git command.

**Root Cause:**
The function allocates significant memory through `syscall.SlicePtrFromStrings` (4716.84kB), which converts Go string slices to C-style string arrays for process execution. These allocations occur when spawning Git processes via `os/exec.Command`.

**Recommendation:**
1. Cache Git-ignored status for paths within the same test run
2. Consider using a Go-native Git ignore implementation instead of spawning Git processes
3. Batch Git ignore checks to reduce the number of process spawns
4. Investigate if the entire Git ignore functionality can be optional for tests

### 3. File Operations

**Function:** `fileutil.GatherProjectContext` and related functions  
**Resource Consumption:** Significant memory allocations in both unit and integration tests  
**Location:** `/internal/fileutil/fileutil.go`

**Context:**
File gathering and processing functions are central to the Architect tool, which processes project files to generate context for AI models.

**Root Cause:**
- Directory traversal via `path/filepath.WalkDir` involves numerous small allocations
- File reading via `os.ReadFile` allocates memory for file contents
- String splitting during path processing allocates memory

**Recommendation:**
1. Use buffered I/O when appropriate to reduce allocations
2. Consider a custom, more efficient directory walker
3. Implement a path cache to avoid redundant path processing
4. For tests: Use smaller file sets or mock the filesystem

### 4. Integration Test Memory Usage

**Function:** Various (regexp compilation, JSON marshaling)  
**Resource Consumption:** ~20% of memory in integration tests  
**Location:** Integration test framework

**Context:**
Integration tests use regular expressions for matching and validation, along with JSON processing for model responses.

**Root Cause:**
- Regular expression compilation (`regexp.Compile`) allocates 19.85% of memory
- JSON encoding/decoding (`encoding/json`) accounts for ~10% of memory allocations

**Recommendation:**
1. Pre-compile regular expressions where possible (they're currently being recompiled in tests)
2. Reuse JSON encoder/decoder instances rather than creating new ones
3. Consider using lighter validation mechanisms than full regexp where appropriate

### 5. Multi-Model Processing

**Function:** `orchestrator.processModelWithRateLimit`  
**Resource Consumption:** Shows up in CPU profiles of integration tests  
**Location:** `/internal/architect/orchestrator/orchestrator.go`

**Context:**
The orchestrator manages concurrent processing of multiple AI models, handling rate limiting and synchronization.

**Root Cause:**
While not consuming a large percentage in our limited profiles, this area could become more critical with more complex scenarios. The current implementation spawns goroutines for each model.

**Recommendation:**
1. Consider a worker pool pattern for model processing instead of spawning goroutines for each model
2. Implement better timeout handling to avoid goroutine leaks
3. Add metrics to track concurrency performance in real usage

## Action Items

Based on the analysis, here are the recommended optimization tasks, ordered by priority:

1. **Optimize Token Counting**
   - Refactor `estimateTokenCount` to use more efficient string processing
   - Explore caching mechanisms for repeated token counting operations
   - Consider using specialized text processing libraries

2. **Improve Git Integration**
   - Implement a caching layer for Git ignore status
   - Investigate Go-native Git ignore rule processing to eliminate process spawning
   - Add configuration to disable Git integration in tests when not needed

3. **Enhance File Processing**
   - Implement buffered I/O for file reading operations
   - Add path caching to reduce repetitive path processing
   - Create test-specific file gatherers that avoid disk I/O

4. **Optimize Regular Expression Usage**
   - Pre-compile and reuse regular expressions in the test framework
   - Consider replacing regexp with simpler string operations where appropriate

5. **Improve Concurrency Management**
   - Implement worker pools for model processing
   - Add better resource cleanup for goroutines
   - Enhance monitoring of concurrent operations

## Next Steps

1. Implement the highest priority optimizations (token counting, Git integration)
2. Run follow-up profiling to measure improvements
3. Iterate on the remaining action items based on the updated profiles
4. Document performance best practices for future development