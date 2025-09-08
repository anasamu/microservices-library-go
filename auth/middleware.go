package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthMiddleware handles authentication and authorization for HTTP and gRPC
type AuthMiddleware struct {
	jwtManager   *JWTManager
	policyEngine *PolicyEngine
	logger       *logrus.Logger
	config       *MiddlewareConfig
}

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	// Token extraction
	TokenHeader     string `json:"token_header"`      // HTTP header name for token
	TokenQueryParam string `json:"token_query_param"` // Query parameter name for token
	TokenCookie     string `json:"token_cookie"`      // Cookie name for token
	BearerPrefix    string `json:"bearer_prefix"`     // Bearer token prefix

	// Authorization
	RequireAuth        bool     `json:"require_auth"`        // Require authentication
	RequireRoles       []string `json:"require_roles"`       // Required roles
	RequirePermissions []string `json:"require_permissions"` // Required permissions
	AllowAnonymous     bool     `json:"allow_anonymous"`     // Allow anonymous access

	// Policy evaluation
	PolicyEvaluationOrder *PolicyEvaluationOrder `json:"policy_evaluation_order"`

	// Error handling
	ErrorResponseFormat string                                                  `json:"error_response_format"` // JSON or HTML
	CustomErrorHandler  func(w http.ResponseWriter, r *http.Request, err error) `json:"-"`

	// Logging
	LogRequests  bool `json:"log_requests"`
	LogDecisions bool `json:"log_decisions"`
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		TokenHeader:           "Authorization",
		TokenQueryParam:       "token",
		TokenCookie:           "auth_token",
		BearerPrefix:          "Bearer ",
		RequireAuth:           true,
		RequireRoles:          []string{},
		RequirePermissions:    []string{},
		AllowAnonymous:        false,
		PolicyEvaluationOrder: DefaultPolicyEvaluationOrder(),
		ErrorResponseFormat:   "json",
		LogRequests:           true,
		LogDecisions:          true,
	}
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtManager *JWTManager, policyEngine *PolicyEngine, config *MiddlewareConfig, logger *logrus.Logger) *AuthMiddleware {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	return &AuthMiddleware{
		jwtManager:   jwtManager,
		policyEngine: policyEngine,
		logger:       logger,
		config:       config,
	}
}

// HTTPMiddleware returns HTTP middleware function
func (am *AuthMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Log request if enabled
		if am.config.LogRequests {
			am.logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			}).Debug("HTTP request received")
		}

		// Extract token
		token, err := am.extractTokenFromHTTP(r)
		if err != nil {
			if am.config.AllowAnonymous {
				// Allow anonymous access
				next.ServeHTTP(w, r)
				return
			}
			am.handleHTTPError(w, r, fmt.Errorf("token extraction failed: %w", err))
			return
		}

		// Validate token
		claims, err := am.jwtManager.ValidateToken(ctx, token)
		if err != nil {
			if am.config.AllowAnonymous {
				// Allow anonymous access
				next.ServeHTTP(w, r)
				return
			}
			am.handleHTTPError(w, r, fmt.Errorf("token validation failed: %w", err))
			return
		}

		// Add claims to context
		ctx = context.WithValue(ctx, "auth_claims", claims)
		ctx = context.WithValue(ctx, "user_id", claims.UserID.String())
		ctx = context.WithValue(ctx, "tenant_id", claims.TenantID.String())

		// Check authorization if required
		if am.config.RequireAuth {
			if err := am.checkHTTPAuthorization(ctx, r, claims); err != nil {
				am.handleHTTPError(w, r, err)
				return
			}
		}

		// Log decision if enabled
		if am.config.LogDecisions {
			am.logger.WithFields(logrus.Fields{
				"user_id":   claims.UserID,
				"tenant_id": claims.TenantID,
				"method":    r.Method,
				"path":      r.URL.Path,
				"decision":  "allow",
			}).Info("HTTP authorization decision")
		}

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GRPCUnaryInterceptor returns gRPC unary interceptor
func (am *AuthMiddleware) GRPCUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Log request if enabled
		if am.config.LogRequests {
			am.logger.WithFields(logrus.Fields{
				"method": info.FullMethod,
			}).Debug("gRPC request received")
		}

		// Extract token from metadata
		token, err := am.extractTokenFromGRPC(ctx)
		if err != nil {
			if am.config.AllowAnonymous {
				// Allow anonymous access
				return handler(ctx, req)
			}
			return nil, status.Errorf(codes.Unauthenticated, "token extraction failed: %v", err)
		}

		// Validate token
		claims, err := am.jwtManager.ValidateToken(ctx, token)
		if err != nil {
			if am.config.AllowAnonymous {
				// Allow anonymous access
				return handler(ctx, req)
			}
			return nil, status.Errorf(codes.Unauthenticated, "token validation failed: %v", err)
		}

		// Add claims to context
		ctx = context.WithValue(ctx, "auth_claims", claims)
		ctx = context.WithValue(ctx, "user_id", claims.UserID.String())
		ctx = context.WithValue(ctx, "tenant_id", claims.TenantID.String())

		// Check authorization if required
		if am.config.RequireAuth {
			if err := am.checkGRPCAuthorization(ctx, info, claims); err != nil {
				return nil, status.Errorf(codes.PermissionDenied, "authorization failed: %v", err)
			}
		}

		// Log decision if enabled
		if am.config.LogDecisions {
			am.logger.WithFields(logrus.Fields{
				"user_id":   claims.UserID,
				"tenant_id": claims.TenantID,
				"method":    info.FullMethod,
				"decision":  "allow",
			}).Info("gRPC authorization decision")
		}

		// Continue to handler
		return handler(ctx, req)
	}
}

// GRPCStreamInterceptor returns gRPC stream interceptor
func (am *AuthMiddleware) GRPCStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// Log request if enabled
		if am.config.LogRequests {
			am.logger.WithFields(logrus.Fields{
				"method": info.FullMethod,
			}).Debug("gRPC stream request received")
		}

		// Extract token from metadata
		token, err := am.extractTokenFromGRPC(ctx)
		if err != nil {
			if am.config.AllowAnonymous {
				// Allow anonymous access
				return handler(srv, ss)
			}
			return status.Errorf(codes.Unauthenticated, "token extraction failed: %v", err)
		}

		// Validate token
		claims, err := am.jwtManager.ValidateToken(ctx, token)
		if err != nil {
			if am.config.AllowAnonymous {
				// Allow anonymous access
				return handler(srv, ss)
			}
			return status.Errorf(codes.Unauthenticated, "token validation failed: %v", err)
		}

		// Add claims to context
		ctx = context.WithValue(ctx, "auth_claims", claims)
		ctx = context.WithValue(ctx, "user_id", claims.UserID.String())
		ctx = context.WithValue(ctx, "tenant_id", claims.TenantID.String())

		// Check authorization if required
		if am.config.RequireAuth {
			if err := am.checkGRPCAuthorization(ctx, &grpc.UnaryServerInfo{FullMethod: info.FullMethod}, claims); err != nil {
				return status.Errorf(codes.PermissionDenied, "authorization failed: %v", err)
			}
		}

		// Log decision if enabled
		if am.config.LogDecisions {
			am.logger.WithFields(logrus.Fields{
				"user_id":   claims.UserID,
				"tenant_id": claims.TenantID,
				"method":    info.FullMethod,
				"decision":  "allow",
			}).Info("gRPC stream authorization decision")
		}

		// Create new stream with updated context
		wrappedStream := &wrappedServerStream{ServerStream: ss, ctx: ctx}

		// Continue to handler
		return handler(srv, wrappedStream)
	}
}

// extractTokenFromHTTP extracts token from HTTP request
func (am *AuthMiddleware) extractTokenFromHTTP(r *http.Request) (string, error) {
	// Try header first
	if token := r.Header.Get(am.config.TokenHeader); token != "" {
		if strings.HasPrefix(token, am.config.BearerPrefix) {
			return strings.TrimPrefix(token, am.config.BearerPrefix), nil
		}
		return token, nil
	}

	// Try query parameter
	if token := r.URL.Query().Get(am.config.TokenQueryParam); token != "" {
		return token, nil
	}

	// Try cookie
	if cookie, err := r.Cookie(am.config.TokenCookie); err == nil {
		return cookie.Value, nil
	}

	return "", fmt.Errorf("no token found in request")
}

// extractTokenFromGRPC extracts token from gRPC metadata
func (am *AuthMiddleware) extractTokenFromGRPC(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata found")
	}

	// Try authorization header
	if values := md.Get("authorization"); len(values) > 0 {
		token := values[0]
		if strings.HasPrefix(token, am.config.BearerPrefix) {
			return strings.TrimPrefix(token, am.config.BearerPrefix), nil
		}
		return token, nil
	}

	// Try custom header
	if values := md.Get(am.config.TokenHeader); len(values) > 0 {
		return values[0], nil
	}

	return "", fmt.Errorf("no token found in metadata")
}

// checkHTTPAuthorization checks HTTP authorization
func (am *AuthMiddleware) checkHTTPAuthorization(ctx context.Context, r *http.Request, claims *JWTClaims) error {
	// Check required roles
	if len(am.config.RequireRoles) > 0 {
		if !claims.HasAnyRole(am.config.RequireRoles...) {
			return fmt.Errorf("insufficient roles: required %v, got %v", am.config.RequireRoles, claims.Roles)
		}
	}

	// Check required permissions
	if len(am.config.RequirePermissions) > 0 {
		if !claims.HasAnyPermission(am.config.RequirePermissions...) {
			return fmt.Errorf("insufficient permissions: required %v, got %v", am.config.RequirePermissions, claims.Permissions)
		}
	}

	// Use policy engine if available
	if am.policyEngine != nil {
		policyRequest := &PolicyRequest{
			UserID:   claims.UserID.String(),
			TenantID: claims.TenantID.String(),
			Resource: r.URL.Path,
			Action:   r.Method,
			Context: map[string]interface{}{
				"user_roles":       claims.Roles,
				"user_permissions": claims.Permissions,
				"ip_address":       r.RemoteAddr,
				"user_agent":       r.UserAgent(),
			},
		}

		decision, err := am.policyEngine.EvaluatePolicy(ctx, policyRequest, am.config.PolicyEvaluationOrder)
		if err != nil {
			return fmt.Errorf("policy evaluation failed: %w", err)
		}

		if decision.Decision == PolicyEngineEffectDeny {
			return fmt.Errorf("access denied: %s", decision.Reason)
		}
	}

	return nil
}

// checkGRPCAuthorization checks gRPC authorization
func (am *AuthMiddleware) checkGRPCAuthorization(ctx context.Context, info *grpc.UnaryServerInfo, claims *JWTClaims) error {
	// Check required roles
	if len(am.config.RequireRoles) > 0 {
		if !claims.HasAnyRole(am.config.RequireRoles...) {
			return fmt.Errorf("insufficient roles: required %v, got %v", am.config.RequireRoles, claims.Roles)
		}
	}

	// Check required permissions
	if len(am.config.RequirePermissions) > 0 {
		if !claims.HasAnyPermission(am.config.RequirePermissions...) {
			return fmt.Errorf("insufficient permissions: required %v, got %v", am.config.RequirePermissions, claims.Permissions)
		}
	}

	// Use policy engine if available
	if am.policyEngine != nil {
		policyRequest := &PolicyRequest{
			UserID:   claims.UserID.String(),
			TenantID: claims.TenantID.String(),
			Resource: info.FullMethod,
			Action:   "grpc_call",
			Context: map[string]interface{}{
				"user_roles":       claims.Roles,
				"user_permissions": claims.Permissions,
				"grpc_method":      info.FullMethod,
			},
		}

		decision, err := am.policyEngine.EvaluatePolicy(ctx, policyRequest, am.config.PolicyEvaluationOrder)
		if err != nil {
			return fmt.Errorf("policy evaluation failed: %w", err)
		}

		if decision.Decision == PolicyEngineEffectDeny {
			return fmt.Errorf("access denied: %s", decision.Reason)
		}
	}

	return nil
}

// handleHTTPError handles HTTP errors
func (am *AuthMiddleware) handleHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	if am.config.CustomErrorHandler != nil {
		am.config.CustomErrorHandler(w, r, err)
		return
	}

	am.logger.WithError(err).WithFields(logrus.Fields{
		"method":      r.Method,
		"path":        r.URL.Path,
		"remote_addr": r.RemoteAddr,
	}).Error("HTTP authentication/authorization error")

	w.WriteHeader(http.StatusUnauthorized)

	if am.config.ErrorResponseFormat == "json" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"error":"authentication_failed","message":"%s","timestamp":"%s"}`, err.Error(), time.Now().UTC().Format(time.RFC3339))
	} else {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><h1>Authentication Failed</h1><p>%s</p></body></html>`, err.Error())
	}
}

// wrappedServerStream wraps gRPC ServerStream with custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the custom context
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// GetClaimsFromContext extracts claims from context
func GetClaimsFromContext(ctx context.Context) (*JWTClaims, bool) {
	claims, ok := ctx.Value("auth_claims").(*JWTClaims)
	return claims, ok
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	return tenantID, ok
}
