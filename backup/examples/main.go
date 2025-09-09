package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/backup/gateway"
	"github.com/anasamu/microservices-library-go/backup/providers/local"
	"github.com/anasamu/microservices-library-go/backup/types"
)

func main() {
	// Create a new backup manager
	manager := gateway.NewBackupManager()

	// Create a local file provider
	localProvider := local.NewLocalProvider("/tmp/backups")

	// Set the provider
	manager.SetProvider(localProvider)

	// Example 1: Create a backup
	fmt.Println("=== Creating Backup ===")

	// Sample data to backup
	data := strings.NewReader("This is sample data to backup")

	// Backup options
	opts := &types.BackupOptions{
		Compression: true,
		Encryption:  false,
		Tags: map[string]string{
			"environment": "development",
			"service":     "example",
		},
		Description: "Sample backup for demonstration",
	}

	// Create backup
	metadata, err := manager.CreateBackup(context.Background(), "sample-backup", data, opts)
	if err != nil {
		log.Fatalf("Failed to create backup: %v", err)
	}

	fmt.Printf("Backup created successfully: %s (ID: %s)\n", metadata.Name, metadata.ID)
	fmt.Printf("Size: %d bytes, Created: %s\n", metadata.Size, metadata.CreatedAt.Format(time.RFC3339))

	// Example 2: List all backups
	fmt.Println("\n=== Listing Backups ===")

	backups, err := manager.ListBackups(context.Background())
	if err != nil {
		log.Fatalf("Failed to list backups: %v", err)
	}

	fmt.Printf("Found %d backups:\n", len(backups))
	for _, backup := range backups {
		fmt.Printf("- %s (ID: %s, Size: %d bytes)\n", backup.Name, backup.ID, backup.Size)
	}

	// Example 3: Get specific backup metadata
	fmt.Println("\n=== Getting Backup Metadata ===")

	backupMeta, err := manager.GetBackup(context.Background(), metadata.ID)
	if err != nil {
		log.Fatalf("Failed to get backup metadata: %v", err)
	}

	fmt.Printf("Backup Details:\n")
	fmt.Printf("  Name: %s\n", backupMeta.Name)
	fmt.Printf("  ID: %s\n", backupMeta.ID)
	fmt.Printf("  Size: %d bytes\n", backupMeta.Size)
	fmt.Printf("  Created: %s\n", backupMeta.CreatedAt.Format(time.RFC3339))
	fmt.Printf("  Description: %s\n", backupMeta.Description)
	fmt.Printf("  Tags: %v\n", backupMeta.Tags)

	// Example 4: Restore backup
	fmt.Println("\n=== Restoring Backup ===")

	var restoredData strings.Builder
	restoreOpts := &types.RestoreOptions{
		Overwrite: true,
	}

	err = manager.RestoreBackup(context.Background(), metadata.ID, &restoredData, restoreOpts)
	if err != nil {
		log.Fatalf("Failed to restore backup: %v", err)
	}

	fmt.Printf("Backup restored successfully. Content: %s\n", restoredData.String())

	// Example 5: Health check
	fmt.Println("\n=== Health Check ===")

	err = manager.HealthCheck(context.Background())
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}

	fmt.Println("Backup provider is healthy!")

	// Example 6: Delete backup
	fmt.Println("\n=== Deleting Backup ===")

	err = manager.DeleteBackup(context.Background(), metadata.ID)
	if err != nil {
		log.Fatalf("Failed to delete backup: %v", err)
	}

	fmt.Println("Backup deleted successfully!")

	// Verify deletion
	backups, err = manager.ListBackups(context.Background())
	if err != nil {
		log.Fatalf("Failed to list backups after deletion: %v", err)
	}

	fmt.Printf("Remaining backups: %d\n", len(backups))
}
