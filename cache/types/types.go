package types

import (
	"time"
)

// CacheFeature represents a cache feature
type CacheFeature string

const (
	// Basic features
	FeatureSet    CacheFeature = "set"
	FeatureGet    CacheFeature = "get"
	FeatureDelete CacheFeature = "delete"
	FeatureExists CacheFeature = "exists"
	FeatureFlush  CacheFeature = "flush"
	FeatureStats  CacheFeature = "stats"

	// Advanced features
	FeatureTags        CacheFeature = "tags"
	FeatureTTL         CacheFeature = "ttl"
	FeatureBatch       CacheFeature = "batch"
	FeaturePattern     CacheFeature = "pattern"
	FeaturePersistence CacheFeature = "persistence"
	FeatureClustering  CacheFeature = "clustering"
	FeaturePubSub      CacheFeature = "pubsub"
	FeatureLuaScripts  CacheFeature = "lua_scripts"
)

// ConnectionInfo holds connection information for a cache provider
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

// CacheStats represents cache statistics
type CacheStats struct {
	Hits       int64         `json:"hits"`
	Misses     int64         `json:"misses"`
	Keys       int64         `json:"keys"`
	Memory     int64         `json:"memory"`
	Uptime     time.Duration `json:"uptime"`
	LastUpdate time.Time     `json:"last_update"`
	Provider   string        `json:"provider"`
}

// ProviderInfo holds information about a cache provider
type ProviderInfo struct {
	Name              string          `json:"name"`
	SupportedFeatures []CacheFeature  `json:"supported_features"`
	ConnectionInfo    *ConnectionInfo `json:"connection_info"`
	IsConnected       bool            `json:"is_connected"`
}

// CacheItem represents a cached item with metadata
type CacheItem struct {
	Key        string                 `json:"key"`
	Value      interface{}            `json:"value"`
	Expiration time.Time              `json:"expiration"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// CacheConfig holds general cache configuration
type CacheConfig struct {
	DefaultTTL    time.Duration `json:"default_ttl"`
	MaxKeyLength  int           `json:"max_key_length"`
	MaxValueSize  int64         `json:"max_value_size"`
	Compression   bool          `json:"compression"`
	Serialization string        `json:"serialization"` // json, msgpack, protobuf
	KeyPrefix     string        `json:"key_prefix"`
	Namespace     string        `json:"namespace"`
}

// BatchOperation represents a batch cache operation
type BatchOperation struct {
	Type  BatchOperationType `json:"type"`
	Key   string             `json:"key"`
	Value interface{}        `json:"value,omitempty"`
	TTL   time.Duration      `json:"ttl,omitempty"`
	Tags  []string           `json:"tags,omitempty"`
}

// BatchOperationType represents the type of batch operation
type BatchOperationType string

const (
	BatchSet    BatchOperationType = "set"
	BatchGet    BatchOperationType = "get"
	BatchDelete BatchOperationType = "delete"
)

// CacheEvent represents a cache event
type CacheEvent struct {
	Type      CacheEventType `json:"type"`
	Key       string         `json:"key"`
	Value     interface{}    `json:"value,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Provider  string         `json:"provider"`
}

// CacheEventType represents the type of cache event
type CacheEventType string

const (
	EventSet    CacheEventType = "set"
	EventGet    CacheEventType = "get"
	EventDelete CacheEventType = "delete"
	EventExpire CacheEventType = "expire"
	EventFlush  CacheEventType = "flush"
)

// CacheError represents a cache-specific error
type CacheError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Key     string `json:"key,omitempty"`
}

func (e *CacheError) Error() string {
	return e.Message
}

// Common cache error codes
const (
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeInvalidKey      = "INVALID_KEY"
	ErrCodeInvalidValue    = "INVALID_VALUE"
	ErrCodeExpired         = "EXPIRED"
	ErrCodeConnection      = "CONNECTION_ERROR"
	ErrCodeTimeout         = "TIMEOUT"
	ErrCodeQuota           = "QUOTA_EXCEEDED"
	ErrCodeSerialization   = "SERIALIZATION_ERROR"
	ErrCodeDeserialization = "DESERIALIZATION_ERROR"
)
