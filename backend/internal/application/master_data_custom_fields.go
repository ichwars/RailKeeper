package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/domain"
)

func (s *MasterDataService) updateAccessoryCustomField(
	ctx context.Context,
	key string,
	input MasterDataInput,
) (*MasterDataEntry, error) {
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	sortOrder := 0
	if input.SortOrder != nil {
		sortOrder = *input.SortOrder
	}
	metadata, err := normalizeAccessoryCustomFieldMetadata(key, active, input.Metadata)
	if err != nil {
		return nil, err
	}
	definition, err := ParseAccessoryCustomAttributeDefinition(key, active, metadata)
	if err != nil {
		return nil, err
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal master data metadata: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin custom field update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := reserveMasterDataWriteTransaction(ctx, tx); err != nil {
		return nil, err
	}
	if definition.Active {
		if err := validateAccessoryCustomFieldValues(ctx, tx, definition); err != nil {
			return nil, err
		}
	}
	result, err := tx.ExecContext(ctx, `
UPDATE master_data_entries
SET label=?, active=?, sort_order=?, source_url=?, metadata_json=?, updated_at=?
WHERE type=? AND key=?
`, input.Label, boolToInt(active), sortOrder, input.SourceURL, string(metadataJSON),
		time.Now().UTC().Format(time.RFC3339), accessoryCustomField, key)
	if err != nil {
		return nil, fmt.Errorf("update master data: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read master data update result: %w", err)
	}
	if affected == 0 {
		return nil, ErrMasterDataNotFound
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit custom field update: %w", err)
	}
	s.invalidateCache()
	return s.Get(ctx, accessoryCustomField, key)
}

func (s *MasterDataService) deleteAccessoryCustomField(ctx context.Context, key string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin custom field delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := reserveMasterDataWriteTransaction(ctx, tx); err != nil {
		return err
	}
	var referenced bool
	if err := tx.QueryRowContext(ctx, `
SELECT EXISTS(
    SELECT 1
    FROM accessory_product_attributes attributes
    JOIN accessory_products products ON products.id=attributes.product_id
    WHERE attributes.attribute_key=? AND products.article_type=?
)
`, key, domain.AccessoryArticleOther).Scan(&referenced); err != nil {
		return fmt.Errorf("check custom field references: %w", err)
	}
	if referenced {
		return fmt.Errorf("%w: referenced custom field %q cannot be deleted", ErrMasterDataValidation, key)
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM master_data_entries WHERE type=? AND key=?`,
		accessoryCustomField, key)
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
		return fmt.Errorf("commit custom field delete: %w", err)
	}
	s.invalidateCache()
	return nil
}

func validateImportedAccessoryCustomFieldReferences(
	ctx context.Context,
	tx *sql.Tx,
	entriesByType map[string][]MasterDataEntry,
) error {
	definitions := map[string]domain.AccessoryAttributeDefinition{}
	for typeName, entries := range entriesByType {
		for _, entry := range entries {
			if effectiveMasterDataType(typeName, entry) != accessoryCustomField {
				continue
			}
			key := strings.TrimSpace(entry.Key)
			if key == "" {
				key = slugKey(entry.Label)
			}
			if _, duplicate := definitions[key]; duplicate {
				return fmt.Errorf("%w: duplicate custom field %q", ErrMasterDataValidation, key)
			}
			definition, err := ParseAccessoryCustomAttributeDefinition(key, entry.Active, entry.Metadata)
			if err != nil {
				return err
			}
			definitions[key] = definition
		}
	}

	references, err := loadReferencedAccessoryCustomAttributes(ctx, tx, "")
	if err != nil {
		return err
	}
	for _, reference := range references {
		definition, exists := definitions[reference.Value.Key]
		if !exists {
			return fmt.Errorf("%w: referenced custom field %q is missing from import",
				ErrMasterDataValidation, reference.Value.Key)
		}
		if !definition.Active {
			continue
		}
		if err := domain.ValidateControlledAccessoryAttributeValues(
			[]domain.AccessoryAttributeValue{reference.Value},
			[]domain.AccessoryAttributeDefinition{definition},
		); err != nil {
			return fmt.Errorf("%w: referenced custom field %q rejects stored attribute for product %s: %v",
				ErrMasterDataValidation, definition.Key, reference.ProductID, err)
		}
	}
	return nil
}

func validateAccessoryCustomFieldValues(
	ctx context.Context,
	tx *sql.Tx,
	definition domain.AccessoryAttributeDefinition,
) error {
	references, err := loadReferencedAccessoryCustomAttributes(ctx, tx, definition.Key)
	if err != nil {
		return err
	}
	for _, reference := range references {
		if err := domain.ValidateControlledAccessoryAttributeValues(
			[]domain.AccessoryAttributeValue{reference.Value},
			[]domain.AccessoryAttributeDefinition{definition},
		); err != nil {
			return fmt.Errorf("%w: referenced custom field %q rejects stored attribute for product %s: %v",
				ErrMasterDataValidation, definition.Key, reference.ProductID, err)
		}
	}
	return nil
}

type referencedAccessoryCustomAttribute struct {
	ProductID string
	Value     domain.AccessoryAttributeValue
}

func loadReferencedAccessoryCustomAttributes(
	ctx context.Context,
	tx *sql.Tx,
	key string,
) ([]referencedAccessoryCustomAttribute, error) {
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

	references := []referencedAccessoryCustomAttribute{}
	for rows.Next() {
		var reference referencedAccessoryCustomAttribute
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
					ErrMasterDataValidation, reference.Value.Key)
			}
		}
		if err := reference.Value.Validate(); err != nil {
			return nil, fmt.Errorf("%w: referenced custom field %q has invalid stored data: %v",
				ErrMasterDataValidation, reference.Value.Key, err)
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
