package vault

import (
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/config/types"
	"github.com/hashicorp/vault/api"
)

// Provider implements configuration provider for HashiCorp Vault
type Provider struct {
	client   *api.Client
	path     string
	watchers []func(*types.Config)
	stopChan chan struct{}
}

// NewProvider creates a new Vault-based configuration provider
func NewProvider(address, token, path string) (*Provider, error) {
	client, err := api.NewClient(&api.Config{
		Address: address,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating vault client: %w", err)
	}

	client.SetToken(token)

	return &Provider{
		client:   client,
		path:     path,
		watchers: make([]func(*types.Config), 0),
		stopChan: make(chan struct{}),
	}, nil
}

// Load loads configuration from Vault
func (p *Provider) Load() (*types.Config, error) {
	config := &types.Config{}

	// Load secrets from Vault
	if p.path != "" {
		secret, err := p.client.Logical().Read(p.path)
		if err != nil {
			return nil, fmt.Errorf("error reading vault secret: %w", err)
		}

		if secret != nil && secret.Data != nil {
			// Load server configuration
			config.Server = types.ServerConfig{
				Port:         p.getString(secret.Data, "server_port", "8080"),
				Host:         p.getString(secret.Data, "server_host", "0.0.0.0"),
				Environment:  p.getString(secret.Data, "server_environment", "development"),
				ServiceName:  p.getString(secret.Data, "server_service_name", ""),
				Version:      p.getString(secret.Data, "server_version", ""),
				ReadTimeout:  p.getInt(secret.Data, "server_read_timeout", 30),
				WriteTimeout: p.getInt(secret.Data, "server_write_timeout", 30),
				IdleTimeout:  p.getInt(secret.Data, "server_idle_timeout", 120),
			}

			// Load database configuration
			config.Database = types.DatabaseConfig{
				PostgreSQL: types.PostgreSQLConfig{
					Host:     p.getString(secret.Data, "db_postgresql_host", "localhost"),
					Port:     p.getInt(secret.Data, "db_postgresql_port", 5432),
					User:     p.getString(secret.Data, "db_postgresql_user", ""),
					Password: p.getString(secret.Data, "db_postgresql_password", ""),
					DBName:   p.getString(secret.Data, "db_postgresql_dbname", ""),
					SSLMode:  p.getString(secret.Data, "db_postgresql_sslmode", "disable"),
					MaxConns: p.getInt(secret.Data, "db_postgresql_max_conns", 25),
					MinConns: p.getInt(secret.Data, "db_postgresql_min_conns", 5),
				},
				MongoDB: types.MongoDBConfig{
					URI:      p.getString(secret.Data, "db_mongodb_uri", "mongodb://localhost:27017"),
					Database: p.getString(secret.Data, "db_mongodb_database", ""),
					MaxPool:  p.getInt(secret.Data, "db_mongodb_max_pool", 100),
					MinPool:  p.getInt(secret.Data, "db_mongodb_min_pool", 10),
				},
			}

			// Load Redis configuration
			config.Redis = types.RedisConfig{
				Host:     p.getString(secret.Data, "redis_host", "localhost"),
				Port:     p.getInt(secret.Data, "redis_port", 6379),
				Password: p.getString(secret.Data, "redis_password", ""),
				DB:       p.getInt(secret.Data, "redis_db", 0),
				PoolSize: p.getInt(secret.Data, "redis_pool_size", 10),
			}

			// Load Vault configuration
			config.Vault = types.VaultConfig{
				Address: p.getString(secret.Data, "vault_address", ""),
				Token:   p.getString(secret.Data, "vault_token", ""),
				Path:    p.getString(secret.Data, "vault_path", ""),
			}

			// Load logging configuration
			config.Logging = types.LoggingConfig{
				Level:      p.getString(secret.Data, "logging_level", "info"),
				Format:     p.getString(secret.Data, "logging_format", "json"),
				Output:     p.getString(secret.Data, "logging_output", "stdout"),
				ElasticURL: p.getString(secret.Data, "logging_elastic_url", ""),
				Index:      p.getString(secret.Data, "logging_index", ""),
			}

			// Load monitoring configuration
			config.Monitoring = types.MonitoringConfig{
				Prometheus: types.PrometheusConfig{
					Enabled: p.getBool(secret.Data, "monitoring_prometheus_enabled", true),
					Port:    p.getString(secret.Data, "monitoring_prometheus_port", "9090"),
					Path:    p.getString(secret.Data, "monitoring_prometheus_path", "/metrics"),
				},
				Jaeger: types.JaegerConfig{
					Enabled:  p.getBool(secret.Data, "monitoring_jaeger_enabled", false),
					Endpoint: p.getString(secret.Data, "monitoring_jaeger_endpoint", ""),
					Service:  p.getString(secret.Data, "monitoring_jaeger_service", ""),
				},
			}

			// Load storage configuration
			config.Storage = types.StorageConfig{
				MinIO: types.MinIOConfig{
					Endpoint:        p.getString(secret.Data, "storage_minio_endpoint", "localhost:9000"),
					AccessKeyID:     p.getString(secret.Data, "storage_minio_access_key_id", ""),
					SecretAccessKey: p.getString(secret.Data, "storage_minio_secret_access_key", ""),
					UseSSL:          p.getBool(secret.Data, "storage_minio_use_ssl", false),
					BucketName:      p.getString(secret.Data, "storage_minio_bucket_name", ""),
				},
				S3: types.S3Config{
					Region:          p.getString(secret.Data, "storage_s3_region", ""),
					AccessKeyID:     p.getString(secret.Data, "storage_s3_access_key_id", ""),
					SecretAccessKey: p.getString(secret.Data, "storage_s3_secret_access_key", ""),
					BucketName:      p.getString(secret.Data, "storage_s3_bucket_name", ""),
				},
			}

			// Load search configuration
			config.Search = types.SearchConfig{
				Elasticsearch: types.ElasticsearchConfig{
					URL:      p.getString(secret.Data, "search_elasticsearch_url", "http://localhost:9200"),
					Username: p.getString(secret.Data, "search_elasticsearch_username", ""),
					Password: p.getString(secret.Data, "search_elasticsearch_password", ""),
					Index:    p.getString(secret.Data, "search_elasticsearch_index", ""),
				},
			}

			// Load auth configuration
			config.Auth = types.AuthConfig{
				JWT: types.JWTConfig{
					SecretKey:  p.getString(secret.Data, "auth_jwt_secret_key", ""),
					Expiration: p.getInt(secret.Data, "auth_jwt_expiration", 3600),
					RefreshExp: p.getInt(secret.Data, "auth_jwt_refresh_exp", 86400),
					Issuer:     p.getString(secret.Data, "auth_jwt_issuer", "siakad"),
					Audience:   p.getString(secret.Data, "auth_jwt_audience", ""),
				},
			}

			// Load RabbitMQ configuration
			config.RabbitMQ = types.RabbitMQConfig{
				URL:      p.getString(secret.Data, "rabbitmq_url", "amqp://guest:guest@localhost:5672/"),
				Exchange: p.getString(secret.Data, "rabbitmq_exchange", ""),
				Queue:    p.getString(secret.Data, "rabbitmq_queue", ""),
			}

			// Load Kafka configuration
			config.Kafka = types.KafkaConfig{
				Brokers: p.getStringSlice(secret.Data, "kafka_brokers", []string{"localhost:9092"}),
				Topic:   p.getString(secret.Data, "kafka_topic", ""),
				GroupID: p.getString(secret.Data, "kafka_group_id", ""),
			}

			// Load gRPC configuration
			config.GRPC = types.GRPCConfig{
				Port:    p.getString(secret.Data, "grpc_port", "50051"),
				Host:    p.getString(secret.Data, "grpc_host", "0.0.0.0"),
				Timeout: p.getInt(secret.Data, "grpc_timeout", 30),
			}
		}
	}

	return config, nil
}

// Save saves configuration to Vault
func (p *Provider) Save(config *types.Config) error {
	data := make(map[string]interface{})

	// Convert config to map
	data["server_port"] = config.Server.Port
	data["server_host"] = config.Server.Host
	data["server_environment"] = config.Server.Environment
	data["server_service_name"] = config.Server.ServiceName
	data["server_version"] = config.Server.Version
	data["server_read_timeout"] = config.Server.ReadTimeout
	data["server_write_timeout"] = config.Server.WriteTimeout
	data["server_idle_timeout"] = config.Server.IdleTimeout

	// Database
	data["db_postgresql_host"] = config.Database.PostgreSQL.Host
	data["db_postgresql_port"] = config.Database.PostgreSQL.Port
	data["db_postgresql_user"] = config.Database.PostgreSQL.User
	data["db_postgresql_password"] = config.Database.PostgreSQL.Password
	data["db_postgresql_dbname"] = config.Database.PostgreSQL.DBName
	data["db_postgresql_sslmode"] = config.Database.PostgreSQL.SSLMode
	data["db_postgresql_max_conns"] = config.Database.PostgreSQL.MaxConns
	data["db_postgresql_min_conns"] = config.Database.PostgreSQL.MinConns

	data["db_mongodb_uri"] = config.Database.MongoDB.URI
	data["db_mongodb_database"] = config.Database.MongoDB.Database
	data["db_mongodb_max_pool"] = config.Database.MongoDB.MaxPool
	data["db_mongodb_min_pool"] = config.Database.MongoDB.MinPool

	// Redis
	data["redis_host"] = config.Redis.Host
	data["redis_port"] = config.Redis.Port
	data["redis_password"] = config.Redis.Password
	data["redis_db"] = config.Redis.DB
	data["redis_pool_size"] = config.Redis.PoolSize

	// Auth
	data["auth_jwt_secret_key"] = config.Auth.JWT.SecretKey
	data["auth_jwt_expiration"] = config.Auth.JWT.Expiration
	data["auth_jwt_refresh_exp"] = config.Auth.JWT.RefreshExp
	data["auth_jwt_issuer"] = config.Auth.JWT.Issuer
	data["auth_jwt_audience"] = config.Auth.JWT.Audience

	// Write to Vault
	_, err := p.client.Logical().Write(p.path, data)
	if err != nil {
		return fmt.Errorf("error writing to vault: %w", err)
	}

	return nil
}

// Watch watches for configuration changes in Vault
func (p *Provider) Watch(callback func(*types.Config)) error {
	p.watchers = append(p.watchers, callback)

	if len(p.watchers) == 1 {
		go p.watchVault()
	}

	return nil
}

// watchVault watches for changes in Vault
func (p *Provider) watchVault() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			config, err := p.Load()
			if err != nil {
				continue
			}

			for _, watcher := range p.watchers {
				go watcher(config)
			}

		case <-p.stopChan:
			return
		}
	}
}

// Close closes the provider
func (p *Provider) Close() error {
	close(p.stopChan)
	return nil
}

// Helper methods for type conversion
func (p *Provider) getString(data map[string]interface{}, key string, defaultValue string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return defaultValue
}

func (p *Provider) getInt(data map[string]interface{}, key string, defaultValue int) int {
	if value, ok := data[key].(int); ok {
		return value
	}
	if value, ok := data[key].(float64); ok {
		return int(value)
	}
	return defaultValue
}

func (p *Provider) getBool(data map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := data[key].(bool); ok {
		return value
	}
	return defaultValue
}

func (p *Provider) getStringSlice(data map[string]interface{}, key string, defaultValue []string) []string {
	if value, ok := data[key].([]string); ok {
		return value
	}
	if value, ok := data[key].([]interface{}); ok {
		result := make([]string, len(value))
		for i, v := range value {
			if str, ok := v.(string); ok {
				result[i] = str
			}
		}
		return result
	}
	return defaultValue
}
