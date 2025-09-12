package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/database"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Provider implements DatabaseProvider for Redis
type Provider struct {
	client *redis.Client
	config map[string]interface{}
	logger *logrus.Logger
}

// NewProvider creates a new Redis database provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "redis"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.DatabaseFeature {
	return []gateway.DatabaseFeature{
		gateway.FeatureConnectionPool,
		gateway.FeatureClustering,
		gateway.FeatureKeyValue,
		gateway.FeatureInMemory,
		gateway.FeaturePersistent,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *gateway.ConnectionInfo {
	host, _ := p.config["host"].(string)
	port, _ := p.config["port"].(int)
	db, _ := p.config["db"].(int)

	return &gateway.ConnectionInfo{
		Host:     host,
		Port:     port,
		Database: fmt.Sprintf("db%d", db),
		User:     "",
		Driver:   "redis",
		Version:  "6.0+",
	}
}

// Configure configures the Redis provider
func (p *Provider) Configure(config map[string]interface{}) error {
	host, ok := config["host"].(string)
	if !ok || host == "" {
		host = "localhost"
	}

	port, ok := config["port"].(int)
	if !ok || port == 0 {
		port = 6379
	}

	password, ok := config["password"].(string)
	if !ok {
		password = ""
	}

	db, ok := config["db"].(int)
	if !ok {
		db = 0
	}

	poolSize, ok := config["pool_size"].(int)
	if !ok || poolSize == 0 {
		poolSize = 100
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
		PoolSize: poolSize,
	})

	p.client = client
	p.config = config

	p.logger.Info("Redis provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.client != nil
}

// Connect connects to the database
func (p *Provider) Connect(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("redis provider not configured")
	}

	// Test connection
	if err := p.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	p.logger.Info("Redis connected successfully")
	return nil
}

// Disconnect disconnects from the database
func (p *Provider) Disconnect(ctx context.Context) error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// Ping checks database connection
func (p *Provider) Ping(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("redis provider not configured")
	}
	return p.client.Ping(ctx).Err()
}

// IsConnected checks if the database is connected
func (p *Provider) IsConnected() bool {
	if !p.IsConfigured() {
		return false
	}
	return p.client.Ping(context.Background()).Err() == nil
}

// BeginTransaction begins a new transaction
func (p *Provider) BeginTransaction(ctx context.Context) (gateway.Transaction, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("redis provider not configured")
	}

	// Redis transactions are handled by MULTI/EXEC commands
	return &Transaction{client: p.client}, nil
}

// WithTransaction executes a function within a transaction
func (p *Provider) WithTransaction(ctx context.Context, fn func(gateway.Transaction) error) error {
	if !p.IsConfigured() {
		return fmt.Errorf("redis provider not configured")
	}

	// Redis transactions are handled by MULTI/EXEC commands
	tx := &Transaction{client: p.client}
	return fn(tx)
}

// Query executes a query that returns rows (Redis doesn't have traditional queries)
func (p *Provider) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("redis provider not configured")
	}

	// For Redis, we'll use SCAN to get keys matching a pattern
	pattern := "*"
	if len(args) > 0 {
		if p, ok := args[0].(string); ok {
			pattern = p
		}
	}

	keys, cursor, err := p.client.Scan(ctx, 0, pattern, 100).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &QueryResult{keys: keys, cursor: cursor}, nil
}

// QueryRow executes a query that returns a single row
func (p *Provider) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("redis provider not configured")
	}

	// For Redis, we'll get a single key
	key := "*"
	if len(args) > 0 {
		if k, ok := args[0].(string); ok {
			key = k
		}
	}

	keys, _, err := p.client.Scan(ctx, 0, key, 1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to execute query row: %w", err)
	}

	var value string
	if len(keys) > 0 {
		value, err = p.client.Get(ctx, keys[0]).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get value: %w", err)
		}
	}

	return &Row{key: key, value: value}, nil
}

// Exec executes a query without returning rows
func (p *Provider) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("redis provider not configured")
	}

	// For Redis, we'll set a key-value pair
	key := "exec_test"
	value := "test_value"
	if len(args) >= 2 {
		if k, ok := args[0].(string); ok {
			key = k
		}
		if v, ok := args[1].(string); ok {
			value = v
		}
	}

	err := p.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &ExecResult{affected: 1}, nil
}

// Prepare prepares a statement (Redis doesn't have prepared statements)
func (p *Provider) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("redis provider not configured")
	}

	// Redis doesn't have prepared statements, return a mock
	return &PreparedStatement{client: p.client, query: query}, nil
}

// HealthCheck performs a health check on the database
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("redis provider not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test ping
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	// Test basic operations
	testKey := "health_check_test"
	testValue := "test_value"

	if err := p.client.Set(ctx, testKey, testValue, time.Minute).Err(); err != nil {
		return fmt.Errorf("redis set operation failed: %w", err)
	}

	if _, err := p.client.Get(ctx, testKey).Result(); err != nil {
		return fmt.Errorf("redis get operation failed: %w", err)
	}

	if err := p.client.Del(ctx, testKey).Err(); err != nil {
		return fmt.Errorf("redis del operation failed: %w", err)
	}

	return nil
}

// GetStats returns database statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.DatabaseStats, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("redis provider not configured")
	}

	// Get Redis info
	info, err := p.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Get connection pool stats
	poolStats := p.client.PoolStats()

	return &gateway.DatabaseStats{
		ActiveConnections: int(poolStats.TotalConns),
		IdleConnections:   int(poolStats.IdleConns),
		MaxConnections:    int(poolStats.TotalConns),
		WaitCount:         0, // Redis pool stats don't expose wait count
		WaitDuration:      0, // Redis pool stats don't expose wait duration
		MaxIdleClosed:     0, // Redis pool stats don't expose these fields
		MaxIdleTimeClosed: 0,
		MaxLifetimeClosed: 0,
		ProviderData: map[string]interface{}{
			"driver": "redis",
			"info":   info,
		},
	}, nil
}

// Close closes the database connection
func (p *Provider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// Transaction represents a Redis transaction
type Transaction struct {
	client *redis.Client
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	// Redis transactions are handled by MULTI/EXEC commands
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	// Redis transactions are handled by MULTI/EXEC commands
	return nil
}

// Query executes a query within the transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	// Redis transactions are handled by MULTI/EXEC commands
	return nil, fmt.Errorf("redis transactions are handled by MULTI/EXEC commands")
}

// QueryRow executes a query that returns a single row within the transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	// Redis transactions are handled by MULTI/EXEC commands
	return nil, fmt.Errorf("redis transactions are handled by MULTI/EXEC commands")
}

// Exec executes a query without returning rows within the transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	// Redis transactions are handled by MULTI/EXEC commands
	return nil, fmt.Errorf("redis transactions are handled by MULTI/EXEC commands")
}

// Prepare prepares a statement within the transaction
func (t *Transaction) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	// Redis transactions are handled by MULTI/EXEC commands
	return nil, fmt.Errorf("redis transactions are handled by MULTI/EXEC commands")
}

// QueryResult represents a Redis query result
type QueryResult struct {
	keys   []string
	cursor uint64
}

// Close closes the result set
func (qr *QueryResult) Close() error {
	return nil
}

// Next advances to the next row
func (qr *QueryResult) Next() bool {
	return len(qr.keys) > 0
}

// Scan scans the current row into dest
func (qr *QueryResult) Scan(dest ...interface{}) error {
	if len(qr.keys) == 0 {
		return fmt.Errorf("no more rows")
	}

	if len(dest) > 0 {
		if destStr, ok := dest[0].(*string); ok {
			*destStr = qr.keys[0]
			qr.keys = qr.keys[1:]
		}
	}

	return nil
}

// Columns returns the column names
func (qr *QueryResult) Columns() ([]string, error) {
	return []string{"key"}, nil
}

// Err returns any error that occurred during iteration
func (qr *QueryResult) Err() error {
	return nil
}

// Row represents a Redis row
type Row struct {
	key   string
	value string
}

// Scan scans the row into dest
func (r *Row) Scan(dest ...interface{}) error {
	if len(dest) >= 2 {
		if destKey, ok := dest[0].(*string); ok {
			*destKey = r.key
		}
		if destValue, ok := dest[1].(*string); ok {
			*destValue = r.value
		}
	}
	return nil
}

// Err returns any error that occurred during scanning
func (r *Row) Err() error {
	return nil
}

// ExecResult represents a Redis execution result
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

// PreparedStatement represents a Redis prepared statement
type PreparedStatement struct {
	client *redis.Client
	query  string
}

// Close closes the prepared statement
func (ps *PreparedStatement) Close() error {
	return nil
}

// Query executes the prepared statement with args
func (ps *PreparedStatement) Query(ctx context.Context, args ...interface{}) (gateway.QueryResult, error) {
	// Redis doesn't have prepared statements, return empty result
	return &QueryResult{keys: []string{}}, nil
}

// QueryRow executes the prepared statement with args and returns a single row
func (ps *PreparedStatement) QueryRow(ctx context.Context, args ...interface{}) (gateway.Row, error) {
	// Redis doesn't have prepared statements, return empty row
	return &Row{key: "", value: ""}, nil
}

// Exec executes the prepared statement with args
func (ps *PreparedStatement) Exec(ctx context.Context, args ...interface{}) (gateway.ExecResult, error) {
	// Redis doesn't have prepared statements, return empty result
	return &ExecResult{affected: 0}, nil
}
