package backupdata

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"railkeeper/backend/internal/backup"
)

func (s *Store) exportTable(ctx context.Context, tx *sql.Tx, table string) ([]map[string]any, error) {
	rows, err := tx.QueryContext(ctx, "SELECT * FROM "+QuoteIdentifier(table))
	if err != nil {
		return nil, fmt.Errorf("export %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("read %s columns: %w", table, err)
	}

	out := []map[string]any{}
	for rows.Next() {
		values := make([]any, len(columns))
		targets := make([]any, len(columns))
		for i := range values {
			targets[i] = &values[i]
		}
		if err := rows.Scan(targets...); err != nil {
			return nil, fmt.Errorf("scan %s: %w", table, err)
		}
		row := map[string]any{}
		for i, column := range columns {
			row[column] = NormalizeBackupValue(table, column, values[i])
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s: %w", table, err)
	}
	return out, nil
}

func (s *Store) exportDataTransferProfiles(ctx context.Context, tx *sql.Tx) ([]map[string]any, error) {
	profiles, err := s.exportTable(ctx, tx, "data_transfer_profiles")
	if err != nil {
		return nil, err
	}
	for _, profile := range profiles {
		delete(profile, "name_key")
		creatorID, valid := backup.BackupNonEmptyString(profile["created_by_user_id"])
		if !valid {
			profile["created_by_user_id"] = nil
			continue
		}
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=?)`, creatorID).
			Scan(&exists); err != nil {
			return nil, fmt.Errorf("check exported transfer profile creator: %w", err)
		}
		if exists == 0 {
			profile["created_by_user_id"] = nil
		}
	}
	return profiles, nil
}

func TableColumns(ctx context.Context, tx *sql.Tx, table string) (map[string]bool, error) {
	rows, err := tx.QueryContext(ctx, "PRAGMA table_info("+QuoteIdentifier(table)+")")
	if err != nil {
		return nil, fmt.Errorf("read %s schema: %w", table, err)
	}
	defer func() { _ = rows.Close() }()

	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return nil, fmt.Errorf("scan %s schema: %w", table, err)
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s schema: %w", table, err)
	}
	return columns, nil
}

func InsertBackupRow(ctx context.Context, tx *sql.Tx, table string, allowedColumns map[string]bool, row map[string]any) (bool, error) {
	columns := make([]string, 0, len(row))
	for column := range row {
		if allowedColumns[column] {
			columns = append(columns, column)
		}
	}
	sort.Strings(columns)
	if len(columns) == 0 {
		return false, nil
	}

	placeholders := make([]string, len(columns))
	values := make([]any, len(columns))
	for i, column := range columns {
		placeholders[i] = "?"
		value, err := NormalizeImportValue(table, column, row[column])
		if err != nil {
			return false, err
		}
		values[i] = value
	}

	query := "INSERT INTO " + QuoteIdentifier(table) +
		" (" + strings.Join(QuoteIdentifiers(columns), ", ") + ")" +
		" VALUES (" + strings.Join(placeholders, ", ") + ")"
	if _, err := tx.ExecContext(ctx, query, values...); err != nil {
		return false, fmt.Errorf("insert %s: %w", table, err)
	}
	return true, nil
}

func NormalizeBackupValue(table, column string, value any) any {
	switch typed := value.(type) {
	case []byte:
		if table == "file_blobs" && column == "data" {
			return base64.StdEncoding.EncodeToString(typed)
		}
		return string(typed)
	default:
		return typed
	}
}

func NormalizeImportValue(table, column string, value any) (any, error) {
	if table == "file_blobs" && column == "data" {
		if encoded, ok := value.(string); ok {
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return nil, ErrBackupInvalid
			}
			return decoded, nil
		}
		return nil, ErrBackupInvalid
	}
	switch typed := value.(type) {
	case json.Number:
		if intValue, err := typed.Int64(); err == nil {
			return intValue, nil
		}
		if floatValue, err := typed.Float64(); err == nil {
			return floatValue, nil
		}
		return typed.String(), nil
	default:
		return typed, nil
	}
}

func QuoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func QuoteIdentifiers(values []string) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = QuoteIdentifier(value)
	}
	return out
}
