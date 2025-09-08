package types

import (
	"time"
)

// MonitoringFeature represents a monitoring feature
type MonitoringFeature string

const (
	// Basic monitoring features
	FeatureMetrics     MonitoringFeature = "metrics"
	FeatureLogging     MonitoringFeature = "logging"
	FeatureTracing     MonitoringFeature = "tracing"
	FeatureHealthCheck MonitoringFeature = "health_check"
	FeatureAlerting    MonitoringFeature = "alerting"

	// Advanced features
	FeatureCustomMetrics         MonitoringFeature = "custom_metrics"
	FeatureDistributedTracing    MonitoringFeature = "distributed_tracing"
	FeaturePerformanceMonitoring MonitoringFeature = "performance_monitoring"
	FeatureErrorTracking         MonitoringFeature = "error_tracking"
	FeatureUptimeMonitoring      MonitoringFeature = "uptime_monitoring"
	FeatureResourceMonitoring    MonitoringFeature = "resource_monitoring"
	FeatureAPMMonitoring         MonitoringFeature = "apm_monitoring"
	FeatureRealTimeMonitoring    MonitoringFeature = "real_time_monitoring"
	FeatureHistoricalData        MonitoringFeature = "historical_data"
	FeatureDashboard             MonitoringFeature = "dashboard"
	FeatureNotification          MonitoringFeature = "notification"
	FeatureSLA                   MonitoringFeature = "sla"
	FeatureCapacityPlanning      MonitoringFeature = "capacity_planning"
)

// ConnectionInfo represents monitoring provider connection information
type ConnectionInfo struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Protocol string            `json:"protocol"`
	Version  string            `json:"version"`
	Secure   bool              `json:"secure"`
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

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
	MetricTypeTimer     MetricType = "timer"
)

// LogLevel represents the log level
type LogLevel string

const (
	LogLevelTrace LogLevel = "trace"
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

// HealthStatus represents the health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// Metric represents a monitoring metric
type Metric struct {
	Name        string                 `json:"name"`
	Type        MetricType             `json:"type"`
	Value       float64                `json:"value"`
	Labels      map[string]string      `json:"labels"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description,omitempty"`
	Unit        string                 `json:"unit,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Service   string                 `json:"service"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Trace represents a distributed trace
type Trace struct {
	TraceID     string                 `json:"trace_id"`
	ServiceName string                 `json:"service_name"`
	Operation   string                 `json:"operation"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	Status      string                 `json:"status"`
	Tags        map[string]string      `json:"tags,omitempty"`
	Spans       []Span                 `json:"spans,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Span represents a span in a trace
type Span struct {
	SpanID    string                 `json:"span_id"`
	ParentID  string                 `json:"parent_id,omitempty"`
	Operation string                 `json:"operation"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  time.Duration          `json:"duration"`
	Status    string                 `json:"status"`
	Tags      map[string]string      `json:"tags,omitempty"`
	Logs      []LogEntry             `json:"logs,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// HealthCheck represents a health check
type HealthCheck struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Alert represents an alert
type Alert struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Severity    AlertSeverity          `json:"severity"`
	Status      string                 `json:"status"` // active, resolved, acknowledged
	Source      string                 `json:"source"`
	Service     string                 `json:"service"`
	Timestamp   time.Time              `json:"timestamp"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Annotations map[string]string      `json:"annotations,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MonitoringStats represents monitoring statistics
type MonitoringStats struct {
	MetricsCount int64                  `json:"metrics_count"`
	LogsCount    int64                  `json:"logs_count"`
	TracesCount  int64                  `json:"traces_count"`
	AlertsCount  int64                  `json:"alerts_count"`
	ActiveAlerts int64                  `json:"active_alerts"`
	Uptime       time.Duration          `json:"uptime"`
	LastUpdate   time.Time              `json:"last_update"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// ProviderInfo holds information about a monitoring provider
type ProviderInfo struct {
	Name              string              `json:"name"`
	SupportedFeatures []MonitoringFeature `json:"supported_features"`
	ConnectionInfo    *ConnectionInfo     `json:"connection_info"`
	IsConnected       bool                `json:"is_connected"`
}

// MonitoringConfig holds general monitoring configuration
type MonitoringConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	BatchSize       int               `json:"batch_size"`
	FlushInterval   time.Duration     `json:"flush_interval"`
	BufferSize      int               `json:"buffer_size"`
	SamplingRate    float64           `json:"sampling_rate"`
	EnableTracing   bool              `json:"enable_tracing"`
	EnableMetrics   bool              `json:"enable_metrics"`
	EnableLogging   bool              `json:"enable_logging"`
	EnableAlerting  bool              `json:"enable_alerting"`
	Metadata        map[string]string `json:"metadata"`
}

// MonitoringError represents a monitoring-specific error
type MonitoringError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Source  string `json:"source,omitempty"`
}

func (e *MonitoringError) Error() string {
	return e.Message
}

// Common monitoring error codes
const (
	ErrCodeConnection      = "CONNECTION_ERROR"
	ErrCodeTimeout         = "TIMEOUT"
	ErrCodeInvalidMetric   = "INVALID_METRIC"
	ErrCodeInvalidLog      = "INVALID_LOG"
	ErrCodeInvalidTrace    = "INVALID_TRACE"
	ErrCodeInvalidAlert    = "INVALID_ALERT"
	ErrCodeSerialization   = "SERIALIZATION_ERROR"
	ErrCodeDeserialization = "DESERIALIZATION_ERROR"
	ErrCodeQuota           = "QUOTA_EXCEEDED"
	ErrCodeRateLimit       = "RATE_LIMIT_EXCEEDED"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
)

// MetricRequest represents a metric submission request
type MetricRequest struct {
	Metrics []Metric               `json:"metrics"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// LogRequest represents a log submission request
type LogRequest struct {
	Logs    []LogEntry             `json:"logs"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// TraceRequest represents a trace submission request
type TraceRequest struct {
	Traces  []Trace                `json:"traces"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// AlertRequest represents an alert submission request
type AlertRequest struct {
	Alerts  []Alert                `json:"alerts"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// HealthCheckRequest represents a health check request
type HealthCheckRequest struct {
	Service string        `json:"service"`
	Timeout time.Duration `json:"timeout,omitempty"`
}

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Service   string                 `json:"service"`
	Status    HealthStatus           `json:"status"`
	Checks    []HealthCheck          `json:"checks"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// QueryRequest represents a query request
type QueryRequest struct {
	Query   string                 `json:"query"`
	Start   time.Time              `json:"start"`
	End     time.Time              `json:"end"`
	Step    time.Duration          `json:"step,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// QueryResponse represents a query response
type QueryResponse struct {
	Data      interface{}            `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AddField adds a field to a log entry
func (l *LogEntry) AddField(key string, value interface{}) {
	if l.Fields == nil {
		l.Fields = make(map[string]interface{})
	}
	l.Fields[key] = value
}

// AddMetadata adds metadata to a log entry
func (l *LogEntry) AddMetadata(key string, value interface{}) {
	if l.Metadata == nil {
		l.Metadata = make(map[string]interface{})
	}
	l.Metadata[key] = value
}

// SetTraceInfo sets trace information for a log entry
func (l *LogEntry) SetTraceInfo(traceID, spanID string) {
	l.TraceID = traceID
	l.SpanID = spanID
}

// FinishTrace finishes a trace
func (t *Trace) FinishTrace(status string) {
	t.EndTime = time.Now()
	t.Duration = t.EndTime.Sub(t.StartTime)
	t.Status = status
}

// AddSpan adds a span to a trace
func (t *Trace) AddSpan(span *Span) {
	t.Spans = append(t.Spans, *span)
}

// AddTag adds a tag to a trace
func (t *Trace) AddTag(key, value string) {
	if t.Tags == nil {
		t.Tags = make(map[string]string)
	}
	t.Tags[key] = value
}

// FinishSpan finishes a span
func (s *Span) FinishSpan(status string) {
	s.EndTime = time.Now()
	s.Duration = s.EndTime.Sub(s.StartTime)
	s.Status = status
}

// AddTag adds a tag to a span
func (s *Span) AddTag(key, value string) {
	if s.Tags == nil {
		s.Tags = make(map[string]string)
	}
	s.Tags[key] = value
}

// AddLog adds a log entry to a span
func (s *Span) AddLog(level LogLevel, message string, fields map[string]interface{}) {
	logEntry := LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Source:    s.Operation,
		Fields:    fields,
		Metadata:  make(map[string]interface{}),
	}

	if s.Logs == nil {
		s.Logs = make([]LogEntry, 0)
	}
	s.Logs = append(s.Logs, logEntry)
}
