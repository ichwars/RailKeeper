package application

import (
	"encoding/json"
	"fmt"
)

// VehicleCatalogFields contains catalog descriptions only. Ownership, maintenance,
// decoder assignments and technical modifications remain on the owned vehicle.
type VehicleCatalogFields map[string]string

func (fields *VehicleCatalogFields) UnmarshalJSON(raw []byte) error {
	value, err := decodeVehicleCatalogFields(string(raw), false)
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

func decodeVehicleCatalogFields(raw string, complete bool) (VehicleCatalogFields, error) {
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
