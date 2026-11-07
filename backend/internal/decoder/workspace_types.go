package decoder

type DecoderCVWorkspaceInput struct {
	ID               string `json:"id"`
	ExpectedRevision string `json:"expectedRevision"`
	Profile          string `json:"profile"`
}

type DecoderCVCheckpointInput struct {
	ID               string `json:"id"`
	ExpectedRevision string `json:"expectedRevision"`
	Note             string `json:"note"`
}

type DecoderCVAssignmentInput struct {
	ExpectedRevision string  `json:"expectedRevision"`
	InstallationID   *string `json:"installationId"`
	SnapshotID       string  `json:"snapshotId"`
}
