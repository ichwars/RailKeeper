package decoder

type VehicleCVChange struct {
	ID                        string               `json:"id,omitempty"`
	ExpectedRevision          string               `json:"expectedRevision"`
	Delete                    bool                 `json:"delete,omitempty"`
	Value                     *VehicleCVValueInput `json:"value,omitempty"`
	WorkspaceID               string               `json:"workspaceId,omitempty"`
	ExpectedWorkspaceRevision string               `json:"expectedWorkspaceRevision,omitempty"`
}

type VehicleCVChangesInput struct {
	Changes []VehicleCVChange `json:"changes"`
}
