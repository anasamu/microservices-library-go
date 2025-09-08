package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/database/gateway"
	"github.com/anasamu/microservices-library-go/database/providers/elasticsearch"
	"github.com/anasamu/microservices-library-go/database/providers/mongodb"
	"github.com/anasamu/microservices-library-go/database/providers/mysql"
	"github.com/anasamu/microservices-library-go/database/providers/postgresql"
	"github.com/anasamu/microservices-library-go/database/providers/redis"
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
	config.RetryAttempts = 3
	config.RetryDelay = 5 * time.Second

	databaseManager := gateway.NewDatabaseManager(config, logger)

	// Register database providers
	registerProviders(databaseManager, logger)

	// Example 1: Connect to PostgreSQL
	examplePostgreSQLConnection(databaseManager)

	// Example 2: Connect to MongoDB
	exampleMongoDBConnection(databaseManager)

	// Example 3: Connect to Redis
	exampleRedisConnection(databaseManager)

	// Example 4: Connect to MySQL
	exampleMySQLConnection(databaseManager)

	// Example 5: Connect to Elasticsearch
	exampleElasticsearchConnection(databaseManager)

	// Example 6: Query operations
	exampleQueryOperations(databaseManager)

	// Example 7: Transaction operations
	exampleTransactionOperations(databaseManager)

	// Example 8: Health checks
	exampleHealthChecks(databaseManager)

	// Example 9: Database statistics
	exampleDatabaseStats(databaseManager)
}

// registerProviders registers all database providers
func registerProviders(databaseManager *gateway.DatabaseManager, logger *logrus.Logger) {
	// Register PostgreSQL provider
	postgresProvider := postgresql.NewProvider(logger)
	postgresConfig := map[string]interface{}{
		"host":            "localhost",
		"port":            5432,
		"user":            "postgres",
		"password":        "password",
		"database":        "testdb",
		"ssl_mode":        "disable",
		"max_connections": 100,
		"min_connections": 10,
	}
	if err := postgresProvider.Configure(postgresConfig); err != nil {
		log.Printf("Failed to configure PostgreSQL: %v", err)
	} else {
		databaseManager.RegisterProvider(postgresProvider)
	}

	// Register MongoDB provider
	mongoProvider := mongodb.NewProvider(logger)
	mongoConfig := map[string]interface{}{
		"uri":      "mongodb://localhost:27017",
		"database": "testdb",
		"max_pool": 100,
		"min_pool": 10,
		"timeout":  10 * time.Second,
	}
	if err := mongoProvider.Configure(mongoConfig); err != nil {
		log.Printf("Failed to configure MongoDB: %v", err)
	} else {
		databaseManager.RegisterProvider(mongoProvider)
	}

	// Register Redis provider
	redisProvider := redis.NewProvider(logger)
	redisConfig := map[string]interface{}{
		"host":      "localhost",
		"port":      6379,
		"password":  "",
		"db":        0,
		"pool_size": 100,
	}
	if err := redisProvider.Configure(redisConfig); err != nil {
		log.Printf("Failed to configure Redis: %v", err)
	} else {
		databaseManager.RegisterProvider(redisProvider)
	}

	// Register MySQL provider
	mysqlProvider := mysql.NewProvider(logger)
	mysqlConfig := map[string]interface{}{
		"host":            "localhost",
		"port":            3306,
		"user":            "root",
		"password":        "password",
		"database":        "testdb",
		"max_connections": 100,
		"min_connections": 10,
	}
	if err := mysqlProvider.Configure(mysqlConfig); err != nil {
		log.Printf("Failed to configure MySQL: %v", err)
	} else {
		databaseManager.RegisterProvider(mysqlProvider)
	}

	// Register Elasticsearch provider
	elasticsearchProvider := elasticsearch.NewProvider(logger)
	elasticsearchConfig := map[string]interface{}{
		"addresses": []string{"http://localhost:9200"},
		"username":  "",
		"password":  "",
		"index":     "testdb",
	}
	if err := elasticsearchProvider.Configure(elasticsearchConfig); err != nil {
		log.Printf("Failed to configure Elasticsearch: %v", err)
	} else {
		databaseManager.RegisterProvider(elasticsearchProvider)
	}
}

// examplePostgreSQLConnection demonstrates PostgreSQL connection
func examplePostgreSQLConnection(databaseManager *gateway.DatabaseManager) {
	fmt.Println("=== PostgreSQL Connection Example ===")

	ctx := context.Background()

	// Connect to PostgreSQL
	err := databaseManager.Connect(ctx, "postgresql")
	if err != nil {
		log.Printf("Failed to connect to PostgreSQL: %v", err)
		return
	}

	// Check if connected
	if databaseManager.IsProviderConnected("postgresql") {
		fmt.Println("PostgreSQL connected successfully")
	}

	// Ping the database
	err = databaseManager.Ping(ctx, "postgresql")
	if err != nil {
		log.Printf("Failed to ping PostgreSQL: %v", err)
	} else {
		fmt.Println("PostgreSQL ping successful")
	}
}

// exampleMongoDBConnection demonstrates MongoDB connection
func exampleMongoDBConnection(databaseManager *gateway.DatabaseManager) {
	fmt.Println("\n=== MongoDB Connection Example ===")

	ctx := context.Background()

	// Connect to MongoDB
	err := databaseManager.Connect(ctx, "mongodb")
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		return
	}

	// Check if connected
	if databaseManager.IsProviderConnected("mongodb") {
		fmt.Println("MongoDB connected successfully")
	}

	// Ping the database
	err = databaseManager.Ping(ctx, "mongodb")
	if err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
	} else {
		fmt.Println("MongoDB ping successful")
	}
}

// exampleRedisConnection demonstrates Redis connection
func exampleRedisConnection(databaseManager *gateway.DatabaseManager) {
	fmt.Println("\n=== Redis Connection Example ===")

	ctx := context.Background()

	// Connect to Redis
	err := databaseManager.Connect(ctx, "redis")
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
		return
	}

	// Check if connected
	if databaseManager.IsProviderConnected("redis") {
		fmt.Println("Redis connected successfully")
	}

	// Ping the database
	err = databaseManager.Ping(ctx, "redis")
	if err != nil {
		log.Printf("Failed to ping Redis: %v", err)
	} else {
		fmt.Println("Redis ping successful")
	}
}

// exampleMySQLConnection demonstrates MySQL connection
func exampleMySQLConnection(databaseManager *gateway.DatabaseManager) {
	fmt.Println("\n=== MySQL Connection Example ===")

	ctx := context.Background()

	// Connect to MySQL
	err := databaseManager.Connect(ctx, "mysql")
	if err != nil {
		log.Printf("Failed to connect to MySQL: %v", err)
		return
	}

	// Check if connected
	if databaseManager.IsProviderConnected("mysql") {
		fmt.Println("MySQL connected successfully")
	}

	// Ping the database
	err = databaseManager.Ping(ctx, "mysql")
	if err != nil {
		log.Printf("Failed to ping MySQL: %v", err)
	} else {
		fmt.Println("MySQL ping successful")
	}
}

// exampleElasticsearchConnection demonstrates Elasticsearch connection
func exampleElasticsearchConnection(databaseManager *gateway.DatabaseManager) {
	fmt.Println("\n=== Elasticsearch Connection Example ===")

	ctx := context.Background()

	// Connect to Elasticsearch
	err := databaseManager.Connect(ctx, "elasticsearch")
	if err != nil {
		log.Printf("Failed to connect to Elasticsearch: %v", err)
		return
	}

	// Check if connected
	if databaseManager.IsProviderConnected("elasticsearch") {
		fmt.Println("Elasticsearch connected successfully")
	}

	// Ping the database
	err = databaseManager.Ping(ctx, "elasticsearch")
	if err != nil {
		log.Printf("Failed to ping Elasticsearch: %v", err)
	} else {
		fmt.Println("Elasticsearch ping successful")
	}
}

// exampleQueryOperations demonstrates query operations
func exampleQueryOperations(databaseManager *gateway.DatabaseManager) {
	fmt.Println("\n=== Query Operations Example ===")

	ctx := context.Background()

	// Example with PostgreSQL
	if databaseManager.IsProviderConnected("postgresql") {
		fmt.Println("PostgreSQL Query Operations:")

		// Create a test table
		createTableQuery := `
			CREATE TABLE IF NOT EXISTS users (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(100) UNIQUE NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
		_, err := databaseManager.Exec(ctx, "postgresql", createTableQuery)
		if err != nil {
			log.Printf("Failed to create table: %v", err)
		} else {
			fmt.Println("Table created successfully")
		}

		// Insert a user
		insertQuery := "INSERT INTO users (name, email) VALUES ($1, $2)"
		result, err := databaseManager.Exec(ctx, "postgresql", insertQuery, "John Doe", "john@example.com")
		if err != nil {
			log.Printf("Failed to insert user: %v", err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			fmt.Printf("User inserted successfully, rows affected: %d\n", rowsAffected)
		}

		// Query users
		selectQuery := "SELECT id, name, email, created_at FROM users WHERE name = $1"
		rows, err := databaseManager.Query(ctx, "postgresql", selectQuery, "John Doe")
		if err != nil {
			log.Printf("Failed to query users: %v", err)
		} else {
			defer rows.Close()
			fmt.Println("Users found:")
			for rows.Next() {
				var id int
				var name, email string
				var createdAt time.Time
				if err := rows.Scan(&id, &name, &email, &createdAt); err != nil {
					log.Printf("Failed to scan row: %v", err)
					continue
				}
				fmt.Printf("  ID: %d, Name: %s, Email: %s, Created: %s\n", id, name, email, createdAt.Format(time.RFC3339))
			}
		}

		// Query single row
		selectOneQuery := "SELECT id, name, email FROM users WHERE email = $1"
		row, err := databaseManager.QueryRow(ctx, "postgresql", selectOneQuery, "john@example.com")
		if err != nil {
			log.Printf("Failed to query single user: %v", err)
		} else {
			var id int
			var name, email string
			if err := row.Scan(&id, &name, &email); err != nil {
				log.Printf("Failed to scan single row: %v", err)
			} else {
				fmt.Printf("Single user: ID: %d, Name: %s, Email: %s\n", id, name, email)
			}
		}
	}

	// Example with Redis
	if databaseManager.IsProviderConnected("redis") {
		fmt.Println("\nRedis Operations:")

		// Set a key-value pair
		_, err := databaseManager.Exec(ctx, "redis", "SET", "user:1", "John Doe")
		if err != nil {
			log.Printf("Failed to set Redis key: %v", err)
		} else {
			fmt.Println("Redis key set successfully")
		}

		// Get a value
		row, err := databaseManager.QueryRow(ctx, "redis", "GET", "user:1")
		if err != nil {
			log.Printf("Failed to get Redis key: %v", err)
		} else {
			var key, value string
			if err := row.Scan(&key, &value); err != nil {
				log.Printf("Failed to scan Redis row: %v", err)
			} else {
				fmt.Printf("Redis value: %s = %s\n", key, value)
			}
		}
	}
}

// exampleTransactionOperations demonstrates transaction operations
func exampleTransactionOperations(databaseManager *gateway.DatabaseManager) {
	fmt.Println("\n=== Transaction Operations Example ===")

	ctx := context.Background()

	// Example with PostgreSQL
	if databaseManager.IsProviderConnected("postgresql") {
		fmt.Println("PostgreSQL Transaction Operations:")

		// Execute transaction
		err := databaseManager.WithTransaction(ctx, "postgresql", func(tx gateway.Transaction) error {
			// Insert multiple users within transaction
			insertQuery := "INSERT INTO users (name, email) VALUES ($1, $2)"

			_, err := tx.Exec(ctx, insertQuery, "Alice Smith", "alice@example.com")
			if err != nil {
				return fmt.Errorf("failed to insert Alice: %w", err)
			}

			_, err = tx.Exec(ctx, insertQuery, "Bob Johnson", "bob@example.com")
			if err != nil {
				return fmt.Errorf("failed to insert Bob: %w", err)
			}

			fmt.Println("Transaction operations completed successfully")
			return nil
		})

		if err != nil {
			log.Printf("Transaction failed: %v", err)
		} else {
			fmt.Println("Transaction committed successfully")
		}
	}
}

// exampleHealthChecks demonstrates health checks
func exampleHealthChecks(databaseManager *gateway.DatabaseManager) {
	fmt.Println("\n=== Health Checks Example ===")

	ctx := context.Background()

	// Perform health checks on all providers
	results := databaseManager.HealthCheck(ctx)

	fmt.Printf("Health Check Results:\n")
	for provider, err := range results {
		if err != nil {
			fmt.Printf("  %s: ❌ %v\n", provider, err)
		} else {
			fmt.Printf("  %s: ✅ Healthy\n", provider)
		}
	}

	// Get connected providers
	connected := databaseManager.GetConnectedProviders()
	fmt.Printf("\nConnected Providers: %v\n", connected)
}

// exampleDatabaseStats demonstrates database statistics
func exampleDatabaseStats(databaseManager *gateway.DatabaseManager) {
	fmt.Println("\n=== Database Statistics Example ===")

	ctx := context.Background()

	// Get statistics for each provider
	providers := databaseManager.GetSupportedProviders()
	for _, providerName := range providers {
		if databaseManager.IsProviderConnected(providerName) {
			stats, err := databaseManager.GetStats(ctx, providerName)
			if err != nil {
				log.Printf("Failed to get stats for %s: %v", providerName, err)
				continue
			}

			fmt.Printf("%s Statistics:\n", providerName)
			fmt.Printf("  Active Connections: %d\n", stats.ActiveConnections)
			fmt.Printf("  Idle Connections: %d\n", stats.IdleConnections)
			fmt.Printf("  Max Connections: %d\n", stats.MaxConnections)
			fmt.Printf("  Wait Count: %d\n", stats.WaitCount)
			fmt.Printf("  Wait Duration: %v\n", stats.WaitDuration)
		}
	}
}
