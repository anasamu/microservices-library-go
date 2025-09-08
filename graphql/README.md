# GraphQL Library

Library GraphQL untuk microservices SIAKAD yang menyediakan server, resolvers, directives, dan utilities untuk implementasi GraphQL.

## Fitur

- **GraphQL Server**: Server GraphQL dengan middleware dan interceptors
- **Resolvers**: Registry dan builder untuk resolvers
- **Directives**: Custom directives untuk authentication, authorization, rate limiting, caching
- **Schema Builder**: Builder untuk schema GraphQL programmatically
- **Validation**: Input validation dan error handling
- **Pagination**: Support untuk pagination
- **Filtering**: Support untuk filtering dan sorting
- **Caching**: In-memory caching untuk resolvers
- **Multi-tenant**: Support untuk multi-tenant architecture
- **Logging**: Structured logging untuk semua operasi
- **Metrics**: Metrics collection untuk monitoring

## Instalasi

```bash
go get github.com/siakad/microservices/libs/graphql
```

## Penggunaan

### 1. Setup GraphQL Server

```go
package main

import (
    "github.com/siakad/microservices/libs/graphql"
    "github.com/sirupsen/logrus"
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
    
    // Create resolver (implement your schema)
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

### 2. Schema Definition

```go
// Define your GraphQL schema
type User struct {
    ID       string `json:"id" gql:"id"`
    Name     string `json:"name" gql:"name"`
    Email    string `json:"email" gql:"email"`
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

### 3. Resolver Implementation

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

### 4. Custom Directives

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

### 5. Schema dengan Directives

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

### 6. Pagination

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

### 7. Filtering dan Search

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

### 8. Caching

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

### 9. Error Handling

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

### 10. Multi-tenant Support

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

### 11. Middleware dan Interceptors

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

### 12. CORS Support

```go
r := gin.Default()

// Add CORS middleware
allowedOrigins := []string{"http://localhost:3000", "https://yourdomain.com"}
r.Use(graphql.CORSHandler(allowedOrigins))

r.POST("/graphql", server.GinHandler())
r.GET("/playground", server.PlaygroundHandler())
```

### 13. Health Check

```go
r := gin.Default()

// Add health check endpoint
r.GET("/health", graphql.HealthCheckHandler())

r.POST("/graphql", server.GinHandler())
r.GET("/playground", server.PlaygroundHandler())
```

## Konfigurasi

### Environment Variables

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

### Config File

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

## Testing

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

## Monitoring

Library ini menyediakan metrics untuk monitoring:

- **Operation Duration**: Durasi operasi GraphQL
- **Error Count**: Jumlah error per operasi
- **Cache Hit/Miss**: Cache performance
- **Rate Limit**: Rate limiting metrics
- **Authentication**: Authentication metrics

## Security

- **Authentication**: JWT token validation
- **Authorization**: Role dan permission-based access control
- **Rate Limiting**: Request rate limiting
- **Input Validation**: Input sanitization dan validation
- **CORS**: Cross-origin resource sharing
- **Tenant Isolation**: Multi-tenant data isolation

## Best Practices

1. **Use Directives**: Gunakan directives untuk cross-cutting concerns
2. **Implement Pagination**: Selalu implement pagination untuk list queries
3. **Add Caching**: Cache expensive operations
4. **Validate Input**: Validasi semua input
5. **Handle Errors**: Proper error handling dan logging
6. **Use Context**: Pass context untuk tracing dan logging
7. **Monitor Performance**: Monitor query complexity dan performance
8. **Secure Endpoints**: Implement proper authentication dan authorization

## Dependencies

- `github.com/99designs/gqlgen` - GraphQL library
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/sirupsen/logrus` - Logging
- `github.com/vektah/gqlparser/v2` - GraphQL parser

## License

MIT License
