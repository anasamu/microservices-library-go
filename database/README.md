# Database Gateway Library

A comprehensive, modular, and production-ready database library for Go microservices. This library provides a unified interface for multiple database providers including PostgreSQL, MongoDB, Redis, and MySQL.

## 🚀 Features

### 🔧 Multi-Provider Support
- **PostgreSQL**: Full PostgreSQL support with transactions, prepared statements, and connection pooling
- **MongoDB**: Complete MongoDB integration with document operations and aggregation
- **Redis**: Redis support for key-value operations, pub/sub, and data structures
- **MySQL**: Full MySQL support with transactions and prepared statements
- **SQLite**: Embedded SQLite database with WAL mode and connection pooling
- **Cassandra**: Apache Cassandra support with CQL operations and clustering
- **InfluxDB**: Time-series database support with Flux queries and measurements
- **Elasticsearch**: Search and analytics engine integration

### 📊 Core Operations
- **Query Operations**: Execute queries, query rows, and execute statements
- **Transaction Management**: Begin, commit, and rollback transactions
- **Prepared Statements**: Prepare and execute parameterized queries
- **Connection Management**: Connect, disconnect, and ping databases
- **Health Monitoring**: Comprehensive health checks and statistics

### 🔗 Advanced Features
- **Connection Pooling**: Configurable connection pools for optimal performance
- **Retry Logic**: Automatic retry with exponential backoff
- **Health Checks**: Real-time database health monitoring
- **Statistics**: Detailed connection and performance statistics
- **Error Handling**: Comprehensive error reporting with context
- **Database Migrations**: Version-controlled schema migrations with rollback support
- **CLI Tools**: Command-line interface for migration management
- **Multi-Database Support**: Run migrations across different database types

### 🏥 Production Features
- **Connection Management**: Automatic connection lifecycle management
- **Timeout Handling**: Configurable timeouts for all operations
- **Logging**: Structured logging with detailed context
- **Configuration**: Environment variable and file-based configuration
- **Monitoring**: Built-in metrics and health monitoring

## 📁 Project Structure

```
libs/database/
├── gateway/                    # Core database gateway
│   ├── manager.go             # Database manager implementation
│   ├── example.go             # Usage examples
│   └── go.mod                 # Gateway dependencies
├── providers/                 # Database provider implementations
│   ├── postgresql/            # PostgreSQL provider
│   │   ├── provider.go        # PostgreSQL implementation
│   │   └── go.mod             # PostgreSQL dependencies
│   ├── mongodb/               # MongoDB provider
│   │   ├── provider.go        # MongoDB implementation
│   │   └── go.mod             # MongoDB dependencies
│   ├── redis/                 # Redis provider
│   │   ├── provider.go        # Redis implementation
│   │   └── go.mod             # Redis dependencies
│   ├── mysql/                 # MySQL provider
│   │   ├── provider.go        # MySQL implementation
│   │   └── go.mod             # MySQL dependencies
│   ├── sqlite/                # SQLite provider
│   │   ├── provider.go        # SQLite implementation
│   │   └── go.mod             # SQLite dependencies
│   ├── cassandra/             # Cassandra provider
│   │   ├── provider.go        # Cassandra implementation
│   │   └── go.mod             # Cassandra dependencies
│   ├── influxdb/              # InfluxDB provider
│   │   ├── provider.go        # InfluxDB implementation
│   │   └── go.mod             # InfluxDB dependencies
│   └── elasticsearch/         # Elasticsearch provider
│       ├── provider.go        # Elasticsearch implementation
│       └── go.mod             # Elasticsearch dependencies
├── migrations/                # Database migration system
│   ├── migration.go           # Migration manager
│   └── cli.go                 # CLI tools
├── cmd/                       # Command-line tools
│   └── migrate/               # Migration CLI
│       └── main.go            # CLI implementation
├── go.mod                     # Main module dependencies
└── README.md                  # This file
```

## 🛠️ Installation

### Prerequisites
- Go 1.21 or higher
- Git

### Basic Installation

```bash
# Clone the repository
git clone https://github.com/anasamu/microservices-library-go.git
cd microservices-library-go/libs/database

# Install dependencies
go mod tidy
```

### Using Specific Providers

```bash
# For PostgreSQL support
go get github.com/anasamu/microservices-library-go/libs/database/providers/postgresql

# For MongoDB support
go get github.com/anasamu/microservices-library-go/libs/database/providers/mongodb

# For Redis support
go get github.com/anasamu/microservices-library-go/libs/database/providers/redis

# For MySQL support
go get github.com/anasamu/microservices-library-go/libs/database/providers/mysql

# For SQLite support
go get github.com/anasamu/microservices-library-go/libs/database/providers/sqlite

# For Cassandra support
go get github.com/anasamu/microservices-library-go/libs/database/providers/cassandra

# For InfluxDB support
go get github.com/anasamu/microservices-library-go/libs/database/providers/influxdb

# For Elasticsearch support
go get github.com/anasamu/microservices-library-go/libs/database/providers/elasticsearch
```

## 📖 Usage Examples

### Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/anasamu/microservices-library-go/libs/database/gateway"
    "github.com/anasamu/microservices-library-go/libs/database/providers/postgresql"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create database manager
    config := gateway.DefaultManagerConfig()
    config.DefaultProvider = "postgresql"
    config.MaxConnections = 100
    
    databaseManager := gateway.NewDatabaseManager(config, logger)
    
    // Register PostgreSQL provider
    postgresProvider := postgresql.NewProvider(logger)
    postgresConfig := map[string]interface{}{
        "host":     "localhost",
        "port":     5432,
        "user":     "postgres",
        "password": "password",
        "database": "testdb",
        "ssl_mode": "disable",
    }
    
    if err := postgresProvider.Configure(postgresConfig); err != nil {
        log.Fatal(err)
    }
    
    databaseManager.RegisterProvider(postgresProvider)
    
    // Use the database manager...
}
```

### Connect to Database

```go
// Connect to database
ctx := context.Background()
err := databaseManager.Connect(ctx, "postgresql")
if err != nil {
    log.Fatal(err)
}

// Check connection
if databaseManager.IsProviderConnected("postgresql") {
    log.Println("Database connected successfully")
}

// Ping database
err = databaseManager.Ping(ctx, "postgresql")
if err != nil {
    log.Fatal(err)
}
```

### Execute Queries

```go
// Execute a query
rows, err := databaseManager.Query(ctx, "postgresql", 
    "SELECT id, name, email FROM users WHERE active = $1", true)
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

// Process results
for rows.Next() {
    var id int
    var name, email string
    if err := rows.Scan(&id, &name, &email); err != nil {
        log.Fatal(err)
    }
    log.Printf("User: %d, %s, %s", id, name, email)
}
```

### Query Single Row

```go
// Query a single row
row, err := databaseManager.QueryRow(ctx, "postgresql",
    "SELECT id, name, email FROM users WHERE id = $1", userID)
if err != nil {
    log.Fatal(err)
}

var id int
var name, email string
if err := row.Scan(&id, &name, &email); err != nil {
    log.Fatal(err)
}

log.Printf("User: %d, %s, %s", id, name, email)
```

### SQLite Provider

```go
// Register SQLite provider
sqliteProvider := sqlite.NewProvider(logger)
sqliteConfig := map[string]interface{}{
    "file":                "./app.db",
    "journal_mode":        "WAL",
    "max_connections":     1,
    "max_idle_connections": 1,
}

if err := sqliteProvider.Configure(sqliteConfig); err != nil {
    log.Fatal(err)
}

databaseManager.RegisterProvider(sqliteProvider)

// Connect to SQLite
err := databaseManager.Connect(ctx, "sqlite")
if err != nil {
    log.Fatal(err)
}

// Create table
_, err = databaseManager.Exec(ctx, "sqlite", `
    CREATE TABLE IF NOT EXISTS products (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        price REAL NOT NULL
    )
`)

// Insert data
_, err = databaseManager.Exec(ctx, "sqlite", 
    "INSERT INTO products (name, price) VALUES (?, ?)", "Laptop", 999.99)
```

### Cassandra Provider

```go
// Register Cassandra provider
cassandraProvider := cassandra.NewProvider(logger)
cassandraConfig := map[string]interface{}{
    "hosts":          []string{"localhost"},
    "keyspace":       "myapp",
    "username":       "",
    "password":       "",
    "consistency":    "quorum",
    "timeout":        10,
    "num_connections": 2,
}

if err := cassandraProvider.Configure(cassandraConfig); err != nil {
    log.Fatal(err)
}

databaseManager.RegisterProvider(cassandraProvider)

// Connect to Cassandra
err := databaseManager.Connect(ctx, "cassandra")
if err != nil {
    log.Fatal(err)
}

// Create table
_, err = databaseManager.Exec(ctx, "cassandra", `
    CREATE TABLE IF NOT EXISTS myapp.events (
        id UUID PRIMARY KEY,
        event_type TEXT,
        timestamp TIMESTAMP,
        data TEXT
    )
`)

// Insert data
_, err = databaseManager.Exec(ctx, "cassandra", `
    INSERT INTO myapp.events (id, event_type, timestamp, data)
    VALUES (now(), ?, toTimestamp(now()), ?)
`, "user_login", "User logged in successfully")
```

### InfluxDB Provider

```go
// Register InfluxDB provider
influxProvider := influxdb.NewProvider(logger)
influxConfig := map[string]interface{}{
    "url":     "http://localhost:8086",
    "token":   "your-token-here",
    "org":     "your-org",
    "bucket":  "metrics",
    "timeout": 30,
}

if err := influxProvider.Configure(influxConfig); err != nil {
    log.Fatal(err)
}

databaseManager.RegisterProvider(influxProvider)

// Connect to InfluxDB
err := databaseManager.Connect(ctx, "influxdb")
if err != nil {
    log.Fatal(err)
}

// Query time-series data
rows, err := databaseManager.Query(ctx, "influxdb", `
    from(bucket: "metrics")
    |> range(start: -1h)
    |> filter(fn: (r) => r._measurement == "cpu_usage")
    |> limit(n: 10)
`)

for rows.Next() {
    var result map[string]interface{}
    if err := rows.Scan(&result); err != nil {
        log.Fatal(err)
    }
    log.Printf("Time series data: %+v", result)
}
```

### Execute Statements

```go
// Execute a statement
result, err := databaseManager.Exec(ctx, "postgresql",
    "INSERT INTO users (name, email) VALUES ($1, $2)", "John Doe", "john@example.com")
if err != nil {
    log.Fatal(err)
}

rowsAffected, err := result.RowsAffected()
if err != nil {
    log.Fatal(err)
}

log.Printf("Inserted %d rows", rowsAffected)
```

### Database Migrations

The library includes a comprehensive migration system that supports version-controlled schema changes across all database providers.

#### Creating Migrations

```go
// Create a new migration
cliManager, err := databaseManager.GetCLIManager("postgresql", "./migrations")
if err != nil {
    log.Fatal(err)
}

err = cliManager.CreateMigration("create_users_table")
if err != nil {
    log.Fatal(err)
}
```

#### Running Migrations

```go
// Apply all pending migrations
err = cliManager.Up(ctx)
if err != nil {
    log.Fatal(err)
}

// Rollback the last migration
err = cliManager.Down(ctx)
if err != nil {
    log.Fatal(err)
}

// Check migration status
err = cliManager.Status(ctx)
if err != nil {
    log.Fatal(err)
}
```

#### Migration File Format

Migrations are stored as JSON files with the following structure:

```json
{
  "version": "20231201120000",
  "description": "create_users_table",
  "up_sql": "CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(100) NOT NULL)",
  "down_sql": "DROP TABLE users",
  "created_at": "2023-12-01T12:00:00Z",
  "checksum": "abc123..."
}
```

#### CLI Tool Usage

```bash
# Create a new migration
./migrate -provider=postgresql -cmd=create -name="add_user_email_index"

# Apply all pending migrations
./migrate -provider=postgresql -cmd=up

# Rollback the last migration
./migrate -provider=postgresql -cmd=down

# Check migration status
./migrate -provider=postgresql -cmd=status

# Validate all migration files
./migrate -provider=postgresql -cmd=validate

# Reset database (drop all and reapply)
./migrate -provider=postgresql -cmd=reset
```

### Transactions

```go
// Execute within a transaction
err := databaseManager.WithTransaction(ctx, "postgresql", func(tx gateway.Transaction) error {
    // Insert user
    _, err := tx.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", 
        "Alice Smith", "alice@example.com")
    if err != nil {
        return err
    }
    
    // Insert user profile
    _, err = tx.Exec(ctx, "INSERT INTO user_profiles (user_id, bio) VALUES ($1, $2)", 
        userID, "Software Engineer")
    if err != nil {
        return err
    }
    
    return nil
})

if err != nil {
    log.Fatal(err)
}
```

### Prepared Statements

```go
// Prepare a statement
stmt, err := databaseManager.Prepare(ctx, "postgresql",
    "SELECT id, name, email FROM users WHERE id = $1")
if err != nil {
    log.Fatal(err)
}
defer stmt.Close()

// Execute prepared statement
row, err := stmt.QueryRow(ctx, userID)
if err != nil {
    log.Fatal(err)
}

var id int
var name, email string
if err := row.Scan(&id, &name, &email); err != nil {
    log.Fatal(err)
}
```

### Health Checks

```go
// Perform health checks
results := databaseManager.HealthCheck(ctx)

for provider, err := range results {
    if err != nil {
        log.Printf("%s: ❌ %v", provider, err)
    } else {
        log.Printf("%s: ✅ Healthy", provider)
    }
}
```

### Database Statistics

```go
// Get database statistics
stats, err := databaseManager.GetStats(ctx, "postgresql")
if err != nil {
    log.Fatal(err)
}

log.Printf("Active Connections: %d", stats.ActiveConnections)
log.Printf("Idle Connections: %d", stats.IdleConnections)
log.Printf("Max Connections: %d", stats.MaxConnections)
log.Printf("Wait Count: %d", stats.WaitCount)
```

## 🔧 Configuration

### Environment Variables

The library supports configuration through environment variables:

```bash
# Database Manager Configuration
export DB_DEFAULT_PROVIDER="postgresql"
export DB_MAX_CONNECTIONS="100"
export DB_RETRY_ATTEMPTS="3"
export DB_RETRY_DELAY="5s"
export DB_TIMEOUT="30s"

# PostgreSQL Configuration
export POSTGRES_HOST="localhost"
export POSTGRES_PORT="5432"
export POSTGRES_USER="postgres"
export POSTGRES_PASSWORD="password"
export POSTGRES_DATABASE="testdb"
export POSTGRES_SSL_MODE="disable"

# MongoDB Configuration
export MONGO_URI="mongodb://localhost:27017"
export MONGO_DATABASE="testdb"
export MONGO_MAX_POOL="100"
export MONGO_MIN_POOL="10"

# Redis Configuration
export REDIS_HOST="localhost"
export REDIS_PORT="6379"
export REDIS_PASSWORD=""
export REDIS_DB="0"
export REDIS_POOL_SIZE="100"

# MySQL Configuration
export MYSQL_HOST="localhost"
export MYSQL_PORT="3306"
export MYSQL_USER="root"
export MYSQL_PASSWORD="password"
export MYSQL_DATABASE="testdb"

# SQLite Configuration
export SQLITE_FILE="./app.db"
export SQLITE_JOURNAL_MODE="WAL"

# Cassandra Configuration
export CASSANDRA_HOST="localhost"
export CASSANDRA_KEYSPACE="myapp"
export CASSANDRA_USERNAME=""
export CASSANDRA_PASSWORD=""
export CASSANDRA_CONSISTENCY="quorum"

# InfluxDB Configuration
export INFLUXDB_URL="http://localhost:8086"
export INFLUXDB_TOKEN="your-token-here"
export INFLUXDB_ORG="your-org"
export INFLUXDB_BUCKET="metrics"
```

### Configuration Files

You can also use configuration files:

```json
{
  "database": {
    "default_provider": "postgresql",
    "max_connections": 100,
    "retry_attempts": 3,
    "retry_delay": "5s",
    "timeout": "30s"
  },
  "providers": {
    "postgresql": {
      "host": "localhost",
      "port": 5432,
      "user": "postgres",
      "password": "password",
      "database": "testdb",
      "ssl_mode": "disable"
    },
    "mongodb": {
      "uri": "mongodb://localhost:27017",
      "database": "testdb",
      "max_pool": 100,
      "min_pool": 10
    },
    "redis": {
      "host": "localhost",
      "port": 6379,
      "password": "",
      "db": 0,
      "pool_size": 100
    },
    "mysql": {
      "host": "localhost",
      "port": 3306,
      "user": "root",
      "password": "password",
      "database": "testdb"
    },
    "sqlite": {
      "file": "./app.db",
      "journal_mode": "WAL"
    },
    "cassandra": {
      "hosts": ["localhost"],
      "keyspace": "myapp",
      "username": "",
      "password": "",
      "consistency": "quorum"
    },
    "influxdb": {
      "url": "http://localhost:8086",
      "token": "your-token-here",
      "org": "your-org",
      "bucket": "metrics"
    }
  }
}
```

## 🧪 Testing

Run tests for all modules:

```bash
# Run all tests
go test ./...

# Run tests for specific provider
go test ./providers/postgresql/...
go test ./providers/mongodb/...
go test ./providers/redis/...
go test ./providers/mysql/...

# Run gateway tests
go test ./gateway/...
```

## 📚 API Documentation

### Database Manager API

- `NewDatabaseManager(config, logger)` - Create database manager
- `RegisterProvider(provider)` - Register a database provider
- `Connect(ctx, provider)` - Connect to a database
- `Disconnect(ctx, provider)` - Disconnect from a database
- `Ping(ctx, provider)` - Ping a database
- `Query(ctx, provider, query, args...)` - Execute a query
- `QueryRow(ctx, provider, query, args...)` - Execute a query returning a single row
- `Exec(ctx, provider, query, args...)` - Execute a statement
- `BeginTransaction(ctx, provider)` - Begin a transaction
- `WithTransaction(ctx, provider, fn)` - Execute within a transaction
- `Prepare(ctx, provider, query)` - Prepare a statement
- `HealthCheck(ctx)` - Check provider health
- `GetStats(ctx, provider)` - Get database statistics
- `GetMigrationManager(provider)` - Get migration manager for provider
- `GetCLIManager(provider, dir)` - Get CLI manager for migrations
- `RunMigrations(ctx, provider, dir)` - Run all pending migrations
- `GetMigrationStatus(ctx, provider, dir)` - Get migration status

### Provider Interface

All providers implement the `DatabaseProvider` interface:

```go
type DatabaseProvider interface {
    GetName() string
    GetSupportedFeatures() []DatabaseFeature
    GetConnectionInfo() *ConnectionInfo
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    Ping(ctx context.Context) error
    IsConnected() bool
    BeginTransaction(ctx context.Context) (Transaction, error)
    WithTransaction(ctx context.Context, fn func(Transaction) error) error
    Query(ctx context.Context, query string, args ...interface{}) (QueryResult, error)
    QueryRow(ctx context.Context, query string, args ...interface{}) (Row, error)
    Exec(ctx context.Context, query string, args ...interface{}) (ExecResult, error)
    Prepare(ctx context.Context, query string) (PreparedStatement, error)
    HealthCheck(ctx context.Context) error
    GetStats(ctx context.Context) (*DatabaseStats, error)
    Configure(config map[string]interface{}) error
    IsConfigured() bool
    Close() error
}
```

### Supported Features

- `FeatureTransactions` - Transaction support
- `FeaturePreparedStmts` - Prepared statement support
- `FeatureConnectionPool` - Connection pooling
- `FeatureReadReplicas` - Read replica support
- `FeatureClustering` - Database clustering
- `FeatureSharding` - Database sharding
- `FeatureFullTextSearch` - Full-text search
- `FeatureJSONSupport` - JSON data type support
- `FeatureGeoSpatial` - Geospatial data support
- `FeatureTimeSeries` - Time series data support
- `FeatureGraphDB` - Graph database support
- `FeatureKeyValue` - Key-value store support
- `FeatureDocumentStore` - Document store support
- `FeatureColumnFamily` - Column family support
- `FeatureInMemory` - In-memory database support
- `FeaturePersistent` - Persistent storage support
- `FeatureSharding` - Database sharding support
- `FeatureClustering` - Database clustering support

## 🔒 Security Considerations

### Connection Security

- **SSL/TLS**: All providers support encrypted connections
- **Authentication**: Multiple authentication methods per provider
- **Connection Pooling**: Secure connection management
- **Timeout Handling**: Prevents hanging connections

### Data Security

- **Prepared Statements**: Protection against SQL injection
- **Parameter Binding**: Safe parameter handling
- **Transaction Isolation**: Proper transaction handling
- **Connection Encryption**: Encrypted data transmission

## 🚀 Performance

### Optimization Features

- **Connection Pooling**: Efficient connection management
- **Prepared Statements**: Optimized query execution
- **Batch Operations**: Efficient bulk operations
- **Retry Logic**: Automatic retry with backoff
- **Connection Reuse**: Minimize connection overhead

### Monitoring

- **Health Checks**: Real-time database health monitoring
- **Statistics**: Detailed performance metrics
- **Connection Monitoring**: Connection pool statistics
- **Query Performance**: Query execution monitoring

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📧 Email: support@example.com
- 💬 Discord: [Join our Discord](https://discord.gg/example)
- 📖 Documentation: [Full Documentation](https://docs.example.com)
- 🐛 Issues: [GitHub Issues](https://github.com/anasamu/microservices-library-go/issues)

## 🙏 Acknowledgments

- [PostgreSQL](https://www.postgresql.org/) for the advanced relational database
- [MongoDB](https://www.mongodb.com/) for the document database
- [Redis](https://redis.io/) for the in-memory data store
- [MySQL](https://www.mysql.com/) for the popular relational database
- [SQLite](https://www.sqlite.org/) for the embedded database
- [Apache Cassandra](https://cassandra.apache.org/) for the distributed NoSQL database
- [InfluxDB](https://www.influxdata.com/) for the time-series database
- [Elasticsearch](https://www.elastic.co/) for the search and analytics engine
- [Go Database Drivers](https://github.com/golang/go/wiki/SQLDrivers) for database connectivity
- [SQLx](https://github.com/jmoiron/sqlx) for enhanced SQL operations
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) for MongoDB integration
- [Redis Go Client](https://github.com/redis/go-redis) for Redis integration
- [Cassandra Go Driver](https://github.com/gocql/gocql) for Cassandra integration
- [InfluxDB Go Client](https://github.com/influxdata/influxdb-client-go) for InfluxDB integration

---

Made with ❤️ for the Go microservices community
