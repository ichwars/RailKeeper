package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"railkeeper/backend/internal/infrastructure"
)

func TestServerShutdownDrainsActiveDatabaseWrite(t *testing.T) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()

	dataDir := t.TempDir()
	database, err := infrastructure.OpenSQLite(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = database.Close() }()
	if _, err := database.Exec("CREATE TABLE shutdown_test (value TEXT NOT NULL)"); err != nil {
		t.Fatal(err)
	}

	started := make(chan struct{})
	finish, release := context.WithCancel(context.Background())
	defer release()
	writeResult := make(chan error, 1)
	serverStopped := make(chan error, 1)
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-finish.Done()
		_, err := database.ExecContext(r.Context(), "INSERT INTO shutdown_test VALUES ('committed')")
		writeResult <- err
		w.WriteHeader(http.StatusNoContent)
	})}
	go func() {
		err := serveUntilShutdown(ctx, server, listener)
		serverStopped <- errors.Join(err, database.Close())
	}()

	requestDone := make(chan error, 1)
	go func() {
		client := http.Client{Timeout: 5 * time.Second}
		response, err := client.Get("http://" + listener.Addr().String())
		if err == nil {
			_ = response.Body.Close()
		}
		requestDone <- err
	}()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("request did not reach handler")
	}
	stop()
	deadline := time.Now().Add(5 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", listener.Addr().String(), 100*time.Millisecond)
		if err != nil {
			break
		}
		_ = conn.Close()
		if time.Now().After(deadline) {
			t.Fatal("listener did not close after shutdown")
		}
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case err := <-serverStopped:
		t.Fatalf("process finished while a request was active: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	release()
	if err := waitForShutdownResult(t, writeResult); err != nil {
		t.Fatalf("write during shutdown: %v", err)
	}
	if err := waitForShutdownResult(t, requestDone); err != nil {
		t.Fatalf("request failed during shutdown: %v", err)
	}
	if err := waitForShutdownResult(t, serverStopped); err != nil {
		t.Fatalf("server shutdown: %v", err)
	}

	reopened, err := infrastructure.OpenSQLite(dataDir)
	if err != nil {
		t.Fatalf("reopen %s: %v", filepath.Join(dataDir, "railkeeper.db"), err)
	}
	defer func() { _ = reopened.Close() }()
	var value string
	if err := reopened.QueryRow("SELECT value FROM shutdown_test").Scan(&value); err != nil || value != "committed" {
		t.Fatalf("reopened database value = %q, error = %v", value, err)
	}
}

func waitForShutdownResult(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not complete")
		return nil
	}
}
