package infrastructure

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"

	"modernc.org/sqlite"

	"railkeeper/backend/internal/application"
)

func init() {
	// SQLite's built-in lower function only handles ASCII. Use the same Unicode conversion as the Go search.
	sqlite.MustRegisterDeterministicScalarFunction("railkeeper_work_item_contains", 6, digitalCenterWorkItemContains)
}

func digitalCenterWorkItemContains(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
	needle, _ := args[0].([]byte)
	if len(needle) > 0 {
		query := string(needle)
		for _, value := range args[1:] {
			field, _ := value.([]byte)
			if strings.Contains(strings.ToLower(string(field)), query) {
				return int64(1), nil
			}
		}
	}
	return int64(0), nil
}

// ListWorkItemPage reads the totals and the requested rows in the same SQLite snapshot.
// The application supplies normalized pagination and a lowercased search term.
func (repository *DigitalCenterWorkspaceRepository) ListWorkItemPage(
	ctx context.Context,
	sessionID string,
	filter application.DigitalCenterWorkItemFilter,
) (application.DigitalCenterWorkItemPage, error) {
	match := "1"
	args := make([]any, 0, 5)
	if filter.CompareStatus != "" {
		match = "compare_status=?"
		args = append(args, filter.CompareStatus)
	}
	if filter.Query != "" {
		// Bind text as bytes so the callback receives the full value even for embedded NULs.
		match += ` AND railkeeper_work_item_contains(?, CAST(name AS BLOB), CAST(protocol AS BLOB),
			CAST(vehicle_id AS BLOB), CAST(center_object_id AS BLOB), CAST(decoder_address AS BLOB))`
		args = append(args, []byte(filter.Query))
	}

	tx, err := repository.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return application.DigitalCenterWorkItemPage{}, fmt.Errorf("begin digital center worklist snapshot: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	page := application.DigitalCenterWorkItemPage{Page: filter.Page, PageSize: filter.PageSize}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE `+match+`)
FROM digital_center_work_items WHERE session_id=?`, append(args, sessionID)...).Scan(&page.SessionTotal, &page.Total); err != nil {
		return application.DigitalCenterWorkItemPage{}, fmt.Errorf("count digital center work items: %w", err)
	}
	if page.Total > 0 {
		page.TotalPages = 1 + (page.Total-1)/page.PageSize
	}
	if page.Page > page.TotalPages {
		return page, nil
	}

	// Checking the page against the total first also bounds the offset multiplication.
	offset := (page.Page - 1) * page.PageSize
	pageArgs := append([]any{sessionID}, args...)
	pageArgs = append(pageArgs, page.PageSize, offset)
	rows, err := tx.QueryContext(ctx, digitalCenterWorkItemSelect+` WHERE session_id=? AND `+match+
		digitalCenterWorkItemOrder+` LIMIT ? OFFSET ?`, pageArgs...)
	if err != nil {
		return application.DigitalCenterWorkItemPage{}, fmt.Errorf("list digital center work item page: %w", err)
	}
	page.Items, err = scanDigitalCenterWorkItems(rows)
	if err != nil {
		return application.DigitalCenterWorkItemPage{}, err
	}
	return page, nil
}
