package interfaces

import (
	"context"
	"time"
)

// MetricsCollector interface defines metrics collection operations
type MetricsCollector interface {
	// Counter metrics
	IncrementCounter(name string, labels map[string]string, value float64)
	AddCounter(name string, labels map[string]string, value float64)

	// Gauge metrics
	SetGauge(name string, labels map[string]string, value float64)
	AddGauge(name string, labels map[string]string, value float64)

	// Histogram metrics
	ObserveHistogram(name string, labels map[string]string, value float64)

	// Summary metrics
	ObserveSummary(name string, labels map[string]string, value float64)

	// Custom metrics
	RecordCustomMetric(name string, metricType MetricType, labels map[string]string, value float64)

	// Business metrics
	RecordBusinessOperation(operation, resource, status string, duration time.Duration)
	RecordBusinessEvent(event string, labels map[string]string)

	// Health metrics
	RecordHealthCheck(service string, status HealthStatus, duration time.Duration)
	RecordDependencyHealth(dependency string, status HealthStatus, duration time.Duration)

	// Start/Stop
	Start() error
	Stop() error
}

// Tracer interface defines distributed tracing operations
type Tracer interface {
	// Span operations
	StartSpan(ctx context.Context, operationName string) (context.Context, Span)
	StartSpanWithOptions(ctx context.Context, operationName string, opts SpanOptions) (context.Context, Span)
	StartSpanFromContext(ctx context.Context, operationName string) (context.Context, Span)

	// Context operations
	Extract(ctx context.Context, carrier interface{}) (context.Context, error)
	Inject(ctx context.Context, carrier interface{}) error

	// Global tracer operations
	SetGlobalTracer(tracer Tracer)
	GetGlobalTracer() Tracer

	// Start/Stop
	Start() error
	Stop() error
}

// Span interface defines span operations
type Span interface {
	// Span information
	GetSpanContext() SpanContext
	GetOperationName() string
	GetStartTime() time.Time
	GetDuration() time.Duration

	// Span operations
	SetTag(key string, value interface{})
	SetTags(tags map[string]interface{})
	LogEvent(event string)
	LogFields(fields map[string]interface{})
	SetBaggageItem(key, value string)
	GetBaggageItem(key string) string

	// Span lifecycle
	Finish()
	FinishWithOptions(opts SpanFinishOptions)

	// Error handling
	SetError(err error)
	IsError() bool
}

// SpanContext represents span context information
type SpanContext struct {
	TraceID string
	SpanID  string
	Baggage map[string]string
}

// SpanOptions represents options for creating a span
type SpanOptions struct {
	Parent    Span
	Tags      map[string]interface{}
	StartTime time.Time
}

// SpanFinishOptions represents options for finishing a span
type SpanFinishOptions struct {
	FinishTime time.Time
	Error      error
}

// Logger interface defines logging operations
type Logger interface {
	// Log levels
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	// Formatted logging
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	// Structured logging
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger

	// Context logging
	WithContext(ctx context.Context) Logger

	// Level management
	SetLevel(level LogLevel)
	GetLevel() LogLevel
	IsLevelEnabled(level LogLevel) bool
}

// HealthChecker interface defines health check operations
type HealthChecker interface {
	// Health check operations
	CheckHealth(ctx context.Context) HealthStatus
	CheckDependencyHealth(ctx context.Context, dependency string) HealthStatus
	RegisterHealthCheck(name string, check HealthCheckFunc)
	UnregisterHealthCheck(name string)

	// Health status
	GetHealthStatus() map[string]HealthStatus
	IsHealthy() bool
	IsDependencyHealthy(dependency string) bool

	// Start/Stop
	Start() error
	Stop() error
}

// HealthCheckFunc defines the signature for health check functions
type HealthCheckFunc func(ctx context.Context) error

// Enums and types

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// HealthStatus represents the health status of a service
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// LogLevel represents the log level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)

// CircuitBreaker interface defines circuit breaker operations
type CircuitBreaker interface {
	// Circuit breaker operations
	Execute(ctx context.Context, operation func() (interface{}, error)) (interface{}, error)
	ExecuteAsync(ctx context.Context, operation func() (interface{}, error)) <-chan CircuitBreakerResult

	// State management
	GetState() CircuitBreakerState
	GetMetrics() CircuitBreakerMetrics
	Reset()

	// Configuration
	SetConfig(config CircuitBreakerConfig)
	GetConfig() CircuitBreakerConfig
}

// CircuitBreakerResult represents the result of a circuit breaker operation
type CircuitBreakerResult struct {
	Result interface{}
	Error  error
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerStateClosed   CircuitBreakerState = "closed"
	CircuitBreakerStateOpen     CircuitBreakerState = "open"
	CircuitBreakerStateHalfOpen CircuitBreakerState = "half-open"
)

// CircuitBreakerMetrics represents circuit breaker metrics
type CircuitBreakerMetrics struct {
	Requests             int64
	Successes            int64
	Failures             int64
	Rejections           int64
	State                CircuitBreakerState
	LastStateChange      time.Time
	ConsecutiveFailures  int64
	ConsecutiveSuccesses int64
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxRequests      int64
	Interval         time.Duration
	Timeout          time.Duration
	MaxFailures      int64
	SuccessThreshold int64
	FailureThreshold int64
	FailureRate      float64
	RecoveryTimeout  time.Duration
}

// RateLimiter interface defines rate limiting operations
type RateLimiter interface {
	// Rate limiting operations
	Allow(ctx context.Context, key string) (bool, error)
	AllowN(ctx context.Context, key string, n int) (bool, error)
	Wait(ctx context.Context, key string) error
	WaitN(ctx context.Context, key string, n int) error

	// Configuration
	SetLimit(limit int)
	SetBurst(burst int)
	SetWindow(window time.Duration)

	// Information
	GetLimit() int
	GetBurst() int
	GetWindow() time.Duration
	GetRemaining(ctx context.Context, key string) (int, error)
	GetResetTime(ctx context.Context, key string) (time.Time, error)
}
