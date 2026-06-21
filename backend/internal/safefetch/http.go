package safefetch

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrInvalidURL     = errors.New("remote URL is invalid")
	ErrPrivateAddress = errors.New("remote URL resolves to a private address")
)

type Resolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type Options struct {
	Timeout      time.Duration
	MaxRedirects int
	Resolver     Resolver
	Dialer       Dialer
}

func NewHTTPClient(_ context.Context, options Options) *http.Client {
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	maxRedirects := options.MaxRedirects
	if maxRedirects <= 0 {
		maxRedirects = 5
	}
	resolver := options.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	dialer := options.Dialer
	if dialer == nil {
		dialer = &net.Dialer{Timeout: timeout}
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialPublicAddress(ctx, network, address, resolver, dialer)
	}

	return &http.Client{
		Timeout: timeout,
		Transport: publicURLTransport{
			base:     transport,
			resolver: resolver,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return http.ErrUseLastResponse
			}
			if err := ValidatePublicHTTPURL(req.Context(), req.URL.String(), resolver); err != nil {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func IsPublicHTTPURL(ctx context.Context, value string) bool {
	return ValidatePublicHTTPURL(ctx, value, net.DefaultResolver) == nil
}

func ValidatePublicHTTPURL(ctx context.Context, value string, resolver Resolver) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed == nil {
		return ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURL
	}
	host := parsed.Hostname()
	if host == "" {
		return ErrInvalidURL
	}
	_, err = publicIPsForHost(ctx, host, resolver)
	return err
}

type publicURLTransport struct {
	base     http.RoundTripper
	resolver Resolver
}

func (t publicURLTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil || req.URL == nil {
		return nil, ErrInvalidURL
	}
	if err := ValidatePublicHTTPURL(req.Context(), req.URL.String(), t.resolver); err != nil {
		return nil, err
	}
	return t.base.RoundTrip(req)
}

func dialPublicAddress(ctx context.Context, network, address string, resolver Resolver, dialer Dialer) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidURL, address)
	}
	ips, err := publicIPsForHost(ctx, host, resolver)
	if err != nil {
		return nil, err
	}
	var firstErr error
	for _, ip := range ips {
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return nil, fmt.Errorf("%w: %s", ErrPrivateAddress, host)
}

func publicIPsForHost(ctx context.Context, host string, resolver Resolver) ([]net.IP, error) {
	host = strings.TrimSpace(strings.Trim(host, "[]"))
	if host == "" {
		return nil, ErrInvalidURL
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return nil, fmt.Errorf("%w: %s", ErrPrivateAddress, host)
	}
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicIP(ip) {
			return nil, fmt.Errorf("%w: %s", ErrPrivateAddress, host)
		}
		return []net.IP{ip}, nil
	}
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	addresses, err := resolver.LookupIPAddr(lookupCtx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve remote host: %w", err)
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("resolve remote host: no addresses")
	}
	ips := make([]net.IP, 0, len(addresses))
	for _, address := range addresses {
		if !isPublicIP(address.IP) {
			return nil, fmt.Errorf("%w: %s", ErrPrivateAddress, host)
		}
		ips = append(ips, address.IP)
	}
	return ips, nil
}

func isPublicIP(ip net.IP) bool {
	return ip != nil &&
		!ip.IsLoopback() &&
		!ip.IsPrivate() &&
		!ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() &&
		!ip.IsMulticast() &&
		!ip.IsUnspecified()
}
