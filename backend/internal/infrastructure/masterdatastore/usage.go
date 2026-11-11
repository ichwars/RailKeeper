package masterdatastore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"railkeeper/backend/internal/inventory/masterdata"
)

type masterDataUsageRule struct {
	source string
	match  string
}

var masterDataUsageRules = map[string][]masterDataUsageRule{
	"manufacturer": {
		{"vehicles", "manufacturer=requested.label COLLATE NOCASE"},
		{"accessory_products", "manufacturer=requested.label COLLATE NOCASE"},
		{"exhibition_entries", "manufacturer=requested.label COLLATE NOCASE"},
	},
	"gauge": {
		{"vehicles", "gauge=requested.label COLLATE NOCASE"},
		{"accessory_products products JOIN json_each(products.gauges_json) gauges", "gauges.value=requested.label COLLATE NOCASE"},
	},
	"epoch": {
		{"vehicles", "epoch=requested.label COLLATE NOCASE"},
		{"exhibition_entries", "epoch=requested.label COLLATE NOCASE"},
	},
	"railway_company": {
		{"vehicles", "railway_company=requested.label COLLATE NOCASE"},
		{"exhibition_entries", "railway_company=requested.label COLLATE NOCASE"},
	},
	"vehicle_category": {{"vehicles", "category=requested.label COLLATE NOCASE"}},
	"vehicle_gattung": {
		{"vehicles", "gattung=requested.label COLLATE NOCASE"},
		{"exhibition_entries", "gattung=requested.label COLLATE NOCASE"},
	},
	"symbols":           {{"vehicle_functions", "symbol_key=requested.key"}},
	"stock_unit":        {{"accessory_products", "stock_unit=requested.key"}},
	"article_type":      {{"accessory_products", "article_type=requested.key"}},
	"accessory_subtype": {{"accessory_products", "subtype=requested.key"}},
	"accessory_custom_field": {{
		"accessory_product_attributes attributes JOIN accessory_products products ON products.id=attributes.product_id",
		"attributes.attribute_key=requested.key AND products.article_type='other'",
	}},
	"cv8_manufacturer": {},
}

type masterDataQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func readMasterDataUsage(
	ctx context.Context, queryer masterDataQueryer, entries []masterdata.MasterDataEntry,
) (map[masterdata.Identity]bool, error) {
	usage := make(map[masterdata.Identity]bool, len(entries))
	var requested [][3]string
	types := map[string]bool{}
	for _, entry := range entries {
		rules, known := masterDataUsageRules[entry.Type]
		if !known {
			continue
		}
		usage[masterdata.IdentityOf(entry)] = false
		if len(rules) != 0 {
			requested = append(requested, [3]string{entry.Type, entry.Key, entry.Label})
			types[entry.Type] = true
		}
	}
	if len(requested) == 0 {
		return usage, nil
	}

	payload, err := json.Marshal(requested)
	if err != nil {
		return nil, err
	}
	var branches []string
	args := []any{string(payload)}
	for _, typeName := range slices.Sorted(maps.Keys(types)) {
		for _, rule := range masterDataUsageRules[typeName] {
			branches = append(branches, `
SELECT requested.type, requested.key
FROM requested JOIN `+rule.source+`
WHERE requested.type=? AND `+rule.match)
			args = append(args, typeName)
		}
	}
	rows, err := queryer.QueryContext(ctx, `
WITH requested(type, key, label) AS (
  SELECT value->>0, value->>1, value->>2 FROM json_each(?)
)
`+strings.Join(branches, "\nUNION\n"), args...)
	if err != nil {
		return nil, fmt.Errorf("query master data usage: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var entry masterdata.MasterDataEntry
		if err := rows.Scan(&entry.Type, &entry.Key); err != nil {
			return nil, fmt.Errorf("scan master data usage: %w", err)
		}
		usage[masterdata.IdentityOf(entry)] = true
	}
	return usage, rows.Err()
}

func masterDataIsUsed(ctx context.Context, queryer masterDataQueryer, entry masterdata.MasterDataEntry) (bool, error) {
	usage, err := readMasterDataUsage(ctx, queryer, []masterdata.MasterDataEntry{entry})
	if err != nil {
		return false, err
	}
	used, known := usage[masterdata.IdentityOf(entry)]
	if !known {
		return false, masterdata.ErrMasterDataUsageUnknown
	}
	return used, nil
}

func (s *Store) Usage(ctx context.Context, entries []masterdata.MasterDataEntry) (map[masterdata.Identity]bool, error) {
	return readMasterDataUsage(ctx, s.db, entries)
}
