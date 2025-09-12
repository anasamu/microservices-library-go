# Auth Library Examples

This directory contains comprehensive examples demonstrating how to use the Auth library in different scenarios.

## Examples Overview

### 1. Simple Example (`simple_example.go`)
A basic example showing the core functionality of the auth library:
- Creating users
- Authenticating users
- Token validation
- Permission checking
- Health checks and statistics

**Run:**
```bash
go run simple_example.go
```

### 2. HTTP Server Example (`http_server_example.go`)
A complete HTTP server implementation with REST API endpoints:
- `/health` - Health check endpoint
- `/register` - User registration (simplified)
- `/login` - User authentication
- `/validate` - Token validation
- `/protected` - Protected resource access

**Run:**
```bash
go run http_server_example.go
```

**Test the API:**
```bash
# Register a user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "secure_password",
    "first_name": "John",
    "last_name": "Doe",
    "roles": ["user"],
    "permissions": ["read", "write"]
  }'

# Login
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "secure_password",
    "service_id": "user-service"
  }'

# Access protected resource (replace TOKEN with actual token)
curl -X GET http://localhost:8080/protected \
  -H "Authorization: Bearer TOKEN"
```

### 3. OAuth Example (`oauth_example.go`)
OAuth integration with Google and GitHub:
- OAuth initiation
- OAuth callback handling
- Provider-specific configurations

**Run:**
```bash
go run oauth_example.go
```

**Test OAuth:**
```bash
# Initiate Google OAuth
curl http://localhost:8080/auth/google

# Initiate GitHub OAuth
curl http://localhost:8080/auth/github
```

### 4. Two-Factor Authentication Example (`twofa_example.go`)
Complete 2FA implementation:
- 2FA setup with QR code generation
- 2FA verification
- 2FA disable functionality
- Login with 2FA

**Run:**
```bash
go run twofa_example.go
```

**Test 2FA:**
```bash
# Setup 2FA for a user
curl -X POST http://localhost:8080/setup-2fa \
  -d "user_id=john_doe"

# Verify 2FA setup
curl -X POST http://localhost:8080/verify-2fa \
  -d "user_id=john_doe&code=123456"

# Login with 2FA
curl -X POST http://localhost:8080/login-with-2fa \
  -d "username=john_doe&password=secure_password&code=123456"
```

### 5. HTTP Server Example (`http_server_example.go`)
A complete HTTP server implementation with REST API endpoints:
- `/health` - Health check endpoint
- `/register` - User registration (simplified)
- `/login` - User authentication
- `/validate` - Token validation
- `/protected` - Protected resource access

**Run:**
```bash
go run http_server_example.go
```

## Configuration

### Environment Variables

Set these environment variables for production use:

```bash
# JWT Configuration
export JWT_SECRET="your-super-secret-jwt-key"
export JWT_EXPIRATION="24h"

# OAuth Configuration
export GOOGLE_CLIENT_ID="your-google-client-id"
export GOOGLE_CLIENT_SECRET="your-google-client-secret"
export GITHUB_CLIENT_ID="your-github-client-id"
export GITHUB_CLIENT_SECRET="your-github-client-secret"

# 2FA Configuration
export TOTP_ISSUER="YourApp"
export TOTP_ALGORITHM="SHA1"
```

### Provider Configuration

Each example shows how to configure different authentication providers:

#### JWT Provider
```go
jwtConfig := &jwt.JWTConfig{
    Secret:     "your-secret-key",
    Expiration: 24 * time.Hour,
    Issuer:     "your-app",
    Algorithm:  "HS256",
}
```

#### OAuth Provider
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
```

#### 2FA Provider
```go
twofaConfig := &twofa.TwoFAConfig{
    Issuer:    "YourApp",
    Algorithm: "SHA1",
    Digits:    6,
    Period:    30,
}
```

## Error Handling

All examples include proper error handling:

```go
response, err := authManager.Authenticate(ctx, "jwt", authRequest)
if err != nil {
    log.Printf("Authentication failed: %v", err)
    return
}
```

## Security Best Practices

1. **Use HTTPS in production** - Never send credentials over HTTP
2. **Validate all inputs** - Sanitize and validate user inputs
3. **Use secure secrets** - Generate strong, random secrets for JWT
4. **Implement rate limiting** - Prevent brute force attacks
5. **Log security events** - Monitor authentication attempts
6. **Use secure cookies** - Set appropriate cookie flags
7. **Implement CSRF protection** - Protect against cross-site request forgery

## Testing

Run the examples and test the functionality:

```bash
# Test simple example
go run simple_example.go

# Test HTTP server
go run http_server_example.go &
curl http://localhost:8080/health

# Test OAuth (requires valid OAuth credentials)
go run oauth_example.go

# Test 2FA
go run twofa_example.go
```

## Integration with Other Libraries

The auth library integrates well with other microservices libraries:

```go
import (
    "github.com/anasamu/microservices-library-go/auth"
    "github.com/anasamu/microservices-library-go/cache"
    "github.com/anasamu/microservices-library-go/ratelimit"
)

// Use with cache for session management
cacheManager := cache.NewCacheManager(nil, logger)
authManager := auth.NewAuthManager(auth.DefaultManagerConfig(), logger)

// Use with rate limiting
rateLimitManager := ratelimit.NewRateLimitManager(nil, logger)
```

## Troubleshooting

### Common Issues

1. **"Provider not found" error**
   - Make sure to register the provider before using it
   - Check provider name matches exactly

2. **"Invalid token" error**
   - Verify JWT secret is consistent
   - Check token expiration
   - Ensure proper token format

3. **OAuth callback errors**
   - Verify redirect URLs match OAuth provider configuration
   - Check client ID and secret
   - Ensure proper scopes are requested

4. **2FA setup issues**
   - Verify TOTP configuration
   - Check QR code generation
   - Ensure proper time synchronization

### Debug Mode

Enable debug logging:

```go
logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)
```

## Contributing

When adding new examples:

1. Follow the existing code structure
2. Include proper error handling
3. Add comprehensive comments
4. Update this README
5. Test thoroughly

## License

These examples are part of the microservices library and follow the same MIT license.
