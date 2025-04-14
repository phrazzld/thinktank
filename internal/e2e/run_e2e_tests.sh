#!/bin/bash

# Script to run the E2E tests for Architect CLI

# Display usage instructions
function show_usage {
  echo "Usage: $0 [options]"
  echo
  echo "Options:"
  echo "  -h, --help         Show this help message and exit"
  echo "  -v, --verbose      Run tests with verbose output"
  echo "  -s, --short        Run tests in short mode"
  echo "  -r, --run PATTERN  Run only tests matching the specified pattern"
  echo "  -k, --key KEY      Use the specified API key (default: test-api-key)"
  echo
  echo "Examples:"
  echo "  $0                   Run all E2E tests"
  echo "  $0 -v                Run all E2E tests with verbose output"
  echo "  $0 -r TestBasic      Run only tests with 'TestBasic' in their name"
  echo "  $0 -v -r TestToken   Run only tests with 'TestToken' in their name with verbose output"
}

# Get script directory for relative paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Default options
VERBOSE=""
TEST_PATTERN=""
SHORT=""
API_KEY="test-api-key"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      show_usage
      exit 0
      ;;
    -v|--verbose)
      VERBOSE="-v"
      shift
      ;;
    -s|--short)
      SHORT="-short"
      shift
      ;;
    -r|--run)
      if [[ -z "$2" || "$2" == -* ]]; then
        echo "Error: --run requires a test pattern"
        show_usage
        exit 1
      fi
      TEST_PATTERN="-run=$2"
      shift 2
      ;;
    -k|--key)
      if [[ -z "$2" || "$2" == -* ]]; then
        echo "Error: --key requires an API key"
        show_usage
        exit 1
      fi
      API_KEY="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      show_usage
      exit 1
      ;;
  esac
done

# Print header
echo
echo "=== Running Architect E2E Tests ==="
echo

# Set environment to make tests run reliably
export ARCHITECT_DEBUG=true
export GEMINI_API_KEY_SOURCE=env
export GEMINI_API_KEY="$API_KEY"

# Build the command
TEST_CMD="go test -tags=manual_api_test $VERBOSE $SHORT $TEST_PATTERN $SCRIPT_DIR/..."

# Echo the command for transparency
echo "Running: $TEST_CMD"
echo "API Key: ${API_KEY:0:3}...${API_KEY: -3}"
echo

# Run the tests
cd "$PROJECT_ROOT" && eval "$TEST_CMD"

# Get the exit code
EXIT_CODE=$?

# Print footer
echo
if [ $EXIT_CODE -eq 0 ]; then
  echo "=== E2E Tests Passed Successfully ==="
else
  echo "=== E2E Tests Failed with Exit Code $EXIT_CODE ==="
fi
echo

exit $EXIT_CODE
