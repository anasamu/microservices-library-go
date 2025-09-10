package cockroachdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/database/gateway"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Provider implements the DatabaseProvider interface for CockroachDB
type Provider struct {
	db     *sql.DB
	config map[string]interface{}
	logger *logrus.Logger
}

// NewProvider creates a new CockroachDB provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		logger: logger,
		config: make(map[string]interface{}),
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "cockroachdb"
}

// GetSupportedFeatures returns the supported database features
func (p *Provider) GetSupportedFeatures() []gateway.DatabaseFeature {
	return []gateway.DatabaseFeature{
		gateway.FeatureTransactions,
		gateway.FeaturePreparedStmts,
		gateway.FeatureConnectionPool,
		gateway.FeatureClustering,
		gateway.FeaturePersistent,
	}
}

// Configure configures the CockroachDB provider
func (p *Provider) Configure(config map[string]interface{}) error {
	p.config = config
	return nil
}

// IsConfigured returns true if the provider is configured
func (p *Provider) IsConfigured() bool {
	return len(p.config) > 0
}

// Connect establishes a connection to CockroachDB
func (p *Provider) Connect(ctx context.Context) error {
	// Build connection string
	connStr := p.buildConnectionString()

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open CockroachDB connection: %w", err)
	}

	// Configure connection pool
	if maxConns, ok := p.config["max_connections"].(int); ok {
		db.SetMaxOpenConns(maxConns)
	}
	if minConns, ok := p.config["min_connections"].(int); ok {
		db.SetMaxIdleConns(minConns)
	}
	if maxLifetime, ok := p.config["max_connection_lifetime"].(time.Duration); ok {
		db.SetConnMaxLifetime(maxLifetime)
	}
	if maxIdleTime, ok := p.config["max_connection_idle_time"].(time.Duration); ok {
		db.SetConnMaxIdleTime(maxIdleTime)
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping CockroachDB: %w", err)
	}

	p.db = db
	p.logger.Info("Connected to CockroachDB successfully")
	return nil
}

// Disconnect closes the connection to CockroachDB
func (p *Provider) Disconnect(ctx context.Context) error {
	if p.db != nil {
		if err := p.db.Close(); err != nil {
			return fmt.Errorf("failed to close CockroachDB connection: %w", err)
		}
		p.db = nil
		p.logger.Info("Disconnected from CockroachDB")
	}
	return nil
}

// Close closes the connection to CockroachDB
func (p *Provider) Close() error {
	return p.Disconnect(context.Background())
}

// IsConnected returns true if the provider is connected
func (p *Provider) IsConnected() bool {
	return p.db != nil
}

// Ping tests the connection to CockroachDB
func (p *Provider) Ping(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("not connected to CockroachDB")
	}
	return p.db.PingContext(ctx)
}

// Exec executes a query without returning rows
func (p *Provider) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected to CockroachDB")
	}

	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &Result{result: result}, nil
}

// Query executes a query that returns rows
func (p *Provider) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected to CockroachDB")
	}

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &Rows{rows: rows}, nil
}

// QueryRow executes a query that returns a single row
func (p *Provider) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected to CockroachDB")
	}

	row := p.db.QueryRowContext(ctx, query, args...)
	return &Row{row: row}, nil
}

// BeginTransaction starts a new transaction
func (p *Provider) BeginTransaction(ctx context.Context) (gateway.Transaction, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected to CockroachDB")
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{tx: tx, logger: p.logger}, nil
}

// Prepare prepares a statement
func (p *Provider) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected to CockroachDB")
	}

	stmt, err := p.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	return &PreparedStatement{stmt: stmt}, nil
}

// WithTransaction executes a function within a transaction
func (p *Provider) WithTransaction(ctx context.Context, fn func(gateway.Transaction) error) error {
	tx, err := p.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rollbackErr)
		}
		return err
	}

	return tx.Commit()
}

// GetStats returns database statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.DatabaseStats, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected to CockroachDB")
	}

	stats := p.db.Stats()
	return &gateway.DatabaseStats{
		ActiveConnections: stats.OpenConnections,
		IdleConnections:   stats.Idle,
		MaxConnections:    stats.MaxOpenConnections,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
	}, nil
}

// Health checks the health of the CockroachDB connection
func (p *Provider) Health(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("not connected to CockroachDB")
	}

	// Simple health check query
	var result int
	err := p.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// HealthCheck performs a health check on the database connection
func (p *Provider) HealthCheck(ctx context.Context) error {
	return p.Health(ctx)
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *gateway.ConnectionInfo {
	return &gateway.ConnectionInfo{
		Host:     getString(p.config, "host", "localhost"),
		Port:     getInt(p.config, "port", 26257),
		Database: getString(p.config, "database", "defaultdb"),
		User:     getString(p.config, "user", "root"),
	}
}

// buildConnectionString builds the PostgreSQL connection string for CockroachDB
func (p *Provider) buildConnectionString() string {
	host := getString(p.config, "host", "localhost")
	port := getInt(p.config, "port", 26257)
	user := getString(p.config, "user", "root")
	password := getString(p.config, "password", "")
	database := getString(p.config, "database", "defaultdb")
	sslMode := getString(p.config, "ssl_mode", "require")

	// CockroachDB specific parameters
	applicationName := getString(p.config, "application_name", "microservices-library")
	connectTimeout := getInt(p.config, "connect_timeout", 10)

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s application_name=%s connect_timeout=%d",
		host, port, user, password, database, sslMode, applicationName, connectTimeout,
	)

	// Add additional parameters if specified
	if cluster, ok := p.config["cluster"].(string); ok && cluster != "" {
		connStr += fmt.Sprintf(" options='--cluster=%s'", cluster)
	}

	return connStr
}

// Result implements the gateway.Result interface
type Result struct {
	result sql.Result
}

// LastInsertId returns the last inserted ID
func (r *Result) LastInsertId() (int64, error) {
	return r.result.LastInsertId()
}

// RowsAffected returns the number of affected rows
func (r *Result) RowsAffected() (int64, error) {
	return r.result.RowsAffected()
}

// Rows implements the gateway.Rows interface
type Rows struct {
	rows *sql.Rows
}

// Close closes the rows
func (r *Rows) Close() error {
	return r.rows.Close()
}

// Next advances to the next row
func (r *Rows) Next() bool {
	return r.rows.Next()
}

// Scan scans the current row into dest
func (r *Rows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

// Err returns any error that occurred during iteration
func (r *Rows) Err() error {
	return r.rows.Err()
}

// Columns returns the column names
func (r *Rows) Columns() ([]string, error) {
	return r.rows.Columns()
}

// Row implements the gateway.Row interface
type Row struct {
	row *sql.Row
}

// Scan scans the row into dest
func (r *Row) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

// Err returns any error that occurred during scanning
func (r *Row) Err() error {
	return r.row.Err()
}

// Transaction implements the gateway.Transaction interface
type Transaction struct {
	tx     *sql.Tx
	logger *logrus.Logger
}

// Exec executes a query within the transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	result, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query in transaction: %w", err)
	}
	return &Result{result: result}, nil
}

// Query executes a query that returns rows within the transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query in transaction: %w", err)
	}
	return &Rows{rows: rows}, nil
}

// QueryRow executes a query that returns a single row within the transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	row := t.tx.QueryRowContext(ctx, query, args...)
	return &Row{row: row}, nil
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	if err := t.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	t.logger.Debug("Transaction committed successfully")
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	if err := t.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	t.logger.Debug("Transaction rolled back successfully")
	return nil
}

// Prepare prepares a statement within the transaction
func (t *Transaction) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	stmt, err := t.tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement in transaction: %w", err)
	}
	return &PreparedStatement{stmt: stmt}, nil
}

// PreparedStatement implements the gateway.PreparedStatement interface
type PreparedStatement struct {
	stmt *sql.Stmt
}

// Exec executes the prepared statement
func (ps *PreparedStatement) Exec(ctx context.Context, args ...interface{}) (gateway.ExecResult, error) {
	result, err := ps.stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute prepared statement: %w", err)
	}
	return &Result{result: result}, nil
}

// Query executes the prepared statement and returns rows
func (ps *PreparedStatement) Query(ctx context.Context, args ...interface{}) (gateway.QueryResult, error) {
	rows, err := ps.stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query prepared statement: %w", err)
	}
	return &Rows{rows: rows}, nil
}

// QueryRow executes the prepared statement and returns a single row
func (ps *PreparedStatement) QueryRow(ctx context.Context, args ...interface{}) (gateway.Row, error) {
	row := ps.stmt.QueryRowContext(ctx, args...)
	return &Row{row: row}, nil
}

// Close closes the prepared statement
func (ps *PreparedStatement) Close() error {
	return ps.stmt.Close()
}

// Helper functions
func getString(config map[string]interface{}, key, defaultValue string) string {
	if value, ok := config[key].(string); ok {
		return value
	}
	return defaultValue
}

func getInt(config map[string]interface{}, key string, defaultValue int) int {
	if value, ok := config[key].(int); ok {
		return value
	}
	return defaultValue
}
