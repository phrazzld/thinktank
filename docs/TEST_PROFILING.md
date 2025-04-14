# Test Profiling Guide

This document explains how to profile tests in the Architect project to identify performance bottlenecks and optimize test execution.

## Running Tests with Profiling

### CI/GitHub Actions

The project includes a dedicated GitHub Actions workflow for test profiling:

1. Navigate to the "Actions" tab in the GitHub repository
2. Click on the "Go CI" workflow
3. Click "Run workflow" dropdown
4. Check the "Run tests with profiling" option
5. Click "Run workflow"

The workflow will run all tests with CPU, memory, and block profiling enabled and upload the profiling data as artifacts.

### Local Profiling

For local profiling, you can run tests with the following flags:

```bash
# CPU profiling
go test ./... -cpuprofile=cpu.prof

# Memory profiling
go test ./... -memprofile=mem.prof

# Block profiling
go test ./... -blockprofile=block.prof

# With parallel execution
go test ./... -cpuprofile=cpu.prof -parallel 4

# Specific package
go test ./internal/integration/... -cpuprofile=cpu.prof -parallel 4
```

## Analyzing Profile Data

To analyze the profile data, use the Go pprof tool:

```bash
# For CPU profiling
go tool pprof cpu.prof

# For memory profiling
go tool pprof mem.prof

# For block profiling
go tool pprof block.prof
```

### Common pprof Commands

Once inside the pprof interactive shell, you can use these commands:

- `top` - Shows the top functions consuming resources
- `top N` - Shows the top N functions
- `list function_name` - Shows source code for a function with profiling data
- `web` - Generates a graph visualization (requires graphviz)
- `png` - Generates a PNG visualization (requires graphviz)
- `pdf` - Generates a PDF visualization (requires graphviz)

### Web Interface

For a more interactive analysis:

```bash
go tool pprof -http=:8080 cpu.prof
```

This will start a local web server on port 8080 with an interactive UI for analyzing the profile data.

## Optimizing Based on Profile Results

When analyzing profiles, look for:

1. Functions with high cumulative times
2. Unexpected memory allocations
3. CPU-intensive operations that could be optimized
4. Contention points in parallel tests

Common optimization strategies include:

- Reducing allocations in hot paths
- Using more efficient data structures
- Caching results instead of recomputing
- Parallelizing independent operations

## Continuous Monitoring

After making optimizations based on profiling data, run the profiling workflow again to verify improvements. Document significant optimizations and their impact on test execution time and resource consumption.