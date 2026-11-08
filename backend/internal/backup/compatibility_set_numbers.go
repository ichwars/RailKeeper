package backup

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/inventory/numbering"
)

func prepareVehicleSetInventoryNumbers(ctx context.Context, store LegacyRestoreStore, doc *BackupDocument) error {
	if doc.Version >= 18 {
		return nil
	}
	scheme, err := store.VehicleSetScheme(ctx)
	if errors.Is(err, numbering.ErrNotFound) {
		now := time.Now().UTC().Format(time.RFC3339)
		if err = store.CreateLegacyVehicleSetScheme(ctx, credentials.RandomID(), now); err != nil {
			return fmt.Errorf("create legacy vehicle set inventory scheme: %w", err)
		}
		scheme, err = store.VehicleSetScheme(ctx)
	}
	if err != nil {
		return fmt.Errorf("read legacy vehicle set inventory scheme: %w", err)
	}
	if !scheme.Active {
		if err = store.ActivateLegacyVehicleSetScheme(ctx, time.Now().UTC().Format(time.RFC3339)); err != nil {
			return fmt.Errorf("activate legacy vehicle set inventory scheme: %w", err)
		}
		scheme.Active = true
	}
	if len(doc.Tables["vehicle_sets"]) == 0 {
		return nil
	}
	originalRows := doc.Tables["vehicle_sets"]
	normalizedRows := make([]map[string]any, len(originalRows))
	for index, row := range originalRows {
		copyRow := make(map[string]any, len(row)+1)
		for column, value := range row {
			copyRow[column] = value
		}
		normalizedRows[index] = copyRow
	}
	sort.SliceStable(normalizedRows, func(left, right int) bool {
		leftCreated := fmt.Sprint(normalizedRows[left]["created_at"])
		rightCreated := fmt.Sprint(normalizedRows[right]["created_at"])
		if leftCreated != rightCreated {
			return leftCreated < rightCreated
		}
		return fmt.Sprint(normalizedRows[left]["id"]) < fmt.Sprint(normalizedRows[right]["id"])
	})
	for index, row := range normalizedRows {
		row["inventory_number"] = numbering.Format(scheme.Prefix, scheme.NextNumber+index, scheme.Padding)
	}
	doc.Tables["vehicle_sets"] = normalizedRows

	if err := store.AdvanceVehicleSetScheme(ctx, scheme.NextNumber+len(normalizedRows), time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("advance legacy vehicle set inventory scheme: %w", err)
	}
	return nil
}
