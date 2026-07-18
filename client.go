package wise

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	errorfamily "github.com/larsartmann/go-error-family"
)

const (
	defaultTimeout           = 30 * time.Second
	defaultMaxRetries        = 3
	defaultRetryBackoffStart = 100 * time.Millisecond
	defaultRetryBackoffCap   = 5 * time.Second
)

// Doer is the interface for an HTTP client. *http.Client satisfies this.
// Inject a custom implementation via WithHTTPClient for testing or middleware
// (tracing, logging, retries at the transport layer).
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client is the Wise API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient Doer
	executor   failsafe.Executor[*http.Response]
}

// New creates a new Wise API client with the given API key and options.
func New(apiKey string, opts ...Option) *Client {
	cfg := defaultConfig()
	cfg.apiKey = apiKey

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.baseURL == "" {
		cfg.baseURL = ProductionURL
	}

	timeout := defaultTimeout
	if cfg.timeout > 0 {
		timeout = cfg.timeout
	}

	retryMax := defaultMaxRetries
	if cfg.maxRetries > 0 {
		retryMax = cfg.maxRetries
	}

	retryMin := defaultRetryBackoffStart
	if cfg.retryMin > 0 {
		retryMin = cfg.retryMin
	}

	retryMaxDelay := defaultRetryBackoffCap
	if cfg.retryMax > 0 {
		retryMaxDelay = cfg.retryMax
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	retry := retrypolicy.NewBuilder[*http.Response]().
		WithMaxRetries(retryMax).
		WithBackoff(retryMin, retryMaxDelay).
		HandleIf(isRetryable).
		Build()

	return &Client{
		apiKey:     apiKey,
		baseURL:    cfg.baseURL,
		executor:   failsafe.With(retry),
		httpClient: httpClient,
	}
}

// isRetryable determines if an HTTP response should be retried.
func isRetryable(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}

	// Retry on 429 (rate limit) and 5xx (server errors)
	return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
}

// Authenticate validates the API key by calling ListProfiles.
func (c *Client) Authenticate(ctx context.Context) error {
	_, err := c.ListProfiles(ctx)
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	return nil
}

// Health checks if the Wise API is reachable.
func (c *Client) Health(ctx context.Context) error {
	return c.Authenticate(ctx)
}

// --- HTTP helpers ---

func (c *Client) get(ctx context.Context, path string, target any) error {
	return c.getWithQuery(ctx, path, nil, target)
}

type responseCloser struct {
	resp *http.Response
}

func (rc *responseCloser) close() {
	if rc.resp != nil && rc.resp.Body != nil {
		_ = rc.resp.Body.Close()
	}
}

func (c *Client) getWithQuery(
	ctx context.Context,
	path string,
	query func() string,
	target any,
) error {
	fullURL := c.baseURL + path

	if query != nil {
		if q := query(); q != "" {
			fullURL += "?" + q
		}
	}

	resp, err := c.executor.WithContext(ctx).
		//nolint:contextcheck
		GetWithExecution(func(exec failsafe.Execution[*http.Response]) (*http.Response, error) {
			req, reqErr := http.NewRequestWithContext(exec.Context(), http.MethodGet, fullURL, nil)
			if reqErr != nil {
				return nil, fmt.Errorf("create request: %w", reqErr)
			}

			c.setAuth(req)

			return c.httpClient.Do(req)
		})
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	rc := &responseCloser{resp: resp}
	defer rc.close()

	if err := c.checkError(resp); err != nil {
		return err
	}

	if target != nil {
		if err := jsonDecode(resp, target); err != nil {
			return errorfamily.WrapCorruption(err, "wise.response.decode", "decode response")
		}
	}

	return nil
}

func (c *Client) setAuth(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

func (c *Client) checkError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := readBody(resp)

	return newAPIError(resp.StatusCode, body, parseRetryAfter(resp.Header.Get("Retry-After")))
}
