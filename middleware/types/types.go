package types

import (
	"context"
	"net/http"
	"time"
)

// MiddlewareFeature represents a middleware feature
type MiddlewareFeature string

const (
	// Authentication features
	FeatureJWT           MiddlewareFeature = "jwt"
	FeatureOAuth2        MiddlewareFeature = "oauth2"
	FeatureBasicAuth     MiddlewareFeature = "basic_auth"
	FeatureAPIKey        MiddlewareFeature = "api_key"
	FeatureSessionAuth   MiddlewareFeature = "session_auth"
	FeatureTwoFactor     MiddlewareFeature = "two_factor"
	FeatureSSO           MiddlewareFeature = "sso"
	FeatureLDAP          MiddlewareFeature = "ldap"
	FeatureSAML          MiddlewareFeature = "saml"
	FeatureOpenIDConnect MiddlewareFeature = "openid_connect"

	// Authorization features
	FeatureRBAC            MiddlewareFeature = "rbac"
	FeatureABAC            MiddlewareFeature = "abac"
	FeatureACL             MiddlewareFeature = "acl"
	FeaturePolicyEngine    MiddlewareFeature = "policy_engine"
	FeatureAttributeBased  MiddlewareFeature = "attribute_based"
	FeatureContextAware    MiddlewareFeature = "context_aware"
	FeatureDynamicPolicies MiddlewareFeature = "dynamic_policies"

	// Logging features
	FeatureStructuredLogging  MiddlewareFeature = "structured_logging"
	FeatureRequestLogging     MiddlewareFeature = "request_logging"
	FeatureResponseLogging    MiddlewareFeature = "response_logging"
	FeatureAuditLogging       MiddlewareFeature = "audit_logging"
	FeatureErrorLogging       MiddlewareFeature = "error_logging"
	FeaturePerformanceLogging MiddlewareFeature = "performance_logging"

	// Monitoring features
	FeatureMetrics      MiddlewareFeature = "metrics"
	FeatureTracing      MiddlewareFeature = "tracing"
	FeatureHealthChecks MiddlewareFeature = "health_checks"
	FeatureProfiling    MiddlewareFeature = "profiling"
	FeatureAlerting     MiddlewareFeature = "alerting"
	FeatureDashboard    MiddlewareFeature = "dashboard"

	// Rate limiting features
	FeatureTokenBucket   MiddlewareFeature = "token_bucket"
	FeatureSlidingWindow MiddlewareFeature = "sliding_window"
	FeatureFixedWindow   MiddlewareFeature = "fixed_window"
	FeatureLeakyBucket   MiddlewareFeature = "leaky_bucket"
	FeatureDistributedRL MiddlewareFeature = "distributed_rate_limit"

	// Circuit breaker features
	FeatureCircuitBreaker MiddlewareFeature = "circuit_breaker"
	FeatureRetryLogic     MiddlewareFeature = "retry_logic"
	FeatureTimeout        MiddlewareFeature = "timeout"
	FeatureBulkhead       MiddlewareFeature = "bulkhead"
	FeatureFallback       MiddlewareFeature = "fallback"

	// Caching features
	FeatureMemoryCache       MiddlewareFeature = "memory_cache"
	FeatureRedisCache        MiddlewareFeature = "redis_cache"
	FeatureDistributedCache  MiddlewareFeature = "distributed_cache"
	FeatureCacheInvalidation MiddlewareFeature = "cache_invalidation"
	FeatureCacheWarming      MiddlewareFeature = "cache_warming"

	// Storage features
	FeatureFileStorage     MiddlewareFeature = "file_storage"
	FeatureObjectStorage   MiddlewareFeature = "object_storage"
	FeatureBlobStorage     MiddlewareFeature = "blob_storage"
	FeatureFileUpload      MiddlewareFeature = "file_upload"
	FeatureFileDownload    MiddlewareFeature = "file_download"
	FeatureFileMetadata    MiddlewareFeature = "file_metadata"
	FeatureFileCompression MiddlewareFeature = "file_compression"
	FeatureFileEncryption  MiddlewareFeature = "file_encryption"

	// Communication features
	FeatureHTTP      MiddlewareFeature = "http"
	FeatureGRPC      MiddlewareFeature = "grpc"
	FeatureWebSocket MiddlewareFeature = "websocket"
	FeatureGraphQL   MiddlewareFeature = "graphql"
	FeatureSSE       MiddlewareFeature = "sse"
	FeatureQUIC      MiddlewareFeature = "quic"

	// Messaging features
	FeatureKafka           MiddlewareFeature = "kafka"
	FeatureNATS            MiddlewareFeature = "nats"
	FeatureRabbitMQ        MiddlewareFeature = "rabbitmq"
	FeatureSQS             MiddlewareFeature = "sqs"
	FeatureRedis           MiddlewareFeature = "redis"
	FeatureMessageQueue    MiddlewareFeature = "message_queue"
	FeaturePubSub          MiddlewareFeature = "pubsub"
	FeatureEventSourcing   MiddlewareFeature = "event_sourcing"
	FeatureCQRS            MiddlewareFeature = "cqrs"
	FeatureDeadLetterQueue MiddlewareFeature = "dead_letter_queue"

	// Chaos engineering features
	FeatureLatencyInjection   MiddlewareFeature = "latency_injection"
	FeatureErrorInjection     MiddlewareFeature = "error_injection"
	FeatureResourceExhaustion MiddlewareFeature = "resource_exhaustion"
	FeatureNetworkPartition   MiddlewareFeature = "network_partition"
	FeatureServiceFailure     MiddlewareFeature = "service_failure"

	// Failover features
	FeatureLoadBalancing    MiddlewareFeature = "load_balancing"
	FeatureHealthChecking   MiddlewareFeature = "health_checking"
	FeatureServiceDiscovery MiddlewareFeature = "service_discovery"
	FeatureFailover         MiddlewareFeature = "failover"
	FeatureGracefulShutdown MiddlewareFeature = "graceful_shutdown"
)

// ConnectionInfo represents middleware provider connection information
type ConnectionInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Version  string `json:"version"`
	Secure   bool   `json:"secure"`
}

// MiddlewareConfig represents middleware configuration
type MiddlewareConfig struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Enabled    bool                   `json:"enabled"`
	Priority   int                    `json:"priority"`
	Timeout    time.Duration          `json:"timeout"`
	RetryCount int                    `json:"retry_count"`
	RetryDelay time.Duration          `json:"retry_delay"`
	Metadata   map[string]interface{} `json:"metadata"`
	Rules      []MiddlewareRule       `json:"rules"`
	Conditions []MiddlewareCondition  `json:"conditions"`
}

// MiddlewareRule represents a middleware rule
type MiddlewareRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Conditions  []MiddlewareCondition  `json:"conditions"`
	Actions     []MiddlewareAction     `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// MiddlewareCondition represents a middleware condition
type MiddlewareCondition struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Field    string                 `json:"field"`
	Operator string                 `json:"operator"`
	Value    interface{}            `json:"value"`
	Metadata map[string]interface{} `json:"metadata"`
}

// MiddlewareAction represents a middleware action
type MiddlewareAction struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// MiddlewareRequest represents a middleware request
type MiddlewareRequest struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Method    string                 `json:"method"`
	Path      string                 `json:"path"`
	Headers   map[string]string      `json:"headers"`
	Body      []byte                 `json:"body"`
	Query     map[string]string      `json:"query"`
	UserID    string                 `json:"user_id,omitempty"`
	ServiceID string                 `json:"service_id,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// MiddlewareResponse represents a middleware response
type MiddlewareResponse struct {
	ID         string                 `json:"id"`
	Success    bool                   `json:"success"`
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       []byte                 `json:"body"`
	Message    string                 `json:"message,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Duration   time.Duration          `json:"duration"`
}

// MiddlewareContext represents middleware execution context
type MiddlewareContext struct {
	Request     *MiddlewareRequest     `json:"request"`
	Response    *MiddlewareResponse    `json:"response"`
	User        *User                  `json:"user,omitempty"`
	Service     *Service               `json:"service,omitempty"`
	Environment map[string]interface{} `json:"environment,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// User represents user information in middleware context
type User struct {
	ID          string                 `json:"id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Roles       []string               `json:"roles,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Service represents service information in middleware context
type Service struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Region      string                 `json:"region,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MiddlewareStats represents middleware statistics
type MiddlewareStats struct {
	TotalRequests      int64                  `json:"total_requests"`
	SuccessfulRequests int64                  `json:"successful_requests"`
	FailedRequests     int64                  `json:"failed_requests"`
	AverageLatency     time.Duration          `json:"average_latency"`
	MaxLatency         time.Duration          `json:"max_latency"`
	MinLatency         time.Duration          `json:"min_latency"`
	ErrorRate          float64                `json:"error_rate"`
	Throughput         float64                `json:"throughput"`
	ActiveConnections  int64                  `json:"active_connections"`
	ProviderData       map[string]interface{} `json:"provider_data"`
}

// MiddlewareHandler represents a middleware handler function
type MiddlewareHandler func(ctx context.Context, req *MiddlewareRequest) (*MiddlewareResponse, error)

// HTTPMiddlewareHandler represents an HTTP middleware handler
type HTTPMiddlewareHandler func(next http.Handler) http.Handler

// MiddlewareChain represents a chain of middleware handlers
type MiddlewareChain struct {
	handlers []MiddlewareHandler
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain(handlers ...MiddlewareHandler) *MiddlewareChain {
	return &MiddlewareChain{
		handlers: handlers,
	}
}

// Add adds a handler to the chain
func (mc *MiddlewareChain) Add(handler MiddlewareHandler) *MiddlewareChain {
	mc.handlers = append(mc.handlers, handler)
	return mc
}

// Execute executes the middleware chain
func (mc *MiddlewareChain) Execute(ctx context.Context, req *MiddlewareRequest) (*MiddlewareResponse, error) {
	if len(mc.handlers) == 0 {
		return &MiddlewareResponse{
			ID:        req.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	// Execute handlers in order
	for _, handler := range mc.handlers {
		resp, err := handler(ctx, req)
		if err != nil {
			return &MiddlewareResponse{
				ID:        req.ID,
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			}, err
		}

		// If handler returns a response, use it as the new request for next handler
		if resp != nil {
			req = &MiddlewareRequest{
				ID:        req.ID,
				Type:      req.Type,
				Method:    req.Method,
				Path:      req.Path,
				Headers:   resp.Headers,
				Body:      resp.Body,
				Query:     req.Query,
				UserID:    req.UserID,
				ServiceID: req.ServiceID,
				Context:   resp.Context,
				Metadata:  resp.Metadata,
				Timestamp: time.Now(),
			}
		}
	}

	return &MiddlewareResponse{
		ID:        req.ID,
		Success:   true,
		Timestamp: time.Now(),
	}, nil
}
