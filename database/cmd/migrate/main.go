package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/anasamu/microservices-library-go/database/gateway"
	"github.com/anasamu/microservices-library-go/database/migrations"
	"github.com/anasamu/microservices-library-go/database/providers/cassandra"
	"github.com/anasamu/microservices-library-go/database/providers/cockroachdb"
	"github.com/anasamu/microservices-library-go/database/providers/influxdb"
	"github.com/anasamu/microservices-library-go/database/providers/mongodb"
	"github.com/anasamu/microservices-library-go/database/providers/mysql"
	"github.com/anasamu/microservices-library-go/database/providers/postgresql"
	"github.com/anasamu/microservices-library-go/database/providers/redis"
	"github.com/anasamu/microservices-library-go/database/providers/sqlite"
	"github.com/sirupsen/logrus"
)

var (
	provider      = flag.String("provider", "postgresql", "Database provider (postgresql, mysql, sqlite, cassandra, cockroachdb, influxdb, mongodb, redis)")
	migrationsDir = flag.String("dir", "./migrations", "Migrations directory")
	command       = flag.String("cmd", "status", "Command (create, up, down, status, reset, validate)")
	name          = flag.String("name", "", "Migration name (for create command)")
	config        = flag.String("config", "", "Configuration file path")
	verbose       = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()

	// Setup logger
	logger := logrus.New()
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	// Create database manager
	config := gateway.DefaultManagerConfig()
	databaseManager := gateway.NewDatabaseManager(config, logger)

	// Register and configure provider
	providerName := *provider
	dbProvider, err := createProvider(providerName, logger)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	if err := databaseManager.RegisterProvider(dbProvider); err != nil {
		log.Fatalf("Failed to register provider: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	if err := databaseManager.Connect(ctx, providerName); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer databaseManager.Close()

	// Get CLI manager
	provider, err := databaseManager.GetProvider(providerName)
	if err != nil {
		log.Fatalf("Failed to get provider: %v", err)
	}

	cliManager := migrations.NewCLIManager(provider, *migrationsDir, logger)

	// Execute command
	switch *command {
	case "create":
		if *name == "" {
			log.Fatal("Migration name is required for create command")
		}
		if err := cliManager.CreateMigration(*name); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
		fmt.Printf("Migration '%s' created successfully\n", *name)

	case "up":
		if err := cliManager.Up(ctx); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := cliManager.Down(ctx); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		fmt.Println("Migration rolled back successfully")

	case "status":
		if err := cliManager.Status(ctx); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}

	case "reset":
		if err := cliManager.Reset(ctx); err != nil {
			log.Fatalf("Failed to reset migrations: %v", err)
		}
		fmt.Println("Database reset successfully")

	case "validate":
		if err := cliManager.Validate(); err != nil {
			log.Fatalf("Migration validation failed: %v", err)
		}
		fmt.Println("All migrations are valid")

	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func createProvider(providerName string, logger *logrus.Logger) (gateway.DatabaseProvider, error) {
	switch providerName {
	case "postgresql":
		provider := postgresql.NewProvider(logger)
		config := map[string]interface{}{
			"host":     getEnv("POSTGRES_HOST", "localhost"),
			"port":     getEnvInt("POSTGRES_PORT", 5432),
			"user":     getEnv("POSTGRES_USER", "postgres"),
			"password": getEnv("POSTGRES_PASSWORD", "password"),
			"database": getEnv("POSTGRES_DATABASE", "testdb"),
			"ssl_mode": getEnv("POSTGRES_SSL_MODE", "disable"),
		}
		if err := provider.Configure(config); err != nil {
			return nil, err
		}
		return provider, nil

	case "mysql":
		provider := mysql.NewProvider(logger)
		config := map[string]interface{}{
			"host":     getEnv("MYSQL_HOST", "localhost"),
			"port":     getEnvInt("MYSQL_PORT", 3306),
			"user":     getEnv("MYSQL_USER", "root"),
			"password": getEnv("MYSQL_PASSWORD", "password"),
			"database": getEnv("MYSQL_DATABASE", "testdb"),
		}
		if err := provider.Configure(config); err != nil {
			return nil, err
		}
		return provider, nil

	case "sqlite":
		provider := sqlite.NewProvider(logger)
		config := map[string]interface{}{
			"file": getEnv("SQLITE_FILE", "./test.db"),
		}
		if err := provider.Configure(config); err != nil {
			return nil, err
		}
		return provider, nil

	case "cassandra":
		provider := cassandra.NewProvider(logger)
		config := map[string]interface{}{
			"hosts":       []string{getEnv("CASSANDRA_HOST", "localhost")},
			"keyspace":    getEnv("CASSANDRA_KEYSPACE", "test_keyspace"),
			"username":    getEnv("CASSANDRA_USERNAME", ""),
			"password":    getEnv("CASSANDRA_PASSWORD", ""),
			"consistency": getEnv("CASSANDRA_CONSISTENCY", "quorum"),
		}
		if err := provider.Configure(config); err != nil {
			return nil, err
		}
		return provider, nil

	case "cockroachdb":
		provider := cockroachdb.NewProvider(logger)
		config := map[string]interface{}{
			"host":     getEnv("COCKROACHDB_HOST", "localhost"),
			"port":     getEnvInt("COCKROACHDB_PORT", 26257),
			"user":     getEnv("COCKROACHDB_USER", "root"),
			"password": getEnv("COCKROACHDB_PASSWORD", ""),
			"database": getEnv("COCKROACHDB_DATABASE", "defaultdb"),
			"ssl_mode": getEnv("COCKROACHDB_SSL_MODE", "require"),
			"cluster":  getEnv("COCKROACHDB_CLUSTER", ""),
		}
		if err := provider.Configure(config); err != nil {
			return nil, err
		}
		return provider, nil

	case "influxdb":
		provider := influxdb.NewProvider(logger)
		config := map[string]interface{}{
			"url":    getEnv("INFLUXDB_URL", "http://localhost:8086"),
			"token":  getEnv("INFLUXDB_TOKEN", ""),
			"org":    getEnv("INFLUXDB_ORG", ""),
			"bucket": getEnv("INFLUXDB_BUCKET", ""),
		}
		if err := provider.Configure(config); err != nil {
			return nil, err
		}
		return provider, nil

	case "mongodb":
		provider := mongodb.NewProvider(logger)
		config := map[string]interface{}{
			"uri":      getEnv("MONGO_URI", "mongodb://localhost:27017"),
			"database": getEnv("MONGO_DATABASE", "testdb"),
		}
		if err := provider.Configure(config); err != nil {
			return nil, err
		}
		return provider, nil

	case "redis":
		provider := redis.NewProvider(logger)
		config := map[string]interface{}{
			"host": getEnv("REDIS_HOST", "localhost"),
			"port": getEnvInt("REDIS_PORT", 6379),
			"db":   getEnvInt("REDIS_DB", 0),
		}
		if err := provider.Configure(config); err != nil {
			return nil, err
		}
		return provider, nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
