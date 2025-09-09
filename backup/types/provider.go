package types

import (
	"context"
	"io"
	"time"
)

// BackupMetadata contains metadata about a backup
type BackupMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Size        int64             `json:"size"`
	CreatedAt   time.Time         `json:"created_at"`
	Tags        map[string]string `json:"tags,omitempty"`
	Description string            `json:"description,omitempty"`
}

// BackupOptions contains options for backup operations
type BackupOptions struct {
	Compression bool              `json:"compression"`
	Encryption  bool              `json:"encryption"`
	Tags        map[string]string `json:"tags,omitempty"`
	Description string            `json:"description,omitempty"`
}

// RestoreOptions contains options for restore operations
type RestoreOptions struct {
	Overwrite bool `json:"overwrite"`
}

// Provider defines the interface for backup providers
type Provider interface {
	// CreateBackup creates a backup from the given data
	CreateBackup(ctx context.Context, name string, data io.Reader, opts *BackupOptions) (*BackupMetadata, error)

	// RestoreBackup restores a backup to the given writer
	RestoreBackup(ctx context.Context, backupID string, writer io.Writer, opts *RestoreOptions) error

	// ListBackups lists all available backups
	ListBackups(ctx context.Context) ([]*BackupMetadata, error)

	// GetBackup retrieves backup metadata
	GetBackup(ctx context.Context, backupID string) (*BackupMetadata, error)

	// DeleteBackup removes a backup
	DeleteBackup(ctx context.Context, backupID string) error

	// HealthCheck checks if the provider is healthy
	HealthCheck(ctx context.Context) error
}

// Manager defines the interface for backup management
type Manager interface {
	// SetProvider sets the backup provider
	SetProvider(provider Provider)

	// GetProvider returns the current backup provider
	GetProvider() Provider

	// CreateBackup creates a backup using the current provider
	CreateBackup(ctx context.Context, name string, data io.Reader, opts *BackupOptions) (*BackupMetadata, error)

	// RestoreBackup restores a backup using the current provider
	RestoreBackup(ctx context.Context, backupID string, writer io.Writer, opts *RestoreOptions) error

	// ListBackups lists all available backups
	ListBackups(ctx context.Context) ([]*BackupMetadata, error)

	// GetBackup retrieves backup metadata
	GetBackup(ctx context.Context, backupID string) (*BackupMetadata, error)

	// DeleteBackup removes a backup
	DeleteBackup(ctx context.Context, backupID string) error

	// HealthCheck checks if the current provider is healthy
	HealthCheck(ctx context.Context) error
}
