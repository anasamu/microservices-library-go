package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
)

// Logger represents the application logger
type Logger struct {
	*logrus.Logger
	elasticClient *elasticsearch.Client
	config        *LoggingConfig
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string
	Format     string
	Output     string
	ElasticURL string
	Index      string
	Service    string
	Version    string
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
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
}

// NewLogger creates a new logger instance
func NewLogger(config *LoggingConfig) (*Logger, error) {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set formatter
	switch strings.ToLower(config.Format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	}

	// Set output
	switch strings.ToLower(config.Output) {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	case "file":
		file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		logger.SetOutput(file)
	default:
		logger.SetOutput(os.Stdout)
	}

	l := &Logger{
		Logger: logger,
		config: config,
	}

	// Initialize Elasticsearch client if configured
	if config.ElasticURL != "" {
		esClient, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: []string{config.ElasticURL},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
		}
		l.elasticClient = esClient
	}

	return l, nil
}

// WithContext creates a logger with context information
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.Logger.WithContext(ctx)

	// Add trace information if available
	if traceID := ctx.Value("trace_id"); traceID != nil {
		entry = entry.WithField("trace_id", traceID)
	}
	if spanID := ctx.Value("span_id"); spanID != nil {
		entry = entry.WithField("span_id", spanID)
	}
	if requestID := ctx.Value("request_id"); requestID != nil {
		entry = entry.WithField("request_id", requestID)
	}
	if userID := ctx.Value("user_id"); userID != nil {
		entry = entry.WithField("user_id", userID)
	}
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		entry = entry.WithField("tenant_id", tenantID)
	}

	// Add service information
	entry = entry.WithFields(logrus.Fields{
		"service": l.config.Service,
		"version": l.config.Version,
	})

	return entry
}

// WithFields creates a logger with additional fields
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	entry := l.Logger.WithFields(fields)

	// Add service information
	entry = entry.WithFields(logrus.Fields{
		"service": l.config.Service,
		"version": l.config.Version,
	})

	return entry
}

// WithField creates a logger with a single field
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.WithFields(logrus.Fields{key: value})
}

// WithError creates a logger with error information
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.WithField("error", err.Error())
}

// WithTenant creates a logger with tenant information
func (l *Logger) WithTenant(tenantID string) *logrus.Entry {
	return l.WithField("tenant_id", tenantID)
}

// WithUser creates a logger with user information
func (l *Logger) WithUser(userID string) *logrus.Entry {
	return l.WithField("user_id", userID)
}

// WithRequest creates a logger with request information
func (l *Logger) WithRequest(method, path, requestID string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"method":     method,
		"path":       path,
		"request_id": requestID,
	})
}

// WithResponse creates a logger with response information
func (l *Logger) WithResponse(statusCode int, duration int64) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"status_code": statusCode,
		"duration":    duration,
	})
}

// LogToElasticsearch sends log entry to Elasticsearch
func (l *Logger) LogToElasticsearch(entry *LogEntry) error {
	if l.elasticClient == nil {
		return nil
	}

	// Create index name with date suffix
	indexName := fmt.Sprintf("%s-%s", l.config.Index, time.Now().Format("2006.01.02"))

	// Marshal log entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Send to Elasticsearch
	res, err := l.elasticClient.Index(
		indexName,
		strings.NewReader(string(data)),
		l.elasticClient.Index.WithRefresh("false"),
	)
	if err != nil {
		return fmt.Errorf("failed to send log to elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch error: %s", res.String())
	}

	return nil
}

// Info logs an info level message
func (l *Logger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

// Infof logs a formatted info level message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logger.Infof(format, args...)
}

// Debug logs a debug level message
func (l *Logger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

// Debugf logs a formatted debug level message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

// Warn logs a warning level message
func (l *Logger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

// Warnf logs a formatted warning level message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Logger.Warnf(format, args...)
}

// Error logs an error level message
func (l *Logger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

// Errorf logs a formatted error level message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

// Fatal logs a fatal level message and exits
func (l *Logger) Fatal(args ...interface{}) {
	l.Logger.Fatal(args...)
}

// Fatalf logs a formatted fatal level message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Logger.Fatalf(format, args...)
}

// Panic logs a panic level message and panics
func (l *Logger) Panic(args ...interface{}) {
	l.Logger.Panic(args...)
}

// Panicf logs a formatted panic level message and panics
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.Logger.Panicf(format, args...)
}

// LogRequest logs HTTP request information
func (l *Logger) LogRequest(ctx context.Context, method, path string, statusCode int, duration int64, err error) {
	entry := l.WithContext(ctx).WithFields(logrus.Fields{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration":    duration,
	})

	if err != nil {
		entry.WithError(err).Error("HTTP request failed")
	} else {
		entry.Info("HTTP request completed")
	}
}

// LogDatabase logs database operation information
func (l *Logger) LogDatabase(ctx context.Context, operation, table string, duration int64, err error) {
	entry := l.WithContext(ctx).WithFields(logrus.Fields{
		"operation": operation,
		"table":     table,
		"duration":  duration,
	})

	if err != nil {
		entry.WithError(err).Error("Database operation failed")
	} else {
		entry.Debug("Database operation completed")
	}
}

// LogBusiness logs business logic information
func (l *Logger) LogBusiness(ctx context.Context, action string, entity string, entityID string, err error) {
	entry := l.WithContext(ctx).WithFields(logrus.Fields{
		"action":    action,
		"entity":    entity,
		"entity_id": entityID,
	})

	if err != nil {
		entry.WithError(err).Error("Business operation failed")
	} else {
		entry.Info("Business operation completed")
	}
}

// LogSecurity logs security-related events
func (l *Logger) LogSecurity(ctx context.Context, event string, userID string, ipAddress string, details map[string]interface{}) {
	entry := l.WithContext(ctx).WithFields(logrus.Fields{
		"event":      event,
		"user_id":    userID,
		"ip_address": ipAddress,
	})

	if details != nil {
		entry = entry.WithFields(details)
	}

	entry.Warn("Security event occurred")
}

// LogPerformance logs performance-related information
func (l *Logger) LogPerformance(ctx context.Context, operation string, duration int64, metrics map[string]interface{}) {
	entry := l.WithContext(ctx).WithFields(logrus.Fields{
		"operation": operation,
		"duration":  duration,
	})

	if metrics != nil {
		entry = entry.WithFields(metrics)
	}

	entry.Info("Performance metric recorded")
}

// Close closes the logger and any associated resources
func (l *Logger) Close() error {
	// Close any file handles or connections
	return nil
}
