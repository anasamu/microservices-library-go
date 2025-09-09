package custom

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/circuitbreaker/types"
	"github.com/sirupsen/logrus"
)

// CustomProvider implements a custom circuit breaker provider
type CustomProvider struct {
	breakers  map[string]*CustomCircuitBreaker
	configs   map[string]*types.CircuitBreakerConfig
	mu        sync.RWMutex
	logger    *logrus.Logger
	connected bool
}

// CustomCircuitBreaker represents a custom circuit breaker implementation
type CustomCircuitBreaker struct {
	name        string
	config      *types.CircuitBreakerConfig
	state       types.CircuitState
	counts      types.Counts
	lastFailure time.Time
	lastSuccess time.Time
	mu          sync.RWMutex
	logger      *logrus.Logger
}

// CustomConfig holds custom provider-specific configuration
type CustomConfig struct {
	DefaultMaxRequests uint32        `json:"default_max_requests"`
	DefaultInterval    time.Duration `json:"default_interval"`
	DefaultTimeout     time.Duration `json:"default_timeout"`
	KeyPrefix          string        `json:"key_prefix"`
	Namespace          string        `json:"namespace"`
	EnableMetrics      bool          `json:"enable_metrics"`
	EnableLogging      bool          `json:"enable_logging"`
}

// NewCustomProvider creates a new custom circuit breaker provider
func NewCustomProvider(config *CustomConfig, logger *logrus.Logger) *CustomProvider {
	if config == nil {
		config = &CustomConfig{
			DefaultMaxRequests: 10,
			DefaultInterval:    10 * time.Second,
			DefaultTimeout:     60 * time.Second,
			EnableMetrics:      true,
			EnableLogging:      true,
		}
	}

	return &CustomProvider{
		breakers:  make(map[string]*CustomCircuitBreaker),
		configs:   make(map[string]*types.CircuitBreakerConfig),
		logger:    logger,
		connected: false,
	}
}

// GetName returns the provider name
func (c *CustomProvider) GetName() string {
	return "custom"
}

// GetSupportedFeatures returns the features supported by custom provider
func (c *CustomProvider) GetSupportedFeatures() []types.CircuitBreakerFeature {
	return []types.CircuitBreakerFeature{
		types.FeatureExecute,
		types.FeatureState,
		types.FeatureMetrics,
		types.FeatureFallback,
		types.FeatureTimeout,
		types.FeatureRetry,
		types.FeatureCustomState,
		types.FeatureBulkhead,
		types.FeatureRateLimit,
		types.FeatureHealthCheck,
		types.FeatureMonitoring,
	}
}

// GetConnectionInfo returns connection information
func (c *CustomProvider) GetConnectionInfo() *types.ConnectionInfo {
	status := types.StatusDisconnected
	if c.connected {
		status = types.StatusConnected
	}

	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     0,
		Database: "custom",
		Status:   status,
		Metadata: map[string]string{
			"provider": "custom",
			"version":  "1.0.0",
			"breakers": fmt.Sprintf("%d", len(c.breakers)),
			"metrics":  "enabled",
		},
	}
}

// Connect establishes connection (always succeeds for custom provider)
func (c *CustomProvider) Connect(ctx context.Context) error {
	c.connected = true
	c.logger.Info("Connected to custom circuit breaker successfully")
	return nil
}

// Disconnect closes the connection
func (c *CustomProvider) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false
	c.breakers = make(map[string]*CustomCircuitBreaker)
	c.configs = make(map[string]*types.CircuitBreakerConfig)
	c.logger.Info("Disconnected from custom circuit breaker")
	return nil
}

// Ping tests the connection (always succeeds for custom provider)
func (c *CustomProvider) Ping(ctx context.Context) error {
	if !c.connected {
		return fmt.Errorf("custom circuit breaker not connected")
	}
	return nil
}

// IsConnected returns the connection status
func (c *CustomProvider) IsConnected() bool {
	return c.connected
}

// Execute runs a function through the circuit breaker
func (c *CustomProvider) Execute(ctx context.Context, name string, fn func() (interface{}, error)) (*types.ExecutionResult, error) {
	if !c.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Custom circuit breaker not connected"}
	}

	breaker := c.getOrCreateBreaker(name)
	start := time.Now()

	result, err := breaker.Execute(ctx, fn)

	duration := time.Since(start)

	executionResult := &types.ExecutionResult{
		Result:   result,
		Error:    err,
		Duration: duration,
		State:    breaker.GetState(),
		Fallback: false,
		Retry:    false,
		Attempts: 1,
	}

	c.logger.WithFields(logrus.Fields{
		"name":     name,
		"state":    executionResult.State,
		"duration": duration,
		"error":    err != nil,
	}).Debug("Custom circuit breaker execution completed")

	return executionResult, nil
}

// ExecuteWithFallback runs a function with fallback through the circuit breaker
func (c *CustomProvider) ExecuteWithFallback(ctx context.Context, name string, fn func() (interface{}, error), fallback func() (interface{}, error)) (*types.ExecutionResult, error) {
	if !c.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Custom circuit breaker not connected"}
	}

	breaker := c.getOrCreateBreaker(name)
	start := time.Now()

	result, err := breaker.Execute(ctx, fn)

	// If the main function fails, try fallback
	if err != nil {
		c.logger.WithError(err).WithField("name", name).Debug("Main function failed, trying fallback")

		fallbackResult, fallbackErr := fallback()
		if fallbackErr != nil {
			c.logger.WithError(fallbackErr).WithField("name", name).Error("Fallback function also failed")
			return &types.ExecutionResult{
				Result:   nil,
				Error:    fallbackErr,
				Duration: time.Since(start),
				State:    breaker.GetState(),
				Fallback: true,
				Retry:    false,
				Attempts: 1,
			}, fallbackErr
		}

		return &types.ExecutionResult{
			Result:   fallbackResult,
			Error:    nil,
			Duration: time.Since(start),
			State:    breaker.GetState(),
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
		State:    breaker.GetState(),
		Fallback: false,
		Retry:    false,
		Attempts: 1,
	}

	c.logger.WithFields(logrus.Fields{
		"name":     name,
		"state":    executionResult.State,
		"duration": duration,
		"fallback": false,
	}).Debug("Custom circuit breaker execution with fallback completed")

	return executionResult, nil
}

// GetState returns the state of a circuit breaker
func (c *CustomProvider) GetState(ctx context.Context, name string) (types.CircuitState, error) {
	if !c.IsConnected() {
		return "", &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Custom circuit breaker not connected"}
	}

	breaker := c.getBreaker(name)
	if breaker == nil {
		return types.StateClosed, nil // Default state for non-existent breakers
	}

	return breaker.GetState(), nil
}

// GetStats returns circuit breaker statistics
func (c *CustomProvider) GetStats(ctx context.Context, name string) (*types.CircuitBreakerStats, error) {
	if !c.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Custom circuit breaker not connected"}
	}

	breaker := c.getBreaker(name)
	if breaker == nil {
		return &types.CircuitBreakerStats{
			Provider:   c.GetName(),
			State:      types.StateClosed,
			LastUpdate: time.Now(),
		}, nil
	}

	return breaker.GetStats(), nil
}

// Reset resets a circuit breaker
func (c *CustomProvider) Reset(ctx context.Context, name string) error {
	if !c.IsConnected() {
		return &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Custom circuit breaker not connected"}
	}

	breaker := c.getBreaker(name)
	if breaker == nil {
		return &types.CircuitBreakerError{Code: types.ErrCodeInvalidState, Message: "Circuit breaker not found"}
	}

	breaker.Reset()
	c.logger.WithField("name", name).Info("Custom circuit breaker reset")
	return nil
}

// Configure configures a circuit breaker
func (c *CustomProvider) Configure(ctx context.Context, name string, config *types.CircuitBreakerConfig) error {
	if !c.IsConnected() {
		return &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Custom circuit breaker not connected"}
	}

	if config == nil {
		return &types.CircuitBreakerError{Code: types.ErrCodeInvalidConfig, Message: "Configuration cannot be nil"}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Store the configuration
	c.configs[name] = config

	// Create or update the breaker with new configuration
	breaker := c.getOrCreateBreakerUnsafe(name)
	breaker.UpdateConfig(config)

	c.logger.WithFields(logrus.Fields{
		"name":              name,
		"max_requests":      config.MaxRequests,
		"interval":          config.Interval,
		"timeout":           config.Timeout,
		"failure_threshold": config.FailureThreshold,
	}).Info("Custom circuit breaker configured")

	return nil
}

// GetConfig returns circuit breaker configuration
func (c *CustomProvider) GetConfig(ctx context.Context, name string) (*types.CircuitBreakerConfig, error) {
	if !c.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Custom circuit breaker not connected"}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	config, exists := c.configs[name]
	if !exists {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeInvalidConfig, Message: "Configuration not found"}
	}

	return config, nil
}

// GetAllStates returns states of all circuit breakers
func (c *CustomProvider) GetAllStates(ctx context.Context) (map[string]types.CircuitState, error) {
	if !c.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Custom circuit breaker not connected"}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	states := make(map[string]types.CircuitState)
	for name, breaker := range c.breakers {
		states[name] = breaker.GetState()
	}

	return states, nil
}

// GetAllStats returns statistics of all circuit breakers
func (c *CustomProvider) GetAllStats(ctx context.Context) (map[string]*types.CircuitBreakerStats, error) {
	if !c.IsConnected() {
		return nil, &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Custom circuit breaker not connected"}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]*types.CircuitBreakerStats)
	for name, breaker := range c.breakers {
		stats[name] = breaker.GetStats()
	}

	return stats, nil
}

// ResetAll resets all circuit breakers
func (c *CustomProvider) ResetAll(ctx context.Context) error {
	if !c.IsConnected() {
		return &types.CircuitBreakerError{Code: types.ErrCodeConnection, Message: "Custom circuit breaker not connected"}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for name, breaker := range c.breakers {
		breaker.Reset()
		c.logger.WithField("name", name).Debug("Custom circuit breaker reset")
	}

	c.logger.Info("All custom circuit breakers reset")
	return nil
}

// Close closes the provider
func (c *CustomProvider) Close(ctx context.Context) error {
	return c.Disconnect(ctx)
}

// getOrCreateBreaker gets an existing breaker or creates a new one
func (c *CustomProvider) getOrCreateBreaker(name string) *CustomCircuitBreaker {
	c.mu.RLock()
	breaker, exists := c.breakers[name]
	c.mu.RUnlock()

	if exists {
		return breaker
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.getOrCreateBreakerUnsafe(name)
}

// getOrCreateBreakerUnsafe gets an existing breaker or creates a new one (without locks)
func (c *CustomProvider) getOrCreateBreakerUnsafe(name string) *CustomCircuitBreaker {
	// Double-check after acquiring write lock
	if breaker, exists := c.breakers[name]; exists {
		return breaker
	}

	// Create with default configuration
	config := types.DefaultConfig
	config.Name = name
	c.configs[name] = config
	breaker := NewCustomCircuitBreaker(name, config, c.logger)
	c.breakers[name] = breaker

	return breaker
}

// getBreaker gets an existing breaker
func (c *CustomProvider) getBreaker(name string) *CustomCircuitBreaker {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.breakers[name]
}

// NewCustomCircuitBreaker creates a new custom circuit breaker
func NewCustomCircuitBreaker(name string, config *types.CircuitBreakerConfig, logger *logrus.Logger) *CustomCircuitBreaker {
	return &CustomCircuitBreaker{
		name:   name,
		config: config,
		state:  types.StateClosed,
		counts: types.Counts{},
		logger: logger,
	}
}

// Execute runs a function through the custom circuit breaker
func (cb *CustomCircuitBreaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit is open
	if cb.state == types.StateOpen {
		if time.Since(cb.lastFailure) < cb.config.Timeout {
			cb.counts.Requests++
			return nil, &types.CircuitBreakerError{
				Code:    types.ErrCodeCircuitOpen,
				Message: "Circuit breaker is open",
				State:   cb.state,
			}
		}
		// Timeout has passed, move to half-open
		cb.state = types.StateHalfOpen
		cb.logger.WithField("name", cb.name).Info("Circuit breaker moved to half-open state")
	}

	// Check max requests in half-open state
	if cb.state == types.StateHalfOpen && cb.counts.Requests >= cb.config.MaxRequests {
		cb.counts.Requests++
		return nil, &types.CircuitBreakerError{
			Code:    types.ErrCodeMaxRequests,
			Message: "Maximum requests exceeded in half-open state",
			State:   cb.state,
		}
	}

	// Execute the function
	cb.counts.Requests++
	result, err := fn()

	// Update counts and state based on result
	if cb.config.IsSuccessful(err) {
		cb.onSuccess()
	} else {
		cb.onFailure()
	}

	return result, err
}

// GetState returns the current state
func (cb *CustomCircuitBreaker) GetState() types.CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CustomCircuitBreaker) GetStats() *types.CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return &types.CircuitBreakerStats{
		Requests:             int64(cb.counts.Requests),
		Successes:            int64(cb.counts.TotalSuccesses),
		Failures:             int64(cb.counts.TotalFailures),
		Timeouts:             0, // Custom implementation doesn't track timeouts separately
		Rejects:              int64(cb.counts.Requests - cb.counts.TotalSuccesses - cb.counts.TotalFailures),
		State:                cb.state,
		LastFailureTime:      cb.counts.LastFailureTime,
		LastSuccessTime:      cb.counts.LastSuccessTime,
		ConsecutiveFailures:  int64(cb.counts.ConsecutiveFailures),
		ConsecutiveSuccesses: int64(cb.counts.ConsecutiveSuccesses),
		Uptime:               time.Since(cb.counts.LastSuccessTime),
		LastUpdate:           time.Now(),
		Provider:             "custom",
	}
}

// Reset resets the circuit breaker
func (cb *CustomCircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = types.StateClosed
	cb.counts = types.Counts{}
	cb.lastFailure = time.Time{}
	cb.lastSuccess = time.Time{}

	cb.logger.WithField("name", cb.name).Info("Custom circuit breaker reset")
}

// UpdateConfig updates the circuit breaker configuration
func (cb *CustomCircuitBreaker) UpdateConfig(config *types.CircuitBreakerConfig) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.config = config
	cb.logger.WithField("name", cb.name).Debug("Custom circuit breaker configuration updated")
}

// onSuccess handles successful execution
func (cb *CustomCircuitBreaker) onSuccess() {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0
	cb.lastSuccess = time.Now()
	cb.counts.LastSuccessTime = cb.lastSuccess

	// If we're in half-open state and have enough successes, close the circuit
	if cb.state == types.StateHalfOpen && cb.counts.ConsecutiveSuccesses >= cb.config.SuccessThreshold {
		oldState := cb.state
		cb.state = types.StateClosed
		cb.counts.ConsecutiveSuccesses = 0

		if cb.config.OnStateChange != nil {
			cb.config.OnStateChange(cb.name, oldState, cb.state)
		}

		cb.logger.WithFields(logrus.Fields{
			"name": cb.name,
			"from": oldState,
			"to":   cb.state,
		}).Info("Custom circuit breaker moved to closed state")
	}
}

// onFailure handles failed execution
func (cb *CustomCircuitBreaker) onFailure() {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0
	cb.lastFailure = time.Now()
	cb.counts.LastFailureTime = cb.lastFailure

	// Check if we should open the circuit
	if cb.config.ReadyToTrip(cb.counts) {
		oldState := cb.state
		cb.state = types.StateOpen

		if cb.config.OnStateChange != nil {
			cb.config.OnStateChange(cb.name, oldState, cb.state)
		}

		cb.logger.WithFields(logrus.Fields{
			"name": cb.name,
			"from": oldState,
			"to":   cb.state,
		}).Warn("Custom circuit breaker moved to open state")
	}
}
