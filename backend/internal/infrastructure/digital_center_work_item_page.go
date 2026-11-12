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
	sqlite.MustRegisterFunction("railkeeper_work_item_contains", &sqlite.FunctionImpl{
		NArgs:         6,
		Deterministic: true,
		// The callback reads arguments only during the call. Avoid copying the query for each row.
		VolatileArgs: true,
		Scalar:       digitalCenterWorkItemContains,
	})
}

func digitalCenterWorkItemContains(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
	query, _ := args[0].(string)
	if query == "" {
		return int64(0), nil
	}
	for _, value := range args[1:] {
		field, _ := value.(string)
		if strings.Contains(strings.ToLower(field), query) {
			return int64(1), nil
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
		match += ` AND railkeeper_work_item_contains(?, CAST(name AS TEXT), CAST(protocol AS TEXT),
			CAST(vehicle_id AS TEXT), CAST(center_object_id AS TEXT), CAST(decoder_address AS TEXT))`
		args = append(args, filter.Query)
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
