package types

import (
	"context"
	"time"
)

// EventSourcingProvider represents an event sourcing provider interface
type EventSourcingProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []EventSourcingFeature
	GetConnectionInfo() *ConnectionInfo

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Event operations
	AppendEvent(ctx context.Context, request *AppendEventRequest) (*AppendEventResponse, error)
	GetEvents(ctx context.Context, request *GetEventsRequest) (*GetEventsResponse, error)
	GetEventByID(ctx context.Context, request *GetEventByIDRequest) (*Event, error)
	GetEventsByStream(ctx context.Context, request *GetEventsByStreamRequest) (*GetEventsResponse, error)

	// Stream management
	CreateStream(ctx context.Context, request *CreateStreamRequest) error
	DeleteStream(ctx context.Context, request *DeleteStreamRequest) error
	StreamExists(ctx context.Context, request *StreamExistsRequest) (bool, error)
	ListStreams(ctx context.Context) ([]StreamInfo, error)

	// Snapshot operations
	CreateSnapshot(ctx context.Context, request *CreateSnapshotRequest) (*CreateSnapshotResponse, error)
	GetSnapshot(ctx context.Context, request *GetSnapshotRequest) (*Snapshot, error)
	DeleteSnapshot(ctx context.Context, request *DeleteSnapshotRequest) error

	// Advanced operations
	AppendEventsBatch(ctx context.Context, request *AppendEventsBatchRequest) (*AppendEventsBatchResponse, error)
	GetStreamInfo(ctx context.Context, request *GetStreamInfoRequest) (*StreamInfo, error)
	GetStats(ctx context.Context) (*EventSourcingStats, error)

	// Health and monitoring
	HealthCheck(ctx context.Context) error

	// Configuration
	Configure(config map[string]interface{}) error
	IsConfigured() bool
	Close() error
}

// EventSourcingFeature represents an event sourcing feature
type EventSourcingFeature string

const (
	// Basic features
	FeatureEventAppend      EventSourcingFeature = "event_append"
	FeatureEventRetrieval   EventSourcingFeature = "event_retrieval"
	FeatureStreamManagement EventSourcingFeature = "stream_management"
	FeatureEventQuery       EventSourcingFeature = "event_query"
	FeatureEventFiltering   EventSourcingFeature = "event_filtering"

	// Advanced features
	FeatureSnapshots          EventSourcingFeature = "snapshots"
	FeatureEventReplay        EventSourcingFeature = "event_replay"
	FeatureEventProjection    EventSourcingFeature = "event_projection"
	FeatureEventVersioning    EventSourcingFeature = "event_versioning"
	FeatureEventCompression   EventSourcingFeature = "event_compression"
	FeatureEventEncryption    EventSourcingFeature = "event_encryption"
	FeatureEventBatching      EventSourcingFeature = "event_batching"
	FeatureEventPartitioning  EventSourcingFeature = "event_partitioning"
	FeatureEventRetention     EventSourcingFeature = "event_retention"
	FeatureEventOrdering      EventSourcingFeature = "event_ordering"
	FeatureEventDeduplication EventSourcingFeature = "event_deduplication"
	FeatureEventCorrelation   EventSourcingFeature = "event_correlation"
	FeatureEventAggregation   EventSourcingFeature = "event_aggregation"
	FeatureEventStreaming     EventSourcingFeature = "event_streaming"
	FeatureEventReplication   EventSourcingFeature = "event_replication"
	FeatureEventClustering    EventSourcingFeature = "event_clustering"
)

// ConnectionInfo holds connection information for an event sourcing provider
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

// Event represents a domain event in event sourcing
type Event struct {
	ID            string                 `json:"id"`
	StreamID      string                 `json:"stream_id"`
	EventType     string                 `json:"event_type"`
	EventData     map[string]interface{} `json:"event_data"`
	EventMetadata map[string]interface{} `json:"event_metadata,omitempty"`
	Version       int64                  `json:"version"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	CausationID   string                 `json:"causation_id,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	TenantID      string                 `json:"tenant_id,omitempty"`
	AggregateID   string                 `json:"aggregate_id,omitempty"`
	AggregateType string                 `json:"aggregate_type,omitempty"`
	ProviderData  map[string]interface{} `json:"provider_data,omitempty"`
}

// Stream represents an event stream
type Stream struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int64                  `json:"version"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ProviderData  map[string]interface{} `json:"provider_data,omitempty"`
}

// Snapshot represents a snapshot of an aggregate state
type Snapshot struct {
	ID            string                 `json:"id"`
	StreamID      string                 `json:"stream_id"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int64                  `json:"version"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	ProviderData  map[string]interface{} `json:"provider_data,omitempty"`
}

// StreamInfo represents stream information
type StreamInfo struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int64                  `json:"version"`
	EventCount    int64                  `json:"event_count"`
	Size          int64                  `json:"size"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	LastEventAt   *time.Time             `json:"last_event_at,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ProviderData  map[string]interface{} `json:"provider_data,omitempty"`
}

// EventSourcingConfig holds event sourcing configuration
type EventSourcingConfig struct {
	DefaultProvider   string                 `json:"default_provider"`
	RetryAttempts     int                    `json:"retry_attempts"`
	RetryDelay        time.Duration          `json:"retry_delay"`
	Timeout           time.Duration          `json:"timeout"`
	MaxEventSize      int64                  `json:"max_event_size"`
	MaxBatchSize      int                    `json:"max_batch_size"`
	SnapshotThreshold int64                  `json:"snapshot_threshold"`
	RetentionPeriod   time.Duration          `json:"retention_period"`
	Compression       bool                   `json:"compression"`
	Encryption        bool                   `json:"encryption"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// EventSourcingStats represents event sourcing statistics
type EventSourcingStats struct {
	TotalEvents    int64                  `json:"total_events"`
	TotalStreams   int64                  `json:"total_streams"`
	TotalSnapshots int64                  `json:"total_snapshots"`
	EventsByType   map[string]int64       `json:"events_by_type"`
	EventsByStream map[string]int64       `json:"events_by_stream"`
	StorageSize    int64                  `json:"storage_size"`
	Uptime         time.Duration          `json:"uptime"`
	LastUpdate     time.Time              `json:"last_update"`
	Provider       string                 `json:"provider"`
	ProviderData   map[string]interface{} `json:"provider_data,omitempty"`
}

// ProviderInfo holds information about an event sourcing provider
type ProviderInfo struct {
	Name              string                 `json:"name"`
	SupportedFeatures []EventSourcingFeature `json:"supported_features"`
	ConnectionInfo    *ConnectionInfo        `json:"connection_info"`
	IsConnected       bool                   `json:"is_connected"`
}

// Request/Response types

// AppendEventRequest represents an append event request
type AppendEventRequest struct {
	StreamID        string                 `json:"stream_id"`
	EventType       string                 `json:"event_type"`
	EventData       map[string]interface{} `json:"event_data"`
	EventMetadata   map[string]interface{} `json:"event_metadata,omitempty"`
	ExpectedVersion *int64                 `json:"expected_version,omitempty"`
	CorrelationID   string                 `json:"correlation_id,omitempty"`
	CausationID     string                 `json:"causation_id,omitempty"`
	UserID          string                 `json:"user_id,omitempty"`
	TenantID        string                 `json:"tenant_id,omitempty"`
	AggregateID     string                 `json:"aggregate_id,omitempty"`
	AggregateType   string                 `json:"aggregate_type,omitempty"`
	Options         map[string]interface{} `json:"options,omitempty"`
}

// AppendEventResponse represents an append event response
type AppendEventResponse struct {
	EventID      string                 `json:"event_id"`
	StreamID     string                 `json:"stream_id"`
	Version      int64                  `json:"version"`
	Timestamp    time.Time              `json:"timestamp"`
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
}

// GetEventsRequest represents a get events request
type GetEventsRequest struct {
	StreamID     string                 `json:"stream_id,omitempty"`
	EventTypes   []string               `json:"event_types,omitempty"`
	StartVersion *int64                 `json:"start_version,omitempty"`
	EndVersion   *int64                 `json:"end_version,omitempty"`
	StartTime    *time.Time             `json:"start_time,omitempty"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Offset       int                    `json:"offset,omitempty"`
	SortOrder    string                 `json:"sort_order,omitempty"`
	Filter       map[string]interface{} `json:"filter,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// GetEventsResponse represents a get events response
type GetEventsResponse struct {
	Events       []*Event               `json:"events"`
	TotalCount   int64                  `json:"total_count"`
	HasMore      bool                   `json:"has_more"`
	NextOffset   int                    `json:"next_offset,omitempty"`
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
}

// GetEventByIDRequest represents a get event by ID request
type GetEventByIDRequest struct {
	EventID string                 `json:"event_id"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// GetEventsByStreamRequest represents a get events by stream request
type GetEventsByStreamRequest struct {
	StreamID     string                 `json:"stream_id"`
	StartVersion *int64                 `json:"start_version,omitempty"`
	EndVersion   *int64                 `json:"end_version,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// CreateStreamRequest represents a create stream request
type CreateStreamRequest struct {
	StreamID      string                 `json:"stream_id"`
	Name          string                 `json:"name"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Options       map[string]interface{} `json:"options,omitempty"`
}

// DeleteStreamRequest represents a delete stream request
type DeleteStreamRequest struct {
	StreamID string                 `json:"stream_id"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// StreamExistsRequest represents a stream exists request
type StreamExistsRequest struct {
	StreamID string                 `json:"stream_id"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// CreateSnapshotRequest represents a create snapshot request
type CreateSnapshotRequest struct {
	StreamID      string                 `json:"stream_id"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int64                  `json:"version"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Options       map[string]interface{} `json:"options,omitempty"`
}

// CreateSnapshotResponse represents a create snapshot response
type CreateSnapshotResponse struct {
	SnapshotID   string                 `json:"snapshot_id"`
	StreamID     string                 `json:"stream_id"`
	Version      int64                  `json:"version"`
	Timestamp    time.Time              `json:"timestamp"`
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
}

// GetSnapshotRequest represents a get snapshot request
type GetSnapshotRequest struct {
	StreamID string                 `json:"stream_id"`
	Version  *int64                 `json:"version,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// DeleteSnapshotRequest represents a delete snapshot request
type DeleteSnapshotRequest struct {
	SnapshotID string                 `json:"snapshot_id"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// AppendEventsBatchRequest represents a batch append events request
type AppendEventsBatchRequest struct {
	StreamID        string                 `json:"stream_id"`
	Events          []*Event               `json:"events"`
	ExpectedVersion *int64                 `json:"expected_version,omitempty"`
	Options         map[string]interface{} `json:"options,omitempty"`
}

// AppendEventsBatchResponse represents a batch append events response
type AppendEventsBatchResponse struct {
	AppendedCount int                    `json:"appended_count"`
	FailedCount   int                    `json:"failed_count"`
	FailedEvents  []*Event               `json:"failed_events,omitempty"`
	StartVersion  int64                  `json:"start_version"`
	EndVersion    int64                  `json:"end_version"`
	ProviderData  map[string]interface{} `json:"provider_data,omitempty"`
}

// GetStreamInfoRequest represents a get stream info request
type GetStreamInfoRequest struct {
	StreamID string                 `json:"stream_id"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// EventSourcingError represents an event sourcing-specific error
type EventSourcingError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	StreamID string `json:"stream_id,omitempty"`
	EventID  string `json:"event_id,omitempty"`
}

func (e *EventSourcingError) Error() string {
	return e.Message
}

// Common event sourcing error codes
const (
	ErrCodeStreamNotFound         = "STREAM_NOT_FOUND"
	ErrCodeEventNotFound          = "EVENT_NOT_FOUND"
	ErrCodeSnapshotNotFound       = "SNAPSHOT_NOT_FOUND"
	ErrCodeInvalidStreamID        = "INVALID_STREAM_ID"
	ErrCodeInvalidEventType       = "INVALID_EVENT_TYPE"
	ErrCodeInvalidEventData       = "INVALID_EVENT_DATA"
	ErrCodeVersionMismatch        = "VERSION_MISMATCH"
	ErrCodeInvalidVersion         = "INVALID_VERSION"
	ErrCodeInvalidStreamName      = "INVALID_STREAM_NAME"
	ErrCodeInvalidStreamType      = "INVALID_STREAM_TYPE"
	ErrCodeInvalidStreamData      = "INVALID_STREAM_DATA"
	ErrCodeInvalidStreamMetadata  = "INVALID_STREAM_METADATA"
	ErrCodeInvalidStreamOptions   = "INVALID_STREAM_OPTIONS"
	ErrCodeInvalidStreamQuery     = "INVALID_STREAM_QUERY"
	ErrCodeInvalidStreamFilter    = "INVALID_STREAM_FILTER"
	ErrCodeInvalidStreamSort      = "INVALID_STREAM_SORT"
	ErrCodeInvalidStreamLimit     = "INVALID_STREAM_LIMIT"
	ErrCodeInvalidStreamOffset    = "INVALID_STREAM_OFFSET"
	ErrCodeInvalidStreamSortOrder = "INVALID_STREAM_SORT_ORDER"
)
