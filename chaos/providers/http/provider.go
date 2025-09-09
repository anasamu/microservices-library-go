package http

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/chaos/types"
)

// Provider implements HTTP-based chaos engineering
type Provider struct {
	experiments map[string]*types.ExperimentResult
	mu          sync.RWMutex
	transport   *http.Transport
	client      *http.Client
}

// NewProvider creates a new HTTP chaos provider
func NewProvider() *Provider {
	transport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &Provider{
		experiments: make(map[string]*types.ExperimentResult),
		transport:   transport,
		client:      client,
	}
}

// Initialize initializes the HTTP provider
func (p *Provider) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Configure transport based on config
	if maxIdleConns, ok := config["max_idle_conns"].(int); ok {
		p.transport.MaxIdleConns = maxIdleConns
	}

	if idleConnTimeout, ok := config["idle_conn_timeout"].(string); ok {
		if duration, err := time.ParseDuration(idleConnTimeout); err == nil {
			p.transport.IdleConnTimeout = duration
		}
	}

	if timeout, ok := config["timeout"].(string); ok {
		if duration, err := time.ParseDuration(timeout); err == nil {
			p.client.Timeout = duration
		}
	}

	return nil
}

// StartExperiment starts an HTTP chaos experiment
func (p *Provider) StartExperiment(ctx context.Context, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	experimentID := generateExperimentID()

	switch config.Experiment {
	case types.HTTPLatency:
		return p.startLatencyExperiment(ctx, experimentID, config)
	case types.HTTPError:
		return p.startErrorExperiment(ctx, experimentID, config)
	case types.HTTPTimeout:
		return p.startTimeoutExperiment(ctx, experimentID, config)
	default:
		return nil, fmt.Errorf("unsupported experiment type: %s", config.Experiment)
	}
}

// StopExperiment stops an HTTP chaos experiment
func (p *Provider) StopExperiment(ctx context.Context, experimentID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	experiment, exists := p.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	// Update experiment status
	experiment.Status = "stopped"
	experiment.EndTime = time.Now().Format(time.RFC3339)

	return nil
}

// GetExperimentStatus gets the status of an HTTP chaos experiment
func (p *Provider) GetExperimentStatus(ctx context.Context, experimentID string) (*types.ExperimentResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	experiment, exists := p.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	return experiment, nil
}

// ListExperiments lists all HTTP chaos experiments
func (p *Provider) ListExperiments(ctx context.Context) ([]*types.ExperimentResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	experiments := make([]*types.ExperimentResult, 0, len(p.experiments))
	for _, experiment := range p.experiments {
		experiments = append(experiments, experiment)
	}

	return experiments, nil
}

// Cleanup cleans up the HTTP provider
func (p *Provider) Cleanup(ctx context.Context) error {
	// Stop all running experiments
	for experimentID := range p.experiments {
		if err := p.StopExperiment(ctx, experimentID); err != nil {
			// Log error but continue cleanup
			fmt.Printf("Failed to stop experiment %s: %v\n", experimentID, err)
		}
	}

	// Close transport connections
	p.transport.CloseIdleConnections()

	return nil
}

// startLatencyExperiment starts an HTTP latency experiment
func (p *Provider) startLatencyExperiment(ctx context.Context, experimentID string, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	latencyMs := 1000
	if l, ok := config.Parameters["latency_ms"].(int); ok {
		latencyMs = l
	}

	probability := 0.3
	if prob, ok := config.Parameters["probability"].(float64); ok {
		probability = prob
	}

	// Create a custom transport with latency injection
	latencyTransport := &LatencyTransport{
		Base:        p.transport,
		LatencyMs:   latencyMs,
		Probability: probability,
	}

	latencyClient := &http.Client{
		Transport: latencyTransport,
		Timeout:   p.client.Timeout,
	}

	// Store the custom client for this experiment
	result := &types.ExperimentResult{
		ID:        experimentID,
		Status:    "running",
		Message:   "HTTP latency experiment started",
		StartTime: time.Now().Format(time.RFC3339),
		Metrics: map[string]interface{}{
			"type":        "HTTPLatency",
			"latency_ms":  latencyMs,
			"probability": probability,
			"client":      latencyClient,
		},
	}

	p.experiments[experimentID] = result
	return result, nil
}

// startErrorExperiment starts an HTTP error experiment
func (p *Provider) startErrorExperiment(ctx context.Context, experimentID string, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	errorCode := 500
	if code, ok := config.Parameters["error_code"].(int); ok {
		errorCode = code
	}

	probability := 0.2
	if prob, ok := config.Parameters["probability"].(float64); ok {
		probability = prob
	}

	// Create a custom transport with error injection
	errorTransport := &ErrorTransport{
		Base:        p.transport,
		ErrorCode:   errorCode,
		Probability: probability,
	}

	errorClient := &http.Client{
		Transport: errorTransport,
		Timeout:   p.client.Timeout,
	}

	result := &types.ExperimentResult{
		ID:        experimentID,
		Status:    "running",
		Message:   "HTTP error experiment started",
		StartTime: time.Now().Format(time.RFC3339),
		Metrics: map[string]interface{}{
			"type":        "HTTPError",
			"error_code":  errorCode,
			"probability": probability,
			"client":      errorClient,
		},
	}

	p.experiments[experimentID] = result
	return result, nil
}

// startTimeoutExperiment starts an HTTP timeout experiment
func (p *Provider) startTimeoutExperiment(ctx context.Context, experimentID string, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	timeoutMs := 5000
	if t, ok := config.Parameters["timeout_ms"].(int); ok {
		timeoutMs = t
	}

	probability := 0.1
	if prob, ok := config.Parameters["probability"].(float64); ok {
		probability = prob
	}

	// Create a custom transport with timeout injection
	timeoutTransport := &TimeoutTransport{
		Base:        p.transport,
		TimeoutMs:   timeoutMs,
		Probability: probability,
	}

	timeoutClient := &http.Client{
		Transport: timeoutTransport,
		Timeout:   p.client.Timeout,
	}

	result := &types.ExperimentResult{
		ID:        experimentID,
		Status:    "running",
		Message:   "HTTP timeout experiment started",
		StartTime: time.Now().Format(time.RFC3339),
		Metrics: map[string]interface{}{
			"type":        "HTTPTimeout",
			"timeout_ms":  timeoutMs,
			"probability": probability,
			"client":      timeoutClient,
		},
	}

	p.experiments[experimentID] = result
	return result, nil
}

// LatencyTransport wraps an HTTP transport to inject latency
type LatencyTransport struct {
	Base        http.RoundTripper
	LatencyMs   int
	Probability float64
}

func (t *LatencyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Inject latency based on probability
	if rand.Float64() < t.Probability {
		time.Sleep(time.Duration(t.LatencyMs) * time.Millisecond)
	}

	return t.Base.RoundTrip(req)
}

// ErrorTransport wraps an HTTP transport to inject errors
type ErrorTransport struct {
	Base        http.RoundTripper
	ErrorCode   int
	Probability float64
}

func (t *ErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Inject error based on probability
	if rand.Float64() < t.Probability {
		return &http.Response{
			StatusCode: t.ErrorCode,
			Status:     http.StatusText(t.ErrorCode),
			Body:       http.NoBody,
			Request:    req,
		}, nil
	}

	return t.Base.RoundTrip(req)
}

// TimeoutTransport wraps an HTTP transport to inject timeouts
type TimeoutTransport struct {
	Base        http.RoundTripper
	TimeoutMs   int
	Probability float64
}

func (t *TimeoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Inject timeout based on probability
	if rand.Float64() < t.Probability {
		time.Sleep(time.Duration(t.TimeoutMs) * time.Millisecond)
		return nil, fmt.Errorf("request timeout after %dms", t.TimeoutMs)
	}

	return t.Base.RoundTrip(req)
}

// generateExperimentID generates a unique experiment ID
func generateExperimentID() string {
	return fmt.Sprintf("http-chaos-%d", time.Now().UnixNano())
}
