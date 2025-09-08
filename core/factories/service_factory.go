package factories

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/core/interfaces"
)

// ServiceType represents the type of service
type ServiceType string

const (
	ServiceTypeHTTP       ServiceType = "http"
	ServiceTypeGRPC       ServiceType = "grpc"
	ServiceTypeMessage    ServiceType = "message"
	ServiceTypeBackground ServiceType = "background"
	ServiceTypeHybrid     ServiceType = "hybrid"
)

// ServiceFactory creates service instances
type ServiceFactory struct {
	configs map[ServiceType]interface{}
}

// NewServiceFactory creates a new service factory
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{
		configs: make(map[ServiceType]interface{}),
	}
}

// RegisterConfig registers a service configuration
func (f *ServiceFactory) RegisterConfig(serviceType ServiceType, config interface{}) {
	f.configs[serviceType] = config
}

// CreateService creates a service instance based on type
func (f *ServiceFactory) CreateService(ctx context.Context, serviceType ServiceType, name string) (interfaces.Service, error) {
	config, exists := f.configs[serviceType]
	if !exists {
		return nil, fmt.Errorf("no configuration found for service type: %s", serviceType)
	}

	switch serviceType {
	case ServiceTypeHTTP:
		httpConfig, ok := config.(*HTTPServiceConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for HTTP service")
		}
		return NewHTTPService(name, httpConfig)

	case ServiceTypeGRPC:
		grpcConfig, ok := config.(*GRPCServiceConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for gRPC service")
		}
		return NewGRPCService(name, grpcConfig)

	case ServiceTypeMessage:
		messageConfig, ok := config.(*MessageServiceConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for message service")
		}
		return NewMessageService(name, messageConfig)

	case ServiceTypeBackground:
		backgroundConfig, ok := config.(*BackgroundServiceConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for background service")
		}
		return NewBackgroundService(name, backgroundConfig)

	case ServiceTypeHybrid:
		hybridConfig, ok := config.(*HybridServiceConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for hybrid service")
		}
		return NewHybridService(name, hybridConfig)

	default:
		return nil, fmt.Errorf("unsupported service type: %s", serviceType)
	}
}

// CreateMultipleServices creates multiple service instances
func (f *ServiceFactory) CreateMultipleServices(ctx context.Context, serviceTypes []ServiceType, names []string) (map[string]interfaces.Service, error) {
	if len(serviceTypes) != len(names) {
		return nil, fmt.Errorf("service types and names must have the same length")
	}

	services := make(map[string]interfaces.Service)

	for i, serviceType := range serviceTypes {
		service, err := f.CreateService(ctx, serviceType, names[i])
		if err != nil {
			// Stop already created services
			for _, createdService := range services {
				createdService.Stop(ctx)
			}
			return nil, fmt.Errorf("failed to create service %s: %w", names[i], err)
		}
		services[names[i]] = service
	}

	return services, nil
}

// GetSupportedTypes returns supported service types
func (f *ServiceFactory) GetSupportedTypes() []ServiceType {
	return []ServiceType{
		ServiceTypeHTTP,
		ServiceTypeGRPC,
		ServiceTypeMessage,
		ServiceTypeBackground,
		ServiceTypeHybrid,
	}
}

// IsTypeSupported checks if a service type is supported
func (f *ServiceFactory) IsTypeSupported(serviceType ServiceType) bool {
	supportedTypes := f.GetSupportedTypes()
	for _, supportedType := range supportedTypes {
		if supportedType == serviceType {
			return true
		}
	}
	return false
}

// Configuration structures for different service types

// HTTPServiceConfig represents HTTP service configuration
type HTTPServiceConfig struct {
	Port        string
	Host        string
	BasePath    string
	Middlewares []string
	Routes      []RouteConfig
}

// GRPCServiceConfig represents gRPC service configuration
type GRPCServiceConfig struct {
	Port         string
	Host         string
	Services     []string
	Interceptors []string
}

// MessageServiceConfig represents message service configuration
type MessageServiceConfig struct {
	Topics         []string
	ConsumerGroups []string
	Handlers       []HandlerConfig
}

// BackgroundServiceConfig represents background service configuration
type BackgroundServiceConfig struct {
	Interval time.Duration
	Jobs     []JobConfig
}

// HybridServiceConfig represents hybrid service configuration
type HybridServiceConfig struct {
	HTTP       *HTTPServiceConfig
	GRPC       *GRPCServiceConfig
	Message    *MessageServiceConfig
	Background *BackgroundServiceConfig
}

// RouteConfig represents route configuration
type RouteConfig struct {
	Method  string
	Path    string
	Handler string
	Auth    bool
}

// HandlerConfig represents message handler configuration
type HandlerConfig struct {
	Topic   string
	Group   string
	Handler string
}

// JobConfig represents background job configuration
type JobConfig struct {
	Name     string
	Schedule string
	Handler  string
}

// Service factory functions (to be implemented in respective packages)

// NewHTTPService creates a new HTTP service
func NewHTTPService(name string, config *HTTPServiceConfig) (interfaces.Service, error) {
	// Implementation will be in http package
	return nil, fmt.Errorf("not implemented")
}

// NewGRPCService creates a new gRPC service
func NewGRPCService(name string, config *GRPCServiceConfig) (interfaces.Service, error) {
	// Implementation will be in grpc package
	return nil, fmt.Errorf("not implemented")
}

// NewMessageService creates a new message service
func NewMessageService(name string, config *MessageServiceConfig) (interfaces.Service, error) {
	// Implementation will be in messaging package
	return nil, fmt.Errorf("not implemented")
}

// NewBackgroundService creates a new background service
func NewBackgroundService(name string, config *BackgroundServiceConfig) (interfaces.Service, error) {
	// Implementation will be in background package
	return nil, fmt.Errorf("not implemented")
}

// NewHybridService creates a new hybrid service
func NewHybridService(name string, config *HybridServiceConfig) (interfaces.Service, error) {
	// Implementation will combine multiple service types
	return nil, fmt.Errorf("not implemented")
}
