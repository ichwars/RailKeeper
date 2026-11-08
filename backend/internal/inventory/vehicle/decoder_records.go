package vehicle

import (
	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/domain"
)

type VehicleExternalMap struct {
	ID               string `json:"id"`
	VehicleID        string `json:"vehicleId"`
	Provider         string `json:"provider"`
	ExternalID       string `json:"externalId"`
	ExternalName     string `json:"externalName,omitempty"`
	ExternalAddress  string `json:"externalAddress,omitempty"`
	ExternalProtocol string `json:"externalProtocol,omitempty"`
	SyncStatus       string `json:"syncStatus"`
	LastSeenAt       string `json:"lastSeenAt,omitempty"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}
type VehicleExternalMapInput struct {
	Provider         string `json:"provider"`
	ExternalID       string `json:"externalId"`
	ExternalName     string `json:"externalName"`
	ExternalAddress  string `json:"externalAddress"`
	ExternalProtocol string `json:"externalProtocol"`
	SyncStatus       string `json:"syncStatus"`
}
type VehicleMaintenance = domain.VehicleMaintenance
type VehicleMaintenanceInput = domain.VehicleMaintenanceInput
type VehicleSparePart = domain.VehicleSparePart
type VehicleSparePartInput = domain.VehicleSparePartInput
type VehicleFunction struct {
	ID                       string `json:"id"`
	VehicleID                string `json:"vehicleId"`
	FunctionKey              string `json:"functionKey"`
	Name                     string `json:"name,omitempty"`
	SymbolKey                string `json:"symbolKey,omitempty"`
	FunctionType             string `json:"functionType"`
	Mode                     string `json:"mode"`
	DirectionDependent       *bool  `json:"directionDependent"`
	Inverted                 *bool  `json:"inverted"`
	LegacyDirectionDependent *bool  `json:"legacyDirectionDependent,omitempty"`
	FlagsVersion             int    `json:"flagsVersion"`
	Notes                    string `json:"notes,omitempty"`
	SortOrder                int    `json:"sortOrder"`
	CreatedAt                string `json:"createdAt"`
	UpdatedAt                string `json:"updatedAt"`
	Revision                 string `json:"revision"`
}
type VehicleFunctionInput struct {
	Name                     string `json:"name"`
	SymbolKey                string `json:"symbolKey"`
	FunctionType             string `json:"functionType"`
	Mode                     string `json:"mode"`
	DirectionDependent       *bool  `json:"directionDependent"`
	Inverted                 *bool  `json:"inverted"`
	LegacyDirectionDependent *bool  `json:"legacyDirectionDependent,omitempty"`
	FlagsVersion             int    `json:"flagsVersion"`
	Notes                    string `json:"notes"`
}

// An absent expected revision means the key must not exist yet.
type VehicleFunctionChange struct {
	FunctionKey      string                `json:"functionKey"`
	ExpectedRevision string                `json:"expectedRevision"`
	Value            *VehicleFunctionInput `json:"value,omitempty"`
	Delete           bool                  `json:"delete,omitempty"`
}
type VehicleCVValue = decoder.VehicleCVValue
type VehicleCVValueHistory = decoder.VehicleCVValueHistory
type VehicleCVValueInput = decoder.VehicleCVValueInput
type VehicleCVFile = decoder.VehicleCVFile
type VehicleCVFileInput = decoder.VehicleCVFileInput

// Source context is derived from the saved comparison workspace. It never
// replaces decoder_type or inserts a manufacturer/firmware CV reading.
type ECoSDecoderProfileEvidence = decoder.ECoSDecoderProfileEvidence
