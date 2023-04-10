package httpx_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"git.sr.ht/~jamesponddotco/httpx-go"
)

func TestNewRequest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		method  string
		url     string
		headers map[string]string
		body    io.Reader
		want    *http.Request
		err     error
	}{
		{
			name:   "valid request",
			method: "GET",
			url:    "https://example.com",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			want: &http.Request{
				Method: "GET",
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
				},
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			err: nil,
		},
		{
			name:    "invalid method",
			method:  "invalid_method",
			url:     "https://example.com",
			headers: nil,
			body:    nil,
			want:    nil,
			err:     httpx.ErrRequest,
		},
		{
			name:    "invalid url",
			method:  "GET",
			url:     "::",
			headers: nil,
			body:    nil,
			want:    nil,
			err:     httpx.ErrRequest,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			got, err := httpx.NewRequest(ctx, tt.method, tt.url, tt.headers, tt.body)

			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("got error %v, want %v", err, tt.err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reqEqual(t, got.Req, tt.want) {
				t.Errorf("got request %+v, want %+v", got.Req, tt.want)
			}
		})
	}
}

func TestSetBearerToken(t *testing.T) {
	t.Parallel()

	req := &httpx.Request{Req: &http.Request{Header: http.Header{}}}
	token := "test-token"

	req.SetBearerToken(token)

	got := req.Req.Header.Get("Authorization")
	want := "Bearer " + token

	if got != want {
		t.Errorf("got Authorization header %q, want %q", got, want)
	}
}

func TestSetIdempotencyKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		method string
		key    string
		err    error
	}{
		{
			name:   "valid POST",
			method: "POST",
			key:    "",
			err:    nil,
		},
		{
			name:   "valid PATCH",
			method: "PATCH",
			key:    "",
			err:    nil,
		},
		{
			name:   "valid non-POST non-PATCH",
			method: "GET",
			key:    "",
			err:    nil,
		},
		{
			name:   "valid custom key",
			method: "POST",
			key:    "custom-idempotency-key",
			err:    nil,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := &httpx.Request{Req: &http.Request{Header: http.Header{}, Method: tt.method}}
			err := req.SetIdempotencyKey(tt.key)

			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("got error %v, want %v", err, tt.err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.method == "POST" || tt.method == "PATCH" {
				got := req.Req.Header.Get("Idempotency-Key")
				if tt.key == "" {
					if got == "" {
						t.Error("expected non-empty Idempotency-Key header, got empty")
					}
				} else {
					if got != tt.key {
						t.Errorf("got Idempotency-Key header %q, want %q", got, tt.key)
					}
				}
			} else {
				if req.Req.Header.Get("Idempotency-Key") != "" {
					t.Error("expected empty Idempotency-Key header for non-POST and non-PATCH methods")
				}
			}
		})
	}
}

func reqEqual(t *testing.T, a, b *http.Request) bool {
	t.Helper()

	if a == b {
		return true
	}

	if a.Method != b.Method {
		return false
	}

	if a.URL.String() != b.URL.String() {
		return false
	}

	if !headersEqual(t, a.Header, b.Header) {
		return false
	}

	return true
}

func headersEqual(t *testing.T, a, b http.Header) bool {
	t.Helper()

	if len(a) != len(b) {
		return false
	}

	for key, vals := range a {
		bVals, ok := b[key]
		if !ok || len(vals) != len(bVals) {
			return false
		}

		for i := range vals {
			if vals[i] != bVals[i] {
				return false
			}
		}
	}

	return true
}
