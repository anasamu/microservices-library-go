package main

import (
	"context"
	"log"
	"time"

	"github.com/siakad/microservices/libs/auth"
	"github.com/siakad/microservices/libs/cache"
	"github.com/siakad/microservices/libs/config"
	"github.com/siakad/microservices/libs/database"
	"github.com/siakad/microservices/libs/logging"
	"github.com/siakad/microservices/libs/monitoring"
	"github.com/siakad/microservices/libs/utils"
)

// ExampleService demonstrates how to use the microservices libraries
type ExampleService struct {
	config       *config.Config
	logger       *logrus.Logger
	database     *database.PostgreSQL
	cacheManager *cache.CacheManager
	jwtManager   *auth.JWTManager
	rbacManager  *auth.RBACManager
	metrics      *monitoring.Metrics
}

// NewExampleService creates a new example service
func NewExampleService() (*ExampleService, error) {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		return nil, err
	}

	// Initialize logger
	logger := logging.NewLogger(cfg.Logging)

	// Initialize database
	db, err := database.NewPostgreSQL(&cfg.Database.PostgreSQL, logger)
	if err != nil {
		return nil, err
	}

	// Initialize cache
	cacheManager, err := cache.NewCacheManager(&cfg.Cache, logger)
	if err != nil {
		return nil, err
	}

	// Initialize auth
	jwtManager := auth.NewJWTManager(&cfg.Auth.JWT, logger)
	rbacManager := auth.NewRBACManager(logger)

	// Initialize monitoring
	metrics := monitoring.NewMetrics(&cfg.Monitoring, logger)
	metrics.Start()

	return &ExampleService{
		config:       cfg,
		logger:       logger,
		database:     db,
		cacheManager: cacheManager,
		jwtManager:   jwtManager,
		rbacManager:  rbacManager,
		metrics:      metrics,
	}, nil
}

// Start starts the example service
func (s *ExampleService) Start(ctx context.Context) error {
	s.logger.Info("Starting example service")

	// Example: Create a user
	userID := utils.GenerateUUID()
	tenantID := utils.GenerateUUID()
	email := "user@example.com"

	// Example: Generate JWT token
	tokenPair, err := s.jwtManager.GenerateTokenPair(
		ctx,
		userID,
		tenantID,
		email,
		[]string{"user"},
		[]string{"read", "write"},
		map[string]interface{}{
			"source": "example",
		},
	)
	if err != nil {
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   email,
	}).Info("Generated JWT token")

	// Example: Create RBAC role
	err = s.rbacManager.CreateRole(
		ctx,
		"user",
		"Regular user role",
		[]string{"read", "write"},
		map[string]interface{}{
			"created_by": "system",
		},
	)
	if err != nil {
		return err
	}

	// Example: Cache user data
	userData := map[string]interface{}{
		"id":         userID,
		"email":      email,
		"name":       "Example User",
		"created_at": time.Now(),
	}

	err = s.cacheManager.Set(ctx, cache.CreateUserCacheKey(userID.String()), userData, 5*time.Minute)
	if err != nil {
		return err
	}

	// Example: Retrieve from cache
	var cachedUserData map[string]interface{}
	err = s.cacheManager.Get(ctx, cache.CreateUserCacheKey(userID.String()), &cachedUserData)
	if err != nil {
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"cached_data": cachedUserData,
	}).Info("Retrieved user data from cache")

	// Example: Database operation
	err = s.database.Transaction(ctx, func(tx *sqlx.Tx) error {
		// Insert user
		query := `INSERT INTO users (id, email, name, created_at) VALUES ($1, $2, $3, $4)`
		_, err := tx.ExecContext(ctx, query, userID, email, "Example User", time.Now())
		if err != nil {
			return err
		}

		// Insert user role
		roleQuery := `INSERT INTO user_roles (user_id, role, created_at) VALUES ($1, $2, $3)`
		_, err = tx.ExecContext(ctx, roleQuery, userID, "user", time.Now())
		return err
	})
	if err != nil {
		return err
	}

	// Example: Record metrics
	s.metrics.RecordBusinessOperation("user_created", "user", "success", time.Since(time.Now()))

	s.logger.Info("Example service started successfully")
	return nil
}

// Stop stops the example service
func (s *ExampleService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping example service")

	// Close database connection
	if err := s.database.Close(); err != nil {
		s.logger.WithError(err).Error("Failed to close database connection")
	}

	// Close cache connection
	if err := s.cacheManager.Close(); err != nil {
		s.logger.WithError(err).Error("Failed to close cache connection")
	}

	s.logger.Info("Example service stopped")
	return nil
}

func main() {
	// Create example service
	service, err := NewExampleService()
	if err != nil {
		log.Fatal(err)
	}

	// Create context
	ctx := context.Background()

	// Start service
	if err := service.Start(ctx); err != nil {
		log.Fatal(err)
	}

	// Wait for interrupt signal
	<-ctx.Done()

	// Stop service
	if err := service.Stop(ctx); err != nil {
		log.Fatal(err)
	}
}
