// Package main provides the command-line interface for the thinktank tool
package main

import (
	"github.com/misty-step/thinktank/internal/cli"
)

// All error handling functions are now in internal/cli package to eliminate duplication

// main is the entry point for the Go runtime
func main() {
	cli.Main()
}
