// main.go - Simple entry point for thinktank
// This file exists for backward compatibility - the main CLI is in cmd/thinktank/
package main

import (
	"os"
	"os/exec"
)

func main() {
	// Simply execute the main thinktank binary from cmd/thinktank
	cmd := exec.Command("go", append([]string{"run", "./cmd/thinktank"}, os.Args[1:]...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}
}
