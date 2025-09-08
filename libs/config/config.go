package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Vault      VaultConfig      `mapstructure:"vault"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Search     SearchConfig     `mapstructure:"search"`
	Auth       AuthConfig       `mapstructure:"auth"`
	RabbitMQ   RabbitMQConfig   `mapstructure:"rabbitmq"`
	Kafka      KafkaConfig      `mapstructure:"kafka"`
	GRPC       GRPCConfig       `mapstructure:"grpc"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	Environment  string `mapstructure:"environment"`
	ServiceName  string `mapstructure:"service_name"`
	Version      string `mapstructure:"version"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	PostgreSQL PostgreSQLConfig `mapstructure:"postgresql"`
	MongoDB    MongoDBConfig    `mapstructure:"mongodb"`
}

// PostgreSQLConfig holds PostgreSQL configuration
type PostgreSQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	MaxConns int    `mapstructure:"max_conns"`
	MinConns int    `mapstructure:"min_conns"`
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
	MaxPool  int    `mapstructure:"max_pool"`
	MinPool  int    `mapstructure:"min_pool"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// VaultConfig holds Vault configuration
type VaultConfig struct {
	Address string `mapstructure:"address"`
	Token   string `mapstructure:"token"`
	Path    string `mapstructure:"path"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	ElasticURL string `mapstructure:"elastic_url"`
	Index      string `mapstructure:"index"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Jaeger     JaegerConfig     `mapstructure:"jaeger"`
}

// PrometheusConfig holds Prometheus configuration
type PrometheusConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    string `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// JaegerConfig holds Jaeger configuration
type JaegerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
	Service  string `mapstructure:"service"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	MinIO MinIOConfig `mapstructure:"minio"`
	S3    S3Config    `mapstructure:"s3"`
}

// MinIOConfig holds MinIO configuration
type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
}

// S3Config holds S3 configuration
type S3Config struct {
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	BucketName      string `mapstructure:"bucket_name"`
}

// SearchConfig holds search configuration
type SearchConfig struct {
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
}

// ElasticsearchConfig holds Elasticsearch configuration
type ElasticsearchConfig struct {
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Index    string `mapstructure:"index"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWT JWTConfig `mapstructure:"jwt"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey  string `mapstructure:"secret_key"`
	Expiration int    `mapstructure:"expiration"`
	RefreshExp int    `mapstructure:"refresh_exp"`
	Issuer     string `mapstructure:"issuer"`
	Audience   string `mapstructure:"audience"`
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	URL      string `mapstructure:"url"`
	Exchange string `mapstructure:"exchange"`
	Queue    string `mapstructure:"queue"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
	GroupID string   `mapstructure:"group_id"`
}

// GRPCConfig holds gRPC configuration
type GRPCConfig struct {
	Port    string `mapstructure:"port"`
	Host    string `mapstructure:"host"`
	Timeout int    `mapstructure:"timeout"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Set default values
	setDefaults()

	// Enable reading from environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Load secrets from Vault if configured
	if config.Vault.Address != "" {
		if err := loadSecretsFromVault(&config); err != nil {
			return nil, fmt.Errorf("error loading secrets from vault: %w", err)
		}
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 120)

	// Database defaults
	viper.SetDefault("database.postgresql.host", "localhost")
	viper.SetDefault("database.postgresql.port", 5432)
	viper.SetDefault("database.postgresql.sslmode", "disable")
	viper.SetDefault("database.postgresql.max_conns", 25)
	viper.SetDefault("database.postgresql.min_conns", 5)

	viper.SetDefault("database.mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("database.mongodb.max_pool", 100)
	viper.SetDefault("database.mongodb.min_pool", 10)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	// Monitoring defaults
	viper.SetDefault("monitoring.prometheus.enabled", true)
	viper.SetDefault("monitoring.prometheus.port", "9090")
	viper.SetDefault("monitoring.prometheus.path", "/metrics")

	viper.SetDefault("monitoring.jaeger.enabled", false)

	// Storage defaults
	viper.SetDefault("storage.minio.endpoint", "localhost:9000")
	viper.SetDefault("storage.minio.use_ssl", false)

	// Search defaults
	viper.SetDefault("search.elasticsearch.url", "http://localhost:9200")

	// Auth defaults
	viper.SetDefault("auth.jwt.expiration", 3600)
	viper.SetDefault("auth.jwt.refresh_exp", 86400)
	viper.SetDefault("auth.jwt.issuer", "siakad")

	// RabbitMQ defaults
	viper.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")

	// Kafka defaults
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})

	// gRPC defaults
	viper.SetDefault("grpc.port", "50051")
	viper.SetDefault("grpc.host", "0.0.0.0")
	viper.SetDefault("grpc.timeout", 30)
}

// loadSecretsFromVault loads secrets from HashiCorp Vault
func loadSecretsFromVault(config *Config) error {
	client, err := api.NewClient(&api.Config{
		Address: config.Vault.Address,
	})
	if err != nil {
		return fmt.Errorf("error creating vault client: %w", err)
	}

	client.SetToken(config.Vault.Token)

	// Load database secrets
	if config.Vault.Path != "" {
		secret, err := client.Logical().Read(config.Vault.Path)
		if err != nil {
			return fmt.Errorf("error reading vault secret: %w", err)
		}

		if secret != nil && secret.Data != nil {
			// Update database passwords
			if dbPassword, ok := secret.Data["database_password"].(string); ok {
				config.Database.PostgreSQL.Password = dbPassword
			}
			if redisPassword, ok := secret.Data["redis_password"].(string); ok {
				config.Redis.Password = redisPassword
			}
			if jwtSecret, ok := secret.Data["jwt_secret"].(string); ok {
				config.Auth.JWT.SecretKey = jwtSecret
			}
			if minioAccessKey, ok := secret.Data["minio_access_key"].(string); ok {
				config.Storage.MinIO.AccessKeyID = minioAccessKey
			}
			if minioSecretKey, ok := secret.Data["minio_secret_key"].(string); ok {
				config.Storage.MinIO.SecretAccessKey = minioSecretKey
			}
		}
	}

	return nil
}

// GetEnv returns environment variable value or default
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevelopment returns true if environment is development
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if environment is production
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// GetDatabaseURL returns PostgreSQL connection string
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.PostgreSQL.User,
		c.Database.PostgreSQL.Password,
		c.Database.PostgreSQL.Host,
		c.Database.PostgreSQL.Port,
		c.Database.PostgreSQL.DBName,
		c.Database.PostgreSQL.SSLMode,
	)
}

// GetRedisURL returns Redis connection string
func (c *Config) GetRedisURL() string {
	if c.Redis.Password != "" {
		return fmt.Sprintf("redis://:%s@%s:%d/%d",
			c.Redis.Password,
			c.Redis.Host,
			c.Redis.Port,
			c.Redis.DB,
		)
	}
	return fmt.Sprintf("redis://%s:%d/%d",
		c.Redis.Host,
		c.Redis.Port,
		c.Redis.DB,
	)
}
