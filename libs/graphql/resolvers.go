package graphql

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sirupsen/logrus"
)

// ResolverRegistry manages GraphQL resolvers
type ResolverRegistry struct {
	resolvers map[string]interface{}
	logger    *logrus.Logger
}

// NewResolverRegistry creates a new resolver registry
func NewResolverRegistry(logger *logrus.Logger) *ResolverRegistry {
	return &ResolverRegistry{
		resolvers: make(map[string]interface{}),
		logger:    logger,
	}
}

// Register registers a resolver
func (rr *ResolverRegistry) Register(name string, resolver interface{}) {
	rr.resolvers[name] = resolver
	rr.logger.WithField("resolver", name).Info("GraphQL resolver registered")
}

// Get gets a resolver by name
func (rr *ResolverRegistry) Get(name string) (interface{}, bool) {
	resolver, exists := rr.resolvers[name]
	return resolver, exists
}

// List lists all registered resolvers
func (rr *ResolverRegistry) List() []string {
	var names []string
	for name := range rr.resolvers {
		names = append(names, name)
	}
	return names
}

// ResolverBuilder helps build GraphQL resolvers
type ResolverBuilder struct {
	registry *ResolverRegistry
	logger   *logrus.Logger
}

// NewResolverBuilder creates a new resolver builder
func NewResolverBuilder(logger *logrus.Logger) *ResolverBuilder {
	return &ResolverBuilder{
		registry: NewResolverRegistry(logger),
		logger:   logger,
	}
}

// AddQuery adds a query resolver
func (rb *ResolverBuilder) AddQuery(name string, resolver interface{}) *ResolverBuilder {
	rb.registry.Register(fmt.Sprintf("Query.%s", name), resolver)
	return rb
}

// AddMutation adds a mutation resolver
func (rb *ResolverBuilder) AddMutation(name string, resolver interface{}) *ResolverBuilder {
	rb.registry.Register(fmt.Sprintf("Mutation.%s", name), resolver)
	return rb
}

// AddSubscription adds a subscription resolver
func (rb *ResolverBuilder) AddSubscription(name string, resolver interface{}) *ResolverBuilder {
	rb.registry.Register(fmt.Sprintf("Subscription.%s", name), resolver)
	return rb
}

// AddFieldResolver adds a field resolver
func (rb *ResolverBuilder) AddFieldResolver(typeName, fieldName string, resolver interface{}) *ResolverBuilder {
	rb.registry.Register(fmt.Sprintf("%s.%s", typeName, fieldName), resolver)
	return rb
}

// Build builds the resolver registry
func (rb *ResolverBuilder) Build() *ResolverRegistry {
	return rb.registry
}

// BaseResolver provides common resolver functionality
type BaseResolver struct {
	logger *logrus.Logger
}

// NewBaseResolver creates a new base resolver
func NewBaseResolver(logger *logrus.Logger) *BaseResolver {
	return &BaseResolver{
		logger: logger,
	}
}

// LogOperation logs a GraphQL operation
func (br *BaseResolver) LogOperation(ctx context.Context, operation string, args interface{}) {
	br.logger.WithFields(logrus.Fields{
		"operation": operation,
		"args":      args,
		"user":      GetUserFromContext(ctx),
		"tenant":    GetTenantFromContext(ctx),
	}).Info("GraphQL operation executed")
}

// HandleError handles GraphQL errors
func (br *BaseResolver) HandleError(ctx context.Context, operation string, err error) error {
	br.logger.WithFields(logrus.Fields{
		"operation": operation,
		"error":     err.Error(),
		"user":      GetUserFromContext(ctx),
		"tenant":    GetTenantFromContext(ctx),
	}).Error("GraphQL operation failed")

	return &graphql.Error{
		Message: err.Error(),
		Extensions: map[string]interface{}{
			"operation": operation,
		},
	}
}

// ValidateInput validates input parameters
func (br *BaseResolver) ValidateInput(ctx context.Context, input interface{}) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	// Use reflection to validate struct fields
	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("input must be a struct")
	}

	// Validate required fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Check for required tag
		if required := field.Tag.Get("required"); required == "true" {
			if fieldValue.IsZero() {
				return fmt.Errorf("field %s is required", field.Name)
			}
		}

		// Check for validation tag
		if validation := field.Tag.Get("validate"); validation != "" {
			if err := br.validateField(fieldValue, validation); err != nil {
				return fmt.Errorf("field %s validation failed: %w", field.Name, err)
			}
		}
	}

	return nil
}

// validateField validates a single field
func (br *BaseResolver) validateField(fieldValue reflect.Value, validation string) error {
	// Simple validation - can be extended with more complex rules
	rules := strings.Split(validation, ",")

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)

		switch {
		case rule == "email":
			if fieldValue.Kind() == reflect.String {
				email := fieldValue.String()
				if !strings.Contains(email, "@") {
					return fmt.Errorf("invalid email format")
				}
			}
		case rule == "min=1":
			if fieldValue.Kind() == reflect.String {
				if len(fieldValue.String()) < 1 {
					return fmt.Errorf("field cannot be empty")
				}
			}
		case strings.HasPrefix(rule, "min="):
			// Extract minimum length
			minLen := strings.TrimPrefix(rule, "min=")
			if fieldValue.Kind() == reflect.String {
				if len(fieldValue.String()) < len(minLen) {
					return fmt.Errorf("field must be at least %s characters", minLen)
				}
			}
		case strings.HasPrefix(rule, "max="):
			// Extract maximum length
			maxLen := strings.TrimPrefix(rule, "max=")
			if fieldValue.Kind() == reflect.String {
				if len(fieldValue.String()) > len(maxLen) {
					return fmt.Errorf("field must be at most %s characters", maxLen)
				}
			}
		}
	}

	return nil
}

// PaginationResolver provides pagination functionality
type PaginationResolver struct {
	*BaseResolver
}

// NewPaginationResolver creates a new pagination resolver
func NewPaginationResolver(logger *logrus.Logger) *PaginationResolver {
	return &PaginationResolver{
		BaseResolver: NewBaseResolver(logger),
	}
}

// PaginationInput represents pagination input
type PaginationInput struct {
	Page     int `json:"page" gql:"page" validate:"min=1"`
	PageSize int `json:"pageSize" gql:"pageSize" validate:"min=1,max=100"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Page       int   `json:"page" gql:"page"`
	PageSize   int   `json:"pageSize" gql:"pageSize"`
	Total      int64 `json:"total" gql:"total"`
	TotalPages int   `json:"totalPages" gql:"totalPages"`
	HasNext    bool  `json:"hasNext" gql:"hasNext"`
	HasPrev    bool  `json:"hasPrev" gql:"hasPrev"`
}

// PaginatedResult represents a paginated result
type PaginatedResult struct {
	Data       interface{}     `json:"data" gql:"data"`
	Pagination *PaginationInfo `json:"pagination" gql:"pagination"`
}

// CalculatePagination calculates pagination info
func (pr *PaginationResolver) CalculatePagination(page, pageSize int, total int64) *PaginationInfo {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginationInfo{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// FilterResolver provides filtering functionality
type FilterResolver struct {
	*BaseResolver
}

// NewFilterResolver creates a new filter resolver
func NewFilterResolver(logger *logrus.Logger) *FilterResolver {
	return &FilterResolver{
		BaseResolver: NewBaseResolver(logger),
	}
}

// FilterInput represents filter input
type FilterInput struct {
	Field    string      `json:"field" gql:"field"`
	Operator string      `json:"operator" gql:"operator"`
	Value    interface{} `json:"value" gql:"value"`
}

// SortInput represents sort input
type SortInput struct {
	Field string `json:"field" gql:"field"`
	Order string `json:"order" gql:"order"` // "asc" or "desc"
}

// SearchInput represents search input
type SearchInput struct {
	Query      string           `json:"query" gql:"query"`
	Fields     []string         `json:"fields" gql:"fields"`
	Filters    []*FilterInput   `json:"filters" gql:"filters"`
	Sort       []*SortInput     `json:"sort" gql:"sort"`
	Pagination *PaginationInput `json:"pagination" gql:"pagination"`
}

// ApplyFilters applies filters to data
func (fr *FilterResolver) ApplyFilters(data interface{}, filters []*FilterInput) (interface{}, error) {
	// This is a simplified implementation
	// In a real application, you would apply filters based on your data source
	return data, nil
}

// ApplySort applies sorting to data
func (fr *FilterResolver) ApplySort(data interface{}, sort []*SortInput) (interface{}, error) {
	// This is a simplified implementation
	// In a real application, you would apply sorting based on your data source
	return data, nil
}

// ApplySearch applies search to data
func (fr *FilterResolver) ApplySearch(data interface{}, search *SearchInput) (interface{}, error) {
	// This is a simplified implementation
	// In a real application, you would apply search based on your data source
	return data, nil
}

// CacheResolver provides caching functionality
type CacheResolver struct {
	*BaseResolver
	cache map[string]interface{}
}

// NewCacheResolver creates a new cache resolver
func NewCacheResolver(logger *logrus.Logger) *CacheResolver {
	return &CacheResolver{
		BaseResolver: NewBaseResolver(logger),
		cache:        make(map[string]interface{}),
	}
}

// Get gets a value from cache
func (cr *CacheResolver) Get(key string) (interface{}, bool) {
	value, exists := cr.cache[key]
	return value, exists
}

// Set sets a value in cache
func (cr *CacheResolver) Set(key string, value interface{}) {
	cr.cache[key] = value
}

// Delete deletes a value from cache
func (cr *CacheResolver) Delete(key string) {
	delete(cr.cache, key)
}

// Clear clears the cache
func (cr *CacheResolver) Clear() {
	cr.cache = make(map[string]interface{})
}

// ResolveWithCache resolves with caching
func (cr *CacheResolver) ResolveWithCache(ctx context.Context, key string, resolver func() (interface{}, error)) (interface{}, error) {
	// Check cache first
	if value, exists := cr.Get(key); exists {
		cr.logger.WithField("cache_key", key).Debug("Cache hit")
		return value, nil
	}

	// Resolve and cache
	value, err := resolver()
	if err != nil {
		return nil, err
	}

	cr.Set(key, value)
	cr.logger.WithField("cache_key", key).Debug("Cache miss, value cached")

	return value, nil
}

// ErrorHandler handles GraphQL errors
type ErrorHandler struct {
	logger *logrus.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *logrus.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// Handle handles an error
func (eh *ErrorHandler) Handle(ctx context.Context, err error) *graphql.Error {
	// Log the error
	eh.logger.WithFields(logrus.Fields{
		"error":  err.Error(),
		"user":   GetUserFromContext(ctx),
		"tenant": GetTenantFromContext(ctx),
	}).Error("GraphQL error occurred")

	// Return GraphQL error
	return &graphql.Error{
		Message: err.Error(),
		Extensions: map[string]interface{}{
			"code": "INTERNAL_ERROR",
		},
	}
}

// HandleValidationError handles validation errors
func (eh *ErrorHandler) HandleValidationError(ctx context.Context, err error) *graphql.Error {
	eh.logger.WithFields(logrus.Fields{
		"error":  err.Error(),
		"user":   GetUserFromContext(ctx),
		"tenant": GetTenantFromContext(ctx),
	}).Warn("GraphQL validation error")

	return &graphql.Error{
		Message: err.Error(),
		Extensions: map[string]interface{}{
			"code": "VALIDATION_ERROR",
		},
	}
}

// HandleNotFoundError handles not found errors
func (eh *ErrorHandler) HandleNotFoundError(ctx context.Context, resource string) *graphql.Error {
	message := fmt.Sprintf("%s not found", resource)

	eh.logger.WithFields(logrus.Fields{
		"resource": resource,
		"user":     GetUserFromContext(ctx),
		"tenant":   GetTenantFromContext(ctx),
	}).Warn("GraphQL not found error")

	return &graphql.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code": "NOT_FOUND",
		},
	}
}

// HandleUnauthorizedError handles unauthorized errors
func (eh *ErrorHandler) HandleUnauthorizedError(ctx context.Context) *graphql.Error {
	eh.logger.WithFields(logrus.Fields{
		"user":   GetUserFromContext(ctx),
		"tenant": GetTenantFromContext(ctx),
	}).Warn("GraphQL unauthorized error")

	return &graphql.Error{
		Message: "Unauthorized",
		Extensions: map[string]interface{}{
			"code": "UNAUTHORIZED",
		},
	}
}

// HandleForbiddenError handles forbidden errors
func (eh *ErrorHandler) HandleForbiddenError(ctx context.Context) *graphql.Error {
	eh.logger.WithFields(logrus.Fields{
		"user":   GetUserFromContext(ctx),
		"tenant": GetTenantFromContext(ctx),
	}).Warn("GraphQL forbidden error")

	return &graphql.Error{
		Message: "Forbidden",
		Extensions: map[string]interface{}{
			"code": "FORBIDDEN",
		},
	}
}
