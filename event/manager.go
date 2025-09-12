package event

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/event/types"
	"github.com/sirupsen/logrus"
)

// EventSourcingManager manages multiple event sourcing providers
type EventSourcingManager struct {
	providers map[string]types.EventSourcingProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds event sourcing manager configuration
type ManagerConfig struct {
	DefaultProvider   string            `json:"default_provider"`
	RetryAttempts     int               `json:"retry_attempts"`
	RetryDelay        time.Duration     `json:"retry_delay"`
	Timeout           time.Duration     `json:"timeout"`
	MaxEventSize      int64             `json:"max_event_size"`
	MaxBatchSize      int               `json:"max_batch_size"`
	SnapshotThreshold int64             `json:"snapshot_threshold"`
	RetentionPeriod   time.Duration     `json:"retention_period"`
	Compression       bool              `json:"compression"`
	Encryption        bool              `json:"encryption"`
	Metadata          map[string]string `json:"metadata"`
}

// NewEventSourcingManager creates a new event sourcing manager
func NewEventSourcingManager(config *ManagerConfig, logger *logrus.Logger) *EventSourcingManager {
	if config == nil {
		config = &ManagerConfig{
			DefaultProvider:   "postgresql",
			RetryAttempts:     3,
			RetryDelay:        time.Second,
			Timeout:           30 * time.Second,
			MaxEventSize:      1024 * 1024, // 1MB
			MaxBatchSize:      100,
			SnapshotThreshold: 1000,
			RetentionPeriod:   365 * 24 * time.Hour, // 1 year
			Compression:       false,
			Encryption:        false,
			Metadata:          make(map[string]string),
		}
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &EventSourcingManager{
		providers: make(map[string]types.EventSourcingProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers an event sourcing provider
func (esm *EventSourcingManager) RegisterProvider(provider types.EventSourcingProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := esm.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	esm.providers[name] = provider
	esm.logger.WithField("provider", name).Info("Event sourcing provider registered")

	return nil
}

// GetProvider returns a specific provider by name
func (esm *EventSourcingManager) GetProvider(name string) (types.EventSourcingProvider, error) {
	provider, exists := esm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default provider
func (esm *EventSourcingManager) GetDefaultProvider() (types.EventSourcingProvider, error) {
	if esm.config.DefaultProvider == "" {
		// Return first available provider
		for _, provider := range esm.providers {
			return provider, nil
		}
		return nil, fmt.Errorf("no providers registered")
	}

	return esm.GetProvider(esm.config.DefaultProvider)
}

// AppendEvent appends an event to a stream using the default provider
func (esm *EventSourcingManager) AppendEvent(ctx context.Context, request *types.AppendEventRequest) (*types.AppendEventResponse, error) {
	return esm.AppendEventWithProvider(ctx, "", request)
}

// AppendEventWithProvider appends an event to a stream using a specific provider
func (esm *EventSourcingManager) AppendEventWithProvider(ctx context.Context, providerName string, request *types.AppendEventRequest) (*types.AppendEventResponse, error) {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Validate request
	if err := esm.validateAppendEventRequest(request); err != nil {
		return nil, fmt.Errorf("invalid append event request: %w", err)
	}

	// Check event size limit
	if esm.getEventSize(request) > esm.config.MaxEventSize {
		return nil, fmt.Errorf("event size %d exceeds maximum allowed size %d", esm.getEventSize(request), esm.config.MaxEventSize)
	}

	result, err := esm.executeWithRetry(ctx, func() (interface{}, error) {
		return provider.AppendEvent(ctx, request)
	})
	if err != nil {
		return nil, err
	}
	return result.(*types.AppendEventResponse), nil
}

// GetEvents retrieves events using the default provider
func (esm *EventSourcingManager) GetEvents(ctx context.Context, request *types.GetEventsRequest) (*types.GetEventsResponse, error) {
	return esm.GetEventsWithProvider(ctx, "", request)
}

// GetEventsWithProvider retrieves events using a specific provider
func (esm *EventSourcingManager) GetEventsWithProvider(ctx context.Context, providerName string, request *types.GetEventsRequest) (*types.GetEventsResponse, error) {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	result, err := esm.executeWithRetry(ctx, func() (interface{}, error) {
		return provider.GetEvents(ctx, request)
	})
	if err != nil {
		return nil, err
	}
	return result.(*types.GetEventsResponse), nil
}

// GetEventByID retrieves an event by ID using the default provider
func (esm *EventSourcingManager) GetEventByID(ctx context.Context, request *types.GetEventByIDRequest) (*types.Event, error) {
	return esm.GetEventByIDWithProvider(ctx, "", request)
}

// GetEventByIDWithProvider retrieves an event by ID using a specific provider
func (esm *EventSourcingManager) GetEventByIDWithProvider(ctx context.Context, providerName string, request *types.GetEventByIDRequest) (*types.Event, error) {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	result, err := esm.executeWithRetry(ctx, func() (interface{}, error) {
		return provider.GetEventByID(ctx, request)
	})
	if err != nil {
		return nil, err
	}
	return result.(*types.Event), nil
}

// GetEventsByStream retrieves events by stream using the default provider
func (esm *EventSourcingManager) GetEventsByStream(ctx context.Context, request *types.GetEventsByStreamRequest) (*types.GetEventsResponse, error) {
	return esm.GetEventsByStreamWithProvider(ctx, "", request)
}

// GetEventsByStreamWithProvider retrieves events by stream using a specific provider
func (esm *EventSourcingManager) GetEventsByStreamWithProvider(ctx context.Context, providerName string, request *types.GetEventsByStreamRequest) (*types.GetEventsResponse, error) {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	result, err := esm.executeWithRetry(ctx, func() (interface{}, error) {
		return provider.GetEventsByStream(ctx, request)
	})
	if err != nil {
		return nil, err
	}
	return result.(*types.GetEventsResponse), nil
}

// CreateStream creates a stream using the default provider
func (esm *EventSourcingManager) CreateStream(ctx context.Context, request *types.CreateStreamRequest) error {
	return esm.CreateStreamWithProvider(ctx, "", request)
}

// CreateStreamWithProvider creates a stream using a specific provider
func (esm *EventSourcingManager) CreateStreamWithProvider(ctx context.Context, providerName string, request *types.CreateStreamRequest) error {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return err
	}

	// Validate request
	if err := esm.validateCreateStreamRequest(request); err != nil {
		return fmt.Errorf("invalid create stream request: %w", err)
	}

	return esm.executeWithRetryError(ctx, func() error {
		return provider.CreateStream(ctx, request)
	})
}

// DeleteStream deletes a stream using the default provider
func (esm *EventSourcingManager) DeleteStream(ctx context.Context, request *types.DeleteStreamRequest) error {
	return esm.DeleteStreamWithProvider(ctx, "", request)
}

// DeleteStreamWithProvider deletes a stream using a specific provider
func (esm *EventSourcingManager) DeleteStreamWithProvider(ctx context.Context, providerName string, request *types.DeleteStreamRequest) error {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return err
	}

	return esm.executeWithRetryError(ctx, func() error {
		return provider.DeleteStream(ctx, request)
	})
}

// StreamExists checks if a stream exists using the default provider
func (esm *EventSourcingManager) StreamExists(ctx context.Context, request *types.StreamExistsRequest) (bool, error) {
	return esm.StreamExistsWithProvider(ctx, "", request)
}

// StreamExistsWithProvider checks if a stream exists using a specific provider
func (esm *EventSourcingManager) StreamExistsWithProvider(ctx context.Context, providerName string, request *types.StreamExistsRequest) (bool, error) {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return false, err
	}

	var result bool
	err = esm.executeWithRetryError(ctx, func() error {
		var err error
		result, err = provider.StreamExists(ctx, request)
		return err
	})

	return result, err
}

// ListStreams lists all streams using the default provider
func (esm *EventSourcingManager) ListStreams(ctx context.Context) ([]types.StreamInfo, error) {
	return esm.ListStreamsWithProvider(ctx, "")
}

// ListStreamsWithProvider lists all streams using a specific provider
func (esm *EventSourcingManager) ListStreamsWithProvider(ctx context.Context, providerName string) ([]types.StreamInfo, error) {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result []types.StreamInfo
	err = esm.executeWithRetryError(ctx, func() error {
		var err error
		result, err = provider.ListStreams(ctx)
		return err
	})

	return result, err
}

// CreateSnapshot creates a snapshot using the default provider
func (esm *EventSourcingManager) CreateSnapshot(ctx context.Context, request *types.CreateSnapshotRequest) (*types.CreateSnapshotResponse, error) {
	return esm.CreateSnapshotWithProvider(ctx, "", request)
}

// CreateSnapshotWithProvider creates a snapshot using a specific provider
func (esm *EventSourcingManager) CreateSnapshotWithProvider(ctx context.Context, providerName string, request *types.CreateSnapshotRequest) (*types.CreateSnapshotResponse, error) {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	result, err := esm.executeWithRetry(ctx, func() (interface{}, error) {
		return provider.CreateSnapshot(ctx, request)
	})
	if err != nil {
		return nil, err
	}
	return result.(*types.CreateSnapshotResponse), nil
}

// GetSnapshot retrieves a snapshot using the default provider
func (esm *EventSourcingManager) GetSnapshot(ctx context.Context, request *types.GetSnapshotRequest) (*types.Snapshot, error) {
	return esm.GetSnapshotWithProvider(ctx, "", request)
}

// GetSnapshotWithProvider retrieves a snapshot using a specific provider
func (esm *EventSourcingManager) GetSnapshotWithProvider(ctx context.Context, providerName string, request *types.GetSnapshotRequest) (*types.Snapshot, error) {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	result, err := esm.executeWithRetry(ctx, func() (interface{}, error) {
		return provider.GetSnapshot(ctx, request)
	})
	if err != nil {
		return nil, err
	}
	return result.(*types.Snapshot), nil
}

// DeleteSnapshot deletes a snapshot using the default provider
func (esm *EventSourcingManager) DeleteSnapshot(ctx context.Context, request *types.DeleteSnapshotRequest) error {
	return esm.DeleteSnapshotWithProvider(ctx, "", request)
}

// DeleteSnapshotWithProvider deletes a snapshot using a specific provider
func (esm *EventSourcingManager) DeleteSnapshotWithProvider(ctx context.Context, providerName string, request *types.DeleteSnapshotRequest) error {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return err
	}

	return esm.executeWithRetryError(ctx, func() error {
		return provider.DeleteSnapshot(ctx, request)
	})
}

// AppendEventsBatch appends multiple events using the default provider
func (esm *EventSourcingManager) AppendEventsBatch(ctx context.Context, request *types.AppendEventsBatchRequest) (*types.AppendEventsBatchResponse, error) {
	return esm.AppendEventsBatchWithProvider(ctx, "", request)
}

// AppendEventsBatchWithProvider appends multiple events using a specific provider
func (esm *EventSourcingManager) AppendEventsBatchWithProvider(ctx context.Context, providerName string, request *types.AppendEventsBatchRequest) (*types.AppendEventsBatchResponse, error) {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Validate request
	if err := esm.validateAppendEventsBatchRequest(request); err != nil {
		return nil, fmt.Errorf("invalid append events batch request: %w", err)
	}

	result, err := esm.executeWithRetry(ctx, func() (interface{}, error) {
		return provider.AppendEventsBatch(ctx, request)
	})
	if err != nil {
		return nil, err
	}
	return result.(*types.AppendEventsBatchResponse), nil
}

// GetStreamInfo gets stream information using the default provider
func (esm *EventSourcingManager) GetStreamInfo(ctx context.Context, request *types.GetStreamInfoRequest) (*types.StreamInfo, error) {
	return esm.GetStreamInfoWithProvider(ctx, "", request)
}

// GetStreamInfoWithProvider gets stream information using a specific provider
func (esm *EventSourcingManager) GetStreamInfoWithProvider(ctx context.Context, providerName string, request *types.GetStreamInfoRequest) (*types.StreamInfo, error) {
	provider, err := esm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	result, err := esm.executeWithRetry(ctx, func() (interface{}, error) {
		return provider.GetStreamInfo(ctx, request)
	})
	if err != nil {
		return nil, err
	}
	return result.(*types.StreamInfo), nil
}

// GetStats returns event sourcing statistics from all providers
func (esm *EventSourcingManager) GetStats(ctx context.Context) (map[string]*types.EventSourcingStats, error) {
	stats := make(map[string]*types.EventSourcingStats)

	for name, provider := range esm.providers {
		stat, err := provider.GetStats(ctx)
		if err != nil {
			esm.logger.WithError(err).WithField("provider", name).Warn("Failed to get stats from provider")
			continue
		}
		stats[name] = stat
	}

	return stats, nil
}

// HealthCheck performs health check on all providers
func (esm *EventSourcingManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for name, provider := range esm.providers {
		err := provider.HealthCheck(ctx)
		results[name] = err
	}

	return results
}

// Close closes all providers
func (esm *EventSourcingManager) Close() error {
	var lastErr error

	for name, provider := range esm.providers {
		if err := provider.Close(); err != nil {
			esm.logger.WithError(err).WithField("provider", name).Error("Failed to close provider")
			lastErr = err
		}
	}

	return lastErr
}

// ListProviders returns a list of registered provider names
func (esm *EventSourcingManager) ListProviders() []string {
	providers := make([]string, 0, len(esm.providers))
	for name := range esm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderInfo returns information about all registered providers
func (esm *EventSourcingManager) GetProviderInfo() map[string]*types.ProviderInfo {
	info := make(map[string]*types.ProviderInfo)

	for name, provider := range esm.providers {
		info[name] = &types.ProviderInfo{
			Name:              provider.GetName(),
			SupportedFeatures: provider.GetSupportedFeatures(),
			ConnectionInfo:    provider.GetConnectionInfo(),
			IsConnected:       provider.IsConnected(),
		}
	}

	return info
}

// getProvider returns a provider by name or the default provider
func (esm *EventSourcingManager) getProvider(name string) (types.EventSourcingProvider, error) {
	if name == "" {
		return esm.GetDefaultProvider()
	}
	return esm.GetProvider(name)
}

// executeWithRetry executes a function with retry logic
func (esm *EventSourcingManager) executeWithRetry(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error

	for attempt := 0; attempt <= esm.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(esm.config.RetryDelay * time.Duration(attempt)):
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		esm.logger.WithError(err).WithField("attempt", attempt+1).Debug("Operation failed, retrying")
	}

	return nil, fmt.Errorf("operation failed after %d attempts: %w", esm.config.RetryAttempts+1, lastErr)
}

// executeWithRetryError executes a function that returns only an error with retry logic
func (esm *EventSourcingManager) executeWithRetryError(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= esm.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(esm.config.RetryDelay * time.Duration(attempt)):
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		esm.logger.WithError(err).WithField("attempt", attempt+1).Debug("Operation failed, retrying")
	}

	return fmt.Errorf("operation failed after %d attempts: %w", esm.config.RetryAttempts+1, lastErr)
}

// validateAppendEventRequest validates an append event request
func (esm *EventSourcingManager) validateAppendEventRequest(request *types.AppendEventRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.StreamID == "" {
		return fmt.Errorf("stream_id is required")
	}

	if request.EventType == "" {
		return fmt.Errorf("event_type is required")
	}

	if request.EventData == nil {
		return fmt.Errorf("event_data is required")
	}

	return nil
}

// validateCreateStreamRequest validates a create stream request
func (esm *EventSourcingManager) validateCreateStreamRequest(request *types.CreateStreamRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.StreamID == "" {
		return fmt.Errorf("stream_id is required")
	}

	if request.Name == "" {
		return fmt.Errorf("name is required")
	}

	if request.AggregateID == "" {
		return fmt.Errorf("aggregate_id is required")
	}

	if request.AggregateType == "" {
		return fmt.Errorf("aggregate_type is required")
	}

	return nil
}

// validateAppendEventsBatchRequest validates a batch append events request
func (esm *EventSourcingManager) validateAppendEventsBatchRequest(request *types.AppendEventsBatchRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.StreamID == "" {
		return fmt.Errorf("stream_id is required")
	}

	if len(request.Events) == 0 {
		return fmt.Errorf("events cannot be empty")
	}

	if len(request.Events) > esm.config.MaxBatchSize {
		return fmt.Errorf("batch size %d exceeds maximum allowed size %d", len(request.Events), esm.config.MaxBatchSize)
	}

	// Validate each event in the batch
	for i, event := range request.Events {
		if event == nil {
			return fmt.Errorf("event at index %d cannot be nil", i)
		}

		if event.EventType == "" {
			return fmt.Errorf("event at index %d: event_type is required", i)
		}

		if event.EventData == nil {
			return fmt.Errorf("event at index %d: event_data is required", i)
		}
	}

	return nil
}

// getEventSize calculates the approximate size of an event
func (esm *EventSourcingManager) getEventSize(request *types.AppendEventRequest) int64 {
	size := int64(0)

	// Add basic fields
	size += int64(len(request.StreamID))
	size += int64(len(request.EventType))
	size += int64(len(request.CorrelationID))
	size += int64(len(request.CausationID))
	size += int64(len(request.UserID))
	size += int64(len(request.TenantID))
	size += int64(len(request.AggregateID))
	size += int64(len(request.AggregateType))

	// Add event data size (approximate)
	if request.EventData != nil {
		for key, value := range request.EventData {
			size += int64(len(key))
			if str, ok := value.(string); ok {
				size += int64(len(str))
			} else {
				size += 64 // Approximate size for non-string values
			}
		}
	}

	// Add event metadata size (approximate)
	if request.EventMetadata != nil {
		for key, value := range request.EventMetadata {
			size += int64(len(key))
			if str, ok := value.(string); ok {
				size += int64(len(str))
			} else {
				size += 64 // Approximate size for non-string values
			}
		}
	}

	// Add options size (approximate)
	if request.Options != nil {
		for key, value := range request.Options {
			size += int64(len(key))
			if str, ok := value.(string); ok {
				size += int64(len(str))
			} else {
				size += 64 // Approximate size for non-string values
			}
		}
	}

	return size
}

// GetConfig returns the manager configuration
func (esm *EventSourcingManager) GetConfig() *ManagerConfig {
	return esm.config
}

// UpdateConfig updates the manager configuration
func (esm *EventSourcingManager) UpdateConfig(config *ManagerConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	esm.config = config
	esm.logger.Info("Event sourcing manager configuration updated")

	return nil
}

// GetLogger returns the manager logger
func (esm *EventSourcingManager) GetLogger() *logrus.Logger {
	return esm.logger
}

// SetLogger sets the manager logger
func (esm *EventSourcingManager) SetLogger(logger *logrus.Logger) {
	if logger != nil {
		esm.logger = logger
	}
}
