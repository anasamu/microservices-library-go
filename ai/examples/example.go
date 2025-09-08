package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anasamu/microservices-library-go/ai/gateway"
	"github.com/anasamu/microservices-library-go/ai/types"
)

func main() {
	// Create AI manager
	manager := gateway.NewAIManager()

	// Example 1: OpenAI Provider
	fmt.Println("=== OpenAI Provider Example ===")
	openaiConfig := &types.ProviderConfig{
		Name:         "openai",
		APIKey:       os.Getenv("OPENAI_API_KEY"),
		DefaultModel: "gpt-4",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
	}

	if openaiConfig.APIKey != "" {
		if err := manager.AddProvider(openaiConfig); err != nil {
			log.Printf("Failed to add OpenAI provider: %v", err)
		} else {
			fmt.Println("OpenAI provider added successfully")

			// Test chat
			chatReq := &types.ChatRequest{
				Messages: []types.Message{
					{
						Role:    "user",
						Content: "Hello! Can you tell me a short joke?",
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
				log.Printf("Chat error: %v", err)
			} else if resp.Error != nil {
				log.Printf("API error: %v", resp.Error)
			} else {
				fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
			}
		}
	} else {
		fmt.Println("OPENAI_API_KEY not set, skipping OpenAI example")
	}

	// Example 2: Anthropic Provider
	fmt.Println("\n=== Anthropic Provider Example ===")
	anthropicConfig := &types.ProviderConfig{
		Name:         "anthropic",
		APIKey:       os.Getenv("ANTHROPIC_API_KEY"),
		DefaultModel: "claude-3-sonnet-20240229",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
	}

	if anthropicConfig.APIKey != "" {
		if err := manager.AddProvider(anthropicConfig); err != nil {
			log.Printf("Failed to add Anthropic provider: %v", err)
		} else {
			fmt.Println("Anthropic provider added successfully")

			// Test text generation
			textReq := &types.TextGenerationRequest{
				Prompt:      "Write a haiku about programming",
				Model:       "claude-3-sonnet-20240229",
				Temperature: 0.7,
				MaxTokens:   50,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := manager.GenerateText(ctx, "anthropic", textReq)
			if err != nil {
				log.Printf("Text generation error: %v", err)
			} else if resp.Error != nil {
				log.Printf("API error: %v", resp.Error)
			} else {
				fmt.Printf("Generated text: %s\n", resp.Choices[0].Text)
			}
		}
	} else {
		fmt.Println("ANTHROPIC_API_KEY not set, skipping Anthropic example")
	}

	// Example 3: Google Provider
	fmt.Println("\n=== Google Provider Example ===")
	googleConfig := &types.ProviderConfig{
		Name:         "google",
		APIKey:       os.Getenv("GOOGLE_API_KEY"),
		DefaultModel: "gemini-pro",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
	}

	if googleConfig.APIKey != "" {
		if err := manager.AddProvider(googleConfig); err != nil {
			log.Printf("Failed to add Google provider: %v", err)
		} else {
			fmt.Println("Google provider added successfully")

			// Test chat
			chatReq := &types.ChatRequest{
				Messages: []types.Message{
					{
						Role:    "user",
						Content: "What is the capital of France?",
					},
				},
				Model:       "gemini-pro",
				Temperature: 0.3,
				MaxTokens:   50,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := manager.Chat(ctx, "google", chatReq)
			if err != nil {
				log.Printf("Chat error: %v", err)
			} else if resp.Error != nil {
				log.Printf("API error: %v", resp.Error)
			} else {
				fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
			}
		}
	} else {
		fmt.Println("GOOGLE_API_KEY not set, skipping Google example")
	}

	// Example 4: DeepSeek Provider
	fmt.Println("\n=== DeepSeek Provider Example ===")
	deepseekConfig := &types.ProviderConfig{
		Name:         "deepseek",
		APIKey:       os.Getenv("DEEPSEEK_API_KEY"),
		DefaultModel: "deepseek-chat",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
	}

	if deepseekConfig.APIKey != "" {
		if err := manager.AddProvider(deepseekConfig); err != nil {
			log.Printf("Failed to add DeepSeek provider: %v", err)
		} else {
			fmt.Println("DeepSeek provider added successfully")

			// Test chat
			chatReq := &types.ChatRequest{
				Messages: []types.Message{
					{
						Role:    "user",
						Content: "Explain quantum computing in simple terms",
					},
				},
				Model:       "deepseek-chat",
				Temperature: 0.5,
				MaxTokens:   150,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := manager.Chat(ctx, "deepseek", chatReq)
			if err != nil {
				log.Printf("Chat error: %v", err)
			} else if resp.Error != nil {
				log.Printf("API error: %v", resp.Error)
			} else {
				fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
			}
		}
	} else {
		fmt.Println("DEEPSEEK_API_KEY not set, skipping DeepSeek example")
	}

	// Example 5: X.AI Provider
	fmt.Println("\n=== X.AI Provider Example ===")
	xaiConfig := &types.ProviderConfig{
		Name:         "xai",
		APIKey:       os.Getenv("XAI_API_KEY"),
		DefaultModel: "grok-beta",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
	}

	if xaiConfig.APIKey != "" {
		if err := manager.AddProvider(xaiConfig); err != nil {
			log.Printf("Failed to add X.AI provider: %v", err)
		} else {
			fmt.Println("X.AI provider added successfully")

			// Test chat
			chatReq := &types.ChatRequest{
				Messages: []types.Message{
					{
						Role:    "user",
						Content: "What's the latest in AI technology?",
					},
				},
				Model:       "grok-beta",
				Temperature: 0.7,
				MaxTokens:   100,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := manager.Chat(ctx, "xai", chatReq)
			if err != nil {
				log.Printf("Chat error: %v", err)
			} else if resp.Error != nil {
				log.Printf("API error: %v", resp.Error)
			} else {
				fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
			}
		}
	} else {
		fmt.Println("XAI_API_KEY not set, skipping X.AI example")
	}

	// Example 6: Health Check
	fmt.Println("\n=== Health Check Example ===")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	healthStatus, err := manager.HealthCheck(ctx)
	if err != nil {
		log.Printf("Health check error: %v", err)
	} else {
		for provider, status := range healthStatus {
			fmt.Printf("Provider %s: %s\n", provider,
				map[bool]string{true: "Healthy", false: "Unhealthy"}[status.Healthy])
			if !status.Healthy {
				fmt.Printf("  Error: %s\n", status.Message)
			}
		}
	}

	// Example 7: Get All Model Info
	fmt.Println("\n=== Model Information Example ===")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	allModels, err := manager.GetAllModelInfo(ctx)
	if err != nil {
		log.Printf("Get model info error: %v", err)
	} else {
		for provider, models := range allModels {
			fmt.Printf("Provider %s has %d models:\n", provider, len(models))
			for _, model := range models {
				fmt.Printf("  - %s: %s\n", model.ID, model.Description)
			}
		}
	}

	// Example 8: Statistics
	fmt.Println("\n=== Statistics Example ===")
	stats := manager.GetStats()
	for provider, stat := range stats {
		fmt.Printf("Provider %s:\n", provider)
		fmt.Printf("  Total Requests: %d\n", stat.TotalRequests)
		fmt.Printf("  Total Tokens: %d\n", stat.TotalTokens)
		fmt.Printf("  Success Rate: %.2f%%\n", stat.SuccessRate*100)
		fmt.Printf("  Average Latency: %.2f ms\n", stat.AvgLatency)
		fmt.Printf("  Last Used: %s\n", stat.LastUsed.Format(time.RFC3339))
	}

	// Example 9: Fallback Example
	fmt.Println("\n=== Fallback Example ===")
	if len(manager.ListProviders()) > 1 {
		chatReq := &types.ChatRequest{
			Messages: []types.Message{
				{
					Role:    "user",
					Content: "This is a test message for fallback",
				},
			},
			Temperature: 0.7,
			MaxTokens:   50,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Try with fallback (will try first provider, then others if it fails)
		providers := manager.ListProviders()
		resp, err := manager.ChatWithFallback(ctx, providers[0], chatReq)
		if err != nil {
			log.Printf("Fallback chat error: %v", err)
		} else if resp.Error != nil {
			log.Printf("Fallback API error: %v", resp.Error)
		} else {
			fmt.Printf("Fallback response: %s\n", resp.Choices[0].Message.Content)
		}
	}

	fmt.Println("\n=== Example completed ===")
}
