package transfer

type TransferAccessory struct {
	ID                 string                            `json:"id,omitempty"`
	InventoryNumber    string                            `json:"inventoryNumber"`
	Manufacturer       string                            `json:"manufacturer"`
	ArticleNumber      string                            `json:"articleNumber,omitempty"`
	Name               string                            `json:"name"`
	Category           string                            `json:"category"`
	TrackingMode       string                            `json:"trackingMode"`
	Description        string                            `json:"description,omitempty"`
	EAN                string                            `json:"ean,omitempty"`
	ManufacturerStatus string                            `json:"manufacturerStatus,omitempty"`
	ArticleType        string                            `json:"articleType"`
	Subtype            string                            `json:"subtype,omitempty"`
	Gauges             []string                          `json:"gauges"`
	Scale              string                            `json:"scale,omitempty"`
	ListPrice          string                            `json:"listPrice,omitempty"`
	PackageQuantity    int                               `json:"packageQuantity"`
	StockUnit          string                            `json:"stockUnit"`
	MinimumStock       int                               `json:"minimumStock"`
	InventoryStrategy  string                            `json:"inventoryStrategy"`
	ManufacturerURL    string                            `json:"manufacturerUrl,omitempty"`
	ProductURL         string                            `json:"productUrl,omitempty"`
	AlternativeNumbers []string                          `json:"alternativeNumbers"`
	Keywords           []string                          `json:"keywords"`
	CompatibilityNotes string                            `json:"compatibilityNotes,omitempty"`
	InternalNotes      string                            `json:"internalNotes,omitempty"`
	Archived           bool                              `json:"archived"`
	Stock              []TransferAccessoryStock          `json:"stock"`
	Assets             []TransferAccessoryAsset          `json:"assets"`
	CreatedAt          string                            `json:"createdAt,omitempty"`
	UpdatedAt          string                            `json:"updatedAt,omitempty"`
	FingerprintState   TransferAccessoryFingerprintState `json:"-"`
}

type TransferAccessoryFingerprintState struct {
	Reservations     []TransferAccessoryReservationFingerprint      `json:"reservations"`
	Installations    []TransferAccessoryInstallationFingerprint     `json:"installations"`
	ConditionHistory []TransferAccessoryConditionHistoryFingerprint `json:"conditionHistory"`
}

type TransferAccessoryReservationFingerprint struct {
	ID         string `json:"id"`
	AssetID    string `json:"assetId,omitempty"`
	LocationID string `json:"locationId"`
	Quantity   int    `json:"quantity"`
	Status     string `json:"status"`
	UpdatedAt  string `json:"updatedAt"`
}

type TransferAccessoryInstallationFingerprint struct {
	ID               string `json:"id"`
	AssetID          string `json:"assetId,omitempty"`
	SourceLocationID string `json:"sourceLocationId"`
	Quantity         int    `json:"quantity"`
	Condition        string `json:"condition"`
	InstalledAt      string `json:"installedAt"`
	RemovedAt        string `json:"removedAt,omitempty"`
}

type TransferAccessoryConditionHistoryFingerprint struct {
	ID             string `json:"id"`
	InstallationID string `json:"installationId"`
	Condition      string `json:"condition"`
	ChangedAt      string `json:"changedAt"`
}

type TransferAccessoryStock struct {
	LocationID   string `json:"locationId"`
	LocationName string `json:"locationName"`
	Quantity     int    `json:"quantity"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}

type TransferAccessoryAsset struct {
	ID                  string `json:"id,omitempty"`
	InventoryNumber     string `json:"inventoryNumber,omitempty"`
	SerialNumber        string `json:"serialNumber,omitempty"`
	Condition           string `json:"condition"`
	Lifecycle           string `json:"lifecycle"`
	StorageLocationID   string `json:"storageLocationId,omitempty"`
	StorageLocationName string `json:"storageLocationName,omitempty"`
	PurchaseDate        string `json:"purchaseDate,omitempty"`
	PurchasePrice       string `json:"purchasePrice,omitempty"`
	WarrantyUntil       string `json:"warrantyUntil,omitempty"`
	Notes               string `json:"notes,omitempty"`
	CreatedAt           string `json:"createdAt,omitempty"`
	UpdatedAt           string `json:"updatedAt,omitempty"`
}
