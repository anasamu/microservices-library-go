package migrations

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/anasamu/microservices-library-go/database/gateway"
	"github.com/sirupsen/logrus"
)

// CLIManager handles CLI operations for migrations
type CLIManager struct {
	manager       *MigrationManager
	migrationsDir string
	logger        *logrus.Logger
}

// NewCLIManager creates a new CLI manager
func NewCLIManager(provider gateway.DatabaseProvider, migrationsDir string, logger *logrus.Logger) *CLIManager {
	return &CLIManager{
		manager:       NewMigrationManager(provider, logger),
		migrationsDir: migrationsDir,
		logger:        logger,
	}
}

// CreateMigration creates a new migration file
func (cm *CLIManager) CreateMigration(name string) error {
	// Generate version timestamp
	version := time.Now().Format("20060102150405")

	// Clean the name
	cleanName := strings.ReplaceAll(strings.ToLower(name), " ", "_")
	cleanName = strings.ReplaceAll(cleanName, "-", "_")

	// Create migration filename
	filename := fmt.Sprintf("%s_%s.json", version, cleanName)
	filepath := filepath.Join(cm.migrationsDir, filename)

	// Create migration structure
	migration := Migration{
		Version:     version,
		Description: name,
		UpSQL:       "-- Add your up migration SQL here",
		DownSQL:     "-- Add your down migration SQL here",
		CreatedAt:   time.Now(),
		Checksum:    "",
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(migration, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal migration: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write migration file: %w", err)
	}

	cm.logger.WithFields(logrus.Fields{
		"file":    filepath,
		"version": version,
	}).Info("Migration file created successfully")

	return nil
}

// LoadMigrations loads all migration files from the migrations directory
func (cm *CLIManager) LoadMigrations() ([]Migration, error) {
	var migrations []Migration

	// Read migration files
	files, err := os.ReadDir(cm.migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filepath := filepath.Join(cm.migrationsDir, file.Name())
		data, err := os.ReadFile(filepath)
		if err != nil {
			cm.logger.WithError(err).WithField("file", filepath).Warn("Failed to read migration file")
			continue
		}

		var migration Migration
		if err := json.Unmarshal(data, &migration); err != nil {
			cm.logger.WithError(err).WithField("file", filepath).Warn("Failed to parse migration file")
			continue
		}

		migrations = append(migrations, migration)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// Up applies all pending migrations
func (cm *CLIManager) Up(ctx context.Context) error {
	// Initialize migration table
	if err := cm.manager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize migration table: %w", err)
	}

	// Load available migrations
	migrations, err := cm.LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := cm.manager.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, migration := range appliedMigrations {
		appliedMap[migration.Version] = true
	}

	// Apply pending migrations
	appliedCount := 0
	for _, migration := range migrations {
		if appliedMap[migration.Version] {
			cm.logger.WithField("version", migration.Version).Debug("Migration already applied, skipping")
			continue
		}

		cm.logger.WithFields(logrus.Fields{
			"version":     migration.Version,
			"description": migration.Description,
		}).Info("Applying migration")

		if err := cm.manager.ApplyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		appliedCount++
	}

	cm.logger.WithField("count", appliedCount).Info("Migrations applied successfully")
	return nil
}

// Down rolls back the last migration
func (cm *CLIManager) Down(ctx context.Context) error {
	// Initialize migration table
	if err := cm.manager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize migration table: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := cm.manager.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(appliedMigrations) == 0 {
		cm.logger.Info("No migrations to rollback")
		return nil
	}

	// Get the last applied migration
	lastMigration := appliedMigrations[len(appliedMigrations)-1]

	cm.logger.WithFields(logrus.Fields{
		"version":     lastMigration.Version,
		"description": lastMigration.Description,
	}).Info("Rolling back migration")

	if err := cm.manager.RollbackMigration(ctx, lastMigration); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", lastMigration.Version, err)
	}

	cm.logger.Info("Migration rolled back successfully")
	return nil
}

// Status shows the status of all migrations
func (cm *CLIManager) Status(ctx context.Context) error {
	// Initialize migration table
	if err := cm.manager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize migration table: %w", err)
	}

	// Load available migrations
	migrations, err := cm.LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get migration status
	statuses, err := cm.manager.GetMigrationStatus(ctx, migrations)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	// Print status table
	fmt.Println("Migration Status:")
	fmt.Println("=================")
	fmt.Printf("%-20s %-50s %-10s %-20s\n", "Version", "Description", "Status", "Applied At")
	fmt.Println(strings.Repeat("-", 100))

	for _, status := range statuses {
		statusText := "Pending"
		appliedAt := "-"

		if status.Applied {
			statusText = "Applied"
			if status.AppliedAt != nil {
				appliedAt = status.AppliedAt.Format("2006-01-02 15:04:05")
			}
		}

		fmt.Printf("%-20s %-50s %-10s %-20s\n",
			status.Migration.Version,
			status.Migration.Description,
			statusText,
			appliedAt,
		)
	}

	return nil
}

// Reset drops all tables and reapplies all migrations
func (cm *CLIManager) Reset(ctx context.Context) error {
	cm.logger.Warn("Resetting database - this will drop all data!")

	// Get applied migrations in reverse order
	appliedMigrations, err := cm.manager.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Rollback all migrations
	for i := len(appliedMigrations) - 1; i >= 0; i-- {
		migration := appliedMigrations[i]

		cm.logger.WithFields(logrus.Fields{
			"version":     migration.Version,
			"description": migration.Description,
		}).Info("Rolling back migration")

		if err := cm.manager.RollbackMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.Version, err)
		}
	}

	// Reapply all migrations
	return cm.Up(ctx)
}

// Validate checks if all migration files are valid
func (cm *CLIManager) Validate() error {
	migrations, err := cm.LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	validCount := 0
	invalidCount := 0

	for _, migration := range migrations {
		if migration.Version == "" || migration.Description == "" {
			cm.logger.WithField("version", migration.Version).Error("Invalid migration: missing version or description")
			invalidCount++
			continue
		}

		if migration.UpSQL == "" {
			cm.logger.WithField("version", migration.Version).Error("Invalid migration: missing up SQL")
			invalidCount++
			continue
		}

		validCount++
	}

	cm.logger.WithFields(logrus.Fields{
		"valid":   validCount,
		"invalid": invalidCount,
	}).Info("Migration validation completed")

	if invalidCount > 0 {
		return fmt.Errorf("found %d invalid migrations", invalidCount)
	}

	return nil
}

// CreateMigrationsDirectory creates the migrations directory if it doesn't exist
func (cm *CLIManager) CreateMigrationsDirectory() error {
	if err := os.MkdirAll(cm.migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	cm.logger.WithField("directory", cm.migrationsDir).Info("Migrations directory created")
	return nil
}
