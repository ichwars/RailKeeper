package api

import (
	"net/http"

	"railkeeper/backend/internal/application"
)

func cvDeletionComparison(r *http.Request) (application.VehicleCVDeletionInput, bool) {
	query := r.URL.Query()
	input := application.VehicleCVDeletionInput{WorkspaceID: query.Get("workspaceId"), ExpectedWorkspaceRevision: query.Get("expectedWorkspaceRevision")}
	for _, field := range []string{"expectedRevision", "workspaceId", "expectedWorkspaceRevision"} {
		if len(query[field]) > 1 {
			return input, false
		}
	}
	if query.Has("expectedRevision") {
		value := query.Get("expectedRevision")
		input.ExpectedRevision = &value
	}
	if query.Has("workspaceId") || query.Has("expectedWorkspaceRevision") {
		if input.WorkspaceID == "" || input.ExpectedWorkspaceRevision == "" || input.ExpectedRevision == nil {
			return input, false
		}
	}
	return input, true
}
