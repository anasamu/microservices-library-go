# Microservices Authentication Library

A comprehensive, dynamic, and flexible authentication and authorization library for Go microservices.

## Features

- **Multiple Authentication Providers**: JWT, OAuth2, Two-Factor Authentication
- **Multiple Authorization Models**: RBAC, ABAC, ACL
- **Dynamic Configuration**: Runtime provider registration and configuration
- **Microservice-Ready**: Built for distributed systems with service identification
- **Comprehensive Logging**: Structured logging with context
- **Health Monitoring**: Built-in health checks and statistics
- **Type Safety**: Strong typing with comprehensive interfaces
- **Focused Design**: Authentication and authorization only, no user management

## Architecture

```
auth/
├── gateway/           # Main auth manager and gateway
├── types/            # Shared types and interfaces
├── providers/
│   ├── authentication/
│   │   ├── jwt/      # JWT authentication provider
│   │   ├── oauth/    # OAuth2 authentication provider
│   │   └── twofa/    # Two-Factor Authentication provider
│   └── authorization/
│       ├── rbac/     # Role-Based Access Control
│       ├── abac/     # Attribute-Based Access Control
│       └── acl/      # Access Control Lists
└── examples/         # Usage examples
```

## Quick Start

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/sirupsen/logrus"
    "github.com/anasamu/microservices-library-go/auth"
    "github.com/anasamu/microservices-library-go/auth/providers/authentication/jwt"
    "github.com/anasamu/microservices-library-go/auth/types"
)

func main() {
    // Initialize logger
    logger := logrus.New()
    
    // Create auth manager
    authManager := gateway.NewAuthManager(gateway.DefaultManagerConfig(), logger)
    
    // Register JWT provider
    jwtProvider := jwt.NewJWTProvider(jwt.DefaultJWTConfig(), logger)
    authManager.RegisterProvider(jwtProvider)
    
    // Authenticate user
    ctx := context.Background()
    authRequest := &types.AuthRequest{
        Username:  "user@example.com",
        Password:  "password",
        ServiceID: "user-service",
        Context:   map[string]interface{}{"ip": "192.168.1.1"},
    }
    
    response, err := authManager.Authenticate(ctx, "jwt", authRequest)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Authentication successful: %+v", response)
}
```


## Configuration

### Environment Variables

The library supports configuration through environment variables:

```bash
# JWT Configuration
export JWT_SECRET_KEY="your-secret-key"
export JWT_ACCESS_EXPIRY="15m"
export JWT_REFRESH_EXPIRY="168h"
export JWT_ISSUER="your-service"
export JWT_AUDIENCE="your-clients"

# OAuth Configuration
export GOOGLE_CLIENT_ID="your-google-client-id"
export GOOGLE_CLIENT_SECRET="your-google-client-secret"
```

### Dynamic Configuration

```go
// Configure JWT provider
jwtConfig := map[string]interface{}{
    "secret_key":     "your-secret-key",
    "access_expiry":  time.Minute * 15,
    "refresh_expiry": time.Hour * 24 * 7,
    "issuer":         "your-service",
    "audience":       "your-clients",
}
jwtProvider.Configure(jwtConfig)

// Configure OAuth provider
oauthConfig := map[string]interface{}{
    "google": map[string]interface{}{
        "client_id":     "your-google-client-id",
        "client_secret": "your-google-client-secret",
        "redirect_url":  "http://localhost:8080/auth/callback/google",
        "scopes":        []string{"openid", "profile", "email"},
    },
}
oauthProvider.Configure(oauthConfig)
```

## Authentication Providers

### JWT Provider

```go
jwtProvider := jwt.NewJWTProvider(jwt.DefaultJWTConfig(), logger)
authManager.RegisterProvider(jwtProvider)

// Generate token pair
tokenPair, err := jwtProvider.GenerateTokenPair(ctx, userID, email, roles, permissions, serviceID, context, metadata)

// Validate token
claims, err := jwtProvider.ValidateToken(ctx, tokenString)
```

### OAuth Provider

```go
oauthConfigs := map[string]*oauth.OAuthConfig{
    "google": {
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURL:  "http://localhost:8080/auth/callback/google",
        Scopes:       []string{"openid", "profile", "email"},
        Provider:     "google",
    },
}
oauthProvider := oauth.NewOAuthProvider(oauthConfigs, logger)
authManager.RegisterProvider(oauthProvider)

// Get authorization URL
authURL, err := oauthProvider.GetAuthURL(ctx, "google", state)

// Exchange code for tokens
tokenResponse, err := oauthProvider.ExchangeCode(ctx, "google", code)
```

### Two-Factor Authentication Provider

```go
twofaProvider := twofa.NewTwoFAProvider(twofa.DefaultTwoFAConfig(), logger)
authManager.RegisterProvider(twofaProvider)

// Generate 2FA secret
secret, err := twofaProvider.GenerateSecret(ctx, userID, userEmail)

// Verify 2FA code
verification, err := twofaProvider.VerifyCode(ctx, secret.Secret, code)
```

## Authorization Models

### Role-Based Access Control (RBAC)

```go
rbacManager := rbac.NewRBACManager(logger)

// Create role
role, err := rbacManager.CreateRole(ctx, "admin", "Administrator role", 
    []string{"user:read", "user:write", "admin:all"}, nil)

// Create permission
permission, err := rbacManager.CreatePermission(ctx, "user:read", "users", "read", 
    "Read user data", nil)

// Check access
accessRequest := &rbac.AccessRequest{
    Principal: "user123",
    Action:    "read",
    Resource:  "users",
    Context:   map[string]interface{}{"service": "user-service"},
}
decision, err := rbacManager.CheckAccess(ctx, accessRequest)
```

### Attribute-Based Access Control (ABAC)

```go
abacManager := abac.NewABACManager(logger)

// Create attribute
attribute, err := abacManager.CreateAttribute(ctx, "department", abac.AttributeTypeString, 
    abac.AttributeCategorySubject, "User department", nil, nil, false, "admin", nil)

// Create policy
policy, err := abacManager.CreatePolicy(ctx, "hr-policy", "HR access policy", 
    abac.ABACEffectAllow, rules, target, conditions, 100, "admin", nil)

// Check access
abacRequest := &abac.ABACRequest{
    Subject: map[string]interface{}{
        "id":         "user123",
        "department": "hr",
        "role":       "manager",
    },
    Resource: map[string]interface{}{
        "id":   "employee-data",
        "type": "sensitive",
    },
    Action: map[string]interface{}{
        "name": "read",
    },
    Environment: map[string]interface{}{
        "time": "business-hours",
    },
}
decision, err := abacManager.CheckAccess(ctx, abacRequest)
```

## Microservice Integration

### Service Identification

```go
authRequest := &types.AuthRequest{
    Username:  "user@example.com",
    Password:  "password",
    ServiceID: "user-service", // Identifies the requesting service
    Context: map[string]interface{}{
        "tenant_id": "tenant123",
        "region":    "us-east-1",
    },
    Metadata: map[string]interface{}{
        "version": "1.0.0",
        "build":   "abc123",
    },
}
```

### Multi-Tenant Support

```go
// Context can include tenant information
context := map[string]interface{}{
    "tenant_id": "tenant123",
    "organization": "acme-corp",
    "environment": "production",
}

// Service ID can be tenant-specific
serviceID := "user-service-tenant123"
```

## Monitoring and Health Checks

```go
// Health check
healthResults := authManager.HealthCheck(ctx)
for provider, err := range healthResults {
    if err != nil {
        log.Printf("Provider %s is unhealthy: %v", provider, err)
    }
}

// Get statistics
stats := authManager.GetStats(ctx)
log.Printf("Auth statistics: %+v", stats)
```

## Error Handling

The library provides comprehensive error handling:

```go
response, err := authManager.Authenticate(ctx, "jwt", authRequest)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "invalid credentials"):
        // Handle authentication failure
    case strings.Contains(err.Error(), "provider not found"):
        // Handle provider not found
    case strings.Contains(err.Error(), "token expired"):
        // Handle token expiration
    default:
        // Handle other errors
    }
}
```

## Security Best Practices

1. **Use Strong Secrets**: Generate cryptographically secure secrets for JWT signing
2. **Token Expiration**: Set appropriate token expiration times
3. **HTTPS Only**: Always use HTTPS in production
4. **Input Validation**: Validate all inputs before processing
5. **Rate Limiting**: Implement rate limiting for authentication endpoints
6. **Audit Logging**: Log all authentication and authorization events
7. **Regular Rotation**: Rotate secrets and keys regularly

## Examples

See the `examples/` directory for complete working examples:

- `microservice_example.go`: Complete microservice with authentication
- Authentication and authorization integration
- Multiple provider usage
- Error handling patterns

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
