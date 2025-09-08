# SIAKAD Microservices Libraries - Examples

Contoh penggunaan library-library microservices SIAKAD untuk membantu developer memahami cara mengintegrasikan dan menggunakan library-library tersebut dalam aplikasi mereka.

## Struktur Examples

```
examples/
├── basic_service.go          # Contoh service dasar dengan semua library
├── event_driven_service.go   # Contoh service event-driven
├── go.mod                    # Go module untuk examples
└── README.md                 # Dokumentasi examples
```

## Contoh Penggunaan

### 1. Basic Service (`basic_service.go`)

Contoh service dasar yang menunjukkan cara menggunakan semua library utama:

- **Configuration Management**: Load config dari file YAML
- **Logging**: Structured logging dengan logrus
- **Database**: PostgreSQL dengan connection pooling
- **Caching**: Multi-backend cache (Redis, Memcache, Memory)
- **Authentication**: JWT token generation dan validation
- **Authorization**: RBAC role dan permission management
- **Monitoring**: Prometheus metrics collection
- **Utilities**: Helper functions untuk common operations

```go
// Contoh penggunaan
service, err := NewExampleService()
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
if err := service.Start(ctx); err != nil {
    log.Fatal(err)
}
```

### 2. Event-Driven Service (`event_driven_service.go`)

Contoh service event-driven yang menunjukkan:

- **Event Sourcing**: Event store dan outbox pattern
- **Message Queuing**: RabbitMQ dan Kafka integration
- **Distributed Tracing**: OpenTelemetry dengan Jaeger
- **Event Handlers**: Custom event handlers
- **Event Publishing**: Publish events ke multiple systems

```go
// Contoh penggunaan
service, err := NewEventDrivenService()
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
if err := service.Start(ctx); err != nil {
    log.Fatal(err)
}
```

## Cara Menjalankan Examples

### Prerequisites

1. **Docker & Docker Compose**: Untuk menjalankan dependencies
2. **Go 1.21+**: Untuk menjalankan aplikasi
3. **Dependencies**: PostgreSQL, Redis, RabbitMQ, Kafka, Jaeger

### Setup Dependencies

```bash
# Clone repository
git clone <repository-url>
cd backend/microservices/libs/examples

# Start dependencies dengan Docker Compose
docker-compose up -d

# Wait for services to be ready
sleep 30
```

### Menjalankan Basic Service

```bash
# Build dan run basic service
go run basic_service.go
```

### Menjalankan Event-Driven Service

```bash
# Build dan run event-driven service
go run event_driven_service.go
```

## Konfigurasi

### Basic Service Configuration

Buat file `config.yaml`:

```yaml
server:
  port: "8080"
  host: "0.0.0.0"
  environment: "development"
  service_name: "example-service"
  version: "1.0.0"

database:
  postgresql:
    host: "localhost"
    port: 5432
    user: "postgres"
    password: "password"
    dbname: "siakad"
    sslmode: "disable"
    max_conns: 25
    min_conns: 5

cache:
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
    pool_size: 10
    enabled: true
  memcache:
    servers: ["localhost:11211"]
    enabled: false
  memory:
    default_expiration: "5m"
    cleanup_interval: "10m"
    enabled: true
  default_ttl: "5m"
  enabled: true

auth:
  jwt:
    secret_key: "your-secret-key"
    expiration: 3600
    refresh_exp: 86400
    issuer: "siakad"
    audience: "siakad-users"

monitoring:
  prometheus:
    enabled: true
    port: "9090"
    path: "/metrics"
  jaeger:
    enabled: false
    endpoint: "http://localhost:14268/api/traces"
    service: "example-service"

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

### Event-Driven Service Configuration

Tambahkan konfigurasi untuk messaging dan tracing:

```yaml
rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"
  exchange: "siakad.events"
  queue: "user-events"

kafka:
  brokers: ["localhost:9092"]
  topic: "user-events"
  group_id: "event-driven-service"

tracing:
  service_name: "event-driven-service"
  service_version: "1.0.0"
  environment: "development"
  jaeger_endpoint: "http://localhost:14268/api/traces"
  sample_rate: 1.0
  enabled: true
```

## Docker Compose untuk Dependencies

Buat file `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: siakad
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  rabbitmq:
    image: rabbitmq:3-management
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq

  kafka:
    image: confluentinc/cp-kafka:latest
    environment:
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    ports:
      - "9092:9092"
    depends_on:
      - zookeeper

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"

  jaeger:
    image: jaegertracing/all-in-one:latest
    environment:
      COLLECTOR_OTLP_ENABLED: true
    ports:
      - "16686:16686"
      - "14268:14268"
      - "14250:14250"

volumes:
  postgres_data:
  redis_data:
  rabbitmq_data:
```

## Testing

### Unit Tests

```bash
# Run unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Integration Tests

```bash
# Run integration tests (requires dependencies)
go test -tags=integration ./...

# Run specific test
go test -run TestBasicService ./...
```

## Best Practices

### 1. Error Handling
- Gunakan structured errors
- Log errors dengan context
- Implement retry logic untuk transient errors
- Use circuit breakers untuk external services

### 2. Configuration
- Gunakan environment-specific config files
- Store secrets di Vault atau environment variables
- Validate configuration on startup
- Use sensible defaults

### 3. Logging
- Gunakan structured logging
- Include request ID dalam logs
- Log business events, bukan hanya technical events
- Use appropriate log levels

### 4. Monitoring
- Instrument semua operations
- Use consistent metric names
- Set up alerting untuk critical metrics
- Monitor business metrics

### 5. Testing
- Write unit tests untuk business logic
- Use mocks untuk external dependencies
- Write integration tests untuk critical paths
- Maintain high test coverage

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check PostgreSQL is running
   - Verify connection parameters
   - Check network connectivity

2. **Redis Connection Failed**
   - Check Redis is running
   - Verify Redis configuration
   - Check Redis memory usage

3. **RabbitMQ Connection Failed**
   - Check RabbitMQ is running
   - Verify credentials
   - Check exchange and queue configuration

4. **Kafka Connection Failed**
   - Check Kafka is running
   - Verify broker addresses
   - Check topic configuration

5. **Jaeger Connection Failed**
   - Check Jaeger is running
   - Verify endpoint configuration
   - Check network connectivity

### Debug Mode

Enable debug logging:

```yaml
logging:
  level: "debug"
  format: "json"
  output: "stdout"
```

### Health Checks

Check service health:

```bash
# Check basic service health
curl http://localhost:8080/health

# Check metrics
curl http://localhost:9090/metrics
```

## Contributing

1. Follow Go coding standards
2. Add comprehensive tests
3. Update documentation
4. Use semantic versioning
5. Add examples for new features

## License

This project is licensed under the MIT License - see the LICENSE file for details.
