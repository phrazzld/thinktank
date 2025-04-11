// Package auditlog provides structured logging for audit purposes
package auditlog

// AuditLogger defines the interface for logging audit events.
// Implementations of this interface will handle persisting audit
// log entries in various formats (e.g., JSON Lines file, no-op).
type AuditLogger interface {
	// Log records a single audit entry.
	// The entry contains information about operations, status, and relevant metadata.
	// Returns an error if the logging operation fails.
	Log(entry AuditEntry) error

	// Close releases any resources used by the logger (e.g., open file handles).
	// Should be called when the logger is no longer needed.
	// Returns an error if the closing operation fails.
	Close() error
}
