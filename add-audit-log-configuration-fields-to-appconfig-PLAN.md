# Add audit log configuration fields to AppConfig

## Goal
Enhance the AppConfig struct in config.go to include configuration fields for controlling audit logging behavior, specifically enabling/disabling audit logging and specifying the log file path.

## Implementation Approach
I'll update the AppConfig struct in `internal/config/config.go` to add the following new fields:

1. **AuditLogEnabled** - A boolean flag that controls whether structured audit logging is enabled
2. **AuditLogFile** - A string field that specifies the path to the audit log file

These fields will be added to the `AppConfig` struct with appropriate mapstructure and TOML tags consistent with the existing configuration pattern. The fields will need to be included in the `DefaultConfig()` function with sensible defaults.

The implementation will focus on the configuration structure only and won't include integration with the actual audit logging system yet (that's covered by subsequent tasks).

## Key Design Decisions

1. **Field Placement** - I'll place the new audit log fields in the "Logging and display settings" section of the AppConfig struct to maintain logical grouping with other logging-related fields.

2. **Field Naming** - I'll use `AuditLogEnabled` and `AuditLogFile` for field names, with snake_case mapstructure tags (`audit_log_enabled` and `audit_log_file`) to maintain consistency with the existing configuration pattern.

3. **Default Values**:
   - `AuditLogEnabled` will default to `false` to ensure audit logging is opt-in and doesn't impact performance by default
   - `AuditLogFile` will default to an empty string, which will be interpreted by later implementations to use a standard location based on XDG paths

4. **TOML Tags** - Both fields will include TOML tags for serialization/deserialization, making them configurable via the config.toml file.

This approach ensures that the audit logging configuration follows the same patterns as the rest of the application configuration system and lays the groundwork for implementing the actual audit logging behavior in subsequent tasks.