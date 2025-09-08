package xai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anasamu/microservices-library-go/ai/types"
)

// XAIProvider implements the AIProvider interface for X.AI (Grok)
type XAIProvider struct {
	config     *types.ProviderConfig
	httpClient *http.Client
}

// NewXAIProvider creates a new X.AI provider
func NewXAIProvider(config *types.ProviderConfig) *XAIProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.x.ai/v1"
	}
	if config.DefaultModel == "" {
		config.DefaultModel = "grok-beta"
	}

	return &XAIProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetProviderName returns the provider name
func (p *XAIProvider) GetProviderName() string {
	return "xai"
}

// Chat sends a chat request to X.AI
func (p *XAIProvider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.config.DefaultModel
	}

	// Convert to X.AI format
	xaiReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
		"stream":      req.Stream,
		"top_p":       req.TopP,
	}

	// Remove empty values
	if req.Temperature == 0 {
		delete(xaiReq, "temperature")
	}
	if req.MaxTokens == 0 {
		delete(xaiReq, "max_tokens")
	}
	if req.TopP == 0 {
		delete(xaiReq, "top_p")
	}

	jsonData, err := json.Marshal(xaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	for key, value := range p.config.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error *types.AIError `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return &types.ChatResponse{Error: errorResp.Error}, nil
	}

	var chatResp types.ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResp, nil
}

// GenerateText generates text based on a prompt
func (p *XAIProvider) GenerateText(ctx context.Context, req *types.TextGenerationRequest) (*types.TextGenerationResponse, error) {
	// Convert to chat format for X.AI
	chatReq := &types.ChatRequest{
		Messages: []types.Message{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		Model:       req.Model,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
	}

	chatResp, err := p.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	if chatResp.Error != nil {
		return &types.TextGenerationResponse{Error: chatResp.Error}, nil
	}

	// Convert to text generation response
	textResp := &types.TextGenerationResponse{
		ID:      chatResp.ID,
		Object:  "text_completion",
		Created: chatResp.Created,
		Model:   chatResp.Model,
		Choices: []types.TextChoice{
			{
				Index:        0,
				Text:         chatResp.Choices[0].Message.Content,
				FinishReason: chatResp.Choices[0].FinishReason,
			},
		},
		Usage: chatResp.Usage,
	}

	return textResp, nil
}

// EmbedText generates embeddings for text
func (p *XAIProvider) EmbedText(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	// X.AI doesn't have a direct embedding API yet
	return nil, fmt.Errorf("embedding not supported by X.AI API")
}

// GetModelInfo returns information about available models
func (p *XAIProvider) GetModelInfo(ctx context.Context) ([]types.ModelInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	for key, value := range p.config.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// If models endpoint is not available, return known models
		models := []types.ModelInfo{
			{
				ID:          "grok-beta",
				Object:      "model",
				Created:     1709251200,
				OwnedBy:     "xai",
				Description: "Grok Beta - X.AI's flagship model",
				MaxTokens:   128000,
				ContextSize: 128000,
			},
		}
		return models, nil
	}

	var modelResp struct {
		Object string            `json:"object"`
		Data   []types.ModelInfo `json:"data"`
	}

	if err := json.Unmarshal(body, &modelResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return modelResp.Data, nil
}

// IsHealthy checks if the provider is healthy
func (p *XAIProvider) IsHealthy(ctx context.Context) error {
	// Simple health check by making a minimal request
	req := &types.ChatRequest{
		Messages: []types.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		MaxTokens: 10,
	}

	_, err := p.Chat(ctx, req)
	return err
}
