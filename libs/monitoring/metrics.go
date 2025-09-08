package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// Metrics represents the application metrics
type Metrics struct {
	config *MetricsConfig
	logger *logrus.Logger

	// HTTP metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestSize     *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec

	// Database metrics
	dbConnectionsActive prometheus.Gauge
	dbConnectionsIdle   prometheus.Gauge
	dbQueriesTotal      *prometheus.CounterVec
	dbQueryDuration     *prometheus.HistogramVec

	// gRPC metrics
	grpcRequestsTotal   *prometheus.CounterVec
	grpcRequestDuration *prometheus.HistogramVec

	// Business metrics
	businessOperationsTotal   *prometheus.CounterVec
	businessOperationDuration *prometheus.HistogramVec

	// System metrics
	systemMemoryUsage prometheus.Gauge
	systemCPUUsage    prometheus.Gauge
	systemGoroutines  prometheus.Gauge

	// Custom metrics
	customMetrics map[string]prometheus.Collector
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled     bool
	Port        string
	Path        string
	ServiceName string
	Version     string
}

// NewMetrics creates a new metrics instance
func NewMetrics(config *MetricsConfig, logger *logrus.Logger) *Metrics {
	if !config.Enabled {
		return &Metrics{
			config: config,
			logger: logger,
		}
	}

	m := &Metrics{
		config:        config,
		logger:        logger,
		customMetrics: make(map[string]prometheus.Collector),
	}

	// Initialize HTTP metrics
	m.httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code", "service", "version"},
	)

	m.httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status_code", "service", "version"},
	)

	m.httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "service", "version"},
	)

	m.httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "status_code", "service", "version"},
	)

	// Initialize database metrics
	m.dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	m.dbConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	m.dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status", "service", "version"},
	)

	m.dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table", "service", "version"},
	)

	// Initialize gRPC metrics
	m.grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status_code", "service", "version"},
	)

	m.grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status_code", "service", "version"},
	)

	// Initialize business metrics
	m.businessOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "business_operations_total",
			Help: "Total number of business operations",
		},
		[]string{"operation", "entity", "status", "service", "version"},
	)

	m.businessOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "business_operation_duration_seconds",
			Help:    "Business operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "entity", "service", "version"},
	)

	// Initialize system metrics
	m.systemMemoryUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "system_memory_usage_bytes",
			Help: "System memory usage in bytes",
		},
	)

	m.systemCPUUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "system_cpu_usage_percent",
			Help: "System CPU usage percentage",
		},
	)

	m.systemGoroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "system_goroutines_total",
			Help: "Total number of goroutines",
		},
	)

	m.logger.Info("Metrics initialized successfully")
	return m
}

// Start starts the metrics server
func (m *Metrics) Start() error {
	if !m.config.Enabled {
		return nil
	}

	http.Handle(m.config.Path, promhttp.Handler())

	m.logger.Infof("Metrics server starting on port %s", m.config.Port)

	go func() {
		if err := http.ListenAndServe(":"+m.config.Port, nil); err != nil {
			m.logger.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	return nil
}

// RecordHTTPRequest records HTTP request metrics
func (m *Metrics) RecordHTTPRequest(method, path, statusCode string, duration time.Duration, requestSize, responseSize int64) {
	if !m.config.Enabled {
		return
	}

	labels := prometheus.Labels{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"service":     m.config.ServiceName,
		"version":     m.config.Version,
	}

	m.httpRequestsTotal.With(labels).Inc()
	m.httpRequestDuration.With(labels).Observe(duration.Seconds())

	requestLabels := prometheus.Labels{
		"method":  method,
		"path":    path,
		"service": m.config.ServiceName,
		"version": m.config.Version,
	}

	m.httpRequestSize.With(requestLabels).Observe(float64(requestSize))
	m.httpResponseSize.With(labels).Observe(float64(responseSize))
}

// RecordDatabaseQuery records database query metrics
func (m *Metrics) RecordDatabaseQuery(operation, table, status string, duration time.Duration) {
	if !m.config.Enabled {
		return
	}

	labels := prometheus.Labels{
		"operation": operation,
		"table":     table,
		"status":    status,
		"service":   m.config.ServiceName,
		"version":   m.config.Version,
	}

	m.dbQueriesTotal.With(labels).Inc()

	queryLabels := prometheus.Labels{
		"operation": operation,
		"table":     table,
		"service":   m.config.ServiceName,
		"version":   m.config.Version,
	}

	m.dbQueryDuration.With(queryLabels).Observe(duration.Seconds())
}

// RecordGRPCRequest records gRPC request metrics
func (m *Metrics) RecordGRPCRequest(method, statusCode string, duration time.Duration) {
	if !m.config.Enabled {
		return
	}

	labels := prometheus.Labels{
		"method":      method,
		"status_code": statusCode,
		"service":     m.config.ServiceName,
		"version":     m.config.Version,
	}

	m.grpcRequestsTotal.With(labels).Inc()
	m.grpcRequestDuration.With(labels).Observe(duration.Seconds())
}

// RecordBusinessOperation records business operation metrics
func (m *Metrics) RecordBusinessOperation(operation, entity, status string, duration time.Duration) {
	if !m.config.Enabled {
		return
	}

	labels := prometheus.Labels{
		"operation": operation,
		"entity":    entity,
		"status":    status,
		"service":   m.config.ServiceName,
		"version":   m.config.Version,
	}

	m.businessOperationsTotal.With(labels).Inc()

	opLabels := prometheus.Labels{
		"operation": operation,
		"entity":    entity,
		"service":   m.config.ServiceName,
		"version":   m.config.Version,
	}

	m.businessOperationDuration.With(opLabels).Observe(duration.Seconds())
}

// UpdateDatabaseConnections updates database connection metrics
func (m *Metrics) UpdateDatabaseConnections(active, idle int) {
	if !m.config.Enabled {
		return
	}

	m.dbConnectionsActive.Set(float64(active))
	m.dbConnectionsIdle.Set(float64(idle))
}

// UpdateSystemMetrics updates system metrics
func (m *Metrics) UpdateSystemMetrics(memoryUsage, cpuUsage float64, goroutines int) {
	if !m.config.Enabled {
		return
	}

	m.systemMemoryUsage.Set(memoryUsage)
	m.systemCPUUsage.Set(cpuUsage)
	m.systemGoroutines.Set(float64(goroutines))
}

// CreateCustomCounter creates a custom counter metric
func (m *Metrics) CreateCustomCounter(name, help string, labels []string) prometheus.CounterVec {
	if !m.config.Enabled {
		return prometheus.CounterVec{}
	}

	counter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)

	m.customMetrics[name] = counter
	return *counter
}

// CreateCustomGauge creates a custom gauge metric
func (m *Metrics) CreateCustomGauge(name, help string, labels []string) prometheus.GaugeVec {
	if !m.config.Enabled {
		return prometheus.GaugeVec{}
	}

	gauge := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)

	m.customMetrics[name] = gauge
	return *gauge
}

// CreateCustomHistogram creates a custom histogram metric
func (m *Metrics) CreateCustomHistogram(name, help string, labels []string, buckets []float64) prometheus.HistogramVec {
	if !m.config.Enabled {
		return prometheus.HistogramVec{}
	}

	histogram := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)

	m.customMetrics[name] = histogram
	return *histogram
}

// CreateCustomSummary creates a custom summary metric
func (m *Metrics) CreateCustomSummary(name, help string, labels []string) prometheus.SummaryVec {
	if !m.config.Enabled {
		return prometheus.SummaryVec{}
	}

	summary := promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: name,
			Help: help,
		},
		labels,
	)

	m.customMetrics[name] = summary
	return *summary
}

// GetCustomMetric returns a custom metric by name
func (m *Metrics) GetCustomMetric(name string) (prometheus.Collector, bool) {
	metric, exists := m.customMetrics[name]
	return metric, exists
}

// HealthCheck performs a health check on the metrics system
func (m *Metrics) HealthCheck(ctx context.Context) error {
	if !m.config.Enabled {
		return nil
	}

	// Check if metrics endpoint is accessible
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%s%s", m.config.Port, m.config.Path))
	if err != nil {
		return fmt.Errorf("metrics health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("metrics endpoint returned status %d", resp.StatusCode)
	}

	return nil
}

// Close closes the metrics system
func (m *Metrics) Close() error {
	// Cleanup any resources if needed
	return nil
}
