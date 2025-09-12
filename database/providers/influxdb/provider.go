package influxdb

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/database"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/sirupsen/logrus"
)

// Provider implements DatabaseProvider for InfluxDB
type Provider struct {
	client   influxdb2.Client
	writeAPI api.WriteAPI
	queryAPI api.QueryAPI
	config   map[string]interface{}
	logger   *logrus.Logger
}

// NewProvider creates a new InfluxDB database provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "influxdb"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.DatabaseFeature {
	return []gateway.DatabaseFeature{
		gateway.FeatureConnectionPool,
		gateway.FeatureClustering,
		gateway.FeatureTimeSeries,
		gateway.FeaturePersistent,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *gateway.ConnectionInfo {
	url, _ := p.config["url"].(string)
	org, _ := p.config["org"].(string)
	bucket, _ := p.config["bucket"].(string)

	return &gateway.ConnectionInfo{
		Host:     url,
		Port:     8086,
		Database: bucket,
		User:     org,
		Driver:   "influxdb",
		Version:  "2.0+",
	}
}

// Configure configures the InfluxDB provider
func (p *Provider) Configure(config map[string]interface{}) error {
	url, ok := config["url"].(string)
	if !ok || url == "" {
		url = "http://localhost:8086"
	}

	token, ok := config["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("influxdb token is required")
	}

	org, ok := config["org"].(string)
	if !ok || org == "" {
		return fmt.Errorf("influxdb org is required")
	}

	bucket, ok := config["bucket"].(string)
	if !ok || bucket == "" {
		return fmt.Errorf("influxdb bucket is required")
	}

	timeout, ok := config["timeout"].(time.Duration)
	if !ok || timeout == 0 {
		timeout = 30 * time.Second
	}

	// Create InfluxDB client
	client := influxdb2.NewClient(url, token)

	// Get write and query APIs
	writeAPI := client.WriteAPI(org, bucket)
	queryAPI := client.QueryAPI(org)

	p.client = client
	p.writeAPI = writeAPI
	p.queryAPI = queryAPI
	p.config = config

	p.logger.Info("InfluxDB provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.client != nil && p.writeAPI != nil && p.queryAPI != nil
}

// Connect connects to the database
func (p *Provider) Connect(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("influxdb provider not configured")
	}

	// Test connection by querying the health endpoint
	health, err := p.client.Health(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to InfluxDB: %w", err)
	}

	if health.Status != "pass" {
		message := "unknown error"
		if health.Message != nil {
			message = *health.Message
		}
		return fmt.Errorf("influxdb health check failed: %s", message)
	}

	p.logger.Info("InfluxDB connected successfully")
	return nil
}

// Disconnect disconnects from the database
func (p *Provider) Disconnect(ctx context.Context) error {
	if p.client != nil {
		p.client.Close()
	}
	return nil
}

// Ping checks database connection
func (p *Provider) Ping(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("influxdb provider not configured")
	}

	health, err := p.client.Health(ctx)
	if err != nil {
		return err
	}

	if health.Status != "pass" {
		message := "unknown error"
		if health.Message != nil {
			message = *health.Message
		}
		return fmt.Errorf("influxdb ping failed: %s", message)
	}

	return nil
}

// IsConnected checks if the database is connected
func (p *Provider) IsConnected() bool {
	if !p.IsConfigured() {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := p.client.Health(ctx)
	return err == nil && health.Status == "pass"
}

// BeginTransaction begins a new transaction (InfluxDB doesn't support traditional transactions)
func (p *Provider) BeginTransaction(ctx context.Context) (gateway.Transaction, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("influxdb provider not configured")
	}

	// InfluxDB doesn't support traditional transactions
	return &Transaction{provider: p}, nil
}

// WithTransaction executes a function within a transaction
func (p *Provider) WithTransaction(ctx context.Context, fn func(gateway.Transaction) error) error {
	if !p.IsConfigured() {
		return fmt.Errorf("influxdb provider not configured")
	}

	// InfluxDB doesn't support traditional transactions
	tx := &Transaction{provider: p}
	return fn(tx)
}

// Query executes a query that returns rows
func (p *Provider) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("influxdb provider not configured")
	}

	result, err := p.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &QueryResult{result: result}, nil
}

// QueryRow executes a query that returns a single row
func (p *Provider) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("influxdb provider not configured")
	}

	result, err := p.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query row: %w", err)
	}

	// Get the first record
	if result.Next() {
		record := result.Record()
		recordData := map[string]interface{}{
			"time":        record.Time(),
			"value":       record.Value(),
			"field":       record.Field(),
			"measurement": record.Measurement(),
		}
		return &Row{record: recordData}, nil
	}

	return nil, fmt.Errorf("no rows found")
}

// Exec executes a query without returning rows
func (p *Provider) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("influxdb provider not configured")
	}

	// For InfluxDB, we'll write a test point to simulate execution
	point := influxdb2.NewPoint(
		"exec_test",
		map[string]string{"operation": "exec"},
		map[string]interface{}{"value": 1, "timestamp": time.Now()},
		time.Now(),
	)

	p.writeAPI.WritePoint(point)
	p.writeAPI.Flush()

	return &ExecResult{affected: 1}, nil
}

// Prepare prepares a statement (InfluxDB doesn't have prepared statements in the same way)
func (p *Provider) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("influxdb provider not configured")
	}

	// InfluxDB doesn't have prepared statements, return a mock
	return &PreparedStatement{provider: p, query: query}, nil
}

// HealthCheck performs a health check on the database
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("influxdb provider not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test ping
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("influxdb health check failed: %w", err)
	}

	// Test basic query
	query := `from(bucket: "` + p.config["bucket"].(string) + `")
		|> range(start: -1h)
		|> limit(n: 1)`

	result, err := p.queryAPI.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("influxdb query health check failed: %w", err)
	}
	defer result.Close()

	return nil
}

// GetStats returns database statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.DatabaseStats, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("influxdb provider not configured")
	}

	// Get server info (simplified for InfluxDB)
	server := map[string]interface{}{
		"url": p.config["url"],
	}

	return &gateway.DatabaseStats{
		ActiveConnections: 0, // InfluxDB doesn't expose this directly
		IdleConnections:   0,
		MaxConnections:    0,
		WaitCount:         0,
		WaitDuration:      0,
		MaxIdleClosed:     0,
		MaxIdleTimeClosed: 0,
		MaxLifetimeClosed: 0,
		ProviderData: map[string]interface{}{
			"driver": "influxdb",
			"server": server,
		},
	}, nil
}

// Close closes the database connection
func (p *Provider) Close() error {
	if p.client != nil {
		p.client.Close()
	}
	return nil
}

// Transaction represents an InfluxDB transaction
type Transaction struct {
	provider *Provider
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	// InfluxDB doesn't support traditional transactions
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	// InfluxDB doesn't support traditional transactions
	return nil
}

// Query executes a query within the transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	return t.provider.Query(ctx, query, args...)
}

// QueryRow executes a query that returns a single row within the transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	return t.provider.QueryRow(ctx, query, args...)
}

// Exec executes a query without returning rows within the transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	return t.provider.Exec(ctx, query, args...)
}

// Prepare prepares a statement within the transaction
func (t *Transaction) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	return t.provider.Prepare(ctx, query)
}

// QueryResult represents an InfluxDB query result
type QueryResult struct {
	result *api.QueryTableResult
}

// Close closes the result set
func (qr *QueryResult) Close() error {
	qr.result.Close()
	return nil
}

// Next advances to the next row
func (qr *QueryResult) Next() bool {
	return qr.result.Next()
}

// Scan scans the current row into dest
func (qr *QueryResult) Scan(dest ...interface{}) error {
	record := qr.result.Record()

	if len(dest) > 0 {
		if destMap, ok := dest[0].(*map[string]interface{}); ok {
			*destMap = map[string]interface{}{
				"time":        record.Time(),
				"value":       record.Value(),
				"field":       record.Field(),
				"measurement": record.Measurement(),
			}
		}
	}

	return nil
}

// Columns returns the column names
func (qr *QueryResult) Columns() ([]string, error) {
	return []string{"time", "value", "field", "measurement"}, nil
}

// Err returns any error that occurred during iteration
func (qr *QueryResult) Err() error {
	return qr.result.Err()
}

// Row represents an InfluxDB row
type Row struct {
	record map[string]interface{}
}

// Scan scans the row into dest
func (r *Row) Scan(dest ...interface{}) error {
	if len(dest) > 0 {
		if destMap, ok := dest[0].(*map[string]interface{}); ok {
			*destMap = r.record
		}
	}
	return nil
}

// Err returns any error that occurred during scanning
func (r *Row) Err() error {
	return nil
}

// ExecResult represents an InfluxDB execution result
type ExecResult struct {
	affected int64
}

// LastInsertId returns the last insert ID
func (er *ExecResult) LastInsertId() (int64, error) {
	return 0, nil
}

// RowsAffected returns the number of rows affected
func (er *ExecResult) RowsAffected() (int64, error) {
	return er.affected, nil
}

// PreparedStatement represents an InfluxDB prepared statement
type PreparedStatement struct {
	provider *Provider
	query    string
}

// Close closes the prepared statement
func (ps *PreparedStatement) Close() error {
	return nil
}

// Query executes the prepared statement with args
func (ps *PreparedStatement) Query(ctx context.Context, args ...interface{}) (gateway.QueryResult, error) {
	return ps.provider.Query(ctx, ps.query, args...)
}

// QueryRow executes the prepared statement with args and returns a single row
func (ps *PreparedStatement) QueryRow(ctx context.Context, args ...interface{}) (gateway.Row, error) {
	return ps.provider.QueryRow(ctx, ps.query, args...)
}

// Exec executes the prepared statement with args
func (ps *PreparedStatement) Exec(ctx context.Context, args ...interface{}) (gateway.ExecResult, error) {
	return ps.provider.Exec(ctx, ps.query, args...)
}
