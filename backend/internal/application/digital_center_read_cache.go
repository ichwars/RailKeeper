package application

import (
	"context"
)

// ListWorkspaceCenters restores completed reads from local storage without contacting a station.
func (service *DigitalCenterWorkspaceService) ListWorkspaceCenters(ctx context.Context) ([]DigitalCenterSummary, error) {
	if service == nil || service.repository == nil {
		return nil, ErrDigitalCenterWorkspaceUnavailable
	}
	centers, err := service.ListConfiguredCenters(ctx)
	if err != nil {
		return nil, err
	}
	sessions, err := service.repository.LatestReadySessions(ctx, centers)
	if err != nil {
		return nil, err
	}
	for i := range centers {
		centers[i].LastReadSession = sessions[centers[i].Provider]
	}
	return centers, nil
}
