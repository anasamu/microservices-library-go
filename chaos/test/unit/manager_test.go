package unit

import (
	"context"
	"fmt"
	"testing"

	"github.com/anasamu/microservices-library-go/chaos/gateway"
	"github.com/anasamu/microservices-library-go/chaos/test/mocks"
)

func TestManager_Initialize(t *testing.T) {
	manager := gateway.NewManager()
	mockProvider := mocks.NewMockProvider()

	manager.RegisterProvider(gateway.ChaosTypeKubernetes, mockProvider)

	ctx := context.Background()
	err := manager.Initialize(ctx)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !mockProvider.IsInitialized() {
		t.Error("Expected provider to be initialized")
	}
}

func TestManager_Initialize_Error(t *testing.T) {
	manager := gateway.NewManager()
	mockProvider := mocks.NewMockProvider()
	mockProvider.SetInitError(fmt.Errorf("init failed"))

	manager.RegisterProvider(gateway.ChaosTypeKubernetes, mockProvider)

	ctx := context.Background()
	err := manager.Initialize(ctx)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != "failed to initialize kubernetes provider: init failed" {
		t.Errorf("Expected specific error message, got %v", err)
	}
}

func TestManager_StartExperiment(t *testing.T) {
	manager := gateway.NewManager()
	mockProvider := mocks.NewMockProvider()

	manager.RegisterProvider(gateway.ChaosTypeKubernetes, mockProvider)

	ctx := context.Background()
	config := gateway.ExperimentConfig{
		Type:       gateway.ChaosTypeKubernetes,
		Experiment: gateway.PodFailure,
		Duration:   "5m",
		Target:     "test-pod",
	}

	result, err := manager.StartExperiment(ctx, config)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Error("Expected result, got nil")
	}

	if result.Status != "running" {
		t.Errorf("Expected status 'running', got %s", result.Status)
	}

	if mockProvider.GetExperimentCount() != 1 {
		t.Errorf("Expected 1 experiment, got %d", mockProvider.GetExperimentCount())
	}
}

func TestManager_StartExperiment_ProviderNotFound(t *testing.T) {
	manager := gateway.NewManager()

	ctx := context.Background()
	config := gateway.ExperimentConfig{
		Type:       gateway.ChaosTypeKubernetes,
		Experiment: gateway.PodFailure,
		Duration:   "5m",
		Target:     "test-pod",
	}

	result, err := manager.StartExperiment(ctx, config)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result, got %v", result)
	}

	if err.Error() != "provider for chaos type kubernetes not found" {
		t.Errorf("Expected specific error message, got %v", err)
	}
}

func TestManager_StopExperiment(t *testing.T) {
	manager := gateway.NewManager()
	mockProvider := mocks.NewMockProvider()

	manager.RegisterProvider(gateway.ChaosTypeKubernetes, mockProvider)

	ctx := context.Background()
	config := gateway.ExperimentConfig{
		Type:       gateway.ChaosTypeKubernetes,
		Experiment: gateway.PodFailure,
		Duration:   "5m",
		Target:     "test-pod",
	}

	// Start experiment first
	result, err := manager.StartExperiment(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start experiment: %v", err)
	}

	// Stop experiment
	err = manager.StopExperiment(ctx, gateway.ChaosTypeKubernetes, result.ID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check experiment status
	experiment, exists := mockProvider.GetExperiment(result.ID)
	if !exists {
		t.Error("Expected experiment to exist")
	}

	if experiment.Status != "stopped" {
		t.Errorf("Expected status 'stopped', got %s", experiment.Status)
	}
}

func TestManager_GetExperimentStatus(t *testing.T) {
	manager := gateway.NewManager()
	mockProvider := mocks.NewMockProvider()

	manager.RegisterProvider(gateway.ChaosTypeKubernetes, mockProvider)

	ctx := context.Background()
	config := gateway.ExperimentConfig{
		Type:       gateway.ChaosTypeKubernetes,
		Experiment: gateway.PodFailure,
		Duration:   "5m",
		Target:     "test-pod",
	}

	// Start experiment first
	result, err := manager.StartExperiment(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start experiment: %v", err)
	}

	// Get status
	status, err := manager.GetExperimentStatus(ctx, gateway.ChaosTypeKubernetes, result.ID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if status == nil {
		t.Error("Expected status, got nil")
	}

	if status.ID != result.ID {
		t.Errorf("Expected ID %s, got %s", result.ID, status.ID)
	}
}

func TestManager_ListExperiments(t *testing.T) {
	manager := gateway.NewManager()
	mockProvider := mocks.NewMockProvider()

	manager.RegisterProvider(gateway.ChaosTypeKubernetes, mockProvider)

	ctx := context.Background()
	config := gateway.ExperimentConfig{
		Type:       gateway.ChaosTypeKubernetes,
		Experiment: gateway.PodFailure,
		Duration:   "5m",
		Target:     "test-pod",
	}

	// Start multiple experiments
	for i := 0; i < 3; i++ {
		_, err := manager.StartExperiment(ctx, config)
		if err != nil {
			t.Fatalf("Failed to start experiment %d: %v", i, err)
		}
	}

	// List experiments
	experiments, err := manager.ListExperiments(ctx, gateway.ChaosTypeKubernetes)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(experiments) != 3 {
		t.Errorf("Expected 3 experiments, got %d", len(experiments))
	}
}

func TestManager_GetAvailableProviders(t *testing.T) {
	manager := gateway.NewManager()
	mockProvider := mocks.NewMockProvider()

	manager.RegisterProvider(gateway.ChaosTypeKubernetes, mockProvider)
	manager.RegisterProvider(gateway.ChaosTypeHTTP, mockProvider)

	providers := manager.GetAvailableProviders()

	if len(providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providers))
	}

	// Check that both providers are in the list
	hasKubernetes := false
	hasHTTP := false
	for _, provider := range providers {
		if provider == gateway.ChaosTypeKubernetes {
			hasKubernetes = true
		}
		if provider == gateway.ChaosTypeHTTP {
			hasHTTP = true
		}
	}

	if !hasKubernetes {
		t.Error("Expected kubernetes provider to be available")
	}

	if !hasHTTP {
		t.Error("Expected http provider to be available")
	}
}

func TestManager_Cleanup(t *testing.T) {
	manager := gateway.NewManager()
	mockProvider := mocks.NewMockProvider()

	manager.RegisterProvider(gateway.ChaosTypeKubernetes, mockProvider)

	ctx := context.Background()

	// Initialize first
	err := manager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Cleanup
	err = manager.Cleanup(ctx)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if mockProvider.IsInitialized() {
		t.Error("Expected provider to be cleaned up")
	}
}

func TestDefaultManager(t *testing.T) {
	manager := gateway.DefaultManager()

	providers := manager.GetAvailableProviders()

	// Should have all three default providers
	if len(providers) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(providers))
	}

	// Check that all expected providers are present
	expectedProviders := []gateway.ChaosType{
		gateway.ChaosTypeKubernetes,
		gateway.ChaosTypeHTTP,
		gateway.ChaosTypeMessaging,
	}

	for _, expected := range expectedProviders {
		found := false
		for _, provider := range providers {
			if provider == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected provider %s not found", expected)
		}
	}
}
