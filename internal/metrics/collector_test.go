package metrics

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDefaultCollector_RecordDuration(t *testing.T) {
	c := NewCollector(nil)
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	c.clock = func() time.Time { return fixedTime }

	c.RecordDuration("api_latency_ms", 150*time.Millisecond, "model", "gpt-4o")

	metrics := c.Metrics()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	m := metrics[0]
	if m.Name != "api_latency_ms" {
		t.Errorf("expected name api_latency_ms, got %s", m.Name)
	}
	if m.Type != TypeDuration {
		t.Errorf("expected type %s, got %s", TypeDuration, m.Type)
	}
	if m.Value != 150 {
		t.Errorf("expected value 150, got %f", m.Value)
	}
	if m.Labels["model"] != "gpt-4o" {
		t.Errorf("expected label model=gpt-4o, got %s", m.Labels["model"])
	}
	if !m.Timestamp.Equal(fixedTime) {
		t.Errorf("expected timestamp %v, got %v", fixedTime, m.Timestamp)
	}
}

func TestDefaultCollector_StartTimer(t *testing.T) {
	c := NewCollector(nil)
	callCount := 0
	times := []time.Time{
		time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),         // StartTimer start
		time.Date(2024, 1, 15, 10, 30, 1, 0, time.UTC),         // stop() calculates duration
		time.Date(2024, 1, 15, 10, 30, 1, 500000000, time.UTC), // RecordDuration timestamp
	}
	c.clock = func() time.Time {
		idx := callCount
		if idx >= len(times) {
			idx = len(times) - 1
		}
		callCount++
		return times[idx]
	}

	stop := c.StartTimer("operation_ms", "op", "test")
	stop()

	metrics := c.Metrics()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	m := metrics[0]
	if m.Name != "operation_ms" {
		t.Errorf("expected name operation_ms, got %s", m.Name)
	}
	if m.Value != 1000 {
		t.Errorf("expected value 1000ms, got %f", m.Value)
	}
	if m.Labels["op"] != "test" {
		t.Errorf("expected label op=test, got %s", m.Labels["op"])
	}
}

func TestDefaultCollector_IncrCounter(t *testing.T) {
	c := NewCollector(nil)

	c.IncrCounter("requests_total", "status", "success")
	c.IncrCounter("requests_total", "status", "success")

	metrics := c.Metrics()
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	for _, m := range metrics {
		if m.Type != TypeCounter {
			t.Errorf("expected type %s, got %s", TypeCounter, m.Type)
		}
		if m.Value != 1 {
			t.Errorf("expected value 1, got %f", m.Value)
		}
	}
}

func TestDefaultCollector_AddCounter(t *testing.T) {
	c := NewCollector(nil)

	c.AddCounter("bytes_processed", 1024, "source", "file")

	metrics := c.Metrics()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	m := metrics[0]
	if m.Value != 1024 {
		t.Errorf("expected value 1024, got %f", m.Value)
	}
}

func TestDefaultCollector_SetGauge(t *testing.T) {
	c := NewCollector(nil)

	c.SetGauge("token_utilization_pct", 75.5, "model", "gpt-4o")

	metrics := c.Metrics()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	m := metrics[0]
	if m.Type != TypeGauge {
		t.Errorf("expected type %s, got %s", TypeGauge, m.Type)
	}
	if m.Value != 75.5 {
		t.Errorf("expected value 75.5, got %f", m.Value)
	}
}

func TestDefaultCollector_Flush(t *testing.T) {
	var buf bytes.Buffer
	exporter := NewJSONLinesExporter(&buf)
	c := NewCollector(exporter)
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	c.clock = func() time.Time { return fixedTime }

	c.RecordDuration("api_latency_ms", 100*time.Millisecond, "model", "gpt-4o")
	c.IncrCounter("requests_total")

	if err := c.Flush(); err != nil {
		t.Fatalf("flush failed: %v", err)
	}

	// Verify buffer was cleared
	if len(c.Metrics()) != 0 {
		t.Errorf("expected metrics to be cleared after flush, got %d", len(c.Metrics()))
	}

	// Verify output
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %s", len(lines), buf.String())
	}

	var m Metric
	if err := json.Unmarshal([]byte(lines[0]), &m); err != nil {
		t.Fatalf("failed to unmarshal first line: %v", err)
	}
	if m.Name != "api_latency_ms" {
		t.Errorf("expected first metric name api_latency_ms, got %s", m.Name)
	}
}

func TestDefaultCollector_FlushWithNilExporter(t *testing.T) {
	c := NewCollector(nil)
	c.IncrCounter("test")

	if err := c.Flush(); err != nil {
		t.Errorf("flush with nil exporter should succeed, got: %v", err)
	}
}

func TestDefaultCollector_ThreadSafety(t *testing.T) {
	c := NewCollector(nil)
	done := make(chan struct{})

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				c.IncrCounter("concurrent_counter", "goroutine", string(rune('0'+id)))
			}
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	metrics := c.Metrics()
	if len(metrics) != 1000 {
		t.Errorf("expected 1000 metrics, got %d", len(metrics))
	}
}

func TestNoopCollector(t *testing.T) {
	c := NewNoopCollector()

	// None of these should panic
	c.RecordDuration("test", time.Second)
	stop := c.StartTimer("test")
	stop()
	c.IncrCounter("test")
	c.AddCounter("test", 10)
	c.SetGauge("test", 1.0)

	if err := c.Flush(); err != nil {
		t.Errorf("noop flush should succeed, got: %v", err)
	}

	if c.Metrics() != nil {
		t.Error("noop metrics should return nil")
	}
}

func TestJSONLinesExporter(t *testing.T) {
	var buf bytes.Buffer
	exporter := NewJSONLinesExporter(&buf)
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	metrics := []Metric{
		{
			Timestamp: fixedTime,
			Name:      "api_latency_ms",
			Type:      TypeDuration,
			Value:     150,
			Labels:    map[string]string{"model": "gpt-4o"},
		},
	}

	if err := exporter.Export(metrics); err != nil {
		t.Fatalf("export failed: %v", err)
	}

	expected := `{"timestamp":"2024-01-15T10:30:00Z","name":"api_latency_ms","type":"duration","value":150,"labels":{"model":"gpt-4o"}}`
	got := strings.TrimSpace(buf.String())
	if got != expected {
		t.Errorf("unexpected output:\ngot:  %s\nwant: %s", got, expected)
	}
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]string
	}{
		{
			name:     "empty",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "single pair",
			input:    []string{"key", "value"},
			expected: map[string]string{"key": "value"},
		},
		{
			name:     "multiple pairs",
			input:    []string{"a", "1", "b", "2"},
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			name:     "odd number",
			input:    []string{"key"},
			expected: map[string]string{"key": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLabels(tt.input)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			for k, v := range tt.expected {
				if got[k] != v {
					t.Errorf("expected %s=%s, got %s=%s", k, v, k, got[k])
				}
			}
		})
	}
}
