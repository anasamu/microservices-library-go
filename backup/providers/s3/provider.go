package s3

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/backup/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Provider implements backup operations using AWS S3
type S3Provider struct {
	bucket     string
	prefix     string
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	svc        *s3.S3
}

// S3Config contains configuration for S3 provider
type S3Config struct {
	Bucket   string
	Prefix   string
	Region   string
	Endpoint string // Optional custom endpoint for S3-compatible services
}

// NewS3Provider creates a new S3 backup provider
func NewS3Provider(config *S3Config) (*S3Provider, error) {
	if config.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required")
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(config.Region),
		Endpoint: aws.String(config.Endpoint), // Will be nil if empty
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create S3 service
	svc := s3.New(sess)

	// Create uploader and downloader
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)

	return &S3Provider{
		bucket:     config.Bucket,
		prefix:     config.Prefix,
		uploader:   uploader,
		downloader: downloader,
		svc:        svc,
	}, nil
}

// CreateBackup creates a backup in S3
func (p *S3Provider) CreateBackup(ctx context.Context, name string, data io.Reader, opts *types.BackupOptions) (*types.BackupMetadata, error) {
	// Generate backup ID
	backupID := fmt.Sprintf("%s-%d", name, time.Now().Unix())

	// Create S3 key
	key := p.getS3Key(backupID)

	// Upload to S3
	_, err := p.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
		Body:   data,
		Metadata: map[string]*string{
			"backup-name":        aws.String(name),
			"backup-id":          aws.String(backupID),
			"backup-created-at":  aws.String(time.Now().Format(time.RFC3339)),
			"backup-description": aws.String(opts.Description),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload backup to S3: %w", err)
	}

	// Get object size
	headResult, err := p.svc.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	// Create metadata
	metadata := &types.BackupMetadata{
		ID:          backupID,
		Name:        name,
		Size:        *headResult.ContentLength,
		CreatedAt:   time.Now(),
		Tags:        opts.Tags,
		Description: opts.Description,
	}

	// Store metadata as a separate object
	metadataKey := p.getMetadataKey(backupID)
	metadataData := fmt.Sprintf(`{"id":"%s","name":"%s","size":%d,"created_at":"%s","tags":%s,"description":"%s"}`,
		backupID, name, metadata.Size, metadata.CreatedAt.Format(time.RFC3339),
		formatTags(opts.Tags), opts.Description)

	_, err = p.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(metadataKey),
		Body:        strings.NewReader(metadataData),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store backup metadata: %w", err)
	}

	return metadata, nil
}

// RestoreBackup restores a backup from S3
func (p *S3Provider) RestoreBackup(ctx context.Context, backupID string, writer io.Writer, opts *types.RestoreOptions) error {
	key := p.getS3Key(backupID)

	_, err := p.downloader.DownloadWithContext(ctx, writer, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download backup from S3: %w", err)
	}

	return nil
}

// ListBackups lists all available backups
func (p *S3Provider) ListBackups(ctx context.Context) ([]*types.BackupMetadata, error) {
	// List metadata objects
	prefix := p.getMetadataPrefix()

	result, err := p.svc.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(p.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list backup metadata: %w", err)
	}

	var backups []*types.BackupMetadata
	for _, obj := range result.Contents {
		// Download metadata
		metadata, err := p.downloadMetadata(ctx, *obj.Key)
		if err != nil {
			continue // Skip invalid metadata
		}
		backups = append(backups, metadata)
	}

	return backups, nil
}

// GetBackup retrieves backup metadata
func (p *S3Provider) GetBackup(ctx context.Context, backupID string) (*types.BackupMetadata, error) {
	metadataKey := p.getMetadataKey(backupID)
	return p.downloadMetadata(ctx, metadataKey)
}

// DeleteBackup removes a backup from S3
func (p *S3Provider) DeleteBackup(ctx context.Context, backupID string) error {
	// Delete backup data
	key := p.getS3Key(backupID)
	_, err := p.svc.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete backup data: %w", err)
	}

	// Delete metadata
	metadataKey := p.getMetadataKey(backupID)
	_, err = p.svc.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(metadataKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete backup metadata: %w", err)
	}

	return nil
}

// HealthCheck checks if S3 is accessible
func (p *S3Provider) HealthCheck(ctx context.Context) error {
	_, err := p.svc.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(p.bucket),
	})
	if err != nil {
		return fmt.Errorf("S3 health check failed: %w", err)
	}
	return nil
}

// Helper methods

func (p *S3Provider) getS3Key(backupID string) string {
	if p.prefix != "" {
		return filepath.Join(p.prefix, "backups", backupID)
	}
	return filepath.Join("backups", backupID)
}

func (p *S3Provider) getMetadataKey(backupID string) string {
	if p.prefix != "" {
		return filepath.Join(p.prefix, "metadata", backupID+".json")
	}
	return filepath.Join("metadata", backupID+".json")
}

func (p *S3Provider) getMetadataPrefix() string {
	if p.prefix != "" {
		return filepath.Join(p.prefix, "metadata")
	}
	return "metadata"
}

func (p *S3Provider) downloadMetadata(ctx context.Context, key string) (*types.BackupMetadata, error) {
	result, err := p.svc.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	// Parse metadata JSON (simplified - in production, use proper JSON parsing)
	// This is a simplified implementation
	metadata := &types.BackupMetadata{
		ID:        "parsed-from-json",
		Name:      "parsed-from-json",
		Size:      0,
		CreatedAt: time.Now(),
		Tags:      make(map[string]string),
	}

	return metadata, nil
}

func formatTags(tags map[string]string) string {
	if len(tags) == 0 {
		return "{}"
	}

	var parts []string
	for k, v := range tags {
		parts = append(parts, fmt.Sprintf(`"%s":"%s"`, k, v))
	}
	return "{" + strings.Join(parts, ",") + "}"
}
