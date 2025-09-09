package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anasamu/microservices-library-go/middleware/gateway"
	"github.com/anasamu/microservices-library-go/middleware/providers/auth"
	"github.com/anasamu/microservices-library-go/middleware/providers/communication"
	"github.com/anasamu/microservices-library-go/middleware/providers/logging"
	"github.com/anasamu/microservices-library-go/middleware/providers/messaging"
	"github.com/anasamu/microservices-library-go/middleware/providers/ratelimit"
	"github.com/anasamu/microservices-library-go/middleware/providers/storage"
	"github.com/anasamu/microservices-library-go/middleware/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareIntegration(t *testing.T) {
	logger := logrus.New()
	manager := gateway.NewMiddlewareManager(gateway.DefaultManagerConfig(), logger)

	// Register providers
	authProvider := auth.NewAuthProvider(auth.DefaultAuthConfig(), logger)
	err := manager.RegisterProvider(authProvider)
	require.NoError(t, err)

	loggingProvider := logging.NewLoggingProvider(logging.DefaultLoggingConfig(), logger)
	err = manager.RegisterProvider(loggingProvider)
	require.NoError(t, err)

	rateLimitProvider := ratelimit.NewRateLimitProvider(ratelimit.DefaultRateLimitConfig(), logger)
	err = manager.RegisterProvider(rateLimitProvider)
	require.NoError(t, err)

	storageProvider := storage.NewStorageProvider(storage.DefaultStorageConfig(), logger)
	err = manager.RegisterProvider(storageProvider)
	require.NoError(t, err)

	communicationProvider := communication.NewCommunicationProvider(communication.DefaultCommunicationConfig(), logger)
	err = manager.RegisterProvider(communicationProvider)
	require.NoError(t, err)

	messagingProvider := messaging.NewMessagingProvider(messaging.DefaultMessagingConfig(), logger)
	err = manager.RegisterProvider(messagingProvider)
	require.NoError(t, err)

	t.Run("HTTP Middleware Chain", func(t *testing.T) {
		// Create a test handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, World!"))
		})

		// Create middleware chain
		authMiddleware, err := manager.CreateHTTPMiddleware("auth", &types.MiddlewareConfig{
			Name:    "auth-middleware",
			Type:    "authentication",
			Enabled: true,
		})
		require.NoError(t, err)

		loggingMiddleware, err := manager.CreateHTTPMiddleware("logging", &types.MiddlewareConfig{
			Name:    "logging-middleware",
			Type:    "logging",
			Enabled: true,
		})
		require.NoError(t, err)

		rateLimitMiddleware, err := manager.CreateHTTPMiddleware("ratelimit", &types.MiddlewareConfig{
			Name:    "ratelimit-middleware",
			Type:    "rate_limiting",
			Enabled: true,
		})
		require.NoError(t, err)

		storageMiddleware, err := manager.CreateHTTPMiddleware("storage", &types.MiddlewareConfig{
			Name:    "storage-middleware",
			Type:    "storage",
			Enabled: true,
		})
		require.NoError(t, err)

		communicationMiddleware, err := manager.CreateHTTPMiddleware("communication", &types.MiddlewareConfig{
			Name:    "communication-middleware",
			Type:    "communication",
			Enabled: true,
		})
		require.NoError(t, err)

		messagingMiddleware, err := manager.CreateHTTPMiddleware("messaging", &types.MiddlewareConfig{
			Name:    "messaging-middleware",
			Type:    "messaging",
			Enabled: true,
		})
		require.NoError(t, err)

		// Chain middlewares
		wrappedHandler := authMiddleware(loggingMiddleware(rateLimitMiddleware(storageMiddleware(communicationMiddleware(messagingMiddleware(handler))))))

		// Test request without auth
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "missing authorization header")

		// Test request with auth
		req = httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		w = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Should still fail due to invalid token, but rate limiting should pass
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid token")
	})

	t.Run("Middleware Chain Execution", func(t *testing.T) {
		// Create middleware chain
		config := &types.MiddlewareConfig{
			Name:     "test-chain",
			Type:     "test",
			Enabled:  true,
			Priority: 1,
			Timeout:  30 * time.Second,
			Metadata: make(map[string]interface{}),
		}

		chain, err := manager.CreateChain(context.Background(), "logging", config)
		require.NoError(t, err)

		// Create test request
		request := &types.MiddlewareRequest{
			ID:        "test-request",
			Type:      "http",
			Method:    "GET",
			Path:      "/test",
			Headers:   map[string]string{"Content-Type": "application/json"},
			Query:     make(map[string]string),
			Context:   make(map[string]interface{}),
			Metadata:  make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		// Execute chain
		response, err := manager.ExecuteChain(context.Background(), "logging", chain, request)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.True(t, response.Success)
	})

	t.Run("Provider Health Check", func(t *testing.T) {
		results := manager.HealthCheck(context.Background())
		assert.NotNil(t, results)
		assert.Contains(t, results, "auth")
		assert.Contains(t, results, "logging")
		assert.Contains(t, results, "ratelimit")
		assert.Contains(t, results, "storage")
		assert.Contains(t, results, "communication")
		assert.Contains(t, results, "messaging")

		for provider, err := range results {
			assert.NoError(t, err, "Provider %s should be healthy", provider)
		}
	})

	t.Run("Provider Statistics", func(t *testing.T) {
		stats := manager.GetStats(context.Background())
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "auth")
		assert.Contains(t, stats, "logging")
		assert.Contains(t, stats, "ratelimit")
		assert.Contains(t, stats, "storage")
		assert.Contains(t, stats, "communication")
		assert.Contains(t, stats, "messaging")

		for provider, stat := range stats {
			assert.NotNil(t, stat, "Provider %s should have statistics", provider)
		}
	})

	t.Run("Provider Configuration", func(t *testing.T) {
		// Test auth provider configuration
		authConfig := map[string]interface{}{
			"jwt_secret_key": "new-secret-key",
			"jwt_expiry":     30 * time.Minute,
			"excluded_paths": []string{"/health", "/metrics", "/docs"},
		}

		err := authProvider.Configure(authConfig)
		assert.NoError(t, err)
		assert.True(t, authProvider.IsConfigured())

		// Test logging provider configuration
		loggingConfig := map[string]interface{}{
			"log_level":       "debug",
			"log_format":      "json",
			"include_headers": true,
			"include_body":    false,
		}

		err = loggingProvider.Configure(loggingConfig)
		assert.NoError(t, err)
		assert.True(t, loggingProvider.IsConfigured())

		// Test rate limit provider configuration
		rateLimitConfig := map[string]interface{}{
			"requests_per_second": 200,
			"burst_size":          400,
			"window_size":         2 * time.Minute,
		}

		err = rateLimitProvider.Configure(rateLimitConfig)
		assert.NoError(t, err)
		assert.True(t, rateLimitProvider.IsConfigured())

		// Test storage provider configuration
		storageConfig := map[string]interface{}{
			"storage_type":        "s3",
			"base_path":           "/uploads",
			"max_file_size":       int64(50 * 1024 * 1024), // 50MB
			"allowed_mime_types":  []string{"image/*", "application/pdf"},
			"cache_enabled":       true,
			"compression_enabled": true,
		}

		err = storageProvider.Configure(storageConfig)
		assert.NoError(t, err)
		assert.True(t, storageProvider.IsConfigured())

		// Test communication provider configuration
		communicationConfig := map[string]interface{}{
			"protocols":        []string{"http", "grpc", "websocket"},
			"default_protocol": "grpc",
			"timeout":          45 * time.Second,
			"retry_attempts":   5,
			"max_connections":  200,
			"compression":      true,
			"encryption":       true,
			"load_balancing":   true,
			"circuit_breaker":  true,
		}

		err = communicationProvider.Configure(communicationConfig)
		assert.NoError(t, err)
		assert.True(t, communicationProvider.IsConfigured())

		// Test messaging provider configuration
		messagingConfig := map[string]interface{}{
			"queues":            []string{"kafka", "nats", "rabbitmq"},
			"default_queue":     "kafka",
			"broker_url":        "localhost:9092",
			"topic_prefix":      "test-service",
			"group_id":          "test-group",
			"batch_size":        200,
			"compression":       true,
			"encryption":        false,
			"dead_letter_queue": true,
			"message_ttl":       2 * time.Hour,
		}

		err = messagingProvider.Configure(messagingConfig)
		assert.NoError(t, err)
		assert.True(t, messagingProvider.IsConfigured())
	})

	t.Run("Close Manager", func(t *testing.T) {
		err := manager.Close()
		assert.NoError(t, err)
	})
}
