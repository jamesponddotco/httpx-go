package httpx_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.sr.ht/~jamesponddotco/httpx-go"
)

func TestRetryPolicy_RetryAfter(t *testing.T) {
	t.Parallel()

	const (
		minDelay    = 1 * time.Second
		maxDelay    = 30 * time.Second
		jitterRange = 0.25
	)

	policy := httpx.DefaultRetryPolicy()

	tests := []struct {
		name       string
		retryAfter string
		min        time.Duration
		max        time.Duration
	}{
		{
			name: "without Retry-After header",
			min:  minDelay,
			max:  maxDelay,
		},
		{
			name:       "with Retry-After header 3 seconds",
			retryAfter: "3",
			min:        3 * time.Second,
			max:        3 * time.Second * (1 + time.Duration(jitterRange*float64(time.Second))),
		},
		{
			name:       "with Retry-After header 10 seconds",
			retryAfter: "10",
			min:        10 * time.Second,
			max:        10 * time.Second * (1 + time.Duration(jitterRange*float64(time.Second))),
		},
		{
			name:       "with invalid Retry-After header",
			retryAfter: "invalid",
			min:        minDelay,
			max:        maxDelay,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := httptest.NewRecorder()
			if tt.retryAfter != "" {
				resp.Header().Set("Retry-After", tt.retryAfter)
			}

			actualResp := resp.Result()
			defer actualResp.Body.Close()

			delay := policy.RetryAfter(actualResp)

			if delay < tt.min || delay > tt.max {
				t.Errorf("RetryAfter() delay = %v, expected between %v and %v", delay, tt.min, tt.max)
			}
		})
	}
}

func TestRetryPolicy_ShouldRetry(t *testing.T) {
	t.Parallel()

	policy := httpx.DefaultRetryPolicy()

	tests := []struct {
		name     string
		status   int
		expected bool
	}{
		{
			name:     "retryable status code",
			status:   http.StatusTooManyRequests,
			expected: true,
		},
		{
			name:     "non-retryable status code",
			status:   http.StatusOK,
			expected: false,
		},
		{
			name:     "another retryable status code",
			status:   http.StatusGatewayTimeout,
			expected: true,
		},
		{
			name:     "another non-retryable status code",
			status:   http.StatusNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := httptest.NewRecorder()
			resp.WriteHeader(tt.status)

			actualResp := resp.Result()
			defer actualResp.Body.Close()

			if got := policy.ShouldRetry(actualResp); got != tt.expected {
				t.Errorf("ShouldRetry() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestRetryPolicy_Wait(t *testing.T) {
	t.Parallel()

	policy := httpx.DefaultRetryPolicy()

	tests := []struct {
		name          string
		retryAfter    string
		timeout       time.Duration
		cancelContext bool
		expectedErr   error
	}{
		{
			name:          "successful wait",
			retryAfter:    "1",
			timeout:       5 * time.Second,
			cancelContext: false,
			expectedErr:   nil,
		},
		{
			name:          "wait timeout",
			retryAfter:    "5",
			timeout:       1 * time.Second,
			cancelContext: false,
			expectedErr:   context.DeadlineExceeded,
		},
		{
			name:          "wait canceled by context",
			retryAfter:    "3",
			timeout:       0,
			cancelContext: true,
			expectedErr:   context.Canceled,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := httptest.NewRecorder()
			if tt.retryAfter != "" {
				resp.Header().Set("Retry-After", tt.retryAfter)
			}

			actualResp := resp.Result()
			defer actualResp.Body.Close()

			var (
				ctx    context.Context
				cancel context.CancelFunc
			)

			if tt.cancelContext {
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), tt.timeout)
			}
			defer cancel()

			err := policy.Wait(ctx, actualResp)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Wait() error = %v, expected %v", err, tt.expectedErr)
			}
		})
	}
}
