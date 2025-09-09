package types

import (
	"time"
)

// ServiceFeature represents a service discovery feature
type ServiceFeature string

const (
	// Basic features
	FeatureRegister   ServiceFeature = "register"
	FeatureDeregister ServiceFeature = "deregister"
	FeatureDiscover   ServiceFeature = "discover"
	FeatureHealth     ServiceFeature = "health"
	FeatureWatch      ServiceFeature = "watch"

	// Advanced features
	FeatureLoadBalancing ServiceFeature = "load_balancing"
	FeatureFailover      ServiceFeature = "failover"
	FeatureTags          ServiceFeature = "tags"
	FeatureMetadata      ServiceFeature = "metadata"
	FeatureTTL           ServiceFeature = "ttl"
	FeatureClustering    ServiceFeature = "clustering"
	FeatureConsistency   ServiceFeature = "consistency"
	FeatureSecurity      ServiceFeature = "security"
)

// ConnectionInfo holds connection information for a service discovery provider
type ConnectionInfo struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Protocol string            `json:"protocol"`
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

// ServiceInstance represents a service instance
type ServiceInstance struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Protocol    string            `json:"protocol"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	Health      HealthStatus      `json:"health"`
	Weight      int               `json:"weight"`
	TTL         time.Duration     `json:"ttl"`
	LastUpdated time.Time         `json:"last_updated"`
}

// HealthStatus represents the health status of a service
type HealthStatus string

const (
	HealthPassing  HealthStatus = "passing"
	HealthWarning  HealthStatus = "warning"
	HealthCritical HealthStatus = "critical"
	HealthUnknown  HealthStatus = "unknown"
)

// Service represents a service with multiple instances
type Service struct {
	Name      string             `json:"name"`
	Instances []*ServiceInstance `json:"instances"`
	Tags      []string           `json:"tags"`
	Metadata  map[string]string  `json:"metadata"`
}

// ServiceQuery represents a query for discovering services
type ServiceQuery struct {
	Name     string            `json:"name"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
	Health   HealthStatus      `json:"health"`
	Limit    int               `json:"limit"`
}

// ServiceRegistration represents a service registration request
type ServiceRegistration struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Protocol string            `json:"protocol"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
	Health   HealthStatus      `json:"health"`
	TTL      time.Duration     `json:"ttl"`
}

// HealthCheck represents a health check configuration
type HealthCheck struct {
	ID       string        `json:"id"`
	Service  string        `json:"service"`
	Type     HealthType    `json:"type"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Path     string        `json:"path,omitempty"`
	Port     int           `json:"port,omitempty"`
	Command  []string      `json:"command,omitempty"`
}

// HealthType represents the type of health check
type HealthType string

const (
	HealthTypeHTTP    HealthType = "http"
	HealthTypeTCP     HealthType = "tcp"
	HealthTypeCommand HealthType = "command"
	HealthTypeScript  HealthType = "script"
)

// ServiceEvent represents a service discovery event
type ServiceEvent struct {
	Type      ServiceEventType `json:"type"`
	Service   *Service         `json:"service"`
	Instance  *ServiceInstance `json:"instance,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
	Provider  string           `json:"provider"`
}

// ServiceEventType represents the type of service event
type ServiceEventType string

const (
	EventServiceRegistered   ServiceEventType = "service_registered"
	EventServiceDeregistered ServiceEventType = "service_deregistered"
	EventServiceUpdated      ServiceEventType = "service_updated"
	EventInstanceAdded       ServiceEventType = "instance_added"
	EventInstanceRemoved     ServiceEventType = "instance_removed"
	EventInstanceUpdated     ServiceEventType = "instance_updated"
	EventHealthChanged       ServiceEventType = "health_changed"
)

// ProviderInfo holds information about a service discovery provider
type ProviderInfo struct {
	Name              string           `json:"name"`
	SupportedFeatures []ServiceFeature `json:"supported_features"`
	ConnectionInfo    *ConnectionInfo  `json:"connection_info"`
	IsConnected       bool             `json:"is_connected"`
}

// DiscoveryStats represents service discovery statistics
type DiscoveryStats struct {
	ServicesRegistered int64         `json:"services_registered"`
	InstancesActive    int64         `json:"instances_active"`
	QueriesProcessed   int64         `json:"queries_processed"`
	Uptime             time.Duration `json:"uptime"`
	LastUpdate         time.Time     `json:"last_update"`
	Provider           string        `json:"provider"`
}

// DiscoveryConfig holds general discovery configuration
type DiscoveryConfig struct {
	DefaultTTL     time.Duration `json:"default_ttl"`
	HealthInterval time.Duration `json:"health_interval"`
	WatchTimeout   time.Duration `json:"watch_timeout"`
	RetryAttempts  int           `json:"retry_attempts"`
	RetryDelay     time.Duration `json:"retry_delay"`
	Namespace      string        `json:"namespace"`
	ServicePrefix  string        `json:"service_prefix"`
	InstancePrefix string        `json:"instance_prefix"`
}

// DiscoveryError represents a discovery-specific error
type DiscoveryError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Service string `json:"service,omitempty"`
}

func (e *DiscoveryError) Error() string {
	return e.Message
}

// Common discovery error codes
const (
	ErrCodeServiceNotFound      = "SERVICE_NOT_FOUND"
	ErrCodeInstanceNotFound     = "INSTANCE_NOT_FOUND"
	ErrCodeServiceExists        = "SERVICE_EXISTS"
	ErrCodeInstanceExists       = "INSTANCE_EXISTS"
	ErrCodeInvalidService       = "INVALID_SERVICE"
	ErrCodeInvalidInstance      = "INVALID_INSTANCE"
	ErrCodeConnection           = "CONNECTION_ERROR"
	ErrCodeTimeout              = "TIMEOUT"
	ErrCodeHealthCheckFailed    = "HEALTH_CHECK_FAILED"
	ErrCodeRegistrationFailed   = "REGISTRATION_FAILED"
	ErrCodeDeregistrationFailed = "DEREGISTRATION_FAILED"
	ErrCodeWatchFailed          = "WATCH_FAILED"
)

// LoadBalancer represents a load balancing strategy
type LoadBalancer interface {
	SelectInstance(instances []*ServiceInstance) (*ServiceInstance, error)
}

// LoadBalancerType represents the type of load balancer
type LoadBalancerType string

const (
	LoadBalancerRoundRobin LoadBalancerType = "round_robin"
	LoadBalancerRandom     LoadBalancerType = "random"
	LoadBalancerWeighted   LoadBalancerType = "weighted"
	LoadBalancerLeastConn  LoadBalancerType = "least_connections"
	LoadBalancerHash       LoadBalancerType = "hash"
)

// WatchOptions represents options for watching service changes
type WatchOptions struct {
	ServiceName string            `json:"service_name"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	HealthOnly  bool              `json:"health_only"`
	Timeout     time.Duration     `json:"timeout"`
}
