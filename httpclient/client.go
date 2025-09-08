package httpclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// HTTPClientManager manages HTTP clients with retry and circuit breaker capabilities
type HTTPClientManager struct {
	clients map[string]*resty.Client
	configs map[string]*HTTPClientConfig
	logger  *logrus.Logger
}

// HTTPClientConfig holds HTTP client configuration
type HTTPClientConfig struct {
	ID                   string
	Name                 string
	BaseURL              string
	Timeout              time.Duration
	RetryCount           int
	RetryWaitTime        time.Duration
	RetryMaxWaitTime     time.Duration
	MaxRetryWaitTime     time.Duration
	RetryOnStatus        []int
	Headers              map[string]string
	Auth                 *AuthConfig
	TLS                  *TLSConfig
	Proxy                string
	UserAgent            string
	EnableRetry          bool
	EnableCircuitBreaker bool
	CircuitBreakerConfig *CircuitBreakerConfig
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type         string // "basic", "bearer", "apikey"
	Username     string
	Password     string
	Token        string
	APIKey       string
	APIKeyHeader string
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	InsecureSkipVerify bool
	CertFile           string
	KeyFile            string
	CAFile             string
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxRequests uint32
	Interval    time.Duration
	Timeout     time.Duration
	ReadyToTrip func(counts map[string]int) bool
}

// HTTPRequest represents an HTTP request
type HTTPRequest struct {
	ID        uuid.UUID              `json:"id"`
	Method    string                 `json:"method"`
	URL       string                 `json:"url"`
	Headers   map[string]string      `json:"headers"`
	Body      interface{}            `json:"body"`
	Query     map[string]string      `json:"query"`
	Timeout   time.Duration          `json:"timeout"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	ID        uuid.UUID              `json:"id"`
	RequestID uuid.UUID              `json:"request_id"`
	Status    int                    `json:"status"`
	Headers   map[string]string      `json:"headers"`
	Body      []byte                 `json:"body"`
	Size      int64                  `json:"size"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// HTTPError represents an HTTP error
type HTTPError struct {
	ID        uuid.UUID              `json:"id"`
	RequestID uuid.UUID              `json:"request_id"`
	Message   string                 `json:"message"`
	Status    int                    `json:"status"`
	Body      []byte                 `json:"body"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// NewHTTPClientManager creates a new HTTP client manager
func NewHTTPClientManager(logger *logrus.Logger) *HTTPClientManager {
	return &HTTPClientManager{
		clients: make(map[string]*resty.Client),
		configs: make(map[string]*HTTPClientConfig),
		logger:  logger,
	}
}

// CreateHTTPClient creates a new HTTP client
func (hcm *HTTPClientManager) CreateHTTPClient(config *HTTPClientConfig) error {
	// Set default values
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}
	if config.RetryWaitTime == 0 {
		config.RetryWaitTime = 1 * time.Second
	}
	if config.RetryMaxWaitTime == 0 {
		config.RetryMaxWaitTime = 10 * time.Second
	}
	if config.UserAgent == "" {
		config.UserAgent = "SIAKAD-HTTP-Client/1.0"
	}
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}
	config.UpdatedAt = time.Now()

	// Create Resty client
	client := resty.New()

	// Set base URL
	if config.BaseURL != "" {
		client.SetBaseURL(config.BaseURL)
	}

	// Set timeout
	client.SetTimeout(config.Timeout)

	// Set retry configuration
	if config.EnableRetry {
		client.SetRetryCount(config.RetryCount)
		client.SetRetryWaitTime(config.RetryWaitTime)
		client.SetRetryMaxWaitTime(config.RetryMaxWaitTime)
		if len(config.RetryOnStatus) > 0 {
			client.SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
				return 0, nil
			})
		}
	}

	// Set headers
	if config.Headers != nil {
		client.SetHeaders(config.Headers)
	}

	// Set user agent
	client.SetHeader("User-Agent", config.UserAgent)

	// Set authentication
	if config.Auth != nil {
		switch config.Auth.Type {
		case "basic":
			client.SetBasicAuth(config.Auth.Username, config.Auth.Password)
		case "bearer":
			client.SetAuthToken(config.Auth.Token)
		case "apikey":
			header := "X-API-Key"
			if config.Auth.APIKeyHeader != "" {
				header = config.Auth.APIKeyHeader
			}
			client.SetHeader(header, config.Auth.APIKey)
		}
	}

	// Set TLS configuration
	if config.TLS != nil {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: config.TLS.InsecureSkipVerify,
		}
		client.SetTLSClientConfig(tlsConfig)
	}

	// Set proxy
	if config.Proxy != "" {
		client.SetProxy(config.Proxy)
	}

	// Add request/response middleware
	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		req.SetHeader("X-Request-ID", uuid.New().String())
		return nil
	})

	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		hcm.logResponse(resp)
		return nil
	})

	client.OnError(func(req *resty.Request, err error) {
		hcm.logError(req, err)
	})

	hcm.clients[config.ID] = client
	hcm.configs[config.ID] = config

	hcm.logger.WithFields(logrus.Fields{
		"id":            config.ID,
		"name":          config.Name,
		"base_url":      config.BaseURL,
		"timeout":       config.Timeout,
		"retry_count":   config.RetryCount,
		"retry_enabled": config.EnableRetry,
	}).Info("HTTP client created successfully")

	return nil
}

// GetClient returns an HTTP client by ID
func (hcm *HTTPClientManager) GetClient(clientID string) (*resty.Client, error) {
	client, exists := hcm.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("HTTP client not found: %s", clientID)
	}
	return client, nil
}

// MakeRequest makes an HTTP request using the specified client
func (hcm *HTTPClientManager) MakeRequest(ctx context.Context, clientID string, request *HTTPRequest) (*HTTPResponse, error) {
	client, err := hcm.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	// Create Resty request
	req := client.R().SetContext(ctx)

	// Set method and URL
	req.Method = request.Method
	req.URL = request.URL

	// Set headers
	if request.Headers != nil {
		req.SetHeaders(request.Headers)
	}

	// Set query parameters
	if request.Query != nil {
		req.SetQueryParams(request.Query)
	}

	// Set body
	if request.Body != nil {
		req.SetBody(request.Body)
	}

	// Set timeout if specified
	if request.Timeout > 0 {
		req.SetTimeout(request.Timeout)
	}

	// Execute request
	start := time.Now()
	resp, err := req.Send()
	duration := time.Since(start)

	if err != nil {
		httpError := &HTTPError{
			ID:        uuid.New(),
			RequestID: request.ID,
			Message:   err.Error(),
			Status:    0,
			Duration:  duration,
			Metadata:  request.Metadata,
			CreatedAt: time.Now(),
		}
		return nil, httpError
	}

	// Create response
	response := &HTTPResponse{
		ID:        uuid.New(),
		RequestID: request.ID,
		Status:    resp.StatusCode(),
		Headers:   resp.Header(),
		Body:      resp.Body(),
		Size:      resp.Size(),
		Duration:  duration,
		Metadata:  request.Metadata,
		CreatedAt: time.Now(),
	}

	return response, nil
}

// MakeRequestAsync makes an HTTP request asynchronously
func (hcm *HTTPClientManager) MakeRequestAsync(ctx context.Context, clientID string, request *HTTPRequest) <-chan HTTPResult {
	result := make(chan HTTPResult, 1)

	go func() {
		defer close(result)
		response, err := hcm.MakeRequest(ctx, clientID, request)
		result <- HTTPResult{
			Response: response,
			Error:    err,
		}
	}()

	return result
}

// MakeRequestWithRetry makes an HTTP request with custom retry logic
func (hcm *HTTPClientManager) MakeRequestWithRetry(ctx context.Context, clientID string, request *HTTPRequest, maxRetries int, retryDelay time.Duration) (*HTTPResponse, error) {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		response, err := hcm.MakeRequest(ctx, clientID, request)
		if err == nil {
			return response, nil
		}

		lastErr = err
		if i < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

// GetConfig returns the configuration of an HTTP client
func (hcm *HTTPClientManager) GetConfig(clientID string) (*HTTPClientConfig, error) {
	config, exists := hcm.configs[clientID]
	if !exists {
		return nil, fmt.Errorf("HTTP client not found: %s", clientID)
	}
	return config, nil
}

// UpdateHTTPClient updates an HTTP client configuration
func (hcm *HTTPClientManager) UpdateHTTPClient(clientID string, config *HTTPClientConfig) error {
	if _, exists := hcm.configs[clientID]; !exists {
		return fmt.Errorf("HTTP client not found: %s", clientID)
	}

	// Update configuration
	config.ID = clientID
	config.UpdatedAt = time.Now()

	// Recreate client with new configuration
	return hcm.CreateHTTPClient(config)
}

// DeleteHTTPClient deletes an HTTP client
func (hcm *HTTPClientManager) DeleteHTTPClient(clientID string) error {
	if _, exists := hcm.clients[clientID]; !exists {
		return fmt.Errorf("HTTP client not found: %s", clientID)
	}

	delete(hcm.clients, clientID)
	delete(hcm.configs, clientID)

	hcm.logger.WithField("id", clientID).Info("HTTP client deleted successfully")
	return nil
}

// logResponse logs HTTP response information
func (hcm *HTTPClientManager) logResponse(resp *resty.Response) {
	hcm.logger.WithFields(logrus.Fields{
		"status":     resp.StatusCode(),
		"method":     resp.Request.Method,
		"url":        resp.Request.URL,
		"size":       resp.Size(),
		"duration":   resp.Time(),
		"request_id": resp.Header().Get("X-Request-ID"),
	}).Debug("HTTP response")
}

// logError logs HTTP error information
func (hcm *HTTPClientManager) logError(req *resty.Request, err error) {
	hcm.logger.WithFields(logrus.Fields{
		"method":     req.Method,
		"url":        req.URL,
		"error":      err,
		"request_id": req.Header.Get("X-Request-ID"),
	}).Error("HTTP request error")
}

// HTTPResult represents the result of an async HTTP request
type HTTPResult struct {
	Response *HTTPResponse
	Error    error
}

// CreateRequest creates a new HTTP request
func CreateRequest(method, url string, body interface{}) *HTTPRequest {
	return &HTTPRequest{
		ID:        uuid.New(),
		Method:    method,
		URL:       url,
		Body:      body,
		Headers:   make(map[string]string),
		Query:     make(map[string]string),
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

// SetHeader sets a header in the request
func (r *HTTPRequest) SetHeader(key, value string) {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers[key] = value
}

// SetQueryParam sets a query parameter in the request
func (r *HTTPRequest) SetQueryParam(key, value string) {
	if r.Query == nil {
		r.Query = make(map[string]string)
	}
	r.Query[key] = value
}

// SetMetadata sets metadata in the request
func (r *HTTPRequest) SetMetadata(key string, value interface{}) {
	if r.Metadata == nil {
		r.Metadata = make(map[string]interface{})
	}
	r.Metadata[key] = value
}

// CreateAPIClient creates an HTTP client for API calls
func (hcm *HTTPClientManager) CreateAPIClient(name, baseURL string, timeout time.Duration) error {
	config := &HTTPClientConfig{
		Name:             name,
		BaseURL:          baseURL,
		Timeout:          timeout,
		RetryCount:       3,
		RetryWaitTime:    1 * time.Second,
		RetryMaxWaitTime: 10 * time.Second,
		RetryOnStatus:    []int{500, 502, 503, 504},
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
		UserAgent:            "SIAKAD-API-Client/1.0",
		EnableRetry:          true,
		EnableCircuitBreaker: false,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	return hcm.CreateHTTPClient(config)
}

// CreateExternalServiceClient creates an HTTP client for external service calls
func (hcm *HTTPClientManager) CreateExternalServiceClient(name, baseURL string, timeout time.Duration, auth *AuthConfig) error {
	config := &HTTPClientConfig{
		Name:             name,
		BaseURL:          baseURL,
		Timeout:          timeout,
		RetryCount:       5,
		RetryWaitTime:    2 * time.Second,
		RetryMaxWaitTime: 30 * time.Second,
		RetryOnStatus:    []int{500, 502, 503, 504, 429},
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
		Auth:                 auth,
		UserAgent:            "SIAKAD-External-Client/1.0",
		EnableRetry:          true,
		EnableCircuitBreaker: true,
		CircuitBreakerConfig: &CircuitBreakerConfig{
			MaxRequests: 3,
			Interval:    10 * time.Second,
			Timeout:     60 * time.Second,
			ReadyToTrip: func(counts map[string]int) bool {
				return counts["failures"] >= 5
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return hcm.CreateHTTPClient(config)
}
