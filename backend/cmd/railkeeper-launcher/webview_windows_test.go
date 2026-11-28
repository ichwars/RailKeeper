package main

import (
	"net/url"
	"testing"
)

func TestDesktopOriginUsesActualLoopbackEndpoint(t *testing.T) {
	tests := []struct {
		address, navigation string
		allowed             bool
	}{
		{"http://127.0.0.1:8081", "http://127.0.0.1:8081/vehicles", true},
		{"http://127.0.0.1:8081", "http://127.0.0.1:8080/", false},
		{"http://127.0.0.1:8081", "http://127.0.0.1.example.com:8081/", false},
		{"http://127.0.0.1:80", "http://127.0.0.1/", true},
		{"http://[::1]:80", "http://[::1]/", true},
		{"http://127.0.0.1:80", "https://127.0.0.1/", false},
		{"http://127.0.0.1:80", "http://untrusted@127.0.0.1/", false},
	}
	for _, tt := range tests {
		t.Run(tt.navigation, func(t *testing.T) {
			window := desktopWindow{address: tt.address}
			target, err := url.Parse(tt.navigation)
			if err != nil {
				t.Fatal(err)
			}
			if allowed := window.sameOrigin(target); allowed != tt.allowed {
				t.Errorf("origin %s allowed navigation to %s = %t, want %t", tt.address, target, allowed, tt.allowed)
			}
		})
	}
}
