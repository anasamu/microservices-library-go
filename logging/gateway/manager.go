package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/logging/types"
	"github.com/sirupsen/logrus"
)

// LoggingManager manages multiple logging providers
type LoggingManager struct {
	providers map[string]types.LoggingProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds logging manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	FallbackEnabled bool              `json:"fallback_enabled"`
	Metadata        map[string]string `json:"metadata"`
}

// NewLoggingManager creates a new logging manager
func NewLoggingManager(config *ManagerConfig, logger *logrus.Logger) *LoggingManager {
	if config == nil {
		config = &ManagerConfig{
			RetryAttempts:   3,
			RetryDelay:      time.Second,
			Timeout:         30 * time.Second,
			FallbackEnabled: true,
		}
	}

	return &LoggingManager{
		providers: make(map[string]types.LoggingProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a logging provider
func (lm *LoggingManager) RegisterProvider(provider types.LoggingProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := lm.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	lm.providers[name] = provider
	lm.logger.WithField("provider", name).Info("Logging provider registered")

	return nil
}

// GetProvider returns a specific provider by name
func (lm *LoggingManager) GetProvider(name string) (types.LoggingProvider, error) {
	provider, exists := lm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default provider
func (lm *LoggingManager) GetDefaultProvider() (types.LoggingProvider, error) {
	if lm.config.DefaultProvider == "" {
		// Return first available provider
		for _, provider := range lm.providers {
			return provider, nil
		}
		return nil, fmt.Errorf("no providers registered")
	}

	return lm.GetProvider(lm.config.DefaultProvider)
}

// Log logs a message using the default provider
func (lm *LoggingManager) Log(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	return lm.LogWithProvider(ctx, "", level, message, fields)
}

// LogWithProvider logs a message using a specific provider
func (lm *LoggingManager) LogWithProvider(ctx context.Context, providerName string, level types.LogLevel, message string, fields map[string]interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Log(ctx, level, message, fields)
	})
}

// LogWithContext logs a message with context using the default provider
func (lm *LoggingManager) LogWithContext(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	return lm.LogWithContextAndProvider(ctx, "", level, message, fields)
}

// LogWithContextAndProvider logs a message with context using a specific provider
func (lm *LoggingManager) LogWithContextAndProvider(ctx context.Context, providerName string, level types.LogLevel, message string, fields map[string]interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.LogWithContext(ctx, level, message, fields)
	})
}

// Info logs an info level message using the default provider
func (lm *LoggingManager) Info(ctx context.Context, message string, fields ...map[string]interface{}) error {
	return lm.InfoWithProvider(ctx, "", message, fields...)
}

// InfoWithProvider logs an info level message using a specific provider
func (lm *LoggingManager) InfoWithProvider(ctx context.Context, providerName string, message string, fields ...map[string]interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Info(ctx, message, fields...)
	})
}

// Debug logs a debug level message using the default provider
func (lm *LoggingManager) Debug(ctx context.Context, message string, fields ...map[string]interface{}) error {
	return lm.DebugWithProvider(ctx, "", message, fields...)
}

// DebugWithProvider logs a debug level message using a specific provider
func (lm *LoggingManager) DebugWithProvider(ctx context.Context, providerName string, message string, fields ...map[string]interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Debug(ctx, message, fields...)
	})
}

// Warn logs a warning level message using the default provider
func (lm *LoggingManager) Warn(ctx context.Context, message string, fields ...map[string]interface{}) error {
	return lm.WarnWithProvider(ctx, "", message, fields...)
}

// WarnWithProvider logs a warning level message using a specific provider
func (lm *LoggingManager) WarnWithProvider(ctx context.Context, providerName string, message string, fields ...map[string]interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Warn(ctx, message, fields...)
	})
}

// Error logs an error level message using the default provider
func (lm *LoggingManager) Error(ctx context.Context, message string, fields ...map[string]interface{}) error {
	return lm.ErrorWithProvider(ctx, "", message, fields...)
}

// ErrorWithProvider logs an error level message using a specific provider
func (lm *LoggingManager) ErrorWithProvider(ctx context.Context, providerName string, message string, fields ...map[string]interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Error(ctx, message, fields...)
	})
}

// Fatal logs a fatal level message using the default provider
func (lm *LoggingManager) Fatal(ctx context.Context, message string, fields ...map[string]interface{}) error {
	return lm.FatalWithProvider(ctx, "", message, fields...)
}

// FatalWithProvider logs a fatal level message using a specific provider
func (lm *LoggingManager) FatalWithProvider(ctx context.Context, providerName string, message string, fields ...map[string]interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Fatal(ctx, message, fields...)
	})
}

// Panic logs a panic level message using the default provider
func (lm *LoggingManager) Panic(ctx context.Context, message string, fields ...map[string]interface{}) error {
	return lm.PanicWithProvider(ctx, "", message, fields...)
}

// PanicWithProvider logs a panic level message using a specific provider
func (lm *LoggingManager) PanicWithProvider(ctx context.Context, providerName string, message string, fields ...map[string]interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Panic(ctx, message, fields...)
	})
}

// Infof logs a formatted info level message using the default provider
func (lm *LoggingManager) Infof(ctx context.Context, format string, args ...interface{}) error {
	return lm.InfofWithProvider(ctx, "", format, args...)
}

// InfofWithProvider logs a formatted info level message using a specific provider
func (lm *LoggingManager) InfofWithProvider(ctx context.Context, providerName string, format string, args ...interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Infof(ctx, format, args...)
	})
}

// Debugf logs a formatted debug level message using the default provider
func (lm *LoggingManager) Debugf(ctx context.Context, format string, args ...interface{}) error {
	return lm.DebugfWithProvider(ctx, "", format, args...)
}

// DebugfWithProvider logs a formatted debug level message using a specific provider
func (lm *LoggingManager) DebugfWithProvider(ctx context.Context, providerName string, format string, args ...interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Debugf(ctx, format, args...)
	})
}

// Warnf logs a formatted warning level message using the default provider
func (lm *LoggingManager) Warnf(ctx context.Context, format string, args ...interface{}) error {
	return lm.WarnfWithProvider(ctx, "", format, args...)
}

// WarnfWithProvider logs a formatted warning level message using a specific provider
func (lm *LoggingManager) WarnfWithProvider(ctx context.Context, providerName string, format string, args ...interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Warnf(ctx, format, args...)
	})
}

// Errorf logs a formatted error level message using the default provider
func (lm *LoggingManager) Errorf(ctx context.Context, format string, args ...interface{}) error {
	return lm.ErrorfWithProvider(ctx, "", format, args...)
}

// ErrorfWithProvider logs a formatted error level message using a specific provider
func (lm *LoggingManager) ErrorfWithProvider(ctx context.Context, providerName string, format string, args ...interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Errorf(ctx, format, args...)
	})
}

// Fatalf logs a formatted fatal level message using the default provider
func (lm *LoggingManager) Fatalf(ctx context.Context, format string, args ...interface{}) error {
	return lm.FatalfWithProvider(ctx, "", format, args...)
}

// FatalfWithProvider logs a formatted fatal level message using a specific provider
func (lm *LoggingManager) FatalfWithProvider(ctx context.Context, providerName string, format string, args ...interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Fatalf(ctx, format, args...)
	})
}

// Panicf logs a formatted panic level message using the default provider
func (lm *LoggingManager) Panicf(ctx context.Context, format string, args ...interface{}) error {
	return lm.PanicfWithProvider(ctx, "", format, args...)
}

// PanicfWithProvider logs a formatted panic level message using a specific provider
func (lm *LoggingManager) PanicfWithProvider(ctx context.Context, providerName string, format string, args ...interface{}) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.Panicf(ctx, format, args...)
	})
}

// LogBatch logs multiple entries using the default provider
func (lm *LoggingManager) LogBatch(ctx context.Context, entries []types.LogEntry) error {
	return lm.LogBatchWithProvider(ctx, "", entries)
}

// LogBatchWithProvider logs multiple entries using a specific provider
func (lm *LoggingManager) LogBatchWithProvider(ctx context.Context, providerName string, entries []types.LogEntry) error {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return err
	}

	return lm.executeWithRetry(ctx, func() error {
		return provider.LogBatch(ctx, entries)
	})
}

// Search searches logs using the default provider
func (lm *LoggingManager) Search(ctx context.Context, query types.LogQuery) ([]types.LogEntry, error) {
	return lm.SearchWithProvider(ctx, "", query)
}

// SearchWithProvider searches logs using a specific provider
func (lm *LoggingManager) SearchWithProvider(ctx context.Context, providerName string, query types.LogQuery) ([]types.LogEntry, error) {
	provider, err := lm.getProvider(providerName)
	if err != nil {
		return nil, err
	}

	var result []types.LogEntry
	err = lm.executeWithRetry(ctx, func() error {
		var err error
		result, err = provider.Search(ctx, query)
		return err
	})

	return result, err
}

// GetStats returns logging statistics from all providers
func (lm *LoggingManager) GetStats(ctx context.Context) (map[string]*types.LoggingStats, error) {
	stats := make(map[string]*types.LoggingStats)

	for name, provider := range lm.providers {
		stat, err := provider.GetStats(ctx)
		if err != nil {
			lm.logger.WithError(err).WithField("provider", name).Warn("Failed to get stats from provider")
			continue
		}
		stats[name] = stat
	}

	return stats, nil
}

// Close closes all providers
func (lm *LoggingManager) Close() error {
	var lastErr error

	for name, provider := range lm.providers {
		if err := provider.Disconnect(context.Background()); err != nil {
			lm.logger.WithError(err).WithField("provider", name).Error("Failed to close provider")
			lastErr = err
		}
	}

	return lastErr
}

// getProvider returns a provider by name or the default provider
func (lm *LoggingManager) getProvider(name string) (types.LoggingProvider, error) {
	if name == "" {
		return lm.GetDefaultProvider()
	}
	return lm.GetProvider(name)
}

// executeWithRetry executes a function with retry logic
func (lm *LoggingManager) executeWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= lm.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(lm.config.RetryDelay):
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			lm.logger.WithError(err).WithField("attempt", attempt+1).Debug("Logging operation failed, retrying")
			continue
		}

		return nil
	}

	return fmt.Errorf("operation failed after %d attempts: %w", lm.config.RetryAttempts+1, lastErr)
}

// ListProviders returns a list of registered provider names
func (lm *LoggingManager) ListProviders() []string {
	providers := make([]string, 0, len(lm.providers))
	for name := range lm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderInfo returns information about all registered providers
func (lm *LoggingManager) GetProviderInfo() map[string]*types.ProviderInfo {
	info := make(map[string]*types.ProviderInfo)

	for name, provider := range lm.providers {
		info[name] = &types.ProviderInfo{
			Name:              provider.GetName(),
			SupportedFeatures: provider.GetSupportedFeatures(),
			ConnectionInfo:    provider.GetConnectionInfo(),
			IsConnected:       provider.IsConnected(),
		}
	}

	return info
}
