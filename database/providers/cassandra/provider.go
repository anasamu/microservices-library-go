package cassandra

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/database"
	"github.com/gocql/gocql"
	"github.com/sirupsen/logrus"
)

// Provider implements DatabaseProvider for Cassandra
type Provider struct {
	session *gocql.Session
	cluster *gocql.ClusterConfig
	config  map[string]interface{}
	logger  *logrus.Logger
}

// NewProvider creates a new Cassandra database provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "cassandra"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.DatabaseFeature {
	return []gateway.DatabaseFeature{
		gateway.FeatureConnectionPool,
		gateway.FeatureClustering,
		gateway.FeatureColumnFamily,
		gateway.FeaturePersistent,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *gateway.ConnectionInfo {
	hosts, _ := p.config["hosts"].([]string)
	keyspace, _ := p.config["keyspace"].(string)

	host := "localhost"
	if len(hosts) > 0 {
		host = hosts[0]
	}

	return &gateway.ConnectionInfo{
		Host:     host,
		Port:     9042,
		Database: keyspace,
		User:     "",
		Driver:   "cassandra",
		Version:  "3.11+",
	}
}

// Configure configures the Cassandra provider
func (p *Provider) Configure(config map[string]interface{}) error {
	hosts, ok := config["hosts"].([]string)
	if !ok || len(hosts) == 0 {
		hosts = []string{"localhost"}
	}

	keyspace, ok := config["keyspace"].(string)
	if !ok || keyspace == "" {
		return fmt.Errorf("cassandra keyspace is required")
	}

	username, ok := config["username"].(string)
	if !ok {
		username = ""
	}

	password, ok := config["password"].(string)
	if !ok {
		password = ""
	}

	consistency, ok := config["consistency"].(string)
	if !ok || consistency == "" {
		consistency = "quorum"
	}

	timeout, ok := config["timeout"].(time.Duration)
	if !ok || timeout == 0 {
		timeout = 10 * time.Second
	}

	numConns, ok := config["num_connections"].(int)
	if !ok || numConns == 0 {
		numConns = 2
	}

	// Create cluster configuration
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Timeout = timeout
	cluster.NumConns = numConns
	cluster.Consistency = gocql.ParseConsistency(consistency)

	if username != "" && password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: username,
			Password: password,
		}
	}

	// Create session
	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create Cassandra session: %w", err)
	}

	p.session = session
	p.cluster = cluster
	p.config = config

	p.logger.Info("Cassandra provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.session != nil
}

// Connect connects to the database
func (p *Provider) Connect(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("cassandra provider not configured")
	}

	// Test connection with a simple query
	var result int
	if err := p.session.Query("SELECT now() FROM system.local").Scan(&result); err != nil {
		return fmt.Errorf("failed to connect to Cassandra: %w", err)
	}

	p.logger.Info("Cassandra connected successfully")
	return nil
}

// Disconnect disconnects from the database
func (p *Provider) Disconnect(ctx context.Context) error {
	if p.session != nil {
		p.session.Close()
	}
	return nil
}

// Ping checks database connection
func (p *Provider) Ping(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("cassandra provider not configured")
	}

	var result int
	return p.session.Query("SELECT now() FROM system.local").Scan(&result)
}

// IsConnected checks if the database is connected
func (p *Provider) IsConnected() bool {
	if !p.IsConfigured() {
		return false
	}

	var result int
	return p.session.Query("SELECT now() FROM system.local").Scan(&result) == nil
}

// BeginTransaction begins a new transaction (Cassandra has limited transaction support)
func (p *Provider) BeginTransaction(ctx context.Context) (gateway.Transaction, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("cassandra provider not configured")
	}

	// Cassandra has limited transaction support, mainly for single partition operations
	return &Transaction{session: p.session}, nil
}

// WithTransaction executes a function within a transaction
func (p *Provider) WithTransaction(ctx context.Context, fn func(gateway.Transaction) error) error {
	if !p.IsConfigured() {
		return fmt.Errorf("cassandra provider not configured")
	}

	// Cassandra has limited transaction support
	tx := &Transaction{session: p.session}
	return fn(tx)
}

// Query executes a query that returns rows
func (p *Provider) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("cassandra provider not configured")
	}

	iter := p.session.Query(query, args...).Iter()
	return &QueryResult{iter: iter}, nil
}

// QueryRow executes a query that returns a single row
func (p *Provider) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("cassandra provider not configured")
	}

	iter := p.session.Query(query, args...).Iter()
	if !iter.Scan() {
		return nil, fmt.Errorf("no rows found")
	}

	// Get the current row data
	rowData := make(map[string]interface{})
	iter.MapScan(rowData)

	return &Row{data: rowData}, nil
}

// Exec executes a query without returning rows
func (p *Provider) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("cassandra provider not configured")
	}

	err := p.session.Query(query, args...).Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &ExecResult{affected: 1}, nil
}

// Prepare prepares a statement
func (p *Provider) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("cassandra provider not configured")
	}

	stmt := p.session.Query(query)
	return &PreparedStatement{stmt: stmt}, nil
}

// HealthCheck performs a health check on the database
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("cassandra provider not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test ping
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("cassandra health check failed: %w", err)
	}

	// Test basic query
	var result int
	if err := p.session.Query("SELECT now() FROM system.local").Scan(&result); err != nil {
		return fmt.Errorf("cassandra query health check failed: %w", err)
	}

	return nil
}

// GetStats returns database statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.DatabaseStats, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("cassandra provider not configured")
	}

	// Get cluster information
	var clusterName string
	if err := p.session.Query("SELECT cluster_name FROM system.local").Scan(&clusterName); err != nil {
		return nil, fmt.Errorf("failed to get cluster name: %w", err)
	}

	return &gateway.DatabaseStats{
		ActiveConnections: p.cluster.NumConns,
		IdleConnections:   0,
		MaxConnections:    p.cluster.NumConns,
		WaitCount:         0,
		WaitDuration:      0,
		MaxIdleClosed:     0,
		MaxIdleTimeClosed: 0,
		MaxLifetimeClosed: 0,
		ProviderData: map[string]interface{}{
			"driver":       "cassandra",
			"cluster_name": clusterName,
			"consistency":  p.cluster.Consistency,
		},
	}, nil
}

// Close closes the database connection
func (p *Provider) Close() error {
	if p.session != nil {
		p.session.Close()
	}
	return nil
}

// Transaction represents a Cassandra transaction
type Transaction struct {
	session *gocql.Session
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	// Cassandra transactions are handled by the session
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	// Cassandra transactions are handled by the session
	return nil
}

// Query executes a query within the transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	iter := t.session.Query(query, args...).Iter()
	return &QueryResult{iter: iter}, nil
}

// QueryRow executes a query that returns a single row within the transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	iter := t.session.Query(query, args...).Iter()
	if !iter.Scan() {
		return nil, fmt.Errorf("no rows found")
	}

	rowData := make(map[string]interface{})
	iter.MapScan(rowData)

	return &Row{data: rowData}, nil
}

// Exec executes a query without returning rows within the transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	err := t.session.Query(query, args...).Exec()
	if err != nil {
		return nil, err
	}
	return &ExecResult{affected: 1}, nil
}

// Prepare prepares a statement within the transaction
func (t *Transaction) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	stmt := t.session.Query(query)
	return &PreparedStatement{stmt: stmt}, nil
}

// QueryResult represents a Cassandra query result
type QueryResult struct {
	iter *gocql.Iter
}

// Close closes the result set
func (qr *QueryResult) Close() error {
	return qr.iter.Close()
}

// Next advances to the next row
func (qr *QueryResult) Next() bool {
	return qr.iter.Scan()
}

// Scan scans the current row into dest
func (qr *QueryResult) Scan(dest ...interface{}) error {
	if !qr.iter.Scan(dest...) {
		return qr.iter.Close()
	}
	return nil
}

// Columns returns the column names
func (qr *QueryResult) Columns() ([]string, error) {
	columns := qr.iter.Columns()
	columnNames := make([]string, len(columns))
	for i, col := range columns {
		columnNames[i] = col.Name
	}
	return columnNames, nil
}

// Err returns any error that occurred during iteration
func (qr *QueryResult) Err() error {
	return qr.iter.Close()
}

// Row represents a Cassandra row
type Row struct {
	data map[string]interface{}
}

// Scan scans the row into dest
func (r *Row) Scan(dest ...interface{}) error {
	if len(dest) > 0 {
		if destMap, ok := dest[0].(*map[string]interface{}); ok {
			*destMap = r.data
		}
	}
	return nil
}

// Err returns any error that occurred during scanning
func (r *Row) Err() error {
	return nil
}

// ExecResult represents a Cassandra execution result
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

// PreparedStatement represents a Cassandra prepared statement
type PreparedStatement struct {
	stmt *gocql.Query
}

// Close closes the prepared statement
func (ps *PreparedStatement) Close() error {
	return nil
}

// Query executes the prepared statement with args
func (ps *PreparedStatement) Query(ctx context.Context, args ...interface{}) (gateway.QueryResult, error) {
	iter := ps.stmt.Bind(args...).Iter()
	return &QueryResult{iter: iter}, nil
}

// QueryRow executes the prepared statement with args and returns a single row
func (ps *PreparedStatement) QueryRow(ctx context.Context, args ...interface{}) (gateway.Row, error) {
	iter := ps.stmt.Bind(args...).Iter()
	if !iter.Scan() {
		return nil, fmt.Errorf("no rows found")
	}

	rowData := make(map[string]interface{})
	iter.MapScan(rowData)

	return &Row{data: rowData}, nil
}

// Exec executes the prepared statement with args
func (ps *PreparedStatement) Exec(ctx context.Context, args ...interface{}) (gateway.ExecResult, error) {
	err := ps.stmt.Bind(args...).Exec()
	if err != nil {
		return nil, err
	}
	return &ExecResult{affected: 1}, nil
}
