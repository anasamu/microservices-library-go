package consul

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/config/types"
	"github.com/hashicorp/consul/api"
)

// Provider implements configuration provider for HashiCorp Consul
type Provider struct {
	client   *api.Client
	prefix   string
	watchers []func(*types.Config)
	stopChan chan struct{}
}

// NewProvider creates a new Consul-based configuration provider
func NewProvider(address, prefix string) (*Provider, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating consul client: %w", err)
	}

	return &Provider{
		client:   client,
		prefix:   prefix,
		watchers: make([]func(*types.Config), 0),
		stopChan: make(chan struct{}),
	}, nil
}

// Load loads configuration from Consul KV store
func (p *Provider) Load() (*types.Config, error) {
	config := &types.Config{}

	// Load configuration from Consul KV store
	kv := p.client.KV()

	// Get all keys with the prefix
	pairs, _, err := kv.List(p.prefix, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing consul keys: %w", err)
	}

	// Convert KV pairs to map
	configMap := make(map[string]string)
	for _, pair := range pairs {
		key := pair.Key
		if len(p.prefix) > 0 {
			key = key[len(p.prefix)+1:] // Remove prefix and slash
		}
		configMap[key] = string(pair.Value)
	}

	// Parse configuration
	config.Server = types.ServerConfig{
		Port:         p.getString(configMap, "server/port", "8080"),
		Host:         p.getString(configMap, "server/host", "0.0.0.0"),
		Environment:  p.getString(configMap, "server/environment", "development"),
		ServiceName:  p.getString(configMap, "server/service_name", ""),
		Version:      p.getString(configMap, "server/version", ""),
		ReadTimeout:  p.getInt(configMap, "server/read_timeout", 30),
		WriteTimeout: p.getInt(configMap, "server/write_timeout", 30),
		IdleTimeout:  p.getInt(configMap, "server/idle_timeout", 120),
	}

	config.Database = types.DatabaseConfig{
		PostgreSQL: types.PostgreSQLConfig{
			Host:     p.getString(configMap, "database/postgresql/host", "localhost"),
			Port:     p.getInt(configMap, "database/postgresql/port", 5432),
			User:     p.getString(configMap, "database/postgresql/user", ""),
			Password: p.getString(configMap, "database/postgresql/password", ""),
			DBName:   p.getString(configMap, "database/postgresql/dbname", ""),
			SSLMode:  p.getString(configMap, "database/postgresql/sslmode", "disable"),
			MaxConns: p.getInt(configMap, "database/postgresql/max_conns", 25),
			MinConns: p.getInt(configMap, "database/postgresql/min_conns", 5),
		},
		MongoDB: types.MongoDBConfig{
			URI:      p.getString(configMap, "database/mongodb/uri", "mongodb://localhost:27017"),
			Database: p.getString(configMap, "database/mongodb/database", ""),
			MaxPool:  p.getInt(configMap, "database/mongodb/max_pool", 100),
			MinPool:  p.getInt(configMap, "database/mongodb/min_pool", 10),
		},
	}

	config.Redis = types.RedisConfig{
		Host:     p.getString(configMap, "redis/host", "localhost"),
		Port:     p.getInt(configMap, "redis/port", 6379),
		Password: p.getString(configMap, "redis/password", ""),
		DB:       p.getInt(configMap, "redis/db", 0),
		PoolSize: p.getInt(configMap, "redis/pool_size", 10),
	}

	config.Vault = types.VaultConfig{
		Address: p.getString(configMap, "vault/address", ""),
		Token:   p.getString(configMap, "vault/token", ""),
		Path:    p.getString(configMap, "vault/path", ""),
	}

	config.Logging = types.LoggingConfig{
		Level:      p.getString(configMap, "logging/level", "info"),
		Format:     p.getString(configMap, "logging/format", "json"),
		Output:     p.getString(configMap, "logging/output", "stdout"),
		ElasticURL: p.getString(configMap, "logging/elastic_url", ""),
		Index:      p.getString(configMap, "logging/index", ""),
	}

	config.Monitoring = types.MonitoringConfig{
		Prometheus: types.PrometheusConfig{
			Enabled: p.getBool(configMap, "monitoring/prometheus/enabled", true),
			Port:    p.getString(configMap, "monitoring/prometheus/port", "9090"),
			Path:    p.getString(configMap, "monitoring/prometheus/path", "/metrics"),
		},
		Jaeger: types.JaegerConfig{
			Enabled:  p.getBool(configMap, "monitoring/jaeger/enabled", false),
			Endpoint: p.getString(configMap, "monitoring/jaeger/endpoint", ""),
			Service:  p.getString(configMap, "monitoring/jaeger/service", ""),
		},
	}

	config.Storage = types.StorageConfig{
		MinIO: types.MinIOConfig{
			Endpoint:        p.getString(configMap, "storage/minio/endpoint", "localhost:9000"),
			AccessKeyID:     p.getString(configMap, "storage/minio/access_key_id", ""),
			SecretAccessKey: p.getString(configMap, "storage/minio/secret_access_key", ""),
			UseSSL:          p.getBool(configMap, "storage/minio/use_ssl", false),
			BucketName:      p.getString(configMap, "storage/minio/bucket_name", ""),
		},
		S3: types.S3Config{
			Region:          p.getString(configMap, "storage/s3/region", ""),
			AccessKeyID:     p.getString(configMap, "storage/s3/access_key_id", ""),
			SecretAccessKey: p.getString(configMap, "storage/s3/secret_access_key", ""),
			BucketName:      p.getString(configMap, "storage/s3/bucket_name", ""),
		},
	}

	config.Search = types.SearchConfig{
		Elasticsearch: types.ElasticsearchConfig{
			URL:      p.getString(configMap, "search/elasticsearch/url", "http://localhost:9200"),
			Username: p.getString(configMap, "search/elasticsearch/username", ""),
			Password: p.getString(configMap, "search/elasticsearch/password", ""),
			Index:    p.getString(configMap, "search/elasticsearch/index", ""),
		},
	}

	config.Auth = types.AuthConfig{
		JWT: types.JWTConfig{
			SecretKey:  p.getString(configMap, "auth/jwt/secret_key", ""),
			Expiration: p.getInt(configMap, "auth/jwt/expiration", 3600),
			RefreshExp: p.getInt(configMap, "auth/jwt/refresh_exp", 86400),
			Issuer:     p.getString(configMap, "auth/jwt/issuer", "siakad"),
			Audience:   p.getString(configMap, "auth/jwt/audience", ""),
		},
	}

	config.RabbitMQ = types.RabbitMQConfig{
		URL:      p.getString(configMap, "rabbitmq/url", "amqp://guest:guest@localhost:5672/"),
		Exchange: p.getString(configMap, "rabbitmq/exchange", ""),
		Queue:    p.getString(configMap, "rabbitmq/queue", ""),
	}

	config.Kafka = types.KafkaConfig{
		Brokers: p.getStringSlice(configMap, "kafka/brokers", []string{"localhost:9092"}),
		Topic:   p.getString(configMap, "kafka/topic", ""),
		GroupID: p.getString(configMap, "kafka/group_id", ""),
	}

	config.GRPC = types.GRPCConfig{
		Port:    p.getString(configMap, "grpc/port", "50051"),
		Host:    p.getString(configMap, "grpc/host", "0.0.0.0"),
		Timeout: p.getInt(configMap, "grpc/timeout", 30),
	}

	return config, nil
}

// Save saves configuration to Consul KV store
func (p *Provider) Save(config *types.Config) error {
	kv := p.client.KV()

	// Convert config to KV pairs
	pairs := []*api.KVPair{
		{Key: p.prefix + "/server/port", Value: []byte(config.Server.Port)},
		{Key: p.prefix + "/server/host", Value: []byte(config.Server.Host)},
		{Key: p.prefix + "/server/environment", Value: []byte(config.Server.Environment)},
		{Key: p.prefix + "/server/service_name", Value: []byte(config.Server.ServiceName)},
		{Key: p.prefix + "/server/version", Value: []byte(config.Server.Version)},
		{Key: p.prefix + "/server/read_timeout", Value: []byte(fmt.Sprintf("%d", config.Server.ReadTimeout))},
		{Key: p.prefix + "/server/write_timeout", Value: []byte(fmt.Sprintf("%d", config.Server.WriteTimeout))},
		{Key: p.prefix + "/server/idle_timeout", Value: []byte(fmt.Sprintf("%d", config.Server.IdleTimeout))},

		// Database
		{Key: p.prefix + "/database/postgresql/host", Value: []byte(config.Database.PostgreSQL.Host)},
		{Key: p.prefix + "/database/postgresql/port", Value: []byte(fmt.Sprintf("%d", config.Database.PostgreSQL.Port))},
		{Key: p.prefix + "/database/postgresql/user", Value: []byte(config.Database.PostgreSQL.User)},
		{Key: p.prefix + "/database/postgresql/password", Value: []byte(config.Database.PostgreSQL.Password)},
		{Key: p.prefix + "/database/postgresql/dbname", Value: []byte(config.Database.PostgreSQL.DBName)},
		{Key: p.prefix + "/database/postgresql/sslmode", Value: []byte(config.Database.PostgreSQL.SSLMode)},
		{Key: p.prefix + "/database/postgresql/max_conns", Value: []byte(fmt.Sprintf("%d", config.Database.PostgreSQL.MaxConns))},
		{Key: p.prefix + "/database/postgresql/min_conns", Value: []byte(fmt.Sprintf("%d", config.Database.PostgreSQL.MinConns))},

		{Key: p.prefix + "/database/mongodb/uri", Value: []byte(config.Database.MongoDB.URI)},
		{Key: p.prefix + "/database/mongodb/database", Value: []byte(config.Database.MongoDB.Database)},
		{Key: p.prefix + "/database/mongodb/max_pool", Value: []byte(fmt.Sprintf("%d", config.Database.MongoDB.MaxPool))},
		{Key: p.prefix + "/database/mongodb/min_pool", Value: []byte(fmt.Sprintf("%d", config.Database.MongoDB.MinPool))},

		// Redis
		{Key: p.prefix + "/redis/host", Value: []byte(config.Redis.Host)},
		{Key: p.prefix + "/redis/port", Value: []byte(fmt.Sprintf("%d", config.Redis.Port))},
		{Key: p.prefix + "/redis/password", Value: []byte(config.Redis.Password)},
		{Key: p.prefix + "/redis/db", Value: []byte(fmt.Sprintf("%d", config.Redis.DB))},
		{Key: p.prefix + "/redis/pool_size", Value: []byte(fmt.Sprintf("%d", config.Redis.PoolSize))},

		// Auth
		{Key: p.prefix + "/auth/jwt/secret_key", Value: []byte(config.Auth.JWT.SecretKey)},
		{Key: p.prefix + "/auth/jwt/expiration", Value: []byte(fmt.Sprintf("%d", config.Auth.JWT.Expiration))},
		{Key: p.prefix + "/auth/jwt/refresh_exp", Value: []byte(fmt.Sprintf("%d", config.Auth.JWT.RefreshExp))},
		{Key: p.prefix + "/auth/jwt/issuer", Value: []byte(config.Auth.JWT.Issuer)},
		{Key: p.prefix + "/auth/jwt/audience", Value: []byte(config.Auth.JWT.Audience)},
	}

	// Write all pairs
	for _, pair := range pairs {
		_, err := kv.Put(pair, nil)
		if err != nil {
			return fmt.Errorf("error writing to consul: %w", err)
		}
	}

	return nil
}

// Watch watches for configuration changes in Consul
func (p *Provider) Watch(callback func(*types.Config)) error {
	p.watchers = append(p.watchers, callback)

	if len(p.watchers) == 1 {
		go p.watchConsul()
	}

	return nil
}

// watchConsul watches for changes in Consul
func (p *Provider) watchConsul() {
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
func (p *Provider) getString(configMap map[string]string, key string, defaultValue string) string {
	if value, ok := configMap[key]; ok {
		return value
	}
	return defaultValue
}

func (p *Provider) getInt(configMap map[string]string, key string, defaultValue int) int {
	if value, ok := configMap[key]; ok {
		if intValue, err := json.Number(value).Int64(); err == nil {
			return int(intValue)
		}
	}
	return defaultValue
}

func (p *Provider) getBool(configMap map[string]string, key string, defaultValue bool) bool {
	if value, ok := configMap[key]; ok {
		if boolValue, err := json.Marshal(value); err == nil {
			var result bool
			if err := json.Unmarshal(boolValue, &result); err == nil {
				return result
			}
		}
	}
	return defaultValue
}

func (p *Provider) getStringSlice(configMap map[string]string, key string, defaultValue []string) []string {
	if value, ok := configMap[key]; ok {
		var result []string
		if err := json.Unmarshal([]byte(value), &result); err == nil {
			return result
		}
	}
	return defaultValue
}
