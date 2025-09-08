package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ExampleUsage demonstrates how to use the dynamic authentication system
func ExampleUsage() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration from environment variables
	config := LoadAuthConfigFromEnv()

	// Create authentication manager
	authManager, err := NewAuthManager(config, logger)
	if err != nil {
		log.Fatalf("Failed to create auth manager: %v", err)
	}
	defer authManager.Shutdown(context.Background())

	// Example 1: JWT Token Generation and Validation
	exampleJWTUsage(authManager)

	// Example 2: RBAC Usage
	exampleRBACUsage(authManager)

	// Example 3: ACL Usage
	exampleACLUsage(authManager)

	// Example 4: ABAC Usage
	exampleABACUsage(authManager)

	// Example 5: Policy Engine Usage
	examplePolicyEngineUsage(authManager)

	// Example 6: HTTP Middleware Usage
	exampleHTTPMiddlewareUsage(authManager)

	// Example 7: OAuth2 Usage
	exampleOAuth2Usage(authManager)
}

// exampleJWTUsage demonstrates JWT token operations
func exampleJWTUsage(am *AuthManager) {
	fmt.Println("=== JWT Usage Example ===")

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	email := "user@example.com"
	roles := []string{"user", "admin"}
	permissions := []string{"read:all", "write:own", "admin:all"}
	metadata := map[string]interface{}{
		"department": "engineering",
		"location":   "remote",
	}

	// Generate token pair
	tokenPair, err := am.AuthenticateUser(ctx, userID, tenantID, email, roles, permissions, metadata)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		return
	}

	fmt.Printf("Generated access token: %s\n", tokenPair.AccessToken[:50]+"...")
	fmt.Printf("Token expires at: %s\n", tokenPair.ExpiresAt)

	// Validate token
	claims, err := am.ValidateUserToken(ctx, tokenPair.AccessToken)
	if err != nil {
		log.Printf("Failed to validate token: %v", err)
		return
	}

	fmt.Printf("Token validated for user: %s\n", claims.Email)
	fmt.Printf("User roles: %v\n", claims.Roles)
	fmt.Printf("User permissions: %v\n", claims.Permissions)
}

// exampleRBACUsage demonstrates RBAC operations
func exampleRBACUsage(am *AuthManager) {
	fmt.Println("\n=== RBAC Usage Example ===")

	ctx := context.Background()
	rbacManager := am.GetRBACManager()
	if rbacManager == nil {
		fmt.Println("RBAC manager not available")
		return
	}

	// Create custom role
	role, err := rbacManager.CreateRole(ctx, "developer", "Software developer role", []string{"read:own", "write:own", "deploy:staging"}, nil)
	if err != nil {
		log.Printf("Failed to create role: %v", err)
		return
	}

	fmt.Printf("Created role: %s\n", role.Name)

	// Create custom permission
	permission, err := rbacManager.CreatePermission(ctx, "deploy:staging", "deployment", "deploy", "Deploy to staging environment", nil)
	if err != nil {
		log.Printf("Failed to create permission: %v", err)
		return
	}

	fmt.Printf("Created permission: %s\n", permission.Name)

	// Check access
	accessRequest := &AccessRequest{
		Principal: "developer",
		Action:    "deploy",
		Resource:  "staging",
		Context:   map[string]interface{}{},
	}

	decision, err := rbacManager.CheckAccess(ctx, accessRequest)
	if err != nil {
		log.Printf("Failed to check access: %v", err)
		return
	}

	fmt.Printf("Access decision: %s - %s\n", decision.Decision, decision.Reason)
}

// exampleACLUsage demonstrates ACL operations
func exampleACLUsage(am *AuthManager) {
	fmt.Println("\n=== ACL Usage Example ===")

	ctx := context.Background()
	aclManager := am.GetACLManager()
	if aclManager == nil {
		fmt.Println("ACL manager not available")
		return
	}

	// Create ACL group
	group, err := aclManager.CreateACLGroup(ctx, "developers", "Development team", []string{"user1", "user2", "user3"}, "admin", nil)
	if err != nil {
		log.Printf("Failed to create ACL group: %v", err)
		return
	}

	fmt.Printf("Created ACL group: %s with %d members\n", group.Name, len(group.Members))

	// Create ACL entry
	entry, err := aclManager.CreateACLEntry(ctx, "developers", PrincipalTypeGroup, "api:staging", "deploy", ACLEffectAllow, map[string]interface{}{
		"time_of_day": "business_hours",
	}, 100, "Allow developers to deploy to staging during business hours", "admin", nil)
	if err != nil {
		log.Printf("Failed to create ACL entry: %v", err)
		return
	}

	fmt.Printf("Created ACL entry: %s\n", entry.Description)

	// Check access
	aclRequest := &ACLRequest{
		Principal: "user1",
		Resource:  "api:staging",
		Action:    "deploy",
		Context: map[string]interface{}{
			"time_of_day": "business_hours",
		},
	}

	decision, err := aclManager.CheckAccess(ctx, aclRequest)
	if err != nil {
		log.Printf("Failed to check ACL access: %v", err)
		return
	}

	fmt.Printf("ACL access decision: %s - %s\n", decision.Decision, decision.Reason)
}

// exampleABACUsage demonstrates ABAC operations
func exampleABACUsage(am *AuthManager) {
	fmt.Println("\n=== ABAC Usage Example ===")

	ctx := context.Background()
	abacManager := am.GetABACManager()
	if abacManager == nil {
		fmt.Println("ABAC manager not available")
		return
	}

	// Create ABAC policy
	rules := []ABACRule{
		{
			ID:          uuid.New(),
			Name:        "business_hours_rule",
			Description: "Allow access during business hours",
			Effect:      ABACEffectAllow,
			Condition: map[string]interface{}{
				"environment.time_of_day": 9, // 9 AM
			},
			Action:   "deploy",
			Resource: "staging",
			Priority: 100,
		},
	}

	target := ABACTarget{
		Subjects:  []string{"developer"},
		Resources: []string{"staging"},
		Actions:   []string{"deploy"},
	}

	policy, err := abacManager.CreatePolicy(ctx, "staging_deploy_policy", "Policy for staging deployments", ABACEffectAllow, rules, target, nil, 100, "admin", nil)
	if err != nil {
		log.Printf("Failed to create ABAC policy: %v", err)
		return
	}

	fmt.Printf("Created ABAC policy: %s\n", policy.Name)

	// Check access
	abacRequest := &ABACRequest{
		Subject: map[string]interface{}{
			"id":   "developer",
			"role": "developer",
		},
		Resource: map[string]interface{}{
			"id":   "staging",
			"type": "environment",
		},
		Action: map[string]interface{}{
			"name": "deploy",
		},
		Environment: map[string]interface{}{
			"time_of_day": 9,
		},
	}

	decision, err := abacManager.CheckAccess(ctx, abacRequest)
	if err != nil {
		log.Printf("Failed to check ABAC access: %v", err)
		return
	}

	fmt.Printf("ABAC access decision: %s - %s\n", decision.Decision, decision.Reason)
}

// examplePolicyEngineUsage demonstrates policy engine operations
func examplePolicyEngineUsage(am *AuthManager) {
	fmt.Println("\n=== Policy Engine Usage Example ===")

	ctx := context.Background()
	policyEngine := am.GetPolicyEngine()
	if policyEngine == nil {
		fmt.Println("Policy engine not available")
		return
	}

	// Create policy request
	policyRequest := &PolicyRequest{
		UserID:   "developer",
		TenantID: "tenant1",
		Resource: "api:staging",
		Action:   "deploy",
		Context: map[string]interface{}{
			"user_roles":       []string{"developer"},
			"user_permissions": []string{"deploy:staging"},
			"ip_address":       "192.168.1.100",
			"time_of_day":      9,
		},
	}

	// Evaluate policy
	decision, err := policyEngine.EvaluatePolicy(ctx, policyRequest, nil)
	if err != nil {
		log.Printf("Failed to evaluate policy: %v", err)
		return
	}

	fmt.Printf("Policy evaluation result: %s - %s (Source: %s)\n", decision.Decision, decision.Reason, decision.Source)
}

// exampleHTTPMiddlewareUsage demonstrates HTTP middleware usage
func exampleHTTPMiddlewareUsage(am *AuthManager) {
	fmt.Println("\n=== HTTP Middleware Usage Example ===")

	middleware := am.GetMiddleware()
	if middleware == nil {
		fmt.Println("Middleware not available")
		return
	}

	// Create a simple HTTP handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user info from context
		claims, ok := GetClaimsFromContext(r.Context())
		if ok {
			fmt.Fprintf(w, "Hello %s! You have roles: %v\n", claims.Email, claims.Roles)
		} else {
			fmt.Fprintf(w, "Hello anonymous user!\n")
		}
	})

	// Wrap with auth middleware
	protectedHandler := middleware.HTTPMiddleware(handler)

	// Start HTTP server (in a real application)
	fmt.Println("HTTP server would be started with protected endpoints")
	fmt.Printf("Middleware configured with token header: %s\n", middleware.config.TokenHeader)
	fmt.Printf("Require auth: %t\n", middleware.config.RequireAuth)
	fmt.Printf("Allow anonymous: %t\n", middleware.config.AllowAnonymous)

	// Note: In a real application, you would use:
	// http.ListenAndServe(":8080", protectedHandler)
	_ = protectedHandler
}

// exampleOAuth2Usage demonstrates OAuth2 usage
func exampleOAuth2Usage(am *AuthManager) {
	fmt.Println("\n=== OAuth2 Usage Example ===")

	ctx := context.Background()
	oauth2Manager := am.GetOAuth2Manager()
	if oauth2Manager == nil {
		fmt.Println("OAuth2 manager not available")
		return
	}

	// Register Google OAuth2 provider
	err := oauth2Manager.RegisterGoogleProvider(
		"your-google-client-id",
		"your-google-client-secret",
		"http://localhost:8080/auth/google/callback",
		[]string{"openid", "profile", "email"},
	)
	if err != nil {
		log.Printf("Failed to register Google provider: %v", err)
		return
	}

	fmt.Println("Google OAuth2 provider registered")

	// Get authorization URL
	state := "random-state-string"
	authURL, err := oauth2Manager.GetAuthURL(ctx, "google", state)
	if err != nil {
		log.Printf("Failed to get auth URL: %v", err)
		return
	}

	fmt.Printf("Authorization URL: %s\n", authURL)

	// Note: In a real application, you would:
	// 1. Redirect user to authURL
	// 2. Handle callback with authorization code
	// 3. Exchange code for token
	// 4. Get user info
}

// ExampleConfiguration demonstrates configuration usage
func ExampleConfiguration() {
	fmt.Println("\n=== Configuration Example ===")

	// Load from environment variables
	config := LoadAuthConfigFromEnv()
	fmt.Printf("JWT Secret Key: %s\n", config.JWT.SecretKey[:10]+"...")
	fmt.Printf("JWT Access Expiry: %s\n", config.JWT.AccessExpiry)
	fmt.Printf("RBAC Enabled: %t\n", config.RBAC.Enabled)
	fmt.Printf("ACL Enabled: %t\n", config.ACL.Enabled)
	fmt.Printf("ABAC Enabled: %t\n", config.ABAC.Enabled)
	fmt.Printf("Audit Enabled: %t\n", config.Audit.Enabled)

	// Validate configuration
	if err := ValidateAuthConfig(config); err != nil {
		log.Printf("Configuration validation failed: %v", err)
		return
	}

	fmt.Println("Configuration is valid")

	// Save configuration to file
	if err := SaveAuthConfigToFile(config, "/tmp/auth-config.json"); err != nil {
		log.Printf("Failed to save config: %v", err)
		return
	}

	fmt.Println("Configuration saved to /tmp/auth-config.json")
}
