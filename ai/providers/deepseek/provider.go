package deepseek

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

// DeepSeekProvider implements the AIProvider interface for DeepSeek
type DeepSeekProvider struct {
	config     *types.ProviderConfig
	httpClient *http.Client
}

// NewDeepSeekProvider creates a new DeepSeek provider
func NewDeepSeekProvider(config *types.ProviderConfig) *DeepSeekProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.deepseek.com/v1"
	}
	if config.DefaultModel == "" {
		config.DefaultModel = "deepseek-chat"
	}

	return &DeepSeekProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetProviderName returns the provider name
func (p *DeepSeekProvider) GetProviderName() string {
	return "deepseek"
}

// Chat sends a chat request to DeepSeek
func (p *DeepSeekProvider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.config.DefaultModel
	}

	// Convert to DeepSeek format
	deepseekReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
		"stream":      req.Stream,
		"top_p":       req.TopP,
	}

	// Remove empty values
	if req.Temperature == 0 {
		delete(deepseekReq, "temperature")
	}
	if req.MaxTokens == 0 {
		delete(deepseekReq, "max_tokens")
	}
	if req.TopP == 0 {
		delete(deepseekReq, "top_p")
	}

	jsonData, err := json.Marshal(deepseekReq)
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
func (p *DeepSeekProvider) GenerateText(ctx context.Context, req *types.TextGenerationRequest) (*types.TextGenerationResponse, error) {
	// Convert to chat format for DeepSeek
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
func (p *DeepSeekProvider) EmbedText(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "deepseek-embedding" // Default embedding model
	}

	// Convert to DeepSeek format
	deepseekReq := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	}

	jsonData, err := json.Marshal(deepseekReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/embeddings", bytes.NewBuffer(jsonData))
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
		return &types.EmbeddingResponse{Error: errorResp.Error}, nil
	}

	var embedResp types.EmbeddingResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &embedResp, nil
}

// GetModelInfo returns information about available models
func (p *DeepSeekProvider) GetModelInfo(ctx context.Context) ([]types.ModelInfo, error) {
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
				ID:          "deepseek-chat",
				Object:      "model",
				Created:     1709251200,
				OwnedBy:     "deepseek",
				Description: "DeepSeek Chat - General purpose chat model",
				MaxTokens:   32000,
				ContextSize: 32000,
			},
			{
				ID:          "deepseek-coder",
				Object:      "model",
				Created:     1709251200,
				OwnedBy:     "deepseek",
				Description: "DeepSeek Coder - Specialized for coding tasks",
				MaxTokens:   32000,
				ContextSize: 32000,
			},
			{
				ID:          "deepseek-embedding",
				Object:      "model",
				Created:     1709251200,
				OwnedBy:     "deepseek",
				Description: "DeepSeek Embedding - Text embedding model",
				MaxTokens:   0,
				ContextSize: 0,
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
func (p *DeepSeekProvider) IsHealthy(ctx context.Context) error {
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
