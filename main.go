// main.go - Simple entry point for thinktank
package main

import "github.com/phrazzld/thinktank/cmd/thinktank"

func main() {
	// Delegate directly to the new main function in the cmd package
	thinktank.Main()
}
