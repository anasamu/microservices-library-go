package kubernetes

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	chaosmesh "chaos-mesh.org/api/v1alpha1"
	"github.com/anasamu/microservices-library-go/chaos/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Provider implements chaos engineering for Kubernetes using Chaos Mesh
type Provider struct {
	experiments map[string]*types.ExperimentResult
	mu          sync.RWMutex
	client      *kubernetes.Clientset
	chaosClient *chaosmesh.ChaosClient
}

// NewProvider creates a new Kubernetes chaos provider
func NewProvider() *Provider {
	return &Provider{
		experiments: make(map[string]*types.ExperimentResult),
	}
}

// Initialize initializes the Kubernetes provider
func (p *Provider) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Initialize Kubernetes client
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig file
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			return fmt.Errorf("failed to create kubernetes config: %w", err)
		}
	}

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	p.client = client

	// Initialize Chaos Mesh client
	chaosClient, err := chaosmesh.NewForConfig(kubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create chaos mesh client: %w", err)
	}
	p.chaosClient = chaosClient

	return nil
}

// StartExperiment starts a Kubernetes chaos experiment
func (p *Provider) StartExperiment(ctx context.Context, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	experimentID := generateExperimentID()

	switch config.Experiment {
	case types.PodFailure:
		return p.startPodFailureExperiment(ctx, experimentID, config)
	case types.NetworkLatency:
		return p.startNetworkLatencyExperiment(ctx, experimentID, config)
	case types.CPUStress:
		return p.startCPUStressExperiment(ctx, experimentID, config)
	case types.MemoryStress:
		return p.startMemoryStressExperiment(ctx, experimentID, config)
	default:
		return nil, fmt.Errorf("unsupported experiment type: %s", config.Experiment)
	}
}

// StopExperiment stops a Kubernetes chaos experiment
func (p *Provider) StopExperiment(ctx context.Context, experimentID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	experiment, exists := p.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	// Delete the Chaos Mesh experiment
	namespace := "chaos-mesh"
	if ns, ok := experiment.Metrics["namespace"].(string); ok {
		namespace = ns
	}

	switch experiment.Metrics["type"] {
	case "PodChaos":
		err := p.chaosClient.ChaosV1alpha1().PodChaos(namespace).Delete(ctx, experimentID, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete PodChaos: %w", err)
		}
	case "NetworkChaos":
		err := p.chaosClient.ChaosV1alpha1().NetworkChaos(namespace).Delete(ctx, experimentID, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete NetworkChaos: %w", err)
		}
	case "StressChaos":
		err := p.chaosClient.ChaosV1alpha1().StressChaos(namespace).Delete(ctx, experimentID, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete StressChaos: %w", err)
		}
	}

	// Update experiment status
	experiment.Status = "stopped"
	experiment.EndTime = time.Now().Format(time.RFC3339)

	return nil
}

// GetExperimentStatus gets the status of a Kubernetes chaos experiment
func (p *Provider) GetExperimentStatus(ctx context.Context, experimentID string) (*types.ExperimentResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	experiment, exists := p.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	// Check actual status from Kubernetes
	namespace := "chaos-mesh"
	if ns, ok := experiment.Metrics["namespace"].(string); ok {
		namespace = ns
	}

	switch experiment.Metrics["type"] {
	case "PodChaos":
		podChaos, err := p.chaosClient.ChaosV1alpha1().PodChaos(namespace).Get(ctx, experimentID, metav1.GetOptions{})
		if err != nil {
			experiment.Status = "error"
			experiment.Message = err.Error()
		} else {
			experiment.Status = string(podChaos.Status.Phase)
		}
	case "NetworkChaos":
		networkChaos, err := p.chaosClient.ChaosV1alpha1().NetworkChaos(namespace).Get(ctx, experimentID, metav1.GetOptions{})
		if err != nil {
			experiment.Status = "error"
			experiment.Message = err.Error()
		} else {
			experiment.Status = string(networkChaos.Status.Phase)
		}
	case "StressChaos":
		stressChaos, err := p.chaosClient.ChaosV1alpha1().StressChaos(namespace).Get(ctx, experimentID, metav1.GetOptions{})
		if err != nil {
			experiment.Status = "error"
			experiment.Message = err.Error()
		} else {
			experiment.Status = string(stressChaos.Status.Phase)
		}
	}

	return experiment, nil
}

// ListExperiments lists all Kubernetes chaos experiments
func (p *Provider) ListExperiments(ctx context.Context) ([]*types.ExperimentResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	experiments := make([]*types.ExperimentResult, 0, len(p.experiments))
	for _, experiment := range p.experiments {
		experiments = append(experiments, experiment)
	}

	return experiments, nil
}

// Cleanup cleans up the Kubernetes provider
func (p *Provider) Cleanup(ctx context.Context) error {
	// Stop all running experiments
	for experimentID := range p.experiments {
		if err := p.StopExperiment(ctx, experimentID); err != nil {
			// Log error but continue cleanup
			fmt.Printf("Failed to stop experiment %s: %v\n", experimentID, err)
		}
	}

	return nil
}

// startPodFailureExperiment starts a pod failure experiment
func (p *Provider) startPodFailureExperiment(ctx context.Context, experimentID string, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	namespace := "default"
	if ns, ok := config.Parameters["namespace"].(string); ok {
		namespace = ns
	}

	selector := make(map[string]string)
	if sel, ok := config.Parameters["selector"].(string); ok {
		// Parse selector string like "app=my-app"
		parts := strings.Split(sel, "=")
		if len(parts) == 2 {
			selector[parts[0]] = parts[1]
		}
	}

	podChaos := &chaosmesh.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      experimentID,
			Namespace: "chaos-mesh",
		},
		Spec: chaosmesh.PodChaosSpec{
			Action: chaosmesh.PodFailureAction,
			Mode:   chaosmesh.OneMode,
			Selector: chaosmesh.PodSelector{
				Namespaces:     []string{namespace},
				LabelSelectors: selector,
			},
			Duration: &config.Duration,
		},
	}

	_, err := p.chaosClient.ChaosV1alpha1().PodChaos("chaos-mesh").Create(ctx, podChaos, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create PodChaos: %w", err)
	}

	result := &types.ExperimentResult{
		ID:        experimentID,
		Status:    "running",
		Message:   "Pod failure experiment started",
		StartTime: time.Now().Format(time.RFC3339),
		Metrics: map[string]interface{}{
			"type":      "PodChaos",
			"namespace": namespace,
			"selector":  selector,
		},
	}

	p.experiments[experimentID] = result
	return result, nil
}

// startNetworkLatencyExperiment starts a network latency experiment
func (p *Provider) startNetworkLatencyExperiment(ctx context.Context, experimentID string, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	latency := "100ms"
	if l, ok := config.Parameters["latency"].(string); ok {
		latency = l
	}

	networkChaos := &chaosmesh.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      experimentID,
			Namespace: "chaos-mesh",
		},
		Spec: chaosmesh.NetworkChaosSpec{
			Action: chaosmesh.DelayAction,
			Mode:   chaosmesh.OneMode,
			Selector: chaosmesh.PodSelector{
				Namespaces: []string{"default"},
			},
			Delay: &chaosmesh.DelaySpec{
				Latency: latency,
			},
			Duration: &config.Duration,
		},
	}

	_, err := p.chaosClient.ChaosV1alpha1().NetworkChaos("chaos-mesh").Create(ctx, networkChaos, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create NetworkChaos: %w", err)
	}

	result := &types.ExperimentResult{
		ID:        experimentID,
		Status:    "running",
		Message:   "Network latency experiment started",
		StartTime: time.Now().Format(time.RFC3339),
		Metrics: map[string]interface{}{
			"type":    "NetworkChaos",
			"latency": latency,
		},
	}

	p.experiments[experimentID] = result
	return result, nil
}

// startCPUStressExperiment starts a CPU stress experiment
func (p *Provider) startCPUStressExperiment(ctx context.Context, experimentID string, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	workers := 1
	if w, ok := config.Parameters["workers"].(int); ok {
		workers = w
	}

	stressChaos := &chaosmesh.StressChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      experimentID,
			Namespace: "chaos-mesh",
		},
		Spec: chaosmesh.StressChaosSpec{
			Mode: chaosmesh.OneMode,
			Selector: chaosmesh.PodSelector{
				Namespaces: []string{"default"},
			},
			Stressors: &chaosmesh.Stressors{
				CPUStressor: &chaosmesh.CPUStressor{
					Workers: workers,
				},
			},
			Duration: &config.Duration,
		},
	}

	_, err := p.chaosClient.ChaosV1alpha1().StressChaos("chaos-mesh").Create(ctx, stressChaos, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create StressChaos: %w", err)
	}

	result := &types.ExperimentResult{
		ID:        experimentID,
		Status:    "running",
		Message:   "CPU stress experiment started",
		StartTime: time.Now().Format(time.RFC3339),
		Metrics: map[string]interface{}{
			"type":    "StressChaos",
			"workers": workers,
		},
	}

	p.experiments[experimentID] = result
	return result, nil
}

// startMemoryStressExperiment starts a memory stress experiment
func (p *Provider) startMemoryStressExperiment(ctx context.Context, experimentID string, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	size := "256Mi"
	if s, ok := config.Parameters["size"].(string); ok {
		size = s
	}

	stressChaos := &chaosmesh.StressChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      experimentID,
			Namespace: "chaos-mesh",
		},
		Spec: chaosmesh.StressChaosSpec{
			Mode: chaosmesh.OneMode,
			Selector: chaosmesh.PodSelector{
				Namespaces: []string{"default"},
			},
			Stressors: &chaosmesh.Stressors{
				MemoryStressor: &chaosmesh.MemoryStressor{
					Size: size,
				},
			},
			Duration: &config.Duration,
		},
	}

	_, err := p.chaosClient.ChaosV1alpha1().StressChaos("chaos-mesh").Create(ctx, stressChaos, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create StressChaos: %w", err)
	}

	result := &types.ExperimentResult{
		ID:        experimentID,
		Status:    "running",
		Message:   "Memory stress experiment started",
		StartTime: time.Now().Format(time.RFC3339),
		Metrics: map[string]interface{}{
			"type": "StressChaos",
			"size": size,
		},
	}

	p.experiments[experimentID] = result
	return result, nil
}

// generateExperimentID generates a unique experiment ID
func generateExperimentID() string {
	return fmt.Sprintf("chaos-exp-%d", time.Now().UnixNano())
}
