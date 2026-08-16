package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

type MigrationPlan struct {
	Applied []string
	Pending []string
	From    string
	To      string
}

type MigrationSafetyResult struct {
	Plan       MigrationPlan
	BackupPath string
}

type MigrationSafetyOptions struct {
	DatabaseExisted bool
	Now             func() time.Time
	Snapshot        func(context.Context, *sql.DB, string) error
}

func InspectMigrationPlan(ctx context.Context, db *sql.DB, migrationsDir string) (MigrationPlan, error) {
	if db == nil {
		return MigrationPlan{}, errors.New("database is required")
	}
	versions, err := migrationVersions(migrationsDir)
	if err != nil {
		return MigrationPlan{}, err
	}

	applied, err := appliedMigrationVersions(ctx, db)
	if err != nil {
		return MigrationPlan{}, err
	}
	appliedSet := make(map[string]struct{}, len(applied))
	for _, version := range applied {
		appliedSet[version] = struct{}{}
	}
	pending := make([]string, 0, len(versions))
	for _, version := range versions {
		if _, ok := appliedSet[version]; !ok {
			pending = append(pending, version)
		}
	}

	from := "unversioned"
	if len(applied) > 0 {
		from = applied[len(applied)-1]
	}
	to := from
	if len(versions) > 0 {
		to = versions[len(versions)-1]
	}
	return MigrationPlan{Applied: applied, Pending: pending, From: from, To: to}, nil
}

func MigrateSafely(
	ctx context.Context,
	db *sql.DB,
	dataDir string,
	migrationsDir string,
	options MigrationSafetyOptions,
) (MigrationSafetyResult, error) {
	plan, err := InspectMigrationPlan(ctx, db, migrationsDir)
	if err != nil {
		return MigrationSafetyResult{}, err
	}
	result := MigrationSafetyResult{Plan: plan}
	if options.DatabaseExisted && len(plan.Pending) > 0 {
		now := time.Now
		if options.Now != nil {
			now = options.Now
		}
		snapshot := CreateSQLiteSnapshot
		if options.Snapshot != nil {
			snapshot = options.Snapshot
		}
		backupDir := filepath.Join(dataDir, "safety-backups")
		if err := os.MkdirAll(backupDir, 0o700); err != nil {
			return MigrationSafetyResult{}, fmt.Errorf("create migration safety directory: %w", err)
		}
		backupName := fmt.Sprintf(
			"railkeeper-pre-migration-%s-to-%s-%s.db",
			safeMigrationName(plan.From),
			safeMigrationName(plan.To),
			now().UTC().Format("20060102T150405Z"),
		)
		result.BackupPath = filepath.Join(backupDir, backupName)
		if _, err := os.Stat(result.BackupPath); err == nil {
			return MigrationSafetyResult{}, fmt.Errorf("migration safety copy already exists: %s", result.BackupPath)
		} else if !errors.Is(err, os.ErrNotExist) {
			return MigrationSafetyResult{}, fmt.Errorf("inspect migration safety copy: %w", err)
		}
		if err := snapshot(ctx, db, result.BackupPath); err != nil {
			_ = os.Remove(result.BackupPath)
			return MigrationSafetyResult{}, fmt.Errorf("create migration safety copy: %w", err)
		}
		if err := ValidateSQLiteSnapshot(ctx, result.BackupPath); err != nil {
			_ = os.Remove(result.BackupPath)
			return MigrationSafetyResult{}, fmt.Errorf("validate migration safety copy: %w", err)
		}
		backup, err := sql.Open("sqlite", readOnlySQLiteDSN(result.BackupPath))
		if err != nil {
			return MigrationSafetyResult{}, fmt.Errorf("open migration safety copy: %w", err)
		}
		backup.SetMaxOpenConns(1)
		backupPlan, planErr := InspectMigrationPlan(ctx, backup, migrationsDir)
		closeErr := backup.Close()
		if planErr != nil {
			return MigrationSafetyResult{}, fmt.Errorf("inspect migration safety copy: %w", planErr)
		}
		if closeErr != nil {
			return MigrationSafetyResult{}, fmt.Errorf("close migration safety copy: %w", closeErr)
		}
		if !slices.Equal(backupPlan.Applied, plan.Applied) {
			return MigrationSafetyResult{}, errors.New("migration safety copy has a different migration state")
		}
	}

	if err := Migrate(db, migrationsDir); err != nil {
		return MigrationSafetyResult{}, err
	}
	return result, nil
}

func migrationVersions(migrationsDir string) ([]string, error) {
	if migrationsDir == "" {
		return nil, errors.New("migrations directory is required")
	}
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}
	versions := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		versions = append(versions, strings.TrimSuffix(entry.Name(), ".sql"))
	}
	sort.Strings(versions)
	return versions, nil
}

func appliedMigrationVersions(ctx context.Context, db *sql.DB) ([]string, error) {
	var tableCount int
	if err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM sqlite_schema WHERE type='table' AND name='schema_migrations'`,
	).Scan(&tableCount); err != nil {
		return nil, fmt.Errorf("inspect migration table: %w", err)
	}
	if tableCount == 0 {
		return []string{}, nil
	}
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("read applied migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	versions := []string{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}
	return versions, nil
}

func safeMigrationName(value string) string {
	var out strings.Builder
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' ||
			char >= '0' && char <= '9' || char == '_' || char == '-' || char == '.' {
			out.WriteRune(char)
		} else {
			out.WriteByte('-')
		}
	}
	if out.Len() == 0 {
		return "unversioned"
	}
	return out.String()
}
