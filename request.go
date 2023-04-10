package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"git.sr.ht/~jamesponddotco/xstd-go/xcrypto/xuuid"
	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
)

const (
	// ErrRequest is returned when a request cannot be created.
	ErrRequest xerrors.Error = "unable to create request"

	// ErrIdempotencyKey is returned when an idempotency key is not provided and a
	// random one cannot be generated.
	ErrIdempotencyKey xerrors.Error = "unable to generate idempotency key"
)

// Request represents an HTTP request to be sent by a client. It wraps an
// http.Request and provides additional methods for setting common headers.
type Request struct {
	// Req is the underlying http.Request.
	Req *http.Request
}

// NewRequest returns a new Request given a method, URL, and optional headers
// and body.
func NewRequest(ctx context.Context, method, url string, headers map[string]string, body io.Reader) (*Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequest, err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return &Request{Req: req}, nil
}

// SetBearerToken sets the Authorization header to use the given bearer token.
func (r *Request) SetBearerToken(token string) {
	r.Req.Header.Set("Authorization", "Bearer "+token)
}

// SetIdempotencyKey sets the Idempotency-Key header for POST and PATCH
// requests with the given key. If no key is provided, a random one is
// generated using a V4 UUID.
func (r *Request) SetIdempotencyKey(key string) error {
	if strings.TrimSpace(key) == "" {
		uuid, err := xuuid.New()
		if err != nil {
			return fmt.Errorf("%w: %w", ErrIdempotencyKey, err)
		}

		key = uuid.String()
	}

	if r.Req.Method == "POST" || r.Req.Method == "PATCH" {
		r.Req.Header.Set("Idempotency-Key", key)
	}

	return nil
}

// SetUserAgent sets the User-Agent header to the given value.
func (r *Request) SetUserAgent(ua string) {
	r.Req.Header.Set("User-Agent", ua)
}
