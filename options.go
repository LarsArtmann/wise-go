package wise

import (
	"net/http"
	"time"
)

// config holds the internal client configuration.
type config struct {
	apiKey     string
	baseURL    string
	timeout    time.Duration
	maxRetries int
	retryMin   time.Duration
	retryMax   time.Duration
	httpClient *http.Client
	now        func() time.Time
}

func defaultConfig() config {
	return config{}
}

// Option configures a Client.
type Option func(*config)

// WithSandbox sets the client to use the Wise sandbox environment.
func WithSandbox() Option {
	return func(c *config) {
		c.baseURL = SandboxURL
	}
}

// WithBaseURL sets a custom base URL (overrides production/sandbox).
func WithBaseURL(url string) Option {
	return func(c *config) {
		c.baseURL = url
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.timeout = timeout
	}
}

// WithRetry configures the retry policy.
// maxRetries is the maximum number of retry attempts (not including the initial request).
func WithRetry(maxRetries int, minDelay, maxDelay time.Duration) Option {
	return func(c *config) {
		c.maxRetries = maxRetries
		c.retryMin = minDelay
		c.retryMax = maxDelay
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		c.httpClient = client
	}
}

// withNow sets the clock function for testing.
func withNow(fn func() time.Time) Option {
	return func(c *config) {
		c.now = fn
	}
}
