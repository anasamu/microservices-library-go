package minio

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/anasamu/microservices-library-go/storage/gateway"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/encrypt"
	"github.com/sirupsen/logrus"
)

// Provider implements StorageProvider for MinIO
type Provider struct {
	client *minio.Client
	config map[string]interface{}
	logger *logrus.Logger
}

// NewProvider creates a new MinIO storage provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "minio"
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

// GetMaxFileSize returns maximum file size (5TB for MinIO)
func (p *Provider) GetMaxFileSize() int64 {
	return 5 * 1024 * 1024 * 1024 * 1024 // 5TB
}

// GetAllowedTypes returns allowed content types
func (p *Provider) GetAllowedTypes() []string {
	return []string{"*"} // MinIO allows all types
}

// Configure configures the MinIO provider
func (p *Provider) Configure(config map[string]interface{}) error {
	endpoint, ok := config["endpoint"].(string)
	if !ok || endpoint == "" {
		return fmt.Errorf("minio endpoint is required")
	}

	accessKeyID, ok := config["access_key_id"].(string)
	if !ok || accessKeyID == "" {
		return fmt.Errorf("minio access_key_id is required")
	}

	secretAccessKey, ok := config["secret_access_key"].(string)
	if !ok || secretAccessKey == "" {
		return fmt.Errorf("minio secret_access_key is required")
	}

	useSSL, _ := config["use_ssl"].(bool)
	region, _ := config["region"].(string)

	// Create MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %w", err)
	}

	p.client = client
	p.config = config

	p.logger.Info("MinIO provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.client != nil
}

// PutObject uploads an object to MinIO
func (p *Provider) PutObject(ctx context.Context, request *gateway.PutObjectRequest) (*gateway.PutObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("minio provider not configured")
	}

	// Prepare options
	opts := minio.PutObjectOptions{
		ContentType: request.ContentType,
	}

	// Add metadata
	if request.Metadata != nil {
		opts.UserMetadata = request.Metadata
	}

	// Add server-side encryption
	if request.Encryption != nil {
		opts.ServerSideEncryption = encrypt.NewSSE()
	}

	// Upload object
	result, err := p.client.PutObject(ctx, request.Bucket, request.Key, request.Content, request.Size, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to put object: %w", err)
	}

	response := &gateway.PutObjectResponse{
		Key:          request.Key,
		ETag:         strings.Trim(result.ETag, "\""),
		Size:         result.Size,
		LastModified: result.LastModified,
		Metadata:     request.Metadata,
		ProviderData: map[string]interface{}{
			"version_id": result.VersionID,
		},
	}

	if result.VersionID != "" {
		response.VersionID = result.VersionID
	}

	return response, nil
}

// GetObject downloads an object from MinIO
func (p *Provider) GetObject(ctx context.Context, request *gateway.GetObjectRequest) (*gateway.GetObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("minio provider not configured")
	}

	// Prepare options
	opts := minio.GetObjectOptions{}

	// Add version ID if specified
	if request.VersionID != "" {
		opts.VersionID = request.VersionID
	}

	// Add range if specified
	if request.Range != nil {
		opts.SetRange(request.Range.Start, request.Range.End)
	}

	// Get object
	object, err := p.client.GetObject(ctx, request.Bucket, request.Key, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	// Get object info
	info, err := object.Stat()
	if err != nil {
		object.Close()
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	response := &gateway.GetObjectResponse{
		Content:      object,
		Size:         info.Size,
		ContentType:  info.ContentType,
		ETag:         strings.Trim(info.ETag, "\""),
		LastModified: info.LastModified,
		Metadata:     info.UserMetadata,
		ProviderData: map[string]interface{}{
			"version_id": info.VersionID,
		},
	}

	return response, nil
}

// DeleteObject deletes an object from MinIO
func (p *Provider) DeleteObject(ctx context.Context, request *gateway.DeleteObjectRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("minio provider not configured")
	}

	// Prepare options
	opts := minio.RemoveObjectOptions{}

	// Add version ID if specified
	if request.VersionID != "" {
		opts.VersionID = request.VersionID
	}

	// Delete object
	err := p.client.RemoveObject(ctx, request.Bucket, request.Key, opts)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// DeleteObjects deletes multiple objects from MinIO
func (p *Provider) DeleteObjects(ctx context.Context, request *gateway.DeleteObjectsRequest) (*gateway.DeleteObjectsResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("minio provider not configured")
	}

	// Prepare objects to delete
	objectsCh := make(chan minio.ObjectInfo, len(request.Keys))

	go func() {
		defer close(objectsCh)
		for _, key := range request.Keys {
			objectsCh <- minio.ObjectInfo{Key: key}
		}
	}()

	// Delete objects
	response := &gateway.DeleteObjectsResponse{
		Deleted: make([]gateway.DeletedObject, 0, len(request.Keys)),
		Errors:  make([]gateway.DeleteError, 0),
	}

	for err := range p.client.RemoveObjects(ctx, request.Bucket, objectsCh, minio.RemoveObjectsOptions{}) {
		if err.Err != nil {
			response.Errors = append(response.Errors, gateway.DeleteError{
				Key:     err.ObjectName,
				Code:    "DeleteFailed",
				Message: err.Err.Error(),
			})
		} else {
			response.Deleted = append(response.Deleted, gateway.DeletedObject{
				Key: err.ObjectName,
			})
		}
	}

	return response, nil
}

// ListObjects lists objects in MinIO
func (p *Provider) ListObjects(ctx context.Context, request *gateway.ListObjectsRequest) (*gateway.ListObjectsResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("minio provider not configured")
	}

	// Prepare options
	opts := minio.ListObjectsOptions{
		Prefix:    request.Prefix,
		Recursive: request.Delimiter == "",
	}

	// Add max keys
	if request.MaxKeys > 0 {
		opts.MaxKeys = request.MaxKeys
	}

	// Add start after
	if request.StartAfter != "" {
		opts.StartAfter = request.StartAfter
	}

	// List objects
	objectCh := p.client.ListObjects(ctx, request.Bucket, opts)

	response := &gateway.ListObjectsResponse{
		Objects:        make([]gateway.ObjectInfo, 0),
		CommonPrefixes: make([]string, 0),
		ProviderData:   make(map[string]interface{}),
	}

	// Process objects
	for obj := range objectCh {
		if obj.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", obj.Err)
		}

		// Add object info
		objInfo := gateway.ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
			ETag:         strings.Trim(obj.ETag, "\""),
			ContentType:  obj.ContentType,
			Metadata:     obj.UserMetadata,
			StorageClass: obj.StorageClass,
			ProviderData: map[string]interface{}{
				"version_id": obj.VersionID,
			},
		}

		if obj.VersionID != "" {
			objInfo.VersionID = obj.VersionID
		}

		response.Objects = append(response.Objects, objInfo)
	}

	return response, nil
}

// ObjectExists checks if an object exists in MinIO
func (p *Provider) ObjectExists(ctx context.Context, request *gateway.ObjectExistsRequest) (bool, error) {
	if !p.IsConfigured() {
		return false, fmt.Errorf("minio provider not configured")
	}

	// Prepare options
	opts := minio.StatObjectOptions{}

	// Add version ID if specified
	if request.VersionID != "" {
		opts.VersionID = request.VersionID
	}

	// Check if object exists
	_, err := p.client.StatObject(ctx, request.Bucket, request.Key, opts)
	if err != nil {
		// Check if it's a "not found" error
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}

// CopyObject copies an object within MinIO
func (p *Provider) CopyObject(ctx context.Context, request *gateway.CopyObjectRequest) (*gateway.CopyObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("minio provider not configured")
	}

	// Prepare source and destination
	source := minio.CopySrcOptions{
		Bucket: request.SourceBucket,
		Object: request.SourceKey,
	}

	dest := minio.CopyDestOptions{
		Bucket: request.DestBucket,
		Object: request.DestKey,
	}

	// Add metadata if specified
	if request.Metadata != nil {
		dest.UserMetadata = request.Metadata
	}

	// Copy object
	result, err := p.client.CopyObject(ctx, dest, source)
	if err != nil {
		return nil, fmt.Errorf("failed to copy object: %w", err)
	}

	response := &gateway.CopyObjectResponse{
		Key:          request.DestKey,
		ETag:         strings.Trim(result.ETag, "\""),
		LastModified: result.LastModified,
		Size:         result.Size,
		ProviderData: map[string]interface{}{
			"version_id": result.VersionID,
		},
	}

	return response, nil
}

// MoveObject moves an object within MinIO
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

// GetObjectInfo gets object information from MinIO
func (p *Provider) GetObjectInfo(ctx context.Context, request *gateway.GetObjectInfoRequest) (*gateway.ObjectInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("minio provider not configured")
	}

	// Prepare options
	opts := minio.StatObjectOptions{}

	// Add version ID if specified
	if request.VersionID != "" {
		opts.VersionID = request.VersionID
	}

	// Get object info
	info, err := p.client.StatObject(ctx, request.Bucket, request.Key, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	objInfo := &gateway.ObjectInfo{
		Key:          request.Key,
		Size:         info.Size,
		LastModified: info.LastModified,
		ETag:         strings.Trim(info.ETag, "\""),
		ContentType:  info.ContentType,
		Metadata:     info.UserMetadata,
		StorageClass: info.StorageClass,
		ProviderData: map[string]interface{}{
			"version_id": info.VersionID,
		},
	}

	if info.VersionID != "" {
		objInfo.VersionID = info.VersionID
	}

	return objInfo, nil
}

// GeneratePresignedURL generates a presigned URL for MinIO
func (p *Provider) GeneratePresignedURL(ctx context.Context, request *gateway.PresignedURLRequest) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("minio provider not configured")
	}

	// Generate presigned URL based on method
	var url *url.URL
	var err error

	switch strings.ToUpper(request.Method) {
	case "GET":
		url, err = p.client.PresignedGetObject(ctx, request.Bucket, request.Key, request.Expires, nil)
	case "PUT":
		url, err = p.client.PresignedPutObject(ctx, request.Bucket, request.Key, request.Expires)
	case "DELETE":
		url, err = p.client.Presign(ctx, "DELETE", request.Bucket, request.Key, request.Expires, nil)
	default:
		return "", fmt.Errorf("unsupported method: %s", request.Method)
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// GeneratePublicURL generates a public URL for MinIO
func (p *Provider) GeneratePublicURL(ctx context.Context, request *gateway.PublicURLRequest) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("minio provider not configured")
	}

	// Get endpoint from config
	endpoint, _ := p.config["endpoint"].(string)
	useSSL, _ := p.config["use_ssl"].(bool)

	// Generate public URL
	protocol := "http"
	if useSSL {
		protocol = "https"
	}

	url := fmt.Sprintf("%s://%s/%s/%s", protocol, endpoint, request.Bucket, request.Key)

	return url, nil
}

// CreateBucket creates a bucket in MinIO
func (p *Provider) CreateBucket(ctx context.Context, request *gateway.CreateBucketRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("minio provider not configured")
	}

	// Create bucket
	err := p.client.MakeBucket(ctx, request.Bucket, minio.MakeBucketOptions{
		Region: request.Region,
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

// DeleteBucket deletes a bucket from MinIO
func (p *Provider) DeleteBucket(ctx context.Context, request *gateway.DeleteBucketRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("minio provider not configured")
	}

	// Delete bucket
	err := p.client.RemoveBucket(ctx, request.Bucket)
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	return nil
}

// BucketExists checks if a bucket exists in MinIO
func (p *Provider) BucketExists(ctx context.Context, request *gateway.BucketExistsRequest) (bool, error) {
	if !p.IsConfigured() {
		return false, fmt.Errorf("minio provider not configured")
	}

	// Check if bucket exists
	exists, err := p.client.BucketExists(ctx, request.Bucket)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	return exists, nil
}

// ListBuckets lists buckets in MinIO
func (p *Provider) ListBuckets(ctx context.Context) ([]gateway.BucketInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("minio provider not configured")
	}

	// List buckets
	buckets, err := p.client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	// Prepare response
	bucketInfos := make([]gateway.BucketInfo, len(buckets))
	for i, bucket := range buckets {
		bucketInfos[i] = gateway.BucketInfo{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate,
			ProviderData: map[string]interface{}{},
		}
	}

	return bucketInfos, nil
}

// HealthCheck performs a health check on MinIO
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("minio provider not configured")
	}

	// List buckets as a health check
	_, err := p.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("minio health check failed: %w", err)
	}

	return nil
}

// Close closes the MinIO provider
func (p *Provider) Close() error {
	// MinIO client doesn't need explicit closing
	return nil
}
