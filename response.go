package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
)

const (
	// ErrCannotDecodeJSON is returned when a JSON response cannot be decoded.
	ErrCannotDecodeJSON xerrors.Error = "cannot decode JSON response"

	// ErrCannotEncodeJSON is returned when a JSON request cannot be encoded.
	ErrCannotEncodeJSON xerrors.Error = "cannot encode JSON request"

	// ErrCannotDrainResponse is returned when a response body cannot be drained.
	ErrCannotDrainResponse xerrors.Error = "cannot drain response body"

	// ErrCannotCloseResponse is returned when a response body cannot be closed.
	ErrCannotCloseResponse xerrors.Error = "cannot close response body"
)

// ReadJSON reads the body of an HTTP response and unmarshals it into the given
// struct. The provided val parameter should be a pointer to a struct where the
// JSON data will be unmarshalled.
func ReadJSON(resp *http.Response, val any) error {
	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(val); err != nil {
		return fmt.Errorf("%w: %w", ErrCannotDecodeJSON, err)
	}

	return nil
}

// WriteJSON writes a given struct to a JSON payload that can be used for HTTP
// requests. The provided val parameter should be a pointer to a struct where
// the JSON data will be marshaled.
func WriteJSON(val any) (*bytes.Buffer, error) {
	var (
		payload *bytes.Buffer
		encoder = json.NewEncoder(payload)
	)

	if err := encoder.Encode(val); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotEncodeJSON, err)
	}

	return payload, nil
}

// DrainResponseBody drains the response body until EOF and closes it. It
// returns an error if the drain or close fails.
//
// DrainResponseBody reads and discards the remaining content of the response
// body until EOF, then closes it. If an error occurs while draining or closing
// the response body, an error is returned.
func DrainResponseBody(resp *http.Response) error {
	_, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCannotDrainResponse, err)
	}

	if err = resp.Body.Close(); err != nil {
		return fmt.Errorf("%w: %w", ErrCannotCloseResponse, err)
	}

	return nil
}

// IsSuccess checks if the HTTP response has a successful status code (2xx).
func IsSuccess(resp *http.Response) bool {
	return resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
}
