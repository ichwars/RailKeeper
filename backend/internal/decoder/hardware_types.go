package decoder

type DecoderInstanceInput struct {
	ID               string `json:"id"`
	ExpectedRevision string `json:"expectedRevision"`
	Label            string `json:"label"`
	Manufacturer     string `json:"manufacturer"`
	Model            string `json:"model"`
	SerialNumber     string `json:"serialNumber"`
	Notes            string `json:"notes"`
}

type DecoderInstallationInput struct {
	ID               string  `json:"id"`
	ExpectedRevision string  `json:"expectedRevision"`
	DecoderID        string  `json:"decoderId"`
	InstalledAt      *string `json:"installedAt"`
	Notes            string  `json:"notes"`
}

type DecoderRemovalInput struct {
	ExpectedRevision string  `json:"expectedRevision"`
	RemovedAt        *string `json:"removedAt"`
	Note             string  `json:"note"`
}

type DecoderInstallationDatesInput struct {
	ExpectedRevision string  `json:"expectedRevision"`
	InstalledAt      *string `json:"installedAt"`
	RemovedAt        *string `json:"removedAt"`
}
