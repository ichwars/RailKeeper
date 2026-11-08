package catalog

import (
	"maps"
)

type FieldChange struct{ Name, Value string }
type Projection struct {
	Overrides       VehicleCatalogFields
	SourceAuthority *string
	Changes         []FieldChange
}

func PlanProjection(before *VehicleCatalogContext, chosen *VehicleCatalogRevision, selected []string) (Projection, error) {
	fields := maps.Clone(before.CurrentFields)
	for _, field := range selected {
		if !validCatalogRequiredField(field, chosen.Fields[field]) {
			return Projection{}, ErrVehicleCatalogValidation
		}
		fields[field] = chosen.Fields[field]
	}
	overrides := VehicleCatalogFields{}
	for field, value := range fields {
		if chosen.Fields[field] != value {
			overrides[field] = value
		}
	}
	source := before.SourceAuthority
	if source == nil && before.Authority == "legacy" {
		legacy := "legacy"
		source = &legacy
	}
	result := Projection{Overrides: overrides, SourceAuthority: source}
	for _, name := range vehicleCatalogFieldNames {
		if fields[name] != before.CurrentFields[name] {
			result.Changes = append(result.Changes, FieldChange{Name: name, Value: fields[name]})
		}
	}
	return result, nil
}
