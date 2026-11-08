package masterdatastore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/inventory/masterdata"
)

func (s *Store) UpdateCustomField(ctx context.Context, mutation masterdata.Mutation) error {
	input := mutation.Entry
	key := input.Key
	active, sortOrder := input.Active, input.SortOrder
	metadataJSON := mutation.MetadataJSON
	definition := *mutation.Definition

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin custom field update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := reserveMasterDataWriteTransaction(ctx, tx); err != nil {
		return err
	}
	if definition.Active {
		if err := masterdata.ValidateAccessoryCustomFieldValues(ctx, func(ctx context.Context, key string) ([]masterdata.ReferencedAttribute, error) {
			return loadReferencedAccessoryCustomAttributes(ctx, tx, key)
		}, definition); err != nil {
			return err
		}
	}
	result, err := tx.ExecContext(ctx, `
UPDATE master_data_entries
SET label=?, active=?, sort_order=?, source_url=?, metadata_json=?, updated_at=?
WHERE type=? AND key=?
`, input.Label, boolToInt(active), sortOrder, input.SourceURL, string(metadataJSON),
		time.Now().UTC().Format(time.RFC3339), masterdata.AccessoryCustomField, key)
	if err != nil {
		return fmt.Errorf("update master data: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read master data update result: %w", err)
	}
	if affected == 0 {
		return masterdata.ErrMasterDataNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit custom field update: %w", err)
	}
	return nil
}
func loadReferencedAccessoryCustomAttributes(
	ctx context.Context,
	tx *sql.Tx,
	key string,
) ([]masterdata.ReferencedAttribute, error) {
	query := `
SELECT attributes.product_id, attributes.attribute_key, attributes.value_type, attributes.text_value,
       attributes.number_value, attributes.boolean_value, attributes.date_value,
       attributes.single_select_value, attributes.multi_select_value, attributes.unit
FROM accessory_product_attributes attributes
JOIN accessory_products products ON products.id=attributes.product_id
WHERE products.article_type=?`
	args := []any{domain.AccessoryArticleOther}
	if key != "" {
		query += ` AND attributes.attribute_key=?`
		args = append(args, key)
	}
	query += ` ORDER BY attributes.product_id, attributes.attribute_key`
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list referenced custom field values: %w", err)
	}
	defer func() { _ = rows.Close() }()

	references := []masterdata.ReferencedAttribute{}
	for rows.Next() {
		var reference masterdata.ReferencedAttribute
		var kind string
		var textValue, dateValue, singleValue, multiValue, unit sql.NullString
		var numberValue sql.NullFloat64
		var booleanValue sql.NullInt64
		if err := rows.Scan(
			&reference.ProductID, &reference.Value.Key, &kind, &textValue, &numberValue, &booleanValue,
			&dateValue, &singleValue, &multiValue, &unit,
		); err != nil {
			return nil, fmt.Errorf("scan referenced custom field value: %w", err)
		}
		reference.Value.Kind = domain.AccessoryAttributeKind(kind)
		if unit.Valid {
			reference.Value.Unit = &unit.String
		}
		switch reference.Value.Kind {
		case domain.AccessoryAttributeText:
			if textValue.Valid {
				reference.Value.TextValue = &textValue.String
			}
		case domain.AccessoryAttributeNumber:
			if numberValue.Valid {
				reference.Value.NumberValue = &numberValue.Float64
			}
		case domain.AccessoryAttributeBoolean:
			if booleanValue.Valid {
				value := booleanValue.Int64 != 0
				reference.Value.BooleanValue = &value
			}
		case domain.AccessoryAttributeDate:
			if dateValue.Valid {
				reference.Value.DateValue = &dateValue.String
			}
		case domain.AccessoryAttributeSingleSelect:
			if singleValue.Valid {
				reference.Value.OptionValues = []string{singleValue.String}
			}
		case domain.AccessoryAttributeMultiSelect:
			if !multiValue.Valid || json.Unmarshal([]byte(multiValue.String), &reference.Value.OptionValues) != nil {
				return nil, fmt.Errorf("%w: referenced custom field %q has invalid multi-select data",
					masterdata.ErrMasterDataValidation, reference.Value.Key)
			}
		}
		if err := reference.Value.Validate(); err != nil {
			return nil, fmt.Errorf("%w: referenced custom field %q has invalid stored data: %v",
				masterdata.ErrMasterDataValidation, reference.Value.Key, err)
		}
		references = append(references, reference)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate referenced custom field values: %w", err)
	}
	return references, nil
}
func reserveMasterDataWriteTransaction(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `UPDATE master_data_entries SET updated_at=updated_at WHERE 0`); err != nil {
		return fmt.Errorf("reserve master data write transaction: %w", err)
	}
	return nil
}
