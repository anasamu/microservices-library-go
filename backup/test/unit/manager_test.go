package unit

import (
	"context"
	"strings"
	"testing"

	"github.com/anasamu/microservices-library-go/backup"
	"github.com/anasamu/microservices-library-go/backup/test/mocks"
	"github.com/anasamu/microservices-library-go/backup/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupManager_CreateBackup(t *testing.T) {
	manager := gateway.NewBackupManager()
	mockProvider := mocks.NewMockProvider()
	manager.SetProvider(mockProvider)

	ctx := context.Background()
	data := strings.NewReader("test backup data")
	opts := &types.BackupOptions{
		Compression: true,
		Tags: map[string]string{
			"test": "true",
		},
		Description: "Test backup",
	}

	metadata, err := manager.CreateBackup(ctx, "test-backup", data, opts)

	require.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, "test-backup", metadata.Name)
	assert.Equal(t, int64(18), metadata.Size) // "test backup data" = 18 bytes
	assert.Equal(t, "Test backup", metadata.Description)
	assert.Equal(t, "true", metadata.Tags["test"])
	assert.NotEmpty(t, metadata.ID)
}

func TestBackupManager_CreateBackup_NoProvider(t *testing.T) {
	manager := gateway.NewBackupManager()

	ctx := context.Background()
	data := strings.NewReader("test data")
	opts := &types.BackupOptions{}

	_, err := manager.CreateBackup(ctx, "test-backup", data, opts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no backup provider set")
}

func TestBackupManager_RestoreBackup(t *testing.T) {
	manager := gateway.NewBackupManager()
	mockProvider := mocks.NewMockProvider()
	manager.SetProvider(mockProvider)

	ctx := context.Background()

	// First create a backup
	data := strings.NewReader("test restore data")
	opts := &types.BackupOptions{}
	metadata, err := manager.CreateBackup(ctx, "restore-test", data, opts)
	require.NoError(t, err)

	// Now restore it
	var restoredData strings.Builder
	restoreOpts := &types.RestoreOptions{Overwrite: true}

	err = manager.RestoreBackup(ctx, metadata.ID, &restoredData, restoreOpts)

	require.NoError(t, err)
	assert.Contains(t, restoredData.String(), "Mock backup data")
	assert.Contains(t, restoredData.String(), "restore-test")
}

func TestBackupManager_ListBackups(t *testing.T) {
	manager := gateway.NewBackupManager()
	mockProvider := mocks.NewMockProvider()
	manager.SetProvider(mockProvider)

	ctx := context.Background()

	// Create multiple backups
	for i := 0; i < 3; i++ {
		data := strings.NewReader("test data")
		opts := &types.BackupOptions{}
		_, err := manager.CreateBackup(ctx, "test-backup", data, opts)
		require.NoError(t, err)
	}

	backups, err := manager.ListBackups(ctx)

	require.NoError(t, err)
	assert.Len(t, backups, 3)
}

func TestBackupManager_GetBackup(t *testing.T) {
	manager := gateway.NewBackupManager()
	mockProvider := mocks.NewMockProvider()
	manager.SetProvider(mockProvider)

	ctx := context.Background()

	// Create a backup
	data := strings.NewReader("test data")
	opts := &types.BackupOptions{Description: "Test backup"}
	metadata, err := manager.CreateBackup(ctx, "get-test", data, opts)
	require.NoError(t, err)

	// Get the backup
	retrievedMetadata, err := manager.GetBackup(ctx, metadata.ID)

	require.NoError(t, err)
	assert.Equal(t, metadata.ID, retrievedMetadata.ID)
	assert.Equal(t, metadata.Name, retrievedMetadata.Name)
	assert.Equal(t, metadata.Description, retrievedMetadata.Description)
}

func TestBackupManager_DeleteBackup(t *testing.T) {
	manager := gateway.NewBackupManager()
	mockProvider := mocks.NewMockProvider()
	manager.SetProvider(mockProvider)

	ctx := context.Background()

	// Create a backup
	data := strings.NewReader("test data")
	opts := &types.BackupOptions{}
	metadata, err := manager.CreateBackup(ctx, "delete-test", data, opts)
	require.NoError(t, err)

	// Verify it exists
	_, err = manager.GetBackup(ctx, metadata.ID)
	require.NoError(t, err)

	// Delete it
	err = manager.DeleteBackup(ctx, metadata.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = manager.GetBackup(ctx, metadata.ID)
	assert.Error(t, err)
}

func TestBackupManager_HealthCheck(t *testing.T) {
	manager := gateway.NewBackupManager()
	mockProvider := mocks.NewMockProvider()
	manager.SetProvider(mockProvider)

	ctx := context.Background()

	err := manager.HealthCheck(ctx)

	require.NoError(t, err)
}

func TestBackupManager_HealthCheck_NoProvider(t *testing.T) {
	manager := gateway.NewBackupManager()

	ctx := context.Background()

	err := manager.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no backup provider set")
}

func TestBackupManager_SetGetProvider(t *testing.T) {
	manager := gateway.NewBackupManager()
	mockProvider := mocks.NewMockProvider()

	// Initially no provider
	assert.Nil(t, manager.GetProvider())

	// Set provider
	manager.SetProvider(mockProvider)
	assert.Equal(t, mockProvider, manager.GetProvider())
}
