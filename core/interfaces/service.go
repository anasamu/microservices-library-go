package interfaces

import (
	"context"
	"time"
)

// Service interface defines the basic service operations
type Service interface {
	// Lifecycle operations
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	HealthCheck(ctx context.Context) error

	// Service information
	GetName() string
	GetVersion() string
	GetDescription() string
	GetMetadata() map[string]interface{}

	// Dependencies
	GetDependencies() []string
	SetDependency(name string, service Service) error
	GetDependency(name string) (Service, error)
}

// HTTPService interface defines HTTP service operations
type HTTPService interface {
	Service

	// HTTP operations
	RegisterRoutes(router HTTPRouter) error
	GetPort() string
	GetHost() string
	GetBasePath() string

	// Middleware
	AddMiddleware(middleware HTTPMiddleware) error
	GetMiddlewares() []HTTPMiddleware
}

// GRPCService interface defines gRPC service operations
type GRPCService interface {
	Service

	// gRPC operations
	RegisterServices(server GRPCServer) error
	GetPort() string
	GetHost() string

	// Interceptors
	AddUnaryInterceptor(interceptor GRPCUnaryInterceptor) error
	AddStreamInterceptor(interceptor GRPCStreamInterceptor) error
	GetUnaryInterceptors() []GRPCUnaryInterceptor
	GetStreamInterceptors() []GRPCStreamInterceptor
}

// MessageService interface defines message-driven service operations
type MessageService interface {
	Service

	// Message operations
	RegisterHandlers(registry MessageHandlerRegistry) error
	GetTopics() []string
	GetConsumerGroups() []string

	// Message processing
	ProcessMessage(ctx context.Context, message Message) error
	ProcessBatch(ctx context.Context, messages []Message) error
}

// BackgroundService interface defines background service operations
type BackgroundService interface {
	Service

	// Background operations
	StartBackground(ctx context.Context) error
	StopBackground(ctx context.Context) error
	GetInterval() time.Duration
	SetInterval(interval time.Duration)

	// Job operations
	ScheduleJob(job Job) error
	CancelJob(jobID string) error
	GetJobs() []Job
}

// ServiceRegistry interface defines service registry operations
type ServiceRegistry interface {
	// Service registration
	Register(ctx context.Context, service ServiceInfo) error
	Unregister(ctx context.Context, serviceID string) error
	Update(ctx context.Context, service ServiceInfo) error

	// Service discovery
	Discover(ctx context.Context, serviceName string) ([]ServiceInfo, error)
	DiscoverByTag(ctx context.Context, tag string) ([]ServiceInfo, error)
	GetService(ctx context.Context, serviceID string) (*ServiceInfo, error)

	// Health monitoring
	MarkHealthy(ctx context.Context, serviceID string) error
	MarkUnhealthy(ctx context.Context, serviceID string) error
	GetHealthyServices(ctx context.Context, serviceName string) ([]ServiceInfo, error)

	// Watch operations
	Watch(ctx context.Context, serviceName string) (<-chan ServiceEvent, error)
	StopWatch(ctx context.Context, serviceName string) error
}

// ServiceInfo represents service information
type ServiceInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Host        string                 `json:"host"`
	Port        int                    `json:"port"`
	Protocol    string                 `json:"protocol"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	HealthCheck HealthCheckInfo        `json:"health_check"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// HealthCheckInfo represents health check information
type HealthCheckInfo struct {
	Path     string        `json:"path"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Retries  int           `json:"retries"`
}

// ServiceEvent represents a service registry event
type ServiceEvent struct {
	Type    ServiceEventType `json:"type"`
	Service ServiceInfo      `json:"service"`
	Time    time.Time        `json:"time"`
}

// ServiceEventType represents the type of service event
type ServiceEventType string

const (
	ServiceEventTypeRegistered   ServiceEventType = "registered"
	ServiceEventTypeUnregistered ServiceEventType = "unregistered"
	ServiceEventTypeUpdated      ServiceEventType = "updated"
	ServiceEventTypeHealthy      ServiceEventType = "healthy"
	ServiceEventTypeUnhealthy    ServiceEventType = "unhealthy"
)

// Job represents a background job
type Job struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Payload    interface{}            `json:"payload"`
	Schedule   string                 `json:"schedule"`
	NextRun    time.Time              `json:"next_run"`
	LastRun    *time.Time             `json:"last_run"`
	Status     JobStatus              `json:"status"`
	Retries    int                    `json:"retries"`
	MaxRetries int                    `json:"max_retries"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// HTTPRouter interface defines HTTP router operations
type HTTPRouter interface {
	// Route registration
	GET(path string, handler HTTPHandler)
	POST(path string, handler HTTPHandler)
	PUT(path string, handler HTTPHandler)
	DELETE(path string, handler HTTPHandler)
	PATCH(path string, handler HTTPHandler)
	OPTIONS(path string, handler HTTPHandler)
	HEAD(path string, handler HTTPHandler)

	// Group operations
	Group(prefix string) HTTPRouter

	// Middleware
	Use(middleware HTTPMiddleware)

	// Static files
	Static(relativePath, root string)
	StaticFile(relativePath, filepath string)

	// Start/Stop
	Run(addr string) error
	Shutdown(ctx context.Context) error
}

// HTTPHandler defines the signature for HTTP handlers
type HTTPHandler func(ctx HTTPContext) error

// HTTPContext represents HTTP request context
type HTTPContext interface {
	// Request information
	GetMethod() string
	GetPath() string
	GetURL() string
	GetHeader(key string) string
	GetHeaders() map[string]string
	GetQuery(key string) string
	GetQueryParams() map[string]string
	GetBody() []byte
	GetFormValue(key string) string
	GetFormParams() map[string]string

	// Response operations
	SetStatus(code int)
	SetHeader(key, value string)
	SetHeaders(headers map[string]string)
	Write(data []byte) error
	WriteString(data string) error
	WriteJSON(data interface{}) error
	WriteXML(data interface{}) error

	// Context operations
	GetContext() context.Context
	SetContext(ctx context.Context)

	// User information
	GetUser() interface{}
	SetUser(user interface{})

	// Path parameters
	GetParam(key string) string
	GetParams() map[string]string
}

// HTTPMiddleware defines the signature for HTTP middleware
type HTTPMiddleware func(next HTTPHandler) HTTPHandler

// GRPCServer interface defines gRPC server operations
type GRPCServer interface {
	// Service registration
	RegisterService(desc *ServiceDesc, impl interface{})
	RegisterStreamService(desc *StreamServiceDesc, impl interface{})

	// Interceptors
	AddUnaryInterceptor(interceptor GRPCUnaryInterceptor)
	AddStreamInterceptor(interceptor GRPCStreamInterceptor)

	// Start/Stop
	Start(addr string) error
	Stop() error
	GracefulStop() error
}

// ServiceDesc represents gRPC service description
type ServiceDesc struct {
	ServiceName string
	Methods     []MethodDesc
	Streams     []StreamDesc
	Metadata    interface{}
}

// StreamServiceDesc represents gRPC stream service description
type StreamServiceDesc struct {
	ServiceName string
	HandlerType interface{}
	Streams     []StreamDesc
	Metadata    interface{}
}

// MethodDesc represents gRPC method description
type MethodDesc struct {
	MethodName string
	Handler    interface{}
}

// StreamDesc represents gRPC stream description
type StreamDesc struct {
	StreamName    string
	Handler       interface{}
	ServerStreams bool
	ClientStreams bool
}

// GRPCUnaryInterceptor defines the signature for gRPC unary interceptors
type GRPCUnaryInterceptor func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (interface{}, error)

// GRPCStreamInterceptor defines the signature for gRPC stream interceptors
type GRPCStreamInterceptor func(srv interface{}, ss ServerStream, info *StreamServerInfo, handler StreamHandler) error

// UnaryServerInfo represents unary server information
type UnaryServerInfo struct {
	Server     interface{}
	FullMethod string
}

// StreamServerInfo represents stream server information
type StreamServerInfo struct {
	FullMethod     string
	IsClientStream bool
	IsServerStream bool
}

// UnaryHandler defines the signature for unary handlers
type UnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

// StreamHandler defines the signature for stream handlers
type StreamHandler func(srv interface{}, stream ServerStream) error

// ServerStream interface defines server stream operations
type ServerStream interface {
	SetHeader(metadata map[string]string) error
	SendHeader(metadata map[string]string) error
	SetTrailer(metadata map[string]string)
	Context() context.Context
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
}

// MessageHandlerRegistry interface defines message handler registry operations
type MessageHandlerRegistry interface {
	// Handler registration
	RegisterHandler(topic string, handler MessageHandler) error
	RegisterHandlerWithGroup(topic, group string, handler MessageHandler) error
	UnregisterHandler(topic string) error

	// Handler management
	GetHandler(topic string) (MessageHandler, error)
	ListHandlers() map[string]MessageHandler

	// Start/Stop
	Start() error
	Stop() error
}
