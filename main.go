// main.go - Simple entry point for architect
package main

import "github.com/phrazzld/architect/cmd/architect"

func main() {
	// Delegate directly to the new main function in the cmd package
	architect.Main()
}