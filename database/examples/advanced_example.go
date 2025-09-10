package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/anasamu/microservices-library-go/database/gateway"
	"github.com/anasamu/microservices-library-go/database/migrations"
	"github.com/anasamu/microservices-library-go/database/providers/cassandra"
	"github.com/anasamu/microservices-library-go/database/providers/influxdb"
	"github.com/anasamu/microservices-library-go/database/providers/postgresql"
	"github.com/anasamu/microservices-library-go/database/providers/sqlite"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create database manager
	config := gateway.DefaultManagerConfig()
	config.DefaultProvider = "postgresql"
	config.MaxConnections = 100

	databaseManager := gateway.NewDatabaseManager(config, logger)

	// Register multiple providers
	registerProviders(databaseManager, logger)

	// Connect to databases
	ctx := context.Background()
	if err := connectToDatabases(ctx, databaseManager); err != nil {
		log.Fatal(err)
	}

	// Demonstrate migration system
	if err := demonstrateMigrations(ctx, databaseManager, logger); err != nil {
		log.Fatal(err)
	}

	// Demonstrate different database operations
	if err := demonstrateOperations(ctx, databaseManager, logger); err != nil {
		log.Fatal(err)
	}

	// Clean up
	if err := databaseManager.Close(); err != nil {
		log.Printf("Error closing database connections: %v", err)
	}
}

func registerProviders(databaseManager *gateway.DatabaseManager, logger *logrus.Logger) {
	// PostgreSQL Provider
	postgresProvider := postgresql.NewProvider(logger)
	postgresConfig := map[string]interface{}{
		"host":            "localhost",
		"port":            5432,
		"user":            "postgres",
		"password":        "password",
		"database":        "testdb",
		"ssl_mode":        "disable",
		"max_connections": 50,
		"min_connections": 5,
	}

	if err := postgresProvider.Configure(postgresConfig); err != nil {
		log.Printf("Warning: Failed to configure PostgreSQL: %v", err)
	} else {
		databaseManager.RegisterProvider(postgresProvider)
		logger.Info("PostgreSQL provider registered")
	}

	// SQLite Provider
	sqliteProvider := sqlite.NewProvider(logger)
	sqliteConfig := map[string]interface{}{
		"file":                 "./test.db",
		"journal_mode":         "WAL",
		"max_connections":      1,
		"max_idle_connections": 1,
	}

	if err := sqliteProvider.Configure(sqliteConfig); err != nil {
		log.Printf("Warning: Failed to configure SQLite: %v", err)
	} else {
		databaseManager.RegisterProvider(sqliteProvider)
		logger.Info("SQLite provider registered")
	}

	// Cassandra Provider
	cassandraProvider := cassandra.NewProvider(logger)
	cassandraConfig := map[string]interface{}{
		"hosts":           []string{"localhost"},
		"keyspace":        "test_keyspace",
		"username":        "",
		"password":        "",
		"consistency":     "quorum",
		"timeout":         10,
		"num_connections": 2,
	}

	if err := cassandraProvider.Configure(cassandraConfig); err != nil {
		log.Printf("Warning: Failed to configure Cassandra: %v", err)
	} else {
		databaseManager.RegisterProvider(cassandraProvider)
		logger.Info("Cassandra provider registered")
	}

	// InfluxDB Provider
	influxProvider := influxdb.NewProvider(logger)
	influxConfig := map[string]interface{}{
		"url":     "http://localhost:8086",
		"token":   "your-token-here",
		"org":     "your-org",
		"bucket":  "test-bucket",
		"timeout": 30,
	}

	if err := influxProvider.Configure(influxConfig); err != nil {
		log.Printf("Warning: Failed to configure InfluxDB: %v", err)
	} else {
		databaseManager.RegisterProvider(influxProvider)
		logger.Info("InfluxDB provider registered")
	}
}

func connectToDatabases(ctx context.Context, databaseManager *gateway.DatabaseManager) error {
	providers := []string{"postgresql", "sqlite", "cassandra", "influxdb"}

	for _, provider := range providers {
		if databaseManager.IsProviderConnected(provider) {
			log.Printf("Connecting to %s...", provider)
			if err := databaseManager.Connect(ctx, provider); err != nil {
				log.Printf("Warning: Failed to connect to %s: %v", provider, err)
			} else {
				log.Printf("Successfully connected to %s", provider)
			}
		}
	}

	return nil
}

func demonstrateMigrations(ctx context.Context, databaseManager *gateway.DatabaseManager, logger *logrus.Logger) error {
	logger.Info("=== Demonstrating Migration System ===")

	// Create migrations directory
	migrationsDir := "./migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Get provider and create CLI manager
	provider, err := databaseManager.GetProvider("postgresql")
	if err != nil {
		log.Printf("Warning: Failed to get PostgreSQL provider: %v", err)
		return nil
	}

	cliManager := migrations.NewCLIManager(provider, migrationsDir, logger)

	// Create a sample migration
	if err := cliManager.CreateMigration("create_users_table"); err != nil {
		log.Printf("Warning: Failed to create migration: %v", err)
		return nil
	}

	// Load and display migrations
	migrations, err := cliManager.LoadMigrations()
	if err != nil {
		log.Printf("Warning: Failed to load migrations: %v", err)
		return nil
	}

	logger.WithField("count", len(migrations)).Info("Loaded migrations")

	// Show migration status
	if err := cliManager.Status(ctx); err != nil {
		log.Printf("Warning: Failed to get migration status: %v", err)
	}

	return nil
}

func demonstrateOperations(ctx context.Context, databaseManager *gateway.DatabaseManager, logger *logrus.Logger) error {
	logger.Info("=== Demonstrating Database Operations ===")

	// Demonstrate PostgreSQL operations
	if err := demonstratePostgreSQL(ctx, databaseManager, logger); err != nil {
		log.Printf("Warning: PostgreSQL operations failed: %v", err)
	}

	// Demonstrate SQLite operations
	if err := demonstrateSQLite(ctx, databaseManager, logger); err != nil {
		log.Printf("Warning: SQLite operations failed: %v", err)
	}

	// Demonstrate Cassandra operations
	if err := demonstrateCassandra(ctx, databaseManager, logger); err != nil {
		log.Printf("Warning: Cassandra operations failed: %v", err)
	}

	// Demonstrate InfluxDB operations
	if err := demonstrateInfluxDB(ctx, databaseManager, logger); err != nil {
		log.Printf("Warning: InfluxDB operations failed: %v", err)
	}

	return nil
}

func demonstratePostgreSQL(ctx context.Context, databaseManager *gateway.DatabaseManager, logger *logrus.Logger) error {
	if !databaseManager.IsProviderConnected("postgresql") {
		return fmt.Errorf("PostgreSQL not connected")
	}

	logger.Info("--- PostgreSQL Operations ---")

	// Create a test table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS test_users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := databaseManager.Exec(ctx, "postgresql", createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Insert data
	insertSQL := "INSERT INTO test_users (name, email) VALUES ($1, $2)"
	_, err = databaseManager.Exec(ctx, "postgresql", insertSQL, "John Doe", "john@example.com")
	if err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}

	// Query data
	querySQL := "SELECT id, name, email FROM test_users WHERE name = $1"
	rows, err := databaseManager.Query(ctx, "postgresql", querySQL, "John Doe")
	if err != nil {
		return fmt.Errorf("failed to query data: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		logger.WithFields(logrus.Fields{
			"id":    id,
			"name":  name,
			"email": email,
		}).Info("PostgreSQL query result")
	}

	// Get statistics
	stats, err := databaseManager.GetStats(ctx, "postgresql")
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"active_connections": stats.ActiveConnections,
		"max_connections":    stats.MaxConnections,
	}).Info("PostgreSQL statistics")

	return nil
}

func demonstrateSQLite(ctx context.Context, databaseManager *gateway.DatabaseManager, logger *logrus.Logger) error {
	if !databaseManager.IsProviderConnected("sqlite") {
		return fmt.Errorf("SQLite not connected")
	}

	logger.Info("--- SQLite Operations ---")

	// Create a test table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS test_products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := databaseManager.Exec(ctx, "sqlite", createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Insert data
	insertSQL := "INSERT INTO test_products (name, price) VALUES (?, ?)"
	_, err = databaseManager.Exec(ctx, "sqlite", insertSQL, "Laptop", 999.99)
	if err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}

	// Query data
	querySQL := "SELECT id, name, price FROM test_products WHERE name = ?"
	row, err := databaseManager.QueryRow(ctx, "sqlite", querySQL, "Laptop")
	if err != nil {
		return fmt.Errorf("failed to query data: %w", err)
	}

	var id int
	var name string
	var price float64
	if err := row.Scan(&id, &name, &price); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"id":    id,
		"name":  name,
		"price": price,
	}).Info("SQLite query result")

	return nil
}

func demonstrateCassandra(ctx context.Context, databaseManager *gateway.DatabaseManager, logger *logrus.Logger) error {
	if !databaseManager.IsProviderConnected("cassandra") {
		return fmt.Errorf("Cassandra not connected")
	}

	logger.Info("--- Cassandra Operations ---")

	// Create a keyspace (this would typically be done outside the application)
	createKeyspaceSQL := `
		CREATE KEYSPACE IF NOT EXISTS test_keyspace 
		WITH REPLICATION = {
			'class': 'SimpleStrategy',
			'replication_factor': 1
		}
	`

	_, err := databaseManager.Exec(ctx, "cassandra", createKeyspaceSQL)
	if err != nil {
		logger.WithError(err).Warn("Failed to create keyspace (may already exist)")
	}

	// Create a table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS test_keyspace.test_events (
			id UUID PRIMARY KEY,
			event_type TEXT,
			timestamp TIMESTAMP,
			data TEXT
		)
	`

	_, err = databaseManager.Exec(ctx, "cassandra", createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Insert data
	insertSQL := `
		INSERT INTO test_keyspace.test_events (id, event_type, timestamp, data)
		VALUES (now(), ?, toTimestamp(now()), ?)
	`

	_, err = databaseManager.Exec(ctx, "cassandra", insertSQL, "user_login", "User logged in successfully")
	if err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}

	// Query data
	querySQL := "SELECT id, event_type, timestamp, data FROM test_keyspace.test_events LIMIT 1"
	row, err := databaseManager.QueryRow(ctx, "cassandra", querySQL)
	if err != nil {
		return fmt.Errorf("failed to query data: %w", err)
	}

	var id, eventType, timestamp, data string
	if err := row.Scan(&id, &eventType, &timestamp, &data); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"id":         id,
		"event_type": eventType,
		"timestamp":  timestamp,
		"data":       data,
	}).Info("Cassandra query result")

	return nil
}

func demonstrateInfluxDB(ctx context.Context, databaseManager *gateway.DatabaseManager, logger *logrus.Logger) error {
	if !databaseManager.IsProviderConnected("influxdb") {
		return fmt.Errorf("InfluxDB not connected")
	}

	logger.Info("--- InfluxDB Operations ---")

	// InfluxDB uses a different query language (Flux)
	// For demonstration, we'll use a simple query
	querySQL := `
		from(bucket: "test-bucket")
		|> range(start: -1h)
		|> limit(n: 1)
	`

	rows, err := databaseManager.Query(ctx, "influxdb", querySQL)
	if err != nil {
		return fmt.Errorf("failed to query data: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var result map[string]interface{}
		if err := rows.Scan(&result); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		logger.WithField("result", result).Info("InfluxDB query result")
	}

	return nil
}

func cleanup() {
	// Clean up test files
	files := []string{"./test.db", "./migrations"}
	for _, file := range files {
		if err := os.RemoveAll(file); err != nil {
			log.Printf("Warning: Failed to remove %s: %v", file, err)
		}
	}
}
