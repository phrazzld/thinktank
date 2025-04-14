#!/bin/bash
# profile_tests.sh - Generate and perform initial analysis of test profiles

set -e

# Default values
PACKAGE="./..."
OUTPUT_DIR="./profile_data"
PARALLEL=4
INTERACTIVE=false
TEST_ARGS=""

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --package=*)
      PACKAGE="${1#*=}"
      shift
      ;;
    --output=*)
      OUTPUT_DIR="${1#*=}"
      shift
      ;;
    --parallel=*)
      PARALLEL="${1#*=}"
      shift
      ;;
    --interactive)
      INTERACTIVE=true
      shift
      ;;
    --test-args=*)
      TEST_ARGS="${1#*=}"
      shift
      ;;
    *)
      echo "Unknown parameter: $1"
      exit 1
      ;;
  esac
done

# Ensure output directory exists
mkdir -p "$OUTPUT_DIR"

echo "Running tests with profiling for package: $PACKAGE"
echo "Profiles will be saved to: $OUTPUT_DIR"

# Run CPU profiling
echo "Generating CPU profile..."
go test $PACKAGE -cpuprofile="$OUTPUT_DIR/cpu.prof" -parallel=$PARALLEL $TEST_ARGS

# Run memory profiling (separate run to avoid interference)
echo "Generating memory profile..."
go test $PACKAGE -memprofile="$OUTPUT_DIR/mem.prof" -parallel=$PARALLEL $TEST_ARGS

# Run block profiling (useful for concurrency bottlenecks)
echo "Generating block profile..."
go test $PACKAGE -blockprofile="$OUTPUT_DIR/block.prof" -parallel=$PARALLEL $TEST_ARGS

# Generate basic text reports
echo "Generating CPU profile report..."
go tool pprof -top "$OUTPUT_DIR/cpu.prof" > "$OUTPUT_DIR/cpu_top.txt"

echo "Generating memory allocation report..."
go tool pprof -top "$OUTPUT_DIR/mem.prof" > "$OUTPUT_DIR/mem_top.txt"

# Check if graphviz is installed for PNG generation
if command -v dot &> /dev/null; then
  echo "Generating CPU profile graph..."
  go tool pprof -png "$OUTPUT_DIR/cpu.prof" > "$OUTPUT_DIR/cpu_graph.png"

  echo "Generating memory allocation graph..."
  go tool pprof -png "$OUTPUT_DIR/mem.prof" > "$OUTPUT_DIR/mem_graph.png"
else
  echo "Graphviz not installed. Skipping graph generation."
  echo "Install Graphviz to enable graph visualization."
fi

echo "Profile reports generated in $OUTPUT_DIR"
echo "To analyze interactively, run:"
echo "  go tool pprof $OUTPUT_DIR/cpu.prof"
echo "  go tool pprof $OUTPUT_DIR/mem.prof"

# Launch interactive mode if requested
if [ "$INTERACTIVE" = true ]; then
  echo "Launching interactive CPU profile analysis..."
  go tool pprof "$OUTPUT_DIR/cpu.prof"
fi
