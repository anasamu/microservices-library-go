package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/monitoring/gateway"
	"github.com/anasamu/microservices-library-go/monitoring/providers/elasticsearch"
	"github.com/anasamu/microservices-library-go/monitoring/providers/jaeger"
	"github.com/anasamu/microservices-library-go/monitoring/providers/prometheus"
	"github.com/anasamu/microservices-library-go/monitoring/types"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create monitoring manager
	config := &gateway.ManagerConfig{
		DefaultProvider: "prometheus",
		RetryAttempts:   3,
		RetryDelay:      5 * time.Second,
		Timeout:         30 * time.Second,
		BatchSize:       1000,
		FlushInterval:   10 * time.Second,
		BufferSize:      10000,
		SamplingRate:    1.0,
		EnableTracing:   true,
		EnableMetrics:   true,
		EnableLogging:   true,
		EnableAlerting:  true,
		Metadata:        make(map[string]string),
	}

	manager := gateway.NewMonitoringManager(config, logger)

	// Create and register providers
	prometheusProvider := prometheus.NewPrometheusProvider(nil, logger)
	jaegerProvider := jaeger.NewJaegerProvider(nil, logger)
	elasticsearchProvider := elasticsearch.NewElasticsearchProvider(nil, logger)

	// Register providers
	if err := manager.RegisterProvider(prometheusProvider); err != nil {
		log.Fatalf("Failed to register Prometheus provider: %v", err)
	}

	if err := manager.RegisterProvider(jaegerProvider); err != nil {
		log.Fatalf("Failed to register Jaeger provider: %v", err)
	}

	if err := manager.RegisterProvider(elasticsearchProvider); err != nil {
		log.Fatalf("Failed to register Elasticsearch provider: %v", err)
	}

	// Connect to providers
	ctx := context.Background()

	if err := manager.Connect(ctx, "prometheus"); err != nil {
		log.Printf("Failed to connect to Prometheus: %v", err)
	}

	if err := manager.Connect(ctx, "jaeger"); err != nil {
		log.Printf("Failed to connect to Jaeger: %v", err)
	}

	if err := manager.Connect(ctx, "elasticsearch"); err != nil {
		log.Printf("Failed to connect to Elasticsearch: %v", err)
	}

	// Example 1: Submit metrics to Prometheus
	fmt.Println("=== Example 1: Submitting Metrics to Prometheus ===")
	metrics := []types.Metric{
		*prometheus.CreateCounterMetric("http_requests_total", 1, map[string]string{
			"method": "GET",
			"path":   "/api/users",
			"status": "200",
		}),
		*prometheus.CreateGaugeMetric("memory_usage_bytes", 1024*1024*100, map[string]string{
			"instance": "server-1",
		}),
		*prometheus.CreateHistogramMetric("request_duration_seconds", 0.5, map[string]string{
			"method": "POST",
			"path":   "/api/orders",
		}),
	}

	metricRequest := &types.MetricRequest{
		Metrics: metrics,
		Options: map[string]interface{}{
			"batch": true,
		},
	}

	if err := manager.SubmitMetrics(ctx, "prometheus", metricRequest); err != nil {
		log.Printf("Failed to submit metrics: %v", err)
	}

	// Example 2: Submit traces to Jaeger
	fmt.Println("\n=== Example 2: Submitting Traces to Jaeger ===")
	trace := jaeger.CreateTrace("user-service", "get-user")
	span1 := jaeger.CreateSpan(trace.TraceID, "", "database-query")
	span2 := jaeger.CreateSpan(trace.TraceID, span1.SpanID, "cache-lookup")

	// Add tags and logs to spans
	span1.AddTag("db.type", "postgresql")
	span1.AddTag("db.table", "users")
	span1.AddLog(types.LogLevelInfo, "Executing user query", map[string]interface{}{
		"query":  "SELECT * FROM users WHERE id = ?",
		"params": []interface{}{123},
	})

	span2.AddTag("cache.type", "redis")
	span2.AddTag("cache.key", "user:123")
	span2.AddLog(types.LogLevelDebug, "Cache miss", map[string]interface{}{
		"key": "user:123",
	})

	// Finish spans
	span1.FinishSpan("ok")
	span2.FinishSpan("ok")

	// Add spans to trace
	trace.AddSpan(span1)
	trace.AddSpan(span2)

	// Finish trace
	trace.FinishTrace("ok")

	traceRequest := &types.TraceRequest{
		Traces: []types.Trace{*trace},
		Options: map[string]interface{}{
			"batch": true,
		},
	}

	if err := manager.SubmitTraces(ctx, "jaeger", traceRequest); err != nil {
		log.Printf("Failed to submit traces: %v", err)
	}

	// Example 3: Submit logs to Elasticsearch
	fmt.Println("\n=== Example 3: Submitting Logs to Elasticsearch ===")
	logs := []types.LogEntry{
		*elasticsearch.CreateLogEntry(types.LogLevelInfo, "User login successful", "auth-service", "auth-service"),
		*elasticsearch.CreateLogEntry(types.LogLevelWarn, "High memory usage detected", "monitoring-service", "monitoring-service"),
		*elasticsearch.CreateLogEntry(types.LogLevelError, "Database connection failed", "user-service", "user-service"),
	}

	// Add fields to logs
	logs[0].AddField("user_id", "12345")
	logs[0].AddField("ip_address", "192.168.1.100")
	logs[0].SetTraceInfo(trace.TraceID, span1.SpanID)

	logs[1].AddField("memory_usage", "85%")
	logs[1].AddField("threshold", "80%")

	logs[2].AddField("error_code", "DB_CONNECTION_FAILED")
	logs[2].AddField("retry_count", 3)

	logRequest := &types.LogRequest{
		Logs: logs,
		Options: map[string]interface{}{
			"batch": true,
		},
	}

	if err := manager.SubmitLogs(ctx, "elasticsearch", logRequest); err != nil {
		log.Printf("Failed to submit logs: %v", err)
	}

	// Example 4: Submit alerts to Prometheus Alertmanager
	fmt.Println("\n=== Example 4: Submitting Alerts to Prometheus Alertmanager ===")
	alerts := []types.Alert{
		*prometheus.CreateAlert("HighMemoryUsage", "Memory usage is above 90%", types.AlertSeverityHigh, "monitoring", "user-service"),
		*prometheus.CreateAlert("DatabaseConnectionFailed", "Database connection failed multiple times", types.AlertSeverityCritical, "monitoring", "user-service"),
		*prometheus.CreateAlert("SlowResponseTime", "API response time is above threshold", types.AlertSeverityMedium, "monitoring", "api-gateway"),
	}

	// Add labels and annotations to alerts
	alerts[0].Labels["instance"] = "server-1"
	alerts[0].Labels["severity"] = "high"
	alerts[0].Annotations["summary"] = "High memory usage detected"
	alerts[0].Annotations["description"] = "Memory usage has exceeded 90% for more than 5 minutes"

	alerts[1].Labels["instance"] = "db-server-1"
	alerts[1].Labels["severity"] = "critical"
	alerts[1].Annotations["summary"] = "Database connection failed"
	alerts[1].Annotations["description"] = "Database connection has failed 5 times in the last 10 minutes"

	alerts[2].Labels["endpoint"] = "/api/users"
	alerts[2].Labels["severity"] = "medium"
	alerts[2].Annotations["summary"] = "Slow API response time"
	alerts[2].Annotations["description"] = "API response time is above 2 seconds"

	alertRequest := &types.AlertRequest{
		Alerts: alerts,
		Options: map[string]interface{}{
			"batch": true,
		},
	}

	if err := manager.SubmitAlerts(ctx, "prometheus", alertRequest); err != nil {
		log.Printf("Failed to submit alerts: %v", err)
	}

	// Example 5: Query data from providers
	fmt.Println("\n=== Example 5: Querying Data from Providers ===")

	// Query metrics from Prometheus
	metricQueryRequest := &types.QueryRequest{
		Query: "http_requests_total",
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now(),
		Step:  5 * time.Minute,
	}

	metricResponse, err := manager.QueryMetrics(ctx, "prometheus", metricQueryRequest)
	if err != nil {
		log.Printf("Failed to query metrics: %v", err)
	} else {
		fmt.Printf("Metrics query result: %+v\n", metricResponse.Data)
	}

	// Query traces from Jaeger
	traceQueryRequest := &types.QueryRequest{
		Query: "service:user-service",
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now(),
	}

	traceResponse, err := manager.QueryTraces(ctx, "jaeger", traceQueryRequest)
	if err != nil {
		log.Printf("Failed to query traces: %v", err)
	} else {
		fmt.Printf("Traces query result: %+v\n", traceResponse.Data)
	}

	// Query logs from Elasticsearch
	logQueryRequest := &types.QueryRequest{
		Query: "level:error",
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now(),
	}

	logResponse, err := manager.QueryLogs(ctx, "elasticsearch", logQueryRequest)
	if err != nil {
		log.Printf("Failed to query logs: %v", err)
	} else {
		fmt.Printf("Logs query result: %+v\n", logResponse.Data)
	}

	// Example 6: Health checks
	fmt.Println("\n=== Example 6: Health Checks ===")

	// Health check for specific provider
	healthRequest := &types.HealthCheckRequest{
		Service: "prometheus",
		Timeout: 10 * time.Second,
	}

	healthResponse, err := manager.HealthCheck(ctx, "prometheus", healthRequest)
	if err != nil {
		log.Printf("Failed to perform health check: %v", err)
	} else {
		fmt.Printf("Prometheus health status: %s\n", healthResponse.Status)
		for _, check := range healthResponse.Checks {
			fmt.Printf("  - %s: %s (%s)\n", check.Name, check.Status, check.Message)
		}
	}

	// Health check for all providers
	allHealthResponses := manager.HealthCheckAll(ctx)
	for provider, response := range allHealthResponses {
		fmt.Printf("%s health status: %s\n", provider, response.Status)
	}

	// Example 7: Get statistics
	fmt.Println("\n=== Example 7: Provider Statistics ===")

	// Get stats from Prometheus
	prometheusStats, err := manager.GetStats(ctx, "prometheus")
	if err != nil {
		log.Printf("Failed to get Prometheus stats: %v", err)
	} else {
		fmt.Printf("Prometheus stats: %+v\n", prometheusStats)
	}

	// Get stats from Jaeger
	jaegerStats, err := manager.GetStats(ctx, "jaeger")
	if err != nil {
		log.Printf("Failed to get Jaeger stats: %v", err)
	} else {
		fmt.Printf("Jaeger stats: %+v\n", jaegerStats)
	}

	// Get stats from Elasticsearch
	elasticsearchStats, err := manager.GetStats(ctx, "elasticsearch")
	if err != nil {
		log.Printf("Failed to get Elasticsearch stats: %v", err)
	} else {
		fmt.Printf("Elasticsearch stats: %+v\n", elasticsearchStats)
	}

	// Example 8: Provider information
	fmt.Println("\n=== Example 8: Provider Information ===")

	// Get supported providers
	providers := manager.GetSupportedProviders()
	fmt.Printf("Supported providers: %v\n", providers)

	// Get connected providers
	connectedProviders := manager.GetConnectedProviders()
	fmt.Printf("Connected providers: %v\n", connectedProviders)

	// Get provider capabilities
	capabilities, connInfo, err := manager.GetProviderCapabilities("prometheus")
	if err != nil {
		log.Printf("Failed to get provider capabilities: %v", err)
	} else {
		fmt.Printf("Prometheus capabilities: %v\n", capabilities)
		fmt.Printf("Prometheus connection info: %+v\n", connInfo)
	}

	// Clean up
	fmt.Println("\n=== Cleaning up ===")
	if err := manager.Close(); err != nil {
		log.Printf("Failed to close monitoring manager: %v", err)
	}

	fmt.Println("Monitoring example completed successfully!")
}
