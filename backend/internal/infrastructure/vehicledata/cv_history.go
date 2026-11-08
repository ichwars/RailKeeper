package vehicledata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"railkeeper/backend/internal/decoder"
)

func ReadCVHistory(ctx context.Context, query interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, cvValueID string) ([]decoder.VehicleCVValueHistory, error) {
	rows, err := query.QueryContext(ctx, `
SELECT id, cv_value_id, vehicle_id, old_value, new_value, changed_at
FROM vehicle_cv_value_history
WHERE cv_value_id=?
ORDER BY changed_at DESC
`, strings.TrimSpace(cvValueID))
	if err != nil {
		return nil, fmt.Errorf("list vehicle cv value history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []decoder.VehicleCVValueHistory{}
	for rows.Next() {
		var item decoder.VehicleCVValueHistory
		if err := rows.Scan(
			&item.ID,
			&item.CVValueID,
			&item.VehicleID,
			&item.OldValue,
			&item.NewValue,
			&item.ChangedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle cv value history: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle cv value history: %w", err)
	}
	return out, nil
}
