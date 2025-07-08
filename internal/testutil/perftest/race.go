//go:build race

package perftest

// raceDetectionEnabledByBuildTag returns true when the race build tag is set
func raceDetectionEnabledByBuildTag() bool {
	return true
}
