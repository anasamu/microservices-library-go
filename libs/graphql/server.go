package graphql

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GraphQLServer represents a GraphQL server
type GraphQLServer struct {
	handler *handler.Server
	logger  *logrus.Logger
	config  *GraphQLConfig
}

// GraphQLConfig holds GraphQL server configuration
type GraphQLConfig struct {
	PlaygroundEnabled    bool
	IntrospectionEnabled bool
	ComplexityLimit      int
	QueryDepthLimit      int
	CacheSize            int
	Timeout              time.Duration
}

// NewGraphQLServer creates a new GraphQL server
func NewGraphQLServer(resolver graphql.ExecutableSchema, config *GraphQLConfig, logger *logrus.Logger) *GraphQLServer {
	// Set default config values
	if config == nil {
		config = &GraphQLConfig{
			PlaygroundEnabled:    true,
			IntrospectionEnabled: true,
			ComplexityLimit:      1000,
			QueryDepthLimit:      15,
			CacheSize:            1000,
			Timeout:              30 * time.Second,
		}
	}

	// Create GraphQL handler
	h := handler.New(resolver)

	// Configure transports
	h.AddTransport(transport.Options{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.POST{})

	// Configure extensions
	if config.IntrospectionEnabled {
		h.Use(extension.Introspection{})
	}

	h.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(config.CacheSize),
	})

	// Add middleware
	h.Use(extension.FixedComplexityLimit(config.ComplexityLimit))
	h.Use(extension.MaxDepth(config.QueryDepthLimit))

	// Add custom middleware
	h.Use(LoggingMiddleware(logger))
	h.Use(RecoveryMiddleware(logger))
	h.Use(TimeoutMiddleware(config.Timeout))

	return &GraphQLServer{
		handler: h,
		logger:  logger,
		config:  config,
	}
}

// ServeHTTP implements http.Handler
func (s *GraphQLServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// GinHandler returns a Gin handler for GraphQL
func (s *GraphQLServer) GinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		s.handler.ServeHTTP(c.Writer, c.Request)
	}
}

// PlaygroundHandler returns a GraphQL playground handler
func (s *GraphQLServer) PlaygroundHandler() gin.HandlerFunc {
	if !s.config.PlaygroundEnabled {
		return func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "GraphQL playground is disabled"})
		}
	}

	h := playground.Handler("GraphQL Playground", "/graphql")
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// LoggingMiddleware logs GraphQL operations
func LoggingMiddleware(logger *logrus.Logger) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return func(ctx context.Context) *graphql.Response {
			start := time.Now()

			// Get operation info
			rc := graphql.GetOperationContext(ctx)
			if rc != nil {
				logger.WithFields(logrus.Fields{
					"operation": rc.OperationName,
					"query":     rc.RawQuery,
					"variables": rc.Variables,
				}).Info("GraphQL operation started")
			}

			// Execute operation
			response := next(ctx)

			// Log completion
			duration := time.Since(start)
			if rc != nil {
				logger.WithFields(logrus.Fields{
					"operation": rc.OperationName,
					"duration":  duration,
					"errors":    len(response.Errors),
				}).Info("GraphQL operation completed")
			}

			return response
		}
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(logger *logrus.Logger) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return func(ctx context.Context) *graphql.Response {
			defer func() {
				if r := recover(); r != nil {
					logger.WithFields(logrus.Fields{
						"panic": r,
					}).Error("GraphQL operation panicked")
				}
			}()

			return next(ctx)
		}
	}
}

// TimeoutMiddleware adds timeout to operations
func TimeoutMiddleware(timeout time.Duration) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return func(ctx context.Context) *graphql.Response {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan *graphql.Response, 1)
			go func() {
				done <- next(ctx)
			}()

			select {
			case response := <-done:
				return response
			case <-ctx.Done():
				return &graphql.Response{
					Errors: []*graphql.Error{
						{
							Message: fmt.Sprintf("operation timeout after %v", timeout),
						},
					},
				}
			}
		}
	}
}

// AuthMiddleware adds authentication to GraphQL operations
func AuthMiddleware(logger *logrus.Logger) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return func(ctx context.Context) *graphql.Response {
			// Get request context
			rc := graphql.GetRequestContext(ctx)
			if rc == nil {
				return &graphql.Response{
					Errors: []*graphql.Error{
						{
							Message: "invalid request context",
						},
					},
				}
			}

			// Check for authorization header
			authHeader := rc.Header.Get("Authorization")
			if authHeader == "" {
				return &graphql.Response{
					Errors: []*graphql.Error{
						{
							Message: "authorization header required",
						},
					},
				}
			}

			// Add user info to context
			ctx = context.WithValue(ctx, "auth_token", authHeader)

			return next(ctx)
		}
	}
}

// TenantMiddleware adds tenant context to GraphQL operations
func TenantMiddleware(logger *logrus.Logger) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return func(ctx context.Context) *graphql.Response {
			// Get request context
			rc := graphql.GetRequestContext(ctx)
			if rc == nil {
				return &graphql.Response{
					Errors: []*graphql.Error{
						{
							Message: "invalid request context",
						},
					},
				}
			}

			// Check for tenant header
			tenantID := rc.Header.Get("X-Tenant-ID")
			if tenantID == "" {
				return &graphql.Response{
					Errors: []*graphql.Error{
						{
							Message: "tenant ID required",
						},
					},
				}
			}

			// Add tenant info to context
			ctx = context.WithValue(ctx, "tenant_id", tenantID)

			return next(ctx)
		}
	}
}

// RateLimitMiddleware adds rate limiting to GraphQL operations
func RateLimitMiddleware(limit int, window time.Duration, logger *logrus.Logger) graphql.OperationMiddleware {
	// Simple in-memory rate limiter (in production, use Redis)
	requests := make(map[string][]time.Time)

	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return func(ctx context.Context) *graphql.Response {
			// Get client IP
			rc := graphql.GetRequestContext(ctx)
			if rc == nil {
				return &graphql.Response{
					Errors: []*graphql.Error{
						{
							Message: "invalid request context",
						},
					},
				}
			}

			clientIP := rc.Header.Get("X-Forwarded-For")
			if clientIP == "" {
				clientIP = rc.Header.Get("X-Real-IP")
			}
			if clientIP == "" {
				clientIP = "unknown"
			}

			// Clean old requests
			now := time.Now()
			if clientRequests, exists := requests[clientIP]; exists {
				var validRequests []time.Time
				for _, reqTime := range clientRequests {
					if now.Sub(reqTime) < window {
						validRequests = append(validRequests, reqTime)
					}
				}
				requests[clientIP] = validRequests
			}

			// Check rate limit
			if len(requests[clientIP]) >= limit {
				logger.WithFields(logrus.Fields{
					"client_ip": clientIP,
					"limit":     limit,
					"window":    window,
				}).Warn("GraphQL rate limit exceeded")

				return &graphql.Response{
					Errors: []*graphql.Error{
						{
							Message: "rate limit exceeded",
						},
					},
				}
			}

			// Add current request
			requests[clientIP] = append(requests[clientIP], now)

			return next(ctx)
		}
	}
}

// MetricsMiddleware adds metrics collection to GraphQL operations
func MetricsMiddleware(logger *logrus.Logger) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return func(ctx context.Context) *graphql.Response {
			start := time.Now()

			// Get operation info
			rc := graphql.GetOperationContext(ctx)
			operationName := "unknown"
			if rc != nil {
				operationName = rc.OperationName
			}

			// Execute operation
			response := next(ctx)

			// Record metrics
			duration := time.Since(start)
			errorCount := len(response.Errors)

			logger.WithFields(logrus.Fields{
				"operation":   operationName,
				"duration_ms": duration.Milliseconds(),
				"error_count": errorCount,
				"success":     errorCount == 0,
			}).Info("GraphQL metrics")

			return response
		}
	}
}

// ValidationMiddleware adds input validation to GraphQL operations
func ValidationMiddleware(logger *logrus.Logger) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return func(ctx context.Context) *graphql.Response {
			// Get operation context
			rc := graphql.GetOperationContext(ctx)
			if rc == nil {
				return &graphql.Response{
					Errors: []*graphql.Error{
						{
							Message: "invalid operation context",
						},
					},
				}
			}

			// Validate query
			if rc.RawQuery == "" {
				return &graphql.Response{
					Errors: []*graphql.Error{
						{
							Message: "query cannot be empty",
						},
					},
				}
			}

			// Validate operation name for mutations
			if rc.Operation != nil && rc.Operation.Operation == "mutation" && rc.OperationName == "" {
				logger.Warn("Mutation without operation name")
			}

			return next(ctx)
		}
	}
}

// CORSHandler adds CORS headers to GraphQL responses
func CORSHandler(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// HealthCheckHandler returns a health check handler for GraphQL
func HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "graphql",
			"timestamp": time.Now().UTC(),
		})
	}
}
