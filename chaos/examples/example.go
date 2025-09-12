package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/chaos"
)

func main() {
	// Create a chaos manager with default providers
	manager := gateway.DefaultManager()

	// Initialize the manager
	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize chaos manager: %v", err)
	}
	defer manager.Cleanup(ctx)

	// Example 1: Kubernetes Pod Failure Experiment
	fmt.Println("=== Kubernetes Pod Failure Experiment ===")
	podFailureConfig := gateway.ExperimentConfig{
		Type:       gateway.ChaosTypeKubernetes,
		Experiment: gateway.PodFailure,
		Duration:   "5m",
		Target:     "my-app-pod",
		Parameters: map[string]interface{}{
			"namespace": "default",
			"selector":  "app=my-app",
		},
	}

	result, err := manager.StartExperiment(ctx, podFailureConfig)
	if err != nil {
		log.Printf("Failed to start pod failure experiment: %v", err)
	} else {
		fmt.Printf("Experiment started: %s\n", result.ID)

		// Check experiment status
		time.Sleep(2 * time.Second)
		status, err := manager.GetExperimentStatus(ctx, gateway.ChaosTypeKubernetes, result.ID)
		if err != nil {
			log.Printf("Failed to get experiment status: %v", err)
		} else {
			fmt.Printf("Experiment status: %s - %s\n", status.Status, status.Message)
		}

		// Stop the experiment
		time.Sleep(3 * time.Second)
		if err := manager.StopExperiment(ctx, gateway.ChaosTypeKubernetes, result.ID); err != nil {
			log.Printf("Failed to stop experiment: %v", err)
		} else {
			fmt.Println("Experiment stopped successfully")
		}
	}

	// Example 2: HTTP Latency Experiment
	fmt.Println("\n=== HTTP Latency Experiment ===")
	httpLatencyConfig := gateway.ExperimentConfig{
		Type:       gateway.ChaosTypeHTTP,
		Experiment: gateway.HTTPLatency,
		Duration:   "2m",
		Intensity:  0.5,
		Target:     "http://api.example.com",
		Parameters: map[string]interface{}{
			"latency_ms":  1000,
			"probability": 0.3,
		},
	}

	result, err = manager.StartExperiment(ctx, httpLatencyConfig)
	if err != nil {
		log.Printf("Failed to start HTTP latency experiment: %v", err)
	} else {
		fmt.Printf("HTTP latency experiment started: %s\n", result.ID)

		// Let it run for a bit
		time.Sleep(5 * time.Second)

		// Stop the experiment
		if err := manager.StopExperiment(ctx, gateway.ChaosTypeHTTP, result.ID); err != nil {
			log.Printf("Failed to stop HTTP experiment: %v", err)
		} else {
			fmt.Println("HTTP latency experiment stopped")
		}
	}

	// Example 3: Messaging Delay Experiment
	fmt.Println("\n=== Messaging Delay Experiment ===")
	messagingDelayConfig := gateway.ExperimentConfig{
		Type:       gateway.ChaosTypeMessaging,
		Experiment: gateway.MessageDelay,
		Duration:   "3m",
		Target:     "order-queue",
		Parameters: map[string]interface{}{
			"delay_ms":    5000,
			"probability": 0.2,
		},
	}

	result, err = manager.StartExperiment(ctx, messagingDelayConfig)
	if err != nil {
		log.Printf("Failed to start messaging delay experiment: %v", err)
	} else {
		fmt.Printf("Messaging delay experiment started: %s\n", result.ID)

		// Let it run for a bit
		time.Sleep(5 * time.Second)

		// Stop the experiment
		if err := manager.StopExperiment(ctx, gateway.ChaosTypeMessaging, result.ID); err != nil {
			log.Printf("Failed to stop messaging experiment: %v", err)
		} else {
			fmt.Println("Messaging delay experiment stopped")
		}
	}

	// List all experiments
	fmt.Println("\n=== Listing Experiments ===")
	providers := manager.GetAvailableProviders()
	for _, providerType := range providers {
		experiments, err := manager.ListExperiments(ctx, providerType)
		if err != nil {
			log.Printf("Failed to list experiments for %s: %v", providerType, err)
			continue
		}
		fmt.Printf("Provider %s has %d experiments\n", providerType, len(experiments))
		for _, exp := range experiments {
			fmt.Printf("  - %s: %s (%s)\n", exp.ID, exp.Status, exp.Message)
		}
	}
}
