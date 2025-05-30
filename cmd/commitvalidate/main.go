// commitvalidate provides a CLI for validating conventional commit messages
// This tool validates commit messages using Go-native validation without external dependencies
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/phrazzld/thinktank/internal/commitvalidate"
)

const (
	exitSuccess = 0
	exitError   = 1
)

type config struct {
	fromSHA   string
	toSHA     string
	commitSHA string
	stdinMode bool
	verbose   bool
	showHelp  bool
}

func main() {
	cfg := parseFlags()

	if cfg.showHelp {
		printUsage()
		os.Exit(exitSuccess)
	}

	validator := commitvalidate.NewRangeValidator()

	// Determine mode of operation
	switch {
	case cfg.stdinMode:
		// Read commit message from stdin (for commit-msg hooks)
		exitCode := validateFromStdin(validator.Validator, cfg.verbose)
		os.Exit(exitCode)

	case cfg.fromSHA != "" && cfg.toSHA != "":
		// Range validation (for pre-push hooks and CI)
		exitCode := validateRange(validator, cfg.fromSHA, cfg.toSHA, cfg.verbose)
		os.Exit(exitCode)

	case cfg.commitSHA != "":
		// Single commit validation
		exitCode := validateCommit(validator, cfg.commitSHA, cfg.verbose)
		os.Exit(exitCode)

	default:
		fmt.Fprintf(os.Stderr, "Error: Must specify either --stdin, --commit <SHA>, or --from <SHA> --to <SHA>\n\n")
		printUsage()
		os.Exit(exitError)
	}
}

func parseFlags() config {
	cfg := config{}

	flag.StringVar(&cfg.fromSHA, "from", "", "Start commit SHA for range validation")
	flag.StringVar(&cfg.toSHA, "to", "", "End commit SHA for range validation")
	flag.StringVar(&cfg.commitSHA, "commit", "", "Single commit SHA to validate")
	flag.BoolVar(&cfg.stdinMode, "stdin", false, "Read commit message from stdin")
	flag.BoolVar(&cfg.verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&cfg.showHelp, "help", false, "Show help")
	flag.BoolVar(&cfg.showHelp, "h", false, "Show help")

	flag.Parse()

	return cfg
}

func printUsage() {
	fmt.Printf(`commitvalidate - Go-native conventional commit validator

USAGE:
    commitvalidate [OPTIONS]

MODES:
    --stdin                     Read commit message from stdin (for commit-msg hooks)
    --commit <SHA>              Validate a single commit
    --from <SHA> --to <SHA>     Validate range of commits (for pre-push hooks and CI)

OPTIONS:
    --verbose                   Show detailed validation information
    --help, -h                  Show this help message

EXAMPLES:
    # Validate commit message from stdin (typical for commit-msg hooks)
    echo "feat: add new feature" | commitvalidate --stdin

    # Validate a single commit
    commitvalidate --commit abc123def

    # Validate range of commits (typical for pre-push hooks)
    commitvalidate --from origin/main --to HEAD

    # Validate with verbose output
    commitvalidate --from abc123 --to def456 --verbose

BASELINE VALIDATION:
    Only commits made after the baseline commit (%s) are validated.
    This preserves git history while ensuring future commits follow conventions.

CONVENTIONAL COMMIT RULES:
    Format: <type>[optional scope]: <description>

    Valid types: feat, fix, docs, style, refactor, perf, test, chore, ci, build, revert
    Scope: optional, lowercase, alphanumeric with hyphens/slashes
    Description: lowercase start, no period at end

    Examples:
        feat: add user authentication
        fix(api): resolve login timeout issue
        feat(auth)!: breaking change to login flow

EXIT CODES:
    0    All commits are valid
    1    One or more commits are invalid or error occurred
`, commitvalidate.BaselineCommit)
}

func validateFromStdin(validator *commitvalidate.Validator, verbose bool) int {
	// Read commit message from stdin
	scanner := bufio.NewScanner(os.Stdin)
	var messageLines []string

	for scanner.Scan() {
		messageLines = append(messageLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
		return exitError
	}

	message := strings.Join(messageLines, "\n")
	if strings.TrimSpace(message) == "" {
		fmt.Fprintf(os.Stderr, "Error: Empty commit message\n")
		return exitError
	}

	// Validate the message
	result := validator.Validate(message)

	if verbose {
		fmt.Printf("Validating commit message with Go-native validator\n")
		fmt.Printf("Baseline commit: %s\n", commitvalidate.BaselineCommit)
		fmt.Printf("Message: %s\n", strings.Split(message, "\n")[0])
	}

	if result.Valid {
		if verbose {
			fmt.Println("✓ Commit message is valid")
		}
		return exitSuccess
	}

	// Print validation errors
	fmt.Fprintf(os.Stderr, "✗ Commit message validation failed:\n")
	for _, err := range result.Errors {
		fmt.Fprintf(os.Stderr, "  - %s\n", err)
	}
	fmt.Fprintf(os.Stderr, "\nCommit message format should be: <type>[optional scope]: <description>\n")
	fmt.Fprintf(os.Stderr, "Valid types: feat, fix, docs, style, refactor, perf, test, chore, ci, build, revert\n")

	return exitError
}

func validateRange(validator *commitvalidate.RangeValidator, fromSHA, toSHA string, verbose bool) int {
	if verbose {
		fmt.Printf("Validating commit range with Go-native validator\n")
		fmt.Printf("Range: %s..%s\n", fromSHA, toSHA)
		fmt.Printf("Baseline commit: %s\n", commitvalidate.BaselineCommit)
		fmt.Printf("Only commits after baseline will be validated\n")
		fmt.Println("--------------------------------------------------------")
	}

	// Validate the range
	result, err := validator.ValidateRange(fromSHA, toSHA)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error validating range: %v\n", err)
		return exitError
	}

	if verbose || result.SkippedCommits > 0 {
		fmt.Printf("Commits checked: %d\n", result.CommitsChecked)
		if result.SkippedCommits > 0 {
			fmt.Printf("Commits skipped (before baseline): %d\n", result.SkippedCommits)
		}
	}

	if result.Valid {
		if verbose {
			fmt.Println("✓ All commits pass conventional commit validation")
		}
		return exitSuccess
	}

	// Print validation errors
	fmt.Printf("✗ Found %d commits with validation errors:\n\n", len(result.Errors))

	for _, commitError := range result.Errors {
		fmt.Printf("Commit %s (%s)\n", commitError.Commit.ShortSHA, commitError.Commit.Date)
		fmt.Printf("Author: %s\n", commitError.Commit.Author)
		fmt.Printf("Message: %s\n", strings.Split(commitError.Commit.Message, "\n")[0])
		fmt.Println("Errors:")
		for _, err := range commitError.Errors {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Println()
	}

	fmt.Println("Fix tips:")
	fmt.Println("  1. Format should be: <type>[optional scope]: <description>")
	fmt.Println("  2. Valid types: feat, fix, docs, style, refactor, test, chore, ci, build, perf, revert")
	fmt.Println("  3. Use 'git commit --amend' to fix the most recent commit")
	fmt.Println("  4. Use 'git rebase -i' to fix older commits")
	fmt.Printf("  5. Only commits after baseline (%s) are validated\n", commitvalidate.BaselineCommit)

	return exitError
}

func validateCommit(validator *commitvalidate.RangeValidator, commitSHA string, verbose bool) int {
	if verbose {
		fmt.Printf("Validating single commit with Go-native validator\n")
		fmt.Printf("Commit: %s\n", commitSHA)
		fmt.Printf("Baseline commit: %s\n", commitvalidate.BaselineCommit)
	}

	// Check if commit is after baseline
	if !validator.IsBaselineAncestor(commitSHA) {
		if verbose {
			fmt.Printf("✓ Commit %s is before baseline - skipping validation\n", commitSHA)
		}
		return exitSuccess
	}

	// Validate the commit
	result, err := validator.ValidateCommit(commitSHA)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error validating commit: %v\n", err)
		return exitError
	}

	if result.Valid {
		if verbose {
			fmt.Printf("✓ Commit %s passes validation\n", commitSHA)
		}
		return exitSuccess
	}

	// Print validation errors
	fmt.Printf("✗ Commit %s validation failed:\n", commitSHA)
	for _, err := range result.Errors {
		fmt.Printf("  - %s\n", err)
	}

	return exitError
}
