package types

import (
	"context"
	"time"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelTrace LogLevel = "trace"
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
	LevelPanic LogLevel = "panic"
)

// LogFormat represents the log output format
type LogFormat string

const (
	FormatJSON LogFormat = "json"
	FormatText LogFormat = "text"
	FormatXML  LogFormat = "xml"
)

// LogOutput represents the log output destination
type LogOutput string

const (
	OutputStdout LogOutput = "stdout"
	OutputStderr LogOutput = "stderr"
	OutputFile   LogOutput = "file"
	OutputSyslog LogOutput = "syslog"
)

// LoggingProvider represents a logging provider interface
type LoggingProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []LoggingFeature
	GetConnectionInfo() *ConnectionInfo

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Basic logging operations
	Log(ctx context.Context, level LogLevel, message string, fields map[string]interface{}) error
	LogWithContext(ctx context.Context, level LogLevel, message string, fields map[string]interface{}) error

	// Structured logging
	Info(ctx context.Context, message string, fields ...map[string]interface{}) error
	Debug(ctx context.Context, message string, fields ...map[string]interface{}) error
	Warn(ctx context.Context, message string, fields ...map[string]interface{}) error
	Error(ctx context.Context, message string, fields ...map[string]interface{}) error
	Fatal(ctx context.Context, message string, fields ...map[string]interface{}) error
	Panic(ctx context.Context, message string, fields ...map[string]interface{}) error

	// Formatted logging
	Infof(ctx context.Context, format string, args ...interface{}) error
	Debugf(ctx context.Context, format string, args ...interface{}) error
	Warnf(ctx context.Context, format string, args ...interface{}) error
	Errorf(ctx context.Context, format string, args ...interface{}) error
	Fatalf(ctx context.Context, format string, args ...interface{}) error
	Panicf(ctx context.Context, format string, args ...interface{}) error

	// Advanced operations
	WithFields(ctx context.Context, fields map[string]interface{}) LoggingProvider
	WithField(ctx context.Context, key string, value interface{}) LoggingProvider
	WithError(ctx context.Context, err error) LoggingProvider
	WithContext(ctx context.Context) LoggingProvider

	// Batch operations
	LogBatch(ctx context.Context, entries []LogEntry) error

	// Query and search
	Search(ctx context.Context, query LogQuery) ([]LogEntry, error)
	GetStats(ctx context.Context) (*LoggingStats, error)
}

// LoggingFeature represents a logging feature
type LoggingFeature string

const (
	// Basic features
	FeatureBasicLogging   LoggingFeature = "basic_logging"
	FeatureStructuredLog  LoggingFeature = "structured_log"
	FeatureFormattedLog   LoggingFeature = "formatted_log"
	FeatureContextLogging LoggingFeature = "context_logging"
	FeatureBatchLogging   LoggingFeature = "batch_logging"

	// Advanced features
	FeatureSearch      LoggingFeature = "search"
	FeatureFiltering   LoggingFeature = "filtering"
	FeatureAggregation LoggingFeature = "aggregation"
	FeatureRetention   LoggingFeature = "retention"
	FeatureCompression LoggingFeature = "compression"
	FeatureEncryption  LoggingFeature = "encryption"
	FeatureReplication LoggingFeature = "replication"
	FeatureClustering  LoggingFeature = "clustering"
	FeatureRealTime    LoggingFeature = "real_time"
	FeatureAlerting    LoggingFeature = "alerting"
	FeatureDashboards  LoggingFeature = "dashboards"
	FeatureMetrics     LoggingFeature = "metrics"
	FeatureTracing     LoggingFeature = "tracing"
)

// ConnectionInfo holds connection information for a logging provider
type ConnectionInfo struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Database string            `json:"database"`
	Username string            `json:"username"`
	Status   ConnectionStatus  `json:"status"`
	Metadata map[string]string `json:"metadata"`
}

// ConnectionStatus represents the connection status
type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusError        ConnectionStatus = "error"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      LogLevel               `json:"level"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service"`
	Version    string                 `json:"version"`
	TraceID    string                 `json:"trace_id,omitempty"`
	SpanID     string                 `json:"span_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	TenantID   string                 `json:"tenant_id,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   int64                  `json:"duration,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Provider   string                 `json:"provider"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level       LogLevel               `json:"level"`
	Format      LogFormat              `json:"format"`
	Output      LogOutput              `json:"output"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	Index       string                 `json:"index"`
	Retention   time.Duration          `json:"retention"`
	Compression bool                   `json:"compression"`
	Encryption  bool                   `json:"encryption"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// LoggingStats represents logging statistics
type LoggingStats struct {
	TotalLogs     int64              `json:"total_logs"`
	LogsByLevel   map[LogLevel]int64 `json:"logs_by_level"`
	LogsByService map[string]int64   `json:"logs_by_service"`
	StorageSize   int64              `json:"storage_size"`
	Uptime        time.Duration      `json:"uptime"`
	LastUpdate    time.Time          `json:"last_update"`
	Provider      string             `json:"provider"`
}

// ProviderInfo holds information about a logging provider
type ProviderInfo struct {
	Name              string           `json:"name"`
	SupportedFeatures []LoggingFeature `json:"supported_features"`
	ConnectionInfo    *ConnectionInfo  `json:"connection_info"`
	IsConnected       bool             `json:"is_connected"`
}

// LogQuery represents a log search query
type LogQuery struct {
	Levels    []LogLevel             `json:"levels,omitempty"`
	Services  []string               `json:"services,omitempty"`
	StartTime *time.Time             `json:"start_time,omitempty"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Limit     int                    `json:"limit,omitempty"`
	Offset    int                    `json:"offset,omitempty"`
	SortBy    string                 `json:"sort_by,omitempty"`
	SortOrder string                 `json:"sort_order,omitempty"`
}

// LoggingError represents a logging-specific error
type LoggingError struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Level   LogLevel `json:"level,omitempty"`
}

func (e *LoggingError) Error() string {
	return e.Message
}

// Common logging error codes
const (
	ErrCodeConnection      = "CONNECTION_ERROR"
	ErrCodeTimeout         = "TIMEOUT"
	ErrCodeInvalidLevel    = "INVALID_LEVEL"
	ErrCodeInvalidFormat   = "INVALID_FORMAT"
	ErrCodeSerialization   = "SERIALIZATION_ERROR"
	ErrCodeDeserialization = "DESERIALIZATION_ERROR"
	ErrCodeQuota           = "QUOTA_EXCEEDED"
	ErrCodeRetention       = "RETENTION_ERROR"
	ErrCodeSearch          = "SEARCH_ERROR"
	ErrCodeBatch           = "BATCH_ERROR"
)

// LogEvent represents a logging event
type LogEvent struct {
	Type      LogEventType `json:"type"`
	Entry     *LogEntry    `json:"entry,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Provider  string       `json:"provider"`
}

// LogEventType represents the type of logging event
type LogEventType string

const (
	EventLog     LogEventType = "log"
	EventError   LogEventType = "error"
	EventWarning LogEventType = "warning"
	EventAlert   LogEventType = "alert"
	EventStats   LogEventType = "stats"
)

// BatchLogEntry represents a batch log entry
type BatchLogEntry struct {
	Level   LogLevel               `json:"level"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// LoggingOptions represents logging options
type LoggingOptions struct {
	Provider     string            `json:"provider"`
	ConfigPath   string            `json:"config_path"`
	Environment  string            `json:"environment"`
	WatchChanges bool              `json:"watch_changes"`
	Metadata     map[string]string `json:"metadata"`
}

// IsValidLevel checks if the log level is valid
func (l LogLevel) IsValid() bool {
	switch l {
	case LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal, LevelPanic:
		return true
	default:
		return false
	}
}

// IsValidFormat checks if the log format is valid
func (f LogFormat) IsValid() bool {
	switch f {
	case FormatJSON, FormatText, FormatXML:
		return true
	default:
		return false
	}
}

// IsValidOutput checks if the log output is valid
func (o LogOutput) IsValid() bool {
	switch o {
	case OutputStdout, OutputStderr, OutputFile, OutputSyslog:
		return true
	default:
		return false
	}
}

// GetDefaultConfig returns default logging configuration
func GetDefaultConfig() *LoggingConfig {
	return &LoggingConfig{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      OutputStdout,
		Service:     "unknown",
		Version:     "1.0.0",
		Index:       "logs",
		Retention:   30 * 24 * time.Hour, // 30 days
		Compression: false,
		Encryption:  false,
		Metadata:    make(map[string]interface{}),
	}
}
