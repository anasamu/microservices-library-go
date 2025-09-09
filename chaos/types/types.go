package types

import "context"

// ChaosType represents the type of chaos experiment
type ChaosType string

const (
	ChaosTypeKubernetes ChaosType = "kubernetes"
	ChaosTypeHTTP       ChaosType = "http"
	ChaosTypeMessaging  ChaosType = "messaging"
)

// ExperimentType represents the specific type of experiment
type ExperimentType string

const (
	// Kubernetes experiments
	PodFailure     ExperimentType = "pod_failure"
	NetworkLatency ExperimentType = "network_latency"
	CPUStress      ExperimentType = "cpu_stress"
	MemoryStress   ExperimentType = "memory_stress"

	// HTTP experiments
	HTTPLatency ExperimentType = "http_latency"
	HTTPError   ExperimentType = "http_error"
	HTTPTimeout ExperimentType = "http_timeout"

	// Messaging experiments
	MessageDelay   ExperimentType = "message_delay"
	MessageLoss    ExperimentType = "message_loss"
	MessageReorder ExperimentType = "message_reorder"
)

// ExperimentConfig holds configuration for a chaos experiment
type ExperimentConfig struct {
	Type       ChaosType              `json:"type"`
	Experiment ExperimentType         `json:"experiment"`
	Duration   string                 `json:"duration,omitempty"`
	Intensity  float64                `json:"intensity,omitempty"`
	Target     string                 `json:"target,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ExperimentResult holds the result of a chaos experiment
type ExperimentResult struct {
	ID        string                 `json:"id"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
	StartTime string                 `json:"start_time,omitempty"`
	EndTime   string                 `json:"end_time,omitempty"`
}

// ChaosProvider defines the interface for chaos providers
type ChaosProvider interface {
	Initialize(ctx context.Context, config map[string]interface{}) error
	StartExperiment(ctx context.Context, config ExperimentConfig) (*ExperimentResult, error)
	StopExperiment(ctx context.Context, experimentID string) error
	GetExperimentStatus(ctx context.Context, experimentID string) (*ExperimentResult, error)
	ListExperiments(ctx context.Context) ([]*ExperimentResult, error)
	Cleanup(ctx context.Context) error
}
