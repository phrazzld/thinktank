//go:build !race

package perftest

// raceDetectionEnabledByBuildTag returns false when the race build tag is not set
func raceDetectionEnabledByBuildTag() bool {
	return false
}
