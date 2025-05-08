//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("hello world")
	fmt.Println("slog package exists")
}
