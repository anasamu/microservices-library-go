package messaging

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/chaos/types"
)

// Message represents a message in the messaging system
type Message struct {
	ID        string                 `json:"id"`
	Topic     string                 `json:"topic"`
	Payload   []byte                 `json:"payload"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// MessageHandler defines the interface for message handlers
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg *Message) error
}

// Provider implements messaging-based chaos engineering
type Provider struct {
	experiments map[string]*types.ExperimentResult
	mu          sync.RWMutex
	handlers    map[string]MessageHandler
	queues      map[string][]*Message
	subscribers map[string][]MessageHandler
}

// NewProvider creates a new messaging chaos provider
func NewProvider() *Provider {
	return &Provider{
		experiments: make(map[string]*types.ExperimentResult),
		handlers:    make(map[string]MessageHandler),
		queues:      make(map[string][]*Message),
		subscribers: make(map[string][]MessageHandler),
	}
}

// Initialize initializes the messaging provider
func (p *Provider) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Initialize message queues and handlers based on config
	if topics, ok := config["topics"].([]string); ok {
		for _, topic := range topics {
			p.queues[topic] = make([]*Message, 0)
			p.subscribers[topic] = make([]MessageHandler, 0)
		}
	}

	return nil
}

// StartExperiment starts a messaging chaos experiment
func (p *Provider) StartExperiment(ctx context.Context, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	experimentID := generateExperimentID()

	switch config.Experiment {
	case types.MessageDelay:
		return p.startDelayExperiment(ctx, experimentID, config)
	case types.MessageLoss:
		return p.startLossExperiment(ctx, experimentID, config)
	case types.MessageReorder:
		return p.startReorderExperiment(ctx, experimentID, config)
	default:
		return nil, fmt.Errorf("unsupported experiment type: %s", config.Experiment)
	}
}

// StopExperiment stops a messaging chaos experiment
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

// GetExperimentStatus gets the status of a messaging chaos experiment
func (p *Provider) GetExperimentStatus(ctx context.Context, experimentID string) (*types.ExperimentResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	experiment, exists := p.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	return experiment, nil
}

// ListExperiments lists all messaging chaos experiments
func (p *Provider) ListExperiments(ctx context.Context) ([]*types.ExperimentResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	experiments := make([]*types.ExperimentResult, 0, len(p.experiments))
	for _, experiment := range p.experiments {
		experiments = append(experiments, experiment)
	}

	return experiments, nil
}

// Cleanup cleans up the messaging provider
func (p *Provider) Cleanup(ctx context.Context) error {
	// Stop all running experiments
	for experimentID := range p.experiments {
		if err := p.StopExperiment(ctx, experimentID); err != nil {
			// Log error but continue cleanup
			fmt.Printf("Failed to stop experiment %s: %v\n", experimentID, err)
		}
	}

	// Clear all queues and subscribers
	p.queues = make(map[string][]*Message)
	p.subscribers = make(map[string][]MessageHandler)

	return nil
}

// PublishMessage publishes a message to a topic
func (p *Provider) PublishMessage(ctx context.Context, topic string, payload []byte, headers map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	message := &Message{
		ID:        generateMessageID(),
		Topic:     topic,
		Payload:   payload,
		Headers:   headers,
		Timestamp: time.Now(),
	}

	// Add to queue
	if p.queues[topic] == nil {
		p.queues[topic] = make([]*Message, 0)
	}
	p.queues[topic] = append(p.queues[topic], message)

	// Process message with any active chaos experiments
	return p.processMessage(ctx, message)
}

// Subscribe subscribes a handler to a topic
func (p *Provider) Subscribe(topic string, handler MessageHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.subscribers[topic] == nil {
		p.subscribers[topic] = make([]MessageHandler, 0)
	}
	p.subscribers[topic] = append(p.subscribers[topic], handler)
}

// startDelayExperiment starts a message delay experiment
func (p *Provider) startDelayExperiment(ctx context.Context, experimentID string, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	delayMs := 5000
	if d, ok := config.Parameters["delay_ms"].(int); ok {
		delayMs = d
	}

	probability := 0.2
	if prob, ok := config.Parameters["probability"].(float64); ok {
		probability = prob
	}

	result := &types.ExperimentResult{
		ID:        experimentID,
		Status:    "running",
		Message:   "Message delay experiment started",
		StartTime: time.Now().Format(time.RFC3339),
		Metrics: map[string]interface{}{
			"type":        "MessageDelay",
			"delay_ms":    delayMs,
			"probability": probability,
		},
	}

	p.experiments[experimentID] = result
	return result, nil
}

// startLossExperiment starts a message loss experiment
func (p *Provider) startLossExperiment(ctx context.Context, experimentID string, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	lossRate := 0.1
	if rate, ok := config.Parameters["loss_rate"].(float64); ok {
		lossRate = rate
	}

	result := &types.ExperimentResult{
		ID:        experimentID,
		Status:    "running",
		Message:   "Message loss experiment started",
		StartTime: time.Now().Format(time.RFC3339),
		Metrics: map[string]interface{}{
			"type":      "MessageLoss",
			"loss_rate": lossRate,
		},
	}

	p.experiments[experimentID] = result
	return result, nil
}

// startReorderExperiment starts a message reorder experiment
func (p *Provider) startReorderExperiment(ctx context.Context, experimentID string, config types.ExperimentConfig) (*types.ExperimentResult, error) {
	reorderRate := 0.05
	if rate, ok := config.Parameters["reorder_rate"].(float64); ok {
		reorderRate = rate
	}

	result := &types.ExperimentResult{
		ID:        experimentID,
		Status:    "running",
		Message:   "Message reorder experiment started",
		StartTime: time.Now().Format(time.RFC3339),
		Metrics: map[string]interface{}{
			"type":         "MessageReorder",
			"reorder_rate": reorderRate,
		},
	}

	p.experiments[experimentID] = result
	return result, nil
}

// processMessage processes a message with active chaos experiments
func (p *Provider) processMessage(ctx context.Context, message *Message) error {
	// Apply chaos experiments
	for _, experiment := range p.experiments {
		if experiment.Status != "running" {
			continue
		}

		switch experiment.Metrics["type"] {
		case "MessageDelay":
			if rand.Float64() < experiment.Metrics["probability"].(float64) {
				delayMs := experiment.Metrics["delay_ms"].(int)
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
			}
		case "MessageLoss":
			if rand.Float64() < experiment.Metrics["loss_rate"].(float64) {
				// Drop the message
				return nil
			}
		case "MessageReorder":
			if rand.Float64() < experiment.Metrics["reorder_rate"].(float64) {
				// Add random delay to simulate reordering
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			}
		}
	}

	// Deliver to subscribers
	subscribers := p.subscribers[message.Topic]
	for _, handler := range subscribers {
		go func(h MessageHandler) {
			if err := h.HandleMessage(ctx, message); err != nil {
				fmt.Printf("Error handling message %s: %v\n", message.ID, err)
			}
		}(handler)
	}

	return nil
}

// generateExperimentID generates a unique experiment ID
func generateExperimentID() string {
	return fmt.Sprintf("msg-chaos-%d", time.Now().UnixNano())
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}
