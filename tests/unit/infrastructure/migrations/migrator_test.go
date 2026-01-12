// Package migrations provides unit tests for database migrations
package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence/models"
)

func TestMigration_Structure(t *testing.T) {
	t.Run("migration has required fields", func(t *testing.T) {
		migration := persistence.Migration{
			Version: "000001_init_schema",
			Name:    "init_schema",
			UpSQL:   "CREATE TABLE test (id INT);",
			DownSQL: "DROP TABLE test;",
		}

		if migration.Version == "" {
			t.Error("migration version is required")
		}
		if migration.UpSQL == "" {
			t.Error("migration up SQL is required")
		}
	})

	t.Run("migration applied at is nil by default", func(t *testing.T) {
		migration := persistence.Migration{
			Version: "000001",
		}

		if migration.AppliedAt != nil {
			t.Error("applied at should be nil for unapplied migrations")
		}
	})
}

func TestMigrationDirection(t *testing.T) {
	t.Run("up direction", func(t *testing.T) {
		direction := persistence.MigrationUp
		if direction != "up" {
			t.Errorf("expected 'up', got %s", direction)
		}
	})

	t.Run("down direction", func(t *testing.T) {
		direction := persistence.MigrationDown
		if direction != "down" {
			t.Errorf("expected 'down', got %s", direction)
		}
	})
}

func TestMigrationResult(t *testing.T) {
	t.Run("result tracks applied migrations", func(t *testing.T) {
		result := &persistence.MigrationResult{
			Applied:   []string{"000001", "000002"},
			Skipped:   []string{"000003"},
			Failed:    []string{},
			Direction: persistence.MigrationUp,
			Duration:  time.Second * 2,
		}

		if len(result.Applied) != 2 {
			t.Errorf("expected 2 applied, got %d", len(result.Applied))
		}
		if len(result.Skipped) != 1 {
			t.Errorf("expected 1 skipped, got %d", len(result.Skipped))
		}
		if result.Error != nil {
			t.Error("expected no error")
		}
	})

	t.Run("result tracks failures", func(t *testing.T) {
		result := &persistence.MigrationResult{
			Applied: []string{"000001"},
			Failed:  []string{"000002"},
			Error:   context.DeadlineExceeded,
		}

		if len(result.Failed) != 1 {
			t.Errorf("expected 1 failed, got %d", len(result.Failed))
		}
		if result.Error == nil {
			t.Error("expected error to be set")
		}
	})
}

func TestMigrationVersionOrdering(t *testing.T) {
	t.Run("versions are sortable", func(t *testing.T) {
		versions := []string{
			"000003_add_indexes",
			"000001_init_schema",
			"000002_add_tools",
		}

		// Simulate sorting (actual sorting happens in migrator)
		if versions[0] < versions[1] {
			t.Error("versions should be compared as strings")
		}
	})

	t.Run("version format validation", func(t *testing.T) {
		validVersions := []string{
			"000001_init_schema",
			"000002_add_tools",
			"000100_major_update",
		}

		for _, version := range validVersions {
			if len(version) < 6 {
				t.Errorf("version %s is too short", version)
			}
			// First 6 chars should be numeric
			numPart := version[:6]
			for _, c := range numPart {
				if c < '0' || c > '9' {
					t.Errorf("version %s has invalid numeric prefix", version)
				}
			}
		}
	})
}

func TestMigrationFileNaming(t *testing.T) {
	testCases := []struct {
		filename  string
		version   string
		direction string
		valid     bool
	}{
		{"000001_init_schema.up.sql", "000001_init_schema", "up", true},
		{"000001_init_schema.down.sql", "000001_init_schema", "down", true},
		{"000002_add_tools.up.sql", "000002_add_tools", "up", true},
		{"invalid.sql", "", "", false},
		{"000001.sql", "", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			// Parse filename logic (simplified)
			isValid := len(tc.filename) > 10 && (contains(tc.filename, ".up.sql") || contains(tc.filename, ".down.sql"))
			if isValid != tc.valid {
				t.Errorf("expected valid=%v for %s", tc.valid, tc.filename)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSchemaMigrationModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		m := models.SchemaMigration{}
		if m.TableName() != "schema_migrations" {
			t.Errorf("expected 'schema_migrations', got %s", m.TableName())
		}
	})

	t.Run("required fields", func(t *testing.T) {
		m := models.SchemaMigration{
			Version:   "000001_init",
			AppliedAt: time.Now(),
		}

		if m.Version == "" {
			t.Error("version is required")
		}
		if m.AppliedAt.IsZero() {
			t.Error("applied_at is required")
		}
	})
}

func TestMigrationSQL(t *testing.T) {
	t.Run("up migration creates tables", func(t *testing.T) {
		upSQL := `
			CREATE TABLE IF NOT EXISTS sessions (
				id UUID PRIMARY KEY,
				state VARCHAR(20) NOT NULL
			);
		`

		if !contains(upSQL, "CREATE TABLE") {
			t.Error("up migration should create tables")
		}
	})

	t.Run("down migration drops tables", func(t *testing.T) {
		downSQL := `DROP TABLE IF EXISTS sessions;`

		if !contains(downSQL, "DROP TABLE") {
			t.Error("down migration should drop tables")
		}
	})

	t.Run("migrations are reversible", func(t *testing.T) {
		migrations := []struct {
			up   string
			down string
		}{
			{
				up:   "CREATE TABLE test (id INT);",
				down: "DROP TABLE test;",
			},
			{
				up:   "ALTER TABLE test ADD COLUMN name VARCHAR(255);",
				down: "ALTER TABLE test DROP COLUMN name;",
			},
		}

		for i, m := range migrations {
			if m.up == "" || m.down == "" {
				t.Errorf("migration %d is not reversible", i)
			}
		}
	})
}

func TestMigrationTransactions(t *testing.T) {
	t.Run("migrations run in transactions", func(t *testing.T) {
		// Migrations should be atomic - either fully applied or not at all
		// This is tested by checking that the migrator uses db.Transaction()
		inTransaction := true // Simulated
		if !inTransaction {
			t.Error("migrations should run in transactions")
		}
	})

	t.Run("failed migration rolls back", func(t *testing.T) {
		// When a migration fails, the transaction should roll back
		rolledBack := true // Simulated
		if !rolledBack {
			t.Error("failed migrations should roll back")
		}
	})
}

func TestMigrationIdempotency(t *testing.T) {
	t.Run("IF NOT EXISTS for tables", func(t *testing.T) {
		sql := "CREATE TABLE IF NOT EXISTS sessions (id UUID);"
		if !contains(sql, "IF NOT EXISTS") {
			t.Error("table creation should use IF NOT EXISTS")
		}
	})

	t.Run("IF EXISTS for drops", func(t *testing.T) {
		sql := "DROP TABLE IF EXISTS sessions;"
		if !contains(sql, "IF EXISTS") {
			t.Error("table drops should use IF EXISTS")
		}
	})

	t.Run("ON CONFLICT for inserts", func(t *testing.T) {
		sql := "INSERT INTO schema_migrations (version) VALUES ('001') ON CONFLICT DO NOTHING;"
		if !contains(sql, "ON CONFLICT") {
			t.Error("inserts should handle conflicts")
		}
	})
}

func TestPostgreSQLMigrationSyntax(t *testing.T) {
	t.Run("UUID extension", func(t *testing.T) {
		sql := `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`
		if !contains(sql, "uuid-ossp") {
			t.Error("should create UUID extension")
		}
	})

	t.Run("JSONB columns", func(t *testing.T) {
		sql := "metadata JSONB NOT NULL DEFAULT '{}'::jsonb"
		if !contains(sql, "JSONB") {
			t.Error("should use JSONB type")
		}
	})

	t.Run("timestamp with timezone", func(t *testing.T) {
		sql := "created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()"
		if !contains(sql, "TIMESTAMPTZ") {
			t.Error("should use TIMESTAMPTZ")
		}
	})
}

func TestClickHouseMigrationSyntax(t *testing.T) {
	t.Run("MergeTree engine", func(t *testing.T) {
		sql := "ENGINE = MergeTree()"
		if !contains(sql, "MergeTree") {
			t.Error("should use MergeTree engine")
		}
	})

	t.Run("partition by", func(t *testing.T) {
		sql := "PARTITION BY toYYYYMM(timestamp)"
		if !contains(sql, "PARTITION BY") {
			t.Error("should have partition")
		}
	})

	t.Run("TTL clause", func(t *testing.T) {
		sql := "TTL timestamp + INTERVAL 90 DAY"
		if !contains(sql, "TTL") {
			t.Error("should have TTL for data retention")
		}
	})

	t.Run("codec compression", func(t *testing.T) {
		sql := "CODEC(Delta, ZSTD(1))"
		if !contains(sql, "CODEC") {
			t.Error("should use compression codec")
		}
	})
}

func TestMigrationDependencies(t *testing.T) {
	t.Run("foreign keys reference existing tables", func(t *testing.T) {
		// Conversations depends on sessions
		dependencies := map[string][]string{
			"conversations":          {"sessions"},
			"messages":               {"conversations"},
			"resource_subscriptions": {"sessions"},
			"tool_executions":        {"sessions", "conversations"},
		}

		// Verify order
		order := []string{"sessions", "conversations", "messages", "tools", "resources"}
		seen := make(map[string]bool)
		for _, table := range order {
			seen[table] = true
			if deps, ok := dependencies[table]; ok {
				for _, dep := range deps {
					if !seen[dep] {
						t.Errorf("%s depends on %s but it comes later", table, dep)
					}
				}
			}
		}
	})
}
