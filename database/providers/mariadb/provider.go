package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/database/gateway"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// Provider implements the DatabaseProvider interface for MariaDB
type Provider struct {
	db     *sqlx.DB
	config map[string]interface{}
	logger *logrus.Logger
}

// NewProvider creates a new MariaDB provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		logger: logger,
		config: make(map[string]interface{}),
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "mariadb"
}

// GetSupportedFeatures returns the supported database features
func (p *Provider) GetSupportedFeatures() []gateway.DatabaseFeature {
	return []gateway.DatabaseFeature{
		gateway.FeatureTransactions,
		gateway.FeaturePreparedStmts,
		gateway.FeatureConnectionPool,
		gateway.FeatureJSONSupport,
		gateway.FeaturePersistent,
	}
}

// Configure configures the MariaDB provider
func (p *Provider) Configure(config map[string]interface{}) error {
	p.config = config
	return nil
}

// IsConfigured returns true if the provider is configured
func (p *Provider) IsConfigured() bool {
	return len(p.config) > 0
}

// Connect establishes a connection to MariaDB
func (p *Provider) Connect(ctx context.Context) error {
	// Build connection string
	connStr := p.buildConnectionString()

	// Open database connection
	db, err := sqlx.ConnectContext(ctx, "mysql", connStr)
	if err != nil {
		return fmt.Errorf("failed to open MariaDB connection: %w", err)
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
		return fmt.Errorf("failed to ping MariaDB: %w", err)
	}

	p.db = db
	p.logger.Info("Connected to MariaDB successfully")
	return nil
}

// Disconnect closes the connection to MariaDB
func (p *Provider) Disconnect(ctx context.Context) error {
	if p.db != nil {
		if err := p.db.Close(); err != nil {
			return fmt.Errorf("failed to close MariaDB connection: %w", err)
		}
		p.db = nil
		p.logger.Info("Disconnected from MariaDB")
	}
	return nil
}

// Close closes the connection to MariaDB
func (p *Provider) Close() error {
	return p.Disconnect(context.Background())
}

// IsConnected returns true if the provider is connected
func (p *Provider) IsConnected() bool {
	return p.db != nil
}

// Ping tests the connection to MariaDB
func (p *Provider) Ping(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("not connected to MariaDB")
	}
	return p.db.PingContext(ctx)
}

// Exec executes a query without returning rows
func (p *Provider) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected to MariaDB")
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
		return nil, fmt.Errorf("not connected to MariaDB")
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
		return nil, fmt.Errorf("not connected to MariaDB")
	}

	row := p.db.QueryRowContext(ctx, query, args...)
	return &Row{row: row}, nil
}

// BeginTransaction starts a new transaction
func (p *Provider) BeginTransaction(ctx context.Context) (gateway.Transaction, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected to MariaDB")
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{tx: tx, logger: p.logger}, nil
}

// Prepare prepares a statement
func (p *Provider) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	if p.db == nil {
		return nil, fmt.Errorf("not connected to MariaDB")
	}

	stmt, err := p.db.PreparexContext(ctx, query)
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
		return nil, fmt.Errorf("not connected to MariaDB")
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

// Health checks the health of the MariaDB connection
func (p *Provider) Health(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("not connected to MariaDB")
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
		Port:     getInt(p.config, "port", 3306),
		Database: getString(p.config, "database", "testdb"),
		User:     getString(p.config, "user", "root"),
	}
}

// buildConnectionString builds the MySQL connection string for MariaDB
func (p *Provider) buildConnectionString() string {
	host := getString(p.config, "host", "localhost")
	port := getInt(p.config, "port", 3306)
	user := getString(p.config, "user", "root")
	password := getString(p.config, "password", "")
	database := getString(p.config, "database", "testdb")

	// MariaDB specific parameters
	charset := getString(p.config, "charset", "utf8mb4")
	collation := getString(p.config, "collation", "utf8mb4_unicode_ci")
	parseTime := getBool(p.config, "parse_time", true)
	loc := getString(p.config, "loc", "Local")

	// Build connection string
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=%t&loc=%s",
		user, password, host, port, database, charset, collation, parseTime, loc,
	)

	// Add additional parameters if specified
	if timeout, ok := p.config["timeout"].(int); ok && timeout > 0 {
		connStr += fmt.Sprintf("&timeout=%ds", timeout)
	}
	if readTimeout, ok := p.config["read_timeout"].(int); ok && readTimeout > 0 {
		connStr += fmt.Sprintf("&readTimeout=%ds", readTimeout)
	}
	if writeTimeout, ok := p.config["write_timeout"].(int); ok && writeTimeout > 0 {
		connStr += fmt.Sprintf("&writeTimeout=%ds", writeTimeout)
	}
	if sslMode, ok := p.config["ssl_mode"].(string); ok && sslMode != "" {
		connStr += fmt.Sprintf("&tls=%s", sslMode)
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
	tx     *sqlx.Tx
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
	stmt, err := t.tx.PreparexContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement in transaction: %w", err)
	}
	return &PreparedStatement{stmt: stmt}, nil
}

// PreparedStatement implements the gateway.PreparedStatement interface
type PreparedStatement struct {
	stmt *sqlx.Stmt
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

func getBool(config map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := config[key].(bool); ok {
		return value
	}
	return defaultValue
}
