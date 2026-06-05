package main

import (
	"net"
	"strings"
	"testing"
)

func TestBrowserURLUsesLocalhostForWildcardAddress(t *testing.T) {
	got := browserURL("[::]:8080")
	if got != "http://127.0.0.1:8080" {
		t.Fatalf("browserURL() = %q, want localhost URL", got)
	}
}

func TestListenFallsBackToNextPortablePort(t *testing.T) {
	blocker, err := net.Listen("tcp", "127.0.0.1:18080")
	if err != nil {
		t.Skipf("test port unavailable: %v", err)
	}
	defer func() { _ = blocker.Close() }()

	listener, appURL, err := listen("127.0.0.1:18080", true)
	if err != nil {
		t.Fatalf("listen() returned error: %v", err)
	}
	defer func() { _ = listener.Close() }()

	if strings.Contains(appURL, ":18080") {
		t.Fatalf("listen() did not fall back, appURL = %q", appURL)
	}
	if !strings.HasPrefix(appURL, "http://127.0.0.1:") {
		t.Fatalf("listen() appURL = %q, want local URL", appURL)
	}
}
