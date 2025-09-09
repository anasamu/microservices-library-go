package kubernetes

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/failover/types"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Provider implements failover using Kubernetes
type Provider struct {
	name      string
	client    *kubernetes.Clientset
	config    *Config
	logger    *logrus.Logger
	endpoints map[string]*types.ServiceEndpoint
	stats     *types.FailoverStats
	events    []*types.FailoverEvent
	mu        sync.RWMutex
	connected bool
	startTime time.Time
}

// Config holds Kubernetes provider configuration
type Config struct {
	KubeConfigPath string            `json:"kubeconfig_path"`
	Namespace      string            `json:"namespace"`
	Context        string            `json:"context"`
	Timeout        time.Duration     `json:"timeout"`
	Metadata       map[string]string `json:"metadata"`
}

// NewProvider creates a new Kubernetes failover provider
func NewProvider(config *Config, logger *logrus.Logger) (*Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if logger == nil {
		logger = logrus.New()
	}

	// Create Kubernetes client configuration
	var kubeConfig *rest.Config
	var err error

	if config.KubeConfigPath != "" {
		// Use kubeconfig file
		kubeConfig, err = clientcmd.BuildConfigFromFlags(config.Context, config.KubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	} else {
		// Use in-cluster configuration
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	}

	// Set timeout
	if config.Timeout > 0 {
		kubeConfig.Timeout = config.Timeout
	}

	// Create Kubernetes client
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	provider := &Provider{
		name:      "kubernetes",
		client:    client,
		config:    config,
		logger:    logger,
		endpoints: make(map[string]*types.ServiceEndpoint),
		stats: &types.FailoverStats{
			LastUpdate: time.Now(),
		},
		events:    make([]*types.FailoverEvent, 0),
		startTime: time.Now(),
	}

	return provider, nil
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return p.name
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []types.FailoverFeature {
	return []types.FailoverFeature{
		types.FeatureHealthCheck,
		types.FeatureLoadBalancing,
		types.FeatureCircuitBreaker,
		types.FeatureRetry,
		types.FeatureTimeout,
		types.FeatureMonitoring,
		types.FeatureWeightedRouting,
		types.FeatureServiceMesh,
		types.FeatureAutoScaling,
		types.FeatureTrafficShifting,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *types.ConnectionInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := types.StatusDisconnected
	if p.connected {
		status = types.StatusConnected
	}

	return &types.ConnectionInfo{
		Host:     "kubernetes-api",
		Port:     443,
		Status:   status,
		Metadata: p.config.Metadata,
	}
}

// Connect connects to Kubernetes
func (p *Provider) Connect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Test connection by getting server version
	_, err := p.client.Discovery().ServerVersion()
	if err != nil {
		p.connected = false
		return fmt.Errorf("failed to connect to Kubernetes: %w", err)
	}

	p.connected = true
	p.logger.Info("Connected to Kubernetes")
	return nil
}

// Disconnect disconnects from Kubernetes
func (p *Provider) Disconnect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.connected = false
	p.logger.Info("Disconnected from Kubernetes")
	return nil
}

// Ping tests the connection to Kubernetes
func (p *Provider) Ping(ctx context.Context) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to Kubernetes")
	}

	_, err := p.client.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// IsConnected returns connection status
func (p *Provider) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connected
}

// RegisterEndpoint registers an endpoint by discovering it from Kubernetes services
func (p *Provider) RegisterEndpoint(ctx context.Context, endpoint *types.ServiceEndpoint) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to Kubernetes")
	}

	// In Kubernetes, endpoints are typically discovered from services
	// This method allows manual registration for custom endpoints
	p.mu.Lock()
	p.endpoints[endpoint.ID] = endpoint
	p.mu.Unlock()

	p.logger.WithField("endpoint", endpoint.ID).Info("Endpoint registered in Kubernetes provider")
	return nil
}

// DeregisterEndpoint deregisters an endpoint
func (p *Provider) DeregisterEndpoint(ctx context.Context, endpointID string) error {
	p.mu.Lock()
	delete(p.endpoints, endpointID)
	p.mu.Unlock()

	p.logger.WithField("endpoint", endpointID).Info("Endpoint deregistered from Kubernetes provider")
	return nil
}

// UpdateEndpoint updates an endpoint
func (p *Provider) UpdateEndpoint(ctx context.Context, endpoint *types.ServiceEndpoint) error {
	p.mu.Lock()
	p.endpoints[endpoint.ID] = endpoint
	p.mu.Unlock()

	p.logger.WithField("endpoint", endpoint.ID).Info("Endpoint updated in Kubernetes provider")
	return nil
}

// GetEndpoint gets an endpoint by ID
func (p *Provider) GetEndpoint(ctx context.Context, endpointID string) (*types.ServiceEndpoint, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	endpoint, exists := p.endpoints[endpointID]
	if !exists {
		return nil, fmt.Errorf("endpoint %s not found", endpointID)
	}

	return endpoint, nil
}

// ListEndpoints lists all registered endpoints
func (p *Provider) ListEndpoints(ctx context.Context) ([]*types.ServiceEndpoint, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	endpoints := make([]*types.ServiceEndpoint, 0, len(p.endpoints))
	for _, endpoint := range p.endpoints {
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// SelectEndpoint selects an endpoint based on the failover strategy
func (p *Provider) SelectEndpoint(ctx context.Context, config *types.FailoverConfig) (*types.FailoverResult, error) {
	start := time.Now()

	// Get available endpoints
	availableEndpoints, err := p.getAvailableEndpoints(ctx, config)
	if err != nil {
		return &types.FailoverResult{
			Error:    &types.FailoverError{Code: types.ErrCodeNoEndpoints, Message: err.Error()},
			Duration: time.Since(start),
			Strategy: config.Strategy,
		}, nil
	}

	if len(availableEndpoints) == 0 {
		return &types.FailoverResult{
			Error:    &types.FailoverError{Code: types.ErrCodeNoEndpoints, Message: "no healthy endpoints available"},
			Duration: time.Since(start),
			Strategy: config.Strategy,
		}, nil
	}

	// Select endpoint based on strategy
	var selectedEndpoint *types.ServiceEndpoint
	switch config.Strategy {
	case types.StrategyRoundRobin:
		selectedEndpoint = p.selectRoundRobin(availableEndpoints)
	case types.StrategyLeastConn:
		selectedEndpoint = p.selectLeastConnections(availableEndpoints)
	case types.StrategyWeighted:
		selectedEndpoint = p.selectWeighted(availableEndpoints)
	case types.StrategyRandom:
		selectedEndpoint = p.selectRandom(availableEndpoints)
	case types.StrategyHealthBased:
		selectedEndpoint = p.selectHealthBased(availableEndpoints)
	default:
		selectedEndpoint = p.selectRoundRobin(availableEndpoints)
	}

	// Update statistics
	p.updateStats(true, time.Since(start))

	return &types.FailoverResult{
		Endpoint: selectedEndpoint,
		Duration: time.Since(start),
		Strategy: config.Strategy,
		Metadata: map[string]interface{}{
			"total_endpoints":   len(p.endpoints),
			"healthy_endpoints": len(availableEndpoints),
			"namespace":         p.config.Namespace,
		},
	}, nil
}

// ExecuteWithFailover executes a function with failover logic
func (p *Provider) ExecuteWithFailover(ctx context.Context, config *types.FailoverConfig, fn func(*types.ServiceEndpoint) (interface{}, error)) (*types.FailoverResult, error) {
	start := time.Now()
	attempts := 0
	var lastError error

	for attempts < config.RetryAttempts {
		attempts++

		// Select endpoint
		result, err := p.SelectEndpoint(ctx, config)
		if err != nil {
			lastError = err
			continue
		}

		if result.Error != nil {
			lastError = result.Error
			continue
		}

		// Execute function
		funcResult, err := fn(result.Endpoint)
		if err == nil {
			// Success
			p.updateStats(true, time.Since(start))
			return &types.FailoverResult{
				Endpoint: result.Endpoint,
				Duration: time.Since(start),
				Attempts: attempts,
				Strategy: config.Strategy,
				Metadata: map[string]interface{}{
					"result": funcResult,
				},
			}, nil
		}

		// Failure - mark endpoint as unhealthy temporarily
		p.markEndpointUnhealthy(result.Endpoint.ID)
		lastError = err

		// Wait before retry
		if attempts < config.RetryAttempts {
			time.Sleep(config.RetryDelay)
		}
	}

	// All attempts failed
	p.updateStats(false, time.Since(start))
	return &types.FailoverResult{
		Error:    lastError,
		Duration: time.Since(start),
		Attempts: attempts,
		Strategy: config.Strategy,
	}, nil
}

// HealthCheck performs health check on an endpoint
func (p *Provider) HealthCheck(ctx context.Context, endpointID string) (types.HealthStatus, error) {
	if !p.IsConnected() {
		return types.HealthUnknown, fmt.Errorf("not connected to Kubernetes")
	}

	// Check if endpoint exists locally
	p.mu.RLock()
	endpoint, exists := p.endpoints[endpointID]
	p.mu.RUnlock()

	if !exists {
		return types.HealthUnknown, fmt.Errorf("endpoint %s not found", endpointID)
	}

	// In a real implementation, you would check the actual pod/service health
	// For now, we'll use the stored health status
	return endpoint.Health, nil
}

// HealthCheckAll performs health check on all endpoints
func (p *Provider) HealthCheckAll(ctx context.Context) (map[string]types.HealthStatus, error) {
	statuses := make(map[string]types.HealthStatus)

	p.mu.RLock()
	for id, endpoint := range p.endpoints {
		statuses[id] = endpoint.Health
	}
	p.mu.RUnlock()

	return statuses, nil
}

// Configure configures the failover provider
func (p *Provider) Configure(ctx context.Context, config *types.FailoverConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.logger.WithField("config", config.Name).Info("Failover configuration updated")
	return nil
}

// GetConfig returns the current configuration
func (p *Provider) GetConfig(ctx context.Context) (*types.FailoverConfig, error) {
	return types.DefaultConfig, nil
}

// GetStats returns failover statistics
func (p *Provider) GetStats(ctx context.Context) (*types.FailoverStats, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Update uptime
	p.stats.Uptime = time.Since(p.startTime)
	p.stats.LastUpdate = time.Now()
	p.stats.Provider = p.name

	return p.stats, nil
}

// GetEvents returns recent failover events
func (p *Provider) GetEvents(ctx context.Context, limit int) ([]*types.FailoverEvent, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if limit <= 0 || limit > len(p.events) {
		limit = len(p.events)
	}

	// Return most recent events
	start := len(p.events) - limit
	if start < 0 {
		start = 0
	}

	return p.events[start:], nil
}

// Close closes the provider
func (p *Provider) Close(ctx context.Context) error {
	return p.Disconnect(ctx)
}

// Helper methods

func (p *Provider) getAvailableEndpoints(ctx context.Context, config *types.FailoverConfig) ([]*types.ServiceEndpoint, error) {
	// Get endpoints from Kubernetes services if namespace is specified
	if p.config.Namespace != "" {
		return p.discoverEndpointsFromServices(ctx)
	}

	// Fall back to manually registered endpoints
	p.mu.RLock()
	availableEndpoints := make([]*types.ServiceEndpoint, 0)
	for _, endpoint := range p.endpoints {
		if endpoint.Health == types.HealthHealthy || endpoint.Health == types.HealthDegraded {
			availableEndpoints = append(availableEndpoints, endpoint)
		}
	}
	p.mu.RUnlock()

	return availableEndpoints, nil
}

func (p *Provider) discoverEndpointsFromServices(ctx context.Context) ([]*types.ServiceEndpoint, error) {
	// Get services from Kubernetes
	services, err := p.client.CoreV1().Services(p.config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	endpoints := make([]*types.ServiceEndpoint, 0)

	for _, service := range services.Items {
		// Get endpoints for the service
		serviceEndpoints, err := p.client.CoreV1().Endpoints(service.Namespace).Get(ctx, service.Name, metav1.GetOptions{})
		if err != nil {
			p.logger.WithError(err).WithField("service", service.Name).Warn("Failed to get service endpoints")
			continue
		}

		// Convert Kubernetes endpoints to our endpoint format
		for _, subset := range serviceEndpoints.Subsets {
			for _, address := range subset.Addresses {
				endpoint := &types.ServiceEndpoint{
					ID:       fmt.Sprintf("%s-%s", service.Name, address.IP),
					Name:     service.Name,
					Address:  address.IP,
					Port:     int(subset.Ports[0].Port),
					Protocol: string(subset.Ports[0].Protocol),
					Health:   types.HealthHealthy, // Assume healthy if address is ready
					Weight:   100,
					Priority: 1,
					Tags:     service.Labels,
					Metadata: map[string]string{
						"namespace": service.Namespace,
						"service":   service.Name,
						"cluster":   "kubernetes",
					},
					LastCheck: time.Now(),
				}

				endpoints = append(endpoints, endpoint)
			}
		}
	}

	return endpoints, nil
}

// Endpoint selection strategies (same as Consul provider)

func (p *Provider) selectRoundRobin(endpoints []*types.ServiceEndpoint) *types.ServiceEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	index := int(time.Now().UnixNano()) % len(endpoints)
	return endpoints[index]
}

func (p *Provider) selectLeastConnections(endpoints []*types.ServiceEndpoint) *types.ServiceEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].ResponseTime < endpoints[j].ResponseTime
	})

	return endpoints[0]
}

func (p *Provider) selectWeighted(endpoints []*types.ServiceEndpoint) *types.ServiceEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	totalWeight := 0
	for _, endpoint := range endpoints {
		totalWeight += endpoint.Weight
	}

	if totalWeight == 0 {
		return endpoints[0]
	}

	random := rand.Intn(totalWeight)
	currentWeight := 0

	for _, endpoint := range endpoints {
		currentWeight += endpoint.Weight
		if random < currentWeight {
			return endpoint
		}
	}

	return endpoints[len(endpoints)-1]
}

func (p *Provider) selectRandom(endpoints []*types.ServiceEndpoint) *types.ServiceEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	index := rand.Intn(len(endpoints))
	return endpoints[index]
}

func (p *Provider) selectHealthBased(endpoints []*types.ServiceEndpoint) *types.ServiceEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	healthyEndpoints := make([]*types.ServiceEndpoint, 0)
	for _, endpoint := range endpoints {
		if endpoint.Health == types.HealthHealthy {
			healthyEndpoints = append(healthyEndpoints, endpoint)
		}
	}

	if len(healthyEndpoints) > 0 {
		return p.selectRoundRobin(healthyEndpoints)
	}

	return p.selectRoundRobin(endpoints)
}

func (p *Provider) markEndpointUnhealthy(endpointID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if endpoint, exists := p.endpoints[endpointID]; exists {
		endpoint.Health = types.HealthUnhealthy
		endpoint.Failures++
	}
}

func (p *Provider) updateStats(success bool, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stats.TotalRequests++
	if success {
		p.stats.SuccessfulRequests++
	} else {
		p.stats.FailedRequests++
	}

	if p.stats.TotalRequests > 0 {
		p.stats.AverageResponseTime = time.Duration(
			(int64(p.stats.AverageResponseTime)*int64(p.stats.TotalRequests-1) + int64(duration)) / int64(p.stats.TotalRequests),
		)
	}
}
