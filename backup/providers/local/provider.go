package local

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/backup/types"
)

// LocalProvider implements backup operations using local file system
type LocalProvider struct {
	basePath string
}

// NewLocalProvider creates a new local file backup provider
func NewLocalProvider(basePath string) *LocalProvider {
	return &LocalProvider{
		basePath: basePath,
	}
}

// CreateBackup creates a backup in local file system
func (p *LocalProvider) CreateBackup(ctx context.Context, name string, data io.Reader, opts *types.BackupOptions) (*types.BackupMetadata, error) {
	// Generate backup ID
	backupID := fmt.Sprintf("%s-%d", name, time.Now().Unix())

	// Create backup directory structure
	backupDir := filepath.Join(p.basePath, "backups")
	metadataDir := filepath.Join(p.basePath, "metadata")

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Create backup file
	backupPath := filepath.Join(backupDir, backupID)
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer backupFile.Close()

	// Copy data to backup file
	size, err := io.Copy(backupFile, data)
	if err != nil {
		os.Remove(backupPath) // Clean up on error
		return nil, fmt.Errorf("failed to write backup data: %w", err)
	}

	// Create metadata
	metadata := &types.BackupMetadata{
		ID:          backupID,
		Name:        name,
		Size:        size,
		CreatedAt:   time.Now(),
		Tags:        opts.Tags,
		Description: opts.Description,
	}

	// Store metadata
	metadataPath := filepath.Join(metadataDir, backupID+".json")
	if err := p.saveMetadata(metadataPath, metadata); err != nil {
		os.Remove(backupPath) // Clean up on error
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	return metadata, nil
}

// RestoreBackup restores a backup from local file system
func (p *LocalProvider) RestoreBackup(ctx context.Context, backupID string, writer io.Writer, opts *types.RestoreOptions) error {
	backupPath := filepath.Join(p.basePath, "backups", backupID)

	backupFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer backupFile.Close()

	_, err = io.Copy(writer, backupFile)
	if err != nil {
		return fmt.Errorf("failed to copy backup data: %w", err)
	}

	return nil
}

// ListBackups lists all available backups
func (p *LocalProvider) ListBackups(ctx context.Context) ([]*types.BackupMetadata, error) {
	metadataDir := filepath.Join(p.basePath, "metadata")

	// Check if metadata directory exists
	if _, err := os.Stat(metadataDir); os.IsNotExist(err) {
		return []*types.BackupMetadata{}, nil
	}

	// Read metadata directory
	entries, err := os.ReadDir(metadataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata directory: %w", err)
	}

	var backups []*types.BackupMetadata
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			metadataPath := filepath.Join(metadataDir, entry.Name())
			metadata, err := p.loadMetadata(metadataPath)
			if err != nil {
				continue // Skip invalid metadata files
			}
			backups = append(backups, metadata)
		}
	}

	return backups, nil
}

// GetBackup retrieves backup metadata
func (p *LocalProvider) GetBackup(ctx context.Context, backupID string) (*types.BackupMetadata, error) {
	metadataPath := filepath.Join(p.basePath, "metadata", backupID+".json")
	return p.loadMetadata(metadataPath)
}

// DeleteBackup removes a backup from local file system
func (p *LocalProvider) DeleteBackup(ctx context.Context, backupID string) error {
	// Delete backup file
	backupPath := filepath.Join(p.basePath, "backups", backupID)
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete backup file: %w", err)
	}

	// Delete metadata file
	metadataPath := filepath.Join(p.basePath, "metadata", backupID+".json")
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete metadata file: %w", err)
	}

	return nil
}

// HealthCheck checks if local file system is accessible
func (p *LocalProvider) HealthCheck(ctx context.Context) error {
	// Check if base path exists and is writable
	if _, err := os.Stat(p.basePath); os.IsNotExist(err) {
		// Try to create the directory
		if err := os.MkdirAll(p.basePath, 0755); err != nil {
			return fmt.Errorf("local provider health check failed: cannot create base directory: %w", err)
		}
	}

	// Test write access
	testFile := filepath.Join(p.basePath, ".health_check")
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("local provider health check failed: cannot write to directory: %w", err)
	}
	file.Close()
	os.Remove(testFile) // Clean up test file

	return nil
}

// Helper methods

func (p *LocalProvider) saveMetadata(path string, metadata *types.BackupMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func (p *LocalProvider) loadMetadata(path string) (*types.BackupMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()

	var metadata types.BackupMetadata
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &metadata, nil
}
