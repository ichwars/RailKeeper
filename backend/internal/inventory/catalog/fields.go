package catalog

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"railkeeper/backend/internal/revision"
)

// VehicleCatalogFields contains catalog descriptions only. Ownership, maintenance,
// decoder assignments and technical modifications remain on the owned vehicle.
type VehicleCatalogFields map[string]string

func (fields *VehicleCatalogFields) UnmarshalJSON(raw []byte) error {
	value, err := DecodeFields(string(raw), false)
	if err != nil {
		return err
	}
	*fields = value
	return nil
}

var vehicleCatalogFieldNames = []string{
	"name", "manufacturer", "articleNumber", "articleSourceUrl", "gauge", "epoch",
	"railwayCompany", "category", "gattung", "description", "ean", "productionPeriod", "listPrice",
}

func DecodeFields(raw string, complete bool) (VehicleCatalogFields, error) {
	var values map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &values); err != nil || values == nil {
		return nil, fmt.Errorf("catalog fields must be an object")
	}
	fields := make(VehicleCatalogFields, len(values))
	for key, value := range values {
		known := false
		for _, allowed := range vehicleCatalogFieldNames {
			if key == allowed {
				known = true
				break
			}
		}
		var text string
		if !known || string(value) == "null" || json.Unmarshal(value, &text) != nil {
			return nil, fmt.Errorf("invalid catalog field %q", key)
		}
		fields[key] = text
	}
	if complete && len(fields) != len(vehicleCatalogFieldNames) {
		return nil, fmt.Errorf("catalog revision must contain every catalog field")
	}
	return fields, nil
}
func Names() []string { return append([]string(nil), vehicleCatalogFieldNames...) }
func knownField(name string) bool {
	for _, field := range vehicleCatalogFieldNames {
		if field == name {
			return true
		}
	}
	return false
}

func validCatalogSelection(fields []string) bool {
	if fields == nil || len(fields) > len(vehicleCatalogFieldNames) {
		return false
	}
	seen := map[string]bool{}
	for _, field := range fields {
		if !knownField(field) || seen[field] {
			return false
		}
		seen[field] = true
	}
	return true
}
func validCatalogRequiredField(field, value string) bool {
	switch field {
	case "name", "manufacturer", "gauge", "category", "gattung":
		return strings.TrimSpace(value) != ""
	default:
		return true
	}
}
func validateCatalogRevisionInput(input VehicleCatalogRevisionInput) error {
	id, err := hex.DecodeString(input.ID)
	if err != nil || len(id) != 16 || hex.EncodeToString(id) != input.ID ||
		!revision.Valid(input.ExpectedRevision) || !validCatalogSelection(input.SelectedFields) ||
		len(input.Fields) != len(vehicleCatalogFieldNames) || utf8.RuneCountInString(input.Note) > 2000 {
		return ErrVehicleCatalogValidation
	}
	for field, value := range input.Fields {
		if !knownField(field) || !validCatalogRequiredField(field, value) {
			return ErrVehicleCatalogValidation
		}
	}
	return nil
}
