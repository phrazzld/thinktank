package metrics

import (
	"io"
	"sync"
	"time"
)

// Collector records metrics during execution.
type Collector interface {
	// RecordDuration records a duration metric with optional labels (key, value pairs).
	RecordDuration(name string, duration time.Duration, labels ...string)

	// StartTimer returns a stop function that records duration when called.
	StartTimer(name string, labels ...string) func()

	// IncrCounter increments a counter by 1.
	IncrCounter(name string, labels ...string)

	// AddCounter adds delta to a counter.
	AddCounter(name string, delta int64, labels ...string)

	// SetGauge sets a gauge value.
	SetGauge(name string, value float64, labels ...string)

	// Flush exports collected metrics and clears the buffer.
	Flush() error

	// Metrics returns all collected metrics.
	Metrics() []Metric
}

// DefaultCollector is a thread-safe metrics collector.
type DefaultCollector struct {
	mu       sync.Mutex
	metrics  []Metric
	exporter Exporter
	clock    func() time.Time // immutable after construction
}

// CollectorOption configures a DefaultCollector.
type CollectorOption func(*DefaultCollector)

// WithClock sets a custom clock function (for testing).
func WithClock(clock func() time.Time) CollectorOption {
	return func(c *DefaultCollector) {
		c.clock = clock
	}
}

// NewCollector creates a new DefaultCollector with the given exporter.
func NewCollector(exporter Exporter, opts ...CollectorOption) *DefaultCollector {
	c := &DefaultCollector{
		metrics:  make([]Metric, 0),
		exporter: exporter,
		clock:    time.Now,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// RecordDuration records a duration metric in milliseconds.
func (c *DefaultCollector) RecordDuration(name string, duration time.Duration, labels ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics = append(c.metrics, Metric{
		Timestamp: c.clock(),
		Name:      name,
		Type:      TypeDuration,
		Value:     float64(duration.Milliseconds()),
		Labels:    parseLabels(labels),
	})
}

// StartTimer returns a stop function that records duration when called.
func (c *DefaultCollector) StartTimer(name string, labels ...string) func() {
	start := c.clock()
	return func() {
		c.RecordDuration(name, c.clock().Sub(start), labels...)
	}
}

// IncrCounter increments a counter by 1.
func (c *DefaultCollector) IncrCounter(name string, labels ...string) {
	c.AddCounter(name, 1, labels...)
}

// AddCounter adds delta to a counter.
func (c *DefaultCollector) AddCounter(name string, delta int64, labels ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics = append(c.metrics, Metric{
		Timestamp: c.clock(),
		Name:      name,
		Type:      TypeCounter,
		Value:     float64(delta),
		Labels:    parseLabels(labels),
	})
}

// SetGauge sets a gauge value.
func (c *DefaultCollector) SetGauge(name string, value float64, labels ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics = append(c.metrics, Metric{
		Timestamp: c.clock(),
		Name:      name,
		Type:      TypeGauge,
		Value:     value,
		Labels:    parseLabels(labels),
	})
}

// Flush exports collected metrics and clears the buffer.
func (c *DefaultCollector) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.exporter == nil {
		return nil
	}

	if err := c.exporter.Export(c.metrics); err != nil {
		return err
	}
	c.metrics = c.metrics[:0]
	return nil
}

// Metrics returns a copy of all collected metrics.
func (c *DefaultCollector) Metrics() []Metric {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]Metric, len(c.metrics))
	copy(result, c.metrics)
	return result
}

// NoopCollector is a no-op implementation for disabled metrics.
type NoopCollector struct{}

// NewNoopCollector creates a new NoopCollector.
func NewNoopCollector() *NoopCollector {
	return &NoopCollector{}
}

func (n *NoopCollector) RecordDuration(string, time.Duration, ...string) {}
func (n *NoopCollector) StartTimer(string, ...string) func()             { return func() {} }
func (n *NoopCollector) IncrCounter(string, ...string)                   {}
func (n *NoopCollector) AddCounter(string, int64, ...string)             {}
func (n *NoopCollector) SetGauge(string, float64, ...string)             {}
func (n *NoopCollector) Flush() error                                    { return nil }
func (n *NoopCollector) Metrics() []Metric                               { return nil }

// Exporter defines the interface for exporting metrics.
type Exporter interface {
	Export(metrics []Metric) error
}

// JSONLinesExporter writes metrics as JSON Lines to an io.Writer.
type JSONLinesExporter struct {
	writer io.Writer
}

// NewJSONLinesExporter creates a new JSONLinesExporter.
func NewJSONLinesExporter(w io.Writer) *JSONLinesExporter {
	return &JSONLinesExporter{writer: w}
}
