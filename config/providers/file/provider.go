package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/config/types"
	"github.com/spf13/viper"
)

// Provider implements configuration provider for file-based configuration
type Provider struct {
	configPath string
	configType string
	viper      *viper.Viper
	watchers   []func(*types.Config)
	stopChan   chan struct{}
}

// NewProvider creates a new file-based configuration provider
func NewProvider(configPath, configType string) *Provider {
	return &Provider{
		configPath: configPath,
		configType: configType,
		viper:      viper.New(),
		watchers:   make([]func(*types.Config), 0),
		stopChan:   make(chan struct{}),
	}
}

// Load loads configuration from file
func (p *Provider) Load() (*types.Config, error) {
	p.viper.SetConfigFile(p.configPath)
	p.viper.SetConfigType(p.configType)

	// Set default values
	p.setDefaults()

	// Enable reading from environment variables
	p.viper.AutomaticEnv()
	p.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := p.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config types.Config
	if err := p.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// Save saves configuration to file
func (p *Provider) Save(config *types.Config) error {
	// Ensure directory exists
	dir := filepath.Dir(p.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a new viper instance for writing
	writeViper := viper.New()
	writeViper.SetConfigFile(p.configPath)
	writeViper.SetConfigType(p.configType)

	// Set all configuration values
	writeViper.Set("server", config.Server)
	writeViper.Set("database", config.Database)
	writeViper.Set("redis", config.Redis)
	writeViper.Set("vault", config.Vault)
	writeViper.Set("logging", config.Logging)
	writeViper.Set("monitoring", config.Monitoring)
	writeViper.Set("storage", config.Storage)
	writeViper.Set("search", config.Search)
	writeViper.Set("auth", config.Auth)
	writeViper.Set("rabbitmq", config.RabbitMQ)
	writeViper.Set("kafka", config.Kafka)
	writeViper.Set("grpc", config.GRPC)
	writeViper.Set("services", config.Services)

	// Set custom configuration values
	if config.Custom != nil {
		for key, value := range config.Custom {
			writeViper.Set(key, value)
		}
	}

	// Write config file
	if err := writeViper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Watch watches for configuration file changes
func (p *Provider) Watch(callback func(*types.Config)) error {
	// Add callback to watchers
	p.watchers = append(p.watchers, callback)

	// Start watching if not already started
	if len(p.watchers) == 1 {
		go p.watchFile()
	}

	return nil
}

// watchFile watches the configuration file for changes
func (p *Provider) watchFile() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastModTime time.Time

	// Get initial modification time
	if info, err := os.Stat(p.configPath); err == nil {
		lastModTime = info.ModTime()
	}

	for {
		select {
		case <-ticker.C:
			info, err := os.Stat(p.configPath)
			if err != nil {
				continue
			}

			if info.ModTime().After(lastModTime) {
				lastModTime = info.ModTime()

				// Reload configuration
				config, err := p.Load()
				if err != nil {
					continue
				}

				// Notify all watchers
				for _, watcher := range p.watchers {
					go watcher(config)
				}
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

// setDefaults sets default configuration values
func (p *Provider) setDefaults() {
	// Server defaults
	p.viper.SetDefault("server.port", "8080")
	p.viper.SetDefault("server.host", "0.0.0.0")
	p.viper.SetDefault("server.environment", "development")
	p.viper.SetDefault("server.read_timeout", 30)
	p.viper.SetDefault("server.write_timeout", 30)
	p.viper.SetDefault("server.idle_timeout", 120)

	// Database defaults
	p.viper.SetDefault("database.postgresql.host", "localhost")
	p.viper.SetDefault("database.postgresql.port", 5432)
	p.viper.SetDefault("database.postgresql.sslmode", "disable")
	p.viper.SetDefault("database.postgresql.max_conns", 25)
	p.viper.SetDefault("database.postgresql.min_conns", 5)

	p.viper.SetDefault("database.mongodb.uri", "mongodb://localhost:27017")
	p.viper.SetDefault("database.mongodb.max_pool", 100)
	p.viper.SetDefault("database.mongodb.min_pool", 10)

	// Redis defaults
	p.viper.SetDefault("redis.host", "localhost")
	p.viper.SetDefault("redis.port", 6379)
	p.viper.SetDefault("redis.db", 0)
	p.viper.SetDefault("redis.pool_size", 10)

	// Logging defaults
	p.viper.SetDefault("logging.level", "info")
	p.viper.SetDefault("logging.format", "json")
	p.viper.SetDefault("logging.output", "stdout")

	// Monitoring defaults
	p.viper.SetDefault("monitoring.prometheus.enabled", true)
	p.viper.SetDefault("monitoring.prometheus.port", "9090")
	p.viper.SetDefault("monitoring.prometheus.path", "/metrics")

	p.viper.SetDefault("monitoring.jaeger.enabled", false)

	// Storage defaults
	p.viper.SetDefault("storage.minio.endpoint", "localhost:9000")
	p.viper.SetDefault("storage.minio.use_ssl", false)

	// Search defaults
	p.viper.SetDefault("search.elasticsearch.url", "http://localhost:9200")

	// Auth defaults
	p.viper.SetDefault("auth.jwt.expiration", 3600)
	p.viper.SetDefault("auth.jwt.refresh_exp", 86400)
	p.viper.SetDefault("auth.jwt.issuer", "siakad")

	// RabbitMQ defaults
	p.viper.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")

	// Kafka defaults
	p.viper.SetDefault("kafka.brokers", []string{"localhost:9092"})

	// gRPC defaults
	p.viper.SetDefault("grpc.port", "50051")
	p.viper.SetDefault("grpc.host", "0.0.0.0")
	p.viper.SetDefault("grpc.timeout", 30)

	// Services defaults
	p.viper.SetDefault("services.user_service.name", "user-service")
	p.viper.SetDefault("services.user_service.host", "localhost")
	p.viper.SetDefault("services.user_service.port", "8001")
	p.viper.SetDefault("services.user_service.protocol", "http")
	p.viper.SetDefault("services.user_service.health_check.enabled", true)
	p.viper.SetDefault("services.user_service.health_check.path", "/health")
	p.viper.SetDefault("services.user_service.health_check.interval", 30)
	p.viper.SetDefault("services.user_service.retry.enabled", true)
	p.viper.SetDefault("services.user_service.retry.max_attempts", 3)
	p.viper.SetDefault("services.user_service.timeout.connect", 5000)
	p.viper.SetDefault("services.user_service.timeout.read", 10000)
	p.viper.SetDefault("services.user_service.circuit_breaker.enabled", true)
	p.viper.SetDefault("services.user_service.circuit_breaker.failure_threshold", 5)

	p.viper.SetDefault("services.auth_service.name", "auth-service")
	p.viper.SetDefault("services.auth_service.host", "localhost")
	p.viper.SetDefault("services.auth_service.port", "8002")
	p.viper.SetDefault("services.auth_service.protocol", "http")
	p.viper.SetDefault("services.auth_service.health_check.enabled", true)
	p.viper.SetDefault("services.auth_service.health_check.path", "/health")

	p.viper.SetDefault("services.payment_service.name", "payment-service")
	p.viper.SetDefault("services.payment_service.host", "localhost")
	p.viper.SetDefault("services.payment_service.port", "8003")
	p.viper.SetDefault("services.payment_service.protocol", "http")
	p.viper.SetDefault("services.payment_service.health_check.enabled", true)
	p.viper.SetDefault("services.payment_service.health_check.path", "/health")

	p.viper.SetDefault("services.notification_service.name", "notification-service")
	p.viper.SetDefault("services.notification_service.host", "localhost")
	p.viper.SetDefault("services.notification_service.port", "8004")
	p.viper.SetDefault("services.notification_service.protocol", "http")
	p.viper.SetDefault("services.notification_service.health_check.enabled", true)
	p.viper.SetDefault("services.notification_service.health_check.path", "/health")

	p.viper.SetDefault("services.file_service.name", "file-service")
	p.viper.SetDefault("services.file_service.host", "localhost")
	p.viper.SetDefault("services.file_service.port", "8005")
	p.viper.SetDefault("services.file_service.protocol", "http")
	p.viper.SetDefault("services.file_service.health_check.enabled", true)
	p.viper.SetDefault("services.file_service.health_check.path", "/health")

	p.viper.SetDefault("services.report_service.name", "report-service")
	p.viper.SetDefault("services.report_service.host", "localhost")
	p.viper.SetDefault("services.report_service.port", "8006")
	p.viper.SetDefault("services.report_service.protocol", "http")
	p.viper.SetDefault("services.report_service.health_check.enabled", true)
	p.viper.SetDefault("services.report_service.health_check.path", "/health")
}
