package backup

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/inventory/numbering"
)

func prepareArticleInventoryNumbers(
	ctx context.Context,
	store LegacyRestoreStore,
	doc *BackupDocument,
	fallback *ArticleInventoryScheme,
) error {
	rows := doc.Tables["accessory_products"]
	if len(rows) == 0 {
		return nil
	}
	scheme, err := ensureArticleScheme(ctx, store, fallback)
	if err != nil {
		return err
	}

	used := make(map[string]struct{}, len(rows))
	missing := make([]map[string]any, 0)
	for _, row := range rows {
		inventoryNumber := strings.TrimSpace(fmt.Sprint(row["inventory_number"]))
		if inventoryNumber == "<nil>" {
			inventoryNumber = ""
		}
		if inventoryNumber == "" {
			missing = append(missing, row)
			continue
		}
		if _, duplicate := used[inventoryNumber]; duplicate {
			return fmt.Errorf("prepare article inventory numbers: duplicate %q", inventoryNumber)
		}
		used[inventoryNumber] = struct{}{}
	}
	if len(missing) == 0 {
		return nil
	}
	if !scheme.Active {
		return numbering.ErrNotFound
	}

	sort.SliceStable(missing, func(i, j int) bool {
		leftCreated := fmt.Sprint(missing[i]["created_at"])
		rightCreated := fmt.Sprint(missing[j]["created_at"])
		if leftCreated != rightCreated {
			return leftCreated < rightCreated
		}
		return fmt.Sprint(missing[i]["id"]) < fmt.Sprint(missing[j]["id"])
	})

	next := 1
	for _, row := range missing {
		for {
			candidate := numbering.Format(scheme.Prefix, next, scheme.Padding)
			next++
			if _, exists := used[candidate]; exists {
				continue
			}
			row["inventory_number"] = candidate
			used[candidate] = struct{}{}
			break
		}
	}
	if next < scheme.NextNumber {
		next = scheme.NextNumber
	}
	if err := store.AdvanceArticleScheme(ctx, next, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("advance restored article inventory scheme: %w", err)
	}
	return nil
}

func ensureArticleScheme(
	ctx context.Context,
	store LegacyRestoreStore,
	fallback *ArticleInventoryScheme,
) (*ArticleInventoryScheme, error) {
	scheme, err := store.ArticleScheme(ctx)
	if err != nil || scheme != nil {
		return scheme, err
	}
	if fallback == nil {
		now := time.Now().UTC().Format(time.RFC3339)
		fallback = &ArticleInventoryScheme{
			ID: credentials.RandomID(), Prefix: "RK-ART", NextNumber: 1, Padding: 6, Active: true,
			CreatedAt: now, UpdatedAt: now,
		}
	}
	if err := store.InsertArticleScheme(ctx, *fallback); err != nil {
		return nil, fmt.Errorf("restore article inventory scheme: %w", err)
	}
	return store.ArticleScheme(ctx)
}
