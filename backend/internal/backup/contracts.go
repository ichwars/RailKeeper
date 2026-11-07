package backup

import (
	"context"
)

type Snapshot struct {
	Tables   map[string][]map[string]any
	Profiles []map[string]any
}

type Store interface {
	Configured() bool
	Schema(context.Context, []string) (map[string]map[string]bool, error)
	Snapshot(context.Context) (Snapshot, error)
	ExportFiles() ([]BackupFile, error)
	Restore(context.Context, *BackupDocument) (*BackupImportResult, error)
}

type Rules interface {
	Table(*BackupDocument, string, []map[string]any, []string) ([]string, []string)
	Document(*BackupDocument) []string
}

type Service struct {
	store Store
	rules Rules
}

func New(store Store, rules Rules) *Service { return &Service{store: store, rules: rules} }
func (s *Service) configured() bool {
	return s != nil && s.store != nil && s.rules != nil && s.store.Configured()
}
