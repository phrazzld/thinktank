// Package auditlog provides structured logging capabilities for the architect tool.
//
// It implements a JSON lines-based logging system that writes structured log entries
// to a configurable file path. This allows for detailed operational logging including
// inputs, outputs, token counts, errors, and other metadata in a format suitable for
// programmatic analysis and auditability.
//
// The package is designed to operate independently from the existing console logging
// system, focusing on machine-readable, persistent structured logs rather than human
// readable output.
package auditlog
