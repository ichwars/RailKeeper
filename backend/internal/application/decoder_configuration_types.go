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
