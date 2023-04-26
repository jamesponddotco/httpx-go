package httpx

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.sr.ht/~jamesponddotco/xstd-go/xcrypto/xrand"
	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
)

// _jitterFraction is the fraction of the jitter to use when calculating the
// jittered delay.
const _jitterFraction float64 = 0.25

// ErrRetryCanceled is returned when the request is canceled while waiting to retry.
const ErrRetryCanceled xerrors.Error = "retry canceled"

// RetryPolicy defines a policy for retrying HTTP requests.
type RetryPolicy struct {
	// retryTimer is a timer used to wait before retrying a request.
	retryTimer *time.Timer

	// retryableStatusCodeMap is a map of HTTP status codes that should trigger a retry.
	retryableStatusCodeMap map[int]bool

	// RetryableStatusCodes is a slice of HTTP status codes that should trigger
	// a retry.
	//
	// If a response's status code is in this slice, the request will be
	// retried according to the policy.
	RetryableStatusCodes []int

	// MaxRetries is the maximum number of times a request will be retried if
	// it encounters a retryable status code.
	MaxRetries int

	// MinRetryDelay is the minimum duration to wait before retrying a request.
	MinRetryDelay time.Duration

	// MaxRetryDelay is the maximum duration to wait before retrying a request.
	MaxRetryDelay time.Duration
}

// DefaultRetryPolicy returns a RetryPolicy with sensible defaults for retrying HTTP requests.
func DefaultRetryPolicy() *RetryPolicy {
	retryableStatusCodes := []int{
		http.StatusTooManyRequests,
		http.StatusServiceUnavailable,
		http.StatusBadGateway,
		http.StatusGatewayTimeout,
		http.StatusRequestTimeout,
		http.StatusConflict,
		http.StatusPreconditionFailed,
		http.StatusLocked,
	}

	retryableStatusCodeMap := make(map[int]bool, len(retryableStatusCodes))
	for _, code := range retryableStatusCodes {
		retryableStatusCodeMap[code] = true
	}

	return &RetryPolicy{
		retryTimer:             time.NewTimer(0),
		retryableStatusCodeMap: retryableStatusCodeMap,
		RetryableStatusCodes:   retryableStatusCodes,
		MaxRetries:             4,
		MinRetryDelay:          1 * time.Second,
		MaxRetryDelay:          30 * time.Second,
	}
}

// RetryAfter returns the amount of time to wait before retrying a request
// based on the "Retry-After" header.
//
// If the header is not present, the returned duration is MinRetryDelay with
// added jitter to prevent thundering herds.
func (p *RetryPolicy) RetryAfter(resp *http.Response) time.Duration {
	delay := p.MinRetryDelay

	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			delay = time.Duration(seconds) * time.Second
		}
	}

	jitteredDelay := delay + p.jitter(delay)

	switch {
	case jitteredDelay < p.MinRetryDelay:
		jitteredDelay = p.MinRetryDelay
	case jitteredDelay > p.MaxRetryDelay:
		jitteredDelay = p.MaxRetryDelay
	}

	return jitteredDelay
}

// ShouldRetry checks if the response's status code indicates that the request
// should be retried.
func (p *RetryPolicy) ShouldRetry(resp *http.Response) bool {
	return p.retryableStatusCodeMap[resp.StatusCode]
}

// Wait blocks until the specified request should be retried or the context is
// canceled.
//
// If the context is canceled, it returns an error.
func (p *RetryPolicy) Wait(ctx context.Context, resp *http.Response) error {
	delay := p.RetryAfter(resp)

	p.retryTimer.Reset(delay)

	select {
	case <-p.retryTimer.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("%w: %w", ErrRetryCanceled, ctx.Err())
	}
}

// jitter calculates a jittered duration based on the specified duration.
func (*RetryPolicy) jitter(delay time.Duration) time.Duration {
	var (
		jitterRange   = int64(float64(delay) * _jitterFraction)
		minJitter     = int64(delay) - jitterRange
		maxJitter     = int64(delay) + jitterRange
		jitteredDelay = minJitter + int64(xrand.IntChaChaCha(int(maxJitter-minJitter), nil))
	)

	return time.Duration(jitteredDelay)
}
