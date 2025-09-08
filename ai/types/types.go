package types

import (
	"context"
	"time"
)

// AIProvider defines the interface for AI providers
type AIProvider interface {
	// Chat sends a chat request to the AI provider
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// GenerateText generates text based on a prompt
	GenerateText(ctx context.Context, req *TextGenerationRequest) (*TextGenerationResponse, error)

	// EmbedText generates embeddings for text
	EmbedText(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// GetModelInfo returns information about available models
	GetModelInfo(ctx context.Context) ([]ModelInfo, error)

	// GetProviderName returns the name of the provider
	GetProviderName() string

	// IsHealthy checks if the provider is healthy
	IsHealthy(ctx context.Context) error
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Messages         []Message `json:"messages"`
	Model            string    `json:"model"`
	Temperature      float64   `json:"temperature,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
	Error   *AIError `json:"error,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// Choice represents a choice in the response
type Choice struct {
	Index        int      `json:"index"`
	Message      Message  `json:"message"`
	FinishReason string   `json:"finish_reason"`
	Delta        *Message `json:"delta,omitempty"` // For streaming
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// TextGenerationRequest represents a text generation request
type TextGenerationRequest struct {
	Prompt      string   `json:"prompt"`
	Model       string   `json:"model"`
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// TextGenerationResponse represents a text generation response
type TextGenerationResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []TextChoice `json:"choices"`
	Usage   Usage        `json:"usage"`
	Error   *AIError     `json:"error,omitempty"`
}

// TextChoice represents a choice in text generation
type TextChoice struct {
	Index        int    `json:"index"`
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
	Error  *AIError    `json:"error,omitempty"`
}

// Embedding represents a text embedding
type Embedding struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// ModelInfo represents information about an AI model
type ModelInfo struct {
	ID          string   `json:"id"`
	Object      string   `json:"object"`
	Created     int64    `json:"created"`
	OwnedBy     string   `json:"owned_by"`
	Permission  []string `json:"permission"`
	Root        string   `json:"root"`
	Parent      string   `json:"parent"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	ContextSize int      `json:"context_size,omitempty"`
	Description string   `json:"description,omitempty"`
}

// AIError represents an AI provider error
type AIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// ProviderConfig represents configuration for an AI provider
type ProviderConfig struct {
	Name         string            `json:"name"`
	APIKey       string            `json:"api_key"`
	BaseURL      string            `json:"base_url,omitempty"`
	Timeout      time.Duration     `json:"timeout,omitempty"`
	MaxRetries   int               `json:"max_retries,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	Models       []string          `json:"models,omitempty"`
	DefaultModel string            `json:"default_model,omitempty"`
}

// StreamCallback represents a callback function for streaming responses
type StreamCallback func(chunk *ChatResponse) error

// HealthStatus represents the health status of a provider
type HealthStatus struct {
	Provider  string    `json:"provider"`
	Healthy   bool      `json:"healthy"`
	Message   string    `json:"message,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

// ProviderStats represents statistics for a provider
type ProviderStats struct {
	Provider      string    `json:"provider"`
	TotalRequests int64     `json:"total_requests"`
	TotalTokens   int64     `json:"total_tokens"`
	SuccessRate   float64   `json:"success_rate"`
	AvgLatency    float64   `json:"avg_latency_ms"`
	LastUsed      time.Time `json:"last_used"`
}
