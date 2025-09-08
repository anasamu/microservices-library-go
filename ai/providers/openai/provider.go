package openai

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

// OpenAIProvider implements the AIProvider interface for OpenAI
type OpenAIProvider struct {
	config     *types.ProviderConfig
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *types.ProviderConfig) *OpenAIProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.DefaultModel == "" {
		config.DefaultModel = "gpt-4"
	}

	return &OpenAIProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetProviderName returns the provider name
func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
}

// Chat sends a chat request to OpenAI
func (p *OpenAIProvider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.config.DefaultModel
	}

	// Convert to OpenAI format
	openaiReq := map[string]interface{}{
		"model":             req.Model,
		"messages":          req.Messages,
		"temperature":       req.Temperature,
		"max_tokens":        req.MaxTokens,
		"stream":            req.Stream,
		"top_p":             req.TopP,
		"frequency_penalty": req.FrequencyPenalty,
		"presence_penalty":  req.PresencePenalty,
	}

	// Remove empty values
	if req.Temperature == 0 {
		delete(openaiReq, "temperature")
	}
	if req.MaxTokens == 0 {
		delete(openaiReq, "max_tokens")
	}
	if req.TopP == 0 {
		delete(openaiReq, "top_p")
	}
	if req.FrequencyPenalty == 0 {
		delete(openaiReq, "frequency_penalty")
	}
	if req.PresencePenalty == 0 {
		delete(openaiReq, "presence_penalty")
	}

	jsonData, err := json.Marshal(openaiReq)
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
func (p *OpenAIProvider) GenerateText(ctx context.Context, req *types.TextGenerationRequest) (*types.TextGenerationResponse, error) {
	if req.Model == "" {
		req.Model = p.config.DefaultModel
	}

	// Convert to OpenAI format
	openaiReq := map[string]interface{}{
		"model":       req.Model,
		"prompt":      req.Prompt,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
		"top_p":       req.TopP,
		"stop":        req.Stop,
	}

	// Remove empty values
	if req.Temperature == 0 {
		delete(openaiReq, "temperature")
	}
	if req.MaxTokens == 0 {
		delete(openaiReq, "max_tokens")
	}
	if req.TopP == 0 {
		delete(openaiReq, "top_p")
	}
	if len(req.Stop) == 0 {
		delete(openaiReq, "stop")
	}

	jsonData, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/completions", bytes.NewBuffer(jsonData))
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
		return &types.TextGenerationResponse{Error: errorResp.Error}, nil
	}

	var textResp types.TextGenerationResponse
	if err := json.Unmarshal(body, &textResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &textResp, nil
}

// EmbedText generates embeddings for text
func (p *OpenAIProvider) EmbedText(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "text-embedding-ada-002" // Default embedding model
	}

	// Convert to OpenAI format
	openaiReq := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	}

	jsonData, err := json.Marshal(openaiReq)
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
func (p *OpenAIProvider) GetModelInfo(ctx context.Context) ([]types.ModelInfo, error) {
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
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
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
func (p *OpenAIProvider) IsHealthy(ctx context.Context) error {
	// Simple health check by getting model info
	_, err := p.GetModelInfo(ctx)
	return err
}
