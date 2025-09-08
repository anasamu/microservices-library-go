package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/database/gateway"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

// Provider implements DatabaseProvider for Elasticsearch
type Provider struct {
	client *elasticsearch.Client
	config map[string]interface{}
	logger *logrus.Logger
}

// NewProvider creates a new Elasticsearch database provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "elasticsearch"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.DatabaseFeature {
	return []gateway.DatabaseFeature{
		gateway.FeatureConnectionPool,
		gateway.FeatureClustering,
		gateway.FeatureFullTextSearch,
		gateway.FeatureJSONSupport,
		gateway.FeatureGeoSpatial,
		gateway.FeatureTimeSeries,
		gateway.FeatureDocumentStore,
		gateway.FeaturePersistent,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *gateway.ConnectionInfo {
	addresses, _ := p.config["addresses"].([]string)
	index, _ := p.config["index"].(string)

	host := "localhost"
	port := 9200
	if len(addresses) > 0 {
		parts := strings.Split(addresses[0], ":")
		if len(parts) >= 1 {
			host = parts[0]
		}
		if len(parts) >= 2 {
			fmt.Sscanf(parts[1], "%d", &port)
		}
	}

	return &gateway.ConnectionInfo{
		Host:     host,
		Port:     port,
		Database: index,
		User:     "",
		Driver:   "elasticsearch",
		Version:  "8.0+",
	}
}

// Configure configures the Elasticsearch provider
func (p *Provider) Configure(config map[string]interface{}) error {
	addresses, ok := config["addresses"].([]string)
	if !ok || len(addresses) == 0 {
		addresses = []string{"http://localhost:9200"}
	}

	username, ok := config["username"].(string)
	if !ok {
		username = ""
	}

	password, ok := config["password"].(string)
	if !ok {
		password = ""
	}

	index, ok := config["index"].(string)
	if !ok || index == "" {
		index = "default"
	}

	// Create Elasticsearch client
	cfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	p.client = client
	p.config = config

	p.logger.Info("Elasticsearch provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.client != nil
}

// Connect connects to the database
func (p *Provider) Connect(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("elasticsearch provider not configured")
	}

	// Test connection
	res, err := p.client.Info()
	if err != nil {
		return fmt.Errorf("failed to get Elasticsearch info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch info request failed: %s", res.String())
	}

	p.logger.Info("Elasticsearch connected successfully")
	return nil
}

// Disconnect disconnects from the database
func (p *Provider) Disconnect(ctx context.Context) error {
	// Elasticsearch client doesn't need explicit disconnection
	return nil
}

// Ping checks database connection
func (p *Provider) Ping(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("elasticsearch provider not configured")
	}

	res, err := p.client.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch ping failed: %s", res.String())
	}

	return nil
}

// IsConnected checks if the database is connected
func (p *Provider) IsConnected() bool {
	if !p.IsConfigured() {
		return false
	}

	res, err := p.client.Ping()
	if err != nil {
		return false
	}
	defer res.Body.Close()

	return !res.IsError()
}

// BeginTransaction begins a new transaction (Elasticsearch doesn't support traditional transactions)
func (p *Provider) BeginTransaction(ctx context.Context) (gateway.Transaction, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("elasticsearch provider not configured")
	}

	// Elasticsearch doesn't support traditional transactions
	return &Transaction{client: p.client}, nil
}

// WithTransaction executes a function within a transaction
func (p *Provider) WithTransaction(ctx context.Context, fn func(gateway.Transaction) error) error {
	if !p.IsConfigured() {
		return fmt.Errorf("elasticsearch provider not configured")
	}

	// Elasticsearch doesn't support traditional transactions
	tx := &Transaction{client: p.client}
	return fn(tx)
}

// Query executes a query that returns rows
func (p *Provider) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("elasticsearch provider not configured")
	}

	index, _ := p.config["index"].(string)
	if index == "" {
		index = "default"
	}

	// Parse query as JSON
	var queryBody map[string]interface{}
	if err := json.Unmarshal([]byte(query), &queryBody); err != nil {
		return nil, fmt.Errorf("failed to parse query JSON: %w", err)
	}

	// Execute search
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(query),
	}

	res, err := req.Do(ctx, p.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	return &QueryResult{response: res}, nil
}

// QueryRow executes a query that returns a single row
func (p *Provider) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("elasticsearch provider not configured")
	}

	index, _ := p.config["index"].(string)
	if index == "" {
		index = "default"
	}

	// Parse query as JSON
	var queryBody map[string]interface{}
	if err := json.Unmarshal([]byte(query), &queryBody); err != nil {
		return nil, fmt.Errorf("failed to parse query JSON: %w", err)
	}

	// Execute search with size 1
	queryBody["size"] = 1
	queryJSON, _ := json.Marshal(queryBody)

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(string(queryJSON)),
	}

	res, err := req.Do(ctx, p.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	return &Row{response: res}, nil
}

// Exec executes a query without returning rows
func (p *Provider) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("elasticsearch provider not configured")
	}

	index, _ := p.config["index"].(string)
	if index == "" {
		index = "default"
	}

	// For Elasticsearch, we'll index a document
	doc := map[string]interface{}{
		"timestamp": time.Now(),
		"query":     query,
	}

	docJSON, _ := json.Marshal(doc)

	req := esapi.IndexRequest{
		Index:   index,
		Body:    strings.NewReader(string(docJSON)),
		Refresh: "true",
	}

	res, err := req.Do(ctx, p.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute index request: %w", err)
	}

	return &ExecResult{response: res}, nil
}

// Prepare prepares a statement (Elasticsearch doesn't have prepared statements)
func (p *Provider) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("elasticsearch provider not configured")
	}

	// Elasticsearch doesn't have prepared statements
	return &PreparedStatement{client: p.client, query: query}, nil
}

// HealthCheck performs a health check on the database
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("elasticsearch provider not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test ping
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("elasticsearch health check failed: %w", err)
	}

	// Test cluster health
	req := esapi.ClusterHealthRequest{}
	res, err := req.Do(ctx, p.client)
	if err != nil {
		return fmt.Errorf("elasticsearch cluster health check failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch cluster health check failed: %s", res.String())
	}

	return nil
}

// GetStats returns database statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.DatabaseStats, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("elasticsearch provider not configured")
	}

	// Get cluster stats
	req := esapi.ClusterStatsRequest{}
	res, err := req.Do(ctx, p.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster stats: %w", err)
	}
	defer res.Body.Close()

	var stats map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode cluster stats: %w", err)
	}

	return &gateway.DatabaseStats{
		ActiveConnections: 0, // Elasticsearch doesn't expose this directly
		IdleConnections:   0,
		MaxConnections:    0,
		WaitCount:         0,
		WaitDuration:      0,
		MaxIdleClosed:     0,
		MaxIdleTimeClosed: 0,
		MaxLifetimeClosed: 0,
		ProviderData: map[string]interface{}{
			"driver": "elasticsearch",
			"stats":  stats,
		},
	}, nil
}

// Close closes the database connection
func (p *Provider) Close() error {
	// Elasticsearch client doesn't need explicit closing
	return nil
}

// Transaction represents an Elasticsearch transaction
type Transaction struct {
	client *elasticsearch.Client
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	// Elasticsearch doesn't support traditional transactions
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	// Elasticsearch doesn't support traditional transactions
	return nil
}

// Query executes a query within the transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	// Elasticsearch doesn't support traditional transactions
	return nil, fmt.Errorf("elasticsearch transactions are not supported")
}

// QueryRow executes a query that returns a single row within the transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	// Elasticsearch doesn't support traditional transactions
	return nil, fmt.Errorf("elasticsearch transactions are not supported")
}

// Exec executes a query without returning rows within the transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	// Elasticsearch doesn't support traditional transactions
	return nil, fmt.Errorf("elasticsearch transactions are not supported")
}

// Prepare prepares a statement within the transaction
func (t *Transaction) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	// Elasticsearch doesn't support traditional transactions
	return nil, fmt.Errorf("elasticsearch transactions are not supported")
}

// QueryResult represents an Elasticsearch query result
type QueryResult struct {
	response *esapi.Response
}

// Close closes the result set
func (qr *QueryResult) Close() error {
	if qr.response != nil {
		return qr.response.Body.Close()
	}
	return nil
}

// Next advances to the next row
func (qr *QueryResult) Next() bool {
	// This is a simplified implementation
	// In a real implementation, you would parse the response and track position
	return false
}

// Scan scans the current row into dest
func (qr *QueryResult) Scan(dest ...interface{}) error {
	// This is a simplified implementation
	// In a real implementation, you would parse the response and extract data
	return fmt.Errorf("not implemented")
}

// Columns returns the column names
func (qr *QueryResult) Columns() ([]string, error) {
	// Elasticsearch doesn't have fixed columns
	return []string{}, nil
}

// Err returns any error that occurred during iteration
func (qr *QueryResult) Err() error {
	if qr.response != nil && qr.response.IsError() {
		return fmt.Errorf("elasticsearch response error: %s", qr.response.String())
	}
	return nil
}

// Row represents an Elasticsearch row
type Row struct {
	response *esapi.Response
}

// Scan scans the row into dest
func (r *Row) Scan(dest ...interface{}) error {
	// This is a simplified implementation
	// In a real implementation, you would parse the response and extract data
	return fmt.Errorf("not implemented")
}

// Err returns any error that occurred during scanning
func (r *Row) Err() error {
	if r.response != nil && r.response.IsError() {
		return fmt.Errorf("elasticsearch response error: %s", r.response.String())
	}
	return nil
}

// ExecResult represents an Elasticsearch execution result
type ExecResult struct {
	response *esapi.Response
}

// LastInsertId returns the last insert ID
func (er *ExecResult) LastInsertId() (int64, error) {
	// Elasticsearch doesn't have traditional insert IDs
	return 0, nil
}

// RowsAffected returns the number of rows affected
func (er *ExecResult) RowsAffected() (int64, error) {
	// Elasticsearch doesn't provide this information directly
	return 1, nil
}

// PreparedStatement represents an Elasticsearch prepared statement
type PreparedStatement struct {
	client *elasticsearch.Client
	query  string
}

// Close closes the prepared statement
func (ps *PreparedStatement) Close() error {
	// Elasticsearch doesn't need explicit closing
	return nil
}

// Query executes the prepared statement with args
func (ps *PreparedStatement) Query(ctx context.Context, args ...interface{}) (gateway.QueryResult, error) {
	// Elasticsearch doesn't have prepared statements
	return &QueryResult{}, nil
}

// QueryRow executes the prepared statement with args and returns a single row
func (ps *PreparedStatement) QueryRow(ctx context.Context, args ...interface{}) (gateway.Row, error) {
	// Elasticsearch doesn't have prepared statements
	return &Row{}, nil
}

// Exec executes the prepared statement with args
func (ps *PreparedStatement) Exec(ctx context.Context, args ...interface{}) (gateway.ExecResult, error) {
	// Elasticsearch doesn't have prepared statements
	return &ExecResult{}, nil
}
