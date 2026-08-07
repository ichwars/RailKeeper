package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

func accessoryProductJSON(input application.CreateAccessoryProductInput) (string, string, string, error) {
	encode := func(name string, values []string) (string, error) {
		encoded, err := json.Marshal(values)
		if err != nil {
			return "", fmt.Errorf("encode accessory %s: %w", name, err)
		}
		return string(encoded), nil
	}
	gauges, err := encode("gauges", input.Gauges)
	if err != nil {
		return "", "", "", err
	}
	alternativeNumbers, err := encode("alternative numbers", input.AlternativeNumbers)
	if err != nil {
		return "", "", "", err
	}
	keywords, err := encode("keywords", input.Keywords)
	return gauges, alternativeNumbers, keywords, err
}

func decodeAccessoryStringArray(value string, target *[]string) error {
	if err := json.Unmarshal([]byte(value), target); err != nil {
		return fmt.Errorf("decode accessory string array: %w", err)
	}
	if *target == nil {
		*target = []string{}
	}
	return nil
}

func replaceAccessoryAttributes(
	ctx context.Context,
	tx *sql.Tx,
	productID string,
	attributes []domain.AccessoryAttributeValue,
	now string,
) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM accessory_product_attributes WHERE product_id=?`, productID); err != nil {
		return fmt.Errorf("delete accessory attributes: %w", err)
	}
	for _, attribute := range attributes {
		var textValue, dateValue, singleSelectValue, multiSelectValue any
		var numberValue, unit, booleanValue any
		switch attribute.Kind {
		case domain.AccessoryAttributeText:
			textValue = *attribute.TextValue
		case domain.AccessoryAttributeNumber:
			numberValue = *attribute.NumberValue
			if attribute.Unit != nil {
				unit = *attribute.Unit
			}
		case domain.AccessoryAttributeBoolean:
			booleanValue = boolToInt(*attribute.BooleanValue)
		case domain.AccessoryAttributeDate:
			dateValue = *attribute.DateValue
		case domain.AccessoryAttributeSingleSelect:
			singleSelectValue = attribute.OptionValues[0]
		case domain.AccessoryAttributeMultiSelect:
			encoded, err := json.Marshal(attribute.OptionValues)
			if err != nil {
				return fmt.Errorf("encode accessory multi-select attribute: %w", err)
			}
			multiSelectValue = string(encoded)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_product_attributes(
  id, product_id, attribute_key, value_type, text_value, number_value, unit, boolean_value,
  date_value, single_select_value, multi_select_value, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, randomID(), productID, attribute.Key, attribute.Kind,
			textValue, numberValue, unit, booleanValue, dateValue, singleSelectValue, multiSelectValue, now, now); err != nil {
			return fmt.Errorf("insert accessory attribute: %w", err)
		}
	}
	return nil
}

type accessoryRowsQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func loadAccessoryAttributes(
	ctx context.Context,
	queryer accessoryRowsQueryer,
	productIDs []string,
) (map[string][]domain.AccessoryAttributeValue, error) {
	result := make(map[string][]domain.AccessoryAttributeValue, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(productIDs)), ",")
	args := make([]any, len(productIDs))
	for index, id := range productIDs {
		args[index] = id
		result[id] = []domain.AccessoryAttributeValue{}
	}
	rows, err := queryer.QueryContext(ctx, `
SELECT product_id, attribute_key, value_type, text_value, number_value, unit, boolean_value,
       date_value, single_select_value, multi_select_value
FROM accessory_product_attributes
WHERE product_id IN (`+placeholders+`)
ORDER BY product_id, rowid`, args...)
	if err != nil {
		return nil, fmt.Errorf("load accessory attributes: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var productID string
		var attribute domain.AccessoryAttributeValue
		var text, unit, date, single, multi sql.NullString
		var number sql.NullFloat64
		var boolean sql.NullInt64
		if err := rows.Scan(&productID, &attribute.Key, &attribute.Kind, &text, &number, &unit, &boolean,
			&date, &single, &multi); err != nil {
			return nil, fmt.Errorf("scan accessory attribute: %w", err)
		}
		switch attribute.Kind {
		case domain.AccessoryAttributeText:
			attribute.TextValue = &text.String
		case domain.AccessoryAttributeNumber:
			attribute.NumberValue = &number.Float64
			if unit.Valid {
				attribute.Unit = &unit.String
			}
		case domain.AccessoryAttributeBoolean:
			value := boolean.Int64 != 0
			attribute.BooleanValue = &value
		case domain.AccessoryAttributeDate:
			attribute.DateValue = &date.String
		case domain.AccessoryAttributeSingleSelect:
			attribute.OptionValues = []string{single.String}
		case domain.AccessoryAttributeMultiSelect:
			if err := json.Unmarshal([]byte(multi.String), &attribute.OptionValues); err != nil {
				return nil, fmt.Errorf("decode accessory multi-select attribute: %w", err)
			}
		}
		result[productID] = append(result[productID], attribute)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory attributes: %w", err)
	}
	return result, nil
}
