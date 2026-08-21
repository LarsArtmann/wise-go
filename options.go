package wise

import (
	"context"
	"time"
)

// config holds the internal client configuration.
type config struct {
	apiKey           string
	baseURL          string
	timeout          time.Duration
	maxRetries       int
	retryMin         time.Duration
	retryMax         time.Duration
	httpClient       Doer
	correlationID    string
	scaApprovalToken string
	logger           Logger
}

func defaultConfig() config {
	return config{}
}

// Option configures a Client.
type Option func(*config)

// RequestLog describes one completed HTTP exchange with the Wise API.
type RequestLog struct {
	Method   string
	URL      string
	Status   int // 0 when the attempt failed at the transport layer.
	Duration time.Duration
	Attempt  int   // 1-based; greater than 1 means this was a retry.
	Error    error // Transport error, if any; API errors are not logged here.
}

// Logger is notified about every HTTP attempt against the Wise API,
// including retries. Implementations must be safe for concurrent use and
// must not block.
type Logger interface {
	LogRequest(entry RequestLog)
}

// RequestLogFunc adapts a plain function to Logger.
type RequestLogFunc func(entry RequestLog)

// LogRequest implements Logger.
func (f RequestLogFunc) LogRequest(entry RequestLog) {
	f(entry)
}

// WithLogger installs a Logger that observes every HTTP attempt, including
// retries. Use it for operability (latency tracking, retry visibility)
// without wrapping the HTTP transport.
func WithLogger(logger Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}

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

// WithHTTPClient sets a custom HTTP client. Accepts any type implementing Doer
// (*http.Client satisfies this implicitly). Use this to inject a client with
// custom Transport (tracing, logging, mTLS), custom Timeout, or a mock for testing.
func WithHTTPClient(client Doer) Option {
	return func(c *config) {
		c.httpClient = client
	}
}

// WithCorrelationID sets a static correlation ID sent as the
// X-External-Correlation-Id header on every request. This is Wise's
// documented global header for distributed tracing across the API.
// If empty, no header is sent.
//
// For per-request correlation IDs, derive a context with
// [WithRequestCorrelationID]; it takes precedence over this value.
func WithCorrelationID(id string) Option {
	return func(c *config) {
		c.correlationID = id
	}
}

// correlationIDContextKey types the context key for per-request correlation
// IDs; a private struct type prevents collisions with other packages' keys.
type correlationIDContextKey struct{}

// WithRequestCorrelationID returns a context whose requests carry the given
// correlation ID in the X-External-Correlation-Id header, overriding the
// client-wide WithCorrelationID value for requests made with this context.
// Use one unique ID per logical operation to trace it across Wise's systems.
func WithRequestCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDContextKey{}, id)
}

// correlationIDFromContext extracts a per-request correlation ID; empty when
// the context carries none.
func correlationIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(correlationIDContextKey{}).(string)

	return id
}

// WithSCAApprovalToken sends the given one-time token (OTT) as the
// x-2fa-approval header on every request. Configure it after receiving an
// [SCAChallengeError] and approving the challenge in the Wise app to clear
// Wise's Strong Customer Authentication (required once per ~90 days for
// SCA-protected endpoints such as balance statements on UK/EEA profiles).
// Once the challenge window is satisfied, remove the token again.
func WithSCAApprovalToken(token string) Option {
	return func(c *config) {
		c.scaApprovalToken = token
	}
}
