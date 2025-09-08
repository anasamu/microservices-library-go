package types

import (
	"fmt"
	"time"
)

// ConfigProvider represents a configuration provider interface
type ConfigProvider interface {
	Load() (*Config, error)
	Save(config *Config) error
	Watch(callback func(*Config)) error
	Close() error
}

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig           `mapstructure:"server"`
	Database   DatabaseConfig         `mapstructure:"database"`
	Redis      RedisConfig            `mapstructure:"redis"`
	Vault      VaultConfig            `mapstructure:"vault"`
	Logging    LoggingConfig          `mapstructure:"logging"`
	Monitoring MonitoringConfig       `mapstructure:"monitoring"`
	Storage    StorageConfig          `mapstructure:"storage"`
	Search     SearchConfig           `mapstructure:"search"`
	Auth       AuthConfig             `mapstructure:"auth"`
	RabbitMQ   RabbitMQConfig         `mapstructure:"rabbitmq"`
	Kafka      KafkaConfig            `mapstructure:"kafka"`
	GRPC       GRPCConfig             `mapstructure:"grpc"`
	Services   ServicesConfig         `mapstructure:"services"`
	Custom     map[string]interface{} `mapstructure:",remain"` // Dynamic custom configuration
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

// ServicesConfig holds microservices configuration
type ServicesConfig struct {
	UserService         ServiceConfig            `mapstructure:"user_service"`
	AuthService         ServiceConfig            `mapstructure:"auth_service"`
	PaymentService      ServiceConfig            `mapstructure:"payment_service"`
	NotificationService ServiceConfig            `mapstructure:"notification_service"`
	FileService         ServiceConfig            `mapstructure:"file_service"`
	ReportService       ServiceConfig            `mapstructure:"report_service"`
	CustomServices      map[string]ServiceConfig `mapstructure:"custom_services"` // Dynamic services
}

// ServiceConfig holds individual service configuration
type ServiceConfig struct {
	Name           string                 `mapstructure:"name"`
	Host           string                 `mapstructure:"host"`
	Port           string                 `mapstructure:"port"`
	Protocol       string                 `mapstructure:"protocol"` // http, grpc, tcp
	Version        string                 `mapstructure:"version"`
	Environment    string                 `mapstructure:"environment"`
	HealthCheck    HealthCheckConfig      `mapstructure:"health_check"`
	Retry          RetryConfig            `mapstructure:"retry"`
	Timeout        TimeoutConfig          `mapstructure:"timeout"`
	CircuitBreaker CircuitBreakerConfig   `mapstructure:"circuit_breaker"`
	LoadBalancer   LoadBalancerConfig     `mapstructure:"load_balancer"`
	Metadata       map[string]interface{} `mapstructure:"metadata"`
	Custom         map[string]interface{} `mapstructure:",remain"` // Dynamic service-specific config
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Path        string `mapstructure:"path"`
	Interval    int    `mapstructure:"interval"` // seconds
	Timeout     int    `mapstructure:"timeout"`  // seconds
	Retries     int    `mapstructure:"retries"`
	GracePeriod int    `mapstructure:"grace_period"` // seconds
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	MaxAttempts  int     `mapstructure:"max_attempts"`
	InitialDelay int     `mapstructure:"initial_delay"` // milliseconds
	MaxDelay     int     `mapstructure:"max_delay"`     // milliseconds
	Multiplier   float64 `mapstructure:"multiplier"`
	Jitter       bool    `mapstructure:"jitter"`
}

// TimeoutConfig holds timeout configuration
type TimeoutConfig struct {
	Connect int `mapstructure:"connect"` // milliseconds
	Read    int `mapstructure:"read"`    // milliseconds
	Write   int `mapstructure:"write"`   // milliseconds
	Total   int `mapstructure:"total"`   // milliseconds
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled               bool    `mapstructure:"enabled"`
	FailureThreshold      int     `mapstructure:"failure_threshold"`
	SuccessThreshold      int     `mapstructure:"success_threshold"`
	Timeout               int     `mapstructure:"timeout"` // seconds
	MaxRequests           int     `mapstructure:"max_requests"`
	Interval              int     `mapstructure:"interval"` // seconds
	ErrorPercentThreshold float64 `mapstructure:"error_percent_threshold"`
}

// LoadBalancerConfig holds load balancer configuration
type LoadBalancerConfig struct {
	Strategy    string   `mapstructure:"strategy"` // round_robin, least_conn, random, weighted
	Servers     []string `mapstructure:"servers"`
	Weights     []int    `mapstructure:"weights"`
	HealthCheck bool     `mapstructure:"health_check"`
}

// ConfigOptions represents configuration options
type ConfigOptions struct {
	Provider     string            `json:"provider"`
	ConfigPath   string            `json:"config_path"`
	Environment  string            `json:"environment"`
	WatchChanges bool              `json:"watch_changes"`
	SecretsPath  string            `json:"secrets_path"`
	Metadata     map[string]string `json:"metadata"`
}

// ConfigChangeEvent represents a configuration change event
type ConfigChangeEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Changes   map[string]interface{} `json:"changes"`
	Source    string                 `json:"source"`
	Metadata  map[string]interface{} `json:"metadata"`
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

// GetCustomValue returns a custom configuration value
func (c *Config) GetCustomValue(key string) (interface{}, bool) {
	if c.Custom == nil {
		return nil, false
	}
	value, exists := c.Custom[key]
	return value, exists
}

// SetCustomValue sets a custom configuration value
func (c *Config) SetCustomValue(key string, value interface{}) {
	if c.Custom == nil {
		c.Custom = make(map[string]interface{})
	}
	c.Custom[key] = value
}

// GetCustomString returns a custom configuration value as string
func (c *Config) GetCustomString(key string, defaultValue string) string {
	if value, exists := c.GetCustomValue(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetCustomInt returns a custom configuration value as int
func (c *Config) GetCustomInt(key string, defaultValue int) int {
	if value, exists := c.GetCustomValue(key); exists {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if intValue, err := fmt.Sscanf(v, "%d", &defaultValue); err == nil && intValue == 1 {
				return defaultValue
			}
		}
	}
	return defaultValue
}

// GetCustomBool returns a custom configuration value as bool
func (c *Config) GetCustomBool(key string, defaultValue bool) bool {
	if value, exists := c.GetCustomValue(key); exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return defaultValue
}

// GetServiceConfig returns configuration for a specific service
func (c *Config) GetServiceConfig(serviceName string) (*ServiceConfig, bool) {
	switch serviceName {
	case "user_service":
		return &c.Services.UserService, true
	case "auth_service":
		return &c.Services.AuthService, true
	case "payment_service":
		return &c.Services.PaymentService, true
	case "notification_service":
		return &c.Services.NotificationService, true
	case "file_service":
		return &c.Services.FileService, true
	case "report_service":
		return &c.Services.ReportService, true
	default:
		if c.Services.CustomServices != nil {
			if service, exists := c.Services.CustomServices[serviceName]; exists {
				return &service, true
			}
		}
		return nil, false
	}
}

// SetServiceConfig sets configuration for a specific service
func (c *Config) SetServiceConfig(serviceName string, serviceConfig ServiceConfig) {
	if c.Services.CustomServices == nil {
		c.Services.CustomServices = make(map[string]ServiceConfig)
	}
	c.Services.CustomServices[serviceName] = serviceConfig
}

// GetServiceURL returns the full URL for a service
func (c *Config) GetServiceURL(serviceName string) (string, error) {
	service, exists := c.GetServiceConfig(serviceName)
	if !exists {
		return "", fmt.Errorf("service %s not found", serviceName)
	}

	protocol := service.Protocol
	if protocol == "" {
		protocol = "http"
	}

	return fmt.Sprintf("%s://%s:%s", protocol, service.Host, service.Port), nil
}

// GetServiceHealthCheckURL returns the health check URL for a service
func (c *Config) GetServiceHealthCheckURL(serviceName string) (string, error) {
	service, exists := c.GetServiceConfig(serviceName)
	if !exists {
		return "", fmt.Errorf("service %s not found", serviceName)
	}

	baseURL, err := c.GetServiceURL(serviceName)
	if err != nil {
		return "", err
	}

	healthPath := service.HealthCheck.Path
	if healthPath == "" {
		healthPath = "/health"
	}

	return fmt.Sprintf("%s%s", baseURL, healthPath), nil
}

// GetAllServices returns all configured services
func (c *Config) GetAllServices() map[string]ServiceConfig {
	services := make(map[string]ServiceConfig)

	// Add predefined services
	services["user_service"] = c.Services.UserService
	services["auth_service"] = c.Services.AuthService
	services["payment_service"] = c.Services.PaymentService
	services["notification_service"] = c.Services.NotificationService
	services["file_service"] = c.Services.FileService
	services["report_service"] = c.Services.ReportService

	// Add custom services
	if c.Services.CustomServices != nil {
		for name, service := range c.Services.CustomServices {
			services[name] = service
		}
	}

	return services
}
