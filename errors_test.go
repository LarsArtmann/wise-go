package wise

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/failsafe-go/failsafe-go/retrypolicy"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{name: "empty falls back to default", value: "", want: defaultRetryAfter},
		{name: "zero seconds", value: "0", want: 0},
		{name: "positive seconds", value: "120", want: 120 * time.Second},
		{name: "garbage falls back to default", value: "abc", want: defaultRetryAfter},
		{name: "negative falls back to default", value: "-5", want: defaultRetryAfter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := parseRetryAfter(tt.value)
			if got != tt.want {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestParseRetryAfterHTTPDate(t *testing.T) {
	t.Parallel()

	future := time.Now().UTC().Add(60 * time.Second).Format(time.RFC1123)
	got := parseRetryAfter(future)

	low, high := 55*time.Second, 65*time.Second
	if got < low || got > high {
		t.Errorf("parseRetryAfter(http-date) = %v, want between %v and %v", got, low, high)
	}
}

func TestErrorClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantCode   string
		wantFamily errorfamily.Family
		wantRetry  bool
	}{
		{
			name: "rate limit", statusCode: http.StatusTooManyRequests,
			body:     `{"errors":[{"code":"RATE_LIMITED","message":"slow down"}]}`,
			wantCode: "wise.rate_limit", wantFamily: errorfamily.Transient, wantRetry: true,
		},
		{
			name: "auth error", statusCode: http.StatusUnauthorized,
			body:     `{"errors":[{"code":"UNAUTHORIZED","message":"bad key"}]}`,
			wantCode: "wise.auth", wantFamily: errorfamily.Rejection, wantRetry: false,
		},
		{
			name: "forbidden", statusCode: http.StatusForbidden,
			body:     `{"errors":[{"code":"FORBIDDEN","message":"no access"}]}`,
			wantCode: "wise.auth", wantFamily: errorfamily.Rejection, wantRetry: false,
		},
		{
			name: "not found", statusCode: http.StatusNotFound,
			body:     `{"errors":[{"code":"NOT_FOUND","message":"missing"}]}`,
			wantCode: "wise.not_found", wantFamily: errorfamily.Rejection, wantRetry: false,
		},
		{
			name: "server error", statusCode: http.StatusInternalServerError,
			body:     `{"errors":[{"code":"SERVER","message":"boom"}]}`,
			wantCode: "wise.server", wantFamily: errorfamily.Transient, wantRetry: true,
		},
		{
			name: "generic api error", statusCode: http.StatusBadRequest,
			body:     `{"errors":[{"code":"BAD","message":"nope"}]}`,
			wantCode: "wise.api_error", wantFamily: errorfamily.Rejection, wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertErrorClassification(t, tt.statusCode, tt.body, tt.wantCode, tt.wantFamily, tt.wantRetry)
		})
	}
}

func assertErrorClassification(
	t *testing.T,
	statusCode int,
	body, wantCode string,
	wantFamily errorfamily.Family,
	wantRetry bool,
) {
	t.Helper()

	err := newAPIError(statusCode, body, nil, time.Second, "")

	coder, ok := err.(interface{ ErrorCode() string })
	if !ok {
		t.Fatalf("error does not implement ErrorCode: %T", err)
	}

	if got := coder.ErrorCode(); got != wantCode {
		t.Errorf("ErrorCode() = %q, want %q", got, wantCode)
	}

	familier, ok := err.(interface{ ErrorFamily() errorfamily.Family })
	if !ok {
		t.Fatalf("error does not implement ErrorFamily: %T", err)
	}

	if got := familier.ErrorFamily(); got != wantFamily {
		t.Errorf("ErrorFamily() = %v, want %v", got, wantFamily)
	}

	if wantRetry {
		retryable, ok := err.(interface{ IsRetryable() bool })
		if !ok || !retryable.IsRetryable() {
			t.Errorf("expected IsRetryable() = true")
		}
	}
}

func TestNewAPIErrorRetryAfter(t *testing.T) {
	t.Parallel()

	err := newAPIError(http.StatusTooManyRequests, "{}", nil, 42*time.Second, "ip")

	rle, ok := errors.AsType[*RateLimitError](err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}

	if rle.RetryAfter != 42*time.Second {
		t.Errorf("RetryAfter = %v, want 42s", rle.RetryAfter)
	}

	if rle.RateLimitedBy != "ip" {
		t.Errorf("RateLimitedBy = %q, want %q", rle.RateLimitedBy, "ip")
	}
}

func TestCheckErrorCapturesRateLimitedBy(t *testing.T) {
	t.Parallel()

	client := &Client{}
	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"Retry-After":       []string{"2"},
			"X-Rate-Limited-By": []string{"ip"},
		},
		Body: http.NoBody,
	}

	err := client.checkError(resp)

	rle, ok := errors.AsType[*RateLimitError](err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}

	if rle.RateLimitedBy != "ip" {
		t.Errorf("RateLimitedBy = %q, want %q", rle.RateLimitedBy, "ip")
	}

	if rle.RetryAfter != 2*time.Second {
		t.Errorf("RetryAfter = %v, want 2s", rle.RetryAfter)
	}
}

func TestCheckErrorWithoutRateLimitedBy(t *testing.T) {
	t.Parallel()

	client := &Client{}
	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"Retry-After": []string{"1"},
		},
		Body: http.NoBody,
	}

	err := client.checkError(resp)

	rle, ok := errors.AsType[*RateLimitError](err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}

	if rle.RateLimitedBy != "" {
		t.Errorf("RateLimitedBy = %q, want empty", rle.RateLimitedBy)
	}

	ctx := rle.ErrorContext()
	if _, ok := ctx["rate_limited_by"]; ok {
		t.Errorf("ErrorContext should not contain rate_limited_by when empty")
	}
}

func TestCheckErrorClassifiesSCAChallenge(t *testing.T) {
	t.Parallel()

	client := &Client{}
	scaHeader := http.Header{}
	scaHeader.Set(HeaderTwoFAApproval, "bb676aeb-7c4d-4930-bb55-ab949fd3fd87")
	scaHeader.Set(HeaderTwoFAApprovalResult, "REJECTED")
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Header:     scaHeader,
		Body:       http.NoBody,
	}

	err := client.checkError(resp)

	sca, ok := errors.AsType[*SCAChallengeError](err)
	if !ok {
		t.Fatalf("expected *SCAChallengeError, got %T: %v", err, err)
	}

	if sca.TwoFAApprovalToken() != "bb676aeb-7c4d-4930-bb55-ab949fd3fd87" {
		t.Errorf("TwoFAApprovalToken() = %q, want the x-2fa-approval value", sca.TwoFAApprovalToken())
	}

	if sca.ErrorFamily() != errorfamily.Rejection {
		t.Errorf("ErrorFamily() = %v, want Rejection", sca.ErrorFamily())
	}

	if !strings.Contains(sca.Error(), "x-2fa-approval") {
		t.Errorf("Error() should surface the SCA header names, got: %s", sca.Error())
	}

	if !strings.Contains(sca.Error(), "bb676aeb-7c4d-4930-bb55-ab949fd3fd87") {
		t.Errorf("Error() should surface the one-time token, got: %s", sca.Error())
	}
}

func TestCheckErrorForbiddenWithoutSCAHeadersStaysAuthError(t *testing.T) {
	t.Parallel()

	client := &Client{}
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Header:     http.Header{},
		Body:       http.NoBody,
	}

	err := client.checkError(resp)

	if _, ok := errors.AsType[*SCAChallengeError](err); ok {
		t.Fatalf("plain 403 without 2FA headers must stay AuthError, got SCAChallengeError")
	}

	auth, ok := errors.AsType[*AuthError](err)
	if !ok {
		t.Fatalf("expected *AuthError, got %T", err)
	}

	if auth.Headers == nil {
		t.Errorf("Headers should be captured (non-nil) even when empty")
	}
}

func TestWithSCAApprovalTokenSendsHeader(t *testing.T) {
	t.Parallel()

	var gotHeader string

	doer := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotHeader = req.Header.Get(HeaderTwoFAApproval)

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`[]`)),
			Request:    req,
		}, nil
	})

	client := New("test-key", WithHTTPClient(doer), WithSCAApprovalToken("ott-123"))

	if err := client.Authenticate(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotHeader != "ott-123" {
		t.Errorf("x-2fa-approval header = %q, want %q", gotHeader, "ott-123")
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

// TestMapperParseErrorsAreCorruption guards the 2026-08-18 incident class:
// permanent response-shape failures must classify as Corruption so consumers
// fail fast instead of retrying with backoff. A blanket Transient wrap in a
// consumer shadows this classification, so keep these assertions exhaustive
// per mapper.
func TestMapperParseErrorsAreCorruption(t *testing.T) {
	t.Parallel()

	const (
		rawTypePersonal = "PERSONAL"
		rawCreatedAt    = "2020-05-27T10:27:22"
	)

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "profile unparseable created_at",
			call: func() error {
				_, err := mapProfile(raw.Profile{ID: 1, Type: rawTypePersonal, CreatedAt: "not-a-timestamp"})

				return err
			},
		},
		{
			name: "profile unknown type",
			call: func() error {
				_, err := mapProfile(raw.Profile{ID: 1, Type: "TRUST", CreatedAt: rawCreatedAt})

				return err
			},
		},
		{
			name: "amount invalid currency",
			call: func() error {
				_, err := toMoney(raw.BalanceAmount{Value: 100, Currency: "xx"})

				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.call()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if family := errorfamily.Classify(err); family != errorfamily.Corruption {
				t.Errorf("Classify() = %v, want Corruption (error: %v)", family, err)
			}
		})
	}
}

func TestMapQuoteParseErrorsAreCorruption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "unparseable created_time",
			call: func() error {
				wireQuote := raw.Quote{
					ID: "quote-1", SourceCurrency: "EUR", TargetCurrency: "USD", CreatedTime: "garbage",
				}
				_, err := mapQuote(wireQuote, ProfileID{})

				return err
			},
		},
		{
			name: "invalid source currency",
			call: func() error {
				wireQuote := raw.Quote{
					ID: "quote-1", SourceCurrency: "euros", TargetCurrency: "USD",
					CreatedTime: "2023-01-15T10:30:00Z", ExpirationTime: "2023-01-15T11:00:00Z",
				}
				_, err := mapQuote(wireQuote, ProfileID{})

				return err
			},
		},
		{
			name: "unparseable payment option delivery",
			call: func() error {
				wireQuote := raw.Quote{
					ID: "quote-1", SourceCurrency: "EUR", TargetCurrency: "USD",
					CreatedTime: "2023-01-15T10:30:00Z", ExpirationTime: "2023-01-15T11:00:00Z",
					PaymentOptions: []raw.QuotePaymentOption{
						{EstimatedDelivery: "not-a-timestamp", SourceCurrency: "EUR", TargetCurrency: "USD"},
					},
				}
				_, err := mapQuote(wireQuote, ProfileID{})

				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.call()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if family := errorfamily.Classify(err); family != errorfamily.Corruption {
				t.Errorf("Classify() = %v, want Corruption (error: %v)", family, err)
			}
		})
	}
}

func TestMapDeliveryEstimateParseErrorIsCorruption(t *testing.T) {
	t.Parallel()

	const dateLayout = "2006-01-02T15:04:05.000Z0700"

	// Assert the parser handles Wise's actual wire layout so a regression in
	// parseWiseTimestamp cannot silently break delivery estimates.
	got, err := time.Parse(dateLayout, "2018-01-10T12:15:00.000+0000")
	if err != nil {
		t.Fatalf("wise delivery-estimate timestamp layout must stay parseable: %v", err)
	}

	if got.UTC() != time.Date(2018, time.January, 10, 12, 15, 0, 0, time.UTC) {
		t.Errorf("parsed delivery estimate = %v, want 2018-01-10T12:15:00Z", got.UTC())
	}

	_, err = mapDeliveryEstimate(raw.DeliveryEstimate{EstimatedDeliveryDate: "garbage"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if family := errorfamily.Classify(err); family != errorfamily.Corruption {
		t.Errorf("Classify() = %v, want Corruption (error: %v)", family, err)
	}
}

func TestMapBalanceParseErrorsAreCorruption(t *testing.T) {
	t.Parallel()

	const (
		rawTypeStandard = "STANDARD"
		rawCreatedAt    = "2020-05-27T10:27:22"
	)

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "unparseable creation_time",
			call: func() error {
				b := raw.Balance{ID: 1, Currency: "EUR", Type: rawTypeStandard, CreationTime: "garbage"}
				_, err := mapBalance(b)

				return err
			},
		},
		{
			name: "invalid currency",
			call: func() error {
				b := raw.Balance{ID: 1, Currency: "euros", Type: rawTypeStandard, CreationTime: rawCreatedAt}
				_, err := mapBalance(b)

				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.call()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if family := errorfamily.Classify(err); family != errorfamily.Corruption {
				t.Errorf("Classify() = %v, want Corruption (error: %v)", family, err)
			}
		})
	}
}

// requirementField builds a single-field form requirement for tests.
func TestMapFundTransferResultParseErrorsAreCorruption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "unknown funding type",
			call: func() error {
				_, err := mapFundTransferResult(raw.FundingResponse{Type: "CRYPTO", Status: "COMPLETED"})

				return err
			},
		},
		{
			name: "unknown funding status",
			call: func() error {
				_, err := mapFundTransferResult(raw.FundingResponse{Type: "BALANCE", Status: "MAYBE"})

				return err
			},
		},
		{
			name: "empty funding type from a shape change",
			call: func() error {
				_, err := mapFundTransferResult(raw.FundingResponse{Status: "COMPLETED"})

				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.call()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if family := errorfamily.Classify(err); family != errorfamily.Corruption {
				t.Errorf("Classify() = %v, want Corruption (error: %v)", family, err)
			}
		})
	}
}

// TestClassifyExhaustedRetriesGuardClaauses covers the unwrap arms directly:
// the guard clauses (non-exceeded error, nil / non-response LastResult) must
// return nil so doRequest keeps its original error.
func TestClassifyExhaustedRetriesGuardClauses(t *testing.T) {
	t.Parallel()

	client := New("test-api-key")

	t.Run("non-exceeded error returns nil", func(t *testing.T) {
		t.Parallel()

		got := client.classifyExhaustedRetries("GET", "http://x", errTestPlain)
		if got != nil {
			t.Errorf("classifyExhaustedRetries(plain error) = %v, want nil", got)
		}
	})

	t.Run("exceeded with nil LastResult returns nil", func(t *testing.T) {
		t.Parallel()

		exceeded := retrypolicy.ExceededError{LastResult: nil, LastError: errTestBoom}

		got := client.classifyExhaustedRetries("GET", "http://x", exceeded)
		if got != nil {
			t.Errorf("classifyExhaustedRetries(nil LastResult) = %v, want nil", got)
		}
	})

	t.Run("exceeded with non-response LastResult returns nil", func(t *testing.T) {
		t.Parallel()

		exceeded := retrypolicy.ExceededError{LastResult: "a string, not a response"}

		got := client.classifyExhaustedRetries("GET", "http://x", exceeded)
		if got != nil {
			t.Errorf("classifyExhaustedRetries(string LastResult) = %v, want nil", got)
		}
	})
}

// TestClassifyExhaustedRetriesRateLimit pins the payoff: a retries-exceeded
// error carrying a 429 response surfaces as *RateLimitError with Retry-After
// and the rate-limit scope.
func TestClassifyExhaustedRetriesRateLimit(t *testing.T) {
	t.Parallel()

	client := New("test-api-key")

	resp := httptest.NewRecorder()
	resp.Header().Set("Retry-After", "7")
	resp.Header().Set("X-Rate-Limited-By", "profile")
	resp.WriteHeader(http.StatusTooManyRequests)

	exceeded := retrypolicy.ExceededError{
		LastResult: resp.Result(),
		LastError:  errTestRateLimits,
	}

	got := client.classifyExhaustedRetries("GET", "http://x", exceeded)
	if got == nil {
		t.Fatal("classifyExhaustedRetries(429) = nil, want wrapped RateLimitError")
	}

	rateLimitErr, ok := errors.AsType[*RateLimitError](got)
	if !ok {
		t.Fatalf("got %T, want *RateLimitError: %v", got, got)
	}

	if rateLimitErr.RetryAfter != 7*time.Second {
		t.Errorf("RetryAfter = %v, want 7s", rateLimitErr.RetryAfter)
	}

	if rateLimitErr.RateLimitedBy != "profile" {
		t.Errorf("RateLimitedBy = %q, want %q", rateLimitErr.RateLimitedBy, "profile")
	}
}

// TestErrorContexts pins the structured-context contract of the error types:
// consumers route on these maps in logs and metrics.
func TestErrorContexts(t *testing.T) {
	t.Parallel()

	t.Run("APIError context carries the status code", func(t *testing.T) {
		t.Parallel()

		apiErr := &APIError{StatusCode: http.StatusBadRequest}
		if ctx := apiErr.ErrorContext(); ctx["status_code"] != "400" {
			t.Errorf("ErrorContext = %v, want status_code 400", ctx)
		}
	})

	t.Run("RateLimitError context carries retry-after and scope", func(t *testing.T) {
		t.Parallel()

		rateLimitErr := &RateLimitError{
			APIError:      APIError{StatusCode: http.StatusTooManyRequests},
			RetryAfter:    3 * time.Second,
			RateLimitedBy: "ip",
		}

		ctx := rateLimitErr.ErrorContext()
		if ctx["retry_after"] != (3 * time.Second).String() {
			t.Errorf("ErrorContext = %v, want retry_after 3s", ctx)
		}

		if ctx["rate_limited_by"] != "ip" {
			t.Errorf("ErrorContext = %v, want rate_limited_by ip", ctx)
		}
	})

	t.Run("SCAChallengeError code identifies the challenge", func(t *testing.T) {
		t.Parallel()

		scaErr := &SCAChallengeError{APIError: APIError{StatusCode: http.StatusForbidden}}
		if got := scaErr.ErrorCode(); got != errorCodeSCA {
			t.Errorf("ErrorCode = %q, want %q", got, errorCodeSCA)
		}
	})

	t.Run("SCAChallengeError context carries the 2FA verdict headers", func(t *testing.T) {
		t.Parallel()

		headers := http.Header{}
		headers.Set(HeaderTwoFAApprovalResult, "REJECTED")
		headers.Set(HeaderTwoFAApproval, "ott-123")
		scaErr := &SCAChallengeError{APIError: APIError{StatusCode: http.StatusForbidden, Headers: headers}}

		ctx := scaErr.ErrorContext()
		if ctx["status_code"] != "403" {
			t.Errorf("ErrorContext = %v, want status_code 403", ctx)
		}

		if ctx["approval_result"] != "REJECTED" {
			t.Errorf("ErrorContext = %v, want approval_result REJECTED", ctx)
		}

		if ctx["approval_token_issued"] != "true" {
			t.Errorf("ErrorContext = %v, want approval_token_issued true", ctx)
		}
	})

	t.Run("AuthError and NotFoundError expose the promoted status-code context", func(t *testing.T) {
		t.Parallel()

		authErr := &AuthError{APIError: APIError{StatusCode: http.StatusUnauthorized}}
		if ctx := authErr.ErrorContext(); ctx["status_code"] != "401" {
			t.Errorf("ErrorContext = %v, want status_code 401", ctx)
		}

		notFoundErr := &NotFoundError{APIError: APIError{StatusCode: http.StatusNotFound}}
		if ctx := notFoundErr.ErrorContext(); ctx["status_code"] != "404" {
			t.Errorf("ErrorContext = %v, want status_code 404", ctx)
		}
	})
}

// TestIsRetryableContract pins the retryability contract of the six error
// types: only RateLimitError and ServerError are retryable. The four
// non-retryable types deliberately do not implement IsRetryable — absence is
// the convention for "not retryable".
func TestIsRetryableContract(t *testing.T) {
	t.Parallel()

	for _, err := range []error{
		&APIError{}, &AuthError{}, &SCAChallengeError{}, &NotFoundError{},
	} {
		if _, implements := err.(interface{ IsRetryable() bool }); implements {
			t.Errorf("%T implements IsRetryable, want absence (non-retryable by convention)", err)
		}
	}

	for _, err := range []error{&RateLimitError{}, &ServerError{}} {
		retryable, implements := err.(interface{ IsRetryable() bool })
		if !implements || !retryable.IsRetryable() {
			t.Errorf("%T should implement IsRetryable()=true", err)
		}
	}
}
