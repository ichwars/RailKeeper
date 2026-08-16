package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrMasterDataBundled      = errors.New("bundled master data cannot be deleted")
	ErrMasterDataInUse        = fmt.Errorf("%w: master data entry is in use", ErrMasterDataValidation)
	ErrMasterDataUsageUnknown = errors.New("master data usage is unknown")
)

type masterDataQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type masterDataUsageValue int

const (
	masterDataUseLabel masterDataUsageValue = iota
	masterDataUseKey
)

type masterDataUsageRule struct {
	value masterDataUsageValue
	query string
}

var masterDataUsageRules = map[string][]masterDataUsageRule{
	"manufacturer": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE manufacturer=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM accessory_products WHERE manufacturer=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM exhibition_entries WHERE manufacturer=? COLLATE NOCASE)`},
	},
	"gauge": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE gauge=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(
  SELECT 1 FROM accessory_products products, json_each(products.gauges_json)
  WHERE json_each.value=? COLLATE NOCASE
)`},
	},
	"epoch": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE epoch=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM exhibition_entries WHERE epoch=? COLLATE NOCASE)`},
	},
	"railway_company": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE railway_company=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM exhibition_entries WHERE railway_company=? COLLATE NOCASE)`},
	},
	"vehicle_category": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE category=? COLLATE NOCASE)`},
	},
	"vehicle_gattung": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE gattung=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM exhibition_entries WHERE gattung=? COLLATE NOCASE)`},
	},
	"symbols": {
		{masterDataUseKey, `SELECT EXISTS(SELECT 1 FROM vehicle_functions WHERE symbol_key=?)`},
	},
	"stock_unit": {
		{masterDataUseKey, `SELECT EXISTS(SELECT 1 FROM accessory_products WHERE stock_unit=?)`},
	},
	"article_type": {
		{masterDataUseKey, `SELECT EXISTS(SELECT 1 FROM accessory_products WHERE article_type=?)`},
	},
	"accessory_subtype": {
		{masterDataUseKey, `SELECT EXISTS(SELECT 1 FROM accessory_products WHERE subtype=?)`},
	},
	"accessory_custom_field": {
		{masterDataUseKey, `SELECT EXISTS(
  SELECT 1 FROM accessory_product_attributes attributes
  JOIN accessory_products products ON products.id=attributes.product_id
  WHERE attributes.attribute_key=? AND products.article_type='other'
)`},
	},
	"cv8_manufacturer": {},
}

func (s *MasterDataService) SetActive(
	ctx context.Context,
	typeName, key string,
	active bool,
) (*MasterDataEntry, error) {
	typeName = strings.TrimSpace(typeName)
	key = strings.TrimSpace(key)
	if typeName == "" || key == "" {
		return nil, ErrMasterDataValidation
	}
	result, err := s.db.ExecContext(ctx, `
UPDATE master_data_entries SET active=?, updated_at=? WHERE type=? AND key=?
`, boolToInt(active), time.Now().UTC().Format(time.RFC3339), typeName, key)
	if err != nil {
		return nil, fmt.Errorf("set master data active state: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read master data active-state result: %w", err)
	}
	if affected == 0 {
		return nil, ErrMasterDataNotFound
	}
	s.invalidateCache()
	return s.Get(ctx, typeName, key)
}

func (s *MasterDataService) ListForManagement(
	ctx context.Context,
	typeName string,
) ([]MasterDataEntry, error) {
	entries, err := s.List(ctx, typeName, false)
	if err != nil {
		return nil, err
	}
	for index := range entries {
		capabilities, err := masterDataCapabilities(ctx, s.db, entries[index])
		if err != nil {
			return nil, err
		}
		entries[index].Capabilities = &capabilities
	}
	return entries, nil
}

func (s *MasterDataService) ListAllForManagement(
	ctx context.Context,
) (map[string][]MasterDataEntry, error) {
	entriesByType, err := s.ListAll(ctx, false)
	if err != nil {
		return nil, err
	}
	for typeName, entries := range entriesByType {
		for index := range entries {
			capabilities, err := masterDataCapabilities(ctx, s.db, entries[index])
			if err != nil {
				return nil, err
			}
			entries[index].Capabilities = &capabilities
		}
		entriesByType[typeName] = entries
	}
	return entriesByType, nil
}

func (s *MasterDataService) Delete(ctx context.Context, typeName, key string) error {
	typeName = strings.TrimSpace(typeName)
	key = strings.TrimSpace(key)
	if typeName == "" || key == "" {
		return ErrMasterDataValidation
	}
	if typeName == standardArticleType && isStandardArticleTypeKey(key) {
		return ErrMasterDataProtected
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin master data delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := reserveMasterDataWriteTransaction(ctx, tx); err != nil {
		return err
	}

	var entry MasterDataEntry
	if err := tx.QueryRowContext(ctx, `
SELECT type, key, label, origin FROM master_data_entries WHERE type=? AND key=?
`, typeName, key).Scan(&entry.Type, &entry.Key, &entry.Label, &entry.Origin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrMasterDataNotFound
		}
		return fmt.Errorf("load master data for deletion: %w", err)
	}
	if entry.Origin == MasterDataOriginBundled {
		return fmt.Errorf("%w: %s/%s", ErrMasterDataBundled, typeName, key)
	}
	used, err := masterDataIsUsed(ctx, tx, entry)
	if err != nil {
		return fmt.Errorf("check master data usage for %s/%s: %w", typeName, key, err)
	}
	if used {
		return fmt.Errorf("%w: referenced master data %s/%s cannot be deleted",
			ErrMasterDataInUse, typeName, key)
	}

	if _, err := tx.ExecContext(ctx, `
DELETE FROM master_data_relations
WHERE (parent_type=? AND parent_key=?) OR (child_type=? AND child_key=?)
`, typeName, key, typeName, key); err != nil {
		return fmt.Errorf("delete master data relations: %w", err)
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM master_data_entries WHERE type=? AND key=?`, typeName, key)
	if err != nil {
		return fmt.Errorf("delete master data: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read master data delete result: %w", err)
	}
	if affected == 0 {
		return ErrMasterDataNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit master data delete: %w", err)
	}
	s.invalidateCache()
	return nil
}

func isStandardArticleTypeKey(key string) bool {
	for _, standardKey := range standardArticleTypeKeys {
		if key == standardKey {
			return true
		}
	}
	return false
}

func masterDataCapabilities(
	ctx context.Context,
	queryer masterDataQueryer,
	entry MasterDataEntry,
) (MasterDataCapabilities, error) {
	capabilities := MasterDataCapabilities{
		CanDeactivate: entry.Active,
		CanReactivate: !entry.Active,
	}
	if entry.Origin != MasterDataOriginCustom {
		return capabilities, nil
	}
	used, err := masterDataIsUsed(ctx, queryer, entry)
	if errors.Is(err, ErrMasterDataUsageUnknown) {
		return capabilities, nil
	}
	if err != nil {
		return capabilities, err
	}
	capabilities.CanDelete = !used
	return capabilities, nil
}

func masterDataIsUsed(
	ctx context.Context,
	queryer masterDataQueryer,
	entry MasterDataEntry,
) (bool, error) {
	rules, known := masterDataUsageRules[entry.Type]
	if !known {
		return false, ErrMasterDataUsageUnknown
	}
	for _, rule := range rules {
		value := entry.Label
		if rule.value == masterDataUseKey {
			value = entry.Key
		}
		var used bool
		if err := queryer.QueryRowContext(ctx, rule.query, value).Scan(&used); err != nil {
			return false, fmt.Errorf("query %s usage: %w", entry.Type, err)
		}
		if used {
			return true, nil
		}
	}
	return false, nil
}
