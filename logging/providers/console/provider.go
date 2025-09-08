package console

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/anasamu/microservices-library-go/logging/types"
	"github.com/sirupsen/logrus"
)

// ConsoleProvider implements logging to console/stdout
type ConsoleProvider struct {
	name      string
	logger    *logrus.Logger
	config    *types.LoggingConfig
	connected bool
}

// NewConsoleProvider creates a new console logging provider
func NewConsoleProvider(config *types.LoggingConfig) *ConsoleProvider {
	if config == nil {
		config = types.GetDefaultConfig()
	}

	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(string(config.Level))
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set formatter
	switch config.Format {
	case types.FormatJSON:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	case types.FormatText:
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
	switch config.Output {
	case types.OutputStdout:
		logger.SetOutput(os.Stdout)
	case types.OutputStderr:
		logger.SetOutput(os.Stderr)
	default:
		logger.SetOutput(os.Stdout)
	}

	return &ConsoleProvider{
		name:      "console",
		logger:    logger,
		config:    config,
		connected: true,
	}
}

// GetName returns the provider name
func (cp *ConsoleProvider) GetName() string {
	return cp.name
}

// GetSupportedFeatures returns supported features
func (cp *ConsoleProvider) GetSupportedFeatures() []types.LoggingFeature {
	return []types.LoggingFeature{
		types.FeatureBasicLogging,
		types.FeatureStructuredLog,
		types.FeatureFormattedLog,
		types.FeatureContextLogging,
	}
}

// GetConnectionInfo returns connection information
func (cp *ConsoleProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "console",
		Port:     0,
		Database: "",
		Username: "",
		Status:   types.StatusConnected,
		Metadata: map[string]string{
			"output": string(cp.config.Output),
			"format": string(cp.config.Format),
		},
	}
}

// Connect connects to the console (always successful)
func (cp *ConsoleProvider) Connect(ctx context.Context) error {
	cp.connected = true
	return nil
}

// Disconnect disconnects from the console
func (cp *ConsoleProvider) Disconnect(ctx context.Context) error {
	cp.connected = false
	return nil
}

// Ping checks if the console is available
func (cp *ConsoleProvider) Ping(ctx context.Context) error {
	if !cp.connected {
		return fmt.Errorf("console provider not connected")
	}
	return nil
}

// IsConnected returns connection status
func (cp *ConsoleProvider) IsConnected() bool {
	return cp.connected
}

// Log logs a message with the specified level
func (cp *ConsoleProvider) Log(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	if !cp.connected {
		return fmt.Errorf("console provider not connected")
	}

	logLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		logLevel = logrus.InfoLevel
	}

	entry := cp.logger.WithFields(logrus.Fields{
		"service": cp.config.Service,
		"version": cp.config.Version,
	})

	if fields != nil {
		entry = entry.WithFields(fields)
	}

	entry.Log(logLevel, message)
	return nil
}

// LogWithContext logs a message with context
func (cp *ConsoleProvider) LogWithContext(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	if !cp.connected {
		return fmt.Errorf("console provider not connected")
	}

	// Add context fields
	contextFields := make(map[string]interface{})
	if traceID := ctx.Value("trace_id"); traceID != nil {
		contextFields["trace_id"] = traceID
	}
	if spanID := ctx.Value("span_id"); spanID != nil {
		contextFields["span_id"] = spanID
	}
	if requestID := ctx.Value("request_id"); requestID != nil {
		contextFields["request_id"] = requestID
	}
	if userID := ctx.Value("user_id"); userID != nil {
		contextFields["user_id"] = userID
	}

	// Merge with provided fields
	if fields != nil {
		for k, v := range fields {
			contextFields[k] = v
		}
	}

	return cp.Log(ctx, level, message, contextFields)
}

// Info logs an info level message
func (cp *ConsoleProvider) Info(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cp.LogWithContext(ctx, types.LevelInfo, message, mergedFields)
}

// Debug logs a debug level message
func (cp *ConsoleProvider) Debug(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cp.LogWithContext(ctx, types.LevelDebug, message, mergedFields)
}

// Warn logs a warning level message
func (cp *ConsoleProvider) Warn(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cp.LogWithContext(ctx, types.LevelWarn, message, mergedFields)
}

// Error logs an error level message
func (cp *ConsoleProvider) Error(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cp.LogWithContext(ctx, types.LevelError, message, mergedFields)
}

// Fatal logs a fatal level message
func (cp *ConsoleProvider) Fatal(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cp.LogWithContext(ctx, types.LevelFatal, message, mergedFields)
}

// Panic logs a panic level message
func (cp *ConsoleProvider) Panic(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cp.LogWithContext(ctx, types.LevelPanic, message, mergedFields)
}

// Infof logs a formatted info level message
func (cp *ConsoleProvider) Infof(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cp.LogWithContext(ctx, types.LevelInfo, message, nil)
}

// Debugf logs a formatted debug level message
func (cp *ConsoleProvider) Debugf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cp.LogWithContext(ctx, types.LevelDebug, message, nil)
}

// Warnf logs a formatted warning level message
func (cp *ConsoleProvider) Warnf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cp.LogWithContext(ctx, types.LevelWarn, message, nil)
}

// Errorf logs a formatted error level message
func (cp *ConsoleProvider) Errorf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cp.LogWithContext(ctx, types.LevelError, message, nil)
}

// Fatalf logs a formatted fatal level message
func (cp *ConsoleProvider) Fatalf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cp.LogWithContext(ctx, types.LevelFatal, message, nil)
}

// Panicf logs a formatted panic level message
func (cp *ConsoleProvider) Panicf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cp.LogWithContext(ctx, types.LevelPanic, message, nil)
}

// WithFields creates a new provider with additional fields
func (cp *ConsoleProvider) WithFields(ctx context.Context, fields map[string]interface{}) types.LoggingProvider {
	// For console provider, we'll return a wrapper that adds fields to each log
	return &ConsoleProviderWithFields{
		provider: cp,
		fields:   fields,
	}
}

// WithField creates a new provider with a single additional field
func (cp *ConsoleProvider) WithField(ctx context.Context, key string, value interface{}) types.LoggingProvider {
	return cp.WithFields(ctx, map[string]interface{}{key: value})
}

// WithError creates a new provider with error information
func (cp *ConsoleProvider) WithError(ctx context.Context, err error) types.LoggingProvider {
	return cp.WithField(ctx, "error", err.Error())
}

// WithContext creates a new provider with context information
func (cp *ConsoleProvider) WithContext(ctx context.Context) types.LoggingProvider {
	// Console provider already handles context in LogWithContext
	return cp
}

// LogBatch logs multiple entries
func (cp *ConsoleProvider) LogBatch(ctx context.Context, entries []types.LogEntry) error {
	for _, entry := range entries {
		fields := make(map[string]interface{})
		if entry.Fields != nil {
			fields = entry.Fields
		}

		// Add entry fields
		if entry.TraceID != "" {
			fields["trace_id"] = entry.TraceID
		}
		if entry.SpanID != "" {
			fields["span_id"] = entry.SpanID
		}
		if entry.UserID != "" {
			fields["user_id"] = entry.UserID
		}
		if entry.TenantID != "" {
			fields["tenant_id"] = entry.TenantID
		}
		if entry.RequestID != "" {
			fields["request_id"] = entry.RequestID
		}
		if entry.IPAddress != "" {
			fields["ip_address"] = entry.IPAddress
		}
		if entry.UserAgent != "" {
			fields["user_agent"] = entry.UserAgent
		}
		if entry.Method != "" {
			fields["method"] = entry.Method
		}
		if entry.Path != "" {
			fields["path"] = entry.Path
		}
		if entry.StatusCode != 0 {
			fields["status_code"] = entry.StatusCode
		}
		if entry.Duration != 0 {
			fields["duration"] = entry.Duration
		}
		if entry.Error != "" {
			fields["error"] = entry.Error
		}

		if err := cp.Log(ctx, entry.Level, entry.Message, fields); err != nil {
			return err
		}
	}
	return nil
}

// Search is not supported for console provider
func (cp *ConsoleProvider) Search(ctx context.Context, query types.LogQuery) ([]types.LogEntry, error) {
	return nil, fmt.Errorf("search not supported for console provider")
}

// GetStats returns basic statistics
func (cp *ConsoleProvider) GetStats(ctx context.Context) (*types.LoggingStats, error) {
	return &types.LoggingStats{
		TotalLogs:     0, // Console doesn't track this
		LogsByLevel:   make(map[types.LogLevel]int64),
		LogsByService: make(map[string]int64),
		StorageSize:   0,
		Uptime:        time.Since(time.Now()), // Placeholder
		LastUpdate:    time.Now(),
		Provider:      cp.name,
	}, nil
}

// ConsoleProviderWithFields wraps ConsoleProvider with additional fields
type ConsoleProviderWithFields struct {
	provider *ConsoleProvider
	fields   map[string]interface{}
}

// Implement all LoggingProvider methods by delegating to the wrapped provider
// and merging the additional fields

func (cpwf *ConsoleProviderWithFields) GetName() string {
	return cpwf.provider.GetName()
}

func (cpwf *ConsoleProviderWithFields) GetSupportedFeatures() []types.LoggingFeature {
	return cpwf.provider.GetSupportedFeatures()
}

func (cpwf *ConsoleProviderWithFields) GetConnectionInfo() *types.ConnectionInfo {
	return cpwf.provider.GetConnectionInfo()
}

func (cpwf *ConsoleProviderWithFields) Connect(ctx context.Context) error {
	return cpwf.provider.Connect(ctx)
}

func (cpwf *ConsoleProviderWithFields) Disconnect(ctx context.Context) error {
	return cpwf.provider.Disconnect(ctx)
}

func (cpwf *ConsoleProviderWithFields) Ping(ctx context.Context) error {
	return cpwf.provider.Ping(ctx)
}

func (cpwf *ConsoleProviderWithFields) IsConnected() bool {
	return cpwf.provider.IsConnected()
}

func (cpwf *ConsoleProviderWithFields) Log(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	mergedFields := make(map[string]interface{})

	// Add wrapper fields first
	for k, v := range cpwf.fields {
		mergedFields[k] = v
	}

	// Add provided fields (they override wrapper fields)
	if fields != nil {
		for k, v := range fields {
			mergedFields[k] = v
		}
	}

	return cpwf.provider.Log(ctx, level, message, mergedFields)
}

func (cpwf *ConsoleProviderWithFields) LogWithContext(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	mergedFields := make(map[string]interface{})

	// Add wrapper fields first
	for k, v := range cpwf.fields {
		mergedFields[k] = v
	}

	// Add provided fields (they override wrapper fields)
	if fields != nil {
		for k, v := range fields {
			mergedFields[k] = v
		}
	}

	return cpwf.provider.LogWithContext(ctx, level, message, mergedFields)
}

func (cpwf *ConsoleProviderWithFields) Info(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cpwf.LogWithContext(ctx, types.LevelInfo, message, mergedFields)
}

func (cpwf *ConsoleProviderWithFields) Debug(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cpwf.LogWithContext(ctx, types.LevelDebug, message, mergedFields)
}

func (cpwf *ConsoleProviderWithFields) Warn(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cpwf.LogWithContext(ctx, types.LevelWarn, message, mergedFields)
}

func (cpwf *ConsoleProviderWithFields) Error(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cpwf.LogWithContext(ctx, types.LevelError, message, mergedFields)
}

func (cpwf *ConsoleProviderWithFields) Fatal(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cpwf.LogWithContext(ctx, types.LevelFatal, message, mergedFields)
}

func (cpwf *ConsoleProviderWithFields) Panic(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return cpwf.LogWithContext(ctx, types.LevelPanic, message, mergedFields)
}

func (cpwf *ConsoleProviderWithFields) Infof(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cpwf.LogWithContext(ctx, types.LevelInfo, message, nil)
}

func (cpwf *ConsoleProviderWithFields) Debugf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cpwf.LogWithContext(ctx, types.LevelDebug, message, nil)
}

func (cpwf *ConsoleProviderWithFields) Warnf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cpwf.LogWithContext(ctx, types.LevelWarn, message, nil)
}

func (cpwf *ConsoleProviderWithFields) Errorf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cpwf.LogWithContext(ctx, types.LevelError, message, nil)
}

func (cpwf *ConsoleProviderWithFields) Fatalf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cpwf.LogWithContext(ctx, types.LevelFatal, message, nil)
}

func (cpwf *ConsoleProviderWithFields) Panicf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return cpwf.LogWithContext(ctx, types.LevelPanic, message, nil)
}

func (cpwf *ConsoleProviderWithFields) WithFields(ctx context.Context, fields map[string]interface{}) types.LoggingProvider {
	// Merge fields
	mergedFields := make(map[string]interface{})
	for k, v := range cpwf.fields {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}

	return &ConsoleProviderWithFields{
		provider: cpwf.provider,
		fields:   mergedFields,
	}
}

func (cpwf *ConsoleProviderWithFields) WithField(ctx context.Context, key string, value interface{}) types.LoggingProvider {
	return cpwf.WithFields(ctx, map[string]interface{}{key: value})
}

func (cpwf *ConsoleProviderWithFields) WithError(ctx context.Context, err error) types.LoggingProvider {
	return cpwf.WithField(ctx, "error", err.Error())
}

func (cpwf *ConsoleProviderWithFields) WithContext(ctx context.Context) types.LoggingProvider {
	return cpwf
}

func (cpwf *ConsoleProviderWithFields) LogBatch(ctx context.Context, entries []types.LogEntry) error {
	return cpwf.provider.LogBatch(ctx, entries)
}

func (cpwf *ConsoleProviderWithFields) Search(ctx context.Context, query types.LogQuery) ([]types.LogEntry, error) {
	return cpwf.provider.Search(ctx, query)
}

func (cpwf *ConsoleProviderWithFields) GetStats(ctx context.Context) (*types.LoggingStats, error) {
	return cpwf.provider.GetStats(ctx)
}
