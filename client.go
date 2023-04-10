package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"git.sr.ht/~jamesponddotco/httpx-go/internal/build"
	"git.sr.ht/~jamesponddotco/httpx-go/internal/httpclient"
	"git.sr.ht/~jamesponddotco/pagecache-go"
	"git.sr.ht/~jamesponddotco/pagecache-go/memorycachex"
	"golang.org/x/time/rate"
)

// _mediaTypeFormURLEncoded is the default Content-Type for POST requests with
// form data.
const _mediaTypeFormURLEncoded string = "application/x-www-form-urlencoded"

// DefaultTimeout is the default timeout for all requests made by the client.
const DefaultTimeout = 10 * time.Second

type Client struct {
	// client is the underlying http.Client used to make requests.
	client *http.Client

	// RateLimiter specifies a client-side requests per second limit.
	//
	// Ultimately, most APIs enforce this limit on their side, but this is a
	// good way to be a good citizen.
	RateLimiter *rate.Limiter

	// RetryPolicy specifies the policy for retrying requests.
	RetryPolicy *RetryPolicy

	// UserAgent is the User-Agent header to use for all requests.
	UserAgent *UserAgent

	// Cache is an optional cache mechanism to store HTTP responses.
	Cache pagecache.Cache

	// Timeout is the timeout for all requests made by the client, overriding
	// the default value set in the underlying http.Client.
	Timeout time.Duration
}

func DefaultClient() *Client {
	return &Client{
		client:      httpclient.NewClient(DefaultTimeout, DefaultTransport()),
		RateLimiter: rate.NewLimiter(rate.Limit(2), 1),
		RetryPolicy: DefaultRetryPolicy(),
		UserAgent:   DefaultUserAgent(),
		Cache:       memorycachex.NewCache(pagecache.DefaultPolicy(), pagecache.DefaultCapacity),
	}
}

func (c *Client) Do(ctx context.Context, req *Request) (*http.Response, error) {
	c.initClient()
	c.setUserAgent(req)

	var (
		resp *http.Response
		key  string
		err  error
	)

	if c.Cache != nil {
		key = pagecache.Key(build.Name, req.Req)

		resp, err = c.Cache.Get(ctx, key)
		if resp != nil && err == nil {
			return resp, nil
		}
	}

	for i := 0; i < c.RetryPolicy.MaxRetries; i++ {
		if i > 0 && c.RateLimiter != nil {
			if err = c.RateLimiter.Wait(req.Req.Context()); err != nil {
				return nil, fmt.Errorf("%w", err)
			}
		}

		resp, err = c.client.Do(req.Req)
		if err != nil {
			select {
			case <-req.Req.Context().Done():
				return nil, fmt.Errorf("%w", req.Req.Context().Err())
			default:
			}

			if errors.Is(err, context.DeadlineExceeded) {
				return nil, fmt.Errorf("%w", err)
			}

			return nil, fmt.Errorf("%w", err)
		}

		if c.RetryPolicy.ShouldRetry(resp) {
			if err = c.RetryPolicy.Wait(req.Req.Context(), resp); err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			continue
		}
	}

	if c.Cache != nil {
		key = pagecache.Key(build.Name, req.Req)

		var (
			ctx    = context.Background()
			policy = c.Cache.Policy()
		)

		if err = c.Cache.Set(ctx, key, resp, policy.TTL(resp)); err != nil {
			return nil, fmt.Errorf("%w", err)
		}
	}

	return resp, nil
}

// Get is a convenience method for making GET requests.
func (c *Client) Get(ctx context.Context, uri string) (resp *http.Response, err error) {
	req, err := NewRequest(ctx, http.MethodGet, uri, nil, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return c.Do(ctx, req)
}

// Head is a convenience method for making HEAD requests.
func (c *Client) Head(ctx context.Context, uri string) (resp *http.Response, err error) {
	req, err := NewRequest(ctx, http.MethodHead, uri, nil, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return c.Do(ctx, req)
}

// Post is a convenience method for making POST requests.
func (c *Client) Post(ctx context.Context, uri, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := NewRequest(ctx, http.MethodPost, uri, nil, body)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	req.Req.Header.Set("Content-Type", contentType)

	return c.Do(ctx, req)
}

// PostForm is a convenience method for making POST requests with form data.
func (c *Client) PostForm(ctx context.Context, uri string, data url.Values) (resp *http.Response, err error) {
	return c.Post(ctx, uri, _mediaTypeFormURLEncoded, strings.NewReader(data.Encode()))
}

// initClient initializes the underlying http.Client if none has been set and
// set the timeout if it's not zero.
func (c *Client) initClient() {
	if c.client == nil {
		c.client = httpclient.NewClient(DefaultTimeout, DefaultTransport())
	}

	if c.client.Timeout == 0 && c.Timeout != 0 {
		c.client.Timeout = c.Timeout
	}
}

// setUserAgent sets the User-Agent header if it's not already set.
func (c *Client) setUserAgent(req *Request) {
	if req.Req.Header.Get("User-Agent") == "" {
		req.Req.Header.Set("User-Agent", c.UserAgent.String())
	}
}
