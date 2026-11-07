package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"railkeeper/backend/internal/domain"
)

func validateBackupAccessoryProductAttributes(
	attributeRows []map[string]any,
	productRows []map[string]any,
	masterDataRows []map[string]any,
	backupVersion int,
) []string {
	productTypes := make(map[string]domain.AccessoryArticleType, len(productRows))
	for _, row := range productRows {
		productID, productIDValid := backupNonEmptyString(row["id"])
		articleType, articleTypeValid := backupNonEmptyString(row["article_type"])
		if productIDValid && articleTypeValid {
			productTypes[productID] = domain.AccessoryArticleType(articleType)
		}
	}

	validationErrors := []string{}
	controlledDefinitions := []domain.AccessoryAttributeDefinition{}
	if backupVersion >= 3 {
		controlledDefinitions, validationErrors = backupControlledAccessoryAttributeDefinitions(masterDataRows)
	}
	attributesByProduct := map[string][]domain.AccessoryAttributeValue{}
	for index, row := range attributeRows {
		if err := validateBackupAccessoryProductAttribute(row); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes enthält in Zeile %d ungültige Attributdaten: %v",
				index+1,
				err,
			))
			continue
		}
		productID, _ := backupNonEmptyString(row["product_id"])
		attributesByProduct[productID] = append(attributesByProduct[productID],
			backupDomainAccessoryAttribute(row))
	}

	productIDs := make([]string, 0, len(attributesByProduct))
	for productID := range attributesByProduct {
		productIDs = append(productIDs, productID)
	}
	sort.Strings(productIDs)
	for _, productID := range productIDs {
		articleType, exists := productTypes[productID]
		if !exists {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes verweist auf Produkt %s ohne gültigen article_type.",
				productID,
			))
			continue
		}
		var err error
		if backupVersion >= 3 && articleType == domain.AccessoryArticleOther {
			err = validateBackupControlledAccessoryAttributeValues(
				attributesByProduct[productID],
				controlledDefinitions,
			)
		} else {
			err = domain.ValidateAccessoryAttributeValues(articleType, attributesByProduct[productID])
		}
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes enthält für Produkt %s ungültige Attributdaten: %v",
				productID,
				err,
			))
		}
	}
	return validationErrors
}

func backupControlledAccessoryAttributeDefinitions(
	rows []map[string]any,
) ([]domain.AccessoryAttributeDefinition, []string) {
	definitions := []domain.AccessoryAttributeDefinition{}
	validationErrors := []string{}
	seen := map[string]struct{}{}
	for index, row := range rows {
		typeName, valid := backupNonEmptyString(row["type"])
		if !valid || typeName != accessoryCustomField {
			continue
		}
		key, keyValid := backupNonEmptyString(row["key"])
		active, activeValid := backupBooleanValue(row["active"])
		metadataJSON, metadataValid := row["metadata_json"].(string)
		metadata := map[string]any{}
		var metadataErr error
		if metadataValid {
			metadataErr = json.Unmarshal([]byte(metadataJSON), &metadata)
		}
		if !keyValid || !activeValid || !metadataValid || metadataErr != nil {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes kann wegen ungültiger Custom-Field-Definition in Zeile %d nicht validiert werden.",
				index+1,
			))
			continue
		}
		if _, duplicate := seen[key]; duplicate {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes enthält eine doppelte Custom-Field-Definition %s.", key,
			))
			continue
		}
		seen[key] = struct{}{}
		definition, err := ParseAccessoryCustomAttributeDefinition(key, active, metadata)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes kann Custom-Field-Definition %s nicht validieren: %v",
				key,
				err,
			))
			continue
		}
		definitions = append(definitions, definition)
	}
	return definitions, validationErrors
}

func validateBackupControlledAccessoryAttributeValues(
	values []domain.AccessoryAttributeValue,
	definitions []domain.AccessoryAttributeDefinition,
) error {
	if err := domain.ValidateAccessoryAttributeValues(domain.AccessoryArticleOther, values); err != nil {
		return err
	}
	activeDefinitions := make([]domain.AccessoryAttributeDefinition, 0, len(definitions))
	activeKeys := make(map[string]struct{}, len(definitions))
	inactiveKeys := make(map[string]struct{}, len(definitions))
	for _, definition := range definitions {
		if definition.Active {
			activeDefinitions = append(activeDefinitions, definition)
			activeKeys[definition.Key] = struct{}{}
		} else {
			inactiveKeys[definition.Key] = struct{}{}
		}
	}
	activeValues := make([]domain.AccessoryAttributeValue, 0, len(values))
	for _, value := range values {
		if _, active := activeKeys[value.Key]; active {
			activeValues = append(activeValues, value)
			continue
		}
		if _, inactive := inactiveKeys[value.Key]; !inactive {
			return fmt.Errorf("%w: undefined controlled key %q",
				domain.ErrAccessoryAttributeValidation, value.Key)
		}
	}
	return domain.ValidateControlledAccessoryAttributeValues(activeValues, activeDefinitions)
}

func backupDomainAccessoryAttribute(row map[string]any) domain.AccessoryAttributeValue {
	key, _ := backupNonEmptyString(row["attribute_key"])
	valueType, _ := backupNonEmptyString(row["value_type"])
	attribute := domain.AccessoryAttributeValue{
		Key:  key,
		Kind: domain.AccessoryAttributeKind(valueType),
	}
	if unit, ok := row["unit"].(string); ok {
		attribute.Unit = &unit
	}
	switch attribute.Kind {
	case domain.AccessoryAttributeText:
		value := row["text_value"].(string)
		attribute.TextValue = &value
	case domain.AccessoryAttributeNumber:
		value, _ := backupNumberValue(row["number_value"])
		attribute.NumberValue = &value
	case domain.AccessoryAttributeBoolean:
		value, _ := backupBooleanValue(row["boolean_value"])
		attribute.BooleanValue = &value
	case domain.AccessoryAttributeDate:
		value := row["date_value"].(string)
		attribute.DateValue = &value
	case domain.AccessoryAttributeSingleSelect:
		attribute.OptionValues = []string{row["single_select_value"].(string)}
	case domain.AccessoryAttributeMultiSelect:
		_ = json.Unmarshal([]byte(row["multi_select_value"].(string)), &attribute.OptionValues)
	}
	return attribute
}

func validateBackupAccessoryProductAttribute(row map[string]any) error {
	if _, valid := backupNonEmptyString(row["product_id"]); !valid {
		return errors.New("product_id fehlt oder ist ungültig")
	}
	if _, valid := backupNonEmptyString(row["attribute_key"]); !valid {
		return errors.New("attribute_key fehlt oder ist ungültig")
	}
	valueType, valueTypeValid := backupNonEmptyString(row["value_type"])
	valueColumnByType := map[string]string{
		"text":          "text_value",
		"number":        "number_value",
		"boolean":       "boolean_value",
		"date":          "date_value",
		"single_select": "single_select_value",
		"multi_select":  "multi_select_value",
	}
	matchingColumn, supported := valueColumnByType[valueType]
	if !valueTypeValid || !supported {
		return fmt.Errorf("value_type %q wird nicht unterstützt", valueType)
	}
	valueColumns := []string{
		"text_value", "number_value", "boolean_value", "date_value", "single_select_value", "multi_select_value",
	}
	for _, column := range valueColumns {
		if column == matchingColumn {
			if row[column] == nil {
				return fmt.Errorf("%s muss gesetzt sein", column)
			}
			continue
		}
		if row[column] != nil {
			return fmt.Errorf("%s darf für value_type %q nicht gesetzt sein", column, valueType)
		}
	}
	if row["unit"] != nil {
		if valueType != "number" {
			return fmt.Errorf("unit ist nur für value_type number erlaubt")
		}
		if _, ok := row["unit"].(string); !ok {
			return errors.New("unit muss eine Zeichenfolge sein")
		}
	}

	value := row[matchingColumn]
	switch valueType {
	case "text", "date", "single_select":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s muss eine Zeichenfolge sein", matchingColumn)
		}
	case "number":
		if !validBackupNumber(value) {
			return errors.New("number_value muss eine endliche Zahl sein")
		}
	case "boolean":
		if !validBackupBoolean(value) {
			return errors.New("boolean_value muss true, false, 0 oder 1 sein")
		}
	case "multi_select":
		encoded, ok := value.(string)
		if !ok {
			return errors.New("multi_select_value muss ein JSON-String sein")
		}
		var options []string
		if err := json.Unmarshal([]byte(encoded), &options); err != nil || len(options) == 0 {
			return errors.New("multi_select_value muss ein nicht-leeres JSON-Array aus Zeichenfolgen sein")
		}
	}
	return nil
}
