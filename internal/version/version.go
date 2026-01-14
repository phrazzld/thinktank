// Package version provides build-time version information for thinktank.
// Values are injected via ldflags during build.
package version

import "fmt"

// Build-time variables injected via ldflags:
//
//	-X github.com/phrazzld/thinktank/internal/version.Version=v1.2.3
//	-X github.com/phrazzld/thinktank/internal/version.Commit=abc1234
//	-X github.com/phrazzld/thinktank/internal/version.BuildDate=2025-01-13T...
var (
	// Version is the semantic version (e.g., "v1.2.3" or "dev" for development builds)
	Version = "dev"

	// Commit is the git commit hash (short form)
	Commit = "unknown"

	// BuildDate is the ISO 8601 build timestamp
	BuildDate = "unknown"
)

// String returns a formatted version string for display.
func String() string {
	return fmt.Sprintf("thinktank %s (%s, %s)", Version, Commit, BuildDate)
}

// Short returns just the version number.
func Short() string {
	return Version
}
