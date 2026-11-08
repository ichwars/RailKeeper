package backupdata

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/backup"
)

type compatibilityPlanner struct{}
type legacyRestoreStore struct{ tx *sql.Tx }
type compatibilityOperation struct {
	plan *backup.CompatibilityOperation
}

func NewCompatibilityPlanner() RestorePlanner { return compatibilityPlanner{} }
func (compatibilityPlanner) Begin(ctx context.Context, tx *sql.Tx, doc *BackupDocument) (RestoreOperation, error) {
	plan, err := backup.BeginCompatibilityRestore(ctx, &legacyRestoreStore{tx: tx}, doc)
	if err != nil {
		return nil, err
	}
	return &compatibilityOperation{plan: plan}, nil
}
func (op *compatibilityOperation) Rows(ctx context.Context, _ *sql.Tx, table string) ([]map[string]any, error) {
	return op.plan.Rows(ctx, table)
}
func (op *compatibilityOperation) ProfileRow(ctx context.Context, _ *sql.Tx, row map[string]any) (map[string]any, error) {
	return op.plan.ProfileRow(ctx, row)
}
func (op *compatibilityOperation) Finish(ctx context.Context, _ *sql.Tx, sourceRoot string) error {
	return op.plan.Finish(ctx, sourceRoot)
}
