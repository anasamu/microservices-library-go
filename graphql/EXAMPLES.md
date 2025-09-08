# GraphQL Library Examples

Contoh penggunaan library GraphQL untuk microservices SIAKAD.

## 1. Basic GraphQL Server Setup

```go
package main

import (
    "github.com/siakad/microservices/libs/graphql"
    "github.com/sirupsen/logrus"
    "github.com/gin-gonic/gin"
)

func main() {
    logger := logrus.New()
    
    // Create GraphQL config
    config := &graphql.GraphQLConfig{
        PlaygroundEnabled:   true,
        IntrospectionEnabled: true,
        ComplexityLimit:     1000,
        QueryDepthLimit:     15,
        CacheSize:           1000,
        Timeout:             30 * time.Second,
    }
    
    // Create resolver
    resolver := &YourResolver{}
    
    // Create GraphQL server
    server := graphql.NewGraphQLServer(resolver, config, logger)
    
    // Setup Gin routes
    r := gin.Default()
    r.POST("/graphql", server.GinHandler())
    r.GET("/playground", server.PlaygroundHandler())
    
    r.Run(":8080")
}
```

## 2. Schema Definition

```go
// Define your GraphQL schema
type User struct {
    ID       string    `json:"id" gql:"id"`
    Name     string    `json:"name" gql:"name"`
    Email    string    `json:"email" gql:"email"`
    CreatedAt time.Time `json:"createdAt" gql:"createdAt"`
}

type Query struct {
    Users []User `json:"users" gql:"users"`
    User  User   `json:"user" gql:"user"`
}

type Mutation struct {
    CreateUser User `json:"createUser" gql:"createUser"`
    UpdateUser User `json:"updateUser" gql:"updateUser"`
    DeleteUser bool `json:"deleteUser" gql:"deleteUser"`
}
```

## 3. Resolver Implementation

```go
type UserResolver struct {
    *graphql.BaseResolver
    userService UserService
}

func NewUserResolver(logger *logrus.Logger, userService UserService) *UserResolver {
    return &UserResolver{
        BaseResolver: graphql.NewBaseResolver(logger),
        userService:  userService,
    }
}

func (r *UserResolver) Users(ctx context.Context) ([]User, error) {
    r.LogOperation(ctx, "users", nil)
    
    users, err := r.userService.GetUsers(ctx)
    if err != nil {
        return nil, r.HandleError(ctx, "users", err)
    }
    
    return users, nil
}

func (r *UserResolver) User(ctx context.Context, id string) (*User, error) {
    r.LogOperation(ctx, "user", map[string]interface{}{"id": id})
    
    user, err := r.userService.GetUser(ctx, id)
    if err != nil {
        return nil, r.HandleError(ctx, "user", err)
    }
    
    return user, nil
}

func (r *UserResolver) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
    r.LogOperation(ctx, "createUser", input)
    
    // Validate input
    if err := r.ValidateInput(ctx, input); err != nil {
        return nil, r.HandleError(ctx, "createUser", err)
    }
    
    user, err := r.userService.CreateUser(ctx, input)
    if err != nil {
        return nil, r.HandleError(ctx, "createUser", err)
    }
    
    return user, nil
}
```

## 4. Custom Directives

```go
// Register custom directives
directiveRegistry := graphql.NewDirectiveRegistry(logger)

// Auth directive
directiveRegistry.Register("auth", graphql.AuthDirective(logger))

// Role-based access control
directiveRegistry.Register("hasRole", graphql.HasRoleDirective(logger))

// Permission-based access control
directiveRegistry.Register("hasPermission", graphql.HasPermissionDirective(logger))

// Rate limiting
directiveRegistry.Register("rateLimit", graphql.RateLimitDirective(logger))

// Caching
directiveRegistry.Register("cache", graphql.CacheDirective(logger))

// Tenant isolation
directiveRegistry.Register("tenant", graphql.TenantDirective(logger))
```

## 5. Schema dengan Directives

```graphql
type Query {
  users: [User!]! @auth @cache(maxAge: 300)
  user(id: ID!): User @auth @hasRole(role: "admin")
}

type Mutation {
  createUser(input: CreateUserInput!): User! @auth @hasPermission(permission: "user.create")
  updateUser(id: ID!, input: UpdateUserInput!): User! @auth @hasPermission(permission: "user.update")
  deleteUser(id: ID!): Boolean! @auth @hasPermission(permission: "user.delete")
}

type User {
  id: ID!
  name: String!
  email: String!
  createdAt: Time!
}
```

## 6. Pagination

```go
type UserResolver struct {
    *graphql.PaginationResolver
    userService UserService
}

func (r *UserResolver) Users(ctx context.Context, pagination *graphql.PaginationInput) (*graphql.PaginatedResult, error) {
    // Validate pagination input
    if err := r.ValidateInput(ctx, pagination); err != nil {
        return nil, r.HandleError(ctx, "users", err)
    }
    
    // Get users with pagination
    users, total, err := r.userService.GetUsersPaginated(ctx, pagination.Page, pagination.PageSize)
    if err != nil {
        return nil, r.HandleError(ctx, "users", err)
    }
    
    // Calculate pagination info
    paginationInfo := r.CalculatePagination(pagination.Page, pagination.PageSize, total)
    
    return &graphql.PaginatedResult{
        Data:       users,
        Pagination: paginationInfo,
    }, nil
}
```

## 7. Filtering dan Search

```go
type UserResolver struct {
    *graphql.FilterResolver
    userService UserService
}

func (r *UserResolver) SearchUsers(ctx context.Context, search *graphql.SearchInput) (*graphql.PaginatedResult, error) {
    // Validate search input
    if err := r.ValidateInput(ctx, search); err != nil {
        return nil, r.HandleError(ctx, "searchUsers", err)
    }
    
    // Apply search
    users, total, err := r.userService.SearchUsers(ctx, search)
    if err != nil {
        return nil, r.HandleError(ctx, "searchUsers", err)
    }
    
    // Calculate pagination
    paginationInfo := r.CalculatePagination(
        search.Pagination.Page,
        search.Pagination.PageSize,
        total,
    )
    
    return &graphql.PaginatedResult{
        Data:       users,
        Pagination: paginationInfo,
    }, nil
}
```

## 8. Caching

```go
type UserResolver struct {
    *graphql.CacheResolver
    userService UserService
}

func (r *UserResolver) User(ctx context.Context, id string) (*User, error) {
    cacheKey := fmt.Sprintf("user:%s", id)
    
    return r.ResolveWithCache(ctx, cacheKey, func() (interface{}, error) {
        return r.userService.GetUser(ctx, id)
    }).(*User), nil
}
```

## 9. Error Handling

```go
type UserResolver struct {
    *graphql.BaseResolver
    errorHandler *graphql.ErrorHandler
    userService  UserService
}

func (r *UserResolver) User(ctx context.Context, id string) (*User, error) {
    user, err := r.userService.GetUser(ctx, id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return nil, r.errorHandler.HandleNotFoundError(ctx, "User")
        }
        if errors.Is(err, ErrUnauthorized) {
            return nil, r.errorHandler.HandleUnauthorizedError(ctx)
        }
        return nil, r.errorHandler.Handle(ctx, err)
    }
    
    return user, nil
}
```

## 10. Multi-tenant Support

```go
func (r *UserResolver) Users(ctx context.Context) ([]User, error) {
    // Get tenant from context
    tenant, exists := graphql.GetTenantFromContext(ctx)
    if !exists {
        return nil, r.errorHandler.HandleUnauthorizedError(ctx)
    }
    
    // Get users for specific tenant
    users, err := r.userService.GetUsersByTenant(ctx, tenant)
    if err != nil {
        return nil, r.HandleError(ctx, "users", err)
    }
    
    return users, nil
}
```

## 11. Subscriptions

```go
func (r *UserResolver) UserCreated(ctx context.Context) (<-chan *User, error) {
    // Create a channel for user creation events
    userChan := make(chan *User, 1)
    
    // Subscribe to user creation events
    go func() {
        defer close(userChan)
        
        // Listen for events from event stream
        for event := range r.userService.UserCreatedEvents() {
            select {
            case userChan <- event.User:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return userChan, nil
}
```

## 12. Middleware dan Interceptors

```go
// Add custom middleware
server := graphql.NewGraphQLServer(resolver, config, logger)

// Add authentication middleware
server.handler.Use(graphql.AuthMiddleware(logger))

// Add tenant middleware
server.handler.Use(graphql.TenantMiddleware(logger))

// Add rate limiting middleware
server.handler.Use(graphql.RateLimitMiddleware(100, time.Minute, logger))

// Add metrics middleware
server.handler.Use(graphql.MetricsMiddleware(logger))

// Add validation middleware
server.handler.Use(graphql.ValidationMiddleware(logger))
```

## 13. CORS Support

```go
r := gin.Default()

// Add CORS middleware
allowedOrigins := []string{"http://localhost:3000", "https://yourdomain.com"}
r.Use(graphql.CORSHandler(allowedOrigins))

r.POST("/graphql", server.GinHandler())
r.GET("/playground", server.PlaygroundHandler())
```

## 14. Health Check

```go
r := gin.Default()

// Add health check endpoint
r.GET("/health", graphql.HealthCheckHandler())

r.POST("/graphql", server.GinHandler())
r.GET("/playground", server.PlaygroundHandler())
```

## 15. Testing

```go
func TestUserResolver(t *testing.T) {
    logger := logrus.New()
    userService := &MockUserService{}
    resolver := NewUserResolver(logger, userService)
    
    // Test query
    ctx := context.Background()
    users, err := resolver.Users(ctx)
    
    assert.NoError(t, err)
    assert.NotNil(t, users)
}

func TestAuthDirective(t *testing.T) {
    logger := logrus.New()
    directive := graphql.AuthDirective(logger)
    
    // Test with authenticated user
    ctx := graphql.SetUserInContext(context.Background(), &User{ID: "1"})
    
    result, err := directive(ctx, func(ctx context.Context) (interface{}, error) {
        return "success", nil
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "success", result)
}
```

## 16. Configuration

```yaml
graphql:
  playground_enabled: true
  introspection_enabled: true
  complexity_limit: 1000
  query_depth_limit: 15
  cache_size: 1000
  timeout: 30s

cors:
  allowed_origins:
    - "http://localhost:3000"
    - "https://yourdomain.com"

rate_limit:
  requests: 100
  window: "1m"
```

## 17. Environment Variables

```bash
# GraphQL Configuration
GRAPHQL_PLAYGROUND_ENABLED=true
GRAPHQL_INTROSPECTION_ENABLED=true
GRAPHQL_COMPLEXITY_LIMIT=1000
GRAPHQL_QUERY_DEPTH_LIMIT=15
GRAPHQL_CACHE_SIZE=1000
GRAPHQL_TIMEOUT=30s

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

## 18. Monitoring

```go
// Metrics are automatically collected
// You can access them via Prometheus endpoint

// Custom metrics
func (r *UserResolver) Users(ctx context.Context) ([]User, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        // Record custom metric
        userQueryDuration.WithLabelValues("users").Observe(duration.Seconds())
    }()
    
    // ... resolver logic
}
```

## 19. Security Best Practices

```go
// Always validate input
func (r *UserResolver) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
    // Validate input
    if err := r.ValidateInput(ctx, input); err != nil {
        return nil, r.errorHandler.HandleValidationError(ctx, err)
    }
    
    // Check permissions
    if !r.hasPermission(ctx, "user.create") {
        return nil, r.errorHandler.HandleForbiddenError(ctx)
    }
    
    // Sanitize input
    input.Email = strings.ToLower(strings.TrimSpace(input.Email))
    
    // ... create user logic
}
```

## 20. Performance Optimization

```go
// Use caching for expensive operations
func (r *UserResolver) User(ctx context.Context, id string) (*User, error) {
    cacheKey := fmt.Sprintf("user:%s", id)
    
    return r.ResolveWithCache(ctx, cacheKey, func() (interface{}, error) {
        return r.userService.GetUser(ctx, id)
    }).(*User), nil
}

// Use pagination for large datasets
func (r *UserResolver) Users(ctx context.Context, pagination *graphql.PaginationInput) (*graphql.PaginatedResult, error) {
    // Always use pagination for list queries
    if pagination == nil {
        pagination = &graphql.PaginationInput{
            Page:     1,
            PageSize: 10,
        }
    }
    
    // ... pagination logic
}
```
