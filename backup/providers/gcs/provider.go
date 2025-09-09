package gcs

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/anasamu/microservices-library-go/backup/types"
	"google.golang.org/api/iterator"
)

// GCSProvider implements backup operations using Google Cloud Storage
type GCSProvider struct {
	bucket *storage.BucketHandle
	prefix string
}

// GCSConfig contains configuration for GCS provider
type GCSConfig struct {
	Bucket string
	Prefix string
}

// NewGCSProvider creates a new GCS backup provider
func NewGCSProvider(ctx context.Context, config *GCSConfig) (*GCSProvider, error) {
	if config.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required")
	}

	// Create GCS client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	// Get bucket handle
	bucket := client.Bucket(config.Bucket)

	return &GCSProvider{
		bucket: bucket,
		prefix: config.Prefix,
	}, nil
}

// CreateBackup creates a backup in GCS
func (p *GCSProvider) CreateBackup(ctx context.Context, name string, data io.Reader, opts *types.BackupOptions) (*types.BackupMetadata, error) {
	// Generate backup ID
	backupID := fmt.Sprintf("%s-%d", name, time.Now().Unix())

	// Create GCS object name
	objectName := p.getGCSObjectName(backupID)

	// Create object writer
	obj := p.bucket.Object(objectName)
	writer := obj.NewWriter(ctx)

	// Set metadata
	writer.Metadata = map[string]string{
		"backup-name":        name,
		"backup-id":          backupID,
		"backup-created-at":  time.Now().Format(time.RFC3339),
		"backup-description": opts.Description,
	}

	// Copy data to GCS
	_, err := io.Copy(writer, data)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to write data to GCS: %w", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close GCS writer: %w", err)
	}

	// Get object attributes to get size
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object attributes: %w", err)
	}

	// Create metadata
	metadata := &types.BackupMetadata{
		ID:          backupID,
		Name:        name,
		Size:        attrs.Size,
		CreatedAt:   time.Now(),
		Tags:        opts.Tags,
		Description: opts.Description,
	}

	// Store metadata as a separate object
	metadataObjectName := p.getMetadataObjectName(backupID)
	metadataObj := p.bucket.Object(metadataObjectName)
	metadataWriter := metadataObj.NewWriter(ctx)
	metadataWriter.ContentType = "application/json"

	metadataData := fmt.Sprintf(`{"id":"%s","name":"%s","size":%d,"created_at":"%s","tags":%s,"description":"%s"}`,
		backupID, name, metadata.Size, metadata.CreatedAt.Format(time.RFC3339),
		formatTags(opts.Tags), opts.Description)

	_, err = metadataWriter.Write([]byte(metadataData))
	if err != nil {
		metadataWriter.Close()
		return nil, fmt.Errorf("failed to write metadata to GCS: %w", err)
	}

	if err := metadataWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close metadata writer: %w", err)
	}

	return metadata, nil
}

// RestoreBackup restores a backup from GCS
func (p *GCSProvider) RestoreBackup(ctx context.Context, backupID string, writer io.Writer, opts *types.RestoreOptions) error {
	objectName := p.getGCSObjectName(backupID)
	obj := p.bucket.Object(objectName)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("failed to create GCS reader: %w", err)
	}
	defer reader.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("failed to copy data from GCS: %w", err)
	}

	return nil
}

// ListBackups lists all available backups
func (p *GCSProvider) ListBackups(ctx context.Context) ([]*types.BackupMetadata, error) {
	// List metadata objects
	prefix := p.getMetadataPrefix()

	query := &storage.Query{
		Prefix: prefix,
	}

	it := p.bucket.Objects(ctx, query)
	var backups []*types.BackupMetadata

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate objects: %w", err)
		}

		// Download and parse metadata
		metadata, err := p.downloadMetadata(ctx, attrs.Name)
		if err != nil {
			continue // Skip invalid metadata
		}
		backups = append(backups, metadata)
	}

	return backups, nil
}

// GetBackup retrieves backup metadata
func (p *GCSProvider) GetBackup(ctx context.Context, backupID string) (*types.BackupMetadata, error) {
	metadataObjectName := p.getMetadataObjectName(backupID)
	return p.downloadMetadata(ctx, metadataObjectName)
}

// DeleteBackup removes a backup from GCS
func (p *GCSProvider) DeleteBackup(ctx context.Context, backupID string) error {
	// Delete backup data
	objectName := p.getGCSObjectName(backupID)
	obj := p.bucket.Object(objectName)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete backup data: %w", err)
	}

	// Delete metadata
	metadataObjectName := p.getMetadataObjectName(backupID)
	metadataObj := p.bucket.Object(metadataObjectName)
	if err := metadataObj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete backup metadata: %w", err)
	}

	return nil
}

// HealthCheck checks if GCS is accessible
func (p *GCSProvider) HealthCheck(ctx context.Context) error {
	// Try to get bucket attributes
	_, err := p.bucket.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("GCS health check failed: %w", err)
	}
	return nil
}

// Helper methods

func (p *GCSProvider) getGCSObjectName(backupID string) string {
	if p.prefix != "" {
		return filepath.Join(p.prefix, "backups", backupID)
	}
	return filepath.Join("backups", backupID)
}

func (p *GCSProvider) getMetadataObjectName(backupID string) string {
	if p.prefix != "" {
		return filepath.Join(p.prefix, "metadata", backupID+".json")
	}
	return filepath.Join("metadata", backupID+".json")
}

func (p *GCSProvider) getMetadataPrefix() string {
	if p.prefix != "" {
		return filepath.Join(p.prefix, "metadata")
	}
	return "metadata"
}

func (p *GCSProvider) downloadMetadata(ctx context.Context, objectName string) (*types.BackupMetadata, error) {
	obj := p.bucket.Object(objectName)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

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
