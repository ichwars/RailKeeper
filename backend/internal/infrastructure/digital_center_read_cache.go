package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/application"
)

func (repository *DigitalCenterWorkspaceRepository) LatestReadySession(
	ctx context.Context, provider, host string, port int,
) (*application.DigitalCenterReadSession, error) {
	session, err := scanDigitalCenterReadSession(repository.db.QueryRowContext(ctx,
		digitalCenterSessionSelect+`
WHERE provider=? AND host=? AND port=? AND state='ready'
  AND read_completed_at IS NOT NULL
  AND json_extract(capability_json, '$.readLocomotives')=1
ORDER BY read_completed_at DESC, rowid DESC LIMIT 1`, provider, host, port))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load latest completed digital center read: %w", err)
	}
	return &session, nil
}
