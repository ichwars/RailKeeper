package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type masterDataSeed struct {
	Entries   []seedEntry    `json:"entries"`
	Relations []seedRelation `json:"relations"`
}

type bundledMasterDataManifest struct {
	Version int                    `json:"version"`
	Entries []bundledMasterDataKey `json:"entries"`
}

type bundledMasterDataKey struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}

type seedEntry struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Key       string         `json:"key"`
	Label     string         `json:"label"`
	Active    bool           `json:"active"`
	SortOrder int            `json:"sortOrder"`
	SourceURL string         `json:"sourceUrl"`
	Metadata  map[string]any `json:"metadata"`
}

type seedRelation struct {
	ID         string `json:"id"`
	ParentType string `json:"parentType"`
	ParentKey  string `json:"parentKey"`
	ChildType  string `json:"childType"`
	ChildKey   string `json:"childKey"`
	SortOrder  int    `json:"sortOrder"`
}

func SeedMasterData(db *sql.DB, seedsDir string) error {
	if seedsDir == "" {
		return errors.New("seeds directory is required")
	}

	path := filepath.Join(seedsDir, "master_data.json")
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read master data seed: %w", err)
	}

	var seed masterDataSeed
	if err := json.Unmarshal(body, &seed); err != nil {
		return fmt.Errorf("parse master data seed: %w", err)
	}

	manifestPath := filepath.Join(seedsDir, "bundled_master_data_manifest.json")
	manifestBody, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read bundled master data manifest: %w", err)
	}
	var manifest bundledMasterDataManifest
	if err := json.Unmarshal(manifestBody, &manifest); err != nil {
		return fmt.Errorf("parse bundled master data manifest: %w", err)
	}
	if manifest.Version != 1 {
		return fmt.Errorf("unsupported bundled master data manifest version %d", manifest.Version)
	}
	for index, item := range manifest.Entries {
		if item.Type == "" || item.Key == "" {
			return fmt.Errorf("invalid bundled master data manifest entry %d", index)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin master data seed: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, item := range seed.Entries {
		metadata, err := json.Marshal(item.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata for %s: %w", item.ID, err)
		}
		if _, err = tx.ExecContext(ctx, `
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json,
  created_at, updated_at, origin
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'bundled')
ON CONFLICT(type, key) DO UPDATE SET origin='bundled'
`, item.ID, item.Type, item.Key, item.Label, boolToInt(item.Active), item.SortOrder,
			item.SourceURL, string(metadata), now, now); err != nil {
			return fmt.Errorf("seed master data %s: %w", item.ID, err)
		}
	}

	for _, item := range manifest.Entries {
		if _, err = tx.ExecContext(ctx, `
UPDATE master_data_entries SET origin='bundled' WHERE type=? AND key=?
`, item.Type, item.Key); err != nil {
			return fmt.Errorf("reconcile bundled master data %s/%s: %w", item.Type, item.Key, err)
		}
	}

	for _, relation := range seed.Relations {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO master_data_relations(id, parent_type, parent_key, child_type, child_key, sort_order, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(parent_type, parent_key, child_type, child_key) DO NOTHING
`, relation.ID, relation.ParentType, relation.ParentKey, relation.ChildType, relation.ChildKey, relation.SortOrder, now); err != nil {
			return fmt.Errorf("seed master data relation %s: %w", relation.ID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit master data seed: %w", err)
	}
	return nil
}
