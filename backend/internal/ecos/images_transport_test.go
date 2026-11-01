package ecos

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
)

type imageFailureTransport struct {
	cause error
	body  bool
}

func (transport imageFailureTransport) RoundTrip(*http.Request) (*http.Response, error) {
	if !transport.body {
		return nil, transport.cause
	}
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{},
		Body: io.NopCloser(imageFailureBody{cause: transport.cause})}, nil
}

type imageFailureBody struct{ cause error }

func (body imageFailureBody) Read([]byte) (int, error) { return 0, body.cause }

func TestImageWebPreservesTransportFailuresForOptionalReadBudget(t *testing.T) {
	for _, body := range []bool{false, true} {
		cause := &net.OpError{Op: "read", Net: "tcp", Err: os.ErrDeadlineExceeded}
		web := &ImageWeb{base: "http://192.168.2.10", client: &http.Client{
			Transport: imageFailureTransport{cause: cause, body: body},
		}}
		_, err := web.request(context.Background(), http.MethodGet, "/test", "", nil, 1024)
		var networkError net.Error
		if !errors.Is(err, ErrImageWeb) || !errors.Is(err, cause) ||
			!errors.As(err, &networkError) || !networkError.Timeout() {
			t.Fatalf("HTTP failure lost its network cause (body=%v): %v", body, err)
		}
	}
}
