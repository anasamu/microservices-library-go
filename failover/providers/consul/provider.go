package consul

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/failover/types"
	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

// Provider implements failover using Consul
type Provider struct {
	name      string
	client    *api.Client
	config    *Config
	logger    *logrus.Logger
	endpoints map[string]*types.ServiceEndpoint
	stats     *types.FailoverStats
	events    []*types.FailoverEvent
	mu        sync.RWMutex
	connected bool
	startTime time.Time
}

// Config holds Consul provider configuration
type Config struct {
	Address    string            `json:"address"`
	Token      string            `json:"token"`
	Datacenter string            `json:"datacenter"`
	Namespace  string            `json:"namespace"`
	Timeout    time.Duration     `json:"timeout"`
	Metadata   map[string]string `json:"metadata"`
}

// NewProvider creates a new Consul failover provider
func NewProvider(config *Config, logger *logrus.Logger) (*Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if logger == nil {
		logger = logrus.New()
	}

	// Create Consul client configuration
	consulConfig := api.DefaultConfig()
	if config.Address != "" {
		consulConfig.Address = config.Address
	}
	if config.Token != "" {
		consulConfig.Token = config.Token
	}
	if config.Datacenter != "" {
		consulConfig.Datacenter = config.Datacenter
	}
	if config.Namespace != "" {
		consulConfig.Namespace = config.Namespace
	}

	// Create Consul client
	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consul client: %w", err)
	}

	provider := &Provider{
		name:      "consul",
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
		Host:     p.config.Address,
		Port:     8500, // Default Consul port
		Status:   status,
		Metadata: p.config.Metadata,
	}
}

// Connect connects to Consul
func (p *Provider) Connect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Test connection by getting agent info
	agent := p.client.Agent()
	_, err := agent.Self()
	if err != nil {
		p.connected = false
		return fmt.Errorf("failed to connect to Consul: %w", err)
	}

	p.connected = true
	p.logger.Info("Connected to Consul")
	return nil
}

// Disconnect disconnects from Consul
func (p *Provider) Disconnect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.connected = false
	p.logger.Info("Disconnected from Consul")
	return nil
}

// Ping tests the connection to Consul
func (p *Provider) Ping(ctx context.Context) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to Consul")
	}

	agent := p.client.Agent()
	_, err := agent.Self()
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

// RegisterEndpoint registers an endpoint in Consul
func (p *Provider) RegisterEndpoint(ctx context.Context, endpoint *types.ServiceEndpoint) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to Consul")
	}

	// Create Consul service registration
	registration := &api.AgentServiceRegistration{
		ID:      endpoint.ID,
		Name:    endpoint.Name,
		Address: endpoint.Address,
		Port:    endpoint.Port,
		Tags:    endpoint.Tags,
		Meta:    endpoint.Metadata,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", endpoint.Address, endpoint.Port),
			Interval:                       "30s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "60s",
		},
	}

	// Register service with Consul
	agent := p.client.Agent()
	err := agent.ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service in Consul: %w", err)
	}

	// Store endpoint locally
	p.mu.Lock()
	p.endpoints[endpoint.ID] = endpoint
	p.mu.Unlock()

	p.logger.WithField("endpoint", endpoint.ID).Info("Endpoint registered in Consul")
	return nil
}

// DeregisterEndpoint deregisters an endpoint from Consul
func (p *Provider) DeregisterEndpoint(ctx context.Context, endpointID string) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to Consul")
	}

	// Deregister service from Consul
	agent := p.client.Agent()
	err := agent.ServiceDeregister(endpointID)
	if err != nil {
		return fmt.Errorf("failed to deregister service from Consul: %w", err)
	}

	// Remove endpoint locally
	p.mu.Lock()
	delete(p.endpoints, endpointID)
	p.mu.Unlock()

	p.logger.WithField("endpoint", endpointID).Info("Endpoint deregistered from Consul")
	return nil
}

// UpdateEndpoint updates an endpoint in Consul
func (p *Provider) UpdateEndpoint(ctx context.Context, endpoint *types.ServiceEndpoint) error {
	// Deregister and re-register to update
	err := p.DeregisterEndpoint(ctx, endpoint.ID)
	if err != nil {
		return err
	}

	return p.RegisterEndpoint(ctx, endpoint)
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

	p.mu.RLock()
	availableEndpoints := make([]*types.ServiceEndpoint, 0)
	for _, endpoint := range p.endpoints {
		if endpoint.Health == types.HealthHealthy || endpoint.Health == types.HealthDegraded {
			availableEndpoints = append(availableEndpoints, endpoint)
		}
	}
	p.mu.RUnlock()

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
		return types.HealthUnknown, fmt.Errorf("not connected to Consul")
	}

	// Get service health from Consul
	health := p.client.Health()
	checks, _, err := health.Service(endpointID, "", false, nil)
	if err != nil {
		return types.HealthUnknown, fmt.Errorf("failed to get health check: %w", err)
	}

	if len(checks) == 0 {
		return types.HealthUnknown, fmt.Errorf("no health checks found for endpoint %s", endpointID)
	}

	// Check if all checks are passing
	allPassing := true
	for _, check := range checks {
		if check.Checks.AggregatedStatus() != api.HealthPassing {
			allPassing = false
			break
		}
	}

	if allPassing {
		return types.HealthHealthy, nil
	}

	return types.HealthUnhealthy, nil
}

// HealthCheckAll performs health check on all endpoints
func (p *Provider) HealthCheckAll(ctx context.Context) (map[string]types.HealthStatus, error) {
	statuses := make(map[string]types.HealthStatus)

	p.mu.RLock()
	endpointIDs := make([]string, 0, len(p.endpoints))
	for id := range p.endpoints {
		endpointIDs = append(endpointIDs, id)
	}
	p.mu.RUnlock()

	for _, endpointID := range endpointIDs {
		status, err := p.HealthCheck(ctx, endpointID)
		if err != nil {
			status = types.HealthUnknown
		}
		statuses[endpointID] = status
	}

	return statuses, nil
}

// Configure configures the failover provider
func (p *Provider) Configure(ctx context.Context, config *types.FailoverConfig) error {
	// Store configuration
	p.mu.Lock()
	defer p.mu.Unlock()

	// Configuration is stored in the config parameter
	// In a real implementation, you might want to store this
	p.logger.WithField("config", config.Name).Info("Failover configuration updated")
	return nil
}

// GetConfig returns the current configuration
func (p *Provider) GetConfig(ctx context.Context) (*types.FailoverConfig, error) {
	// Return default configuration
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

// Helper methods for endpoint selection strategies

func (p *Provider) selectRoundRobin(endpoints []*types.ServiceEndpoint) *types.ServiceEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	// Simple round-robin based on current time
	index := int(time.Now().UnixNano()) % len(endpoints)
	return endpoints[index]
}

func (p *Provider) selectLeastConnections(endpoints []*types.ServiceEndpoint) *types.ServiceEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	// Sort by response time (proxy for connections)
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].ResponseTime < endpoints[j].ResponseTime
	})

	return endpoints[0]
}

func (p *Provider) selectWeighted(endpoints []*types.ServiceEndpoint) *types.ServiceEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	// Calculate total weight
	totalWeight := 0
	for _, endpoint := range endpoints {
		totalWeight += endpoint.Weight
	}

	if totalWeight == 0 {
		return endpoints[0]
	}

	// Select based on weight
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

	// Prefer healthy endpoints
	healthyEndpoints := make([]*types.ServiceEndpoint, 0)
	for _, endpoint := range endpoints {
		if endpoint.Health == types.HealthHealthy {
			healthyEndpoints = append(healthyEndpoints, endpoint)
		}
	}

	if len(healthyEndpoints) > 0 {
		return p.selectRoundRobin(healthyEndpoints)
	}

	// Fall back to degraded endpoints
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

	// Update average response time
	if p.stats.TotalRequests > 0 {
		p.stats.AverageResponseTime = time.Duration(
			(int64(p.stats.AverageResponseTime)*int64(p.stats.TotalRequests-1) + int64(duration)) / int64(p.stats.TotalRequests),
		)
	}
}
