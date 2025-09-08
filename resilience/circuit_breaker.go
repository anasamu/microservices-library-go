package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
)

// CircuitBreakerManager manages circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*gobreaker.CircuitBreaker
	configs  map[string]*CircuitBreakerConfig
	mutex    sync.RWMutex
	logger   *logrus.Logger
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Name          string
	MaxRequests   uint32
	Interval      time.Duration
	Timeout       time.Duration
	ReadyToTrip   func(counts gobreaker.Counts) bool
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState struct {
	Name      string           `json:"name"`
	State     gobreaker.State  `json:"state"`
	Counts    gobreaker.Counts `json:"counts"`
	Expiry    time.Time        `json:"expiry"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// CircuitBreakerEvent represents a circuit breaker event
type CircuitBreakerEvent struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	Event     string                 `json:"event"`
	FromState gobreaker.State        `json:"from_state"`
	ToState   gobreaker.State        `json:"to_state"`
	Counts    gobreaker.Counts       `json:"counts"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(logger *logrus.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		configs:  make(map[string]*CircuitBreakerConfig),
		logger:   logger,
	}
}

// CreateCircuitBreaker creates a new circuit breaker
func (cbm *CircuitBreakerManager) CreateCircuitBreaker(config *CircuitBreakerConfig) error {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	// Set default values
	if config.MaxRequests == 0 {
		config.MaxRequests = 3
	}
	if config.Interval == 0 {
		config.Interval = 10 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.ReadyToTrip == nil {
		config.ReadyToTrip = func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		}
	}
	if config.OnStateChange == nil {
		config.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
			cbm.logger.WithFields(logrus.Fields{
				"circuit_breaker": name,
				"from_state":      from.String(),
				"to_state":        to.String(),
			}).Info("Circuit breaker state changed")
		}
	}

	// Create circuit breaker
	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          config.Name,
		MaxRequests:   config.MaxRequests,
		Interval:      config.Interval,
		Timeout:       config.Timeout,
		ReadyToTrip:   config.ReadyToTrip,
		OnStateChange: config.OnStateChange,
	})

	cbm.breakers[config.Name] = breaker
	cbm.configs[config.Name] = config

	cbm.logger.WithFields(logrus.Fields{
		"name":         config.Name,
		"max_requests": config.MaxRequests,
		"interval":     config.Interval,
		"timeout":      config.Timeout,
	}).Info("Circuit breaker created successfully")

	return nil
}

// Execute executes a function with circuit breaker protection
func (cbm *CircuitBreakerManager) Execute(ctx context.Context, name string, fn func() (interface{}, error)) (interface{}, error) {
	cbm.mutex.RLock()
	breaker, exists := cbm.breakers[name]
	cbm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("circuit breaker not found: %s", name)
	}

	result, err := breaker.Execute(func() (interface{}, error) {
		return fn()
	})

	if err != nil {
		cbm.logEvent(name, "execution_failed", gobreaker.StateClosed, gobreaker.StateClosed, breaker.Counts(), err)
		return nil, err
	}

	cbm.logEvent(name, "execution_success", gobreaker.StateClosed, gobreaker.StateClosed, breaker.Counts(), nil)
	return result, nil
}

// ExecuteAsync executes a function asynchronously with circuit breaker protection
func (cbm *CircuitBreakerManager) ExecuteAsync(ctx context.Context, name string, fn func() (interface{}, error)) <-chan CircuitBreakerResult {
	result := make(chan CircuitBreakerResult, 1)

	go func() {
		defer close(result)
		value, err := cbm.Execute(ctx, name, fn)
		result <- CircuitBreakerResult{
			Value: value,
			Error: err,
		}
	}()

	return result
}

// GetState returns the state of a circuit breaker
func (cbm *CircuitBreakerManager) GetState(name string) (*CircuitBreakerState, error) {
	cbm.mutex.RLock()
	breaker, exists := cbm.breakers[name]
	config, configExists := cbm.configs[name]
	cbm.mutex.RUnlock()

	if !exists || !configExists {
		return nil, fmt.Errorf("circuit breaker not found: %s", name)
	}

	state := breaker.State()
	counts := breaker.Counts()

	return &CircuitBreakerState{
		Name:      name,
		State:     state,
		Counts:    counts,
		Expiry:    time.Now().Add(config.Timeout),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// GetAllStates returns the states of all circuit breakers
func (cbm *CircuitBreakerManager) GetAllStates() map[string]*CircuitBreakerState {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	states := make(map[string]*CircuitBreakerState)
	for name := range cbm.breakers {
		if state, err := cbm.GetState(name); err == nil {
			states[name] = state
		}
	}

	return states
}

// Reset resets a circuit breaker
func (cbm *CircuitBreakerManager) Reset(name string) error {
	cbm.mutex.RLock()
	breaker, exists := cbm.breakers[name]
	cbm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("circuit breaker not found: %s", name)
	}

	// Reset the circuit breaker by creating a new one
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	config := cbm.configs[name]
	breaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          config.Name,
		MaxRequests:   config.MaxRequests,
		Interval:      config.Interval,
		Timeout:       config.Timeout,
		ReadyToTrip:   config.ReadyToTrip,
		OnStateChange: config.OnStateChange,
	})

	cbm.breakers[name] = breaker

	cbm.logger.WithField("name", name).Info("Circuit breaker reset successfully")
	return nil
}

// DeleteCircuitBreaker deletes a circuit breaker
func (cbm *CircuitBreakerManager) DeleteCircuitBreaker(name string) error {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	if _, exists := cbm.breakers[name]; !exists {
		return fmt.Errorf("circuit breaker not found: %s", name)
	}

	delete(cbm.breakers, name)
	delete(cbm.configs, name)

	cbm.logger.WithField("name", name).Info("Circuit breaker deleted successfully")
	return nil
}

// logEvent logs a circuit breaker event
func (cbm *CircuitBreakerManager) logEvent(name, event string, fromState, toState gobreaker.State, counts gobreaker.Counts, err error) {
	eventData := &CircuitBreakerEvent{
		ID:        uuid.New(),
		Name:      name,
		Event:     event,
		FromState: fromState,
		ToState:   toState,
		Counts:    counts,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	if err != nil {
		eventData.Error = err.Error()
	}

	cbm.logger.WithFields(logrus.Fields{
		"event_id":   eventData.ID,
		"name":       name,
		"event":      event,
		"from_state": fromState.String(),
		"to_state":   toState.String(),
		"counts":     counts,
		"error":      err,
	}).Debug("Circuit breaker event")

	// Here you could send the event to a message queue or store it in a database
}

// CircuitBreakerResult represents the result of an async circuit breaker execution
type CircuitBreakerResult struct {
	Value interface{}
	Error error
}

// DefaultCircuitBreakerConfig returns a default circuit breaker configuration
func DefaultCircuitBreakerConfig(name string) *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		Name:        name,
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Default implementation - can be overridden
		},
	}
}

// CreateHTTPCircuitBreaker creates a circuit breaker specifically for HTTP requests
func (cbm *CircuitBreakerManager) CreateHTTPCircuitBreaker(name string, maxRequests uint32, interval, timeout time.Duration) error {
	config := &CircuitBreakerConfig{
		Name:        name,
		MaxRequests: maxRequests,
		Interval:    interval,
		Timeout:     timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			cbm.logger.WithFields(logrus.Fields{
				"circuit_breaker": name,
				"from_state":      from.String(),
				"to_state":        to.String(),
			}).Info("HTTP circuit breaker state changed")
		},
	}

	return cbm.CreateCircuitBreaker(config)
}

// CreateDatabaseCircuitBreaker creates a circuit breaker specifically for database operations
func (cbm *CircuitBreakerManager) CreateDatabaseCircuitBreaker(name string, maxRequests uint32, interval, timeout time.Duration) error {
	config := &CircuitBreakerConfig{
		Name:        name,
		MaxRequests: maxRequests,
		Interval:    interval,
		Timeout:     timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			cbm.logger.WithFields(logrus.Fields{
				"circuit_breaker": name,
				"from_state":      from.String(),
				"to_state":        to.State.String(),
			}).Info("Database circuit breaker state changed")
		},
	}

	return cbm.CreateCircuitBreaker(config)
}

// CreateExternalServiceCircuitBreaker creates a circuit breaker for external service calls
func (cbm *CircuitBreakerManager) CreateExternalServiceCircuitBreaker(name string, maxRequests uint32, interval, timeout time.Duration) error {
	config := &CircuitBreakerConfig{
		Name:        name,
		MaxRequests: maxRequests,
		Interval:    interval,
		Timeout:     timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			cbm.logger.WithFields(logrus.Fields{
				"circuit_breaker": name,
				"from_state":      from.String(),
				"to_state":        to.String(),
			}).Info("External service circuit breaker state changed")
		},
	}

	return cbm.CreateCircuitBreaker(config)
}
