package gateway

import (
	"context"
	"fmt"
	"io"

	"github.com/anasamu/microservices-library-go/backup/types"
)

// BackupManager implements the backup management interface
type BackupManager struct {
	provider types.Provider
}

// NewBackupManager creates a new backup manager instance
func NewBackupManager() *BackupManager {
	return &BackupManager{}
}

// SetProvider sets the backup provider
func (bm *BackupManager) SetProvider(provider types.Provider) {
	bm.provider = provider
}

// GetProvider returns the current backup provider
func (bm *BackupManager) GetProvider() types.Provider {
	return bm.provider
}

// CreateBackup creates a backup using the current provider
func (bm *BackupManager) CreateBackup(ctx context.Context, name string, data io.Reader, opts *types.BackupOptions) (*types.BackupMetadata, error) {
	if bm.provider == nil {
		return nil, fmt.Errorf("no backup provider set")
	}

	return bm.provider.CreateBackup(ctx, name, data, opts)
}

// RestoreBackup restores a backup using the current provider
func (bm *BackupManager) RestoreBackup(ctx context.Context, backupID string, writer io.Writer, opts *types.RestoreOptions) error {
	if bm.provider == nil {
		return fmt.Errorf("no backup provider set")
	}

	return bm.provider.RestoreBackup(ctx, backupID, writer, opts)
}

// ListBackups lists all available backups
func (bm *BackupManager) ListBackups(ctx context.Context) ([]*types.BackupMetadata, error) {
	if bm.provider == nil {
		return nil, fmt.Errorf("no backup provider set")
	}

	return bm.provider.ListBackups(ctx)
}

// GetBackup retrieves backup metadata
func (bm *BackupManager) GetBackup(ctx context.Context, backupID string) (*types.BackupMetadata, error) {
	if bm.provider == nil {
		return nil, fmt.Errorf("no backup provider set")
	}

	return bm.provider.GetBackup(ctx, backupID)
}

// DeleteBackup removes a backup
func (bm *BackupManager) DeleteBackup(ctx context.Context, backupID string) error {
	if bm.provider == nil {
		return fmt.Errorf("no backup provider set")
	}

	return bm.provider.DeleteBackup(ctx, backupID)
}

// HealthCheck checks if the current provider is healthy
func (bm *BackupManager) HealthCheck(ctx context.Context) error {
	if bm.provider == nil {
		return fmt.Errorf("no backup provider set")
	}

	return bm.provider.HealthCheck(ctx)
}
