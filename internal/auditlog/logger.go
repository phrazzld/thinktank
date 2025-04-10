// Package auditlog provides structured logging capabilities for the architect tool.
package auditlog

// StructuredLogger defines the interface for structured audit logging.
// It provides methods for logging structured events and cleaning up resources.
type StructuredLogger interface {
	// Log records a structured audit event.
	// Implementations should ensure this method is safe for concurrent use
	// and should handle any errors internally to prevent disruption to the
	// application flow (e.g., by logging errors to the standard logger).
	Log(event AuditEvent)

	// Close releases any resources held by the logger.
	// This should be called when the logger is no longer needed,
	// typically using the defer pattern after logger creation.
	// Implementations should ensure this method is idempotent and
	// safe to call multiple times.
	// 
	// Returns an error if cleanup fails, which the caller may choose
	// to log but typically should not cause the application to fail.
	Close() error
}
