package application

import (
	"context"
	"fmt"
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
	for index := range centers {
		center := &centers[index]
		center.LastReadSession, err = service.repository.LatestReadySession(ctx, center.Provider, center.Host, center.Port)
		if err != nil {
			return nil, fmt.Errorf("load saved %s read session: %w", center.Provider, err)
		}
	}
	return centers, nil
}
