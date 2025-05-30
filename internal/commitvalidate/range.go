package commitvalidate

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	// BaselineCommit is the commit SHA where conventional commits were established
	// Only commits after this baseline will be validated (forward-only validation)
	BaselineCommit = "1300e4d675ac087783199f1e608409e6853e589f"
)

// RangeValidator handles Git range-based commit validation
type RangeValidator struct {
	Validator *Validator
}

// NewRangeValidator creates a new range validator
func NewRangeValidator() *RangeValidator {
	return &RangeValidator{
		Validator: NewValidator(),
	}
}

// CommitInfo contains information about a Git commit
type CommitInfo struct {
	SHA      string
	Message  string
	Author   string
	Date     string
	ShortSHA string
}

// RangeValidationResult contains results from validating a range of commits
type RangeValidationResult struct {
	Valid          bool
	CommitsChecked int
	Errors         []CommitValidationError
	SkippedCommits int // Commits before baseline
}

// CommitValidationError represents a validation error for a specific commit
type CommitValidationError struct {
	Commit CommitInfo
	Errors []string
}

// ValidateRange validates all commits in the given range
// Only commits after the baseline will be validated
func (rv *RangeValidator) ValidateRange(fromSHA, toSHA string) (*RangeValidationResult, error) {
	result := &RangeValidationResult{
		Valid:          true,
		CommitsChecked: 0,
		Errors:         []CommitValidationError{},
		SkippedCommits: 0,
	}

	// Get commits in range
	commits, err := rv.getCommitsInRange(fromSHA, toSHA)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits in range %s..%s: %w", fromSHA, toSHA, err)
	}

	if len(commits) == 0 {
		return result, nil
	}

	// Validate each commit
	for _, commit := range commits {
		// Check if commit is after baseline
		if !rv.isCommitAfterBaseline(commit.SHA) {
			result.SkippedCommits++
			continue
		}

		// Validate the commit message
		validationResult := rv.Validator.Validate(commit.Message)
		result.CommitsChecked++

		if !validationResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, CommitValidationError{
				Commit: commit,
				Errors: validationResult.Errors,
			})
		}
	}

	return result, nil
}

// ValidateCommit validates a single commit by SHA
func (rv *RangeValidator) ValidateCommit(commitSHA string) (*ValidationResult, error) {
	// Check if commit is after baseline
	if !rv.isCommitAfterBaseline(commitSHA) {
		return &ValidationResult{
			Valid:  true,
			Errors: []string{},
		}, nil
	}

	// Get commit message
	message, err := rv.getCommitMessage(commitSHA)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit message for %s: %w", commitSHA, err)
	}

	return rv.Validator.Validate(message), nil
}

// getCommitsInRange gets all commits in the specified range
func (rv *RangeValidator) getCommitsInRange(fromSHA, toSHA string) ([]CommitInfo, error) {
	// Use git rev-list to get commit SHAs in range
	// #nosec G204 -- Git SHAs come from git itself, not user input
	cmd := exec.Command("git", "rev-list", "--reverse", fromSHA+".."+toSHA)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git rev-list failed: %w", err)
	}

	shas := strings.Fields(strings.TrimSpace(string(output)))
	if len(shas) == 0 {
		return []CommitInfo{}, nil
	}

	commits := make([]CommitInfo, 0, len(shas))
	for _, sha := range shas {
		commit, err := rv.getCommitInfo(sha)
		if err != nil {
			return nil, fmt.Errorf("failed to get commit info for %s: %w", sha, err)
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

// getCommitInfo gets detailed information about a commit
func (rv *RangeValidator) getCommitInfo(sha string) (CommitInfo, error) {
	// Get commit message
	message, err := rv.getCommitMessage(sha)
	if err != nil {
		return CommitInfo{}, err
	}

	// Get commit metadata
	cmd := exec.Command("git", "log", "-1", "--format=%an <%ae>%n%ci%n%h", sha)
	output, err := cmd.Output()
	if err != nil {
		return CommitInfo{}, fmt.Errorf("failed to get commit metadata: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 3 {
		return CommitInfo{}, fmt.Errorf("unexpected git log output format")
	}

	return CommitInfo{
		SHA:      sha,
		Message:  message,
		Author:   lines[0],
		Date:     lines[1],
		ShortSHA: lines[2],
	}, nil
}

// getCommitMessage gets the full commit message for a SHA
func (rv *RangeValidator) getCommitMessage(sha string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%B", sha)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit message: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// isCommitAfterBaseline checks if a commit is after the baseline commit
func (rv *RangeValidator) isCommitAfterBaseline(commitSHA string) bool {
	// Use git merge-base to check if baseline is an ancestor of the commit
	cmd := exec.Command("git", "merge-base", "--is-ancestor", BaselineCommit, commitSHA)
	err := cmd.Run()
	// If no error, baseline is an ancestor (commit is after baseline)
	// If error, either baseline is not an ancestor, or there's another issue
	return err == nil
}

// IsBaselineAncestor checks if the baseline commit is an ancestor of the given commit
func (rv *RangeValidator) IsBaselineAncestor(commitSHA string) bool {
	return rv.isCommitAfterBaseline(commitSHA)
}

// GetBaselineCommit returns the baseline commit SHA
func (rv *RangeValidator) GetBaselineCommit() string {
	return BaselineCommit
}
