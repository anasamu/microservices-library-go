package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// RateLimiterManager manages rate limiters
type RateLimiterManager struct {
	limiters map[string]*rate.Limiter
	configs  map[string]*RateLimitConfig
	mutex    sync.RWMutex
	logger   *logrus.Logger
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	ID          string
	Name        string
	Requests    int
	Window      time.Duration
	Burst       int
	Description string
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed    bool          `json:"allowed"`
	Remaining  int           `json:"remaining"`
	ResetTime  time.Time     `json:"reset_time"`
	RetryAfter time.Duration `json:"retry_after"`
	Limit      int           `json:"limit"`
	Window     time.Duration `json:"window"`
}

// RateLimitEvent represents a rate limit event
type RateLimitEvent struct {
	ID        uuid.UUID              `json:"id"`
	LimiterID string                 `json:"limiter_id"`
	Event     string                 `json:"event"`
	Allowed   bool                   `json:"allowed"`
	Remaining int                    `json:"remaining"`
	Limit     int                    `json:"limit"`
	Window    time.Duration          `json:"window"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// NewRateLimiterManager creates a new rate limiter manager
func NewRateLimiterManager(logger *logrus.Logger) *RateLimiterManager {
	return &RateLimiterManager{
		limiters: make(map[string]*rate.Limiter),
		configs:  make(map[string]*RateLimitConfig),
		logger:   logger,
	}
}

// CreateRateLimiter creates a new rate limiter
func (rlm *RateLimiterManager) CreateRateLimiter(config *RateLimitConfig) error {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()

	// Set default values
	if config.Burst == 0 {
		config.Burst = config.Requests
	}
	if config.Window == 0 {
		config.Window = time.Minute
	}
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}
	config.UpdatedAt = time.Now()

	// Create rate limiter
	limiter := rate.NewLimiter(rate.Every(config.Window/time.Duration(config.Requests)), config.Burst)

	rlm.limiters[config.ID] = limiter
	rlm.configs[config.ID] = config

	rlm.logger.WithFields(logrus.Fields{
		"id":       config.ID,
		"name":     config.Name,
		"requests": config.Requests,
		"window":   config.Window,
		"burst":    config.Burst,
		"enabled":  config.Enabled,
	}).Info("Rate limiter created successfully")

	return nil
}

// Allow checks if a request is allowed by the rate limiter
func (rlm *RateLimiterManager) Allow(ctx context.Context, limiterID string) (*RateLimitResult, error) {
	rlm.mutex.RLock()
	limiter, exists := rlm.limiters[limiterID]
	config, configExists := rlm.configs[limiterID]
	rlm.mutex.RUnlock()

	if !exists || !configExists {
		return nil, fmt.Errorf("rate limiter not found: %s", limiterID)
	}

	if !config.Enabled {
		return &RateLimitResult{
			Allowed:   true,
			Remaining: config.Requests,
			ResetTime: time.Now().Add(config.Window),
			Limit:     config.Requests,
			Window:    config.Window,
		}, nil
	}

	// Check if request is allowed
	allowed := limiter.Allow()
	remaining := limiter.Burst() - limiter.Tokens()
	resetTime := time.Now().Add(config.Window)

	result := &RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: resetTime,
		Limit:     config.Requests,
		Window:    config.Window,
	}

	if !allowed {
		result.RetryAfter = time.Until(resetTime)
	}

	// Log event
	rlm.logEvent(limiterID, "rate_limit_check", allowed, remaining, config.Requests, config.Window, nil)

	return result, nil
}

// AllowN checks if N requests are allowed by the rate limiter
func (rlm *RateLimiterManager) AllowN(ctx context.Context, limiterID string, n int) (*RateLimitResult, error) {
	rlm.mutex.RLock()
	limiter, exists := rlm.limiters[limiterID]
	config, configExists := rlm.configs[limiterID]
	rlm.mutex.RLock()

	if !exists || !configExists {
		return nil, fmt.Errorf("rate limiter not found: %s", limiterID)
	}

	if !config.Enabled {
		return &RateLimitResult{
			Allowed:   true,
			Remaining: config.Requests,
			ResetTime: time.Now().Add(config.Window),
			Limit:     config.Requests,
			Window:    config.Window,
		}, nil
	}

	// Check if N requests are allowed
	allowed := limiter.AllowN(time.Now(), n)
	remaining := limiter.Burst() - limiter.Tokens()
	resetTime := time.Now().Add(config.Window)

	result := &RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: resetTime,
		Limit:     config.Requests,
		Window:    config.Window,
	}

	if !allowed {
		result.RetryAfter = time.Until(resetTime)
	}

	// Log event
	rlm.logEvent(limiterID, "rate_limit_check_n", allowed, remaining, config.Requests, config.Window, map[string]interface{}{
		"requested": n,
	})

	return result, nil
}

// Wait waits for the rate limiter to allow a request
func (rlm *RateLimiterManager) Wait(ctx context.Context, limiterID string) error {
	rlm.mutex.RLock()
	limiter, exists := rlm.limiters[limiterID]
	config, configExists := rlm.configs[limiterID]
	rlm.mutex.RUnlock()

	if !exists || !configExists {
		return fmt.Errorf("rate limiter not found: %s", limiterID)
	}

	if !config.Enabled {
		return nil
	}

	return limiter.Wait(ctx)
}

// WaitN waits for the rate limiter to allow N requests
func (rlm *RateLimiterManager) WaitN(ctx context.Context, limiterID string, n int) error {
	rlm.mutex.RLock()
	limiter, exists := rlm.limiters[limiterID]
	config, configExists := rlm.configs[limiterID]
	rlm.mutex.RUnlock()

	if !exists || !configExists {
		return fmt.Errorf("rate limiter not found: %s", limiterID)
	}

	if !config.Enabled {
		return nil
	}

	return limiter.WaitN(ctx, n)
}

// Reserve reserves a token from the rate limiter
func (rlm *RateLimiterManager) Reserve(ctx context.Context, limiterID string) (*rate.Reservation, error) {
	rlm.mutex.RLock()
	limiter, exists := rlm.limiters[limiterID]
	config, configExists := rlm.configs[limiterID]
	rlm.mutex.RUnlock()

	if !exists || !configExists {
		return nil, fmt.Errorf("rate limiter not found: %s", limiterID)
	}

	if !config.Enabled {
		return &rate.Reservation{}, nil
	}

	return limiter.Reserve(), nil
}

// ReserveN reserves N tokens from the rate limiter
func (rlm *RateLimiterManager) ReserveN(ctx context.Context, limiterID string, n int) (*rate.Reservation, error) {
	rlm.mutex.RLock()
	limiter, exists := rlm.limiters[limiterID]
	config, configExists := rlm.configs[limiterID]
	rlm.mutex.RUnlock()

	if !exists || !configExists {
		return nil, fmt.Errorf("rate limiter not found: %s", limiterID)
	}

	if !config.Enabled {
		return &rate.Reservation{}, nil
	}

	return limiter.ReserveN(time.Now(), n), nil
}

// GetConfig returns the configuration of a rate limiter
func (rlm *RateLimiterManager) GetConfig(limiterID string) (*RateLimitConfig, error) {
	rlm.mutex.RLock()
	defer rlm.mutex.RUnlock()

	config, exists := rlm.configs[limiterID]
	if !exists {
		return nil, fmt.Errorf("rate limiter not found: %s", limiterID)
	}

	return config, nil
}

// GetAllConfigs returns all rate limiter configurations
func (rlm *RateLimiterManager) GetAllConfigs() map[string]*RateLimitConfig {
	rlm.mutex.RLock()
	defer rlm.mutex.RUnlock()

	configs := make(map[string]*RateLimitConfig)
	for id, config := range rlm.configs {
		configs[id] = config
	}

	return configs
}

// UpdateRateLimiter updates a rate limiter configuration
func (rlm *RateLimiterManager) UpdateRateLimiter(limiterID string, config *RateLimitConfig) error {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()

	if _, exists := rlm.configs[limiterID]; !exists {
		return fmt.Errorf("rate limiter not found: %s", limiterID)
	}

	// Update configuration
	config.ID = limiterID
	config.UpdatedAt = time.Now()

	// Create new limiter with updated configuration
	limiter := rate.NewLimiter(rate.Every(config.Window/time.Duration(config.Requests)), config.Burst)

	rlm.limiters[limiterID] = limiter
	rlm.configs[limiterID] = config

	rlm.logger.WithFields(logrus.Fields{
		"id":       limiterID,
		"name":     config.Name,
		"requests": config.Requests,
		"window":   config.Window,
		"burst":    config.Burst,
		"enabled":  config.Enabled,
	}).Info("Rate limiter updated successfully")

	return nil
}

// DeleteRateLimiter deletes a rate limiter
func (rlm *RateLimiterManager) DeleteRateLimiter(limiterID string) error {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()

	if _, exists := rlm.limiters[limiterID]; !exists {
		return fmt.Errorf("rate limiter not found: %s", limiterID)
	}

	delete(rlm.limiters, limiterID)
	delete(rlm.configs, limiterID)

	rlm.logger.WithField("id", limiterID).Info("Rate limiter deleted successfully")
	return nil
}

// EnableRateLimiter enables a rate limiter
func (rlm *RateLimiterManager) EnableRateLimiter(limiterID string) error {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()

	config, exists := rlm.configs[limiterID]
	if !exists {
		return fmt.Errorf("rate limiter not found: %s", limiterID)
	}

	config.Enabled = true
	config.UpdatedAt = time.Now()

	rlm.logger.WithField("id", limiterID).Info("Rate limiter enabled")
	return nil
}

// DisableRateLimiter disables a rate limiter
func (rlm *RateLimiterManager) DisableRateLimiter(limiterID string) error {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()

	config, exists := rlm.configs[limiterID]
	if !exists {
		return fmt.Errorf("rate limiter not found: %s", limiterID)
	}

	config.Enabled = false
	config.UpdatedAt = time.Now()

	rlm.logger.WithField("id", limiterID).Info("Rate limiter disabled")
	return nil
}

// logEvent logs a rate limit event
func (rlm *RateLimiterManager) logEvent(limiterID, event string, allowed bool, remaining, limit int, window time.Duration, metadata map[string]interface{}) {
	eventData := &RateLimitEvent{
		ID:        uuid.New(),
		LimiterID: limiterID,
		Event:     event,
		Allowed:   allowed,
		Remaining: remaining,
		Limit:     limit,
		Window:    window,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	rlm.logger.WithFields(logrus.Fields{
		"event_id":   eventData.ID,
		"limiter_id": limiterID,
		"event":      event,
		"allowed":    allowed,
		"remaining":  remaining,
		"limit":      limit,
		"window":     window,
		"metadata":   metadata,
	}).Debug("Rate limit event")

	// Here you could send the event to a message queue or store it in a database
}

// CreateAPIRateLimiter creates a rate limiter for API endpoints
func (rlm *RateLimiterManager) CreateAPIRateLimiter(name string, requests int, window time.Duration) error {
	config := &RateLimitConfig{
		Name:        name,
		Requests:    requests,
		Window:      window,
		Burst:       requests,
		Description: fmt.Sprintf("API rate limiter: %d requests per %v", requests, window),
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return rlm.CreateRateLimiter(config)
}

// CreateUserRateLimiter creates a rate limiter for user-specific operations
func (rlm *RateLimiterManager) CreateUserRateLimiter(name string, requests int, window time.Duration) error {
	config := &RateLimitConfig{
		Name:        name,
		Requests:    requests,
		Window:      window,
		Burst:       requests,
		Description: fmt.Sprintf("User rate limiter: %d requests per %v", requests, window),
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return rlm.CreateRateLimiter(config)
}

// CreateGlobalRateLimiter creates a global rate limiter
func (rlm *RateLimiterManager) CreateGlobalRateLimiter(name string, requests int, window time.Duration) error {
	config := &RateLimitConfig{
		Name:        name,
		Requests:    requests,
		Window:      window,
		Burst:       requests,
		Description: fmt.Sprintf("Global rate limiter: %d requests per %v", requests, window),
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return rlm.CreateRateLimiter(config)
}
