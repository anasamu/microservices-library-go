package monitoring

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/monitoring/types"
	"github.com/sirupsen/logrus"
)

// MonitoringManager manages multiple monitoring providers
type MonitoringManager struct {
	providers map[string]MonitoringProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds monitoring manager configuration
type ManagerConfig struct {
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

// MonitoringProvider interface for monitoring backends
type MonitoringProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []types.MonitoringFeature
	GetConnectionInfo() *types.ConnectionInfo

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Metrics operations
	SubmitMetrics(ctx context.Context, request *types.MetricRequest) error
	QueryMetrics(ctx context.Context, request *types.QueryRequest) (*types.QueryResponse, error)

	// Logging operations
	SubmitLogs(ctx context.Context, request *types.LogRequest) error
	QueryLogs(ctx context.Context, request *types.QueryRequest) (*types.QueryResponse, error)

	// Tracing operations
	SubmitTraces(ctx context.Context, request *types.TraceRequest) error
	QueryTraces(ctx context.Context, request *types.QueryRequest) (*types.QueryResponse, error)

	// Alerting operations
	SubmitAlerts(ctx context.Context, request *types.AlertRequest) error
	QueryAlerts(ctx context.Context, request *types.QueryRequest) (*types.QueryResponse, error)

	// Health check operations
	HealthCheck(ctx context.Context, request *types.HealthCheckRequest) (*types.HealthCheckResponse, error)

	// Statistics and monitoring
	GetStats(ctx context.Context) (*types.MonitoringStats, error)

	// Configuration
	Configure(config map[string]interface{}) error
	IsConfigured() bool
	Close() error
}

// DefaultManagerConfig returns default monitoring manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		DefaultProvider: "prometheus",
		RetryAttempts:   3,
		RetryDelay:      5 * time.Second,
		Timeout:         30 * time.Second,
		BatchSize:       1000,
		FlushInterval:   10 * time.Second,
		BufferSize:      10000,
		SamplingRate:    1.0,
		EnableTracing:   true,
		EnableMetrics:   true,
		EnableLogging:   true,
		EnableAlerting:  true,
		Metadata:        make(map[string]string),
	}
}

// NewMonitoringManager creates a new monitoring manager
func NewMonitoringManager(config *ManagerConfig, logger *logrus.Logger) *MonitoringManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &MonitoringManager{
		providers: make(map[string]MonitoringProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a monitoring provider
func (mm *MonitoringManager) RegisterProvider(provider MonitoringProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	mm.providers[name] = provider
	mm.logger.WithField("provider", name).Info("Monitoring provider registered")

	return nil
}

// GetProvider returns a monitoring provider by name
func (mm *MonitoringManager) GetProvider(name string) (MonitoringProvider, error) {
	provider, exists := mm.providers[name]
	if !exists {
		return nil, fmt.Errorf("monitoring provider not found: %s", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default monitoring provider
func (mm *MonitoringManager) GetDefaultProvider() (MonitoringProvider, error) {
	return mm.GetProvider(mm.config.DefaultProvider)
}

// Connect connects to a monitoring system using the specified provider
func (mm *MonitoringManager) Connect(ctx context.Context, providerName string) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	// Connect with retry logic
	for attempt := 1; attempt <= mm.config.RetryAttempts; attempt++ {
		err = provider.Connect(ctx)
		if err == nil {
			break
		}

		mm.logger.WithError(err).WithFields(logrus.Fields{
			"provider": providerName,
			"attempt":  attempt,
		}).Warn("Monitoring connection failed, retrying")

		if attempt < mm.config.RetryAttempts {
			time.Sleep(mm.config.RetryDelay)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to monitoring system after %d attempts: %w", mm.config.RetryAttempts, err)
	}

	mm.logger.WithField("provider", providerName).Info("Monitoring system connected successfully")
	return nil
}

// Disconnect disconnects from a monitoring system using the specified provider
func (mm *MonitoringManager) Disconnect(ctx context.Context, providerName string) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("failed to disconnect from monitoring system: %w", err)
	}

	mm.logger.WithField("provider", providerName).Info("Monitoring system disconnected successfully")
	return nil
}

// Ping pings a monitoring system using the specified provider
func (mm *MonitoringManager) Ping(ctx context.Context, providerName string) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping monitoring system: %w", err)
	}

	return nil
}

// SubmitMetrics submits metrics using the specified provider
func (mm *MonitoringManager) SubmitMetrics(ctx context.Context, providerName string, request *types.MetricRequest) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	// Validate request
	if err := mm.validateMetricRequest(request); err != nil {
		return fmt.Errorf("invalid metric request: %w", err)
	}

	// Submit with retry logic
	for attempt := 1; attempt <= mm.config.RetryAttempts; attempt++ {
		err = provider.SubmitMetrics(ctx, request)
		if err == nil {
			break
		}

		mm.logger.WithError(err).WithFields(logrus.Fields{
			"provider": providerName,
			"attempt":  attempt,
			"count":    len(request.Metrics),
		}).Warn("Metrics submission failed, retrying")

		if attempt < mm.config.RetryAttempts {
			time.Sleep(mm.config.RetryDelay)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to submit metrics after %d attempts: %w", mm.config.RetryAttempts, err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"count":    len(request.Metrics),
	}).Info("Metrics submitted successfully")

	return nil
}

// SubmitLogs submits logs using the specified provider
func (mm *MonitoringManager) SubmitLogs(ctx context.Context, providerName string, request *types.LogRequest) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	// Validate request
	if err := mm.validateLogRequest(request); err != nil {
		return fmt.Errorf("invalid log request: %w", err)
	}

	// Submit with retry logic
	for attempt := 1; attempt <= mm.config.RetryAttempts; attempt++ {
		err = provider.SubmitLogs(ctx, request)
		if err == nil {
			break
		}

		mm.logger.WithError(err).WithFields(logrus.Fields{
			"provider": providerName,
			"attempt":  attempt,
			"count":    len(request.Logs),
		}).Warn("Logs submission failed, retrying")

		if attempt < mm.config.RetryAttempts {
			time.Sleep(mm.config.RetryDelay)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to submit logs after %d attempts: %w", mm.config.RetryAttempts, err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"count":    len(request.Logs),
	}).Info("Logs submitted successfully")

	return nil
}

// SubmitTraces submits traces using the specified provider
func (mm *MonitoringManager) SubmitTraces(ctx context.Context, providerName string, request *types.TraceRequest) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	// Validate request
	if err := mm.validateTraceRequest(request); err != nil {
		return fmt.Errorf("invalid trace request: %w", err)
	}

	// Submit with retry logic
	for attempt := 1; attempt <= mm.config.RetryAttempts; attempt++ {
		err = provider.SubmitTraces(ctx, request)
		if err == nil {
			break
		}

		mm.logger.WithError(err).WithFields(logrus.Fields{
			"provider": providerName,
			"attempt":  attempt,
			"count":    len(request.Traces),
		}).Warn("Traces submission failed, retrying")

		if attempt < mm.config.RetryAttempts {
			time.Sleep(mm.config.RetryDelay)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to submit traces after %d attempts: %w", mm.config.RetryAttempts, err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"count":    len(request.Traces),
	}).Info("Traces submitted successfully")

	return nil
}

// SubmitAlerts submits alerts using the specified provider
func (mm *MonitoringManager) SubmitAlerts(ctx context.Context, providerName string, request *types.AlertRequest) error {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return err
	}

	// Validate request
	if err := mm.validateAlertRequest(request); err != nil {
		return fmt.Errorf("invalid alert request: %w", err)
	}

	// Submit with retry logic
	for attempt := 1; attempt <= mm.config.RetryAttempts; attempt++ {
		err = provider.SubmitAlerts(ctx, request)
		if err == nil {
			break
		}

		mm.logger.WithError(err).WithFields(logrus.Fields{
			"provider": providerName,
			"attempt":  attempt,
			"count":    len(request.Alerts),
		}).Warn("Alerts submission failed, retrying")

		if attempt < mm.config.RetryAttempts {
			time.Sleep(mm.config.RetryDelay)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to submit alerts after %d attempts: %w", mm.config.RetryAttempts, err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"count":    len(request.Alerts),
	}).Info("Alerts submitted successfully")

	return nil
}

// QueryMetrics queries metrics using the specified provider
func (mm *MonitoringManager) QueryMetrics(ctx context.Context, providerName string, request *types.QueryRequest) (*types.QueryResponse, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.QueryMetrics(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"query":    request.Query,
	}).Debug("Metrics queried successfully")

	return response, nil
}

// QueryLogs queries logs using the specified provider
func (mm *MonitoringManager) QueryLogs(ctx context.Context, providerName string, request *types.QueryRequest) (*types.QueryResponse, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.QueryLogs(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"query":    request.Query,
	}).Debug("Logs queried successfully")

	return response, nil
}

// QueryTraces queries traces using the specified provider
func (mm *MonitoringManager) QueryTraces(ctx context.Context, providerName string, request *types.QueryRequest) (*types.QueryResponse, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.QueryTraces(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to query traces: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"query":    request.Query,
	}).Debug("Traces queried successfully")

	return response, nil
}

// QueryAlerts queries alerts using the specified provider
func (mm *MonitoringManager) QueryAlerts(ctx context.Context, providerName string, request *types.QueryRequest) (*types.QueryResponse, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.QueryAlerts(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"query":    request.Query,
	}).Debug("Alerts queried successfully")

	return response, nil
}

// HealthCheck performs health check using the specified provider
func (mm *MonitoringManager) HealthCheck(ctx context.Context, providerName string, request *types.HealthCheckRequest) (*types.HealthCheckResponse, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.HealthCheck(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to perform health check: %w", err)
	}

	mm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"service":  request.Service,
		"status":   response.Status,
	}).Debug("Health check completed")

	return response, nil
}

// GetStats gets statistics from a provider
func (mm *MonitoringManager) GetStats(ctx context.Context, providerName string) (*types.MonitoringStats, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	stats, err := provider.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitoring stats: %w", err)
	}

	return stats, nil
}

// HealthCheckAll performs health check on all providers
func (mm *MonitoringManager) HealthCheckAll(ctx context.Context) map[string]*types.HealthCheckResponse {
	results := make(map[string]*types.HealthCheckResponse)

	for name, provider := range mm.providers {
		request := &types.HealthCheckRequest{
			Service: name,
			Timeout: mm.config.Timeout,
		}

		response, err := provider.HealthCheck(ctx, request)
		if err != nil {
			// Create error response
			response = &types.HealthCheckResponse{
				Service:   name,
				Status:    types.HealthStatusUnhealthy,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			}
		}

		results[name] = response
	}

	return results
}

// GetSupportedProviders returns a list of registered providers
func (mm *MonitoringManager) GetSupportedProviders() []string {
	providers := make([]string, 0, len(mm.providers))
	for name := range mm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderCapabilities returns capabilities of a provider
func (mm *MonitoringManager) GetProviderCapabilities(providerName string) ([]types.MonitoringFeature, *types.ConnectionInfo, error) {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return nil, nil, err
	}

	return provider.GetSupportedFeatures(), provider.GetConnectionInfo(), nil
}

// IsProviderConnected checks if a provider is connected
func (mm *MonitoringManager) IsProviderConnected(providerName string) bool {
	provider, err := mm.GetProvider(providerName)
	if err != nil {
		return false
	}
	return provider.IsConnected()
}

// GetConnectedProviders returns a list of connected providers
func (mm *MonitoringManager) GetConnectedProviders() []string {
	connected := make([]string, 0)
	for name, provider := range mm.providers {
		if provider.IsConnected() {
			connected = append(connected, name)
		}
	}
	return connected
}

// Close closes all monitoring connections
func (mm *MonitoringManager) Close() error {
	var lastErr error

	for name, provider := range mm.providers {
		if err := provider.Close(); err != nil {
			mm.logger.WithError(err).WithField("provider", name).Error("Failed to close monitoring provider")
			lastErr = err
		}
	}

	return lastErr
}

// validateMetricRequest validates a metric request
func (mm *MonitoringManager) validateMetricRequest(request *types.MetricRequest) error {
	if request == nil {
		return fmt.Errorf("metric request cannot be nil")
	}

	if len(request.Metrics) == 0 {
		return fmt.Errorf("metrics cannot be empty")
	}

	for i, metric := range request.Metrics {
		if metric.Name == "" {
			return fmt.Errorf("metric name is required at index %d", i)
		}
		if metric.Type == "" {
			return fmt.Errorf("metric type is required at index %d", i)
		}
	}

	return nil
}

// validateLogRequest validates a log request
func (mm *MonitoringManager) validateLogRequest(request *types.LogRequest) error {
	if request == nil {
		return fmt.Errorf("log request cannot be nil")
	}

	if len(request.Logs) == 0 {
		return fmt.Errorf("logs cannot be empty")
	}

	for i, log := range request.Logs {
		if log.Message == "" {
			return fmt.Errorf("log message is required at index %d", i)
		}
		if log.Level == "" {
			return fmt.Errorf("log level is required at index %d", i)
		}
	}

	return nil
}

// validateTraceRequest validates a trace request
func (mm *MonitoringManager) validateTraceRequest(request *types.TraceRequest) error {
	if request == nil {
		return fmt.Errorf("trace request cannot be nil")
	}

	if len(request.Traces) == 0 {
		return fmt.Errorf("traces cannot be empty")
	}

	for i, trace := range request.Traces {
		if trace.TraceID == "" {
			return fmt.Errorf("trace ID is required at index %d", i)
		}
		if trace.ServiceName == "" {
			return fmt.Errorf("service name is required at index %d", i)
		}
	}

	return nil
}

// validateAlertRequest validates an alert request
func (mm *MonitoringManager) validateAlertRequest(request *types.AlertRequest) error {
	if request == nil {
		return fmt.Errorf("alert request cannot be nil")
	}

	if len(request.Alerts) == 0 {
		return fmt.Errorf("alerts cannot be empty")
	}

	for i, alert := range request.Alerts {
		if alert.Name == "" {
			return fmt.Errorf("alert name is required at index %d", i)
		}
		if alert.Severity == "" {
			return fmt.Errorf("alert severity is required at index %d", i)
		}
	}

	return nil
}
