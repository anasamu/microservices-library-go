package anthropic

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

// AnthropicProvider implements the AIProvider interface for Anthropic
type AnthropicProvider struct {
	config     *types.ProviderConfig
	httpClient *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(config *types.ProviderConfig) *AnthropicProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com/v1"
	}
	if config.DefaultModel == "" {
		config.DefaultModel = "claude-3-sonnet-20240229"
	}

	return &AnthropicProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetProviderName returns the provider name
func (p *AnthropicProvider) GetProviderName() string {
	return "anthropic"
}

// Chat sends a chat request to Anthropic
func (p *AnthropicProvider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.config.DefaultModel
	}

	// Convert to Anthropic format
	anthropicReq := map[string]interface{}{
		"model":       req.Model,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
		"messages":    req.Messages,
	}

	// Remove empty values
	if req.Temperature == 0 {
		delete(anthropicReq, "temperature")
	}
	if req.MaxTokens == 0 {
		anthropicReq["max_tokens"] = 1024 // Default for Anthropic
	}
	if req.TopP == 0 {
		delete(anthropicReq, "top_p")
	}

	jsonData, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
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

	// Convert Anthropic response to standard format
	var anthropicResp struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Model        string `json:"model"`
		StopReason   string `json:"stop_reason"`
		StopSequence string `json:"stop_sequence"`
		Usage        struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to standard format
	var content string
	for _, c := range anthropicResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	chatResp := &types.ChatResponse{
		ID:      anthropicResp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   anthropicResp.Model,
		Choices: []types.Choice{
			{
				Index: 0,
				Message: types.Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: anthropicResp.StopReason,
			},
		},
		Usage: types.Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}

	return chatResp, nil
}

// GenerateText generates text based on a prompt
func (p *AnthropicProvider) GenerateText(ctx context.Context, req *types.TextGenerationRequest) (*types.TextGenerationResponse, error) {
	// Convert to chat format for Anthropic
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
func (p *AnthropicProvider) EmbedText(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	// Anthropic doesn't have a direct embedding API, so we'll use a workaround
	// This is a placeholder implementation
	return nil, fmt.Errorf("embedding not directly supported by Anthropic API")
}

// GetModelInfo returns information about available models
func (p *AnthropicProvider) GetModelInfo(ctx context.Context) ([]types.ModelInfo, error) {
	// Anthropic doesn't have a models endpoint, so we return known models
	models := []types.ModelInfo{
		{
			ID:          "claude-3-opus-20240229",
			Object:      "model",
			Created:     1709251200,
			OwnedBy:     "anthropic",
			Description: "Claude 3 Opus - Most powerful model",
			MaxTokens:   200000,
			ContextSize: 200000,
		},
		{
			ID:          "claude-3-sonnet-20240229",
			Object:      "model",
			Created:     1709251200,
			OwnedBy:     "anthropic",
			Description: "Claude 3 Sonnet - Balanced performance and speed",
			MaxTokens:   200000,
			ContextSize: 200000,
		},
		{
			ID:          "claude-3-haiku-20240307",
			Object:      "model",
			Created:     1709251200,
			OwnedBy:     "anthropic",
			Description: "Claude 3 Haiku - Fastest and most cost-effective",
			MaxTokens:   200000,
			ContextSize: 200000,
		},
	}

	return models, nil
}

// IsHealthy checks if the provider is healthy
func (p *AnthropicProvider) IsHealthy(ctx context.Context) error {
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
