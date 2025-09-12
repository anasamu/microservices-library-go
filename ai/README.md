# AI Microservices Library

A comprehensive Go library for integrating multiple AI providers into microservices applications. This library provides a unified interface for working with various AI models including GPT-4, GPT-5, Claude, Grok, DeepSeek, and Gemini.

## Features

- **Multi-Provider Support**: Unified interface for OpenAI, Anthropic, X.AI, DeepSeek, and Google AI
- **Automatic Fallback**: Built-in fallback mechanism when primary providers fail
- **Health Monitoring**: Real-time health checks and statistics tracking
- **Type Safety**: Strongly typed interfaces and responses
- **Context Support**: Full context.Context support for timeouts and cancellation
- **Statistics**: Request tracking, latency monitoring, and success rate metrics
- **Easy Configuration**: Simple configuration management for multiple providers

## Supported Providers

| Provider | Models | Chat | Text Generation | Embeddings |
|----------|--------|------|-----------------|------------|
| OpenAI | GPT-4, GPT-4-turbo, GPT-3.5-turbo | ✅ | ✅ | ✅ |
| Anthropic | Claude-3-opus, Claude-3-sonnet, Claude-3-haiku | ✅ | ✅ | ❌ |
| X.AI | Grok-beta | ✅ | ✅ | ❌ |
| DeepSeek | DeepSeek-chat, DeepSeek-coder | ✅ | ✅ | ✅ |
| Google | Gemini-pro, Gemini-pro-vision | ✅ | ✅ | ✅ |

## Installation

```bash
go get github.com/anasamu/microservices-library-go/ai
```

## Quick Start

### 1. Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/anasamu/microservices-library-go/ai"
    "github.com/anasamu/microservices-library-go/ai/types"
)

func main() {
    // Create AI manager
    manager := ai.NewAIManager()

    // Add OpenAI provider
    openaiConfig := &types.ProviderConfig{
        Name:         "openai",
        APIKey:       "your-openai-api-key",
        DefaultModel: "gpt-4",
        Timeout:      30 * time.Second,
        MaxRetries:   3,
    }

    if err := manager.AddProvider(openaiConfig); err != nil {
        log.Fatal(err)
    }

    // Send a chat request
    chatReq := &types.ChatRequest{
        Messages: []types.Message{
            {
                Role:    "user",
                Content: "Hello! How are you?",
            },
        },
        Model:       "gpt-4",
        Temperature: 0.7,
        MaxTokens:   100,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    resp, err := manager.Chat(ctx, "openai", chatReq)
    if err != nil {
        log.Fatal(err)
    }

    if resp.Error != nil {
        log.Printf("API error: %v", resp.Error)
    } else {
        fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
    }
}
```

### 2. Multiple Providers with Fallback

```go
// Add multiple providers
providers := []*types.ProviderConfig{
    {
        Name:         "openai",
        APIKey:       "your-openai-api-key",
        DefaultModel: "gpt-4",
    },
    {
        Name:         "anthropic",
        APIKey:       "your-anthropic-api-key",
        DefaultModel: "claude-3-sonnet-20240229",
    },
    {
        Name:         "google",
        APIKey:       "your-google-api-key",
        DefaultModel: "gemini-pro",
    },
}

for _, config := range providers {
    if err := manager.AddProvider(config); err != nil {
        log.Printf("Failed to add provider %s: %v", config.Name, err)
    }
}

// Use fallback mechanism
chatReq := &types.ChatRequest{
    Messages: []types.Message{
        {
            Role:    "user",
            Content: "Explain quantum computing",
        },
    },
    Temperature: 0.7,
    MaxTokens:   200,
}

// This will try OpenAI first, then fallback to other providers if it fails
resp, err := manager.ChatWithFallback(ctx, "openai", chatReq)
```

### 3. Text Generation

```go
textReq := &types.TextGenerationRequest{
    Prompt:      "Write a short story about a robot",
    Model:       "gpt-4",
    Temperature: 0.8,
    MaxTokens:   300,
}

resp, err := manager.GenerateText(ctx, "openai", textReq)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Generated text: %s\n", resp.Choices[0].Text)
```

### 4. Embeddings

```go
embedReq := &types.EmbeddingRequest{
    Input: []string{"Hello world", "How are you?"},
    Model: "text-embedding-ada-002",
}

resp, err := manager.EmbedText(ctx, "openai", embedReq)
if err != nil {
    log.Fatal(err)
}

for i, embedding := range resp.Data {
    fmt.Printf("Embedding %d: %v\n", i, embedding.Embedding[:5]) // Show first 5 dimensions
}
```

## Configuration

### Environment Variables

Set your API keys as environment variables:

```bash
export OPENAI_API_KEY="your-openai-api-key"
export ANTHROPIC_API_KEY="your-anthropic-api-key"
export GOOGLE_API_KEY="your-google-api-key"
export DEEPSEEK_API_KEY="your-deepseek-api-key"
export XAI_API_KEY="your-xai-api-key"
```

### Provider Configuration

```go
config := &types.ProviderConfig{
    Name:         "openai",                    // Provider name
    APIKey:       "your-api-key",             // API key
    BaseURL:      "https://api.openai.com/v1", // Custom base URL (optional)
    Timeout:      30 * time.Second,           // Request timeout
    MaxRetries:   3,                          // Maximum retry attempts
    Headers: map[string]string{               // Custom headers
        "User-Agent": "MyApp/1.0",
    },
    Models:        []string{"gpt-4", "gpt-3.5-turbo"}, // Available models
    DefaultModel:  "gpt-4",                   // Default model to use
}
```

## Health Monitoring

### Health Check

```go
// Check health of all providers
healthStatus, err := manager.HealthCheck(ctx)
if err != nil {
    log.Fatal(err)
}

for provider, status := range healthStatus {
    fmt.Printf("Provider %s: %s\n", provider, 
        map[bool]string{true: "Healthy", false: "Unhealthy"}[status.Healthy])
    if !status.Healthy {
        fmt.Printf("  Error: %s\n", status.Message)
    }
}
```

### Statistics

```go
// Get statistics for all providers
stats := manager.GetStats()
for provider, stat := range stats {
    fmt.Printf("Provider %s:\n", provider)
    fmt.Printf("  Total Requests: %d\n", stat.TotalRequests)
    fmt.Printf("  Total Tokens: %d\n", stat.TotalTokens)
    fmt.Printf("  Success Rate: %.2f%%\n", stat.SuccessRate*100)
    fmt.Printf("  Average Latency: %.2f ms\n", stat.AvgLatency)
    fmt.Printf("  Last Used: %s\n", stat.LastUsed.Format(time.RFC3339))
}
```

## Advanced Usage

### Custom Headers

```go
config := &types.ProviderConfig{
    Name:    "openai",
    APIKey:  "your-api-key",
    Headers: map[string]string{
        "X-Custom-Header": "custom-value",
        "User-Agent":      "MyApp/1.0",
    },
}
```

### Model Information

```go
// Get available models for a provider
models, err := manager.GetModelInfo(ctx, "openai")
if err != nil {
    log.Fatal(err)
}

for _, model := range models {
    fmt.Printf("Model: %s\n", model.ID)
    fmt.Printf("Description: %s\n", model.Description)
    fmt.Printf("Max Tokens: %d\n", model.MaxTokens)
}

// Get models for all providers
allModels, err := manager.GetAllModelInfo(ctx)
if err != nil {
    log.Fatal(err)
}

for provider, models := range allModels {
    fmt.Printf("Provider %s has %d models\n", provider, len(models))
}
```

### Error Handling

```go
resp, err := manager.Chat(ctx, "openai", chatReq)
if err != nil {
    // Network or configuration error
    log.Printf("Request failed: %v", err)
    return
}

if resp.Error != nil {
    // API error (rate limit, invalid request, etc.)
    log.Printf("API error: %s (Type: %s)", resp.Error.Message, resp.Error.Type)
    return
}

// Success
fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
```

## Running Examples

1. Set your API keys as environment variables
2. Run the example:

```bash
cd examples
go run example.go
```

## API Reference

### Types

#### ChatRequest
```go
type ChatRequest struct {
    Messages          []Message `json:"messages"`
    Model             string    `json:"model"`
    Temperature       float64   `json:"temperature,omitempty"`
    MaxTokens         int       `json:"max_tokens,omitempty"`
    Stream            bool      `json:"stream,omitempty"`
    TopP              float64   `json:"top_p,omitempty"`
    FrequencyPenalty  float64   `json:"frequency_penalty,omitempty"`
    PresencePenalty   float64   `json:"presence_penalty,omitempty"`
}
```

#### TextGenerationRequest
```go
type TextGenerationRequest struct {
    Prompt      string   `json:"prompt"`
    Model       string   `json:"model"`
    Temperature float64  `json:"temperature,omitempty"`
    MaxTokens   int      `json:"max_tokens,omitempty"`
    TopP        float64  `json:"top_p,omitempty"`
    Stop        []string `json:"stop,omitempty"`
}
```

#### EmbeddingRequest
```go
type EmbeddingRequest struct {
    Input []string `json:"input"`
    Model string   `json:"model"`
}
```

### Manager Methods

- `AddProvider(config *ProviderConfig) error` - Add a new AI provider
- `RemoveProvider(name string) error` - Remove a provider
- `GetProvider(name string) (AIProvider, error)` - Get a specific provider
- `ListProviders() []string` - List all provider names
- `Chat(ctx, providerName, req) (*ChatResponse, error)` - Send chat request
- `GenerateText(ctx, providerName, req) (*TextGenerationResponse, error)` - Generate text
- `EmbedText(ctx, providerName, req) (*EmbeddingResponse, error)` - Generate embeddings
- `ChatWithFallback(ctx, primaryProvider, req) (*ChatResponse, error)` - Chat with fallback
- `GenerateTextWithFallback(ctx, primaryProvider, req) (*TextGenerationResponse, error)` - Text generation with fallback
- `HealthCheck(ctx) (map[string]*HealthStatus, error)` - Check provider health
- `GetStats() map[string]*ProviderStats` - Get provider statistics
- `GetModelInfo(ctx, providerName) ([]ModelInfo, error)` - Get model information

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Check the examples directory for usage patterns
- Review the API documentation

## Changelog

### v1.0.0
- Initial release
- Support for OpenAI, Anthropic, X.AI, DeepSeek, and Google AI
- Unified interface for all providers
- Health monitoring and statistics
- Automatic fallback mechanism
- Comprehensive examples and documentation
