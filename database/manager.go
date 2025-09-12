package database

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// DatabaseManager manages multiple database providers
type DatabaseManager struct {
	providers map[string]DatabaseProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds database manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	MaxConnections  int               `json:"max_connections"`
	Metadata        map[string]string `json:"metadata"`
}

// DatabaseProvider interface for database backends
type DatabaseProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []DatabaseFeature
	GetConnectionInfo() *ConnectionInfo

	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool

	// Transaction management
	BeginTransaction(ctx context.Context) (Transaction, error)
	WithTransaction(ctx context.Context, fn func(Transaction) error) error

	// Query operations
	Query(ctx context.Context, query string, args ...interface{}) (QueryResult, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (Row, error)
	Exec(ctx context.Context, query string, args ...interface{}) (ExecResult, error)

	// Prepared statements
	Prepare(ctx context.Context, query string) (PreparedStatement, error)

	// Health and monitoring
	HealthCheck(ctx context.Context) error
	GetStats(ctx context.Context) (*DatabaseStats, error)

	// Configuration
	Configure(config map[string]interface{}) error
	IsConfigured() bool
	Close() error
}

// DatabaseFeature represents a database feature
type DatabaseFeature string

const (
	FeatureTransactions   DatabaseFeature = "transactions"
	FeaturePreparedStmts  DatabaseFeature = "prepared_statements"
	FeatureConnectionPool DatabaseFeature = "connection_pool"
	FeatureReadReplicas   DatabaseFeature = "read_replicas"
	FeatureClustering     DatabaseFeature = "clustering"
	FeatureSharding       DatabaseFeature = "sharding"
	FeatureFullTextSearch DatabaseFeature = "full_text_search"
	FeatureJSONSupport    DatabaseFeature = "json_support"
	FeatureGeoSpatial     DatabaseFeature = "geo_spatial"
	FeatureTimeSeries     DatabaseFeature = "time_series"
	FeatureGraphDB        DatabaseFeature = "graph_db"
	FeatureKeyValue       DatabaseFeature = "key_value"
	FeatureDocumentStore  DatabaseFeature = "document_store"
	FeatureColumnFamily   DatabaseFeature = "column_family"
	FeatureInMemory       DatabaseFeature = "in_memory"
	FeaturePersistent     DatabaseFeature = "persistent"
)

// ConnectionInfo represents database connection information
type ConnectionInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Driver   string `json:"driver"`
	Version  string `json:"version"`
}

// Transaction represents a database transaction
type Transaction interface {
	Commit() error
	Rollback() error
	Query(ctx context.Context, query string, args ...interface{}) (QueryResult, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (Row, error)
	Exec(ctx context.Context, query string, args ...interface{}) (ExecResult, error)
	Prepare(ctx context.Context, query string) (PreparedStatement, error)
}

// QueryResult represents the result of a query
type QueryResult interface {
	Close() error
	Next() bool
	Scan(dest ...interface{}) error
	Columns() ([]string, error)
	Err() error
}

// Row represents a single row from a query
type Row interface {
	Scan(dest ...interface{}) error
	Err() error
}

// ExecResult represents the result of an execution
type ExecResult interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// PreparedStatement represents a prepared statement
type PreparedStatement interface {
	Close() error
	Query(ctx context.Context, args ...interface{}) (QueryResult, error)
	QueryRow(ctx context.Context, args ...interface{}) (Row, error)
	Exec(ctx context.Context, args ...interface{}) (ExecResult, error)
}

// DatabaseStats represents database statistics
type DatabaseStats struct {
	ActiveConnections int                    `json:"active_connections"`
	IdleConnections   int                    `json:"idle_connections"`
	MaxConnections    int                    `json:"max_connections"`
	WaitCount         int64                  `json:"wait_count"`
	WaitDuration      time.Duration          `json:"wait_duration"`
	MaxIdleClosed     int64                  `json:"max_idle_closed"`
	MaxIdleTimeClosed int64                  `json:"max_idle_time_closed"`
	MaxLifetimeClosed int64                  `json:"max_lifetime_closed"`
	ProviderData      map[string]interface{} `json:"provider_data"`
}

// DefaultManagerConfig returns default database manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		DefaultProvider: "postgresql",
		RetryAttempts:   3,
		RetryDelay:      5 * time.Second,
		Timeout:         30 * time.Second,
		MaxConnections:  100,
		Metadata:        make(map[string]string),
	}
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(config *ManagerConfig, logger *logrus.Logger) *DatabaseManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &DatabaseManager{
		providers: make(map[string]DatabaseProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a database provider
func (dm *DatabaseManager) RegisterProvider(provider DatabaseProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	dm.providers[name] = provider
	dm.logger.WithField("provider", name).Info("Database provider registered")

	return nil
}

// GetProvider returns a database provider by name
func (dm *DatabaseManager) GetProvider(name string) (DatabaseProvider, error) {
	provider, exists := dm.providers[name]
	if !exists {
		return nil, fmt.Errorf("database provider not found: %s", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default database provider
func (dm *DatabaseManager) GetDefaultProvider() (DatabaseProvider, error) {
	return dm.GetProvider(dm.config.DefaultProvider)
}

// Connect connects to a database using the specified provider
func (dm *DatabaseManager) Connect(ctx context.Context, providerName string) error {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return err
	}

	// Connect with retry logic
	for attempt := 1; attempt <= dm.config.RetryAttempts; attempt++ {
		err = provider.Connect(ctx)
		if err == nil {
			break
		}

		dm.logger.WithError(err).WithFields(logrus.Fields{
			"provider": providerName,
			"attempt":  attempt,
		}).Warn("Database connection failed, retrying")

		if attempt < dm.config.RetryAttempts {
			time.Sleep(dm.config.RetryDelay)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database after %d attempts: %w", dm.config.RetryAttempts, err)
	}

	dm.logger.WithField("provider", providerName).Info("Database connected successfully")
	return nil
}

// Disconnect disconnects from a database using the specified provider
func (dm *DatabaseManager) Disconnect(ctx context.Context, providerName string) error {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("failed to disconnect from database: %w", err)
	}

	dm.logger.WithField("provider", providerName).Info("Database disconnected successfully")
	return nil
}

// Ping pings a database using the specified provider
func (dm *DatabaseManager) Ping(ctx context.Context, providerName string) error {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// Query executes a query using the specified provider
func (dm *DatabaseManager) Query(ctx context.Context, providerName, query string, args ...interface{}) (QueryResult, error) {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	result, err := provider.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	dm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"query":    query,
	}).Debug("Query executed successfully")

	return result, nil
}

// QueryRow executes a query that returns a single row using the specified provider
func (dm *DatabaseManager) QueryRow(ctx context.Context, providerName, query string, args ...interface{}) (Row, error) {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	row, err := provider.QueryRow(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query row: %w", err)
	}

	dm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"query":    query,
	}).Debug("Query row executed successfully")

	return row, nil
}

// Exec executes a query without returning rows using the specified provider
func (dm *DatabaseManager) Exec(ctx context.Context, providerName, query string, args ...interface{}) (ExecResult, error) {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	result, err := provider.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	dm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"query":    query,
	}).Debug("Query executed successfully")

	return result, nil
}

// BeginTransaction begins a transaction using the specified provider
func (dm *DatabaseManager) BeginTransaction(ctx context.Context, providerName string) (Transaction, error) {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	tx, err := provider.BeginTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	dm.logger.WithField("provider", providerName).Debug("Transaction started")
	return tx, nil
}

// WithTransaction executes a function within a transaction using the specified provider
func (dm *DatabaseManager) WithTransaction(ctx context.Context, providerName string, fn func(Transaction) error) error {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.WithTransaction(ctx, fn)
	if err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	dm.logger.WithField("provider", providerName).Debug("Transaction completed")
	return nil
}

// Prepare prepares a statement using the specified provider
func (dm *DatabaseManager) Prepare(ctx context.Context, providerName, query string) (PreparedStatement, error) {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	stmt, err := provider.Prepare(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	dm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"query":    query,
	}).Debug("Statement prepared successfully")

	return stmt, nil
}

// HealthCheck performs health check on all providers
func (dm *DatabaseManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for name, provider := range dm.providers {
		err := provider.HealthCheck(ctx)
		results[name] = err
	}

	return results
}

// GetStats gets statistics from a provider
func (dm *DatabaseManager) GetStats(ctx context.Context, providerName string) (*DatabaseStats, error) {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	stats, err := provider.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}

	return stats, nil
}

// GetSupportedProviders returns a list of registered providers
func (dm *DatabaseManager) GetSupportedProviders() []string {
	providers := make([]string, 0, len(dm.providers))
	for name := range dm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderCapabilities returns capabilities of a provider
func (dm *DatabaseManager) GetProviderCapabilities(providerName string) ([]DatabaseFeature, *ConnectionInfo, error) {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return nil, nil, err
	}

	return provider.GetSupportedFeatures(), provider.GetConnectionInfo(), nil
}

// Close closes all database connections
func (dm *DatabaseManager) Close() error {
	var lastErr error

	for name, provider := range dm.providers {
		if err := provider.Close(); err != nil {
			dm.logger.WithError(err).WithField("provider", name).Error("Failed to close database provider")
			lastErr = err
		}
	}

	return lastErr
}

// IsProviderConnected checks if a provider is connected
func (dm *DatabaseManager) IsProviderConnected(providerName string) bool {
	provider, err := dm.GetProvider(providerName)
	if err != nil {
		return false
	}
	return provider.IsConnected()
}

// GetConnectedProviders returns a list of connected providers
func (dm *DatabaseManager) GetConnectedProviders() []string {
	connected := make([]string, 0)
	for name, provider := range dm.providers {
		if provider.IsConnected() {
			connected = append(connected, name)
		}
	}
	return connected
}
