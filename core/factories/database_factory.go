package factories

import (
	"context"
	"fmt"

	"github.com/anasamu/microservices-library-go/core/interfaces"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMongoDB    DatabaseType = "mongodb"
	DatabaseTypeRedis      DatabaseType = "redis"
)

// DatabaseFactory creates database instances
type DatabaseFactory struct {
	configs map[DatabaseType]interface{}
}

// NewDatabaseFactory creates a new database factory
func NewDatabaseFactory() *DatabaseFactory {
	return &DatabaseFactory{
		configs: make(map[DatabaseType]interface{}),
	}
}

// RegisterConfig registers a database configuration
func (f *DatabaseFactory) RegisterConfig(dbType DatabaseType, config interface{}) {
	f.configs[dbType] = config
}

// CreateDatabase creates a database instance based on type
func (f *DatabaseFactory) CreateDatabase(ctx context.Context, dbType DatabaseType) (interfaces.Database, error) {
	config, exists := f.configs[dbType]
	if !exists {
		return nil, fmt.Errorf("no configuration found for database type: %s", dbType)
	}

	// This is a placeholder implementation
	// In a real implementation, you would create actual database instances
	// based on the configuration and return them as interfaces.Database

	switch dbType {
	case DatabaseTypePostgreSQL:
		// Create PostgreSQL instance
		return nil, fmt.Errorf("PostgreSQL implementation not available in core package")

	case DatabaseTypeMongoDB:
		// Create MongoDB instance
		return nil, fmt.Errorf("MongoDB implementation not available in core package")

	case DatabaseTypeRedis:
		// Create Redis instance
		return nil, fmt.Errorf("Redis implementation not available in core package")

	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// CreateMultipleDatabases creates multiple database instances
func (f *DatabaseFactory) CreateMultipleDatabases(ctx context.Context, dbTypes []DatabaseType) (map[DatabaseType]interfaces.Database, error) {
	databases := make(map[DatabaseType]interfaces.Database)

	for _, dbType := range dbTypes {
		db, err := f.CreateDatabase(ctx, dbType)
		if err != nil {
			// Close already created databases
			for _, createdDB := range databases {
				createdDB.Close()
			}
			return nil, fmt.Errorf("failed to create database %s: %w", dbType, err)
		}
		databases[dbType] = db
	}

	return databases, nil
}

// GetSupportedTypes returns supported database types
func (f *DatabaseFactory) GetSupportedTypes() []DatabaseType {
	return []DatabaseType{
		DatabaseTypePostgreSQL,
		DatabaseTypeMongoDB,
		DatabaseTypeRedis,
	}
}

// IsTypeSupported checks if a database type is supported
func (f *DatabaseFactory) IsTypeSupported(dbType DatabaseType) bool {
	supportedTypes := f.GetSupportedTypes()
	for _, supportedType := range supportedTypes {
		if supportedType == dbType {
			return true
		}
	}
	return false
}
