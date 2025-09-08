package gcs

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/anasamu/microservices-library-go/storage/gateway"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Provider implements StorageProvider for Google Cloud Storage
type Provider struct {
	client *storage.Client
	config map[string]interface{}
	logger *logrus.Logger
}

// NewProvider creates a new GCS storage provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "gcs"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.StorageFeature {
	return []gateway.StorageFeature{
		gateway.FeaturePresignedURLs,
		gateway.FeaturePublicURLs,
		gateway.FeatureMultipart,
		gateway.FeatureVersioning,
		gateway.FeatureEncryption,
		gateway.FeatureLifecycle,
		gateway.FeatureCORS,
	}
}

// GetMaxFileSize returns maximum file size (5TB for GCS)
func (p *Provider) GetMaxFileSize() int64 {
	return 5 * 1024 * 1024 * 1024 * 1024 // 5TB
}

// GetAllowedTypes returns allowed content types
func (p *Provider) GetAllowedTypes() []string {
	return []string{"*"} // GCS allows all types
}

// Configure configures the GCS provider
func (p *Provider) Configure(config map[string]interface{}) error {
	projectID, ok := config["project_id"].(string)
	if !ok || projectID == "" {
		return fmt.Errorf("gcs project_id is required")
	}

	// Create client options
	var opts []option.ClientOption

	// Add credentials if provided
	if credentialsPath, ok := config["credentials_path"].(string); ok && credentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsPath))
	} else if credentialsJSON, ok := config["credentials_json"].(string); ok && credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
	}

	// Create GCS client
	client, err := storage.NewClient(context.Background(), opts...)
	if err != nil {
		return fmt.Errorf("failed to create GCS client: %w", err)
	}

	p.client = client
	p.config = config

	p.logger.Info("GCS provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.client != nil
}

// PutObject uploads an object to GCS
func (p *Provider) PutObject(ctx context.Context, request *gateway.PutObjectRequest) (*gateway.PutObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("gcs provider not configured")
	}

	// Get bucket
	bucket := p.client.Bucket(request.Bucket)
	obj := bucket.Object(request.Key)

	// Create writer
	writer := obj.NewWriter(ctx)
	writer.ContentType = request.ContentType

	// Add metadata
	if request.Metadata != nil {
		writer.Metadata = request.Metadata
	}

	// Add ACL
	if request.ACL != "" {
		writer.PredefinedACL = request.ACL
	}

	// Copy content
	_, err := io.Copy(writer, request.Content)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to copy content: %w", err)
	}

	// Close writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Get object attributes
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object attributes: %w", err)
	}

	response := &gateway.PutObjectResponse{
		Key:          request.Key,
		ETag:         hex.EncodeToString(attrs.MD5),
		Size:         attrs.Size,
		LastModified: attrs.Updated,
		Metadata:     attrs.Metadata,
		ProviderData: map[string]interface{}{
			"generation": attrs.Generation,
		},
	}

	return response, nil
}

// GetObject downloads an object from GCS
func (p *Provider) GetObject(ctx context.Context, request *gateway.GetObjectRequest) (*gateway.GetObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("gcs provider not configured")
	}

	// Get bucket and object
	bucket := p.client.Bucket(request.Bucket)
	obj := bucket.Object(request.Key)

	// Add generation if version ID is specified
	if request.VersionID != "" {
		obj = obj.Generation(parseGeneration(request.VersionID))
	}

	// Create reader
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}

	// Get object attributes
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		reader.Close()
		return nil, fmt.Errorf("failed to get object attributes: %w", err)
	}

	response := &gateway.GetObjectResponse{
		Content:      reader,
		Size:         attrs.Size,
		ContentType:  attrs.ContentType,
		ETag:         hex.EncodeToString(attrs.MD5),
		LastModified: attrs.Updated,
		Metadata:     attrs.Metadata,
		ProviderData: map[string]interface{}{
			"generation": attrs.Generation,
		},
	}

	return response, nil
}

// DeleteObject deletes an object from GCS
func (p *Provider) DeleteObject(ctx context.Context, request *gateway.DeleteObjectRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("gcs provider not configured")
	}

	// Get bucket and object
	bucket := p.client.Bucket(request.Bucket)
	obj := bucket.Object(request.Key)

	// Add generation if version ID is specified
	if request.VersionID != "" {
		obj = obj.Generation(parseGeneration(request.VersionID))
	}

	// Delete object
	err := obj.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// DeleteObjects deletes multiple objects from GCS
func (p *Provider) DeleteObjects(ctx context.Context, request *gateway.DeleteObjectsRequest) (*gateway.DeleteObjectsResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("gcs provider not configured")
	}

	response := &gateway.DeleteObjectsResponse{
		Deleted: make([]gateway.DeletedObject, 0, len(request.Keys)),
		Errors:  make([]gateway.DeleteError, 0),
	}

	// Delete objects one by one
	for _, key := range request.Keys {
		deleteRequest := &gateway.DeleteObjectRequest{
			Bucket: request.Bucket,
			Key:    key,
		}

		err := p.DeleteObject(ctx, deleteRequest)
		if err != nil {
			response.Errors = append(response.Errors, gateway.DeleteError{
				Key:     key,
				Code:    "DeleteFailed",
				Message: err.Error(),
			})
		} else {
			response.Deleted = append(response.Deleted, gateway.DeletedObject{
				Key: key,
			})
		}
	}

	return response, nil
}

// ListObjects lists objects in GCS
func (p *Provider) ListObjects(ctx context.Context, request *gateway.ListObjectsRequest) (*gateway.ListObjectsResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("gcs provider not configured")
	}

	// Get bucket
	bucket := p.client.Bucket(request.Bucket)

	// Create query
	query := &storage.Query{
		Prefix: request.Prefix,
	}

	// Add delimiter
	if request.Delimiter != "" {
		query.Delimiter = request.Delimiter
	}

	// List objects
	it := bucket.Objects(ctx, query)

	response := &gateway.ListObjectsResponse{
		Objects:        make([]gateway.ObjectInfo, 0),
		CommonPrefixes: make([]string, 0),
		ProviderData:   make(map[string]interface{}),
	}

	// Process objects
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate objects: %w", err)
		}

		// Check if it's a common prefix
		if attrs.Prefix != "" {
			response.CommonPrefixes = append(response.CommonPrefixes, attrs.Prefix)
			continue
		}

		// Add object info
		objInfo := gateway.ObjectInfo{
			Key:          attrs.Name,
			Size:         attrs.Size,
			LastModified: attrs.Updated,
			ETag:         hex.EncodeToString(attrs.MD5),
			ContentType:  attrs.ContentType,
			Metadata:     attrs.Metadata,
			StorageClass: string(attrs.StorageClass),
			ProviderData: map[string]interface{}{
				"generation": attrs.Generation,
			},
		}

		response.Objects = append(response.Objects, objInfo)

		// Check max keys limit
		if request.MaxKeys > 0 && len(response.Objects) >= request.MaxKeys {
			break
		}
	}

	return response, nil
}

// ObjectExists checks if an object exists in GCS
func (p *Provider) ObjectExists(ctx context.Context, request *gateway.ObjectExistsRequest) (bool, error) {
	if !p.IsConfigured() {
		return false, fmt.Errorf("gcs provider not configured")
	}

	// Get bucket and object
	bucket := p.client.Bucket(request.Bucket)
	obj := bucket.Object(request.Key)

	// Add generation if version ID is specified
	if request.VersionID != "" {
		obj = obj.Generation(parseGeneration(request.VersionID))
	}

	// Check if object exists
	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}

// CopyObject copies an object within GCS
func (p *Provider) CopyObject(ctx context.Context, request *gateway.CopyObjectRequest) (*gateway.CopyObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("gcs provider not configured")
	}

	// Get source and destination objects
	srcBucket := p.client.Bucket(request.SourceBucket)
	srcObj := srcBucket.Object(request.SourceKey)

	destBucket := p.client.Bucket(request.DestBucket)
	destObj := destBucket.Object(request.DestKey)

	// Create copier
	copier := destObj.CopierFrom(srcObj)

	// Add metadata if specified
	if request.Metadata != nil {
		copier.Metadata = request.Metadata
	}

	// Add ACL if specified
	if request.ACL != "" {
		copier.PredefinedACL = request.ACL
	}

	// Copy object
	attrs, err := copier.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to copy object: %w", err)
	}

	response := &gateway.CopyObjectResponse{
		Key:          request.DestKey,
		ETag:         hex.EncodeToString(attrs.MD5),
		LastModified: attrs.Updated,
		Size:         attrs.Size,
		ProviderData: map[string]interface{}{
			"generation": attrs.Generation,
		},
	}

	return response, nil
}

// MoveObject moves an object within GCS
func (p *Provider) MoveObject(ctx context.Context, request *gateway.MoveObjectRequest) (*gateway.MoveObjectResponse, error) {
	// Move is implemented as copy + delete
	copyRequest := &gateway.CopyObjectRequest{
		SourceBucket: request.SourceBucket,
		SourceKey:    request.SourceKey,
		DestBucket:   request.DestBucket,
		DestKey:      request.DestKey,
	}

	copyResponse, err := p.CopyObject(ctx, copyRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to copy object for move: %w", err)
	}

	// Delete source object
	deleteRequest := &gateway.DeleteObjectRequest{
		Bucket: request.SourceBucket,
		Key:    request.SourceKey,
	}

	err = p.DeleteObject(ctx, deleteRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to delete source object after move: %w", err)
	}

	response := &gateway.MoveObjectResponse{
		Key:          copyResponse.Key,
		ETag:         copyResponse.ETag,
		LastModified: copyResponse.LastModified,
		Size:         copyResponse.Size,
		ProviderData: copyResponse.ProviderData,
	}

	return response, nil
}

// GetObjectInfo gets object information from GCS
func (p *Provider) GetObjectInfo(ctx context.Context, request *gateway.GetObjectInfoRequest) (*gateway.ObjectInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("gcs provider not configured")
	}

	// Get bucket and object
	bucket := p.client.Bucket(request.Bucket)
	obj := bucket.Object(request.Key)

	// Add generation if version ID is specified
	if request.VersionID != "" {
		obj = obj.Generation(parseGeneration(request.VersionID))
	}

	// Get object attributes
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object attributes: %w", err)
	}

	info := &gateway.ObjectInfo{
		Key:          request.Key,
		Size:         attrs.Size,
		LastModified: attrs.Updated,
		ETag:         hex.EncodeToString(attrs.MD5),
		ContentType:  attrs.ContentType,
		Metadata:     attrs.Metadata,
		StorageClass: string(attrs.StorageClass),
		VersionID:    fmt.Sprintf("%d", attrs.Generation),
		ProviderData: map[string]interface{}{
			"generation": attrs.Generation,
		},
	}

	return info, nil
}

// GeneratePresignedURL generates a presigned URL for GCS
func (p *Provider) GeneratePresignedURL(ctx context.Context, request *gateway.PresignedURLRequest) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("gcs provider not configured")
	}

	// Get bucket
	bucket := p.client.Bucket(request.Bucket)

	// Generate signed URL
	opts := &storage.SignedURLOptions{
		Method:  request.Method,
		Expires: time.Now().Add(request.Expires),
	}

	// Add headers if specified
	if request.Headers != nil {
		// Convert map[string]string to []string for headers
		headers := make([]string, 0, len(request.Headers)*2)
		for key, value := range request.Headers {
			headers = append(headers, key, value)
		}
		opts.Headers = headers
	}

	url, err := bucket.SignedURL(request.Key, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

// GeneratePublicURL generates a public URL for GCS
func (p *Provider) GeneratePublicURL(ctx context.Context, request *gateway.PublicURLRequest) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("gcs provider not configured")
	}

	// Generate public URL
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", request.Bucket, request.Key)

	return url, nil
}

// CreateBucket creates a bucket in GCS
func (p *Provider) CreateBucket(ctx context.Context, request *gateway.CreateBucketRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("gcs provider not configured")
	}

	// Get project ID from config
	projectID, ok := p.config["project_id"].(string)
	if !ok || projectID == "" {
		return fmt.Errorf("project_id is required for creating buckets")
	}

	// Create bucket
	bucket := p.client.Bucket(request.Bucket)

	// Create bucket with location
	err := bucket.Create(ctx, projectID, &storage.BucketAttrs{
		Location: request.Region,
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

// DeleteBucket deletes a bucket from GCS
func (p *Provider) DeleteBucket(ctx context.Context, request *gateway.DeleteBucketRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("gcs provider not configured")
	}

	// Get bucket
	bucket := p.client.Bucket(request.Bucket)

	// Delete bucket
	err := bucket.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	return nil
}

// BucketExists checks if a bucket exists in GCS
func (p *Provider) BucketExists(ctx context.Context, request *gateway.BucketExistsRequest) (bool, error) {
	if !p.IsConfigured() {
		return false, fmt.Errorf("gcs provider not configured")
	}

	// Get bucket
	bucket := p.client.Bucket(request.Bucket)

	// Check if bucket exists
	_, err := bucket.Attrs(ctx)
	if err != nil {
		if err == storage.ErrBucketNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	return true, nil
}

// ListBuckets lists buckets in GCS
func (p *Provider) ListBuckets(ctx context.Context) ([]gateway.BucketInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("gcs provider not configured")
	}

	// Get project ID from config
	projectID, ok := p.config["project_id"].(string)
	if !ok || projectID == "" {
		return nil, fmt.Errorf("project_id is required for listing buckets")
	}

	// List buckets
	it := p.client.Buckets(ctx, projectID)

	buckets := make([]gateway.BucketInfo, 0)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate buckets: %w", err)
		}

		buckets = append(buckets, gateway.BucketInfo{
			Name:         attrs.Name,
			CreationDate: attrs.Created,
			Region:       attrs.Location,
			ProviderData: map[string]interface{}{
				"storage_class": attrs.StorageClass,
			},
		})
	}

	return buckets, nil
}

// HealthCheck performs a health check on GCS
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("gcs provider not configured")
	}

	// List buckets as a health check
	_, err := p.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("gcs health check failed: %w", err)
	}

	return nil
}

// Close closes the GCS provider
func (p *Provider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// parseGeneration parses generation from version ID
func parseGeneration(versionID string) int64 {
	// For GCS, version ID is typically the generation number
	// This is a simplified implementation
	if versionID == "" {
		return 0
	}

	// Try to parse as int64
	var gen int64
	fmt.Sscanf(versionID, "%d", &gen)
	return gen
}
