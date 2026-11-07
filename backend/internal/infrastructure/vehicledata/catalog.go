package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// LinkCreatedSetMember is only for a member created by this same command.
// Reused inventory keeps its existing catalog identity.
func LinkCreatedSetMember(ctx context.Context, tx *sql.Tx, setID, vehicleID string) error {
	var revision string
	if err := tx.QueryRowContext(ctx, `SELECT revision_id FROM vehicle_set_article_links WHERE vehicle_set_id=?`, setID).Scan(&revision); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("read new set catalog: %w", err)
	}
	_, err := tx.ExecContext(ctx, `UPDATE vehicle_article_links SET revision_id=?,
overrides_json=(SELECT json_group_object(e.key, e.value)
 FROM vehicle_catalog_legacy_fields v, json_each(v.fields_json) e
 JOIN catalog_article_revisions r ON r.id=?
 WHERE v.entity_id=? AND e.value IS NOT json_extract(r.fields_json, '$.' || e.key)),
authority='catalog', source_authority=NULL, updated_at=strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE vehicle_id=?`, revision, revision, vehicleID, vehicleID)
	if err != nil {
		return fmt.Errorf("link new set member catalog: %w", err)
	}
	return nil
}
