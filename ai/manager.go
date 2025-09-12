package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/ai/providers/anthropic"
	"github.com/anasamu/microservices-library-go/ai/providers/deepseek"
	"github.com/anasamu/microservices-library-go/ai/providers/google"
	"github.com/anasamu/microservices-library-go/ai/providers/openai"
	"github.com/anasamu/microservices-library-go/ai/providers/xai"
	"github.com/anasamu/microservices-library-go/ai/types"
)

// AIManager manages multiple AI providers
type AIManager struct {
	providers map[string]types.AIProvider
	configs   map[string]*types.ProviderConfig
	stats     map[string]*types.ProviderStats
	mu        sync.RWMutex
}

// NewAIManager creates a new AI manager
func NewAIManager() *AIManager {
	return &AIManager{
		providers: make(map[string]types.AIProvider),
		configs:   make(map[string]*types.ProviderConfig),
		stats:     make(map[string]*types.ProviderStats),
	}
}

// AddProvider adds a new AI provider
func (m *AIManager) AddProvider(config *types.ProviderConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config.Name == "" {
		return fmt.Errorf("provider name is required")
	}
	if config.APIKey == "" {
		return fmt.Errorf("API key is required for provider %s", config.Name)
	}

	var provider types.AIProvider

	switch config.Name {
	case "openai":
		provider = openai.NewOpenAIProvider(config)
	case "anthropic":
		provider = anthropic.NewAnthropicProvider(config)
	case "xai":
		provider = xai.NewXAIProvider(config)
	case "deepseek":
		provider = deepseek.NewDeepSeekProvider(config)
	case "google":
		provider = google.NewGoogleProvider(config)
	default:
		return fmt.Errorf("unsupported provider: %s", config.Name)
	}

	// Test the provider
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := provider.IsHealthy(ctx); err != nil {
		return fmt.Errorf("provider %s health check failed: %w", config.Name, err)
	}

	m.providers[config.Name] = provider
	m.configs[config.Name] = config
	m.stats[config.Name] = &types.ProviderStats{
		Provider:      config.Name,
		TotalRequests: 0,
		TotalTokens:   0,
		SuccessRate:   0.0,
		AvgLatency:    0.0,
		LastUsed:      time.Now(),
	}

	return nil
}

// RemoveProvider removes an AI provider
func (m *AIManager) RemoveProvider(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	delete(m.providers, name)
	delete(m.configs, name)
	delete(m.stats, name)

	return nil
}

// GetProvider returns a provider by name
func (m *AIManager) GetProvider(name string) (types.AIProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// ListProviders returns a list of all provider names
func (m *AIManager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]string, 0, len(m.providers))
	for name := range m.providers {
		providers = append(providers, name)
	}

	return providers
}

// Chat sends a chat request to a specific provider
func (m *AIManager) Chat(ctx context.Context, providerName string, req *types.ChatRequest) (*types.ChatResponse, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := provider.Chat(ctx, req)
	duration := time.Since(start)

	// Update stats
	m.updateStats(providerName, duration, err == nil, resp)

	return resp, err
}

// GenerateText generates text using a specific provider
func (m *AIManager) GenerateText(ctx context.Context, providerName string, req *types.TextGenerationRequest) (*types.TextGenerationResponse, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := provider.GenerateText(ctx, req)
	duration := time.Since(start)

	// Update stats
	m.updateStats(providerName, duration, err == nil, nil)

	return resp, err
}

// EmbedText generates embeddings using a specific provider
func (m *AIManager) EmbedText(ctx context.Context, providerName string, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := provider.EmbedText(ctx, req)
	duration := time.Since(start)

	// Update stats
	m.updateStats(providerName, duration, err == nil, nil)

	return resp, err
}

// GetModelInfo returns model information for a specific provider
func (m *AIManager) GetModelInfo(ctx context.Context, providerName string) ([]types.ModelInfo, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetModelInfo(ctx)
}

// GetAllModelInfo returns model information for all providers
func (m *AIManager) GetAllModelInfo(ctx context.Context) (map[string][]types.ModelInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allModels := make(map[string][]types.ModelInfo)

	for name, provider := range m.providers {
		models, err := provider.GetModelInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get models for provider %s: %w", name, err)
		}
		allModels[name] = models
	}

	return allModels, nil
}

// HealthCheck checks the health of all providers
func (m *AIManager) HealthCheck(ctx context.Context) (map[string]*types.HealthStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	healthStatus := make(map[string]*types.HealthStatus)

	for name, provider := range m.providers {
		status := &types.HealthStatus{
			Provider:  name,
			Healthy:   false,
			CheckedAt: time.Now(),
		}

		if err := provider.IsHealthy(ctx); err != nil {
			status.Message = err.Error()
		} else {
			status.Healthy = true
		}

		healthStatus[name] = status
	}

	return healthStatus, nil
}

// GetStats returns statistics for all providers
func (m *AIManager) GetStats() map[string]*types.ProviderStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]*types.ProviderStats)
	for name, stat := range m.stats {
		stats[name] = &types.ProviderStats{
			Provider:      stat.Provider,
			TotalRequests: stat.TotalRequests,
			TotalTokens:   stat.TotalTokens,
			SuccessRate:   stat.SuccessRate,
			AvgLatency:    stat.AvgLatency,
			LastUsed:      stat.LastUsed,
		}
	}

	return stats
}

// GetProviderStats returns statistics for a specific provider
func (m *AIManager) GetProviderStats(providerName string) (*types.ProviderStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats, exists := m.stats[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	return &types.ProviderStats{
		Provider:      stats.Provider,
		TotalRequests: stats.TotalRequests,
		TotalTokens:   stats.TotalTokens,
		SuccessRate:   stats.SuccessRate,
		AvgLatency:    stats.AvgLatency,
		LastUsed:      stats.LastUsed,
	}, nil
}

// updateStats updates provider statistics
func (m *AIManager) updateStats(providerName string, duration time.Duration, success bool, resp interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.stats[providerName]
	if !exists {
		return
	}

	stats.TotalRequests++
	stats.LastUsed = time.Now()

	// Update latency
	if stats.AvgLatency == 0 {
		stats.AvgLatency = float64(duration.Milliseconds())
	} else {
		stats.AvgLatency = (stats.AvgLatency + float64(duration.Milliseconds())) / 2
	}

	// Update success rate
	if success {
		stats.SuccessRate = (stats.SuccessRate*float64(stats.TotalRequests-1) + 1.0) / float64(stats.TotalRequests)
	} else {
		stats.SuccessRate = (stats.SuccessRate * float64(stats.TotalRequests-1)) / float64(stats.TotalRequests)
	}

	// Update token count if response contains usage info
	if resp != nil {
		switch r := resp.(type) {
		case *types.ChatResponse:
			if r.Usage.TotalTokens > 0 {
				stats.TotalTokens += int64(r.Usage.TotalTokens)
			}
		case *types.TextGenerationResponse:
			if r.Usage.TotalTokens > 0 {
				stats.TotalTokens += int64(r.Usage.TotalTokens)
			}
		case *types.EmbeddingResponse:
			if r.Usage.TotalTokens > 0 {
				stats.TotalTokens += int64(r.Usage.TotalTokens)
			}
		}
	}
}

// ChatWithFallback sends a chat request with automatic fallback to other providers
func (m *AIManager) ChatWithFallback(ctx context.Context, primaryProvider string, req *types.ChatRequest) (*types.ChatResponse, error) {
	providers := m.ListProviders()
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Try primary provider first
	if primaryProvider != "" {
		if _, err := m.GetProvider(primaryProvider); err == nil {
			resp, err := m.Chat(ctx, primaryProvider, req)
			if err == nil && resp.Error == nil {
				return resp, nil
			}
		}
	}

	// Try other providers
	for _, providerName := range providers {
		if providerName == primaryProvider {
			continue
		}

		resp, err := m.Chat(ctx, providerName, req)
		if err == nil && resp.Error == nil {
			return resp, nil
		}
	}

	return nil, fmt.Errorf("all providers failed")
}

// GenerateTextWithFallback generates text with automatic fallback
func (m *AIManager) GenerateTextWithFallback(ctx context.Context, primaryProvider string, req *types.TextGenerationRequest) (*types.TextGenerationResponse, error) {
	providers := m.ListProviders()
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Try primary provider first
	if primaryProvider != "" {
		if _, err := m.GetProvider(primaryProvider); err == nil {
			resp, err := m.GenerateText(ctx, primaryProvider, req)
			if err == nil && resp.Error == nil {
				return resp, nil
			}
		}
	}

	// Try other providers
	for _, providerName := range providers {
		if providerName == primaryProvider {
			continue
		}

		resp, err := m.GenerateText(ctx, providerName, req)
		if err == nil && resp.Error == nil {
			return resp, nil
		}
	}

	return nil, fmt.Errorf("all providers failed")
}
