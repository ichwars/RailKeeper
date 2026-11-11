package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"

	"railkeeper/backend/internal/application"
)

func (repository *DigitalCenterWorkspaceRepository) LatestReadySessions(
	ctx context.Context, centers []application.DigitalCenterSummary,
) (map[string]*application.DigitalCenterReadSession, error) {
	if len(centers) == 0 {
		return nil, nil
	}
	payload, err := json.Marshal(centers)
	if err != nil {
		return nil, err
	}
	rows, err := repository.db.QueryContext(ctx, `
WITH requested(provider, host, port) AS (
  SELECT value->>'provider', value->>'host', value->>'port' FROM json_each(?)
), latest(id) AS (
  SELECT (
    SELECT session.id
    FROM digital_center_read_sessions session
    WHERE session.provider=requested.provider AND session.host=requested.host AND session.port=requested.port
      AND session.state='ready' AND session.read_completed_at IS NOT NULL
      AND json_extract(session.capability_json, '$.readLocomotives')=1
    ORDER BY session.read_completed_at DESC, session.rowid DESC LIMIT 1
  )
  FROM requested
)
`+digitalCenterSessionSelect+` WHERE id IN (SELECT id FROM latest)`, string(payload))
	if err != nil {
		return nil, fmt.Errorf("load latest completed digital center reads: %w", err)
	}
	defer func() { _ = rows.Close() }()
	sessions := make(map[string]*application.DigitalCenterReadSession, len(centers))
	for rows.Next() {
		session, err := scanDigitalCenterReadSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan latest completed digital center read: %w", err)
		}
		sessions[session.Provider] = &session
	}
	return sessions, rows.Err()
}
