# Benchmarks

This directory contains performance benchmarking tools and baseline data for the thinktank project.

## Overview

The benchmarking infrastructure supports:
- Performance baseline generation
- Regression detection (configurable threshold)
- CPU and memory profiling
- Comparative performance analysis

## Directory Structure

```
benchmarks/
├── README.md                    # This file
├── 20250707_200027/            # Historical baseline (reference only)
│   └── baseline_summary.md     # Performance characteristics summary
└── baseline/                   # Current baseline directory (gitignored)
    ├── *.txt                   # Benchmark output files
    ├── *.cpuprofile           # CPU profiles
    └── *.memprofile           # Memory profiles
```

## Usage

### Generating Performance Baselines

To create a new performance baseline:

```bash
./scripts/benchmark-before.sh
```

This will:
1. Run all benchmarks across the codebase
2. Generate CPU and memory profiles
3. Save results to a timestamped directory
4. Create a summary report

### Running Regression Checks

To check for performance regressions:

```bash
# Using default baseline location (benchmarks/baseline)
./scripts/performance-regression-check.sh

# Using custom baseline location
THINKTANK_BASELINE_DIR=benchmarks/20250707_200027 ./scripts/performance-regression-check.sh

# Using custom regression threshold (default: 5%)
THINKTANK_REGRESSION_THRESHOLD=10 ./scripts/performance-regression-check.sh
```

### Environment Variables

- `THINKTANK_BASELINE_DIR`: Path to baseline benchmarks (default: `benchmarks/baseline`)
- `THINKTANK_REGRESSION_THRESHOLD`: Regression threshold percentage (default: 5)

### Continuous Integration

The regression check is automatically run in CI to prevent performance degradation. If a regression is detected:

1. Review the specific benchmarks that regressed
2. If the regression is expected (e.g., due to new features), update the baseline
3. If unexpected, investigate and fix the performance issue

### Updating Baselines

When performance changes are intentional:

```bash
# Remove old baseline
rm -rf benchmarks/baseline

# Generate new baseline
./scripts/benchmark-before.sh

# Move to baseline location
mv benchmarks/$(ls -t benchmarks/ | head -1) benchmarks/baseline
```

## Key Benchmarks Monitored

- **Token Counting**: Small/Medium/Large file processing
- **File Processing**: Binary detection, filtering operations
- **Core Functions**: Main(), Execute(), console writer operations
- **Memory Allocations**: Per-operation allocation tracking

## Performance Targets

Based on the Carmack-style refactoring (2025-07-08):
- Token counting: 50-79% improvement achieved
- File processing: 19-53% improvement achieved
- Binary detection: 21% improvement achieved
- Zero performance regressions allowed in CI

## Historical Reference

The `20250707_200027` directory contains the baseline summary from the initial Carmack-style refactoring, documenting:
- 90.4% test coverage achievement
- 19-79% performance improvements
- Zero regression validation

For detailed performance analysis, see [baseline_summary.md](20250707_200027/baseline_summary.md).
