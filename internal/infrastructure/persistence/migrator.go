// Package persistence provides database migration functionality
package persistence

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

// ============================================================================
// Migration Types
// ============================================================================

// MigrationDirection represents the direction of a migration
type MigrationDirection string

const (
	MigrationUp   MigrationDirection = "up"
	MigrationDown MigrationDirection = "down"
)

// Migration represents a database migration
type Migration struct {
	Version   string
	Name      string
	UpSQL     string
	DownSQL   string
	AppliedAt *time.Time
}

// MigrationResult represents the result of running migrations
type MigrationResult struct {
	Applied   []string
	Skipped   []string
	Failed    []string
	Error     error
	Duration  time.Duration
	Direction MigrationDirection
}

// ============================================================================
// Migrator
// ============================================================================

// Migrator handles database migrations
type Migrator struct {
	db         *gorm.DB
	migrations []Migration
	tableName  string
}

// NewMigrator creates a new Migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
		tableName:  "schema_migrations",
	}
}

// LoadMigrationsFromFS loads migrations from an embedded filesystem
func (m *Migrator) LoadMigrationsFromFS(fsys embed.FS, dir string) error {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrationMap := make(map[string]*Migration)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		content, err := fs.ReadFile(fsys, dir+"/"+name)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		// Parse migration filename: 000001_init_schema.up.sql
		parts := strings.Split(name, ".")
		if len(parts) < 3 {
			continue
		}

		version := parts[0]
		direction := parts[len(parts)-2]

		if _, ok := migrationMap[version]; !ok {
			migrationMap[version] = &Migration{
				Version: version,
				Name:    strings.TrimSuffix(version, "_"+strings.Join(parts[1:len(parts)-2], "_")),
			}
		}

		switch direction {
		case "up":
			migrationMap[version].UpSQL = string(content)
		case "down":
			migrationMap[version].DownSQL = string(content)
		}
	}

	// Convert map to sorted slice
	for _, migration := range migrationMap {
		m.migrations = append(m.migrations, *migration)
	}

	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	return nil
}

// AddMigration adds a migration to the migrator
func (m *Migrator) AddMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})
}

// EnsureMigrationTable ensures the migration tracking table exists
func (m *Migrator) EnsureMigrationTable() error {
	return m.db.AutoMigrate(&models.SchemaMigration{})
}

// GetAppliedMigrations returns all applied migrations
func (m *Migrator) GetAppliedMigrations() ([]models.SchemaMigration, error) {
	var applied []models.SchemaMigration
	if err := m.db.Order("version ASC").Find(&applied).Error; err != nil {
		return nil, err
	}
	return applied, nil
}

// IsMigrationApplied checks if a migration has been applied
func (m *Migrator) IsMigrationApplied(version string) (bool, error) {
	var count int64
	if err := m.db.Model(&models.SchemaMigration{}).Where("version = ?", version).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) (*MigrationResult, error) {
	start := time.Now()
	result := &MigrationResult{
		Applied:   make([]string, 0),
		Skipped:   make([]string, 0),
		Failed:    make([]string, 0),
		Direction: MigrationUp,
	}

	if err := m.EnsureMigrationTable(); err != nil {
		result.Error = err
		return result, err
	}

	for _, migration := range m.migrations {
		applied, err := m.IsMigrationApplied(migration.Version)
		if err != nil {
			result.Error = err
			result.Failed = append(result.Failed, migration.Version)
			return result, err
		}

		if applied {
			result.Skipped = append(result.Skipped, migration.Version)
			continue
		}

		log.Info().Str("version", migration.Version).Msg("Applying migration")

		if err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// Execute the up migration
			if migration.UpSQL != "" {
				if err := tx.Exec(migration.UpSQL).Error; err != nil {
					return fmt.Errorf("migration %s failed: %w", migration.Version, err)
				}
			}

			// Record the migration
			record := models.SchemaMigration{
				Version:   migration.Version,
				AppliedAt: time.Now(),
			}
			if err := tx.Create(&record).Error; err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
			}

			return nil
		}); err != nil {
			result.Error = err
			result.Failed = append(result.Failed, migration.Version)
			return result, err
		}

		result.Applied = append(result.Applied, migration.Version)
		log.Info().Str("version", migration.Version).Msg("Migration applied successfully")
	}

	result.Duration = time.Since(start)
	return result, nil
}

// Down rolls back the last migration
func (m *Migrator) Down(ctx context.Context) (*MigrationResult, error) {
	start := time.Now()
	result := &MigrationResult{
		Applied:   make([]string, 0),
		Skipped:   make([]string, 0),
		Failed:    make([]string, 0),
		Direction: MigrationDown,
	}

	applied, err := m.GetAppliedMigrations()
	if err != nil {
		result.Error = err
		return result, err
	}

	if len(applied) == 0 {
		log.Info().Msg("No migrations to roll back")
		result.Duration = time.Since(start)
		return result, nil
	}

	// Get the last applied migration
	lastApplied := applied[len(applied)-1]

	// Find the migration
	var migration *Migration
	for i := range m.migrations {
		if m.migrations[i].Version == lastApplied.Version {
			migration = &m.migrations[i]
			break
		}
	}

	if migration == nil {
		result.Error = fmt.Errorf("migration %s not found", lastApplied.Version)
		return result, result.Error
	}

	log.Info().Str("version", migration.Version).Msg("Rolling back migration")

	if err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Execute the down migration
		if migration.DownSQL != "" {
			if err := tx.Exec(migration.DownSQL).Error; err != nil {
				return fmt.Errorf("rollback %s failed: %w", migration.Version, err)
			}
		}

		// Remove the migration record
		if err := tx.Delete(&models.SchemaMigration{}, "version = ?", migration.Version).Error; err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", migration.Version, err)
		}

		return nil
	}); err != nil {
		result.Error = err
		result.Failed = append(result.Failed, migration.Version)
		return result, err
	}

	result.Applied = append(result.Applied, migration.Version)
	result.Duration = time.Since(start)
	log.Info().Str("version", migration.Version).Msg("Migration rolled back successfully")

	return result, nil
}

// DownTo rolls back to a specific version
func (m *Migrator) DownTo(ctx context.Context, targetVersion string) (*MigrationResult, error) {
	start := time.Now()
	result := &MigrationResult{
		Applied:   make([]string, 0),
		Skipped:   make([]string, 0),
		Failed:    make([]string, 0),
		Direction: MigrationDown,
	}

	for {
		applied, err := m.GetAppliedMigrations()
		if err != nil {
			result.Error = err
			return result, err
		}

		if len(applied) == 0 {
			break
		}

		lastApplied := applied[len(applied)-1]
		if lastApplied.Version <= targetVersion {
			break
		}

		singleResult, err := m.Down(ctx)
		if err != nil {
			result.Error = err
			result.Failed = append(result.Failed, singleResult.Failed...)
			return result, err
		}

		result.Applied = append(result.Applied, singleResult.Applied...)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// Reset rolls back all migrations
func (m *Migrator) Reset(ctx context.Context) (*MigrationResult, error) {
	start := time.Now()
	result := &MigrationResult{
		Applied:   make([]string, 0),
		Skipped:   make([]string, 0),
		Failed:    make([]string, 0),
		Direction: MigrationDown,
	}

	for {
		applied, err := m.GetAppliedMigrations()
		if err != nil {
			result.Error = err
			return result, err
		}

		if len(applied) == 0 {
			break
		}

		singleResult, err := m.Down(ctx)
		if err != nil {
			result.Error = err
			result.Failed = append(result.Failed, singleResult.Failed...)
			return result, err
		}

		result.Applied = append(result.Applied, singleResult.Applied...)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// Fresh drops all tables and re-runs migrations
func (m *Migrator) Fresh(ctx context.Context) (*MigrationResult, error) {
	// Reset all migrations
	if _, err := m.Reset(ctx); err != nil {
		return nil, err
	}

	// Run all migrations
	return m.Up(ctx)
}

// Status returns the status of all migrations
func (m *Migrator) Status() ([]Migration, error) {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	appliedMap := make(map[string]time.Time)
	for _, a := range applied {
		appliedMap[a.Version] = a.AppliedAt
	}

	result := make([]Migration, len(m.migrations))
	for i, migration := range m.migrations {
		result[i] = migration
		if appliedAt, ok := appliedMap[migration.Version]; ok {
			result[i].AppliedAt = &appliedAt
		}
	}

	return result, nil
}

// ============================================================================
// GORM Auto Migration
// ============================================================================

// AutoMigrate runs GORM auto-migration for all models
func AutoMigrate(db *gorm.DB) error {
	log.Info().Msg("Running GORM auto-migration...")

	// Enable UUID extension
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		log.Warn().Err(err).Msg("Failed to create uuid-ossp extension (may already exist)")
	}

	// Migrate all models
	if err := db.AutoMigrate(models.AllModels()...); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	log.Info().Msg("GORM auto-migration completed successfully")
	return nil
}

// MigrateWithGORM performs migrations using GORM's AutoMigrate
func MigrateWithGORM(db *gorm.DB) error {
	return AutoMigrate(db)
}
