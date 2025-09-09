package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/discovery/types"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesProvider implements service discovery using Kubernetes
type KubernetesProvider struct {
	client    kubernetes.Interface
	config    *KubernetesConfig
	logger    *logrus.Logger
	connected bool
}

// KubernetesConfig holds Kubernetes-specific configuration
type KubernetesConfig struct {
	KubeConfigPath string            `json:"kube_config_path"`
	Namespace      string            `json:"namespace"`
	InCluster      bool              `json:"in_cluster"`
	Timeout        time.Duration     `json:"timeout"`
	Metadata       map[string]string `json:"metadata"`
}

// NewKubernetesProvider creates a new Kubernetes provider
func NewKubernetesProvider(config *KubernetesConfig, logger *logrus.Logger) (*KubernetesProvider, error) {
	if config == nil {
		config = &KubernetesConfig{
			Namespace: "default",
			Timeout:   30 * time.Second,
			InCluster: true,
		}
	}

	var kubeConfig *rest.Config
	var err error

	if config.InCluster {
		// Use in-cluster configuration
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	} else {
		// Use kubeconfig file
		if config.KubeConfigPath != "" {
			kubeConfig, err = clientcmd.BuildConfigFromFlags("", config.KubeConfigPath)
		} else {
			kubeConfig, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create kubeconfig: %w", err)
		}
	}

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &KubernetesProvider{
		client:    client,
		config:    config,
		logger:    logger,
		connected: false,
	}, nil
}

// GetName returns the provider name
func (kp *KubernetesProvider) GetName() string {
	return "kubernetes"
}

// GetSupportedFeatures returns the features supported by this provider
func (kp *KubernetesProvider) GetSupportedFeatures() []types.ServiceFeature {
	return []types.ServiceFeature{
		types.FeatureDiscover,
		types.FeatureHealth,
		types.FeatureWatch,
		types.FeatureLoadBalancing,
		types.FeatureTags,
		types.FeatureMetadata,
		types.FeatureClustering,
		types.FeatureConsistency,
	}
}

// GetConnectionInfo returns connection information
func (kp *KubernetesProvider) GetConnectionInfo() *types.ConnectionInfo {
	status := types.StatusDisconnected
	if kp.connected {
		status = types.StatusConnected
	}

	host := "kubernetes-api"
	if kp.config.KubeConfigPath != "" {
		host = kp.config.KubeConfigPath
	}

	return &types.ConnectionInfo{
		Host:     host,
		Protocol: "https",
		Status:   status,
		Metadata: kp.config.Metadata,
	}
}

// Connect establishes connection to Kubernetes
func (kp *KubernetesProvider) Connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, kp.config.Timeout)
	defer cancel()

	// Test connection by getting the API version
	_, err := kp.client.Discovery().ServerVersion()
	if err != nil {
		kp.connected = false
		return fmt.Errorf("failed to connect to Kubernetes: %w", err)
	}

	kp.connected = true
	kp.logger.Info("Connected to Kubernetes")
	return nil
}

// Disconnect closes the connection to Kubernetes
func (kp *KubernetesProvider) Disconnect(ctx context.Context) error {
	kp.connected = false
	kp.logger.Info("Disconnected from Kubernetes")
	return nil
}

// Ping checks if the connection to Kubernetes is alive
func (kp *KubernetesProvider) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, kp.config.Timeout)
	defer cancel()

	_, err := kp.client.Discovery().ServerVersion()
	if err != nil {
		kp.connected = false
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// IsConnected returns the connection status
func (kp *KubernetesProvider) IsConnected() bool {
	return kp.connected
}

// RegisterService is not supported in Kubernetes (services are managed by Kubernetes)
func (kp *KubernetesProvider) RegisterService(ctx context.Context, registration *types.ServiceRegistration) error {
	return fmt.Errorf("service registration is not supported in Kubernetes - services are managed by Kubernetes resources")
}

// DeregisterService is not supported in Kubernetes (services are managed by Kubernetes)
func (kp *KubernetesProvider) DeregisterService(ctx context.Context, serviceID string) error {
	return fmt.Errorf("service deregistration is not supported in Kubernetes - services are managed by Kubernetes resources")
}

// UpdateService is not supported in Kubernetes (services are managed by Kubernetes)
func (kp *KubernetesProvider) UpdateService(ctx context.Context, registration *types.ServiceRegistration) error {
	return fmt.Errorf("service update is not supported in Kubernetes - services are managed by Kubernetes resources")
}

// DiscoverServices discovers services from Kubernetes
func (kp *KubernetesProvider) DiscoverServices(ctx context.Context, query *types.ServiceQuery) ([]*types.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, kp.config.Timeout)
	defer cancel()

	services, err := kp.client.CoreV1().Services(kp.config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	result := make([]*types.Service, 0)

	for _, service := range services.Items {
		// Filter by service name if specified
		if query.Name != "" && service.Name != query.Name {
			continue
		}

		// Get endpoints for this service
		endpoints, err := kp.client.CoreV1().Endpoints(kp.config.Namespace).Get(ctx, service.Name, metav1.GetOptions{})
		if err != nil {
			kp.logger.WithError(err).WithField("service", service.Name).Warn("Failed to get endpoints")
			continue
		}

		instances := make([]*types.ServiceInstance, 0)

		// Convert endpoints to service instances
		for _, subset := range endpoints.Subsets {
			for _, address := range subset.Addresses {
				for _, port := range subset.Ports {
					instance := &types.ServiceInstance{
						ID:       fmt.Sprintf("%s-%s-%d", service.Name, address.IP, port.Port),
						Name:     service.Name,
						Address:  address.IP,
						Port:     int(port.Port),
						Protocol: string(port.Protocol),
						Tags:     []string{service.Namespace},
						Metadata: map[string]string{
							"namespace": service.Namespace,
							"port_name": port.Name,
						},
						Health:      types.HealthPassing, // Assume healthy if endpoint exists
						LastUpdated: time.Now(),
					}

					// Add node name if available
					if address.NodeName != nil {
						instance.Metadata["node_name"] = *address.NodeName
					}

					// Add hostname if available
					if address.Hostname != "" {
						instance.Metadata["hostname"] = address.Hostname
					}

					instances = append(instances, instance)
				}
			}
		}

		// Add service labels as metadata
		metadata := make(map[string]string)
		for key, value := range service.Labels {
			metadata[key] = value
		}

		// Add service annotations as metadata
		for key, value := range service.Annotations {
			metadata["annotation_"+key] = value
		}

		serviceObj := &types.Service{
			Name:      service.Name,
			Instances: instances,
			Tags:      []string{service.Namespace},
			Metadata:  metadata,
		}

		result = append(result, serviceObj)
	}

	return result, nil
}

// GetService gets a specific service from Kubernetes
func (kp *KubernetesProvider) GetService(ctx context.Context, serviceName string) (*types.Service, error) {
	query := &types.ServiceQuery{
		Name: serviceName,
	}

	services, err := kp.DiscoverServices(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	return services[0], nil
}

// GetServiceInstance gets a specific service instance from Kubernetes
func (kp *KubernetesProvider) GetServiceInstance(ctx context.Context, serviceName, instanceID string) (*types.ServiceInstance, error) {
	service, err := kp.GetService(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	for _, instance := range service.Instances {
		if instance.ID == instanceID {
			return instance, nil
		}
	}

	return nil, fmt.Errorf("service instance %s not found for service %s", instanceID, serviceName)
}

// SetHealth is not supported in Kubernetes (health is managed by Kubernetes)
func (kp *KubernetesProvider) SetHealth(ctx context.Context, serviceID string, health types.HealthStatus) error {
	return fmt.Errorf("health setting is not supported in Kubernetes - health is managed by Kubernetes")
}

// GetHealth gets the health status of a service from Kubernetes
func (kp *KubernetesProvider) GetHealth(ctx context.Context, serviceID string) (types.HealthStatus, error) {
	// In Kubernetes, we can check if endpoints exist and are ready
	// This is a simplified health check
	ctx, cancel := context.WithTimeout(ctx, kp.config.Timeout)
	defer cancel()

	// Parse service ID to extract service name
	// Assuming service ID format: serviceName-instanceID
	serviceName := serviceID
	if len(serviceID) > 0 {
		// Try to get the service name from the ID
		// This is a simplified approach - in practice, you might need more sophisticated parsing
		serviceName = serviceID
	}

	endpoints, err := kp.client.CoreV1().Endpoints(kp.config.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return types.HealthUnknown, fmt.Errorf("failed to get endpoints: %w", err)
	}

	// Check if there are ready endpoints
	readyEndpoints := 0
	for _, subset := range endpoints.Subsets {
		readyEndpoints += len(subset.Addresses)
	}

	if readyEndpoints > 0 {
		return types.HealthPassing
	}

	return types.HealthCritical
}

// WatchServices watches for service changes in Kubernetes
func (kp *KubernetesProvider) WatchServices(ctx context.Context, options *types.WatchOptions) (<-chan *types.ServiceEvent, error) {
	eventChan := make(chan *types.ServiceEvent, 100)

	go func() {
		defer close(eventChan)

		watch, err := kp.client.CoreV1().Services(kp.config.Namespace).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			kp.logger.WithError(err).Error("Failed to watch services")
			return
		}
		defer watch.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-watch.ResultChan():
				if event.Object == nil {
					continue
				}

				service, ok := event.Object.(*metav1.Object)
				if !ok {
					continue
				}

				// Convert Kubernetes event to our service event
				serviceEvent := &types.ServiceEvent{
					Type:      convertKubernetesEventType(event.Type),
					Timestamp: time.Now(),
					Provider:  kp.GetName(),
				}

				// Get the full service details
				serviceObj, err := kp.GetService(ctx, service.GetName())
				if err != nil {
					kp.logger.WithError(err).WithField("service", service.GetName()).Warn("Failed to get service details")
					continue
				}

				serviceEvent.Service = serviceObj

				select {
				case eventChan <- serviceEvent:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return eventChan, nil
}

// StopWatch stops watching for service changes (not applicable for Kubernetes)
func (kp *KubernetesProvider) StopWatch(ctx context.Context, watchID string) error {
	// Kubernetes doesn't have a specific stop watch mechanism
	// The context cancellation handles this
	return nil
}

// GetStats returns discovery statistics from Kubernetes
func (kp *KubernetesProvider) GetStats(ctx context.Context) (*types.DiscoveryStats, error) {
	ctx, cancel := context.WithTimeout(ctx, kp.config.Timeout)
	defer cancel()

	services, err := kp.client.CoreV1().Services(kp.config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	servicesCount := int64(len(services.Items))
	instancesCount := int64(0)

	// Count instances by checking endpoints
	for _, service := range services.Items {
		endpoints, err := kp.client.CoreV1().Endpoints(kp.config.Namespace).Get(ctx, service.Name, metav1.GetOptions{})
		if err != nil {
			continue
		}

		for _, subset := range endpoints.Subsets {
			instancesCount += int64(len(subset.Addresses))
		}
	}

	return &types.DiscoveryStats{
		ServicesRegistered: servicesCount,
		InstancesActive:    instancesCount,
		QueriesProcessed:   0, // Kubernetes doesn't provide this metric
		Uptime:             0, // Would need to track this separately
		LastUpdate:         time.Now(),
		Provider:           kp.GetName(),
	}, nil
}

// ListServices returns a list of all service names from Kubernetes
func (kp *KubernetesProvider) ListServices(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, kp.config.Timeout)
	defer cancel()

	services, err := kp.client.CoreV1().Services(kp.config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	serviceNames := make([]string, 0, len(services.Items))
	for _, service := range services.Items {
		serviceNames = append(serviceNames, service.Name)
	}

	return serviceNames, nil
}

// convertKubernetesEventType converts Kubernetes event type to our event type
func convertKubernetesEventType(eventType string) types.ServiceEventType {
	switch eventType {
	case "ADDED":
		return types.EventServiceRegistered
	case "DELETED":
		return types.EventServiceDeregistered
	case "MODIFIED":
		return types.EventServiceUpdated
	default:
		return types.EventServiceUpdated
	}
}
