# Performance Baseline Summary

Generated: Mon Jul  7 20:00:27 PDT 2025
Commit: dae0c7ddc170507d766dbfc759967fc517c479ca
Branch: refactor/carmack-function-decomposition

## Test Results

### Main() Function
- Test path: ./internal/cli
- Results: [main_function.txt](main_function.txt)
- Key metrics: Dry run functionality with file processing baseline

### Execute() Function
- Test path: ./internal/thinktank
- Results: [execute_function.txt](execute_function.txt)
- Key metrics: File writer operations and model processing baseline

### Console Writer Functions
- Test path: ./internal/logutil
- Results: [console_writer.txt](console_writer.txt)
- Key metrics: Model processing and output formatting baseline

### GatherProjectContextWithContext() Function
- Test path: ./internal/fileutil
- Results: [gather_project_context.txt](gather_project_context.txt)
- Key metrics: **Token counting performance baselines:**
  - Small files: 657.7 ns/op, 0 B/op, 0 allocs/op
  - Medium files: 12386 ns/op, 0 B/op, 0 allocs/op
  - Large files: 521937 ns/op, 0 B/op, 0 allocs/op

## Profile Analysis

### CPU Profiles
- Main function: [main_function.cpuprofile](main_function.cpuprofile) - 29k
- Execute function: [execute_function.cpuprofile](execute_function.cpuprofile) - 11k
- Console writer: [console_writer.cpuprofile](console_writer.cpuprofile) - 605 bytes
- GatherProjectContext: [gather_project_context.cpuprofile](gather_project_context.cpuprofile) - empty

### Memory Profiles
- Main function: [main_function.memprofile](main_function.memprofile) - 7.9k
- Execute function: [execute_function.memprofile](execute_function.memprofile) - 8.7k
- Console writer: [console_writer.memprofile](console_writer.memprofile) - 2.8k

## Key Performance Indicators

1. **Token Count Performance**: Established baseline for small/medium/large file processing
2. **Memory Allocation**: Zero allocations for token counting (efficient)
3. **CPU Usage**: Profile data captured for main execution paths
4. **File Processing**: Baseline established for file system operations

## Usage

Compare future benchmark results against these baseline measurements to detect performance regressions during refactoring.

Run comparison with:
```bash
./scripts/benchmark-compare.sh benchmarks/20250707_200027
```

## Success Criteria Met

✅ Performance baselines recorded for Main(), Execute(), console_writer.go, GatherProjectContextWithContext()
✅ CPU profiles generated for key functions
✅ Memory profiles captured
✅ Quantitative metrics established for regression detection
