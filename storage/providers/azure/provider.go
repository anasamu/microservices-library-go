package azure

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/anasamu/microservices-library-go/storage/types"
	"github.com/sirupsen/logrus"
)

// Provider implements StorageProvider for Azure Blob Storage
type Provider struct {
	client *azblob.Client
	config map[string]interface{}
	logger *logrus.Logger
}

// NewProvider creates a new Azure Blob Storage provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "azure"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []types.StorageFeature {
	return []types.StorageFeature{
		types.FeaturePresignedURLs,
		types.FeaturePublicURLs,
		types.FeatureMultipart,
		types.FeatureVersioning,
		types.FeatureEncryption,
		types.FeatureLifecycle,
		types.FeatureCORS,
	}
}

// GetMaxFileSize returns maximum file size (5TB for Azure Blob)
func (p *Provider) GetMaxFileSize() int64 {
	return 5 * 1024 * 1024 * 1024 * 1024 // 5TB
}

// GetAllowedTypes returns allowed content types
func (p *Provider) GetAllowedTypes() []string {
	return []string{"*"} // Azure Blob allows all types
}

// Configure configures the Azure Blob Storage provider
func (p *Provider) Configure(config map[string]interface{}) error {
	accountName, ok := config["account_name"].(string)
	if !ok || accountName == "" {
		return fmt.Errorf("azure account_name is required")
	}

	// Get credentials and create client
	var client *azblob.Client

	// Try different authentication methods
	if accountKey, ok := config["account_key"].(string); ok && accountKey != "" {
		// Use account key
		serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
		sharedKeyCred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
		if err != nil {
			return fmt.Errorf("failed to create shared key credential: %w", err)
		}
		client, err = azblob.NewClientWithSharedKeyCredential(serviceURL, sharedKeyCred, nil)
		if err != nil {
			return fmt.Errorf("failed to create client with shared key credential: %w", err)
		}
	} else if clientID, ok := config["client_id"].(string); ok && clientID != "" {
		// Use service principal
		clientSecret, _ := config["client_secret"].(string)
		tenantID, _ := config["tenant_id"].(string)

		cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
		if err != nil {
			return fmt.Errorf("failed to create client secret credential: %w", err)
		}
		serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
		client, err = azblob.NewClient(serviceURL, cred, nil)
		if err != nil {
			return fmt.Errorf("failed to create Azure Blob client: %w", err)
		}
	} else {
		// Use default credential (managed identity, environment variables, etc.)
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return fmt.Errorf("failed to create default credential: %w", err)
		}
		serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
		client, err = azblob.NewClient(serviceURL, cred, nil)
		if err != nil {
			return fmt.Errorf("failed to create Azure Blob client: %w", err)
		}
	}

	p.client = client
	p.config = config

	p.logger.Info("Azure Blob Storage provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.client != nil
}

// PutObject uploads an object to Azure Blob Storage
func (p *Provider) PutObject(ctx context.Context, request *types.PutObjectRequest) (*types.PutObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("azure provider not configured")
	}

	// Prepare blob options
	options := &azblob.UploadStreamOptions{}

	// Add metadata
	if request.Metadata != nil {
		metadata := make(map[string]*string)
		for k, v := range request.Metadata {
			metadata[k] = &v
		}
		options.Metadata = metadata
	}

	// Add access tier
	if _, ok := request.Options["access_tier"].(string); ok {
		// AccessTier is not directly supported in UploadStreamOptions
		// We would need to use a different approach for setting access tier
	}

	// Upload blob
	result, err := p.client.UploadStream(ctx, request.Bucket, request.Key, bytes.NewReader(request.Content), options)
	if err != nil {
		return nil, fmt.Errorf("failed to upload blob: %w", err)
	}

	response := &types.PutObjectResponse{
		Key:          request.Key,
		ETag:         strings.Trim(string(*result.ETag), "\""),
		Size:         request.Size,
		LastModified: *result.LastModified,
		Metadata:     request.Metadata,
		ProviderData: map[string]interface{}{
			"version_id": result.VersionID,
		},
	}

	if result.VersionID != nil {
		response.VersionID = *result.VersionID
	}

	return response, nil
}

// GetObject downloads an object from Azure Blob Storage
func (p *Provider) GetObject(ctx context.Context, request *types.GetObjectRequest) (*types.GetObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("azure provider not configured")
	}

	// Prepare download options
	options := &azblob.DownloadStreamOptions{}

	// Add version ID if specified
	if request.VersionID != "" {
		// Version ID is not directly supported in DownloadStreamOptions
		// We'll need to use a different approach for versioned downloads
	}

	// Add range if specified
	if request.Range != nil {
		offset := request.Range.Start
		count := request.Range.End - request.Range.Start + 1
		options.Range = azblob.HTTPRange{
			Offset: offset,
			Count:  count,
		}
	}

	// Download blob
	result, err := p.client.DownloadStream(ctx, request.Bucket, request.Key, options)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob: %w", err)
	}

	// Convert metadata from map[string]*string to map[string]string
	metadata := make(map[string]string)
	if result.Metadata != nil {
		for k, v := range result.Metadata {
			if v != nil {
				metadata[k] = *v
			}
		}
	}

	// Read the content from the response body
	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	result.Body.Close()

	response := &types.GetObjectResponse{
		Content:      content,
		Size:         *result.ContentLength,
		ContentType:  *result.ContentType,
		ETag:         strings.Trim(string(*result.ETag), "\""),
		LastModified: *result.LastModified,
		Metadata:     metadata,
		ProviderData: map[string]interface{}{
			"version_id": result.VersionID,
		},
	}

	return response, nil
}

// DeleteObject deletes an object from Azure Blob Storage
func (p *Provider) DeleteObject(ctx context.Context, request *types.DeleteObjectRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("azure provider not configured")
	}

	// Prepare delete options
	options := &azblob.DeleteBlobOptions{}

	// Add version ID if specified
	if request.VersionID != "" {
		// Version ID is not directly supported in DeleteBlobOptions
		// We'll need to use a different approach for versioned deletes
	}

	// Delete blob
	_, err := p.client.DeleteBlob(ctx, request.Bucket, request.Key, options)
	if err != nil {
		return fmt.Errorf("failed to delete blob: %w", err)
	}

	return nil
}

// DeleteObjects deletes multiple objects from Azure Blob Storage
func (p *Provider) DeleteObjects(ctx context.Context, request *types.DeleteObjectsRequest) (*types.DeleteObjectsResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("azure provider not configured")
	}

	response := &types.DeleteObjectsResponse{
		Deleted: make([]types.DeletedObject, 0, len(request.Keys)),
		Errors:  make([]types.DeleteError, 0),
	}

	// Delete objects one by one
	for _, key := range request.Keys {
		deleteRequest := &types.DeleteObjectRequest{
			Bucket: request.Bucket,
			Key:    key,
		}

		err := p.DeleteObject(ctx, deleteRequest)
		if err != nil {
			response.Errors = append(response.Errors, types.DeleteError{
				Key:     key,
				Code:    "DeleteFailed",
				Message: err.Error(),
			})
		} else {
			response.Deleted = append(response.Deleted, types.DeletedObject{
				Key: key,
			})
		}
	}

	return response, nil
}

// ListObjects lists objects in Azure Blob Storage
func (p *Provider) ListObjects(ctx context.Context, request *types.ListObjectsRequest) (*types.ListObjectsResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("azure provider not configured")
	}

	// Prepare list options
	options := &azblob.ListBlobsFlatOptions{
		Prefix: &request.Prefix,
	}

	// Add max results
	if request.MaxKeys > 0 {
		maxResults := int32(request.MaxKeys)
		options.MaxResults = &maxResults
	}

	// Add marker for pagination
	if request.ContinuationToken != "" {
		options.Marker = &request.ContinuationToken
	}

	// List blobs
	pager := p.client.NewListBlobsFlatPager(request.Bucket, options)

	response := &types.ListObjectsResponse{
		Objects:        make([]types.ObjectInfo, 0),
		CommonPrefixes: make([]string, 0),
		ProviderData:   make(map[string]interface{}),
	}

	// Process blobs
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next page: %w", err)
		}

		// Process blobs in this page
		for _, blob := range page.Segment.BlobItems {
			// Convert metadata from map[string]*string to map[string]string
			metadata := make(map[string]string)
			if blob.Metadata != nil {
				for k, v := range blob.Metadata {
					if v != nil {
						metadata[k] = *v
					}
				}
			}

			objInfo := types.ObjectInfo{
				Key:          *blob.Name,
				Size:         *blob.Properties.ContentLength,
				LastModified: *blob.Properties.LastModified,
				ETag:         strings.Trim(string(*blob.Properties.ETag), "\""),
				ContentType:  *blob.Properties.ContentType,
				Metadata:     metadata,
				StorageClass: string(*blob.Properties.AccessTier),
				ProviderData: map[string]interface{}{
					"version_id": nil, // VersionID is not available in this context
				},
			}

			// VersionID is not available in blob properties in this context
			// We would need to make additional API calls to get version information

			response.Objects = append(response.Objects, objInfo)
		}

		// Check if there are more pages
		if page.NextMarker != nil {
			response.NextContinuationToken = *page.NextMarker
			response.IsTruncated = true
		}
	}

	return response, nil
}

// ObjectExists checks if an object exists in Azure Blob Storage
func (p *Provider) ObjectExists(ctx context.Context, request *types.ObjectExistsRequest) (bool, error) {
	if !p.IsConfigured() {
		return false, fmt.Errorf("azure provider not configured")
	}

	// Get blob properties
	_, err := p.client.ServiceClient().NewContainerClient(request.Bucket).NewBlobClient(request.Key).GetProperties(ctx, nil)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "BlobNotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check blob existence: %w", err)
	}

	return true, nil
}

// CopyObject copies an object within Azure Blob Storage
func (p *Provider) CopyObject(ctx context.Context, request *types.CopyObjectRequest) (*types.CopyObjectResponse, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("azure provider not configured")
	}

	// Prepare source URL
	sourceURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		p.config["account_name"], request.SourceBucket, request.SourceKey)

	// Start copy operation
	copyResult, err := p.client.ServiceClient().NewContainerClient(request.DestBucket).NewBlobClient(request.DestKey).StartCopyFromURL(ctx, sourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start copy operation: %w", err)
	}

	// Wait for copy to complete
	for {
		props, err := p.client.ServiceClient().NewContainerClient(request.DestBucket).NewBlobClient(request.DestKey).GetProperties(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get blob properties: %w", err)
		}

		if props.CopyStatus != nil && *props.CopyStatus == "success" {
			break
		}

		if props.CopyStatus != nil && *props.CopyStatus == "failed" {
			return nil, fmt.Errorf("copy operation failed")
		}

		// Wait a bit before checking again
		time.Sleep(100 * time.Millisecond)
	}

	response := &types.CopyObjectResponse{
		Key:          request.DestKey,
		ETag:         strings.Trim(string(*copyResult.ETag), "\""),
		LastModified: time.Now(),
		ProviderData: map[string]interface{}{
			"copy_id": copyResult.CopyID,
		},
	}

	return response, nil
}

// MoveObject moves an object within Azure Blob Storage
func (p *Provider) MoveObject(ctx context.Context, request *types.MoveObjectRequest) (*types.MoveObjectResponse, error) {
	// Move is implemented as copy + delete
	copyRequest := &types.CopyObjectRequest{
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
	deleteRequest := &types.DeleteObjectRequest{
		Bucket: request.SourceBucket,
		Key:    request.SourceKey,
	}

	err = p.DeleteObject(ctx, deleteRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to delete source object after move: %w", err)
	}

	response := &types.MoveObjectResponse{
		Key:          copyResponse.Key,
		ETag:         copyResponse.ETag,
		LastModified: copyResponse.LastModified,
		ProviderData: copyResponse.ProviderData,
	}

	return response, nil
}

// GetObjectInfo gets object information from Azure Blob Storage
func (p *Provider) GetObjectInfo(ctx context.Context, request *types.GetObjectInfoRequest) (*types.ObjectInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("azure provider not configured")
	}

	// Get blob properties
	props, err := p.client.ServiceClient().NewContainerClient(request.Bucket).NewBlobClient(request.Key).GetProperties(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob properties: %w", err)
	}

	// Convert metadata from map[string]*string to map[string]string
	metadata := make(map[string]string)
	if props.Metadata != nil {
		for k, v := range props.Metadata {
			if v != nil {
				metadata[k] = *v
			}
		}
	}

	info := &types.ObjectInfo{
		Key:          request.Key,
		Size:         *props.ContentLength,
		LastModified: *props.LastModified,
		ETag:         strings.Trim(string(*props.ETag), "\""),
		ContentType:  *props.ContentType,
		Metadata:     metadata,
		StorageClass: string(*props.AccessTier),
		ProviderData: map[string]interface{}{
			"version_id": nil, // VersionID is not available in this context
		},
	}

	// VersionID is not available in blob properties in this context
	// We would need to make additional API calls to get version information

	return info, nil
}

// GeneratePresignedURL generates a presigned URL for Azure Blob Storage
func (p *Provider) GeneratePresignedURL(ctx context.Context, request *types.PresignedURLRequest) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("azure provider not configured")
	}

	// Generate SAS URL
	var permissions sas.BlobPermissions
	switch strings.ToUpper(request.Method) {
	case "GET":
		permissions = sas.BlobPermissions{Read: true}
	case "PUT":
		permissions = sas.BlobPermissions{Write: true}
	case "DELETE":
		permissions = sas.BlobPermissions{Delete: true}
	default:
		return "", fmt.Errorf("unsupported method: %s", request.Method)
	}

	// Create SAS URL
	expiry := time.Now().Add(request.ExpiresIn)
	if !request.Expires.IsZero() {
		expiry = request.Expires
	}
	sasURL, err := p.client.ServiceClient().NewContainerClient(request.Bucket).NewBlobClient(request.Key).GetSASURL(permissions, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate SAS URL: %w", err)
	}

	return sasURL, nil
}

// GeneratePublicURL generates a public URL for Azure Blob Storage
func (p *Provider) GeneratePublicURL(ctx context.Context, request *types.PublicURLRequest) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("azure provider not configured")
	}

	// Generate public URL
	accountName := p.config["account_name"].(string)
	url := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", accountName, request.Bucket, request.Key)

	return url, nil
}

// CreateBucket creates a container in Azure Blob Storage
func (p *Provider) CreateBucket(ctx context.Context, request *types.CreateBucketRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("azure provider not configured")
	}

	// Create container
	_, err := p.client.CreateContainer(ctx, request.Bucket, nil)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	return nil
}

// DeleteBucket deletes a container from Azure Blob Storage
func (p *Provider) DeleteBucket(ctx context.Context, request *types.DeleteBucketRequest) error {
	if !p.IsConfigured() {
		return fmt.Errorf("azure provider not configured")
	}

	// Delete container
	_, err := p.client.DeleteContainer(ctx, request.Bucket, nil)
	if err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	return nil
}

// BucketExists checks if a container exists in Azure Blob Storage
func (p *Provider) BucketExists(ctx context.Context, request *types.BucketExistsRequest) (bool, error) {
	if !p.IsConfigured() {
		return false, fmt.Errorf("azure provider not configured")
	}

	// Get container properties
	_, err := p.client.ServiceClient().NewContainerClient(request.Bucket).GetProperties(ctx, nil)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "ContainerNotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check container existence: %w", err)
	}

	return true, nil
}

// ListBuckets lists containers in Azure Blob Storage
func (p *Provider) ListBuckets(ctx context.Context) ([]types.BucketInfo, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("azure provider not configured")
	}

	// List containers
	pager := p.client.NewListContainersPager(nil)

	buckets := make([]types.BucketInfo, 0)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next page: %w", err)
		}

		for _, container := range page.ContainerItems {
			buckets = append(buckets, types.BucketInfo{
				Name:         *container.Name,
				CreationDate: time.Now(), // CreationTime is not available in this context
				ProviderData: map[string]interface{}{
					"etag": container.Properties.ETag,
				},
			})
		}
	}

	return buckets, nil
}

// HealthCheck performs a health check on Azure Blob Storage
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("azure provider not configured")
	}

	// List containers as a health check
	_, err := p.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("azure health check failed: %w", err)
	}

	return nil
}

// Close closes the Azure Blob Storage provider
func (p *Provider) Close() error {
	// Azure client doesn't need explicit closing
	return nil
}
