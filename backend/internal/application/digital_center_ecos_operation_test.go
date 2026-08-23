package application

import (
	"bufio"
	"context"
	"net"
	"testing"
	"time"
)

func TestECoSPauseLiveWaitsForSessionShutdown(t *testing.T) {
	service, closed := runningLiveService(t)
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	status, err := service.PauseLive(ctx)
	if err != nil || status.State != ECoSLiveStopped {
		t.Fatalf("status=%#v err=%v", status, err)
	}
	select {
	case <-service.liveDone:
	default:
		t.Fatal("PauseLive returned before the live session completed")
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("server did not observe the closed live connection")
	}
}

func TestDigitalCenterECoSOperationsAreSerialized(t *testing.T) {
	service := &DigitalCenterWorkspaceService{}
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- service.runECoSOperation(t.Context(), func() error {
			close(firstEntered)
			<-releaseFirst
			return nil
		})
	}()
	<-firstEntered

	secondEntered := make(chan struct{})
	secondDone := make(chan error, 1)
	go func() {
		secondDone <- service.runECoSOperation(t.Context(), func() error {
			close(secondEntered)
			return nil
		})
	}()
	select {
	case <-secondEntered:
		t.Fatal("second operation entered before first completed")
	case <-time.After(25 * time.Millisecond):
	}
	close(releaseFirst)
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
	if err := <-secondDone; err != nil {
		t.Fatal(err)
	}
}

func runningLiveService(t *testing.T) (*ECoSService, <-chan struct{}) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closed := make(chan struct{})
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			close(closed)
			return
		}
		defer close(closed)
		defer func() { _ = conn.Close() }()
		reader := bufio.NewReader(conn)
		for {
			if _, readErr := reader.ReadString('\n'); readErr != nil {
				return
			}
		}
	}()
	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	if _, err := service.StartLive(t.Context(), ECoSConnectionInput{Host: host, Port: port}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		service.StopLive()
		_ = listener.Close()
	})
	return service, closed
}
