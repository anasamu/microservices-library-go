package config

import (
	"fmt"
	"regexp"
	"sync"
	"time"
)

// SimpleConfig provides a simple configuration management without external dependencies
type SimpleConfig struct {
	configs    map[string]interface{}
	watchers   map[string][]ConfigWatcher
	mu         sync.RWMutex
	hotReload  bool
	reloadChan chan ConfigChange
}

// ConfigWatcher defines the interface for configuration watchers
type ConfigWatcher interface {
	OnConfigChange(key string, oldValue, newValue interface{})
}

// ConfigChange represents a configuration change event
type ConfigChange struct {
	Key      string
	OldValue interface{}
	NewValue interface{}
	Time     time.Time
}

// NewSimpleConfig creates a new simple configuration manager
func NewSimpleConfig() *SimpleConfig {
	return &SimpleConfig{
		configs:    make(map[string]interface{}),
		watchers:   make(map[string][]ConfigWatcher),
		hotReload:  true,
		reloadChan: make(chan ConfigChange, 100),
	}
}

// Set sets a configuration value
func (sc *SimpleConfig) Set(key string, value interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	oldValue := sc.configs[key]
	sc.configs[key] = value

	// Notify watchers
	sc.notifyWatchers(key, oldValue, value)
}

// Get retrieves a configuration value
func (sc *SimpleConfig) Get(key string) interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return sc.configs[key]
}

// GetString retrieves a string configuration value
func (sc *SimpleConfig) GetString(key string) string {
	value := sc.Get(key)
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// GetInt retrieves an integer configuration value
func (sc *SimpleConfig) GetInt(key string) int {
	value := sc.Get(key)
	if i, ok := value.(int); ok {
		return i
	}
	return 0
}

// GetBool retrieves a boolean configuration value
func (sc *SimpleConfig) GetBool(key string) bool {
	value := sc.Get(key)
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

// GetDuration retrieves a duration configuration value
func (sc *SimpleConfig) GetDuration(key string) time.Duration {
	value := sc.Get(key)
	if d, ok := value.(time.Duration); ok {
		return d
	}
	if str, ok := value.(string); ok {
		if d, err := time.ParseDuration(str); err == nil {
			return d
		}
	}
	return 0
}

// GetStringSlice retrieves a string slice configuration value
func (sc *SimpleConfig) GetStringSlice(key string) []string {
	value := sc.Get(key)
	if slice, ok := value.([]string); ok {
		return slice
	}
	return []string{}
}

// GetWithDefault retrieves a configuration value with a default
func (sc *SimpleConfig) GetWithDefault(key string, defaultValue interface{}) interface{} {
	value := sc.Get(key)
	if value == nil {
		return defaultValue
	}
	return value
}

// Watch registers a watcher for configuration changes
func (sc *SimpleConfig) Watch(key string, watcher ConfigWatcher) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.watchers[key] = append(sc.watchers[key], watcher)
}

// Unwatch removes a watcher for configuration changes
func (sc *SimpleConfig) Unwatch(key string, watcher ConfigWatcher) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	watchers := sc.watchers[key]
	for i, w := range watchers {
		if w == watcher {
			sc.watchers[key] = append(watchers[:i], watchers[i+1:]...)
			break
		}
	}
}

// GetChangeChannel returns the channel for configuration changes
func (sc *SimpleConfig) GetChangeChannel() <-chan ConfigChange {
	return sc.reloadChan
}

// SetHotReload enables or disables hot reload
func (sc *SimpleConfig) SetHotReload(enabled bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.hotReload = enabled
}

// IsHotReloadEnabled returns whether hot reload is enabled
func (sc *SimpleConfig) IsHotReloadEnabled() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.hotReload
}

// GetAll returns all configuration values
func (sc *SimpleConfig) GetAll() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	all := make(map[string]interface{})
	for key, value := range sc.configs {
		all[key] = value
	}
	return all
}

// SetDefaults sets default configuration values
func (sc *SimpleConfig) SetDefaults(defaults map[string]interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for key, value := range defaults {
		if _, exists := sc.configs[key]; !exists {
			sc.configs[key] = value
		}
	}
}

// Validate validates configuration values
func (sc *SimpleConfig) Validate(rules map[string]ValidationRule) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	for key, rule := range rules {
		value := sc.configs[key]
		if err := rule.Validate(value); err != nil {
			return fmt.Errorf("validation failed for key %s: %w", key, err)
		}
	}

	return nil
}

// notifyWatchers notifies all watchers of a configuration change
func (sc *SimpleConfig) notifyWatchers(key string, oldValue, newValue interface{}) {
	// Send to change channel
	select {
	case sc.reloadChan <- ConfigChange{
		Key:      key,
		OldValue: oldValue,
		NewValue: newValue,
		Time:     time.Now(),
	}:
	default:
		// Channel is full, skip
	}

	// Notify registered watchers
	if watchers, exists := sc.watchers[key]; exists {
		for _, watcher := range watchers {
			go watcher.OnConfigChange(key, oldValue, newValue)
		}
	}
}

// ValidationRule defines the interface for configuration validation
type ValidationRule interface {
	Validate(value interface{}) error
}

// StringValidationRule validates string values
type StringValidationRule struct {
	Required bool
	MinLen   int
	MaxLen   int
	Pattern  string
}

// Validate validates a string value
func (r *StringValidationRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	if r.Required && str == "" {
		return fmt.Errorf("value is required")
	}

	if r.MinLen > 0 && len(str) < r.MinLen {
		return fmt.Errorf("minimum length is %d", r.MinLen)
	}

	if r.MaxLen > 0 && len(str) > r.MaxLen {
		return fmt.Errorf("maximum length is %d", r.MaxLen)
	}

	if r.Pattern != "" {
		matched, err := regexp.MatchString(r.Pattern, str)
		if err != nil {
			return fmt.Errorf("invalid pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf("value does not match pattern %s", r.Pattern)
		}
	}

	return nil
}

// IntValidationRule validates integer values
type IntValidationRule struct {
	Required bool
	Min      int
	Max      int
}

// Validate validates an integer value
func (r *IntValidationRule) Validate(value interface{}) error {
	var i int
	switch v := value.(type) {
	case int:
		i = v
	case int64:
		i = int(v)
	case float64:
		i = int(v)
	default:
		return fmt.Errorf("expected integer, got %T", value)
	}

	if r.Required && i == 0 {
		return fmt.Errorf("value is required")
	}

	if r.Min > 0 && i < r.Min {
		return fmt.Errorf("minimum value is %d", r.Min)
	}

	if r.Max > 0 && i > r.Max {
		return fmt.Errorf("maximum value is %d", r.Max)
	}

	return nil
}

// BoolValidationRule validates boolean values
type BoolValidationRule struct {
	Required bool
}

// Validate validates a boolean value
func (r *BoolValidationRule) Validate(value interface{}) error {
	_, ok := value.(bool)
	if !ok {
		return fmt.Errorf("expected boolean, got %T", value)
	}
	return nil
}

// DurationValidationRule validates duration values
type DurationValidationRule struct {
	Required bool
	Min      time.Duration
	Max      time.Duration
}

// Validate validates a duration value
func (r *DurationValidationRule) Validate(value interface{}) error {
	var d time.Duration
	switch v := value.(type) {
	case time.Duration:
		d = v
	case string:
		var err error
		d, err = time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}
	default:
		return fmt.Errorf("expected duration, got %T", value)
	}

	if r.Required && d == 0 {
		return fmt.Errorf("value is required")
	}

	if r.Min > 0 && d < r.Min {
		return fmt.Errorf("minimum duration is %v", r.Min)
	}

	if r.Max > 0 && d > r.Max {
		return fmt.Errorf("maximum duration is %v", r.Max)
	}

	return nil
}

// SimpleConfigManager provides high-level configuration management
type SimpleConfigManager struct {
	config     *SimpleConfig
	validators map[string]ValidationRule
	mu         sync.RWMutex
}

// NewSimpleConfigManager creates a new simple configuration manager
func NewSimpleConfigManager() *SimpleConfigManager {
	return &SimpleConfigManager{
		config:     NewSimpleConfig(),
		validators: make(map[string]ValidationRule),
	}
}

// SetValidator sets a validation rule for a configuration key
func (scm *SimpleConfigManager) SetValidator(key string, rule ValidationRule) {
	scm.mu.Lock()
	defer scm.mu.Unlock()
	scm.validators[key] = rule
}

// Validate validates all configurations
func (scm *SimpleConfigManager) Validate() error {
	scm.mu.RLock()
	validators := make(map[string]ValidationRule)
	for key, rule := range scm.validators {
		validators[key] = rule
	}
	scm.mu.RUnlock()

	return scm.config.Validate(validators)
}

// GetConfig returns the underlying configuration
func (scm *SimpleConfigManager) GetConfig() *SimpleConfig {
	return scm.config
}

// LoadFromMap loads configuration from a map
func (scm *SimpleConfigManager) LoadFromMap(configMap map[string]interface{}) {
	for key, value := range configMap {
		scm.config.Set(key, value)
	}
}

// LoadFromEnv loads configuration from environment variables
func (scm *SimpleConfigManager) LoadFromEnv() {
	// This is a simplified version - in a real implementation,
	// you would use os.Getenv() to read environment variables
	// and set them in the configuration
}

// SetDefaults sets default configuration values
func (scm *SimpleConfigManager) SetDefaults(defaults map[string]interface{}) {
	scm.config.SetDefaults(defaults)
}
