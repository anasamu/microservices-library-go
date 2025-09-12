package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// StorageManager manages multiple storage providers
type StorageManager struct {
	providers map[string]StorageProvider
	logger    *logrus.Logger
	config    *ManagerConfig
}

// ManagerConfig holds storage manager configuration
type ManagerConfig struct {
	DefaultProvider string            `json:"default_provider"`
	RetryAttempts   int               `json:"retry_attempts"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	Timeout         time.Duration     `json:"timeout"`
	MaxFileSize     int64             `json:"max_file_size"`
	AllowedTypes    []string          `json:"allowed_types"`
	Metadata        map[string]string `json:"metadata"`
}

// StorageProvider interface for storage backends
type StorageProvider interface {
	// Provider information
	GetName() string
	GetSupportedFeatures() []StorageFeature
	GetMaxFileSize() int64
	GetAllowedTypes() []string

	// Basic operations
	PutObject(ctx context.Context, request *PutObjectRequest) (*PutObjectResponse, error)
	GetObject(ctx context.Context, request *GetObjectRequest) (*GetObjectResponse, error)
	DeleteObject(ctx context.Context, request *DeleteObjectRequest) error
	DeleteObjects(ctx context.Context, request *DeleteObjectsRequest) (*DeleteObjectsResponse, error)
	ListObjects(ctx context.Context, request *ListObjectsRequest) (*ListObjectsResponse, error)
	ObjectExists(ctx context.Context, request *ObjectExistsRequest) (bool, error)

	// Advanced operations
	CopyObject(ctx context.Context, request *CopyObjectRequest) (*CopyObjectResponse, error)
	MoveObject(ctx context.Context, request *MoveObjectRequest) (*MoveObjectResponse, error)
	GetObjectInfo(ctx context.Context, request *GetObjectInfoRequest) (*ObjectInfo, error)

	// URL operations
	GeneratePresignedURL(ctx context.Context, request *PresignedURLRequest) (string, error)
	GeneratePublicURL(ctx context.Context, request *PublicURLRequest) (string, error)

	// Bucket operations
	CreateBucket(ctx context.Context, request *CreateBucketRequest) error
	DeleteBucket(ctx context.Context, request *DeleteBucketRequest) error
	BucketExists(ctx context.Context, request *BucketExistsRequest) (bool, error)
	ListBuckets(ctx context.Context) ([]BucketInfo, error)

	// Configuration
	Configure(config map[string]interface{}) error
	IsConfigured() bool
	HealthCheck(ctx context.Context) error
	Close() error
}

// StorageFeature represents a storage feature
type StorageFeature string

const (
	FeaturePresignedURLs StorageFeature = "presigned_urls"
	FeaturePublicURLs    StorageFeature = "public_urls"
	FeatureMultipart     StorageFeature = "multipart_upload"
	FeatureVersioning    StorageFeature = "versioning"
	FeatureEncryption    StorageFeature = "encryption"
	FeatureLifecycle     StorageFeature = "lifecycle"
	FeatureCORS          StorageFeature = "cors"
	FeatureCDN           StorageFeature = "cdn"
)

// PutObjectRequest represents a put object request
type PutObjectRequest struct {
	Bucket      string                 `json:"bucket"`
	Key         string                 `json:"key"`
	Content     io.Reader              `json:"-"`
	Size        int64                  `json:"size"`
	ContentType string                 `json:"content_type"`
	Metadata    map[string]string      `json:"metadata"`
	Tags        map[string]string      `json:"tags"`
	ACL         string                 `json:"acl,omitempty"`
	Encryption  *EncryptionConfig      `json:"encryption,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// PutObjectResponse represents a put object response
type PutObjectResponse struct {
	Key          string                 `json:"key"`
	ETag         string                 `json:"etag"`
	VersionID    string                 `json:"version_id,omitempty"`
	Size         int64                  `json:"size"`
	LastModified time.Time              `json:"last_modified"`
	Metadata     map[string]string      `json:"metadata"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// GetObjectRequest represents a get object request
type GetObjectRequest struct {
	Bucket    string                 `json:"bucket"`
	Key       string                 `json:"key"`
	VersionID string                 `json:"version_id,omitempty"`
	Range     *RangeSpec             `json:"range,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// GetObjectResponse represents a get object response
type GetObjectResponse struct {
	Content      io.ReadCloser          `json:"-"`
	Size         int64                  `json:"size"`
	ContentType  string                 `json:"content_type"`
	ETag         string                 `json:"etag"`
	LastModified time.Time              `json:"last_modified"`
	Metadata     map[string]string      `json:"metadata"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// DeleteObjectRequest represents a delete object request
type DeleteObjectRequest struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	VersionID string `json:"version_id,omitempty"`
}

// DeleteObjectsRequest represents a delete objects request
type DeleteObjectsRequest struct {
	Bucket string   `json:"bucket"`
	Keys   []string `json:"keys"`
}

// DeleteObjectsResponse represents a delete objects response
type DeleteObjectsResponse struct {
	Deleted []DeletedObject `json:"deleted"`
	Errors  []DeleteError   `json:"errors"`
}

// DeletedObject represents a deleted object
type DeletedObject struct {
	Key       string `json:"key"`
	VersionID string `json:"version_id,omitempty"`
	ETag      string `json:"etag,omitempty"`
}

// DeleteError represents a delete error
type DeleteError struct {
	Key     string `json:"key"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ListObjectsRequest represents a list objects request
type ListObjectsRequest struct {
	Bucket            string `json:"bucket"`
	Prefix            string `json:"prefix,omitempty"`
	Delimiter         string `json:"delimiter,omitempty"`
	MaxKeys           int    `json:"max_keys,omitempty"`
	ContinuationToken string `json:"continuation_token,omitempty"`
	StartAfter        string `json:"start_after,omitempty"`
}

// ListObjectsResponse represents a list objects response
type ListObjectsResponse struct {
	Objects               []ObjectInfo           `json:"objects"`
	CommonPrefixes        []string               `json:"common_prefixes"`
	IsTruncated           bool                   `json:"is_truncated"`
	NextContinuationToken string                 `json:"next_continuation_token,omitempty"`
	ProviderData          map[string]interface{} `json:"provider_data"`
}

// ObjectExistsRequest represents an object exists request
type ObjectExistsRequest struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	VersionID string `json:"version_id,omitempty"`
}

// CopyObjectRequest represents a copy object request
type CopyObjectRequest struct {
	SourceBucket string                 `json:"source_bucket"`
	SourceKey    string                 `json:"source_key"`
	DestBucket   string                 `json:"dest_bucket"`
	DestKey      string                 `json:"dest_key"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
	ACL          string                 `json:"acl,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// CopyObjectResponse represents a copy object response
type CopyObjectResponse struct {
	Key          string                 `json:"key"`
	ETag         string                 `json:"etag"`
	LastModified time.Time              `json:"last_modified"`
	Size         int64                  `json:"size"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// MoveObjectRequest represents a move object request
type MoveObjectRequest struct {
	SourceBucket string `json:"source_bucket"`
	SourceKey    string `json:"source_key"`
	DestBucket   string `json:"dest_bucket"`
	DestKey      string `json:"dest_key"`
}

// MoveObjectResponse represents a move object response
type MoveObjectResponse struct {
	Key          string                 `json:"key"`
	ETag         string                 `json:"etag"`
	LastModified time.Time              `json:"last_modified"`
	Size         int64                  `json:"size"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// GetObjectInfoRequest represents a get object info request
type GetObjectInfoRequest struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	VersionID string `json:"version_id,omitempty"`
}

// ObjectInfo represents object information
type ObjectInfo struct {
	Key          string                 `json:"key"`
	Size         int64                  `json:"size"`
	LastModified time.Time              `json:"last_modified"`
	ETag         string                 `json:"etag"`
	ContentType  string                 `json:"content_type"`
	Metadata     map[string]string      `json:"metadata"`
	VersionID    string                 `json:"version_id,omitempty"`
	StorageClass string                 `json:"storage_class,omitempty"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// PresignedURLRequest represents a presigned URL request
type PresignedURLRequest struct {
	Bucket  string                 `json:"bucket"`
	Key     string                 `json:"key"`
	Method  string                 `json:"method"` // GET, PUT, POST, DELETE
	Expires time.Duration          `json:"expires"`
	Headers map[string]string      `json:"headers,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// PublicURLRequest represents a public URL request
type PublicURLRequest struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

// CreateBucketRequest represents a create bucket request
type CreateBucketRequest struct {
	Bucket  string                 `json:"bucket"`
	Region  string                 `json:"region,omitempty"`
	ACL     string                 `json:"acl,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// DeleteBucketRequest represents a delete bucket request
type DeleteBucketRequest struct {
	Bucket string `json:"bucket"`
}

// BucketExistsRequest represents a bucket exists request
type BucketExistsRequest struct {
	Bucket string `json:"bucket"`
}

// BucketInfo represents bucket information
type BucketInfo struct {
	Name         string                 `json:"name"`
	CreationDate time.Time              `json:"creation_date"`
	Region       string                 `json:"region,omitempty"`
	ProviderData map[string]interface{} `json:"provider_data"`
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	Algorithm string            `json:"algorithm"`
	KeyID     string            `json:"key_id,omitempty"`
	Context   map[string]string `json:"context,omitempty"`
}

// RangeSpec represents a range specification
type RangeSpec struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// DefaultManagerConfig returns default storage manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		DefaultProvider: "s3",
		RetryAttempts:   3,
		RetryDelay:      5 * time.Second,
		Timeout:         30 * time.Second,
		MaxFileSize:     100 * 1024 * 1024, // 100MB
		AllowedTypes:    []string{"*"},     // Allow all types by default
		Metadata:        make(map[string]string),
	}
}

// NewStorageManager creates a new storage manager
func NewStorageManager(config *ManagerConfig, logger *logrus.Logger) *StorageManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &StorageManager{
		providers: make(map[string]StorageProvider),
		logger:    logger,
		config:    config,
	}
}

// RegisterProvider registers a storage provider
func (sm *StorageManager) RegisterProvider(provider StorageProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.GetName()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	sm.providers[name] = provider
	sm.logger.WithField("provider", name).Info("Storage provider registered")

	return nil
}

// GetProvider returns a storage provider by name
func (sm *StorageManager) GetProvider(name string) (StorageProvider, error) {
	provider, exists := sm.providers[name]
	if !exists {
		return nil, fmt.Errorf("storage provider not found: %s", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default storage provider
func (sm *StorageManager) GetDefaultProvider() (StorageProvider, error) {
	return sm.GetProvider(sm.config.DefaultProvider)
}

// PutObject uploads an object using the specified provider
func (sm *StorageManager) PutObject(ctx context.Context, providerName string, request *PutObjectRequest) (*PutObjectResponse, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Validate request
	if err := sm.validatePutRequest(request); err != nil {
		return nil, fmt.Errorf("invalid put request: %w", err)
	}

	// Check file size limit
	if request.Size > sm.config.MaxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", request.Size, sm.config.MaxFileSize)
	}

	// Check allowed types
	if !sm.isTypeAllowed(request.ContentType) {
		return nil, fmt.Errorf("content type %s is not allowed", request.ContentType)
	}

	// Set default values
	if request.Key == "" {
		request.Key = uuid.New().String()
	}

	// Upload with retry logic
	var response *PutObjectResponse
	for attempt := 1; attempt <= sm.config.RetryAttempts; attempt++ {
		response, err = provider.PutObject(ctx, request)
		if err == nil {
			break
		}

		sm.logger.WithError(err).WithFields(logrus.Fields{
			"provider": providerName,
			"attempt":  attempt,
			"key":      request.Key,
		}).Warn("Object upload failed, retrying")

		if attempt < sm.config.RetryAttempts {
			time.Sleep(sm.config.RetryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to upload object after %d attempts: %w", sm.config.RetryAttempts, err)
	}

	sm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"key":      response.Key,
		"size":     response.Size,
		"bucket":   request.Bucket,
	}).Info("Object uploaded successfully")

	return response, nil
}

// GetObject downloads an object using the specified provider
func (sm *StorageManager) GetObject(ctx context.Context, providerName string, request *GetObjectRequest) (*GetObjectResponse, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.GetObject(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"key":      request.Key,
		"bucket":   request.Bucket,
	}).Debug("Object retrieved successfully")

	return response, nil
}

// DeleteObject deletes an object using the specified provider
func (sm *StorageManager) DeleteObject(ctx context.Context, providerName string, request *DeleteObjectRequest) error {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.DeleteObject(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"key":      request.Key,
		"bucket":   request.Bucket,
	}).Info("Object deleted successfully")

	return nil
}

// DeleteObjects deletes multiple objects using the specified provider
func (sm *StorageManager) DeleteObjects(ctx context.Context, providerName string, request *DeleteObjectsRequest) (*DeleteObjectsResponse, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.DeleteObjects(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to delete objects: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"bucket":   request.Bucket,
		"deleted":  len(response.Deleted),
		"errors":   len(response.Errors),
	}).Info("Objects deletion completed")

	return response, nil
}

// ListObjects lists objects using the specified provider
func (sm *StorageManager) ListObjects(ctx context.Context, providerName string, request *ListObjectsRequest) (*ListObjectsResponse, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.ListObjects(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"bucket":   request.Bucket,
		"count":    len(response.Objects),
	}).Debug("Objects listed successfully")

	return response, nil
}

// ObjectExists checks if an object exists using the specified provider
func (sm *StorageManager) ObjectExists(ctx context.Context, providerName string, request *ObjectExistsRequest) (bool, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return false, err
	}

	exists, err := provider.ObjectExists(ctx, request)
	if err != nil {
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return exists, nil
}

// CopyObject copies an object using the specified provider
func (sm *StorageManager) CopyObject(ctx context.Context, providerName string, request *CopyObjectRequest) (*CopyObjectResponse, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.CopyObject(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to copy object: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"source":   request.SourceKey,
		"dest":     request.DestKey,
	}).Info("Object copied successfully")

	return response, nil
}

// MoveObject moves an object using the specified provider
func (sm *StorageManager) MoveObject(ctx context.Context, providerName string, request *MoveObjectRequest) (*MoveObjectResponse, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	response, err := provider.MoveObject(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to move object: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"source":   request.SourceKey,
		"dest":     request.DestKey,
	}).Info("Object moved successfully")

	return response, nil
}

// GetObjectInfo gets object information using the specified provider
func (sm *StorageManager) GetObjectInfo(ctx context.Context, providerName string, request *GetObjectInfoRequest) (*ObjectInfo, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	info, err := provider.GetObjectInfo(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	return info, nil
}

// GeneratePresignedURL generates a presigned URL using the specified provider
func (sm *StorageManager) GeneratePresignedURL(ctx context.Context, providerName string, request *PresignedURLRequest) (string, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return "", err
	}

	url, err := provider.GeneratePresignedURL(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"method":   request.Method,
		"key":      request.Key,
		"expires":  request.Expires,
	}).Debug("Presigned URL generated successfully")

	return url, nil
}

// GeneratePublicURL generates a public URL using the specified provider
func (sm *StorageManager) GeneratePublicURL(ctx context.Context, providerName string, request *PublicURLRequest) (string, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return "", err
	}

	url, err := provider.GeneratePublicURL(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to generate public URL: %w", err)
	}

	return url, nil
}

// CreateBucket creates a bucket using the specified provider
func (sm *StorageManager) CreateBucket(ctx context.Context, providerName string, request *CreateBucketRequest) error {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.CreateBucket(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"bucket":   request.Bucket,
		"region":   request.Region,
	}).Info("Bucket created successfully")

	return nil
}

// DeleteBucket deletes a bucket using the specified provider
func (sm *StorageManager) DeleteBucket(ctx context.Context, providerName string, request *DeleteBucketRequest) error {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return err
	}

	err = provider.DeleteBucket(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"bucket":   request.Bucket,
	}).Info("Bucket deleted successfully")

	return nil
}

// BucketExists checks if a bucket exists using the specified provider
func (sm *StorageManager) BucketExists(ctx context.Context, providerName string, request *BucketExistsRequest) (bool, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return false, err
	}

	exists, err := provider.BucketExists(ctx, request)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	return exists, nil
}

// ListBuckets lists buckets using the specified provider
func (sm *StorageManager) ListBuckets(ctx context.Context, providerName string) ([]BucketInfo, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	buckets, err := provider.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"count":    len(buckets),
	}).Debug("Buckets listed successfully")

	return buckets, nil
}

// GetSupportedProviders returns a list of registered providers
func (sm *StorageManager) GetSupportedProviders() []string {
	providers := make([]string, 0, len(sm.providers))
	for name := range sm.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetProviderCapabilities returns capabilities of a provider
func (sm *StorageManager) GetProviderCapabilities(providerName string) ([]StorageFeature, int64, []string, error) {
	provider, err := sm.GetProvider(providerName)
	if err != nil {
		return nil, 0, nil, err
	}

	return provider.GetSupportedFeatures(), provider.GetMaxFileSize(), provider.GetAllowedTypes(), nil
}

// HealthCheck performs health check on all providers
func (sm *StorageManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for name, provider := range sm.providers {
		err := provider.HealthCheck(ctx)
		results[name] = err
	}

	return results
}

// validatePutRequest validates a put object request
func (sm *StorageManager) validatePutRequest(request *PutObjectRequest) error {
	if request.Bucket == "" {
		return fmt.Errorf("bucket is required")
	}

	if request.Content == nil {
		return fmt.Errorf("content is required")
	}

	if request.Size <= 0 {
		return fmt.Errorf("size must be positive")
	}

	return nil
}

// isTypeAllowed checks if a content type is allowed
func (sm *StorageManager) isTypeAllowed(contentType string) bool {
	// If no restrictions, allow all
	if len(sm.config.AllowedTypes) == 0 {
		return true
	}

	// Check if wildcard is allowed
	for _, allowedType := range sm.config.AllowedTypes {
		if allowedType == "*" {
			return true
		}
		if allowedType == contentType {
			return true
		}
		// Check for wildcard patterns like "image/*"
		if len(allowedType) > 1 && allowedType[len(allowedType)-1] == '*' {
			prefix := allowedType[:len(allowedType)-1]
			if len(contentType) >= len(prefix) && contentType[:len(prefix)] == prefix {
				return true
			}
		}
	}

	return false
}
