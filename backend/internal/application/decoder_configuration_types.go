package application

// A working-state snapshot preserves recorded values and their individual sources. It is
// never evidence that an entire decoder was measured together at its recording time.
type DecoderCVSnapshotPayload struct {
	Values []VehicleCVValue `json:"values"`
}

type DecoderSnapshotHardware struct {
	DecoderID         string  `json:"decoderId"`
	DecoderLabel      string  `json:"decoderLabel"`
	Manufacturer      string  `json:"manufacturer"`
	Model             string  `json:"model"`
	SerialNumber      string  `json:"serialNumber"`
	InstallationID    string  `json:"installationId"`
	InstallationState string  `json:"installationState"`
	InstalledAt       *string `json:"installedAt"`
	RemovedAt         *string `json:"removedAt"`
}

type DecoderInstance struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	SerialNumber string `json:"serialNumber"`
	Notes        string `json:"notes"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type DecoderInstallation struct {
	ID                     string  `json:"id"`
	DecoderID              string  `json:"decoderId"`
	VehicleID              *string `json:"vehicleId"`
	VehicleReferenceID     string  `json:"vehicleReferenceId"`
	VehicleName            string  `json:"vehicleName"`
	VehicleInventoryNumber string  `json:"vehicleInventoryNumber"`
	State                  string  `json:"state"`
	InstalledAt            *string `json:"installedAt"`
	RemovedAt              *string `json:"removedAt"`
	Notes                  string  `json:"notes"`
	RecordedAt             string  `json:"recordedAt"`
	UpdatedAt              string  `json:"updatedAt"`
}

type DecoderCVWorkspace struct {
	ID              string  `json:"id"`
	VehicleID       string  `json:"vehicleId"`
	Profile         string  `json:"profile"`
	Origin          string  `json:"origin"`
	InstallationID  *string `json:"installationId"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
	ValueCount      int     `json:"valueCount"`
	SourceCount     int     `json:"sourceCount"`
	ContentRevision string  `json:"contentRevision"`
	BindingRevision string  `json:"bindingRevision"`
}

type DecoderCVSnapshotSummary struct {
	ID                     string  `json:"id"`
	WorkspaceID            string  `json:"workspaceId"`
	VehicleReferenceID     string  `json:"vehicleReferenceId"`
	VehicleName            string  `json:"vehicleName"`
	VehicleInventoryNumber string  `json:"vehicleInventoryNumber"`
	Profile                string  `json:"profile"`
	InstallationID         *string `json:"installationId"`
	Kind                   string  `json:"kind"`
	Scope                  string  `json:"scope"`
	Note                   string  `json:"note"`
	RecordedAt             string  `json:"recordedAt"`
	ValueCount             int     `json:"valueCount"`
}

type DecoderConfigurationContext struct {
	VehicleID       string                     `json:"vehicleId"`
	VehicleName     string                     `json:"vehicleName"`
	InventoryNumber string                     `json:"inventoryNumber"`
	Instances       []DecoderInstance          `json:"instances"`
	Installations   []DecoderInstallation      `json:"installations"`
	Workspaces      []DecoderCVWorkspace       `json:"workspaces"`
	Snapshots       []DecoderCVSnapshotSummary `json:"snapshots"`
	Revision        string                     `json:"revision"`
}
