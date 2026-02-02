// Package metrics provides observability infrastructure for collecting
// and exporting runtime metrics like durations, counters, and gauges.
package metrics

import "time"

// MetricType represents the type of metric being recorded.
type MetricType string

const (
	// TypeCounter represents a monotonically increasing counter.
	TypeCounter MetricType = "counter"
	// TypeGauge represents a point-in-time value that can go up or down.
	TypeGauge MetricType = "gauge"
	// TypeDuration represents a time duration metric.
	TypeDuration MetricType = "duration"
)

// Metric represents a single metric data point.
type Metric struct {
	Timestamp time.Time         `json:"timestamp"`
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// parseLabels converts variadic key=value string pairs into a map.
// Labels must be provided as alternating key, value pairs.
// If an odd number of strings is provided, the last key gets an empty value.
func parseLabels(labels []string) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	result := make(map[string]string)
	for i := 0; i < len(labels); i += 2 {
		key := labels[i]
		value := ""
		if i+1 < len(labels) {
			value = labels[i+1]
		}
		result[key] = value
	}
	return result
}
