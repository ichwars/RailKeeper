package infrastructure

import (
	"context"
	"fmt"
	"railkeeper/backend/internal/application"
	"time"
)

func (repository *DigitalCenterWorkspaceRepository) CreateWriteGrant(
	ctx context.Context,
	grant application.DigitalCenterWriteGrant,
) error {
	if grant.ID == "" {
		grant.ID = randomID()
	}
	if grant.CreatedAt == "" {
		grant.CreatedAt = timestamp()
	}
	result, err := repository.db.ExecContext(ctx, `
INSERT INTO digital_center_write_grants(
  id, token_hash, session_id, work_item_id, preview_hash, actor_user_id, expires_at, consumed_at, created_at
) SELECT ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?
WHERE EXISTS(
  SELECT 1 FROM digital_center_work_items WHERE id=? AND session_id=?
)`, grant.ID, grant.TokenHash, grant.SessionID, grant.WorkItemID, grant.PreviewHash, grant.ActorUserID,
		grant.ExpiresAt, grant.ConsumedAt, grant.CreatedAt, grant.WorkItemID, grant.SessionID)
	if err != nil {
		return fmt.Errorf("create digital center write grant: %w", err)
	}
	if err := requireDigitalCenterUpdate(result, "create digital center write grant"); err != nil {
		return err
	}
	return nil
}

func (repository *DigitalCenterWorkspaceRepository) ConsumeWriteGrant(
	ctx context.Context,
	tokenHash string,
	actorUserID string,
) (application.DigitalCenterWriteGrant, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return application.DigitalCenterWriteGrant{}, fmt.Errorf("consume digital center write grant: begin: %w", err)
	}
	rollback := func(operationErr error) (application.DigitalCenterWriteGrant, error) {
		_ = tx.Rollback()
		return application.DigitalCenterWriteGrant{}, operationErr
	}
	consumedAt := timestamp()
	result, err := tx.ExecContext(ctx, `
UPDATE digital_center_write_grants
SET consumed_at=?
WHERE token_hash=? AND consumed_at IS NULL AND actor_user_id=?
  AND julianday(expires_at) > julianday(?)`, consumedAt, tokenHash, actorUserID, consumedAt)
	if err != nil {
		return rollback(fmt.Errorf("consume digital center write grant: update: %w", err))
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return rollback(fmt.Errorf("consume digital center write grant: read update result: %w", err))
	}
	grant, loadErr := scanDigitalCenterWriteGrant(tx.QueryRowContext(ctx, digitalCenterWriteGrantSelect+`
WHERE token_hash=?`, tokenHash))
	if loadErr != nil {
		return rollback(fmt.Errorf("consume digital center write grant: load: %w", loadErr))
	}
	if affected != 1 {
		switch {
		case grant.ActorUserID != actorUserID:
			return rollback(fmt.Errorf("consume digital center write grant: %w",
				application.ErrDigitalCenterGrantActorMismatch))
		case grant.ConsumedAt != "":
			return rollback(fmt.Errorf("consume digital center write grant: %w",
				application.ErrDigitalCenterGrantConsumed))
		}
		expiresAt, parseErr := time.Parse(time.RFC3339, grant.ExpiresAt)
		if parseErr != nil {
			return rollback(fmt.Errorf("consume digital center write grant: parse expiry: %w", parseErr))
		}
		if !time.Now().UTC().Before(expiresAt) {
			return rollback(fmt.Errorf("consume digital center write grant: %w",
				application.ErrDigitalCenterGrantExpired))
		}
		return rollback(fmt.Errorf("consume digital center write grant: %w",
			application.ErrDigitalCenterGrantConsumed))
	}
	if err := tx.Commit(); err != nil {
		return application.DigitalCenterWriteGrant{}, fmt.Errorf("consume digital center write grant: commit: %w", err)
	}
	grant.ConsumedAt = consumedAt
	return grant, nil
}

const digitalCenterWriteGrantSelect = `
SELECT id, token_hash, session_id, work_item_id, preview_hash, actor_user_id, expires_at,
       COALESCE(consumed_at, ''), created_at
FROM digital_center_write_grants`

func scanDigitalCenterWriteGrant(scanner digitalCenterRowScanner) (application.DigitalCenterWriteGrant, error) {
	grant := application.DigitalCenterWriteGrant{}
	if err := scanner.Scan(&grant.ID, &grant.TokenHash, &grant.SessionID, &grant.WorkItemID,
		&grant.PreviewHash, &grant.ActorUserID, &grant.ExpiresAt, &grant.ConsumedAt,
		&grant.CreatedAt); err != nil {
		return application.DigitalCenterWriteGrant{}, err
	}
	return grant, nil
}
