package masterdatastore

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/inventory/masterdata"
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

func masterDataIsUsed(
	ctx context.Context,
	queryer masterDataQueryer,
	entry masterdata.MasterDataEntry,
) (bool, error) {
	rules, known := masterDataUsageRules[entry.Type]
	if !known {
		return false, masterdata.ErrMasterDataUsageUnknown
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
func (s *Store) Usage(ctx context.Context, entry masterdata.MasterDataEntry) (bool, error) {
	return masterDataIsUsed(ctx, s.db, entry)
}
