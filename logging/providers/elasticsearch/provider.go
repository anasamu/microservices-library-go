package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/logging/types"
	"github.com/elastic/go-elasticsearch/v8"
)

// ElasticsearchProvider implements logging to Elasticsearch
type ElasticsearchProvider struct {
	name      string
	client    *elasticsearch.Client
	config    *types.LoggingConfig
	connected bool
	mu        sync.RWMutex
}

// NewElasticsearchProvider creates a new Elasticsearch logging provider
func NewElasticsearchProvider(config *types.LoggingConfig) (*ElasticsearchProvider, error) {
	if config == nil {
		config = types.GetDefaultConfig()
	}

	// Get Elasticsearch URL from config
	elasticURL := ""
	if config.Metadata != nil {
		if url, ok := config.Metadata["elastic_url"].(string); ok {
			elasticURL = url
		}
	}

	if elasticURL == "" {
		return nil, fmt.Errorf("elasticsearch URL not provided in config metadata")
	}

	// Create Elasticsearch client
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{elasticURL},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	return &ElasticsearchProvider{
		name:      "elasticsearch",
		client:    esClient,
		config:    config,
		connected: false,
	}, nil
}

// GetName returns the provider name
func (ep *ElasticsearchProvider) GetName() string {
	return ep.name
}

// GetSupportedFeatures returns supported features
func (ep *ElasticsearchProvider) GetSupportedFeatures() []types.LoggingFeature {
	return []types.LoggingFeature{
		types.FeatureBasicLogging,
		types.FeatureStructuredLog,
		types.FeatureFormattedLog,
		types.FeatureContextLogging,
		types.FeatureBatchLogging,
		types.FeatureSearch,
		types.FeatureFiltering,
		types.FeatureAggregation,
		types.FeatureRetention,
		types.FeatureCompression,
		types.FeatureRealTime,
	}
}

// GetConnectionInfo returns connection information
func (ep *ElasticsearchProvider) GetConnectionInfo() *types.ConnectionInfo {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	status := types.StatusDisconnected
	if ep.connected {
		status = types.StatusConnected
	}

	elasticURL := ""
	if ep.config.Metadata != nil {
		if url, ok := ep.config.Metadata["elastic_url"].(string); ok {
			elasticURL = url
		}
	}

	return &types.ConnectionInfo{
		Host:     elasticURL,
		Port:     9200, // Default Elasticsearch port
		Database: ep.config.Index,
		Username: "",
		Status:   status,
		Metadata: map[string]string{
			"index": ep.config.Index,
			"url":   elasticURL,
		},
	}
}

// Connect connects to Elasticsearch
func (ep *ElasticsearchProvider) Connect(ctx context.Context) error {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	// Test connection
	res, err := ep.client.Info()
	if err != nil {
		return fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch connection error: %s", res.String())
	}

	ep.connected = true
	return nil
}

// Disconnect disconnects from Elasticsearch
func (ep *ElasticsearchProvider) Disconnect(ctx context.Context) error {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.connected = false
	return nil
}

// Ping checks if Elasticsearch is available
func (ep *ElasticsearchProvider) Ping(ctx context.Context) error {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	if !ep.connected {
		return fmt.Errorf("elasticsearch provider not connected")
	}

	res, err := ep.client.Ping()
	if err != nil {
		return fmt.Errorf("elasticsearch ping failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch ping error: %s", res.String())
	}

	return nil
}

// IsConnected returns connection status
func (ep *ElasticsearchProvider) IsConnected() bool {
	ep.mu.RLock()
	defer ep.mu.RUnlock()
	return ep.connected
}

// Log logs a message with the specified level
func (ep *ElasticsearchProvider) Log(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	if !ep.connected {
		return fmt.Errorf("elasticsearch provider not connected")
	}

	// Create log entry
	entry := types.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Service:   ep.config.Service,
		Version:   ep.config.Version,
		Provider:  ep.name,
		Fields:    fields,
	}

	// Add context fields if available
	if traceID := ctx.Value("trace_id"); traceID != nil {
		entry.TraceID = fmt.Sprintf("%v", traceID)
	}
	if spanID := ctx.Value("span_id"); spanID != nil {
		entry.SpanID = fmt.Sprintf("%v", spanID)
	}
	if requestID := ctx.Value("request_id"); requestID != nil {
		entry.RequestID = fmt.Sprintf("%v", requestID)
	}
	if userID := ctx.Value("user_id"); userID != nil {
		entry.UserID = fmt.Sprintf("%v", userID)
	}

	return ep.sendToElasticsearch(entry)
}

// LogWithContext logs a message with context
func (ep *ElasticsearchProvider) LogWithContext(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	return ep.Log(ctx, level, message, fields)
}

// Info logs an info level message
func (ep *ElasticsearchProvider) Info(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return ep.LogWithContext(ctx, types.LevelInfo, message, mergedFields)
}

// Debug logs a debug level message
func (ep *ElasticsearchProvider) Debug(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return ep.LogWithContext(ctx, types.LevelDebug, message, mergedFields)
}

// Warn logs a warning level message
func (ep *ElasticsearchProvider) Warn(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return ep.LogWithContext(ctx, types.LevelWarn, message, mergedFields)
}

// Error logs an error level message
func (ep *ElasticsearchProvider) Error(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return ep.LogWithContext(ctx, types.LevelError, message, mergedFields)
}

// Fatal logs a fatal level message
func (ep *ElasticsearchProvider) Fatal(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return ep.LogWithContext(ctx, types.LevelFatal, message, mergedFields)
}

// Panic logs a panic level message
func (ep *ElasticsearchProvider) Panic(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return ep.LogWithContext(ctx, types.LevelPanic, message, mergedFields)
}

// Infof logs a formatted info level message
func (ep *ElasticsearchProvider) Infof(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return ep.LogWithContext(ctx, types.LevelInfo, message, nil)
}

// Debugf logs a formatted debug level message
func (ep *ElasticsearchProvider) Debugf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return ep.LogWithContext(ctx, types.LevelDebug, message, nil)
}

// Warnf logs a formatted warning level message
func (ep *ElasticsearchProvider) Warnf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return ep.LogWithContext(ctx, types.LevelWarn, message, nil)
}

// Errorf logs a formatted error level message
func (ep *ElasticsearchProvider) Errorf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return ep.LogWithContext(ctx, types.LevelError, message, nil)
}

// Fatalf logs a formatted fatal level message
func (ep *ElasticsearchProvider) Fatalf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return ep.LogWithContext(ctx, types.LevelFatal, message, nil)
}

// Panicf logs a formatted panic level message
func (ep *ElasticsearchProvider) Panicf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return ep.LogWithContext(ctx, types.LevelPanic, message, nil)
}

// WithFields creates a new provider with additional fields
func (ep *ElasticsearchProvider) WithFields(ctx context.Context, fields map[string]interface{}) types.LoggingProvider {
	return &ElasticsearchProviderWithFields{
		provider: ep,
		fields:   fields,
	}
}

// WithField creates a new provider with a single additional field
func (ep *ElasticsearchProvider) WithField(ctx context.Context, key string, value interface{}) types.LoggingProvider {
	return ep.WithFields(ctx, map[string]interface{}{key: value})
}

// WithError creates a new provider with error information
func (ep *ElasticsearchProvider) WithError(ctx context.Context, err error) types.LoggingProvider {
	return ep.WithField(ctx, "error", err.Error())
}

// WithContext creates a new provider with context information
func (ep *ElasticsearchProvider) WithContext(ctx context.Context) types.LoggingProvider {
	return ep
}

// LogBatch logs multiple entries
func (ep *ElasticsearchProvider) LogBatch(ctx context.Context, entries []types.LogEntry) error {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	if !ep.connected {
		return fmt.Errorf("elasticsearch provider not connected")
	}

	// Create bulk request
	var bulkBody strings.Builder
	for _, entry := range entries {
		// Add index action
		indexAction := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": ep.getIndexName(),
			},
		}
		indexActionJSON, _ := json.Marshal(indexAction)
		bulkBody.WriteString(string(indexActionJSON))
		bulkBody.WriteString("\n")

		// Add document
		entryJSON, _ := json.Marshal(entry)
		bulkBody.WriteString(string(entryJSON))
		bulkBody.WriteString("\n")
	}

	// Send bulk request
	res, err := ep.client.Bulk(
		strings.NewReader(bulkBody.String()),
		ep.client.Bulk.WithRefresh("false"),
	)
	if err != nil {
		return fmt.Errorf("failed to send bulk logs to elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch bulk error: %s", res.String())
	}

	return nil
}

// Search searches logs
func (ep *ElasticsearchProvider) Search(ctx context.Context, query types.LogQuery) ([]types.LogEntry, error) {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	if !ep.connected {
		return nil, fmt.Errorf("elasticsearch provider not connected")
	}

	// Build Elasticsearch query
	esQuery := ep.buildElasticsearchQuery(query)

	// Execute search
	res, err := ep.client.Search(
		ep.client.Search.WithIndex(ep.getIndexName()),
		ep.client.Search.WithBody(strings.NewReader(esQuery)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch search error: %s", res.String())
	}

	// Parse response
	var searchResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Extract hits
	hits, ok := searchResponse["hits"].(map[string]interface{})
	if !ok {
		return []types.LogEntry{}, nil
	}

	hitsArray, ok := hits["hits"].([]interface{})
	if !ok {
		return []types.LogEntry{}, nil
	}

	// Convert to LogEntry
	entries := make([]types.LogEntry, 0, len(hitsArray))
	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		entry := ep.convertToLogEntry(source)
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetStats returns logging statistics
func (ep *ElasticsearchProvider) GetStats(ctx context.Context) (*types.LoggingStats, error) {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	if !ep.connected {
		return nil, fmt.Errorf("elasticsearch provider not connected")
	}

	// Get index stats
	res, err := ep.client.Indices.Stats(
		ep.client.Indices.Stats.WithIndex(ep.getIndexName()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get elasticsearch stats: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch stats error: %s", res.String())
	}

	// Parse response
	var statsResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&statsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode stats response: %w", err)
	}

	// Extract stats
	indices, ok := statsResponse["indices"].(map[string]interface{})
	if !ok {
		return &types.LoggingStats{
			Provider: ep.name,
		}, nil
	}

	var totalLogs int64
	var storageSize int64

	for _, indexStats := range indices {
		indexMap, ok := indexStats.(map[string]interface{})
		if !ok {
			continue
		}

		// Get total documents
		if total, ok := indexMap["total"].(map[string]interface{}); ok {
			if docs, ok := total["docs"].(map[string]interface{}); ok {
				if count, ok := docs["count"].(float64); ok {
					totalLogs += int64(count)
				}
			}
		}

		// Get storage size
		if store, ok := indexMap["total"].(map[string]interface{}); ok {
			if size, ok := store["store"].(map[string]interface{}); ok {
				if sizeInBytes, ok := size["size_in_bytes"].(float64); ok {
					storageSize += int64(sizeInBytes)
				}
			}
		}
	}

	return &types.LoggingStats{
		TotalLogs:     totalLogs,
		LogsByLevel:   make(map[types.LogLevel]int64),
		LogsByService: make(map[string]int64),
		StorageSize:   storageSize,
		Uptime:        time.Since(time.Now()), // Placeholder
		LastUpdate:    time.Now(),
		Provider:      ep.name,
	}, nil
}

// sendToElasticsearch sends a log entry to Elasticsearch
func (ep *ElasticsearchProvider) sendToElasticsearch(entry types.LogEntry) error {
	// Marshal log entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Send to Elasticsearch
	res, err := ep.client.Index(
		ep.getIndexName(),
		strings.NewReader(string(data)),
		ep.client.Index.WithRefresh("false"),
	)
	if err != nil {
		return fmt.Errorf("failed to send log to elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch error: %s", res.String())
	}

	return nil
}

// getIndexName returns the index name with date suffix
func (ep *ElasticsearchProvider) getIndexName() string {
	indexName := ep.config.Index
	if indexName == "" {
		indexName = "logs"
	}
	return fmt.Sprintf("%s-%s", indexName, time.Now().Format("2006.01.02"))
}

// buildElasticsearchQuery builds an Elasticsearch query from LogQuery
func (ep *ElasticsearchProvider) buildElasticsearchQuery(query types.LogQuery) string {
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{},
			},
		},
	}

	boolQuery := esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})
	must := boolQuery["must"].([]interface{})

	// Add level filter
	if len(query.Levels) > 0 {
		levelTerms := make([]string, len(query.Levels))
		for i, level := range query.Levels {
			levelTerms[i] = string(level)
		}
		must = append(must, map[string]interface{}{
			"terms": map[string]interface{}{
				"level": levelTerms,
			},
		})
	}

	// Add service filter
	if len(query.Services) > 0 {
		must = append(must, map[string]interface{}{
			"terms": map[string]interface{}{
				"service": query.Services,
			},
		})
	}

	// Add time range filter
	if query.StartTime != nil || query.EndTime != nil {
		timeRange := map[string]interface{}{}
		if query.StartTime != nil {
			timeRange["gte"] = query.StartTime.Format(time.RFC3339)
		}
		if query.EndTime != nil {
			timeRange["lte"] = query.EndTime.Format(time.RFC3339)
		}
		must = append(must, map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": timeRange,
			},
		})
	}

	// Add message search
	if query.Message != "" {
		must = append(must, map[string]interface{}{
			"match": map[string]interface{}{
				"message": query.Message,
			},
		})
	}

	// Add field filters
	for key, value := range query.Fields {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				key: value,
			},
		})
	}

	boolQuery["must"] = must

	// Add sorting
	if query.SortBy != "" {
		sortOrder := "asc"
		if query.SortOrder == "desc" {
			sortOrder = "desc"
		}
		esQuery["sort"] = []interface{}{
			map[string]interface{}{
				query.SortBy: map[string]interface{}{
					"order": sortOrder,
				},
			},
		}
	}

	// Add pagination
	if query.Limit > 0 {
		esQuery["size"] = query.Limit
	}
	if query.Offset > 0 {
		esQuery["from"] = query.Offset
	}

	queryJSON, _ := json.Marshal(esQuery)
	return string(queryJSON)
}

// convertToLogEntry converts Elasticsearch document to LogEntry
func (ep *ElasticsearchProvider) convertToLogEntry(source map[string]interface{}) types.LogEntry {
	entry := types.LogEntry{}

	if timestamp, ok := source["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
			entry.Timestamp = t
		}
	}

	if level, ok := source["level"].(string); ok {
		entry.Level = types.LogLevel(level)
	}

	if message, ok := source["message"].(string); ok {
		entry.Message = message
	}

	if service, ok := source["service"].(string); ok {
		entry.Service = service
	}

	if version, ok := source["version"].(string); ok {
		entry.Version = version
	}

	if traceID, ok := source["trace_id"].(string); ok {
		entry.TraceID = traceID
	}

	if spanID, ok := source["span_id"].(string); ok {
		entry.SpanID = spanID
	}

	if userID, ok := source["user_id"].(string); ok {
		entry.UserID = userID
	}

	if tenantID, ok := source["tenant_id"].(string); ok {
		entry.TenantID = tenantID
	}

	if requestID, ok := source["request_id"].(string); ok {
		entry.RequestID = requestID
	}

	if ipAddress, ok := source["ip_address"].(string); ok {
		entry.IPAddress = ipAddress
	}

	if userAgent, ok := source["user_agent"].(string); ok {
		entry.UserAgent = userAgent
	}

	if method, ok := source["method"].(string); ok {
		entry.Method = method
	}

	if path, ok := source["path"].(string); ok {
		entry.Path = path
	}

	if statusCode, ok := source["status_code"].(float64); ok {
		entry.StatusCode = int(statusCode)
	}

	if duration, ok := source["duration"].(float64); ok {
		entry.Duration = int64(duration)
	}

	if error, ok := source["error"].(string); ok {
		entry.Error = error
	}

	if provider, ok := source["provider"].(string); ok {
		entry.Provider = provider
	}

	// Handle fields
	if fields, ok := source["fields"].(map[string]interface{}); ok {
		entry.Fields = fields
	}

	return entry
}

// ElasticsearchProviderWithFields wraps ElasticsearchProvider with additional fields
type ElasticsearchProviderWithFields struct {
	provider *ElasticsearchProvider
	fields   map[string]interface{}
}

// Implement all LoggingProvider methods by delegating to the wrapped provider
// and merging the additional fields

func (epwf *ElasticsearchProviderWithFields) GetName() string {
	return epwf.provider.GetName()
}

func (epwf *ElasticsearchProviderWithFields) GetSupportedFeatures() []types.LoggingFeature {
	return epwf.provider.GetSupportedFeatures()
}

func (epwf *ElasticsearchProviderWithFields) GetConnectionInfo() *types.ConnectionInfo {
	return epwf.provider.GetConnectionInfo()
}

func (epwf *ElasticsearchProviderWithFields) Connect(ctx context.Context) error {
	return epwf.provider.Connect(ctx)
}

func (epwf *ElasticsearchProviderWithFields) Disconnect(ctx context.Context) error {
	return epwf.provider.Disconnect(ctx)
}

func (epwf *ElasticsearchProviderWithFields) Ping(ctx context.Context) error {
	return epwf.provider.Ping(ctx)
}

func (epwf *ElasticsearchProviderWithFields) IsConnected() bool {
	return epwf.provider.IsConnected()
}

func (epwf *ElasticsearchProviderWithFields) Log(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	mergedFields := make(map[string]interface{})

	// Add wrapper fields first
	for k, v := range epwf.fields {
		mergedFields[k] = v
	}

	// Add provided fields (they override wrapper fields)
	if fields != nil {
		for k, v := range fields {
			mergedFields[k] = v
		}
	}

	return epwf.provider.Log(ctx, level, message, mergedFields)
}

func (epwf *ElasticsearchProviderWithFields) LogWithContext(ctx context.Context, level types.LogLevel, message string, fields map[string]interface{}) error {
	mergedFields := make(map[string]interface{})

	// Add wrapper fields first
	for k, v := range epwf.fields {
		mergedFields[k] = v
	}

	// Add provided fields (they override wrapper fields)
	if fields != nil {
		for k, v := range fields {
			mergedFields[k] = v
		}
	}

	return epwf.provider.LogWithContext(ctx, level, message, mergedFields)
}

func (epwf *ElasticsearchProviderWithFields) Info(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return epwf.LogWithContext(ctx, types.LevelInfo, message, mergedFields)
}

func (epwf *ElasticsearchProviderWithFields) Debug(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return epwf.LogWithContext(ctx, types.LevelDebug, message, mergedFields)
}

func (epwf *ElasticsearchProviderWithFields) Warn(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return epwf.LogWithContext(ctx, types.LevelWarn, message, mergedFields)
}

func (epwf *ElasticsearchProviderWithFields) Error(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return epwf.LogWithContext(ctx, types.LevelError, message, mergedFields)
}

func (epwf *ElasticsearchProviderWithFields) Fatal(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return epwf.LogWithContext(ctx, types.LevelFatal, message, mergedFields)
}

func (epwf *ElasticsearchProviderWithFields) Panic(ctx context.Context, message string, fields ...map[string]interface{}) error {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	return epwf.LogWithContext(ctx, types.LevelPanic, message, mergedFields)
}

func (epwf *ElasticsearchProviderWithFields) Infof(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return epwf.LogWithContext(ctx, types.LevelInfo, message, nil)
}

func (epwf *ElasticsearchProviderWithFields) Debugf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return epwf.LogWithContext(ctx, types.LevelDebug, message, nil)
}

func (epwf *ElasticsearchProviderWithFields) Warnf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return epwf.LogWithContext(ctx, types.LevelWarn, message, nil)
}

func (epwf *ElasticsearchProviderWithFields) Errorf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return epwf.LogWithContext(ctx, types.LevelError, message, nil)
}

func (epwf *ElasticsearchProviderWithFields) Fatalf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return epwf.LogWithContext(ctx, types.LevelFatal, message, nil)
}

func (epwf *ElasticsearchProviderWithFields) Panicf(ctx context.Context, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return epwf.LogWithContext(ctx, types.LevelPanic, message, nil)
}

func (epwf *ElasticsearchProviderWithFields) WithFields(ctx context.Context, fields map[string]interface{}) types.LoggingProvider {
	// Merge fields
	mergedFields := make(map[string]interface{})
	for k, v := range epwf.fields {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}

	return &ElasticsearchProviderWithFields{
		provider: epwf.provider,
		fields:   mergedFields,
	}
}

func (epwf *ElasticsearchProviderWithFields) WithField(ctx context.Context, key string, value interface{}) types.LoggingProvider {
	return epwf.WithFields(ctx, map[string]interface{}{key: value})
}

func (epwf *ElasticsearchProviderWithFields) WithError(ctx context.Context, err error) types.LoggingProvider {
	return epwf.WithField(ctx, "error", err.Error())
}

func (epwf *ElasticsearchProviderWithFields) WithContext(ctx context.Context) types.LoggingProvider {
	return epwf
}

func (epwf *ElasticsearchProviderWithFields) LogBatch(ctx context.Context, entries []types.LogEntry) error {
	return epwf.provider.LogBatch(ctx, entries)
}

func (epwf *ElasticsearchProviderWithFields) Search(ctx context.Context, query types.LogQuery) ([]types.LogEntry, error) {
	return epwf.provider.Search(ctx, query)
}

func (epwf *ElasticsearchProviderWithFields) GetStats(ctx context.Context) (*types.LoggingStats, error) {
	return epwf.provider.GetStats(ctx)
}
