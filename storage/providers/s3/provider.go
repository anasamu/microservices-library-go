package s3

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/storage/gateway"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/sirupsen/logrus"
)

// Provider implements StorageProvider for AWS S3
type Provider struct {
	client *s3.Client
	config map[string]interface{}
	logger *logrus.Logger
}

// NewProvider creates a new S3 storage provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "s3"
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

// GetMaxFileSize returns maximum file size (5TB for S3)
func (p *Provider) GetMaxFileSize() int64 {
	return 5 * 1024 * 1024 * 1024 * 1024 // 5TB
}

// GetAllowedTypes returns allowed content types
func (p *Provider) GetAllowedTypes() []string {
	return []string{"*"} // S3 allows all types
}

// Configure configures the S3 provider
func (p *Provider) Configure(config map[string]interface{}) error {
	region, ok := config["region"].(string)
	if !ok || region == "" {
		region = "us-east-1" // Default region
	}

	accessKeyID, ok := config["access_key_id"].(string)
	if !ok || accessKeyID == "" {
		return fmt.Errorf("s3 access_key_id is required")
	}

	secretAccessKey, ok := config["secret_access_key"].(string)
	if !ok || secretAccessKey == "" {
		return fmt.Errorf("s3 secret_access_key is required")
	}

	endpoint, _ := config["endpoint"].(string)
	useSSL, _ := config["use_ssl"].(bool)

	// Create AWS config
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s3Config := &aws.Config{
		Region:   aws.String(region),
		Endpoint: aws.String(endpoint),
	}

	if endpoint != "" {
		s3Config.UsePathStyle = aws.Bool(true) // For MinIO and other S3-compatible services
	}

	p.client = s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
		if !useSSL {
			o.UseHTTPS = false
		}
	})

	p.config = config

	p.logger.Info("S3 provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.client != nil
}

// PutObject uploads an object to S3
func (p *Provider) PutObject(ctx context.Context, request *gateway.PutObjectRequest) (*gateway.PutObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("s3 provider not configured")
	}

	// Prepare input
	input := &s3.PutObjectInput{
		Bucket:      aws.String(request.Bucket),
		Key:         aws.String(request.Key),
		Body:        request.Content,
		ContentType: aws.String(request.ContentType),
	}

	// Add metadata
	if request.Metadata != nil {
		input.Metadata = request.Metadata
	}

	// Add ACL
	if request.ACL != "" {
		input.ACL = types.ObjectCannedACL(request.ACL)
	}

	// Add server-side encryption
	if request.Encryption != nil {
		input.ServerSideEncryption = types.ServerSideEncryption(request.Encryption.Algorithm)
		if request.Encryption.KeyID != "" {
			input.SSEKMSKeyId = aws.String(request.Encryption.KeyID)
		}
	}

	// Upload object
	result, err := p.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to put object: %w", err)
	}

	response := &gateway.PutObjectResponse{
		Key:          request.Key,
		ETag:         strings.Trim(*result.ETag, "\""),
		Size:         request.Size,
		LastModified: time.Now(),
		Metadata:     request.Metadata,
		ProviderData: map[string]interface{}{
			"version_id": result.VersionId,
		},
	}

	if result.VersionId != nil {
		response.VersionID = *result.VersionId
	}

	return response, nil
}

// GetObject downloads an object from S3
func (p *Provider) GetObject(ctx context.Context, request *gateway.GetObjectRequest) (*gateway.GetObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("s3 provider not configured")
	}

	// Prepare input
	input := &s3.GetObjectInput{
		Bucket: aws.String(request.Bucket),
		Key:    aws.String(request.Key),
	}

	// Add version ID if specified
	if request.VersionID != "" {
		input.VersionId = aws.String(request.VersionID)
	}

	// Add range if specified
	if request.Range != nil {
		rangeHeader := fmt.Sprintf("bytes=%d-%d", request.Range.Start, request.Range.End)
		input.Range = aws.String(rangeHeader)
	}

	// Get object
	result, err := p.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	response := &gateway.GetObjectResponse{
		Content:      result.Body,
		Size:         *result.ContentLength,
		ContentType:  aws.ToString(result.ContentType),
		ETag:         strings.Trim(aws.ToString(result.ETag), "\""),
		LastModified: aws.ToTime(result.LastModified),
		Metadata:     result.Metadata,
		ProviderData: map[string]interface{}{
			"version_id": result.VersionId,
		},
	}

	return response, nil
}

// DeleteObject deletes an object from S3
func (p *Provider) DeleteObject(ctx context.Context, request *gateway.DeleteObjectRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("s3 provider not configured")
	}

	// Prepare input
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(request.Bucket),
		Key:    aws.String(request.Key),
	}

	// Add version ID if specified
	if request.VersionID != "" {
		input.VersionId = aws.String(request.VersionID)
	}

	// Delete object
	_, err := p.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// DeleteObjects deletes multiple objects from S3
func (p *Provider) DeleteObjects(ctx context.Context, request *gateway.DeleteObjectsRequest) (*gateway.DeleteObjectsResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("s3 provider not configured")
	}

	// Prepare objects to delete
	objects := make([]types.ObjectIdentifier, len(request.Keys))
	for i, key := range request.Keys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(key)}
	}

	// Prepare input
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(request.Bucket),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(false),
		},
	}

	// Delete objects
	result, err := p.client.DeleteObjects(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to delete objects: %w", err)
	}

	// Prepare response
	response := &gateway.DeleteObjectsResponse{
		Deleted: make([]gateway.DeletedObject, len(result.Deleted)),
		Errors:  make([]gateway.DeleteError, len(result.Errors)),
	}

	// Process deleted objects
	for i, deleted := range result.Deleted {
		response.Deleted[i] = gateway.DeletedObject{
			Key:       aws.ToString(deleted.Key),
			VersionID: aws.ToString(deleted.VersionId),
			ETag:      strings.Trim(aws.ToString(deleted.ETag), "\""),
		}
	}

	// Process errors
	for i, err := range result.Errors {
		response.Errors[i] = gateway.DeleteError{
			Key:     aws.ToString(err.Key),
			Code:    aws.ToString(err.Code),
			Message: aws.ToString(err.Message),
		}
	}

	return response, nil
}

// ListObjects lists objects in S3
func (p *Provider) ListObjects(ctx context.Context, request *gateway.ListObjectsRequest) (*gateway.ListObjectsResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("s3 provider not configured")
	}

	// Prepare input
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(request.Bucket),
	}

	// Add prefix
	if request.Prefix != "" {
		input.Prefix = aws.String(request.Prefix)
	}

	// Add delimiter
	if request.Delimiter != "" {
		input.Delimiter = aws.String(request.Delimiter)
	}

	// Add max keys
	if request.MaxKeys > 0 {
		input.MaxKeys = aws.Int32(int32(request.MaxKeys))
	}

	// Add continuation token
	if request.ContinuationToken != "" {
		input.ContinuationToken = aws.String(request.ContinuationToken)
	}

	// Add start after
	if request.StartAfter != "" {
		input.StartAfter = aws.String(request.StartAfter)
	}

	// List objects
	result, err := p.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	// Prepare response
	response := &gateway.ListObjectsResponse{
		Objects:        make([]gateway.ObjectInfo, len(result.Contents)),
		CommonPrefixes: make([]string, len(result.CommonPrefixes)),
		IsTruncated:    aws.ToBool(result.IsTruncated),
		ProviderData:   make(map[string]interface{}),
	}

	// Process objects
	for i, obj := range result.Contents {
		response.Objects[i] = gateway.ObjectInfo{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			LastModified: aws.ToTime(obj.LastModified),
			ETag:         strings.Trim(aws.ToString(obj.ETag), "\""),
			StorageClass: string(obj.StorageClass),
		}
	}

	// Process common prefixes
	for i, prefix := range result.CommonPrefixes {
		response.CommonPrefixes[i] = aws.ToString(prefix.Prefix)
	}

	// Add continuation token
	if result.NextContinuationToken != nil {
		response.NextContinuationToken = *result.NextContinuationToken
	}

	return response, nil
}

// ObjectExists checks if an object exists in S3
func (p *Provider) ObjectExists(ctx context.Context, request *gateway.ObjectExistsRequest) (bool, error) {
	if !p.IsConfigured() {
		return false, fmt.Errorf("s3 provider not configured")
	}

	// Prepare input
	input := &s3.HeadObjectInput{
		Bucket: aws.String(request.Bucket),
		Key:    aws.String(request.Key),
	}

	// Add version ID if specified
	if request.VersionID != "" {
		input.VersionId = aws.String(request.VersionID)
	}

	// Check if object exists
	_, err := p.client.HeadObject(ctx, input)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}

// CopyObject copies an object within S3
func (p *Provider) CopyObject(ctx context.Context, request *gateway.CopyObjectRequest) (*gateway.CopyObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("s3 provider not configured")
	}

	// Prepare copy source
	copySource := fmt.Sprintf("%s/%s", request.SourceBucket, request.SourceKey)

	// Prepare input
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(request.DestBucket),
		Key:        aws.String(request.DestKey),
		CopySource: aws.String(copySource),
	}

	// Add metadata
	if request.Metadata != nil {
		input.Metadata = request.Metadata
		input.MetadataDirective = types.MetadataDirectiveReplace
	}

	// Add ACL
	if request.ACL != "" {
		input.ACL = types.ObjectCannedACL(request.ACL)
	}

	// Copy object
	result, err := p.client.CopyObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to copy object: %w", err)
	}

	response := &gateway.CopyObjectResponse{
		Key:          request.DestKey,
		ETag:         strings.Trim(aws.ToString(result.CopyObjectResult.ETag), "\""),
		LastModified: time.Now(),
		ProviderData: map[string]interface{}{
			"version_id": result.VersionId,
		},
	}

	return response, nil
}

// MoveObject moves an object within S3
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
		ProviderData: copyResponse.ProviderData,
	}

	return response, nil
}

// GetObjectInfo gets object information from S3
func (p *Provider) GetObjectInfo(ctx context.Context, request *gateway.GetObjectInfoRequest) (*gateway.ObjectInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("s3 provider not configured")
	}

	// Prepare input
	input := &s3.HeadObjectInput{
		Bucket: aws.String(request.Bucket),
		Key:    aws.String(request.Key),
	}

	// Add version ID if specified
	if request.VersionID != "" {
		input.VersionId = aws.String(request.VersionID)
	}

	// Get object info
	result, err := p.client.HeadObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	info := &gateway.ObjectInfo{
		Key:          request.Key,
		Size:         aws.ToInt64(result.ContentLength),
		LastModified: aws.ToTime(result.LastModified),
		ETag:         strings.Trim(aws.ToString(result.ETag), "\""),
		ContentType:  aws.ToString(result.ContentType),
		Metadata:     result.Metadata,
		StorageClass: string(result.StorageClass),
		ProviderData: map[string]interface{}{
			"version_id": result.VersionId,
		},
	}

	if result.VersionId != nil {
		info.VersionID = *result.VersionId
	}

	return info, nil
}

// GeneratePresignedURL generates a presigned URL for S3
func (p *Provider) GeneratePresignedURL(ctx context.Context, request *gateway.PresignedURLRequest) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("s3 provider not configured")
	}

	// Create presign client
	presignClient := s3.NewPresignClient(p.client)

	// Prepare input based on method
	var input interface{}
	var presignOptions func(*s3.PresignOptions)

	switch strings.ToUpper(request.Method) {
	case "GET":
		getInput := &s3.GetObjectInput{
			Bucket: aws.String(request.Bucket),
			Key:    aws.String(request.Key),
		}
		input = getInput
		presignOptions = func(o *s3.PresignOptions) {
			o.Expires = request.Expires
		}

	case "PUT":
		putInput := &s3.PutObjectInput{
			Bucket: aws.String(request.Bucket),
			Key:    aws.String(request.Key),
		}
		input = putInput
		presignOptions = func(o *s3.PresignOptions) {
			o.Expires = request.Expires
		}

	case "DELETE":
		deleteInput := &s3.DeleteObjectInput{
			Bucket: aws.String(request.Bucket),
			Key:    aws.String(request.Key),
		}
		input = deleteInput
		presignOptions = func(o *s3.PresignOptions) {
			o.Expires = request.Expires
		}

	default:
		return "", fmt.Errorf("unsupported method: %s", request.Method)
	}

	// Generate presigned URL
	var presignResult *s3.PresignedHTTPRequest
	var err error

	switch req := input.(type) {
	case *s3.GetObjectInput:
		presignResult, err = presignClient.PresignGetObject(ctx, req, presignOptions)
	case *s3.PutObjectInput:
		presignResult, err = presignClient.PresignPutObject(ctx, req, presignOptions)
	case *s3.DeleteObjectInput:
		presignResult, err = presignClient.PresignDeleteObject(ctx, req, presignOptions)
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignResult.URL, nil
}

// GeneratePublicURL generates a public URL for S3
func (p *Provider) GeneratePublicURL(ctx context.Context, request *gateway.PublicURLRequest) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("s3 provider not configured")
	}

	// Get region from config
	region, _ := p.config["region"].(string)
	if region == "" {
		region = "us-east-1"
	}

	// Generate public URL
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", request.Bucket, region, request.Key)

	return url, nil
}

// CreateBucket creates a bucket in S3
func (p *Provider) CreateBucket(ctx context.Context, request *gateway.CreateBucketRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("s3 provider not configured")
	}

	// Prepare input
	input := &s3.CreateBucketInput{
		Bucket: aws.String(request.Bucket),
	}

	// Add region if specified
	if request.Region != "" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(request.Region),
		}
	}

	// Add ACL if specified
	if request.ACL != "" {
		input.ACL = types.BucketCannedACL(request.ACL)
	}

	// Create bucket
	_, err := p.client.CreateBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

// DeleteBucket deletes a bucket from S3
func (p *Provider) DeleteBucket(ctx context.Context, request *gateway.DeleteBucketRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("s3 provider not configured")
	}

	// Prepare input
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(request.Bucket),
	}

	// Delete bucket
	_, err := p.client.DeleteBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	return nil
}

// BucketExists checks if a bucket exists in S3
func (p *Provider) BucketExists(ctx context.Context, request *gateway.BucketExistsRequest) (bool, error) {
	if !p.IsConfigured() {
		return false, fmt.Errorf("s3 provider not configured")
	}

	// Prepare input
	input := &s3.HeadBucketInput{
		Bucket: aws.String(request.Bucket),
	}

	// Check if bucket exists
	_, err := p.client.HeadBucket(ctx, input)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "NoSuchBucket") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	return true, nil
}

// ListBuckets lists buckets in S3
func (p *Provider) ListBuckets(ctx context.Context) ([]gateway.BucketInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("s3 provider not configured")
	}

	// List buckets
	result, err := p.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	// Prepare response
	buckets := make([]gateway.BucketInfo, len(result.Buckets))
	for i, bucket := range result.Buckets {
		buckets[i] = gateway.BucketInfo{
			Name:         aws.ToString(bucket.Name),
			CreationDate: aws.ToTime(bucket.CreationDate),
		}
	}

	return buckets, nil
}

// HealthCheck performs a health check on S3
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("s3 provider not configured")
	}

	// List buckets as a health check
	_, err := p.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("s3 health check failed: %w", err)
	}

	return nil
}

// Close closes the S3 provider
func (p *Provider) Close() error {
	// S3 client doesn't need explicit closing
	return nil
}
