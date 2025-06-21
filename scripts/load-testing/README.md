# Thinktank Load Testing Framework

This directory contains scripts for load testing the thinktank CLI to validate performance and reliability under various conditions.

## Setup

1. **Build the thinktank binary**:
   ```bash
   go build -o thinktank cmd/thinktank/main.go
   ```

2. **Copy and configure the settings**:
   ```bash
   cp scripts/load-testing/config.sh.example scripts/load-testing/config.sh
   # Edit config.sh with your API keys
   ```

3. **Ensure prerequisites are installed**:
   - `bc` (for timing calculations)
   - Valid API keys for the providers you want to test

## Usage

### Run all tests
```bash
./scripts/load-testing/run-all-load-tests.sh
```

### Run individual scenarios
```bash
./scripts/load-testing/scenarios/1_reliability_test.sh
./scripts/load-testing/scenarios/2_concurrency_test.sh
./scripts/load-testing/scenarios/3_multi_model_test.sh
./scripts/load-testing/scenarios/4_stress_test.sh
```

## Test Scenarios

### 1. Reliability Test (`1_reliability_test.sh`)
- **Purpose**: Validate model reliability under sustained load
- **Method**: Runs multiple consecutive requests to the same model
- **Metrics**: Success rate, consistency of execution times
- **Focus**: Single model stability

### 2. Concurrency Test (`2_concurrency_test.sh`)
- **Purpose**: Test performance scaling with different concurrency levels
- **Method**: Runs multiple models with varying `--max-concurrent` settings
- **Metrics**: Execution time vs concurrency level
- **Focus**: Concurrent processing efficiency

### 3. Multi-Model Test (`3_multi_model_test.sh`)
- **Purpose**: Test realistic multi-provider and synthesis workflows
- **Method**: Combines models from different providers, tests synthesis
- **Metrics**: Cross-provider reliability, synthesis functionality
- **Focus**: Critical user journeys with multiple models

### 4. Stress Test (`4_stress_test.sh`)
- **Purpose**: Stress test rate limiting and error handling
- **Method**: Triggers rate limits, tests partial failure scenarios
- **Metrics**: Graceful degradation, partial success handling
- **Focus**: System resilience under stress

## Test Workloads

The tests use predefined workloads in `lib/workloads/`:

- **Small Project**: Simple Go project for basic functionality testing
- **Large Project**: Complex multi-file project for stress testing
- **Instructions**: Simple and complex prompts for different scenarios

## Output

Test results are written to `tmp/load-testing/` with timestamped directories for each run. Each test scenario produces:

- Individual output files from thinktank
- Execution time measurements
- Success/failure status logs
- Performance metrics summary

## Configuration

Edit `config.sh` to customize:
- API keys for different providers
- Thinktank binary path
- Optional API endpoint overrides for staging/mock environments

## Best Practices

- Run tests against staging environments when possible
- Monitor API usage and costs during testing
- Use appropriate rate limits to avoid hitting provider quotas
- Review logs after each test run for insights

## Troubleshooting

- **Binary not found**: Ensure thinktank is built and path is correct in config.sh
- **API errors**: Verify API keys are valid and have sufficient quota
- **Permission errors**: Ensure scripts are executable (`chmod +x`)
- **Rate limit errors**: Adjust test timing or use lower concurrency levels
