package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/chaos"
	"github.com/anasamu/microservices-library-go/chaos/providers/http"
	"github.com/anasamu/microservices-library-go/chaos/providers/messaging"
)

// Integration tests require actual providers to be available
// These tests are skipped by default and should be run manually
// when the required infrastructure is available

func TestKubernetesProvider_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if not in Kubernetes environment
	if os.Getenv("KUBECONFIG") == "" && os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
		t.Skip("Skipping Kubernetes integration test - no cluster access")
	}

	// Note: In real integration tests, you would use the actual Kubernetes provider
	// For now, we'll skip this test

	t.Skip("Kubernetes integration test requires actual cluster access")
}

func TestHTTPProvider_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	manager := gateway.NewManager()

	// Register HTTP provider
	httpProvider := http.NewProvider()
	manager.RegisterProvider(gateway.ChaosTypeHTTP, httpProvider)

	ctx := context.Background()

	// Initialize
	err := manager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize manager: %v", err)
	}
	defer manager.Cleanup(ctx)

	// Test HTTP latency experiment
	config := gateway.ExperimentConfig{
		Type:       gateway.ChaosTypeHTTP,
		Experiment: gateway.HTTPLatency,
		Duration:   "30s",
		Intensity:  0.1,
		Target:     "http://httpbin.org/delay/1",
		Parameters: map[string]interface{}{
			"latency_ms":  500,
			"probability": 0.5,
		},
	}

	result, err := manager.StartExperiment(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start HTTP experiment: %v", err)
	}

	// Wait a bit
	time.Sleep(5 * time.Second)

	// Check status
	status, err := manager.GetExperimentStatus(ctx, gateway.ChaosTypeHTTP, result.ID)
	if err != nil {
		t.Fatalf("Failed to get experiment status: %v", err)
	}

	if status.Status != "running" {
		t.Errorf("Expected status 'running', got %s", status.Status)
	}

	// Stop experiment
	err = manager.StopExperiment(ctx, gateway.ChaosTypeHTTP, result.ID)
	if err != nil {
		t.Fatalf("Failed to stop experiment: %v", err)
	}

	// Verify stopped
	status, err = manager.GetExperimentStatus(ctx, gateway.ChaosTypeHTTP, result.ID)
	if err != nil {
		t.Fatalf("Failed to get final status: %v", err)
	}

	if status.Status != "stopped" {
		t.Errorf("Expected final status 'stopped', got %s", status.Status)
	}
}

func TestMessagingProvider_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	manager := gateway.NewManager()

	// Register messaging provider
	messagingProvider := messaging.NewProvider()
	manager.RegisterProvider(gateway.ChaosTypeMessaging, messagingProvider)

	ctx := context.Background()

	// Initialize
	err := manager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize manager: %v", err)
	}
	defer manager.Cleanup(ctx)

	// Test message delay experiment
	config := gateway.ExperimentConfig{
		Type:       gateway.ChaosTypeMessaging,
		Experiment: gateway.MessageDelay,
		Duration:   "30s",
		Target:     "test-topic",
		Parameters: map[string]interface{}{
			"delay_ms":    1000,
			"probability": 0.3,
		},
	}

	result, err := manager.StartExperiment(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start messaging experiment: %v", err)
	}

	// Wait a bit
	time.Sleep(5 * time.Second)

	// Check status
	status, err := manager.GetExperimentStatus(ctx, gateway.ChaosTypeMessaging, result.ID)
	if err != nil {
		t.Fatalf("Failed to get experiment status: %v", err)
	}

	if status.Status != "running" {
		t.Errorf("Expected status 'running', got %s", status.Status)
	}

	// Stop experiment
	err = manager.StopExperiment(ctx, gateway.ChaosTypeMessaging, result.ID)
	if err != nil {
		t.Fatalf("Failed to stop experiment: %v", err)
	}

	// Verify stopped
	status, err = manager.GetExperimentStatus(ctx, gateway.ChaosTypeMessaging, result.ID)
	if err != nil {
		t.Fatalf("Failed to get final status: %v", err)
	}

	if status.Status != "stopped" {
		t.Errorf("Expected final status 'stopped', got %s", status.Status)
	}
}

func TestManager_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	manager := gateway.DefaultManager()

	ctx := context.Background()

	// Initialize
	err := manager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize manager: %v", err)
	}
	defer manager.Cleanup(ctx)

	// Test multiple experiments
	experiments := []gateway.ExperimentConfig{
		{
			Type:       gateway.ChaosTypeHTTP,
			Experiment: gateway.HTTPLatency,
			Duration:   "10s",
			Parameters: map[string]interface{}{
				"latency_ms":  100,
				"probability": 0.1,
			},
		},
		{
			Type:       gateway.ChaosTypeMessaging,
			Experiment: gateway.MessageDelay,
			Duration:   "10s",
			Parameters: map[string]interface{}{
				"delay_ms":    500,
				"probability": 0.1,
			},
		},
	}

	var experimentIDs []string

	// Start all experiments
	for _, config := range experiments {
		result, err := manager.StartExperiment(ctx, config)
		if err != nil {
			t.Fatalf("Failed to start experiment: %v", err)
		}
		experimentIDs = append(experimentIDs, result.ID)
	}

	// Wait a bit
	time.Sleep(3 * time.Second)

	// Check all experiments are running
	for i, experimentID := range experimentIDs {
		status, err := manager.GetExperimentStatus(ctx, experiments[i].Type, experimentID)
		if err != nil {
			t.Fatalf("Failed to get experiment status: %v", err)
		}

		if status.Status != "running" {
			t.Errorf("Expected status 'running', got %s", status.Status)
		}
	}

	// Stop all experiments
	for i, experimentID := range experimentIDs {
		err := manager.StopExperiment(ctx, experiments[i].Type, experimentID)
		if err != nil {
			t.Fatalf("Failed to stop experiment: %v", err)
		}
	}

	// Verify all experiments are stopped
	for i, experimentID := range experimentIDs {
		status, err := manager.GetExperimentStatus(ctx, experiments[i].Type, experimentID)
		if err != nil {
			t.Fatalf("Failed to get final status: %v", err)
		}

		if status.Status != "stopped" {
			t.Errorf("Expected final status 'stopped', got %s", status.Status)
		}
	}
}
