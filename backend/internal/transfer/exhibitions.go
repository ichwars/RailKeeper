package transfer

type TransferExhibitionList struct {
	ID                string                    `json:"id,omitempty"`
	Designation       string                    `json:"designation"`
	Date              string                    `json:"date"`
	EndDate           string                    `json:"endDate,omitempty"`
	Location          string                    `json:"location,omitempty"`
	Description       string                    `json:"description,omitempty"`
	OrganizationNotes string                    `json:"organizationNotes,omitempty"`
	Status            ExhibitionStatus          `json:"status,omitempty"`
	Locked            bool                      `json:"locked"`
	LockReason        string                    `json:"lockReason,omitempty"`
	LockedAt          string                    `json:"lockedAt,omitempty"`
	CompletedAt       string                    `json:"completedAt,omitempty"`
	ArchivedAt        string                    `json:"archivedAt,omitempty"`
	Entries           []TransferExhibitionEntry `json:"entries"`
	CreatedAt         string                    `json:"createdAt,omitempty"`
	UpdatedAt         string                    `json:"updatedAt,omitempty"`
}

type TransferExhibitionEntry struct {
	ID                     string `json:"id,omitempty"`
	VehicleID              string `json:"vehicleId,omitempty"`
	VehicleInventoryNumber string `json:"vehicleInventoryNumber,omitempty"`
	Owner                  string `json:"owner"`
	ImageURL               string `json:"imageUrl,omitempty"`
	LocomotiveName         string `json:"locomotiveName"`
	Gattung                string `json:"gattung,omitempty"`
	Series                 string `json:"series,omitempty"`
	Manufacturer           string `json:"manufacturer,omitempty"`
	Epoch                  string `json:"epoch,omitempty"`
	RailwayCompany         string `json:"railwayCompany,omitempty"`
	DayScope               string `json:"dayScope"`
	DTDecoder              bool   `json:"dtDecoder"`
	DecoderNumber          string `json:"decoderNumber,omitempty"`
	DecoderType            string `json:"decoderType,omitempty"`
	Adapter                string `json:"adapter,omitempty"`
	InterfaceName          string `json:"interfaceName,omitempty"`
	SXAddress              string `json:"sxAddress,omitempty"`
	Analog                 bool   `json:"analog"`
	Availability           string `json:"availability,omitempty"`
	FunctionKeys           string `json:"functionKeys,omitempty"`
	Notes                  string `json:"notes,omitempty"`
	SortOrder              int    `json:"sortOrder"`
	CreatedAt              string `json:"createdAt,omitempty"`
	UpdatedAt              string `json:"updatedAt,omitempty"`
}
