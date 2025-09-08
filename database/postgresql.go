package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// PostgreSQL represents PostgreSQL database connection
type PostgreSQL struct {
	DB     *sqlx.DB
	config *PostgreSQLConfig
	logger *logrus.Logger
}

// PostgreSQLConfig holds PostgreSQL configuration
type PostgreSQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	MinConns int
}

// NewPostgreSQL creates new PostgreSQL connection
func NewPostgreSQL(config *PostgreSQLConfig, logger *logrus.Logger) (*PostgreSQL, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxConns)
	db.SetMaxIdleConns(config.MinConns)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	pg := &PostgreSQL{
		DB:     db,
		config: config,
		logger: logger,
	}

	pg.logger.Info("PostgreSQL connection established successfully")
	return pg, nil
}

// Close closes the database connection
func (pg *PostgreSQL) Close() error {
	if pg.DB != nil {
		pg.logger.Info("Closing PostgreSQL connection")
		return pg.DB.Close()
	}
	return nil
}

// Ping checks database connection
func (pg *PostgreSQL) Ping(ctx context.Context) error {
	return pg.DB.PingContext(ctx)
}

// GetDB returns the underlying sqlx.DB instance
func (pg *PostgreSQL) GetDB() *sqlx.DB {
	return pg.DB
}

// BeginTx starts a new transaction
func (pg *PostgreSQL) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return pg.DB.BeginTxx(ctx, nil)
}

// Exec executes a query without returning any rows
func (pg *PostgreSQL) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return pg.DB.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (pg *PostgreSQL) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return pg.DB.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (pg *PostgreSQL) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return pg.DB.QueryRowContext(ctx, query, args...)
}

// Get executes a query and scans the result into dest
func (pg *PostgreSQL) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pg.DB.GetContext(ctx, dest, query, args...)
}

// Select executes a query and scans the results into dest
func (pg *PostgreSQL) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pg.DB.SelectContext(ctx, dest, query, args...)
}

// NamedExec executes a named query
func (pg *PostgreSQL) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return pg.DB.NamedExecContext(ctx, query, arg)
}

// NamedQuery executes a named query
func (pg *PostgreSQL) NamedQuery(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	return pg.DB.NamedQueryContext(ctx, query, arg)
}

// Stats returns database connection statistics
func (pg *PostgreSQL) Stats() sql.DBStats {
	return pg.DB.Stats()
}

// HealthCheck performs a health check on the database
func (pg *PostgreSQL) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pg.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check if we can execute a simple query
	var result int
	if err := pg.QueryRow(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query health check failed: %w", err)
	}

	return nil
}

// WithRetry executes a function with retry logic
func (pg *PostgreSQL) WithRetry(ctx context.Context, fn func() error, maxRetries int) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if err := fn(); err != nil {
			lastErr = err
			if i < maxRetries {
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

// Transaction executes a function within a database transaction
func (pg *PostgreSQL) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := pg.BeginTx(ctx)
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

	err = fn(tx)
	return err
}
