package ecos

import (
	"context"
	"net/http"
	"time"
)

// ReadBackup uses the same resolved private address as the database reader.
// Saving a station configuration is an HTTP GET, never a device restore.
func (reader *CVDatabaseReader) ReadBackup(ctx context.Context) ([]byte, error) {
	web := *reader.web
	client := *reader.web.client
	if transport, ok := client.Transport.(*http.Transport); ok {
		copyTransport := transport.Clone()
		copyTransport.ResponseHeaderTimeout = 35 * time.Second
		defer copyTransport.CloseIdleConnections()
		client.Transport = copyTransport
	}
	client.Timeout = 40 * time.Second
	web.client = &client
	return web.request(ctx, http.MethodGet, "/cgi/save?lang=en", "", nil, EncryptedBackupLimit)
}
