// Package orchestrator is responsible for coordinating the core application workflow.
package orchestrator

// OutputInfo contains information about generated output files
type OutputInfo struct {
	// Path to the synthesis output file (if synthesis was used)
	SynthesisFilePath string

	// Paths to individual model output files (if saving individual outputs)
	IndividualFilePaths map[string]string
}

// NewOutputInfo creates a new OutputInfo instance
func NewOutputInfo() *OutputInfo {
	return &OutputInfo{
		IndividualFilePaths: make(map[string]string),
	}
}
