package gateway

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/libs/storage/providers/azure"
	"github.com/anasamu/microservices-library-go/libs/storage/providers/gcs"
	"github.com/anasamu/microservices-library-go/libs/storage/providers/minio"
	"github.com/anasamu/microservices-library-go/libs/storage/providers/s3"
	"github.com/sirupsen/logrus"
)

// ExampleUsage demonstrates how to use the storage gateway system
func ExampleUsage() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create storage manager
	config := DefaultManagerConfig()
	config.DefaultProvider = "s3"
	config.MaxFileSize = 100 * 1024 * 1024 // 100MB
	config.AllowedTypes = []string{"image/*", "application/pdf", "text/*"}

	storageManager := NewStorageManager(config, logger)

	// Register storage providers
	registerProviders(storageManager, logger)

	// Example 1: Upload file to S3
	exampleS3Upload(storageManager)

	// Example 2: Upload file to Google Cloud Storage
	exampleGCSUpload(storageManager)

	// Example 3: Upload file to Azure Blob Storage
	exampleAzureUpload(storageManager)

	// Example 4: Upload file to MinIO
	exampleMinIOUpload(storageManager)

	// Example 5: List objects
	exampleListObjects(storageManager)

	// Example 6: Generate presigned URLs
	examplePresignedURLs(storageManager)

	// Example 7: Copy and move objects
	exampleCopyMoveObjects(storageManager)

	// Example 8: Health checks
	exampleHealthChecks(storageManager)
}

// registerProviders registers all storage providers
func registerProviders(storageManager *StorageManager, logger *logrus.Logger) {
	// Register S3 provider
	s3Provider := s3.NewProvider(logger)
	s3Config := map[string]interface{}{
		"region":            "us-east-1",
		"access_key_id":     "your_aws_access_key_id",
		"secret_access_key": "your_aws_secret_access_key",
		"endpoint":          "", // Use AWS S3
		"use_ssl":           true,
	}
	if err := s3Provider.Configure(s3Config); err != nil {
		log.Printf("Failed to configure S3: %v", err)
	} else {
		storageManager.RegisterProvider(s3Provider)
	}

	// Register Google Cloud Storage provider
	gcsProvider := gcs.NewProvider(logger)
	gcsConfig := map[string]interface{}{
		"project_id":       "your-gcp-project-id",
		"credentials_path": "/path/to/service-account.json",
		// Alternative: "credentials_json": "{\"type\":\"service_account\",...}"
	}
	if err := gcsProvider.Configure(gcsConfig); err != nil {
		log.Printf("Failed to configure GCS: %v", err)
	} else {
		storageManager.RegisterProvider(gcsProvider)
	}

	// Register Azure Blob Storage provider
	azureProvider := azure.NewProvider(logger)
	azureConfig := map[string]interface{}{
		"account_name": "your_storage_account",
		"account_key":  "your_storage_account_key",
		// Alternative: use service principal
		// "client_id":     "your_client_id",
		// "client_secret": "your_client_secret",
		// "tenant_id":     "your_tenant_id",
	}
	if err := azureProvider.Configure(azureConfig); err != nil {
		log.Printf("Failed to configure Azure: %v", err)
	} else {
		storageManager.RegisterProvider(azureProvider)
	}

	// Register MinIO provider
	minioProvider := minio.NewProvider(logger)
	minioConfig := map[string]interface{}{
		"endpoint":          "localhost:9000",
		"access_key_id":     "minioadmin",
		"secret_access_key": "minioadmin",
		"use_ssl":           false,
		"region":            "us-east-1",
	}
	if err := minioProvider.Configure(minioConfig); err != nil {
		log.Printf("Failed to configure MinIO: %v", err)
	} else {
		storageManager.RegisterProvider(minioProvider)
	}
}

// exampleS3Upload demonstrates S3 file upload
func exampleS3Upload(storageManager *StorageManager) {
	fmt.Println("=== S3 Upload Example ===")

	ctx := context.Background()

	// Create upload request
	content := strings.NewReader("Hello, S3! This is a test file.")
	request := &PutObjectRequest{
		Bucket:      "my-test-bucket",
		Key:         "test-files/s3-test.txt",
		Content:     content,
		Size:        int64(content.Len()),
		ContentType: "text/plain",
		Metadata: map[string]string{
			"uploaded-by": "storage-manager",
			"environment": "development",
		},
		Tags: map[string]string{
			"type": "test",
			"env":  "dev",
		},
	}

	// Upload to S3
	response, err := storageManager.PutObject(ctx, "s3", request)
	if err != nil {
		log.Printf("Failed to upload to S3: %v", err)
		return
	}

	fmt.Printf("S3 Upload Successful:\n")
	fmt.Printf("  Key: %s\n", response.Key)
	fmt.Printf("  ETag: %s\n", response.ETag)
	fmt.Printf("  Size: %d bytes\n", response.Size)
	fmt.Printf("  Last Modified: %s\n", response.LastModified)

	// Get object info
	infoRequest := &GetObjectInfoRequest{
		Bucket: "my-test-bucket",
		Key:    "test-files/s3-test.txt",
	}

	info, err := storageManager.GetObjectInfo(ctx, "s3", infoRequest)
	if err != nil {
		log.Printf("Failed to get object info: %v", err)
		return
	}

	fmt.Printf("Object Info:\n")
	fmt.Printf("  Content Type: %s\n", info.ContentType)
	fmt.Printf("  Metadata: %v\n", info.Metadata)
}

// exampleGCSUpload demonstrates Google Cloud Storage file upload
func exampleGCSUpload(storageManager *StorageManager) {
	fmt.Println("\n=== Google Cloud Storage Upload Example ===")

	ctx := context.Background()

	// Create upload request
	content := strings.NewReader("Hello, Google Cloud Storage! This is a test file.")
	request := &PutObjectRequest{
		Bucket:      "my-gcs-bucket",
		Key:         "test-files/gcs-test.txt",
		Content:     content,
		Size:        int64(content.Len()),
		ContentType: "text/plain",
		Metadata: map[string]string{
			"uploaded-by": "storage-manager",
			"environment": "development",
		},
	}

	// Upload to GCS
	response, err := storageManager.PutObject(ctx, "gcs", request)
	if err != nil {
		log.Printf("Failed to upload to GCS: %v", err)
		return
	}

	fmt.Printf("GCS Upload Successful:\n")
	fmt.Printf("  Key: %s\n", response.Key)
	fmt.Printf("  ETag: %s\n", response.ETag)
	fmt.Printf("  Size: %d bytes\n", response.Size)
}

// exampleAzureUpload demonstrates Azure Blob Storage file upload
func exampleAzureUpload(storageManager *StorageManager) {
	fmt.Println("\n=== Azure Blob Storage Upload Example ===")

	ctx := context.Background()

	// Create upload request
	content := strings.NewReader("Hello, Azure Blob Storage! This is a test file.")
	request := &PutObjectRequest{
		Bucket:      "my-azure-container",
		Key:         "test-files/azure-test.txt",
		Content:     content,
		Size:        int64(content.Len()),
		ContentType: "text/plain",
		Metadata: map[string]string{
			"uploaded-by": "storage-manager",
			"environment": "development",
		},
		Options: map[string]interface{}{
			"access_tier": "Hot",
		},
	}

	// Upload to Azure
	response, err := storageManager.PutObject(ctx, "azure", request)
	if err != nil {
		log.Printf("Failed to upload to Azure: %v", err)
		return
	}

	fmt.Printf("Azure Upload Successful:\n")
	fmt.Printf("  Key: %s\n", response.Key)
	fmt.Printf("  ETag: %s\n", response.ETag)
	fmt.Printf("  Size: %d bytes\n", response.Size)
}

// exampleMinIOUpload demonstrates MinIO file upload
func exampleMinIOUpload(storageManager *StorageManager) {
	fmt.Println("\n=== MinIO Upload Example ===")

	ctx := context.Background()

	// Create upload request
	content := strings.NewReader("Hello, MinIO! This is a test file.")
	request := &PutObjectRequest{
		Bucket:      "my-minio-bucket",
		Key:         "test-files/minio-test.txt",
		Content:     content,
		Size:        int64(content.Len()),
		ContentType: "text/plain",
		Metadata: map[string]string{
			"uploaded-by": "storage-manager",
			"environment": "development",
		},
	}

	// Upload to MinIO
	response, err := storageManager.PutObject(ctx, "minio", request)
	if err != nil {
		log.Printf("Failed to upload to MinIO: %v", err)
		return
	}

	fmt.Printf("MinIO Upload Successful:\n")
	fmt.Printf("  Key: %s\n", response.Key)
	fmt.Printf("  ETag: %s\n", response.ETag)
	fmt.Printf("  Size: %d bytes\n", response.Size)
}

// exampleListObjects demonstrates listing objects
func exampleListObjects(storageManager *StorageManager) {
	fmt.Println("\n=== List Objects Example ===")

	ctx := context.Background()

	// List objects from S3
	listRequest := &ListObjectsRequest{
		Bucket:  "my-test-bucket",
		Prefix:  "test-files/",
		MaxKeys: 10,
	}

	response, err := storageManager.ListObjects(ctx, "s3", listRequest)
	if err != nil {
		log.Printf("Failed to list objects: %v", err)
		return
	}

	fmt.Printf("Found %d objects:\n", len(response.Objects))
	for _, obj := range response.Objects {
		fmt.Printf("  - %s (%d bytes, %s)\n", obj.Key, obj.Size, obj.LastModified.Format(time.RFC3339))
	}

	if len(response.CommonPrefixes) > 0 {
		fmt.Printf("Common prefixes:\n")
		for _, prefix := range response.CommonPrefixes {
			fmt.Printf("  - %s/\n", prefix)
		}
	}
}

// examplePresignedURLs demonstrates generating presigned URLs
func examplePresignedURLs(storageManager *StorageManager) {
	fmt.Println("\n=== Presigned URLs Example ===")

	ctx := context.Background()

	// Generate presigned GET URL
	getURLRequest := &PresignedURLRequest{
		Bucket:  "my-test-bucket",
		Key:     "test-files/s3-test.txt",
		Method:  "GET",
		Expires: 1 * time.Hour,
	}

	getURL, err := storageManager.GeneratePresignedURL(ctx, "s3", getURLRequest)
	if err != nil {
		log.Printf("Failed to generate presigned GET URL: %v", err)
		return
	}

	fmt.Printf("Presigned GET URL: %s\n", getURL)

	// Generate presigned PUT URL
	putURLRequest := &PresignedURLRequest{
		Bucket:  "my-test-bucket",
		Key:     "test-files/upload-via-url.txt",
		Method:  "PUT",
		Expires: 1 * time.Hour,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}

	putURL, err := storageManager.GeneratePresignedURL(ctx, "s3", putURLRequest)
	if err != nil {
		log.Printf("Failed to generate presigned PUT URL: %v", err)
		return
	}

	fmt.Printf("Presigned PUT URL: %s\n", putURL)

	// Generate public URL
	publicURLRequest := &PublicURLRequest{
		Bucket: "my-test-bucket",
		Key:    "test-files/s3-test.txt",
	}

	publicURL, err := storageManager.GeneratePublicURL(ctx, "s3", publicURLRequest)
	if err != nil {
		log.Printf("Failed to generate public URL: %v", err)
		return
	}

	fmt.Printf("Public URL: %s\n", publicURL)
}

// exampleCopyMoveObjects demonstrates copying and moving objects
func exampleCopyMoveObjects(storageManager *StorageManager) {
	fmt.Println("\n=== Copy and Move Objects Example ===")

	ctx := context.Background()

	// Copy object
	copyRequest := &CopyObjectRequest{
		SourceBucket: "my-test-bucket",
		SourceKey:    "test-files/s3-test.txt",
		DestBucket:   "my-test-bucket",
		DestKey:      "test-files/s3-test-copy.txt",
		Metadata: map[string]string{
			"copied-from": "s3-test.txt",
			"copied-at":   time.Now().Format(time.RFC3339),
		},
	}

	copyResponse, err := storageManager.CopyObject(ctx, "s3", copyRequest)
	if err != nil {
		log.Printf("Failed to copy object: %v", err)
		return
	}

	fmt.Printf("Object copied successfully:\n")
	fmt.Printf("  Source: %s\n", copyRequest.SourceKey)
	fmt.Printf("  Destination: %s\n", copyResponse.Key)
	fmt.Printf("  ETag: %s\n", copyResponse.ETag)

	// Move object
	moveRequest := &MoveObjectRequest{
		SourceBucket: "my-test-bucket",
		SourceKey:    "test-files/s3-test-copy.txt",
		DestBucket:   "my-test-bucket",
		DestKey:      "test-files/s3-test-moved.txt",
	}

	moveResponse, err := storageManager.MoveObject(ctx, "s3", moveRequest)
	if err != nil {
		log.Printf("Failed to move object: %v", err)
		return
	}

	fmt.Printf("Object moved successfully:\n")
	fmt.Printf("  Source: %s\n", moveRequest.SourceKey)
	fmt.Printf("  Destination: %s\n", moveResponse.Key)
	fmt.Printf("  ETag: %s\n", moveResponse.ETag)
}

// exampleHealthChecks demonstrates health checks
func exampleHealthChecks(storageManager *StorageManager) {
	fmt.Println("\n=== Health Checks Example ===")

	ctx := context.Background()

	// Perform health checks on all providers
	results := storageManager.HealthCheck(ctx)

	fmt.Printf("Health Check Results:\n")
	for provider, err := range results {
		if err != nil {
			fmt.Printf("  %s: ❌ %v\n", provider, err)
		} else {
			fmt.Printf("  %s: ✅ Healthy\n", provider)
		}
	}

	// Get provider capabilities
	providers := storageManager.GetSupportedProviders()
	fmt.Printf("\nProvider Capabilities:\n")
	for _, providerName := range providers {
		features, maxSize, allowedTypes, err := storageManager.GetProviderCapabilities(providerName)
		if err != nil {
			log.Printf("Failed to get capabilities for %s: %v", providerName, err)
			continue
		}

		fmt.Printf("  %s:\n", providerName)
		fmt.Printf("    Features: %v\n", features)
		fmt.Printf("    Max File Size: %d bytes\n", maxSize)
		fmt.Printf("    Allowed Types: %v\n", allowedTypes)
	}
}

// ExampleBucketOperations demonstrates bucket operations
func ExampleBucketOperations() {
	fmt.Println("\n=== Bucket Operations Example ===")

	// Create logger and storage manager
	logger := logrus.New()
	config := DefaultManagerConfig()
	storageManager := NewStorageManager(config, logger)

	// Register providers
	registerProviders(storageManager, logger)

	ctx := context.Background()

	// Create bucket
	createRequest := &CreateBucketRequest{
		Bucket: "my-new-bucket",
		Region: "us-east-1",
		ACL:    "private",
	}

	err := storageManager.CreateBucket(ctx, "s3", createRequest)
	if err != nil {
		log.Printf("Failed to create bucket: %v", err)
	} else {
		fmt.Printf("Bucket created successfully: %s\n", createRequest.Bucket)
	}

	// Check if bucket exists
	existsRequest := &BucketExistsRequest{
		Bucket: "my-new-bucket",
	}

	exists, err := storageManager.BucketExists(ctx, "s3", existsRequest)
	if err != nil {
		log.Printf("Failed to check bucket existence: %v", err)
	} else {
		fmt.Printf("Bucket exists: %t\n", exists)
	}

	// List buckets
	buckets, err := storageManager.ListBuckets(ctx, "s3")
	if err != nil {
		log.Printf("Failed to list buckets: %v", err)
	} else {
		fmt.Printf("Found %d buckets:\n", len(buckets))
		for _, bucket := range buckets {
			fmt.Printf("  - %s (created: %s)\n", bucket.Name, bucket.CreationDate.Format(time.RFC3339))
		}
	}

	// Delete bucket
	deleteRequest := &DeleteBucketRequest{
		Bucket: "my-new-bucket",
	}

	err = storageManager.DeleteBucket(ctx, "s3", deleteRequest)
	if err != nil {
		log.Printf("Failed to delete bucket: %v", err)
	} else {
		fmt.Printf("Bucket deleted successfully: %s\n", deleteRequest.Bucket)
	}
}

// ExampleFileOperations demonstrates file operations
func ExampleFileOperations() {
	fmt.Println("\n=== File Operations Example ===")

	// Create logger and storage manager
	logger := logrus.New()
	config := DefaultManagerConfig()
	storageManager := NewStorageManager(config, logger)

	// Register providers
	registerProviders(storageManager, logger)

	ctx := context.Background()

	// Upload a file
	content := strings.NewReader("This is a test file for file operations.")
	uploadRequest := &PutObjectRequest{
		Bucket:      "my-test-bucket",
		Key:         "operations/test-file.txt",
		Content:     content,
		Size:        int64(content.Len()),
		ContentType: "text/plain",
		Metadata: map[string]string{
			"description": "Test file for operations",
		},
	}

	uploadResponse, err := storageManager.PutObject(ctx, "s3", uploadRequest)
	if err != nil {
		log.Printf("Failed to upload file: %v", err)
		return
	}

	fmt.Printf("File uploaded: %s\n", uploadResponse.Key)

	// Check if file exists
	existsRequest := &ObjectExistsRequest{
		Bucket: "my-test-bucket",
		Key:    "operations/test-file.txt",
	}

	exists, err := storageManager.ObjectExists(ctx, "s3", existsRequest)
	if err != nil {
		log.Printf("Failed to check file existence: %v", err)
	} else {
		fmt.Printf("File exists: %t\n", exists)
	}

	// Download file
	downloadRequest := &GetObjectRequest{
		Bucket: "my-test-bucket",
		Key:    "operations/test-file.txt",
	}

	downloadResponse, err := storageManager.GetObject(ctx, "s3", downloadRequest)
	if err != nil {
		log.Printf("Failed to download file: %v", err)
	} else {
		fmt.Printf("File downloaded: %d bytes, content type: %s\n", downloadResponse.Size, downloadResponse.ContentType)
		downloadResponse.Content.Close()
	}

	// Delete file
	deleteRequest := &DeleteObjectRequest{
		Bucket: "my-test-bucket",
		Key:    "operations/test-file.txt",
	}

	err = storageManager.DeleteObject(ctx, "s3", deleteRequest)
	if err != nil {
		log.Printf("Failed to delete file: %v", err)
	} else {
		fmt.Printf("File deleted successfully\n")
	}
}
