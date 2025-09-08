package google

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

// GoogleProvider implements the AIProvider interface for Google (Gemini)
type GoogleProvider struct {
	config     *types.ProviderConfig
	httpClient *http.Client
}

// NewGoogleProvider creates a new Google provider
func NewGoogleProvider(config *types.ProviderConfig) *GoogleProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	if config.DefaultModel == "" {
		config.DefaultModel = "gemini-pro"
	}

	return &GoogleProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetProviderName returns the provider name
func (p *GoogleProvider) GetProviderName() string {
	return "google"
}

// Chat sends a chat request to Google Gemini
func (p *GoogleProvider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.config.DefaultModel
	}

	// Convert to Gemini format
	geminiReq := map[string]interface{}{
		"contents": convertMessagesToGeminiFormat(req.Messages),
		"generationConfig": map[string]interface{}{
			"temperature":     req.Temperature,
			"maxOutputTokens": req.MaxTokens,
			"topP":            req.TopP,
		},
	}

	// Remove empty values
	if req.Temperature == 0 {
		geminiReq["generationConfig"].(map[string]interface{})["temperature"] = 0.7
	}
	if req.MaxTokens == 0 {
		geminiReq["generationConfig"].(map[string]interface{})["maxOutputTokens"] = 1024
	}
	if req.TopP == 0 {
		geminiReq["generationConfig"].(map[string]interface{})["topP"] = 0.8
	}

	jsonData, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.config.BaseURL, req.Model, p.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
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

	// Convert Gemini response to standard format
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to standard format
	var content string
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		content = geminiResp.Candidates[0].Content.Parts[0].Text
	}

	chatResp := &types.ChatResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []types.Choice{
			{
				Index: 0,
				Message: types.Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: geminiResp.Candidates[0].FinishReason,
			},
		},
		Usage: types.Usage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		},
	}

	return chatResp, nil
}

// convertMessagesToGeminiFormat converts standard messages to Gemini format
func convertMessagesToGeminiFormat(messages []types.Message) []map[string]interface{} {
	var contents []map[string]interface{}

	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		} else if msg.Role == "system" {
			role = "user" // Gemini doesn't have system role, treat as user
		}

		content := map[string]interface{}{
			"role": role,
			"parts": []map[string]interface{}{
				{
					"text": msg.Content,
				},
			},
		}
		contents = append(contents, content)
	}

	return contents
}

// GenerateText generates text based on a prompt
func (p *GoogleProvider) GenerateText(ctx context.Context, req *types.TextGenerationRequest) (*types.TextGenerationResponse, error) {
	// Convert to chat format for Google
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
func (p *GoogleProvider) EmbedText(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "embedding-001" // Default embedding model
	}

	// Convert to Gemini format
	geminiReq := map[string]interface{}{
		"model": req.Model,
		"content": map[string]interface{}{
			"parts": []map[string]interface{}{
				{
					"text": req.Input[0], // Gemini embedding API takes single text
				},
			},
		},
	}

	jsonData, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:embedContent?key=%s", p.config.BaseURL, req.Model, p.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
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

	// Convert Gemini response to standard format
	var geminiResp struct {
		Embedding struct {
			Values []float64 `json:"values"`
		} `json:"embedding"`
	}

	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	embedResp := &types.EmbeddingResponse{
		Object: "list",
		Data: []types.Embedding{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: geminiResp.Embedding.Values,
			},
		},
		Model: req.Model,
		Usage: types.Usage{
			PromptTokens: 1, // Gemini doesn't provide detailed usage for embeddings
		},
	}

	return embedResp, nil
}

// GetModelInfo returns information about available models
func (p *GoogleProvider) GetModelInfo(ctx context.Context) ([]types.ModelInfo, error) {
	url := fmt.Sprintf("%s/models?key=%s", p.config.BaseURL, p.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
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
				ID:          "gemini-pro",
				Object:      "model",
				Created:     1709251200,
				OwnedBy:     "google",
				Description: "Gemini Pro - General purpose model",
				MaxTokens:   32000,
				ContextSize: 32000,
			},
			{
				ID:          "gemini-pro-vision",
				Object:      "model",
				Created:     1709251200,
				OwnedBy:     "google",
				Description: "Gemini Pro Vision - Multimodal model",
				MaxTokens:   32000,
				ContextSize: 32000,
			},
			{
				ID:          "embedding-001",
				Object:      "model",
				Created:     1709251200,
				OwnedBy:     "google",
				Description: "Text Embedding 001 - Text embedding model",
				MaxTokens:   0,
				ContextSize: 0,
			},
		}
		return models, nil
	}

	var modelResp struct {
		Models []struct {
			Name                       string   `json:"name"`
			DisplayName                string   `json:"displayName"`
			Description                string   `json:"description"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &modelResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var models []types.ModelInfo
	for _, model := range modelResp.Models {
		models = append(models, types.ModelInfo{
			ID:          model.Name,
			Object:      "model",
			Created:     1709251200,
			OwnedBy:     "google",
			Description: model.Description,
			MaxTokens:   32000, // Default for Gemini models
			ContextSize: 32000,
		})
	}

	return models, nil
}

// IsHealthy checks if the provider is healthy
func (p *GoogleProvider) IsHealthy(ctx context.Context) error {
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
