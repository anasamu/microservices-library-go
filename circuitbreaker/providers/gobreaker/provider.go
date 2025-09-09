package gobreaker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/circuitbreaker/types"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
)

// GobreakerProvider implements the CircuitBreakerProvider interface using sony/gobreaker
type GobreakerProvider struct {
	breakers  map[string]*gobreaker.CircuitBreaker
	configs   map[string]*types.CircuitBreakerConfig
	mu        sync.RWMutex
	logger    *logrus.Logger
	connected bool
}

// GobreakerConfig holds gobreaker-specific configuration
type GobreakerConfig struct {
	DefaultMaxRequests uint32        `json:"default_max_requests"`
	DefaultInterval    time.Duration `json:"default_interval"`
	DefaultTimeout     time.Duration `json:"default_timeout"`
	KeyPrefix          string        `json:"key_prefix"`
	Namespace          string        `json:"namespace"`
}

// NewGobreakerProvider creates a new gobreaker circuit breaker provider
func NewGobreakerProvider(config *GobreakerConfig, logger *logrus.Logger) *GobreakerProvider {
	if config == nil {
		config = &GobreakerConfig{
			DefaultMaxRequests: 10,
			DefaultInterval:    10 * time.Second,
			DefaultTimeout:     60 * time.Second,
		}
	}

	return &GobreakerProvider{
		breakers:  make(map[string]*gobreaker.CircuitBreaker),
		configs:   make(map[string]*types.CircuitBreakerConfig),
		logger:    logger,
		connected: false,
	}
}

// GetName returns the provider name
func (g *GobreakerProvider) GetName() string {
	return "gobreaker"
}

// GetSupportedFeatures returns the features supported by gobreaker
func (g *GobreakerProvider) GetSupportedFeatures() []types.CircuitBreakerFeature {
	return []types.CircuitBreakerFeature{
		types.FeatureExecute,
		types.FeatureState,
		types.FeatureMetrics,
		types.FeatureFallback,
		types.FeatureTimeout,
		types.FeatureRetry,
		types.FeatureCustomState,
		types.FeatureMonitoring,
	}
}

// GetConnectionInfo returns connection information
func (g *GobreakerProvider) GetConnectionInfo() *types.ConnectionInfo {
	status := types.StatusDisconnected
	if g.connected {
		status = types.StatusConnected
	}

	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     0,
		Database: "gobreaker",
		Status:   status,
		Metadata: map[string]string{
			"provider": "gobreaker",
			"version":  "0.5.0",
			"breakers": fmt.Sprintf("%d", len(g.breakers)),
		},
	}
}

// Connect establishes connection (always succeeds for gobreaker)
func (g *GobreakerProvider) Connect(ctx context.Context) error {
	g.connected = true
	g.logger.Info("Connected to gobreaker circuit breaker successfully")
	return nil
}

// Disconnect closes the connection
func (g *GobreakerProvider) Disconnect(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.connected = false
	g.breakers = make(map[string]*gobreaker.CircuitBreaker)
	g.configs = make(map[string]*types.CircuitBreakerConfig)
	g.logger.Info("Disconnected from gobreaker circuit breaker")
	return nil
}

// Ping tests the connection (always succeeds for gobreaker)
func (g *GobreakerProvider) Ping(ctx context.Context) error {
	if !g.connected {
		return fmt.Errorf("gobreaker circuit breaker not connected")
	}
	return nil
}

// IsConnected returns the connection status
func (g *GobreakerProvider) IsConnected() bool {
	return g.connected
}

// Execute runs a function through the circuit breaker
func (g *GobreakerProvider) Execute(ctx context.Context, name string, fn func() (interface{}, error)) (*types.ExecutionResult, error) {
	if !g.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Gobreaker circuit breaker not connected"}
	}

	breaker := g.getOrCreateBreaker(name)
	start := time.Now()

	result, err := breaker.Execute(func() (interface{}, error) {
		return fn()
	})

	duration := time.Since(start)

	executionResult := &types.ExecutionResult{
		Result:   result,
		Error:    err,
		Duration: duration,
		State:    g.mapGobreakerState(breaker.State()),
		Fallback: false,
		Retry:    false,
		Attempts: 1,
	}

	g.logger.WithFields(logrus.Fields{
		"name":     name,
		"state":    executionResult.State,
		"duration": duration,
		"error":    err != nil,
	}).Debug("Circuit breaker execution completed")

	return executionResult, nil
}

// ExecuteWithFallback runs a function with fallback through the circuit breaker
func (g *GobreakerProvider) ExecuteWithFallback(ctx context.Context, name string, fn func() (interface{}, error), fallback func() (interface{}, error)) (*types.ExecutionResult, error) {
	if !g.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Gobreaker circuit breaker not connected"}
	}

	breaker := g.getOrCreateBreaker(name)
	start := time.Now()

	result, err := breaker.Execute(func() (interface{}, error) {
		return fn()
	})

	// If the main function fails, try fallback
	if err != nil {
		g.logger.WithError(err).WithField("name", name).Debug("Main function failed, trying fallback")

		fallbackResult, fallbackErr := fallback()
		if fallbackErr != nil {
			g.logger.WithError(fallbackErr).WithField("name", name).Error("Fallback function also failed")
			return &types.ExecutionResult{
				Result:   nil,
				Error:    fallbackErr,
				Duration: time.Since(start),
				State:    g.mapGobreakerState(breaker.State()),
				Fallback: true,
				Retry:    false,
				Attempts: 1,
			}, fallbackErr
		}

		return &types.ExecutionResult{
			Result:   fallbackResult,
			Error:    nil,
			Duration: time.Since(start),
			State:    g.mapGobreakerState(breaker.State()),
			Fallback: true,
			Retry:    false,
			Attempts: 1,
		}, nil
	}

	duration := time.Since(start)

	executionResult := &types.ExecutionResult{
		Result:   result,
		Error:    nil,
		Duration: duration,
		State:    g.mapGobreakerState(breaker.State()),
		Fallback: false,
		Retry:    false,
		Attempts: 1,
	}

	g.logger.WithFields(logrus.Fields{
		"name":     name,
		"state":    executionResult.State,
		"duration": duration,
		"fallback": false,
	}).Debug("Circuit breaker execution with fallback completed")

	return executionResult, nil
}

// GetState returns the state of a circuit breaker
func (g *GobreakerProvider) GetState(ctx context.Context, name string) (types.CircuitState, error) {
	if !g.IsConnected() {
		return "", &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Gobreaker circuit breaker not connected"}
	}

	breaker := g.getBreaker(name)
	if breaker == nil {
		return types.StateClosed, nil // Default state for non-existent breakers
	}

	return g.mapGobreakerState(breaker.State()), nil
}

// GetStats returns circuit breaker statistics
func (g *GobreakerProvider) GetStats(ctx context.Context, name string) (*types.CircuitBreakerStats, error) {
	if !g.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Gobreaker circuit breaker not connected"}
	}

	breaker := g.getBreaker(name)
	if breaker == nil {
		return &types.CircuitBreakerStats{
			Provider:   g.GetName(),
			State:      types.StateClosed,
			LastUpdate: time.Now(),
		}, nil
	}

	counts := breaker.Counts()
	state := breaker.State()

	stats := &types.CircuitBreakerStats{
		Requests:             int64(counts.Requests),
		Successes:            int64(counts.TotalSuccesses),
		Failures:             int64(counts.TotalFailures),
		Timeouts:             0, // Gobreaker doesn't track timeouts separately
		Rejects:              int64(counts.Requests - counts.TotalSuccesses - counts.TotalFailures),
		State:                g.mapGobreakerState(state),
		LastFailureTime:      counts.LastFailureTime,
		LastSuccessTime:      counts.LastSuccessTime,
		ConsecutiveFailures:  int64(counts.ConsecutiveFailures),
		ConsecutiveSuccesses: int64(counts.ConsecutiveSuccesses),
		Uptime:               time.Since(counts.LastSuccessTime),
		LastUpdate:           time.Now(),
		Provider:             g.GetName(),
	}

	return stats, nil
}

// Reset resets a circuit breaker
func (g *GobreakerProvider) Reset(ctx context.Context, name string) error {
	if !g.IsConnected() {
		return &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Gobreaker circuit breaker not connected"}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if breaker, exists := g.breakers[name]; exists {
		// Gobreaker doesn't have a direct reset method, so we recreate it
		delete(g.breakers, name)
		if config, configExists := g.configs[name]; configExists {
			g.createBreaker(name, config)
		}
		g.logger.WithField("name", name).Info("Circuit breaker reset")
	}

	return nil
}

// Configure configures a circuit breaker
func (g *GobreakerProvider) Configure(ctx context.Context, name string, config *types.CircuitBreakerConfig) error {
	if !g.IsConnected() {
		return &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Gobreaker circuit breaker not connected"}
	}

	if config == nil {
		return &types.CircuitBreakerError{Code: types.ErrCodeInvalidConfig, Message: "Configuration cannot be nil"}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Store the configuration
	g.configs[name] = config

	// Create or recreate the breaker with new configuration
	g.createBreaker(name, config)

	g.logger.WithFields(logrus.Fields{
		"name":              name,
		"max_requests":      config.MaxRequests,
		"interval":          config.Interval,
		"timeout":           config.Timeout,
		"failure_threshold": config.FailureThreshold,
	}).Info("Circuit breaker configured")

	return nil
}

// GetConfig returns circuit breaker configuration
func (g *GobreakerProvider) GetConfig(ctx context.Context, name string) (*types.CircuitBreakerConfig, error) {
	if !g.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Gobreaker circuit breaker not connected"}
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	config, exists := g.configs[name]
	if !exists {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeInvalidConfig, Message: "Configuration not found"}
	}

	return config, nil
}

// GetAllStates returns states of all circuit breakers
func (g *GobreakerProvider) GetAllStates(ctx context.Context) (map[string]types.CircuitState, error) {
	if !g.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Gobreaker circuit breaker not connected"}
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	states := make(map[string]types.CircuitState)
	for name, breaker := range g.breakers {
		states[name] = g.mapGobreakerState(breaker.State())
	}

	return states, nil
}

// GetAllStats returns statistics of all circuit breakers
func (g *GobreakerProvider) GetAllStats(ctx context.Context) (map[string]*types.CircuitBreakerStats, error) {
	if !g.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Gobreaker circuit breaker not connected"}
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	stats := make(map[string]*types.CircuitBreakerStats)
	for name, breaker := range g.breakers {
		counts := breaker.Counts()
		state := breaker.State()

		stats[name] = &types.CircuitBreakerStats{
			Requests:             int64(counts.Requests),
			Successes:            int64(counts.TotalSuccesses),
			Failures:             int64(counts.TotalFailures),
			Timeouts:             0,
			Rejects:              int64(counts.Requests - counts.TotalSuccesses - counts.TotalFailures),
			State:                g.mapGobreakerState(state),
			LastFailureTime:      counts.LastFailureTime,
			LastSuccessTime:      counts.LastSuccessTime,
			ConsecutiveFailures:  int64(counts.ConsecutiveFailures),
			ConsecutiveSuccesses: int64(counts.ConsecutiveSuccesses),
			Uptime:               time.Since(counts.LastSuccessTime),
			LastUpdate:           time.Now(),
			Provider:             g.GetName(),
		}
	}

	return stats, nil
}

// ResetAll resets all circuit breakers
func (g *GobreakerProvider) ResetAll(ctx context.Context) error {
	if !g.IsConnected() {
		return &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Gobreaker circuit breaker not connected"}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.breakers = make(map[string]*gobreaker.CircuitBreaker)

	// Recreate all breakers with their configurations
	for name, config := range g.configs {
		_ = g.createBreaker(name, config)
	}

	g.logger.Info("All circuit breakers reset")
	return nil
}

// Close closes the provider
func (g *GobreakerProvider) Close(ctx context.Context) error {
	return g.Disconnect(ctx)
}

// getOrCreateBreaker gets an existing breaker or creates a new one
func (g *GobreakerProvider) getOrCreateBreaker(name string) *gobreaker.CircuitBreaker {
	g.mu.RLock()
	breaker, exists := g.breakers[name]
	g.mu.RUnlock()

	if exists {
		return breaker
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Double-check after acquiring write lock
	if breaker, exists := g.breakers[name]; exists {
		return breaker
	}

	// Create with default configuration
	config := types.DefaultConfig
	config.Name = name
	g.configs[name] = config
	breaker = g.createBreaker(name, config)

	return breaker
}

// getBreaker gets an existing breaker
func (g *GobreakerProvider) getBreaker(name string) *gobreaker.CircuitBreaker {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.breakers[name]
}

// createBreaker creates a new gobreaker circuit breaker
func (g *GobreakerProvider) createBreaker(name string, config *types.CircuitBreakerConfig) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if config.ReadyToTrip != nil {
				return config.ReadyToTrip(g.mapGobreakerCounts(counts))
			}
			return counts.ConsecutiveFailures >= uint32(config.MaxConsecutiveFails)
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			if config.OnStateChange != nil {
				config.OnStateChange(name, g.mapGobreakerState(from), g.mapGobreakerState(to))
			}
			g.logger.WithFields(logrus.Fields{
				"name": name,
				"from": from,
				"to":   to,
			}).Info("Circuit breaker state changed")
		},
		IsSuccessful: config.IsSuccessful,
	}

	breaker := gobreaker.NewCircuitBreaker(settings)
	g.breakers[name] = breaker

	return breaker
}

// mapGobreakerState maps gobreaker state to our state type
func (g *GobreakerProvider) mapGobreakerState(state gobreaker.State) types.CircuitState {
	switch state {
	case gobreaker.StateClosed:
		return types.StateClosed
	case gobreaker.StateOpen:
		return types.StateOpen
	case gobreaker.StateHalfOpen:
		return types.StateHalfOpen
	default:
		return types.StateClosed
	}
}

// mapGobreakerCounts maps gobreaker counts to our counts type
func (g *GobreakerProvider) mapGobreakerCounts(counts gobreaker.Counts) types.Counts {
	return types.Counts{
		Requests:             counts.Requests,
		TotalSuccesses:       counts.TotalSuccesses,
		TotalFailures:        counts.TotalFailures,
		ConsecutiveSuccesses: counts.ConsecutiveSuccesses,
		ConsecutiveFailures:  counts.ConsecutiveFailures,
		LastFailureTime:      counts.LastFailureTime,
		LastSuccessTime:      counts.LastSuccessTime,
	}
}
