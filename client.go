package wise

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-retry"
)

const (
	defaultTimeout           = 30 * time.Second
	defaultMaxRetries        = 3
	defaultRetryBackoffStart = 100 * time.Millisecond
	defaultRetryBackoffCap   = 5 * time.Second
)

// quarterlyAPIVersion is the quarterly versioned API surface
// (https://api.wise.com/2026Q3). Some endpoint families exist only there —
// the webhook subscription CRUD and the SCA one-time-token endpoints — while
// the rest of the SDK uses the legacy /v1../v4 paths. The OpenAPI spec's
// server URL is authoritative; when the quarter rolls over, update this one
// constant after verifying the paths still resolve.
const quarterlyAPIVersion = "2026Q3"

// Doer is the interface for an HTTP client. *http.Client satisfies this.
// Inject a custom implementation via WithHTTPClient for testing or middleware
// (tracing, logging, retries at the transport layer).
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is the Wise API client.
type Client struct {
	apiKey           string
	baseURL          string
	userAgent        string
	correlationID    string
	scaApprovalToken string
	httpClient       Doer
	logger           Logger
	retryConfig      retry.Config
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

	// go-retry counts total attempts; WithRetry's maxRetries counts retries
	// after the initial request, hence the +1. DelayFunc honors Wise's
	// Retry-After, capped at the configured inter-attempt ceiling so a server
	// hint can never exceed the caller's WithRetry budget (a 0 return falls
	// through to exponential backoff).
	retryCfg := retry.DefaultConfig()
	retryCfg.MaxAttempts = retryMax + 1
	retryCfg.InitialDelay = retryMin
	retryCfg.MaxDelay = retryMaxDelay
	retryCfg.IsRetryable = isRetryableError
	retryCfg.DelayFunc = func(attempt int, err error) time.Duration {
		rle, ok := errors.AsType[*RateLimitError](err)
		if !ok || rle.RetryAfter <= 0 {
			return 0
		}

		return min(rle.RetryAfter, retryMaxDelay)
	}

	return &Client{
		apiKey:           apiKey,
		baseURL:          cfg.baseURL,
		userAgent:        cfg.userAgent,
		correlationID:    cfg.correlationID,
		scaApprovalToken: cfg.scaApprovalToken,
		retryConfig:      retryCfg,
		httpClient:       httpClient,
		logger:           cfg.logger,
	}
}

// isRetryableError decides whether a failed attempt triggers a retry. It
// enumerates every type newAPIError produces: rate limits (429) and server
// errors (5xx) retry; auth, not-found, SCA-challenge, and other client
// errors are terminal. Untyped errors (network failures, request
// construction) retry, matching the transport-error handling this
// classification had before go-retry.
func isRetryableError(err error) bool {
	if _, ok := errors.AsType[*RateLimitError](err); ok {
		return true
	}

	if _, ok := errors.AsType[*ServerError](err); ok {
		return true
	}

	if _, ok := errors.AsType[*AuthError](err); ok {
		return false
	}

	if _, ok := errors.AsType[*NotFoundError](err); ok {
		return false
	}

	if _, ok := errors.AsType[*SCAChallengeError](err); ok {
		return false
	}

	if _, ok := errors.AsType[*APIError](err); ok {
		return false
	}

	return true
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
	return c.request(ctx, http.MethodGet, path, query, nil, target, nil)
}

// getWithQueryHeaders is getWithQuery plus extra request headers, for
// endpoints whose contract depends on negotiation headers
// (e.g. Accept-Minor-Version on account-requirements).
func (c *Client) getWithQueryHeaders(
	ctx context.Context,
	path string,
	query func() string,
	extraHeaders map[string]string,
	target any,
) error {
	return c.request(ctx, http.MethodGet, path, query, nil, target, extraHeaders)
}

func (c *Client) post(ctx context.Context, path string, body, target any) error {
	return c.request(ctx, http.MethodPost, path, nil, body, target, nil)
}

// postWithHeaders is post plus extra request headers, for endpoints whose
// contract depends on operation-specific headers (e.g. One-Time-Token on
// the SCA one-time-token endpoints).
func (c *Client) postWithHeaders(
	ctx context.Context,
	path string,
	body, target any,
	extraHeaders map[string]string,
) error {
	return c.request(ctx, http.MethodPost, path, nil, body, target, extraHeaders)
}

func (c *Client) put(ctx context.Context, path string, body, target any) error {
	return c.request(ctx, http.MethodPut, path, nil, body, target, nil)
}

func (c *Client) delete(ctx context.Context, path string) error {
	return c.request(ctx, http.MethodDelete, path, nil, nil, nil, nil)
}

// getRaw performs a GET and returns the response body without JSON decoding,
// for endpoints that serve files (statement.csv/.pdf/.xlsx and friends).
func (c *Client) getRaw(ctx context.Context, path string, query func() string) ([]byte, error) {
	fullURL := c.baseURL + path

	if query != nil {
		if q := query(); q != "" {
			fullURL += "?" + q
		}
	}

	resp, err := c.doRequest(ctx, http.MethodGet, fullURL, nil, nil)
	if err != nil {
		return nil, err
	}

	rc := &responseCloser{resp: resp}
	defer rc.close()

	if err := c.checkError(resp); err != nil {
		return nil, fmt.Errorf("request %s %s: %w", http.MethodGet, fullURL, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body from %s: %w", fullURL, err)
	}

	return body, nil
}

func (c *Client) request(
	ctx context.Context,
	method string,
	path string,
	query func() string,
	body, target any,
	extraHeaders map[string]string,
) error {
	fullURL := c.baseURL + path

	if query != nil {
		if q := query(); q != "" {
			fullURL += "?" + q
		}
	}

	resp, err := c.doRequest(ctx, method, fullURL, body, extraHeaders)
	if err != nil {
		return err
	}

	rc := &responseCloser{resp: resp}
	defer rc.close()

	if err := c.checkError(resp); err != nil {
		return fmt.Errorf("request %s %s: %w", method, fullURL, err)
	}

	if target != nil {
		if err := jsonDecode(resp, target); err != nil {
			return errorfamily.WrapCorruption(
				err,
				"wise.response.decode",
				"decode response from "+fullURL,
			)
		}
	}

	return nil
}

func (c *Client) doRequest(
	ctx context.Context,
	method string,
	fullURL string,
	body any,
	extraHeaders map[string]string,
) (*http.Response, error) {
	resp, err := retry.DoWithValue(ctx, c.retryConfig,
		func(ctx context.Context, attempt int) (*http.Response, error) {
			var bodyReader io.Reader

			if body != nil {
				b, marshalErr := json.Marshal(body)
				if marshalErr != nil {
					return nil, fmt.Errorf("encode request body for %s %s: %w", method, fullURL, marshalErr)
				}

				bodyReader = bytes.NewReader(b)
			}

			req, reqErr := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
			if reqErr != nil {
				return nil, fmt.Errorf("create request for %s %s: %w", method, fullURL, reqErr)
			}

			c.setHeaders(ctx, req)

			for name, value := range extraHeaders {
				req.Header.Set(name, value)
			}

			if body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			return c.executeWithLogging(method, fullURL, req, attempt)
		})
	if err != nil {
		// On exhaustion the chain wraps the final attempt's typed error
		// (RateLimitError, ServerError), so errors.AsType still reaches it —
		// no unwrap bridge needed.
		return nil, fmt.Errorf("execute %s %s: %w", method, fullURL, err)
	}

	return resp, nil
}

// executeWithLogging performs one HTTP attempt, classifies non-2xx responses
// into the typed errors that drive the retry decision, and reports the
// attempt to the configured Logger (if any) with method, URL, status,
// duration, and the 1-based attempt number.
func (c *Client) executeWithLogging(
	method string,
	fullURL string,
	req *http.Request,
	attempt int,
) (*http.Response, error) {
	start := time.Now()

	resp, err := c.httpClient.Do(req)

	if c.logger != nil {
		entry := RequestLog{
			Method:   method,
			URL:      fullURL,
			Duration: time.Since(start),
			Attempt:  attempt,
			Error:    err,
		}
		if resp != nil {
			entry.Status = resp.StatusCode
		}

		c.logger.LogRequest(entry)
	}

	if err != nil {
		return nil, fmt.Errorf("do %s %s: %w", method, fullURL, err)
	}

	apiErr := c.checkError(resp)
	if apiErr != nil {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}

		return nil, apiErr
	}

	return resp, nil
}

func (c *Client) setHeaders(ctx context.Context, req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Wise asks integrations to identify themselves; a custom User-Agent
	// replaces Go's default "Go-http-client/1.1".
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	// A per-request correlation ID from the context overrides the
	// client-wide value, so a logical operation keeps its own trace even
	// when the client is shared.
	correlationID := correlationIDFromContext(ctx)
	if correlationID == "" {
		correlationID = c.correlationID
	}

	if correlationID != "" {
		req.Header.Set("X-External-Correlation-Id", correlationID)
	}

	if c.scaApprovalToken != "" {
		req.Header.Set(HeaderTwoFAApproval, c.scaApprovalToken)
	}
}

func (c *Client) checkError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, readErr := readBody(resp)
	if readErr != nil && body == "" {
		// An unreadable body must not masquerade as an empty Wise response —
		// surface the read failure in the error's message/body instead.
		body = fmt.Sprintf("(response body could not be read: %v)", readErr)
	}

	return newAPIError(
		resp.StatusCode,
		body,
		resp.Header.Clone(),
		parseRetryAfter(resp.Header.Get("Retry-After")),
		resp.Header.Get("X-Rate-Limited-By"),
	)
}
