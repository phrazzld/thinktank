// Package testutil provides test validation utilities for parallel execution
package testutil

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCPUBoundTestsAreParallel validates that CPU-bound tests use t.Parallel()
// This test defines our requirements for parallel test execution optimization
func TestCPUBoundTestsAreParallel(t *testing.T) {
	// Define packages that should have parallel CPU-bound tests
	cpuBoundPackages := []string{
		"cmd/thinktank",
		"internal/models",
		"internal/config",
		"internal/llm",
	}

	// Define test patterns that indicate CPU-bound tests
	cpuBoundPatterns := []string{
		"TestParseFlags",
		"TestValidate",
		"TestGetModel",
		"TestIsCategory",
		"TestWrap",
		"TestNew",
	}

	for _, pkg := range cpuBoundPackages {
		t.Run(pkg, func(t *testing.T) {
			// Get current working directory to build proper paths
			wd, err := os.Getwd()
			if err != nil {
				t.Skipf("Could not get working directory: %v", err)
				return
			}

			// Build path relative to project root (go up from internal/testutil)
			projectRoot := filepath.Join(wd, "..", "..")
			pkgPath := filepath.Join(projectRoot, pkg)

			// Find test files in package
			testFiles, err := filepath.Glob(filepath.Join(pkgPath, "*_test.go"))
			if err != nil {
				t.Skipf("Could not find test files in %s: %v", pkg, err)
				return
			}

			if len(testFiles) == 0 {
				t.Skipf("No test files found in %s", pkg)
				return
			}

			parallelTestCount := 0
			totalCPUBoundTests := 0

			for _, testFile := range testFiles {
				// Parse the test file
				fset := token.NewFileSet()
				node, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
				if err != nil {
					continue // Skip files that can't be parsed
				}

				// Walk the AST to find test functions
				ast.Inspect(node, func(n ast.Node) bool {
					if fn, ok := n.(*ast.FuncDecl); ok {
						if fn.Name != nil && strings.HasPrefix(fn.Name.Name, "Test") {
							// Check if this looks like a CPU-bound test
							testName := fn.Name.Name
							isCPUBound := false
							for _, pattern := range cpuBoundPatterns {
								if strings.Contains(testName, pattern) {
									isCPUBound = true
									break
								}
							}

							if isCPUBound {
								totalCPUBoundTests++

								// Check if test calls t.Parallel()
								hasParallel := false
								ast.Inspect(fn, func(inner ast.Node) bool {
									if call, ok := inner.(*ast.CallExpr); ok {
										if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
											if ident, ok := sel.X.(*ast.Ident); ok {
												if ident.Name == "t" && sel.Sel.Name == "Parallel" {
													hasParallel = true
													return false
												}
											}
										}
									}
									return true
								})

								if hasParallel {
									parallelTestCount++
								} else {
									t.Errorf("CPU-bound test %s in %s should call t.Parallel()", testName, testFile)
								}
							}
						}
					}
					return true
				})
			}

			// Require that CPU-bound tests use parallel execution
			if totalCPUBoundTests > 0 {
				parallelRatio := float64(parallelTestCount) / float64(totalCPUBoundTests)
				assert.GreaterOrEqual(t, parallelRatio, 0.8,
					"At least 80%% of CPU-bound tests in %s should use t.Parallel() (found %d/%d)",
					pkg, parallelTestCount, totalCPUBoundTests)
			}
		})
	}
}
