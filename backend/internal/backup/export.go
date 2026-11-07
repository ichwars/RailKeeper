package backup

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *Service) Export(ctx context.Context) (*BackupDocument, error) {
	if !s.configured() {
		return nil, errors.New("backup service is not configured")
	}
	createdAt := time.Now().UTC().Format(time.RFC3339)
	snapshot, err := s.store.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	doc := &BackupDocument{Format: backupFormat, Version: backupVersion, CreatedAt: createdAt, Tables: map[string][]map[string]any{}, Profiles: snapshot.Profiles, Files: []BackupFile{}}
	for _, table := range backupTableOrder {
		rows, exists := snapshot.Tables[table]
		if !exists {
			return nil, fmt.Errorf("export %s: snapshot table missing", table)
		}
		doc.Tables[table] = rows
	}
	files, err := s.store.ExportFiles()
	if err != nil {
		return nil, err
	}
	doc.Files = files
	return doc, nil
}
