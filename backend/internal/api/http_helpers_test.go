package api

import (
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestClientIPOnlyTrustsForwardedForFromConfiguredProxy(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}

	trustedRequest := httptest.NewRequest("GET", "/", nil)
	trustedRequest.RemoteAddr = "10.0.0.5:4242"
	trustedRequest.Header.Set("X-Forwarded-For", "198.51.100.20, 10.0.0.5")
	if got := clientIP(trustedRequest, trusted); got != "198.51.100.20" {
		t.Fatalf("trusted proxy clientIP() = %q", got)
	}

	untrustedRequest := httptest.NewRequest("GET", "/", nil)
	untrustedRequest.RemoteAddr = "203.0.113.10:4242"
	untrustedRequest.Header.Set("X-Forwarded-For", "198.51.100.20")
	if got := clientIP(untrustedRequest, trusted); got != "203.0.113.10" {
		t.Fatalf("untrusted proxy clientIP() = %q", got)
	}
}

func TestParseTrustedProxyPrefixesRejectsInvalidCIDR(t *testing.T) {
	if _, err := parseTrustedProxyPrefixes([]string{"not-a-cidr"}); err == nil {
		t.Fatal("expected invalid trusted proxy CIDR to be rejected")
	}
}
