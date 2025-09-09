package mocks

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/anasamu/microservices-library-go/backup/types"
)

// MockProvider is a mock implementation of the Provider interface for testing
type MockProvider struct {
	backups map[string]*types.BackupMetadata
	nextID  int
}

// NewMockProvider creates a new mock provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		backups: make(map[string]*types.BackupMetadata),
		nextID:  1,
	}
}

// CreateBackup creates a mock backup
func (m *MockProvider) CreateBackup(ctx context.Context, name string, data io.Reader, opts *types.BackupOptions) (*types.BackupMetadata, error) {
	// Simulate reading data to get size
	dataBytes, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	backupID := fmt.Sprintf("mock-backup-%d", m.nextID)
	m.nextID++

	metadata := &types.BackupMetadata{
		ID:          backupID,
		Name:        name,
		Size:        int64(len(dataBytes)),
		CreatedAt:   time.Now(),
		Tags:        opts.Tags,
		Description: opts.Description,
	}

	m.backups[backupID] = metadata
	return metadata, nil
}

// RestoreBackup restores a mock backup
func (m *MockProvider) RestoreBackup(ctx context.Context, backupID string, writer io.Writer, opts *types.RestoreOptions) error {
	metadata, exists := m.backups[backupID]
	if !exists {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// Write mock data
	mockData := fmt.Sprintf("Mock backup data for %s (size: %d bytes)", metadata.Name, metadata.Size)
	_, err := writer.Write([]byte(mockData))
	return err
}

// ListBackups lists all mock backups
func (m *MockProvider) ListBackups(ctx context.Context) ([]*types.BackupMetadata, error) {
	var backups []*types.BackupMetadata
	for _, backup := range m.backups {
		backups = append(backups, backup)
	}
	return backups, nil
}

// GetBackup retrieves mock backup metadata
func (m *MockProvider) GetBackup(ctx context.Context, backupID string) (*types.BackupMetadata, error) {
	metadata, exists := m.backups[backupID]
	if !exists {
		return nil, fmt.Errorf("backup not found: %s", backupID)
	}
	return metadata, nil
}

// DeleteBackup removes a mock backup
func (m *MockProvider) DeleteBackup(ctx context.Context, backupID string) error {
	if _, exists := m.backups[backupID]; !exists {
		return fmt.Errorf("backup not found: %s", backupID)
	}
	delete(m.backups, backupID)
	return nil
}

// HealthCheck always returns success for mock
func (m *MockProvider) HealthCheck(ctx context.Context) error {
	return nil
}

// Helper methods for testing

// GetBackupCount returns the number of backups in the mock
func (m *MockProvider) GetBackupCount() int {
	return len(m.backups)
}

// ClearBackups removes all backups from the mock
func (m *MockProvider) ClearBackups() {
	m.backups = make(map[string]*types.BackupMetadata)
	m.nextID = 1
}
