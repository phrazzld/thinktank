package metrics

import (
	"encoding/json"
	"fmt"
)

// Export writes metrics as JSON Lines to the underlying writer.
func (e *JSONLinesExporter) Export(metrics []Metric) error {
	for _, m := range metrics {
		data, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("failed to marshal metric %s: %w", m.Name, err)
		}
		if _, err := e.writer.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("failed to write metric %s: %w", m.Name, err)
		}
	}
	return nil
}
