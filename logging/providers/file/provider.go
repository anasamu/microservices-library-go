package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/logging/types"
	"github.com/sirupsen/logrus"
)

// FileProvider implements logging to file
type FileProvider struct {
	name      string
	logger    *logrus.Logger
	config    *types.LoggingConfig
	connected bool
	file      *os.File
	mu        sync.RWMutex
}

// NewFileProvider creates a new file logging provider
func NewFileProvider(config *types.LoggingConfig) *FileProvider {
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

	return &FileProvider{
		name:      "file",
		logger:    logger,
		config:    config,
		connected: false,
	}
}

// GetName returns the provider name
func (fp *FileProvider) GetName() string {
	return fp.name
}

// GetSupportedFeatures returns supported features
func (fp *FileProvider) GetSupportedFeatures() []types.LoggingFeature {
	return []types.LoggingFeature{
		types.FeatureBasicLogging,
		types.FeatureStructuredLog,
		types.FeatureFormattedLog,
		types.FeatureContextLogging,
		types.FeatureBatchLogging,
		types.FeatureRetention,
	}
}

// GetConnectionInfo returns connection information
func (fp *FileProvider) GetConnectionInfo() *types.ConnectionInfo {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	status := types.StatusDisconnected
	if fp.connected && fp.file != nil {
		status = types.StatusConnected
	}

	filePath := "app.log"
	if fp.config.Metadata != nil {
		if path, ok := fp.config.Metadata["file_path"].(string); ok {
			filePath = path
		}
	}

	return &types.ConnectionInfo{
		Host:     "file",
		Port:     0,
		Database: filePath,
		Username: "",
		Status:   status,
		Metadata: map[string]string{
			"format": string(fp.config.Format),
			"path":   filePath,
		},
	}
}

// Connect connects to the file
func (fp *FileProvider) Connect(ctx context.Context) error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	filePath := "app.log"
	if fp.config.Metadata != nil {
		if path, ok := fp.config.Metadata["file_path"].(string); ok {
			filePath = path
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open file for writing
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	fp.file = file
	fp.logger.SetOutput(file)
	fp.connected = true

	return nil
}

// Disconnect disconnects from the file
func (fp *FileProvider) Disconnect(ctx context.Context) error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	if fp.file != nil {
		if err := fp.file.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
		fp.file = nil
	}

	fp.connected = false
	return nil
}

// Ping checks if the file is available
func (fp *FileProvider) Ping(ctx context.Context) error {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	if !fp.connected || fp.file == nil {
		return fmt.Errorf("file provider not connected")
	}

	// Try to write a test entry
	_, err := fp.file.WriteString("")
	return err
}

// IsConnected returns connection status
func (fp *FileProvider) IsConnected() bool {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.connected && fp.file != nil
}

// Log logs a message with the specified level
func (fp *FileProvider) Log(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	if !fp.connected || fp.file == nil {
		return fmt.Errorf("file provider not connected")
	}

	logLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		logLevel = logrus.InfoLevel
	}

	entry := fp.logger.WithFields(logrus.Fields{
		"service": fp.config.Service,
		"version": fp.config.Version,
	})

	if fields != nil {
		entry = entry.WithFields(fields)
	}

	entry.Log(logLevel, message)
	return nil
}

// LogWithContext logs a message with context
func (fp *FileProvider) LogWithContext(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	if !fp.IsConnected() {
		return fmt.Errorf("file provider not connected")
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

	return fp.Log(ctx, level, message, contextFields)
}

// Info logs an info level message
func (fp *FileProvider) Info(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fp.LogWithContext(ctx, types.LevelInfo, message, mergedFields)
}

// Debug logs a debug level message
func (fp *FileProvider) Debug(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fp.LogWithContext(ctx, types.LevelDebug, message, mergedFields)
}

// Warn logs a warning level message
func (fp *FileProvider) Warn(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fp.LogWithContext(ctx, types.LevelWarn, message, mergedFields)
}

// Error logs an error level message
func (fp *FileProvider) Error(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fp.LogWithContext(ctx, types.LevelError, message, mergedFields)
}

// Fatal logs a fatal level message
func (fp *FileProvider) Fatal(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fp.LogWithContext(ctx, types.LevelFatal, message, mergedFields)
}

// Panic logs a panic level message
func (fp *FileProvider) Panic(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fp.LogWithContext(ctx, types.LevelPanic, message, mergedFields)
}

// Infof logs a formatted info level message
func (fp *FileProvider) Infof(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fp.LogWithContext(ctx, types.LevelInfo, message, nil)
}

// Debugf logs a formatted debug level message
func (fp *FileProvider) Debugf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fp.LogWithContext(ctx, types.LevelDebug, message, nil)
}

// Warnf logs a formatted warning level message
func (fp *FileProvider) Warnf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fp.LogWithContext(ctx, types.LevelWarn, message, nil)
}

// Errorf logs a formatted error level message
func (fp *FileProvider) Errorf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fp.LogWithContext(ctx, types.LevelError, message, nil)
}

// Fatalf logs a formatted fatal level message
func (fp *FileProvider) Fatalf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fp.LogWithContext(ctx, types.LevelFatal, message, nil)
}

// Panicf logs a formatted panic level message
func (fp *FileProvider) Panicf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fp.LogWithContext(ctx, types.LevelPanic, message, nil)
}

// WithFields creates a new provider with additional fields
func (fp *FileProvider) WithFields(ctx context.Context, fields map[string]interface{}) types.LoggingProvider {
	return &FileProviderWithFields{
		provider: fp,
		fields:   fields,
	}
}

// WithField creates a new provider with a single additional field
func (fp *FileProvider) WithField(ctx context.Context, key string, value interface{}) types.LoggingProvider {
	return fp.WithFields(ctx, map[string]interface{}{key: value})
}

// WithError creates a new provider with error information
func (fp *FileProvider) WithError(ctx context.Context, err error) types.LoggingProvider {
	return fp.WithField(ctx, "error", err.Error())
}

// WithContext creates a new provider with context information
func (fp *FileProvider) WithContext(ctx context.Context) types.LoggingProvider {
	return fp
}

// LogBatch logs multiple entries
func (fp *FileProvider) LogBatch(ctx context.Context, entries []types.LogEntry) error {
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

		if err := fp.Log(ctx, entry.Level, entry.Message, fields); err != nil {
			return err
		}
	}
	return nil
}

// Search is not supported for file provider
func (fp *FileProvider) Search(ctx context.Context, query types.LogQuery) ([]types.LogEntry, error) {
	return nil, fmt.Errorf("search not supported for file provider")
}

// GetStats returns basic statistics
func (fp *FileProvider) GetStats(ctx context.Context) (*types.LoggingStats, error) {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	var fileSize int64
	if fp.file != nil {
		if stat, err := fp.file.Stat(); err == nil {
			fileSize = stat.Size()
		}
	}

	return &types.LoggingStats{
		TotalLogs:     0, // File doesn't track this
		LogsByLevel:   make(map[types.LogLevel]int64),
		LogsByService: make(map[string]int64),
		StorageSize:   fileSize,
		Uptime:        time.Since(time.Now()), // Placeholder
		LastUpdate:    time.Now(),
		Provider:      fp.name,
	}, nil
}

// FileProviderWithFields wraps FileProvider with additional fields
type FileProviderWithFields struct {
	provider *FileProvider
	fields   map[string]interface{}
}

// Implement all LoggingProvider methods by delegating to the wrapped provider
// and merging the additional fields

func (fpwf *FileProviderWithFields) GetName() string {
	return fpwf.provider.GetName()
}

func (fpwf *FileProviderWithFields) GetSupportedFeatures() []types.LoggingFeature {
	return fpwf.provider.GetSupportedFeatures()
}

func (fpwf *FileProviderWithFields) GetConnectionInfo() *types.ConnectionInfo {
	return fpwf.provider.GetConnectionInfo()
}

func (fpwf *FileProviderWithFields) Connect(ctx context.Context) error {
	return fpwf.provider.Connect(ctx)
}

func (fpwf *FileProviderWithFields) Disconnect(ctx context.Context) error {
	return fpwf.provider.Disconnect(ctx)
}

func (fpwf *FileProviderWithFields) Ping(ctx context.Context) error {
	return fpwf.provider.Ping(ctx)
}

func (fpwf *FileProviderWithFields) IsConnected() bool {
	return fpwf.provider.IsConnected()
}

func (fpwf *FileProviderWithFields) Log(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	mergedFields := make(map[string]interface{})

	// Add wrapper fields first
	for k, v := range fpwf.fields {
		mergedFields[k] = v
	}

	// Add provided fields (they override wrapper fields)
	if fields != nil {
		for k, v := range fields {
			mergedFields[k] = v
		}
	}

	return fpwf.provider.Log(ctx, level, message, mergedFields)
}

func (fpwf *FileProviderWithFields) LogWithContext(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	mergedFields := make(map[string]interface{})

	// Add wrapper fields first
	for k, v := range fpwf.fields {
		mergedFields[k] = v
	}

	// Add provided fields (they override wrapper fields)
	if fields != nil {
		for k, v := range fields {
			mergedFields[k] = v
		}
	}

	return fpwf.provider.LogWithContext(ctx, level, message, mergedFields)
}

func (fpwf *FileProviderWithFields) Info(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fpwf.LogWithContext(ctx, types.LevelInfo, message, mergedFields)
}

func (fpwf *FileProviderWithFields) Debug(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fpwf.LogWithContext(ctx, types.LevelDebug, message, mergedFields)
}

func (fpwf *FileProviderWithFields) Warn(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fpwf.LogWithContext(ctx, types.LevelWarn, message, mergedFields)
}

func (fpwf *FileProviderWithFields) Error(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fpwf.LogWithContext(ctx, types.LevelError, message, mergedFields)
}

func (fpwf *FileProviderWithFields) Fatal(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fpwf.LogWithContext(ctx, types.LevelFatal, message, mergedFields)
}

func (fpwf *FileProviderWithFields) Panic(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return fpwf.LogWithContext(ctx, types.LevelPanic, message, mergedFields)
}

func (fpwf *FileProviderWithFields) Infof(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fpwf.LogWithContext(ctx, types.LevelInfo, message, nil)
}

func (fpwf *FileProviderWithFields) Debugf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fpwf.LogWithContext(ctx, types.LevelDebug, message, nil)
}

func (fpwf *FileProviderWithFields) Warnf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fpwf.LogWithContext(ctx, types.LevelWarn, message, nil)
}

func (fpwf *FileProviderWithFields) Errorf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fpwf.LogWithContext(ctx, types.LevelError, message, nil)
}

func (fpwf *FileProviderWithFields) Fatalf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fpwf.LogWithContext(ctx, types.LevelFatal, message, nil)
}

func (fpwf *FileProviderWithFields) Panicf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fpwf.LogWithContext(ctx, types.LevelPanic, message, nil)
}

func (fpwf *FileProviderWithFields) WithFields(ctx context.Context, fields map[string]interface{}) types.LoggingProvider {
	// Merge fields
	mergedFields := make(map[string]interface{})
	for k, v := range fpwf.fields {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}

	return &FileProviderWithFields{
		provider: fpwf.provider,
		fields:   mergedFields,
	}
}

func (fpwf *FileProviderWithFields) WithField(ctx context.Context, key string, value interface{}) types.LoggingProvider {
	return fpwf.WithFields(ctx, map[string]interface{}{key: value})
}

func (fpwf *FileProviderWithFields) WithError(ctx context.Context, err error) types.LoggingProvider {
	return fpwf.WithField(ctx, "error", err.Error())
}

func (fpwf *FileProviderWithFields) WithContext(ctx context.Context) types.LoggingProvider {
	return fpwf
}

func (fpwf *FileProviderWithFields) LogBatch(ctx context.Context, entries []types.LogEntry) error {
	return fpwf.provider.LogBatch(ctx, entries)
}

func (fpwf *FileProviderWithFields) Search(ctx context.Context, query types.LogQuery) ([]types.LogEntry, error) {
	return fpwf.provider.Search(ctx, query)
}

func (fpwf *FileProviderWithFields) GetStats(ctx context.Context) (*types.LoggingStats, error) {
	return fpwf.provider.GetStats(ctx)
}
