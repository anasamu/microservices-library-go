module github.com/siakad/microservices/libs

go 1.21

require (
	// Authentication & Authorization
	github.com/golang-jwt/jwt/v5 v5.2.0
	golang.org/x/oauth2 v0.14.0
	golang.org/x/crypto v0.15.0

	// Database
	github.com/lib/pq v1.10.9
	github.com/jmoiron/sqlx v1.3.5
	go.mongodb.org/mongo-driver v1.13.1
	github.com/redis/go-redis/v9 v9.3.0

	// gRPC
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0

	// HTTP & Web
	github.com/gin-gonic/gin v1.9.1
	github.com/go-resty/resty/v2 v2.10.0
	github.com/hashicorp/go-retryablehttp v0.7.5

	// Message Queuing
	github.com/rabbitmq/amqp091-go v1.9.0
	github.com/segmentio/kafka-go v0.4.47

	// Caching
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/go-redis/cache/v9 v9.1.0

	// Service Discovery
	github.com/hashicorp/consul/api v1.25.1

	// Monitoring & Observability
	github.com/prometheus/client_golang v1.17.0
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.21.0

	// Resilience
	github.com/sony/gobreaker v0.5.0
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	golang.org/x/time v0.5.0

	// Search
	github.com/elastic/go-elasticsearch/v8 v8.11.1

	// Storage
	github.com/minio/minio-go/v7 v7.0.66

	// Configuration
	github.com/spf13/viper v1.17.0
	github.com/hashicorp/vault/api v1.10.0

	// Logging
	github.com/sirupsen/logrus v1.9.3

	// Utilities
	github.com/google/uuid v1.5.0
	github.com/go-playground/validator/v10 v10.16.0

	// Testing
	github.com/stretchr/testify v1.8.4
	github.com/testcontainers/testcontainers-go v0.26.0
	github.com/golang/mock v1.6.0
	github.com/DATA-DOG/go-sqlmock v1.5.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/containerd/containerd v1.7.7 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v24.0.6+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/fatih/color v1.14.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.9 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/vmihailenco/go-tinylfu v0.2.2 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231120223509-83a465c0220f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
