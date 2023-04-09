package httpx

import (
	"fmt"
	"net/url"
	"time"
)

// Error is a generic HTTP error type that can be used to return more
// information about a request failure.
type Error struct {
	// URL is the URL that was requested.
	URL *url.URL

	// Method is the HTTP method used.
	Method string

	// Message is a human-readable error message.
	Message string

	// StatusText is the HTTP status text.
	StatusText string

	// StatusCode is the HTTP status code.
	StatusCode int
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s %d %s (%s): %s", e.Method, e.StatusCode, e.StatusText, e.URL, e.Message)
}

// RateLimitExceededError represents an error that occurs when the rate limit
// for an HTTP request has been exceeded.
type RateLimitExceededError struct {
	// ResetTime is the time at which the rate limit will reset.
	ResetTime time.Time

	// Remaining is the number of requests remaining in the current time window.
	Remaining int

	// Limit is the maximum number of requests allowed within a specific time
	// window.
	Limit int
}

// Error returns a human-readable error message describing the rate limit
// exceeded error. It implements the error interface.
func (e *RateLimitExceededError) Error() string {
	return fmt.Sprintf("rate limit exceeded: limit %d, remaining %d, reset time %s", e.Limit, e.Remaining, e.ResetTime.Format(time.RFC1123))
}

// RetryAfterExceededError represents an error that occurs when the maximum
// number of retries for an HTTP request has been exceeded.
type RetryAfterExceededError struct {
	// RetryAfter is the duration specified in the Retry-After header, indicating
	// the time clients should wait before sending another request.
	RetryAfter time.Duration

	// MaxRetries is the maximum number of retries allowed for an HTTP request.
	MaxRetries int
}

// Error returns a human-readable error message describing the retry limit
// exceeded error. It implements the error interface.
func (e *RetryAfterExceededError) Error() string {
	return fmt.Sprintf("retry limit exceeded: max retries %d, retry after %s", e.MaxRetries, e.RetryAfter)
}
