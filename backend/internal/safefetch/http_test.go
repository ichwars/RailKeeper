package safefetch

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

var errDialStopped = errors.New("dial stopped by test")

type staticResolver struct {
	addrs []net.IPAddr
	err   error
}

func (r staticResolver) LookupIPAddr(context.Context, string) ([]net.IPAddr, error) {
	return r.addrs, r.err
}

type captureDialer struct {
	network string
	address string
}

func (d *captureDialer) DialContext(_ context.Context, network, address string) (net.Conn, error) {
	d.network = network
	d.address = address
	return nil, errDialStopped
}

func TestHTTPClientRejectsHostnameResolvingToPrivateAddress(t *testing.T) {
	dialer := &captureDialer{}
	client := NewHTTPClient(t.Context(), Options{
		Timeout:  time.Second,
		Resolver: staticResolver{addrs: []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}},
		Dialer:   dialer,
	})

	resp, err := client.Get("http://example.test/manual.pdf")
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	if !errors.Is(err, ErrPrivateAddress) {
		t.Fatalf("expected private address error, got %v", err)
	}
	if dialer.address != "" {
		t.Fatalf("private address should not be dialed, got %s", dialer.address)
	}
}

func TestHTTPClientDialsValidatedResolvedIP(t *testing.T) {
	dialer := &captureDialer{}
	client := NewHTTPClient(t.Context(), Options{
		Timeout:  time.Second,
		Resolver: staticResolver{addrs: []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}}},
		Dialer:   dialer,
	})

	resp, err := client.Get("http://example.test/manual.pdf")
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	if !errors.Is(err, errDialStopped) {
		t.Fatalf("expected test dialer error, got %v", err)
	}
	if dialer.network != "tcp" || dialer.address != "93.184.216.34:80" {
		t.Fatalf("expected validated IP dial, got network=%q address=%q", dialer.network, dialer.address)
	}
}
