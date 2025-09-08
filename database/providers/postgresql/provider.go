package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/database/gateway"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Provider implements DatabaseProvider for PostgreSQL
type Provider struct {
	db     *sqlx.DB
	config map[string]interface{}
	logger *logrus.Logger
}

// NewProvider creates a new PostgreSQL database provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "postgresql"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.DatabaseFeature {
	return []gateway.DatabaseFeature{
		gateway.FeatureTransactions,
		gateway.FeaturePreparedStmts,
		gateway.FeatureConnectionPool,
		gateway.FeatureReadReplicas,
		gateway.FeatureFullTextSearch,
		gateway.FeatureJSONSupport,
		gateway.FeatureGeoSpatial,
		gateway.FeaturePersistent,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *gateway.ConnectionInfo {
	host, _ := p.config["host"].(string)
	port, _ := p.config["port"].(int)
	database, _ := p.config["database"].(string)
	user, _ := p.config["user"].(string)

	return &gateway.ConnectionInfo{
		Host:     host,
		Port:     port,
		Database: database,
		User:     user,
		Driver:   "postgres",
		Version:  "13+",
	}
}

// Configure configures the PostgreSQL provider
func (p *Provider) Configure(config map[string]interface{}) error {
	host, ok := config["host"].(string)
	if !ok || host == "" {
		host = "localhost"
	}

	port, ok := config["port"].(int)
	if !ok || port == 0 {
		port = 5432
	}

	user, ok := config["user"].(string)
	if !ok || user == "" {
		return fmt.Errorf("postgresql user is required")
	}

	password, ok := config["password"].(string)
	if !ok || password == "" {
		return fmt.Errorf("postgresql password is required")
	}

	database, ok := config["database"].(string)
	if !ok || database == "" {
		return fmt.Errorf("postgresql database is required")
	}

	sslMode, ok := config["ssl_mode"].(string)
	if !ok || sslMode == "" {
		sslMode = "disable"
	}

	maxConns, ok := config["max_connections"].(int)
	if !ok || maxConns == 0 {
		maxConns = 100
	}

	minConns, ok := config["min_connections"].(int)
	if !ok || minConns == 0 {
		minConns = 10
	}

	// Build DSN
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, password, host, port, database, sslMode)

	// Create database connection
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(minConns)
	db.SetConnMaxLifetime(time.Hour)

	p.db = db
	p.config = config

	p.logger.Info("PostgreSQL provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.db != nil
}

// Connect connects to the database
func (p *Provider) Connect(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("postgresql provider not configured")
	}

	// Test connection
	if err := p.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	p.logger.Info("PostgreSQL connected successfully")
	return nil
}

// Disconnect disconnects from the database
func (p *Provider) Disconnect(ctx context.Context) error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping checks database connection
func (p *Provider) Ping(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("postgresql provider not configured")
	}
	return p.db.PingContext(ctx)
}

// IsConnected checks if the database is connected
func (p *Provider) IsConnected() bool {
	if !p.IsConfigured() {
		return false
	}
	return p.db.Ping() == nil
}

// BeginTransaction begins a new transaction
func (p *Provider) BeginTransaction(ctx context.Context) (gateway.Transaction, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("postgresql provider not configured")
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{tx: tx}, nil
}

// WithTransaction executes a function within a transaction
func (p *Provider) WithTransaction(ctx context.Context, fn func(gateway.Transaction) error) error {
	if !p.IsConfigured() {
		return fmt.Errorf("postgresql provider not configured")
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(&Transaction{tx: tx})
	return err
}

// Query executes a query that returns rows
func (p *Provider) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("postgresql provider not configured")
	}

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &QueryResult{rows: rows}, nil
}

// QueryRow executes a query that returns a single row
func (p *Provider) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("postgresql provider not configured")
	}

	row := p.db.QueryRowContext(ctx, query, args...)
	return &Row{row: row}, nil
}

// Exec executes a query without returning rows
func (p *Provider) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("postgresql provider not configured")
	}

	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &ExecResult{result: result}, nil
}

// Prepare prepares a statement
func (p *Provider) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("postgresql provider not configured")
	}

	stmt, err := p.db.PreparexContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	return &PreparedStatement{stmt: stmt}, nil
}

// HealthCheck performs a health check on the database
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("postgresql provider not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test ping
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("postgresql health check failed: %w", err)
	}

	// Test basic query
	var result int
	if err := p.db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("postgresql query health check failed: %w", err)
	}

	return nil
}

// GetStats returns database statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.DatabaseStats, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("postgresql provider not configured")
	}

	stats := p.db.Stats()

	return &gateway.DatabaseStats{
		ActiveConnections: stats.OpenConnections,
		IdleConnections:   stats.Idle,
		MaxConnections:    stats.MaxOpenConnections,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxIdleTimeClosed: stats.MaxIdleTimeClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
		ProviderData: map[string]interface{}{
			"driver": "postgres",
		},
	}, nil
}

// Close closes the database connection
func (p *Provider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Transaction represents a PostgreSQL transaction
type Transaction struct {
	tx *sqlx.Tx
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

// Query executes a query within the transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &QueryResult{rows: rows}, nil
}

// QueryRow executes a query that returns a single row within the transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	row := t.tx.QueryRowContext(ctx, query, args...)
	return &Row{row: row}, nil
}

// Exec executes a query without returning rows within the transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	result, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &ExecResult{result: result}, nil
}

// Prepare prepares a statement within the transaction
func (t *Transaction) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	stmt, err := t.tx.PreparexContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &PreparedStatement{stmt: stmt}, nil
}

// QueryResult represents a PostgreSQL query result
type QueryResult struct {
	rows *sql.Rows
}

// Close closes the result set
func (qr *QueryResult) Close() error {
	return qr.rows.Close()
}

// Next advances to the next row
func (qr *QueryResult) Next() bool {
	return qr.rows.Next()
}

// Scan scans the current row into dest
func (qr *QueryResult) Scan(dest ...interface{}) error {
	return qr.rows.Scan(dest...)
}

// Columns returns the column names
func (qr *QueryResult) Columns() ([]string, error) {
	return qr.rows.Columns()
}

// Err returns any error that occurred during iteration
func (qr *QueryResult) Err() error {
	return qr.rows.Err()
}

// Row represents a PostgreSQL row
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

// ExecResult represents a PostgreSQL execution result
type ExecResult struct {
	result sql.Result
}

// LastInsertId returns the last insert ID
func (er *ExecResult) LastInsertId() (int64, error) {
	return er.result.LastInsertId()
}

// RowsAffected returns the number of rows affected
func (er *ExecResult) RowsAffected() (int64, error) {
	return er.result.RowsAffected()
}

// PreparedStatement represents a PostgreSQL prepared statement
type PreparedStatement struct {
	stmt *sqlx.Stmt
}

// Close closes the prepared statement
func (ps *PreparedStatement) Close() error {
	return ps.stmt.Close()
}

// Query executes the prepared statement with args
func (ps *PreparedStatement) Query(ctx context.Context, args ...interface{}) (gateway.QueryResult, error) {
	rows, err := ps.stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return &QueryResult{rows: rows}, nil
}

// QueryRow executes the prepared statement with args and returns a single row
func (ps *PreparedStatement) QueryRow(ctx context.Context, args ...interface{}) (gateway.Row, error) {
	row := ps.stmt.QueryRowContext(ctx, args...)
	return &Row{row: row}, nil
}

// Exec executes the prepared statement with args
func (ps *PreparedStatement) Exec(ctx context.Context, args ...interface{}) (gateway.ExecResult, error) {
	result, err := ps.stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return &ExecResult{result: result}, nil
}
