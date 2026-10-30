package application

import (
	"context"
	"errors"
	"net"

	ecospkg "railkeeper/backend/internal/ecos"
)

// One failed station HTTP transport must not consume a timeout per locomotive.
// This state belongs to a single sequential roster read; a new read retries HTTP.
type storedRosterWeb struct {
	storedECoSWeb
	transportError error
}

func (web *storedRosterWeb) Read(ctx context.Context, objectID int, name string) ([]ecospkg.DatabaseCV, error) {
	if web.transportError != nil {
		return nil, web.transportError
	}
	values, err := web.storedECoSWeb.Read(ctx, objectID, name)
	web.retainTransportFailure(err)
	return values, err
}

func (web *storedRosterWeb) ReadBackup(ctx context.Context) ([]byte, error) {
	if web.transportError != nil {
		return nil, web.transportError
	}
	data, err := web.storedECoSWeb.ReadBackup(ctx)
	web.retainTransportFailure(err)
	return data, err
}

func (web *storedRosterWeb) retainTransportFailure(err error) {
	var networkError net.Error
	if errors.As(err, &networkError) {
		web.transportError = err
	}
}
