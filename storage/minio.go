package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

// MinIO represents MinIO storage client
type MinIO struct {
	client *minio.Client
	config *MinIOConfig
	logger *logrus.Logger
}

// MinIOConfig holds MinIO configuration
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	Region          string
}

// ObjectInfo represents object information
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
	ContentType  string
	Metadata     map[string]string
}

// NewMinIO creates a new MinIO client
func NewMinIO(config *MinIOConfig, logger *logrus.Logger) (*MinIO, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	m := &MinIO{
		client: client,
		config: config,
		logger: logger,
	}

	// Ensure bucket exists
	if err := m.EnsureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	m.logger.Info("MinIO client initialized successfully")
	return m, nil
}

// EnsureBucket ensures the bucket exists
func (m *MinIO) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.config.BucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.config.BucketName, minio.MakeBucketOptions{
			Region: m.config.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		m.logger.Infof("Created bucket: %s", m.config.BucketName)
	}

	return nil
}

// PutObject uploads an object to MinIO
func (m *MinIO) PutObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string, metadata map[string]string) error {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	if metadata != nil {
		opts.UserMetadata = metadata
	}

	_, err := m.client.PutObject(ctx, m.config.BucketName, objectName, reader, objectSize, opts)
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}

	m.logger.Debugf("Uploaded object: %s", objectName)
	return nil
}

// GetObject downloads an object from MinIO
func (m *MinIO) GetObject(ctx context.Context, objectName string) (*minio.Object, error) {
	object, err := m.client.GetObject(ctx, m.config.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	m.logger.Debugf("Downloaded object: %s", objectName)
	return object, nil
}

// GetObjectInfo gets object information
func (m *MinIO) GetObjectInfo(ctx context.Context, objectName string) (*ObjectInfo, error) {
	info, err := m.client.StatObject(ctx, m.config.BucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	return &ObjectInfo{
		Key:          info.Key,
		Size:         info.Size,
		LastModified: info.LastModified,
		ETag:         info.ETag,
		ContentType:  info.ContentType,
		Metadata:     info.UserMetadata,
	}, nil
}

// DeleteObject deletes an object from MinIO
func (m *MinIO) DeleteObject(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(ctx, m.config.BucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	m.logger.Debugf("Deleted object: %s", objectName)
	return nil
}

// DeleteObjects deletes multiple objects from MinIO
func (m *MinIO) DeleteObjects(ctx context.Context, objectNames []string) error {
	objectsCh := make(chan minio.ObjectInfo, len(objectNames))

	go func() {
		defer close(objectsCh)
		for _, objectName := range objectNames {
			objectsCh <- minio.ObjectInfo{Key: objectName}
		}
	}()

	for err := range m.client.RemoveObjects(ctx, m.config.BucketName, objectsCh, minio.RemoveObjectsOptions{}) {
		if err.Err != nil {
			return fmt.Errorf("failed to delete object %s: %w", err.ObjectName, err.Err)
		}
	}

	m.logger.Debugf("Deleted %d objects", len(objectNames))
	return nil
}

// ListObjects lists objects in the bucket
func (m *MinIO) ListObjects(ctx context.Context, prefix string, recursive bool) (<-chan minio.ObjectInfo, error) {
	objectCh := m.client.ListObjects(ctx, m.config.BucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})

	return objectCh, nil
}

// CopyObject copies an object within MinIO
func (m *MinIO) CopyObject(ctx context.Context, sourceObjectName, destObjectName string, metadata map[string]string) error {
	source := minio.CopySrcOptions{
		Bucket: m.config.BucketName,
		Object: sourceObjectName,
	}

	dest := minio.CopyDestOptions{
		Bucket: m.config.BucketName,
		Object: destObjectName,
	}

	if metadata != nil {
		dest.UserMetadata = metadata
	}

	_, err := m.client.CopyObject(ctx, dest, source)
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	m.logger.Debugf("Copied object from %s to %s", sourceObjectName, destObjectName)
	return nil
}

// PresignedGetObject generates a presigned URL for getting an object
func (m *MinIO) PresignedGetObject(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.config.BucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// PresignedPutObject generates a presigned URL for putting an object
func (m *MinIO) PresignedPutObject(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedPutObject(ctx, m.config.BucketName, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// UploadFile uploads a file from local filesystem
func (m *MinIO) UploadFile(ctx context.Context, filePath, objectName string, contentType string, metadata map[string]string) error {
	// Generate object name if not provided
	if objectName == "" {
		objectName = filepath.Base(filePath)
	}

	// Generate content type if not provided
	if contentType == "" {
		contentType = getContentType(filePath)
	}

	_, err := m.client.FPutObject(ctx, m.config.BucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	m.logger.Debugf("Uploaded file: %s -> %s", filePath, objectName)
	return nil
}

// DownloadFile downloads a file to local filesystem
func (m *MinIO) DownloadFile(ctx context.Context, objectName, filePath string) error {
	err := m.client.FGetObject(ctx, m.config.BucketName, objectName, filePath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	m.logger.Debugf("Downloaded file: %s -> %s", objectName, filePath)
	return nil
}

// ObjectExists checks if an object exists
func (m *MinIO) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.config.BucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}

// GetBucketPolicy gets the bucket policy
func (m *MinIO) GetBucketPolicy(ctx context.Context) (string, error) {
	policy, err := m.client.GetBucketPolicy(ctx, m.config.BucketName)
	if err != nil {
		return "", fmt.Errorf("failed to get bucket policy: %w", err)
	}

	return policy, nil
}

// SetBucketPolicy sets the bucket policy
func (m *MinIO) SetBucketPolicy(ctx context.Context, policy string) error {
	err := m.client.SetBucketPolicy(ctx, m.config.BucketName, policy)
	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	m.logger.Debugf("Set bucket policy for: %s", m.config.BucketName)
	return nil
}

// HealthCheck performs a health check on MinIO
func (m *MinIO) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Check if bucket exists
	exists, err := m.client.BucketExists(ctx, m.config.BucketName)
	if err != nil {
		return fmt.Errorf("minio health check failed: %w", err)
	}

	if !exists {
		return fmt.Errorf("bucket %s does not exist", m.config.BucketName)
	}

	// Test basic operations
	testObjectName := "health_check_test"
	testContent := "health check test content"

	// Put test object
	_, err = m.client.PutObject(ctx, m.config.BucketName, testObjectName, strings.NewReader(testContent), int64(len(testContent)), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio put object test failed: %w", err)
	}

	// Get test object
	object, err := m.client.GetObject(ctx, m.config.BucketName, testObjectName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio get object test failed: %w", err)
	}
	object.Close()

	// Delete test object
	err = m.client.RemoveObject(ctx, m.config.BucketName, testObjectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio delete object test failed: %w", err)
	}

	return nil
}

// Close closes the MinIO client
func (m *MinIO) Close() error {
	// MinIO client doesn't need explicit closing
	return nil
}

// getContentType determines content type based on file extension
func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	contentTypes := map[string]string{
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
		".7z":   "application/x-7z-compressed",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}

	return "application/octet-stream"
}
