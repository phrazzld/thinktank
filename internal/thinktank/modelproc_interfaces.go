package thinktank

import (
	"context"
)

// ModelProcessor defines the interface for interacting with AI models
type ModelProcessor interface {
	// Process handles the entire model processing workflow for a single model
	// Returns the generated content as a string and any error encountered
	Process(ctx context.Context, modelName string, stitchedPrompt string) (string, error)
}
