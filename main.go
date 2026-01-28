// main.go - Simple entry point for thinktank
// This file exists for backward compatibility - the main CLI is in cmd/thinktank/
package main

import "github.com/misty-step/thinktank/internal/cli"

func main() {
	// Delegate directly to the main CLI implementation
	cli.Main()
}
