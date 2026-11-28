package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
)

func serveUntilShutdown(ctx context.Context, server *http.Server, listener net.Listener) error {
	drained := make(chan error, 1)
	stop := context.AfterFunc(ctx, func() {
		drained <- server.Shutdown(context.Background())
	})
	serveErr := server.Serve(listener)

	// Serve returns as soon as Shutdown closes the listener. The process must
	// stay alive until every active handler has returned.
	var shutdownErr error
	if stop() {
		shutdownErr = server.Shutdown(context.Background())
	} else {
		shutdownErr = <-drained
	}
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}
	return errors.Join(serveErr, shutdownErr)
}

type desktopControl struct {
	stop  *os.File
	ready *os.File
}

func (control *desktopControl) watch(cancel context.CancelFunc) {
	// The launcher requests shutdown by closing its end of the pipe.
	// This goroutine owns the reader and the OS also closes it on process exit.
	defer func() { _ = control.stop.Close() }()
	var notification [1]byte
	_, _ = control.stop.Read(notification[:])
	cancel()
}
