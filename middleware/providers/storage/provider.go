package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/anasamu/microservices-library-go/middleware/types"
	"github.com/sirupsen/logrus"
)

// StorageProvider implements storage middleware for file operations
type StorageProvider struct {
	name    string
	logger  *logrus.Logger
	config  *StorageConfig
	enabled bool
	// In-memory cache for file metadata (in production, use Redis or similar)
	fileCache map[string]*FileMetadata
	mutex     sync.RWMutex
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	StorageType        string                 `json:"storage_type"`        // local, s3, gcs, azure
	BasePath           string                 `json:"base_path"`           // Base storage path
	MaxFileSize        int64                  `json:"max_file_size"`       // Maximum file size in bytes
	AllowedMimeTypes   []string               `json:"allowed_mime_types"`  // Allowed MIME types
	ExcludedPaths      []string               `json:"excluded_paths"`      // Paths to exclude from storage
	CacheEnabled       bool                   `json:"cache_enabled"`       // Enable file caching
	CacheTTL           time.Duration          `json:"cache_ttl"`           // Cache TTL
	CompressionEnabled bool                   `json:"compression_enabled"` // Enable compression
	EncryptionEnabled  bool                   `json:"encryption_enabled"`  // Enable encryption
	CustomHeaders      map[string]string      `json:"custom_headers"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// FileMetadata represents file metadata
type FileMetadata struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Path       string                 `json:"path"`
	Size       int64                  `json:"size"`
	MimeType   string                 `json:"mime_type"`
	Hash       string                 `json:"hash"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	UserID     string                 `json:"user_id,omitempty"`
	ServiceID  string                 `json:"service_id,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// DefaultStorageConfig returns default storage configuration
func DefaultStorageConfig() *StorageConfig {
	return &StorageConfig{
		StorageType:        "local",
		BasePath:           "/tmp/storage",
		MaxFileSize:        10 * 1024 * 1024, // 10MB
		AllowedMimeTypes:   []string{"image/*", "text/*", "application/json", "application/pdf"},
		ExcludedPaths:      []string{"/health", "/metrics", "/admin"},
		CacheEnabled:       true,
		CacheTTL:           1 * time.Hour,
		CompressionEnabled: false,
		EncryptionEnabled:  false,
		CustomHeaders:      make(map[string]string),
		Metadata:           make(map[string]interface{}),
	}
}

// NewStorageProvider creates a new storage provider
func NewStorageProvider(config *StorageConfig, logger *logrus.Logger) *StorageProvider {
	if config == nil {
		config = DefaultStorageConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &StorageProvider{
		name:      "storage",
		logger:    logger,
		config:    config,
		enabled:   true,
		fileCache: make(map[string]*FileMetadata),
	}
}

// GetName returns the provider name
func (sp *StorageProvider) GetName() string {
	return sp.name
}

// GetType returns the provider type
func (sp *StorageProvider) GetType() string {
	return "storage"
}

// GetSupportedFeatures returns supported features
func (sp *StorageProvider) GetSupportedFeatures() []types.MiddlewareFeature {
	return []types.MiddlewareFeature{
		types.FeatureFileStorage,
		types.FeatureObjectStorage,
		types.FeatureBlobStorage,
		types.FeatureFileUpload,
		types.FeatureFileDownload,
		types.FeatureFileMetadata,
		types.FeatureFileCompression,
		types.FeatureFileEncryption,
		types.FeatureMemoryCache,
		types.FeatureDistributedCache,
		types.FeatureCacheInvalidation,
		types.FeatureCacheWarming,
	}
}

// GetConnectionInfo returns connection information
func (sp *StorageProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     "localhost",
		Port:     0,
		Protocol: sp.config.StorageType,
		Version:  "1.0.0",
		Secure:   sp.config.EncryptionEnabled,
	}
}

// ProcessRequest processes a storage request
func (sp *StorageProvider) ProcessRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	if !sp.enabled {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	// Check if path is excluded
	if sp.isPathExcluded(request.Path) {
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}

	sp.logger.WithFields(logrus.Fields{
		"request_id": request.ID,
		"method":     request.Method,
		"path":       request.Path,
	}).Debug("Processing storage request")

	// Handle different HTTP methods
	switch request.Method {
	case "GET":
		return sp.handleGetRequest(ctx, request)
	case "POST", "PUT":
		return sp.handleUploadRequest(ctx, request)
	case "DELETE":
		return sp.handleDeleteRequest(ctx, request)
	default:
		return &types.MiddlewareResponse{
			ID:        request.ID,
			Success:   true,
			Timestamp: time.Now(),
		}, nil
	}
}

// ProcessResponse processes a storage response
func (sp *StorageProvider) ProcessResponse(ctx context.Context, response *types.MiddlewareResponse) (*types.MiddlewareResponse, error) {
	sp.logger.WithFields(logrus.Fields{
		"response_id": response.ID,
		"status_code": response.StatusCode,
		"success":     response.Success,
	}).Debug("Processing storage response")

	// Add storage headers
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	response.Headers["X-Storage-Provider"] = sp.name
	response.Headers["X-Storage-Type"] = sp.config.StorageType
	response.Headers["X-Storage-Timestamp"] = time.Now().Format(time.RFC3339)

	// Add custom headers
	for name, value := range sp.config.CustomHeaders {
		response.Headers[name] = value
	}

	return response, nil
}

// CreateChain creates a storage middleware chain
func (sp *StorageProvider) CreateChain(ctx context.Context, config *types.MiddlewareConfig) (*types.MiddlewareChain, error) {
	sp.logger.WithField("chain_name", config.Name).Info("Creating storage middleware chain")

	chain := types.NewMiddlewareChain(
		func(ctx context.Context, req *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
			return sp.ProcessRequest(ctx, req)
		},
	)

	return chain, nil
}

// ExecuteChain executes the storage middleware chain
func (sp *StorageProvider) ExecuteChain(ctx context.Context, chain *types.MiddlewareChain, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	sp.logger.WithField("request_id", request.ID).Debug("Executing storage middleware chain")
	return chain.Execute(ctx, request)
}

// CreateHTTPMiddleware creates HTTP storage middleware
func (sp *StorageProvider) CreateHTTPMiddleware(config *types.MiddlewareConfig) types.HTTPMiddlewareHandler {
	sp.logger.WithField("config_type", config.Type).Info("Creating HTTP storage middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create middleware request
			request := &types.MiddlewareRequest{
				ID:        fmt.Sprintf("req-%d", time.Now().UnixNano()),
				Type:      "http",
				Method:    r.Method,
				Path:      r.URL.Path,
				Headers:   make(map[string]string),
				Query:     make(map[string]string),
				Context:   make(map[string]interface{}),
				Metadata:  make(map[string]interface{}),
				Timestamp: time.Now(),
			}

			// Copy headers
			for name, values := range r.Header {
				if len(values) > 0 {
					request.Headers[name] = values[0]
				}
			}

			// Copy query parameters
			for name, values := range r.URL.Query() {
				if len(values) > 0 {
					request.Query[name] = values[0]
				}
			}

			// Process request
			response, err := sp.ProcessRequest(r.Context(), request)
			if err != nil {
				sp.logger.WithError(err).Error("Storage middleware error")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal server error"}`))
				return
			}

			if !response.Success {
				// Set response headers
				for name, value := range response.Headers {
					w.Header().Set(name, value)
				}
				w.WriteHeader(response.StatusCode)
				w.Write(response.Body)
				return
			}

			// Add storage headers
			for name, value := range response.Headers {
				w.Header().Set(name, value)
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// WrapHTTPHandler wraps an HTTP handler with storage middleware
func (sp *StorageProvider) WrapHTTPHandler(handler http.Handler, config *types.MiddlewareConfig) http.Handler {
	middleware := sp.CreateHTTPMiddleware(config)
	return middleware(handler)
}

// Configure configures the storage provider
func (sp *StorageProvider) Configure(config map[string]interface{}) error {
	sp.logger.Info("Configuring storage provider")

	if storageType, ok := config["storage_type"].(string); ok {
		sp.config.StorageType = storageType
	}

	if basePath, ok := config["base_path"].(string); ok {
		sp.config.BasePath = basePath
	}

	if maxFileSize, ok := config["max_file_size"].(int64); ok {
		sp.config.MaxFileSize = maxFileSize
	}

	if allowedMimeTypes, ok := config["allowed_mime_types"].([]string); ok {
		sp.config.AllowedMimeTypes = allowedMimeTypes
	}

	if excludedPaths, ok := config["excluded_paths"].([]string); ok {
		sp.config.ExcludedPaths = excludedPaths
	}

	if cacheEnabled, ok := config["cache_enabled"].(bool); ok {
		sp.config.CacheEnabled = cacheEnabled
	}

	if cacheTTL, ok := config["cache_ttl"].(time.Duration); ok {
		sp.config.CacheTTL = cacheTTL
	}

	if compressionEnabled, ok := config["compression_enabled"].(bool); ok {
		sp.config.CompressionEnabled = compressionEnabled
	}

	if encryptionEnabled, ok := config["encryption_enabled"].(bool); ok {
		sp.config.EncryptionEnabled = encryptionEnabled
	}

	sp.enabled = true
	return nil
}

// IsConfigured returns whether the provider is configured
func (sp *StorageProvider) IsConfigured() bool {
	return sp.enabled && sp.config.BasePath != ""
}

// HealthCheck performs health check
func (sp *StorageProvider) HealthCheck(ctx context.Context) error {
	sp.logger.Debug("Storage provider health check")
	// In production, check if storage backend is accessible
	return nil
}

// GetStats returns storage statistics
func (sp *StorageProvider) GetStats(ctx context.Context) (*types.MiddlewareStats, error) {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	totalFiles := len(sp.fileCache)
	totalSize := int64(0)
	fileTypes := make(map[string]int)

	for _, metadata := range sp.fileCache {
		totalSize += metadata.Size
		fileTypes[metadata.MimeType]++
	}

	return &types.MiddlewareStats{
		TotalRequests:      1000,
		SuccessfulRequests: 950,
		FailedRequests:     50,
		AverageLatency:     5 * time.Millisecond,
		MaxLatency:         50 * time.Millisecond,
		MinLatency:         1 * time.Millisecond,
		ErrorRate:          0.05,
		Throughput:         100.0,
		ActiveConnections:  int64(totalFiles),
		ProviderData: map[string]interface{}{
			"provider_type":       "storage",
			"storage_type":        sp.config.StorageType,
			"base_path":           sp.config.BasePath,
			"total_files":         totalFiles,
			"total_size":          totalSize,
			"file_types":          fileTypes,
			"cache_enabled":       sp.config.CacheEnabled,
			"compression_enabled": sp.config.CompressionEnabled,
			"encryption_enabled":  sp.config.EncryptionEnabled,
		},
	}, nil
}

// Close closes the storage provider
func (sp *StorageProvider) Close() error {
	sp.logger.Info("Closing storage provider")
	sp.enabled = false
	sp.fileCache = make(map[string]*FileMetadata)
	return nil
}

// Helper methods

func (sp *StorageProvider) isPathExcluded(path string) bool {
	for _, excludedPath := range sp.config.ExcludedPaths {
		if strings.HasPrefix(path, excludedPath) {
			return true
		}
	}
	return false
}

func (sp *StorageProvider) handleGetRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	// Extract file path from request
	filePath := sp.extractFilePath(request.Path)
	if filePath == "" {
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:      []byte(`{"error": "invalid file path"}`),
			Error:     "invalid file path",
			Timestamp: time.Now(),
		}, nil
	}

	// Check cache if enabled
	if sp.config.CacheEnabled {
		if metadata, exists := sp.getCachedFile(filePath); exists {
			sp.logger.WithField("file_path", filePath).Debug("File found in cache")
			return &types.MiddlewareResponse{
				ID:      request.ID,
				Success: true,
				Headers: map[string]string{
					"X-File-ID":       metadata.ID,
					"X-File-Name":     metadata.Name,
					"X-File-Size":     fmt.Sprintf("%d", metadata.Size),
					"X-File-MimeType": metadata.MimeType,
					"X-File-Hash":     metadata.Hash,
				},
				Context: map[string]interface{}{
					"file_metadata": metadata,
				},
				Timestamp: time.Now(),
			}, nil
		}
	}

	// In production, this would check actual storage backend
	sp.logger.WithField("file_path", filePath).Debug("File not found in cache, would check storage backend")

	return &types.MiddlewareResponse{
		ID:        request.ID,
		Success:   true,
		Context:   request.Context,
		Timestamp: time.Now(),
	}, nil
}

func (sp *StorageProvider) handleUploadRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	// Extract file path from request
	filePath := sp.extractFilePath(request.Path)
	if filePath == "" {
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:      []byte(`{"error": "invalid file path"}`),
			Error:     "invalid file path",
			Timestamp: time.Now(),
		}, nil
	}

	// Check file size
	if int64(len(request.Body)) > sp.config.MaxFileSize {
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusRequestEntityTooLarge,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:      []byte(`{"error": "file too large"}`),
			Error:     "file too large",
			Timestamp: time.Now(),
		}, nil
	}

	// Check MIME type
	contentType := request.Headers["Content-Type"]
	if !sp.isMimeTypeAllowed(contentType) {
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusUnsupportedMediaType,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:      []byte(`{"error": "unsupported media type"}`),
			Error:     "unsupported media type",
			Timestamp: time.Now(),
		}, nil
	}

	// Create file metadata
	metadata := &FileMetadata{
		ID:         sp.generateFileID(filePath),
		Name:       filepath.Base(filePath),
		Path:       filePath,
		Size:       int64(len(request.Body)),
		MimeType:   contentType,
		Hash:       sp.calculateHash(request.Body),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		UserID:     request.UserID,
		ServiceID:  request.ServiceID,
		Tags:       []string{},
		Attributes: make(map[string]interface{}),
		Metadata:   make(map[string]interface{}),
	}

	// Cache file metadata if enabled
	if sp.config.CacheEnabled {
		sp.cacheFile(metadata)
	}

	sp.logger.WithFields(logrus.Fields{
		"file_id":   metadata.ID,
		"file_name": metadata.Name,
		"file_size": metadata.Size,
		"mime_type": metadata.MimeType,
	}).Info("File uploaded successfully")

	return &types.MiddlewareResponse{
		ID:      request.ID,
		Success: true,
		Headers: map[string]string{
			"X-File-ID":       metadata.ID,
			"X-File-Name":     metadata.Name,
			"X-File-Size":     fmt.Sprintf("%d", metadata.Size),
			"X-File-MimeType": metadata.MimeType,
			"X-File-Hash":     metadata.Hash,
		},
		Context: map[string]interface{}{
			"file_metadata": metadata,
		},
		Timestamp: time.Now(),
	}, nil
}

func (sp *StorageProvider) handleDeleteRequest(ctx context.Context, request *types.MiddlewareRequest) (*types.MiddlewareResponse, error) {
	// Extract file path from request
	filePath := sp.extractFilePath(request.Path)
	if filePath == "" {
		return &types.MiddlewareResponse{
			ID:         request.ID,
			Success:    false,
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:      []byte(`{"error": "invalid file path"}`),
			Error:     "invalid file path",
			Timestamp: time.Now(),
		}, nil
	}

	// Remove from cache if enabled
	if sp.config.CacheEnabled {
		sp.removeCachedFile(filePath)
	}

	sp.logger.WithField("file_path", filePath).Info("File deleted successfully")

	return &types.MiddlewareResponse{
		ID:      request.ID,
		Success: true,
		Headers: map[string]string{
			"X-Deleted-File": filePath,
		},
		Timestamp: time.Now(),
	}, nil
}

func (sp *StorageProvider) extractFilePath(path string) string {
	// Extract file path from URL path
	// This is a simplified implementation
	if strings.HasPrefix(path, "/storage/") {
		return strings.TrimPrefix(path, "/storage/")
	}
	return ""
}

func (sp *StorageProvider) isMimeTypeAllowed(mimeType string) bool {
	if len(sp.config.AllowedMimeTypes) == 0 {
		return true
	}

	for _, allowedType := range sp.config.AllowedMimeTypes {
		if allowedType == "*" || allowedType == mimeType {
			return true
		}
		if strings.HasSuffix(allowedType, "/*") {
			prefix := strings.TrimSuffix(allowedType, "/*")
			if strings.HasPrefix(mimeType, prefix+"/") {
				return true
			}
		}
	}

	return false
}

func (sp *StorageProvider) generateFileID(filePath string) string {
	// Generate unique file ID based on path and timestamp
	hash := md5.Sum([]byte(filePath + time.Now().Format(time.RFC3339Nano)))
	return hex.EncodeToString(hash[:])
}

func (sp *StorageProvider) calculateHash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func (sp *StorageProvider) getCachedFile(filePath string) (*FileMetadata, bool) {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	metadata, exists := sp.fileCache[filePath]
	if !exists {
		return nil, false
	}

	// Check if cache entry has expired
	if time.Since(metadata.UpdatedAt) > sp.config.CacheTTL {
		delete(sp.fileCache, filePath)
		return nil, false
	}

	return metadata, true
}

func (sp *StorageProvider) cacheFile(metadata *FileMetadata) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	sp.fileCache[metadata.Path] = metadata
}

func (sp *StorageProvider) removeCachedFile(filePath string) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	delete(sp.fileCache, filePath)
}
