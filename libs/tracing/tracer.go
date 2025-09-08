package tracing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerManager handles distributed tracing
type TracerManager struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
	config   *TracingConfig
	logger   *logrus.Logger
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	JaegerEndpoint string
	ZipkinEndpoint string
	SampleRate     float64
	Enabled        bool
}

// SpanContext represents span context information
type SpanContext struct {
	TraceID    string
	SpanID     string
	TraceFlags string
	TraceState string
}

// NewTracerManager creates a new tracer manager
func NewTracerManager(config *TracingConfig, logger *logrus.Logger) (*TracerManager, error) {
	if !config.Enabled {
		return &TracerManager{
			config: config,
			logger: logger,
		}, nil
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter
	var exporter sdktrace.SpanExporter
	if config.JaegerEndpoint != "" {
		exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
		if err != nil {
			return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
		}
	} else if config.ZipkinEndpoint != "" {
		exporter, err = zipkin.New(config.ZipkinEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to create Zipkin exporter: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no tracing endpoint configured")
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SampleRate)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := tp.Tracer(config.ServiceName)

	manager := &TracerManager{
		tracer:   tracer,
		provider: tp,
		config:   config,
		logger:   logger,
	}

	logger.WithFields(logrus.Fields{
		"service_name":    config.ServiceName,
		"service_version": config.ServiceVersion,
		"environment":     config.Environment,
		"jaeger_endpoint": config.JaegerEndpoint,
		"zipkin_endpoint": config.ZipkinEndpoint,
		"sample_rate":     config.SampleRate,
	}).Info("Tracer manager initialized successfully")

	return manager, nil
}

// StartSpan starts a new span
func (tm *TracerManager) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !tm.config.Enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	return tm.tracer.Start(ctx, name, opts...)
}

// StartSpanWithAttributes starts a new span with attributes
func (tm *TracerManager) StartSpanWithAttributes(ctx context.Context, name string, attributes map[string]interface{}, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !tm.config.Enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	// Convert attributes to OpenTelemetry attributes
	otelAttrs := make([]attribute.KeyValue, 0, len(attributes))
	for key, value := range attributes {
		otelAttrs = append(otelAttrs, attribute.String(key, fmt.Sprintf("%v", value)))
	}

	opts = append(opts, trace.WithAttributes(otelAttrs...))
	return tm.tracer.Start(ctx, name, opts...)
}

// AddSpanEvent adds an event to the current span
func (tm *TracerManager) AddSpanEvent(ctx context.Context, name string, attributes map[string]interface{}) {
	if !tm.config.Enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	// Convert attributes to OpenTelemetry attributes
	otelAttrs := make([]attribute.KeyValue, 0, len(attributes))
	for key, value := range attributes {
		otelAttrs = append(otelAttrs, attribute.String(key, fmt.Sprintf("%v", value)))
	}

	span.AddEvent(name, trace.WithAttributes(otelAttrs...))
}

// AddSpanAttributes adds attributes to the current span
func (tm *TracerManager) AddSpanAttributes(ctx context.Context, attributes map[string]interface{}) {
	if !tm.config.Enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	// Convert attributes to OpenTelemetry attributes
	otelAttrs := make([]attribute.KeyValue, 0, len(attributes))
	for key, value := range attributes {
		otelAttrs = append(otelAttrs, attribute.String(key, fmt.Sprintf("%v", value)))
	}

	span.SetAttributes(otelAttrs...)
}

// SetSpanStatus sets the status of the current span
func (tm *TracerManager) SetSpanStatus(ctx context.Context, code trace.StatusCode, description string) {
	if !tm.config.Enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetStatus(code, description)
}

// FinishSpan finishes the current span
func (tm *TracerManager) FinishSpan(ctx context.Context) {
	if !tm.config.Enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.End()
	}
}

// GetSpanContext returns the span context from the current context
func (tm *TracerManager) GetSpanContext(ctx context.Context) *SpanContext {
	if !tm.config.Enabled {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return nil
	}

	spanContext := span.SpanContext()
	return &SpanContext{
		TraceID:    spanContext.TraceID().String(),
		SpanID:     spanContext.SpanID().String(),
		TraceFlags: spanContext.TraceFlags().String(),
		TraceState: spanContext.TraceState().String(),
	}
}

// InjectSpanContext injects span context into a map
func (tm *TracerManager) InjectSpanContext(ctx context.Context, carrier map[string]string) {
	if !tm.config.Enabled {
		return
	}

	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, &mapCarrier{carrier: carrier})
}

// ExtractSpanContext extracts span context from a map
func (tm *TracerManager) ExtractSpanContext(ctx context.Context, carrier map[string]string) context.Context {
	if !tm.config.Enabled {
		return ctx
	}

	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(ctx, &mapCarrier{carrier: carrier})
}

// CreateChildSpan creates a child span
func (tm *TracerManager) CreateChildSpan(ctx context.Context, name string, attributes map[string]interface{}) (context.Context, trace.Span) {
	if !tm.config.Enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	// Convert attributes to OpenTelemetry attributes
	otelAttrs := make([]attribute.KeyValue, 0, len(attributes))
	for key, value := range attributes {
		otelAttrs = append(otelAttrs, attribute.String(key, fmt.Sprintf("%v", value)))
	}

	return tm.tracer.Start(ctx, name, trace.WithAttributes(otelAttrs...))
}

// TraceFunction traces a function execution
func (tm *TracerManager) TraceFunction(ctx context.Context, functionName string, fn func(context.Context) error) error {
	if !tm.config.Enabled {
		return fn(ctx)
	}

	ctx, span := tm.StartSpan(ctx, functionName)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	// Add timing information
	span.SetAttributes(
		attribute.String("function.name", functionName),
		attribute.Int64("function.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		span.SetStatus(trace.StatusCodeError, err.Error())
		span.RecordError(err)
	} else {
		span.SetStatus(trace.StatusCodeOk, "success")
	}

	return err
}

// TraceAsyncFunction traces an async function execution
func (tm *TracerManager) TraceAsyncFunction(ctx context.Context, functionName string, fn func(context.Context) error) <-chan error {
	result := make(chan error, 1)

	go func() {
		defer close(result)
		result <- tm.TraceFunction(ctx, functionName, fn)
	}()

	return result
}

// Close closes the tracer manager
func (tm *TracerManager) Close(ctx context.Context) error {
	if !tm.config.Enabled || tm.provider == nil {
		return nil
	}

	return tm.provider.Shutdown(ctx)
}

// mapCarrier implements the TextMapCarrier interface
type mapCarrier struct {
	carrier map[string]string
}

// Get returns the value for a key
func (mc *mapCarrier) Get(key string) string {
	return mc.carrier[key]
}

// Set sets the value for a key
func (mc *mapCarrier) Set(key, value string) {
	mc.carrier[key] = value
}

// Keys returns all keys
func (mc *mapCarrier) Keys() []string {
	keys := make([]string, 0, len(mc.carrier))
	for key := range mc.carrier {
		keys = append(keys, key)
	}
	return keys
}

// CreateTraceID creates a new trace ID
func CreateTraceID() string {
	return uuid.New().String()
}

// CreateSpanID creates a new span ID
func CreateSpanID() string {
	return uuid.New().String()
}

// GetTracer returns the global tracer
func GetTracer() trace.Tracer {
	return otel.Tracer("siakad")
}

// GetTracerProvider returns the global tracer provider
func GetTracerProvider() trace.TracerProvider {
	return otel.GetTracerProvider()
}

// GetTextMapPropagator returns the global text map propagator
func GetTextMapPropagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}
