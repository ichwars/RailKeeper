package transfer

type VehicleTransferValueKind string

const (
	VehicleTransferString  VehicleTransferValueKind = "string"
	VehicleTransferInteger VehicleTransferValueKind = "integer"
	VehicleTransferBoolean VehicleTransferValueKind = "boolean"
)

type VehicleTransferField struct {
	Key     string                   `json:"key"`
	LabelDE string                   `json:"labelDE"`
	LabelEN string                   `json:"labelEN"`
	Kind    VehicleTransferValueKind `json:"kind"`
	Aliases []string                 `json:"aliases"`
}

type DataTransferCSVMappingOrigin string

const (
	CSVMappingAlias    DataTransferCSVMappingOrigin = "alias"
	CSVMappingProfile  DataTransferCSVMappingOrigin = "profile"
	CSVMappingManual   DataTransferCSVMappingOrigin = "manual"
	CSVMappingIgnored  DataTransferCSVMappingOrigin = "ignored"
	CSVMappingUnmapped DataTransferCSVMappingOrigin = "unmapped"
)

type DataTransferCSVColumnMapping struct {
	Index            int                          `json:"index"`
	SourceHeader     string                       `json:"sourceHeader"`
	NormalizedHeader string                       `json:"normalizedHeader"`
	TargetField      string                       `json:"targetField"`
	Origin           DataTransferCSVMappingOrigin `json:"origin"`
}

type DataTransferCSVMappingInput struct {
	Columns       []DataTransferCSVColumnMapping `json:"columns"`
	SaveToProfile bool                           `json:"saveToProfile"`
}
