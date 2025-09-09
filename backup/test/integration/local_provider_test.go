package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anasamu/microservices-library-go/backup/gateway"
	"github.com/anasamu/microservices-library-go/backup/providers/local"
	"github.com/anasamu/microservices-library-go/backup/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalProviderIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir) // Clean up

	// Create backup manager with local provider
	manager := gateway.NewBackupManager()
	localProvider := local.NewLocalProvider(tempDir)
	manager.SetProvider(localProvider)

	ctx := context.Background()

	t.Run("HealthCheck", func(t *testing.T) {
		err := manager.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("CreateAndRestoreBackup", func(t *testing.T) {
		// Create backup
		testData := "This is test backup data for integration testing"
		data := strings.NewReader(testData)
		opts := &types.BackupOptions{
			Compression: true,
			Tags: map[string]string{
				"test":     "integration",
				"category": "unit",
			},
			Description: "Integration test backup",
		}

		metadata, err := manager.CreateBackup(ctx, "integration-test", data, opts)
		require.NoError(t, err)
		assert.Equal(t, "integration-test", metadata.Name)
		assert.Equal(t, int64(len(testData)), metadata.Size)
		assert.Equal(t, "Integration test backup", metadata.Description)
		assert.Equal(t, "integration", metadata.Tags["test"])

		// Verify backup file exists
		backupPath := filepath.Join(tempDir, "backups", metadata.ID)
		_, err = os.Stat(backupPath)
		assert.NoError(t, err)

		// Verify metadata file exists
		metadataPath := filepath.Join(tempDir, "metadata", metadata.ID+".json")
		_, err = os.Stat(metadataPath)
		assert.NoError(t, err)

		// Restore backup
		var restoredData strings.Builder
		restoreOpts := &types.RestoreOptions{Overwrite: true}

		err = manager.RestoreBackup(ctx, metadata.ID, &restoredData, restoreOpts)
		require.NoError(t, err)
		assert.Equal(t, testData, restoredData.String())
	})

	t.Run("ListBackups", func(t *testing.T) {
		// Create multiple backups
		backupNames := []string{"backup1", "backup2", "backup3"}
		var createdBackups []*types.BackupMetadata

		for _, name := range backupNames {
			data := strings.NewReader("test data for " + name)
			opts := &types.BackupOptions{Description: "Test backup " + name}

			metadata, err := manager.CreateBackup(ctx, name, data, opts)
			require.NoError(t, err)
			createdBackups = append(createdBackups, metadata)
		}

		// List all backups
		backups, err := manager.ListBackups(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(backups), len(backupNames))

		// Verify all created backups are in the list
		backupMap := make(map[string]*types.BackupMetadata)
		for _, backup := range backups {
			backupMap[backup.ID] = backup
		}

		for _, createdBackup := range createdBackups {
			assert.Contains(t, backupMap, createdBackup.ID)
		}
	})

	t.Run("GetBackup", func(t *testing.T) {
		// Create a backup
		testData := "Test data for GetBackup"
		data := strings.NewReader(testData)
		opts := &types.BackupOptions{
			Tags: map[string]string{
				"operation": "get",
			},
			Description: "Get backup test",
		}

		createdMetadata, err := manager.CreateBackup(ctx, "get-test", data, opts)
		require.NoError(t, err)

		// Get the backup metadata
		retrievedMetadata, err := manager.GetBackup(ctx, createdMetadata.ID)
		require.NoError(t, err)

		assert.Equal(t, createdMetadata.ID, retrievedMetadata.ID)
		assert.Equal(t, createdMetadata.Name, retrievedMetadata.Name)
		assert.Equal(t, createdMetadata.Size, retrievedMetadata.Size)
		assert.Equal(t, createdMetadata.Description, retrievedMetadata.Description)
		assert.Equal(t, createdMetadata.Tags, retrievedMetadata.Tags)
	})

	t.Run("DeleteBackup", func(t *testing.T) {
		// Create a backup
		data := strings.NewReader("Test data for deletion")
		opts := &types.BackupOptions{Description: "Delete test backup"}

		metadata, err := manager.CreateBackup(ctx, "delete-test", data, opts)
		require.NoError(t, err)

		// Verify backup exists
		_, err = manager.GetBackup(ctx, metadata.ID)
		require.NoError(t, err)

		// Delete the backup
		err = manager.DeleteBackup(ctx, metadata.ID)
		require.NoError(t, err)

		// Verify backup is deleted
		_, err = manager.GetBackup(ctx, metadata.ID)
		assert.Error(t, err)

		// Verify files are deleted
		backupPath := filepath.Join(tempDir, "backups", metadata.ID)
		_, err = os.Stat(backupPath)
		assert.True(t, os.IsNotExist(err))

		metadataPath := filepath.Join(tempDir, "metadata", metadata.ID+".json")
		_, err = os.Stat(metadataPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("MultipleOperations", func(t *testing.T) {
		// Test a sequence of operations
		operations := []struct {
			name string
			data string
			desc string
		}{
			{"op1", "Operation 1 data", "First operation"},
			{"op2", "Operation 2 data", "Second operation"},
			{"op3", "Operation 3 data", "Third operation"},
		}

		var createdBackups []*types.BackupMetadata

		// Create backups
		for _, op := range operations {
			data := strings.NewReader(op.data)
			opts := &types.BackupOptions{Description: op.desc}

			metadata, err := manager.CreateBackup(ctx, op.name, data, opts)
			require.NoError(t, err)
			createdBackups = append(createdBackups, metadata)
		}

		// List and verify
		backups, err := manager.ListBackups(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(backups), len(operations))

		// Restore each backup
		for i, metadata := range createdBackups {
			var restoredData strings.Builder
			err = manager.RestoreBackup(ctx, metadata.ID, &restoredData, &types.RestoreOptions{})
			require.NoError(t, err)
			assert.Equal(t, operations[i].data, restoredData.String())
		}

		// Delete all backups
		for _, metadata := range createdBackups {
			err = manager.DeleteBackup(ctx, metadata.ID)
			require.NoError(t, err)
		}

		// Verify all are deleted
		finalBackups, err := manager.ListBackups(ctx)
		require.NoError(t, err)
		// Should have fewer backups now (other tests may have created some)
		assert.Less(t, len(finalBackups), len(backups))
	})
}
