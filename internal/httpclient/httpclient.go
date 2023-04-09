// Package httpclient provides functions to control the underlying HTTP client
// for [the httpx package].
//
// [the httpx package]: https://godocs.io/git.sr.ht/~jamesponddotco/httpx-go
package httpclient

import (
	"net/http"
	"time"
)

// NewClient returns a new http.Client with the given timeout and transport.
func NewClient(timeout time.Duration, transport *http.Transport) *http.Client {
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
