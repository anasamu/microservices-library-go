package graphql

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
)

// SchemaBuilder helps build GraphQL schemas programmatically
type SchemaBuilder struct {
	types         map[string]interface{}
	queries       map[string]interface{}
	mutations     map[string]interface{}
	subscriptions map[string]interface{}
	logger        *logrus.Logger
}

// NewSchemaBuilder creates a new schema builder
func NewSchemaBuilder(logger *logrus.Logger) *SchemaBuilder {
	return &SchemaBuilder{
		types:         make(map[string]interface{}),
		queries:       make(map[string]interface{}),
		mutations:     make(map[string]interface{}),
		subscriptions: make(map[string]interface{}),
		logger:        logger,
	}
}

// AddType adds a type to the schema
func (sb *SchemaBuilder) AddType(name string, typeDef interface{}) *SchemaBuilder {
	sb.types[name] = typeDef
	return sb
}

// AddQuery adds a query to the schema
func (sb *SchemaBuilder) AddQuery(name string, resolver interface{}) *SchemaBuilder {
	sb.queries[name] = resolver
	return sb
}

// AddMutation adds a mutation to the schema
func (sb *SchemaBuilder) AddMutation(name string, resolver interface{}) *SchemaBuilder {
	sb.mutations[name] = resolver
	return sb
}

// AddSubscription adds a subscription to the schema
func (sb *SchemaBuilder) AddSubscription(name string, resolver interface{}) *SchemaBuilder {
	sb.subscriptions[name] = resolver
	return sb
}

// Build builds the GraphQL schema
func (sb *SchemaBuilder) Build() (string, error) {
	var schema strings.Builder

	// Add scalar types
	schema.WriteString(`
scalar Time
scalar Upload
scalar JSON

directive @auth on FIELD_DEFINITION
directive @hasRole(role: String!) on FIELD_DEFINITION
directive @hasPermission(permission: String!) on FIELD_DEFINITION
directive @rateLimit(limit: Int!, window: String!) on FIELD_DEFINITION
directive @cache(maxAge: Int!) on FIELD_DEFINITION
directive @deprecated(reason: String) on FIELD_DEFINITION
`)

	// Add types
	for name, typeDef := range sb.types {
		schema.WriteString(fmt.Sprintf("\ntype %s {\n", name))
		schema.WriteString(sb.buildTypeFields(typeDef))
		schema.WriteString("}\n")
	}

	// Add input types
	schema.WriteString("\n# Input Types\n")
	for name, typeDef := range sb.types {
		if strings.HasSuffix(name, "Input") {
			schema.WriteString(fmt.Sprintf("\ninput %s {\n", name))
			schema.WriteString(sb.buildInputFields(typeDef))
			schema.WriteString("}\n")
		}
	}

	// Add enums
	schema.WriteString("\n# Enums\n")
	for name, typeDef := range sb.types {
		if strings.HasSuffix(name, "Enum") {
			schema.WriteString(fmt.Sprintf("\nenum %s {\n", name))
			schema.WriteString(sb.buildEnumValues(typeDef))
			schema.WriteString("}\n")
		}
	}

	// Add queries
	if len(sb.queries) > 0 {
		schema.WriteString("\ntype Query {\n")
		for name := range sb.queries {
			schema.WriteString(fmt.Sprintf("  %s: %s\n", name, sb.getReturnType(name)))
		}
		schema.WriteString("}\n")
	}

	// Add mutations
	if len(sb.mutations) > 0 {
		schema.WriteString("\ntype Mutation {\n")
		for name := range sb.mutations {
			schema.WriteString(fmt.Sprintf("  %s: %s\n", name, sb.getReturnType(name)))
		}
		schema.WriteString("}\n")
	}

	// Add subscriptions
	if len(sb.subscriptions) > 0 {
		schema.WriteString("\ntype Subscription {\n")
		for name := range sb.subscriptions {
			schema.WriteString(fmt.Sprintf("  %s: %s\n", name, sb.getReturnType(name)))
		}
		schema.WriteString("}\n")
	}

	return schema.String(), nil
}

// buildTypeFields builds fields for a type
func (sb *SchemaBuilder) buildTypeFields(typeDef interface{}) string {
	var fields strings.Builder

	// Use reflection to get struct fields
	v := reflect.ValueOf(typeDef)
	t := reflect.TypeOf(typeDef)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return ""
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get GraphQL tag
		gqlTag := field.Tag.Get("gql")
		if gqlTag == "" {
			gqlTag = field.Name
		}

		// Get field type
		gqlType := sb.getGraphQLType(fieldType)

		// Add field
		fields.WriteString(fmt.Sprintf("  %s: %s\n", gqlTag, gqlType))
	}

	return fields.String()
}

// buildInputFields builds fields for an input type
func (sb *SchemaBuilder) buildInputFields(typeDef interface{}) string {
	return sb.buildTypeFields(typeDef)
}

// buildEnumValues builds values for an enum
func (sb *SchemaBuilder) buildEnumValues(typeDef interface{}) string {
	var values strings.Builder

	// Use reflection to get enum values
	v := reflect.ValueOf(typeDef)
	t := reflect.TypeOf(typeDef)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return ""
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get GraphQL tag
		gqlTag := field.Tag.Get("gql")
		if gqlTag == "" {
			gqlTag = strings.ToUpper(field.Name)
		}

		// Add enum value
		values.WriteString(fmt.Sprintf("  %s\n", gqlTag))
	}

	return values.String()
}

// getGraphQLType converts Go type to GraphQL type
func (sb *SchemaBuilder) getGraphQLType(goType reflect.Type) string {
	// Handle pointers
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}

	// Handle slices
	if goType.Kind() == reflect.Slice {
		elementType := sb.getGraphQLType(goType.Elem())
		return fmt.Sprintf("[%s]", elementType)
	}

	// Handle maps
	if goType.Kind() == reflect.Map {
		return "JSON"
	}

	// Handle basic types
	switch goType.Kind() {
	case reflect.String:
		return "String"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "Int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "Int"
	case reflect.Float32, reflect.Float64:
		return "Float"
	case reflect.Bool:
		return "Boolean"
	case reflect.Struct:
		// Check if it's a time.Time
		if goType.String() == "time.Time" {
			return "Time"
		}
		// Return the struct name
		return goType.Name()
	default:
		return "String"
	}
}

// getReturnType gets the return type for a resolver
func (sb *SchemaBuilder) getReturnType(name string) string {
	// Simple heuristic - can be improved
	if strings.HasPrefix(name, "get") || strings.HasPrefix(name, "list") {
		return "String" // Should be actual return type
	}
	if strings.HasPrefix(name, "create") || strings.HasPrefix(name, "update") || strings.HasPrefix(name, "delete") {
		return "String" // Should be actual return type
	}
	return "String"
}

// SchemaValidator validates GraphQL schemas
type SchemaValidator struct {
	logger *logrus.Logger
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator(logger *logrus.Logger) *SchemaValidator {
	return &SchemaValidator{
		logger: logger,
	}
}

// Validate validates a GraphQL schema
func (sv *SchemaValidator) Validate(schema string) error {
	// Basic validation - can be extended
	if strings.TrimSpace(schema) == "" {
		return fmt.Errorf("schema cannot be empty")
	}

	// Check for required types
	requiredTypes := []string{"Query", "Mutation", "Subscription"}
	for _, requiredType := range requiredTypes {
		if !strings.Contains(schema, fmt.Sprintf("type %s", requiredType)) {
			sv.logger.Warnf("Schema missing %s type", requiredType)
		}
	}

	return nil
}

// SchemaGenerator generates GraphQL schemas from Go structs
type SchemaGenerator struct {
	logger *logrus.Logger
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(logger *logrus.Logger) *SchemaGenerator {
	return &SchemaGenerator{
		logger: logger,
	}
}

// GenerateFromStruct generates GraphQL schema from a Go struct
func (sg *SchemaGenerator) GenerateFromStruct(structType interface{}) (string, error) {
	var schema strings.Builder

	// Get struct info
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("expected struct, got %s", t.Kind())
	}

	// Generate type definition
	schema.WriteString(fmt.Sprintf("type %s {\n", t.Name()))

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get GraphQL tag
		gqlTag := field.Tag.Get("gql")
		if gqlTag == "" {
			gqlTag = field.Name
		}

		// Get field type
		gqlType := sg.getGraphQLType(field.Type)

		// Add field
		schema.WriteString(fmt.Sprintf("  %s: %s\n", gqlTag, gqlType))
	}

	schema.WriteString("}\n")

	return schema.String(), nil
}

// getGraphQLType converts Go type to GraphQL type
func (sg *SchemaGenerator) getGraphQLType(goType reflect.Type) string {
	// Handle pointers
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}

	// Handle slices
	if goType.Kind() == reflect.Slice {
		elementType := sg.getGraphQLType(goType.Elem())
		return fmt.Sprintf("[%s]", elementType)
	}

	// Handle basic types
	switch goType.Kind() {
	case reflect.String:
		return "String"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "Int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "Int"
	case reflect.Float32, reflect.Float64:
		return "Float"
	case reflect.Bool:
		return "Boolean"
	case reflect.Struct:
		// Check if it's a time.Time
		if goType.String() == "time.Time" {
			return "Time"
		}
		// Return the struct name
		return goType.Name()
	default:
		return "String"
	}
}

// ContextKey represents a context key
type ContextKey string

const (
	// UserContextKey is the key for user in context
	UserContextKey ContextKey = "user"
	// TenantContextKey is the key for tenant in context
	TenantContextKey ContextKey = "tenant"
	// AuthTokenContextKey is the key for auth token in context
	AuthTokenContextKey ContextKey = "auth_token"
)

// GetUserFromContext gets user from GraphQL context
func GetUserFromContext(ctx context.Context) (interface{}, bool) {
	user, ok := ctx.Value(UserContextKey).(interface{})
	return user, ok
}

// GetTenantFromContext gets tenant from GraphQL context
func GetTenantFromContext(ctx context.Context) (string, bool) {
	tenant, ok := ctx.Value(TenantContextKey).(string)
	return tenant, ok
}

// GetAuthTokenFromContext gets auth token from GraphQL context
func GetAuthTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(AuthTokenContextKey).(string)
	return token, ok
}

// SetUserInContext sets user in GraphQL context
func SetUserInContext(ctx context.Context, user interface{}) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// SetTenantInContext sets tenant in GraphQL context
func SetTenantInContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, TenantContextKey, tenant)
}

// SetAuthTokenInContext sets auth token in GraphQL context
func SetAuthTokenInContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, AuthTokenContextKey, token)
}
