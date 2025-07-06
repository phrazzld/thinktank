#!/bin/bash
# Simple local test runner - run full test suite with coverage

set -e

echo "Running full test suite..."
go test -race -cover ./...

echo ""
echo "To see coverage details, run:"
echo "  go test -coverprofile=coverage.out ./..."
echo "  go tool cover -html=coverage.out"
