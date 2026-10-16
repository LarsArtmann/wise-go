package wise

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

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
			name: "bad request", statusCode: http.StatusBadRequest,
			body:     `{"errors":[{"code":"BAD","message":"nope"}]}`,
			wantCode: "wise.bad_request", wantFamily: errorfamily.Rejection, wantRetry: false,
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

// multiErrorEnvelope is a two-entry Wise error body used across the
// captures and matrix tests.
const multiErrorEnvelope = `{"errors":[{"code":"parameter_missing","message":"source is required"},` +
	`{"code":"not.valid","message":"amount is invalid"}]}`

// TestNewAPIErrorCapturesDetails pins the structured-envelope surface: the
// wire codes/messages land in Details verbatim (unknown codes pass through),
// while Message keeps its historical flattened join format byte-for-byte.
func TestNewAPIErrorCapturesDetails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        string
		wantDetails []WiseErrorDetail
		wantMessage string
	}{
		{
			name: "structured envelope with two entries",
			body: multiErrorEnvelope,
			wantDetails: []WiseErrorDetail{
				{Code: WiseErrorCodeParameterMissing, Message: "source is required"},
				{Code: WiseErrorCodeNotValid, Message: "amount is invalid"},
			},
			wantMessage: "parameter_missing: source is required; not.valid: amount is invalid",
		},
		{
			name: "unknown code passes through",
			body: `{"errors":[{"code":"brand_new_code_from_wise","message":"future failure"}]}`,
			wantDetails: []WiseErrorDetail{
				{Code: WiseErrorCode("brand_new_code_from_wise"), Message: "future failure"},
			},
			wantMessage: "brand_new_code_from_wise: future failure",
		},
		{
			name:        "empty JSON object yields no details",
			body:        `{}`,
			wantDetails: nil,
			wantMessage: `{}`,
		},
		{
			name:        "plain-text body yields no details",
			body:        "Funding failed",
			wantDetails: nil,
			wantMessage: "Funding failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := newAPIError(http.StatusUnprocessableEntity, tt.body, nil, 0, "")

			apiErr, ok := errors.AsType[*APIError](err)
			if !ok {
				t.Fatalf("expected *APIError, got %T", err)
			}

			if len(apiErr.Details) != len(tt.wantDetails) {
				t.Fatalf("Details = %+v, want %+v", apiErr.Details, tt.wantDetails)
			}

			for i, want := range tt.wantDetails {
				if apiErr.Details[i] != want {
					t.Errorf("Details[%d] = %+v, want %+v", i, apiErr.Details[i], want)
				}
			}

			if apiErr.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q (join format must stay byte-identical)", apiErr.Message, tt.wantMessage)
			}
		})
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

	// The token value must stay OUT of the message: error strings are the
	// surface most likely to end up in logs, and the OTT authorizes the
	// challenge. Callers fetch it explicitly via TwoFAApprovalToken().
	if strings.Contains(sca.Error(), "bb676aeb-7c4d-4930-bb55-ab949fd3fd87") {
		t.Errorf("Error() must not embed the one-time token, got: %s", sca.Error())
	}
}

// TestCheckErrorTokenlessSCAChallenge pins Wise's PSD2 flavor: a 403 with
// the verdict header but NO x-2fa-approval OTT (live 2026-10-16). The error
// is still an SCA challenge — blocked — but TwoFAApprovalToken() is zero:
// nothing to clear programmatically, and the context reports issued=false.
func TestCheckErrorTokenlessSCAChallenge(t *testing.T) {
	t.Parallel()

	client := &Client{}
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Header:     http.Header{HeaderTwoFAApprovalResult: []string{"REJECTED"}},
		Body:       http.NoBody,
	}

	err := client.checkError(resp)

	sca, ok := errors.AsType[*SCAChallengeError](err)
	if !ok {
		t.Fatalf("expected *SCAChallengeError, got %T: %v", err, err)
	}

	if !sca.TwoFAApprovalToken().IsZero() {
		t.Errorf("tokenless challenge must yield a zero OneTimeToken, got %q", sca.TwoFAApprovalToken())
	}

	if got := sca.ErrorContext()["approval_token_issued"]; got != "false" {
		t.Errorf("approval_token_issued = %q, want \"false\"", got)
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

// TestBadRequestErrorClassification pins the 400 surface: typed error code,
// promoted Rejection family, no IsRetryable method, promoted status-code
// context, and code-sentinel matching through the embedded APIError.
func TestBadRequestErrorClassification(t *testing.T) {
	t.Parallel()

	err := newAPIError(http.StatusBadRequest,
		`{"errors":[{"code":"parameter_missing","message":"source is required"}]}`, nil, 0, "")

	badRequest, ok := errors.AsType[*BadRequestError](err)
	if !ok {
		t.Fatalf("expected *BadRequestError, got %T", err)
	}

	if got := badRequest.ErrorCode(); got != errorCodeBadRequest {
		t.Errorf("ErrorCode() = %q, want %q", got, errorCodeBadRequest)
	}

	if got := badRequest.ErrorFamily(); got != errorfamily.Rejection {
		t.Errorf("ErrorFamily() = %v, want Rejection (promoted)", got)
	}

	if _, implements := err.(interface{ IsRetryable() bool }); implements {
		t.Errorf("BadRequestError implements IsRetryable, want absence")
	}

	ctx := badRequest.ErrorContext()
	if ctx["status_code"] != "400" {
		t.Errorf("ErrorContext = %v, want promoted status_code 400", ctx)
	}

	if !errors.Is(err, ErrParameterMissing) {
		t.Errorf("BadRequestError did not match ErrParameterMissing through promotion")
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

	t.Run("APIError context carries the first granular wise_code when details exist", func(t *testing.T) {
		t.Parallel()

		apiErr := &APIError{
			StatusCode: http.StatusBadRequest,
			Details: []WiseErrorDetail{
				{Code: WiseErrorCodeParameterMissing, Message: "source is required"},
				{Code: WiseErrorCodeNotValid, Message: "amount is invalid"},
			},
		}

		ctx := apiErr.ErrorContext()
		if got := ctx["wise_code"]; got != "parameter_missing" {
			t.Errorf("wise_code = %q, want the first detail's code parameter_missing", got)
		}
	})

	t.Run("APIError context omits wise_code without details", func(t *testing.T) {
		t.Parallel()

		apiErr := &APIError{StatusCode: http.StatusConflict}
		if _, present := apiErr.ErrorContext()["wise_code"]; present {
			t.Errorf("wise_code present without details, want absent")
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

	t.Run("RateLimitError context carries the granular wise_code", func(t *testing.T) {
		t.Parallel()

		rateLimitErr := &RateLimitError{
			APIError: APIError{
				StatusCode: http.StatusTooManyRequests,
				Details:    []WiseErrorDetail{{Code: WiseErrorCodeRateLimitExceeded, Message: "slow down"}},
			},
			RetryAfter: time.Second,
		}

		if got := rateLimitErr.ErrorContext()["wise_code"]; got != "rate_limit_exceeded" {
			t.Errorf("wise_code = %q, want rate_limit_exceeded", got)
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

// TestIsRetryableContract pins the retryability contract of the seven error
// types: only RateLimitError and ServerError are retryable. The five
// non-retryable types deliberately do not implement IsRetryable — absence is
// the convention for "not retryable".
func TestIsRetryableContract(t *testing.T) {
	t.Parallel()

	for _, err := range []error{
		&APIError{}, &AuthError{}, &SCAChallengeError{}, &NotFoundError{}, &BadRequestError{},
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

// TestErrorCodesUniqueAndWellFormed pins the central error-code catalog:
// one entry per code, every code a lowercase dotted wise.* identifier. The
// list is exhaustive on purpose — adding a code constant without listing it
// here breaks the catalog's completeness contract.
func TestErrorCodesUniqueAndWellFormed(t *testing.T) {
	t.Parallel()

	codes := []string{
		errorCodeAPI,
		errorCodeRateLimit,
		errorCodeAuth,
		errorCodeNotFound,
		errorCodeServer,
		errorCodeSCA,
		errorCodeAccountInvalidRequest,
		errorCodeAccountParseCreationTime,
		errorCodeAccountParseModificationTime,
		errorCodeAccountDetailsInvalidRequest,
		errorCodeAccountDetailsParseIdentification,
		errorCodeBalanceInvalidRequest,
		errorCodeBalanceParseCreationTime,
		errorCodeBalanceParseCurrency,
		errorCodeBalanceParseType,
		errorCodeDeliveryEstimateParseDate,
		errorCodeFundingParseStatus,
		errorCodeFundingParseType,
		errorCodeMoneyParse,
		errorCodeOTTChannelUnsupported,
		errorCodeOTTDecode,
		errorCodeOTTInvalidRequest,
		errorCodeOTTUncleared,
		errorCodeProfileInvalidRequest,
		errorCodeProfileParseCreatedAt,
		errorCodeProfileParseType,
		errorCodeQuoteInvalidRequest,
		errorCodeQuoteParseCreated,
		errorCodeQuoteParseExpiration,
		errorCodeQuoteParsePaymentOptionDelivery,
		errorCodeQuoteParsePaymentOptionSource,
		errorCodeQuoteParsePaymentOptionTarget,
		errorCodeQuoteParseSourceCurrency,
		errorCodeQuoteParseTargetCurrency,
		errorCodeRatesEmptyResponse,
		errorCodeRatesInvalidRequest,
		errorCodeRatesParseSource,
		errorCodeRatesParseTarget,
		errorCodeRatesParseTime,
		errorCodeRecipientInvalidRequest,
		errorCodeRecipientMap,
		errorCodeRecipientParseCurrency,
		errorCodeResponseDecode,
		errorCodeTotalFundsInvalidRequest,
		errorCodeTransactionParseDate,
		errorCodeTransactionsInvalidRequest,
		errorCodeTransferInvalidRequest,
		errorCodeTransferMap,
		errorCodeTransferParseCreated,
		errorCodeTransferParseCurrency,
		errorCodeTransferParsePaymentCreated,
		errorCodeTransferParsePaymentCurrency,
		errorCodeTransferParsePaymentUpdated,
		errorCodeUserInvalidRequest,
		errorCodeUserParseDateOfBirth,
		errorCodeWebhookDecode,
		errorCodeWebhookInvalidRequest,
	}

	wellFormed := regexp.MustCompile(`^wise\.[a-z0-9_]+(\.[a-z0-9_]+)*$`)

	seen := make(map[string]bool, len(codes))
	for _, code := range codes {
		if !wellFormed.MatchString(code) {
			t.Errorf("code %q is not a lowercase dotted wise.* identifier", code)
		}

		if seen[code] {
			t.Errorf("code %q is listed/declared more than once in the catalog", code)
		}

		seen[code] = true
	}

	// Wise's granular wire codes follow the same completeness contract, with
	// their own shape: lowercase segments of letters/digits/underscores,
	// dotted separators, no wise. prefix (these are Wise's codes, not ours).
	wiseCodes := []WiseErrorCode{
		WiseErrorCodeNotValid,
		WiseErrorCodeInvalidValue,
		WiseErrorCodeParameterMissing,
		WiseErrorCodeParameterInvalid,
		WiseErrorCodeMissingValue,
		WiseErrorCodeInvalidRequest,
		WiseErrorCodeUnauthorized,
		WiseErrorCodeForbidden,
		WiseErrorCodeResourceNotFound,
		WiseErrorCodeOperationNotSupported,
		WiseErrorCodeRateLimitExceeded,
		WiseErrorCodeSystemError,
		WiseErrorCodeInvalidFileExtension,
		WiseErrorCodeUnexpectedError,
	}

	wireFormed := regexp.MustCompile(`^[a-z0-9_]+(\.[a-z0-9_]+)*$`)

	seenWise := make(map[WiseErrorCode]bool, len(wiseCodes))
	for _, code := range wiseCodes {
		if !wireFormed.MatchString(string(code)) {
			t.Errorf("Wise error code %q is not a lowercase dotted wire identifier", code)
		}

		if seenWise[code] {
			t.Errorf("Wise error code %q is listed/declared more than once", code)
		}

		seenWise[code] = true
	}
}

// TestSentinelsStayInterfaceTyped pins the sentinel declaration contract:
// package-level sentinels live in errors.go's central var block, declared
// with the error interface type so they are matched by value (errors.Is),
// never by type.
func TestSentinelsStayInterfaceTyped(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("errors.go")
	if err != nil {
		t.Fatalf("read errors.go: %v", err)
	}

	text := string(source)
	for _, name := range []string{
		"ErrEmptyRateList",
		"ErrAccountIdentificationCardinality",
		"ErrOTTAlternativesShape",
		"ErrTransferRefundAmountMissing",
		"ErrNotValid",
		"ErrInvalidValue",
		"ErrParameterMissing",
		"ErrParameterInvalid",
		"ErrInvalidRequest",
		"ErrUnauthorized",
		"ErrForbidden",
		"ErrResourceNotFound",
		"ErrOperationNotSupported",
		"ErrRateLimitExceeded",
		"ErrSystemError",
	} {
		declared := regexp.MustCompile(`(?m)^\t` + name + ` error = errors\.New\(`).MatchString(text)
		if !declared {
			t.Errorf("sentinel %s must stay declared as `%s error = errors.New(...)` in errors.go", name, name)
		}
	}
}

// TestSentinelsMatchThroughProductionWraps pins that every sentinel stays
// matchable with errors.Is through the exact wrap shapes the SDK produces.
func TestSentinelsMatchThroughProductionWraps(t *testing.T) {
	t.Parallel()

	wrapped := errorfamily.WrapCorruption(ErrEmptyRateList, errorCodeRatesEmptyResponse, "no rates")
	if !errors.Is(wrapped, ErrEmptyRateList) {
		t.Error("WrapCorruption wrap lost sentinel match for ErrEmptyRateList")
	}

	wrapped = errorfamily.WrapCorruption(
		ErrAccountIdentificationCardinality,
		errorCodeAccountDetailsParseIdentification,
		"found 2",
	)
	if !errors.Is(wrapped, ErrAccountIdentificationCardinality) {
		t.Error("WrapCorruption wrap lost sentinel match for ErrAccountIdentificationCardinality")
	}

	wrapped = errorfamily.WrapCorruption(ErrTransferRefundAmountMissing, errorCodeWebhookDecode, "refund")
	if !errors.Is(wrapped, ErrTransferRefundAmountMissing) {
		t.Error("WrapCorruption wrap lost sentinel match for ErrTransferRefundAmountMissing")
	}

	fmtWrapped := fmt.Errorf("%w: %s", ErrOTTAlternativesShape, `{"not":"an array"}`)
	if !errors.Is(fmtWrapped, ErrOTTAlternativesShape) {
		t.Error("fmt.Errorf wrap lost sentinel match for ErrOTTAlternativesShape")
	}
}

// TestCodeSentinelsMapCuratedCodes pins the codeSentinels mapping: exactly
// the eleven wave-1 sentinels, each bound to its documented wire code.
func TestCodeSentinelsMapCuratedCodes(t *testing.T) {
	t.Parallel()

	want := map[error]WiseErrorCode{
		ErrNotValid:              WiseErrorCodeNotValid,
		ErrInvalidValue:          WiseErrorCodeInvalidValue,
		ErrParameterMissing:      WiseErrorCodeParameterMissing,
		ErrParameterInvalid:      WiseErrorCodeParameterInvalid,
		ErrInvalidRequest:        WiseErrorCodeInvalidRequest,
		ErrUnauthorized:          WiseErrorCodeUnauthorized,
		ErrForbidden:             WiseErrorCodeForbidden,
		ErrResourceNotFound:      WiseErrorCodeResourceNotFound,
		ErrOperationNotSupported: WiseErrorCodeOperationNotSupported,
		ErrRateLimitExceeded:     WiseErrorCodeRateLimitExceeded,
		ErrSystemError:           WiseErrorCodeSystemError,
	}

	if len(codeSentinels) != len(want) {
		t.Fatalf("codeSentinels has %d entries, want %d; update this pin when curating wave-2",
			len(codeSentinels), len(want))
	}

	for sentinel, wantCode := range want {
		gotCode, ok := codeSentinels[sentinel]
		if !ok {
			t.Errorf("codeSentinels is missing a curated sentinel (%v)", sentinel)

			continue
		}

		if gotCode != wantCode {
			t.Errorf("codeSentinels[%v] = %q, want %q", sentinel, gotCode, wantCode)
		}
	}
}

// TestAPIErrorMatchesCuratedCodeSentinels pins the Is contract end-to-end:
// structured-code matching works bare, through fmt.Errorf chains, through
// Corruption wraps, and through every subtype that embeds APIError — and
// never fires without the wire code or for non-sentinel targets.
func TestAPIErrorMatchesCuratedCodeSentinels(t *testing.T) {
	t.Parallel()

	notFoundBody := `{"errors":[{"code":"resource_not_found","message":"profile not found"}]}`

	t.Run("bare APIError matches", func(t *testing.T) {
		t.Parallel()

		err := newAPIError(http.StatusNotFound, notFoundBody, nil, 0, "")
		if !errors.Is(err, ErrResourceNotFound) {
			t.Errorf("errors.Is(err, ErrResourceNotFound) = false, want true")
		}
	})

	t.Run("fmt.Errorf wrap chain still matches", func(t *testing.T) {
		t.Parallel()

		err := newAPIError(http.StatusNotFound, notFoundBody, nil, 0, "")
		wrapped := fmt.Errorf("get transfer: %w", err)

		if !errors.Is(wrapped, ErrResourceNotFound) {
			t.Errorf("wrapped error lost ErrResourceNotFound match")
		}
	})

	t.Run("Corruption wrap still matches", func(t *testing.T) {
		t.Parallel()

		err := newAPIError(http.StatusNotFound, notFoundBody, nil, 0, "")
		wrapped := errorfamily.WrapCorruption(err, errorCodeResponseDecode, "decode context")

		if !errors.Is(wrapped, ErrResourceNotFound) {
			t.Errorf("Corruption-wrapped error lost ErrResourceNotFound match")
		}
	})

	t.Run("subtype promotion: AuthError matches ErrForbidden", func(t *testing.T) {
		t.Parallel()

		err := newAPIError(http.StatusForbidden,
			`{"errors":[{"code":"forbidden","message":"insufficient permissions"}]}`, nil, 0, "")
		if _, ok := errors.AsType[*AuthError](err); !ok {
			t.Fatalf("expected *AuthError, got %T", err)
		}

		if !errors.Is(err, ErrForbidden) {
			t.Errorf("AuthError did not match ErrForbidden through promotion")
		}
	})

	t.Run("429 matches ErrRateLimitExceeded when the body says so", func(t *testing.T) {
		t.Parallel()

		err := newAPIError(http.StatusTooManyRequests,
			`{"errors":[{"code":"rate_limit_exceeded","message":"slow down"}]}`, nil, time.Second, "ip")
		if !errors.Is(err, ErrRateLimitExceeded) {
			t.Errorf("RateLimitError did not match ErrRateLimitExceeded")
		}
	})

	t.Run("different sentinel does not match", func(t *testing.T) {
		t.Parallel()

		err := newAPIError(http.StatusNotFound, notFoundBody, nil, 0, "")
		if errors.Is(err, ErrRateLimitExceeded) {
			t.Errorf("resource_not_found error matched ErrRateLimitExceeded")
		}
	})

	t.Run("no envelope means no code-sentinel match", func(t *testing.T) {
		t.Parallel()

		err := newAPIError(http.StatusNotFound, "plain 404", nil, 0, "")
		if errors.Is(err, ErrResourceNotFound) {
			t.Errorf("error without structured details matched ErrResourceNotFound")
		}
	})

	t.Run("non-sentinel targets stay unmatched", func(t *testing.T) {
		t.Parallel()

		err := newAPIError(http.StatusNotFound, notFoundBody, nil, 0, "")
		if errors.Is(err, ErrEmptyRateList) {
			t.Errorf("APIError matched the response-shape sentinel ErrEmptyRateList")
		}

		if errors.Is(err, io.EOF) {
			t.Errorf("APIError matched the unrelated stdlib sentinel io.EOF")
		}

		if errors.Is(err, nil) { //nolint:legacyerrors // nil target is not a sentinel
			t.Errorf("errors.Is(err, nil) = true, want false")
		}
	})
}

// TestAPIErrorHasCode pins the open-enum matcher: any wire code works,
// curated or not, and errors without details never do.
func TestAPIErrorHasCode(t *testing.T) {
	t.Parallel()

	err := newAPIError(http.StatusUnprocessableEntity,
		`{"errors":[{"code":"brand_new_code","message":"from the future"}]}`, nil, 0, "")

	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if !apiErr.HasCode(WiseErrorCode("brand_new_code")) {
		t.Errorf("HasCode(brand_new_code) = false, want true (open enum)")
	}

	if apiErr.HasCode(WiseErrorCodeResourceNotFound) {
		t.Errorf("HasCode(resource_not_found) = true, want false")
	}

	empty := &APIError{StatusCode: http.StatusBadRequest}
	if empty.HasCode(WiseErrorCodeNotValid) {
		t.Errorf("HasCode on details-less error = true, want false")
	}
}

// TestAsTypeAndIsAreComplementary pins the matching-idiom split the error
// system commits to: error TYPES (with their fields) extract via
// errors.AsType, curated Wise wire CODES match via errors.Is, and neither
// substitutes for the other. This is the contract the README decision tree
// teaches, so a regression here invalidates the docs.
func TestAsTypeAndIsAreComplementary(t *testing.T) {
	t.Parallel()

	limitedBody := `{"errors":[{"code":"rate_limit_exceeded","message":"slow down"}]}`

	t.Run("429 extracts the type and matches the code", func(t *testing.T) {
		t.Parallel()

		err := newAPIError(http.StatusTooManyRequests, limitedBody, nil, 7*time.Second, "ip")

		rateLimited, ok := errors.AsType[*RateLimitError](err)
		if !ok {
			t.Fatalf("AsType[*RateLimitError] = false, want true")
		}

		if rateLimited.RetryAfter != 7*time.Second {
			t.Errorf("RetryAfter = %v, want 7s (type carries the fields)", rateLimited.RetryAfter)
		}

		if !errors.Is(err, ErrRateLimitExceeded) {
			t.Errorf("Is(ErrRateLimitExceeded) = false, want true (code matches by value)")
		}
	})

	t.Run("type extracts even without the body code", func(t *testing.T) {
		t.Parallel()

		err := newAPIError(http.StatusTooManyRequests, "{}", nil, 5*time.Second, "profile")

		if _, ok := errors.AsType[*RateLimitError](err); !ok {
			t.Fatalf("AsType[*RateLimitError] = false, want true regardless of body codes")
		}

		if errors.Is(err, ErrRateLimitExceeded) {
			t.Errorf("Is(ErrRateLimitExceeded) = true without the wire code, want false")
		}
	})

	t.Run("a code never impersonates a type", func(t *testing.T) {
		t.Parallel()

		err := newAPIError(http.StatusBadRequest,
			`{"errors":[{"code":"system_error","message":"wise is sad"}]}`, nil, 0, "")

		badRequest, ok := errors.AsType[*BadRequestError](err)
		if !ok {
			t.Fatalf("AsType[*BadRequestError] = false, want true (status drives the type)")
		}

		if badRequest.StatusCode != http.StatusBadRequest {
			t.Errorf("StatusCode = %d, want 400", badRequest.StatusCode)
		}

		if _, ok := errors.AsType[*ServerError](err); ok {
			t.Errorf("AsType[*ServerError] = true on a 400, want false (status drives the type)")
		}

		if !errors.Is(err, ErrSystemError) {
			t.Errorf("Is(ErrSystemError) = false, want true (body code matches regardless of type)")
		}
	})
}

// errorMatches reports whether err carries the given concrete error type.
// The matrix rows use it as their wantType matcher.
func errorMatches[T error](err error) bool {
	_, ok := errors.AsType[T](err)

	return ok
}

// TestErrorSurfaceMappingMatrix pins the ENTIRE error surface end-to-end in
// one table: every status x envelope-shape combination maps to its typed
// error, structured details, code-sentinel matches, classification, and
// context keys. A change to newAPIError/classifyAPIError that breaks any
// consumer-visible dimension fails here.
func TestErrorSurfaceMappingMatrix(t *testing.T) {
	t.Parallel()

	const envelope = `{"errors":[{"code":%q,"message":"wire message"}]}`

	tests := []errorMatrixRow{
		{
			name:            "400 envelope maps to BadRequestError with its code",
			statusCode:      http.StatusBadRequest,
			body:            fmt.Sprintf(envelope, "parameter_missing"),
			wantType:        errorMatches[*BadRequestError],
			wantErrorCode:   "wise.bad_request",
			wantFamily:      errorfamily.Rejection,
			wantDetailCount: 1,
			wantIs:          []error{ErrParameterMissing},
			wantNotIs:       []error{ErrRateLimitExceeded},
			wantHasCodes:    []WiseErrorCode{WiseErrorCodeParameterMissing},
			wantContextKeys: []string{"status_code"},
		},
		{
			name:            "400 plain text maps to BadRequestError without details",
			statusCode:      http.StatusBadRequest,
			body:            "validation failed",
			wantType:        errorMatches[*BadRequestError],
			wantErrorCode:   "wise.bad_request",
			wantFamily:      errorfamily.Rejection,
			wantDetailCount: 0,
			wantNotIs:       []error{ErrNotValid, ErrInvalidRequest},
			wantContextKeys: []string{"status_code"},
		},
		{
			name:            "401 envelope maps to AuthError matching ErrUnauthorized",
			statusCode:      http.StatusUnauthorized,
			body:            fmt.Sprintf(envelope, "unauthorized"),
			wantType:        errorMatches[*AuthError],
			wantErrorCode:   "wise.auth",
			wantFamily:      errorfamily.Rejection,
			wantDetailCount: 1,
			wantIs:          []error{ErrUnauthorized},
			wantHasCodes:    []WiseErrorCode{WiseErrorCodeUnauthorized},
			wantContextKeys: []string{"status_code"},
		},
		{
			name:            "403 envelope maps to AuthError matching ErrForbidden",
			statusCode:      http.StatusForbidden,
			body:            fmt.Sprintf(envelope, "forbidden"),
			wantType:        errorMatches[*AuthError],
			wantErrorCode:   "wise.auth",
			wantFamily:      errorfamily.Rejection,
			wantDetailCount: 1,
			wantIs:          []error{ErrForbidden},
			wantNotIs:       []error{ErrUnauthorized},
			wantHasCodes:    []WiseErrorCode{WiseErrorCodeForbidden},
			wantContextKeys: []string{"status_code"},
		},
		{
			name:       "403 empty body with 2FA headers maps to SCAChallengeError",
			statusCode: http.StatusForbidden,
			headers: http.Header{
				HeaderTwoFAApprovalResult: []string{"REJECTED"},
				HeaderTwoFAApproval:       []string{"ott-value"},
			},
			wantType:        errorMatches[*SCAChallengeError],
			wantErrorCode:   "wise.sca_challenge",
			wantFamily:      errorfamily.Rejection,
			wantDetailCount: 0,
			wantNotIs:       []error{ErrForbidden},
			wantContextKeys: []string{"status_code", "approval_result", "approval_token_issued"},
		},
		{
			name:            "404 envelope maps to NotFoundError matching ErrResourceNotFound",
			statusCode:      http.StatusNotFound,
			body:            fmt.Sprintf(envelope, "resource_not_found"),
			wantType:        errorMatches[*NotFoundError],
			wantErrorCode:   "wise.not_found",
			wantFamily:      errorfamily.Rejection,
			wantDetailCount: 1,
			wantIs:          []error{ErrResourceNotFound},
			wantHasCodes:    []WiseErrorCode{WiseErrorCodeResourceNotFound},
			wantContextKeys: []string{"status_code"},
		},
		{
			name:            "409 plain text maps to bare APIError",
			statusCode:      http.StatusConflict,
			body:            "Transfer is already funded",
			wantType:        errorMatches[*APIError],
			wantErrorCode:   "wise.api_error",
			wantFamily:      errorfamily.Rejection,
			wantDetailCount: 0,
			wantContextKeys: []string{"status_code"},
		},
		{
			name:            "429 envelope maps to RateLimitError matching ErrRateLimitExceeded",
			statusCode:      http.StatusTooManyRequests,
			body:            fmt.Sprintf(envelope, "rate_limit_exceeded"),
			retryAfter:      3 * time.Second,
			rateLimitedBy:   "ip",
			wantType:        errorMatches[*RateLimitError],
			wantErrorCode:   "wise.rate_limit",
			wantFamily:      errorfamily.Transient,
			wantRetryable:   true,
			wantDetailCount: 1,
			wantIs:          []error{ErrRateLimitExceeded},
			wantHasCodes:    []WiseErrorCode{WiseErrorCodeRateLimitExceeded},
			wantContextKeys: []string{"status_code", "retry_after", "rate_limited_by"},
		},
		{
			name:            "500 envelope with system_error stays a retryable ServerError",
			statusCode:      http.StatusInternalServerError,
			body:            fmt.Sprintf(envelope, "system_error"),
			wantType:        errorMatches[*ServerError],
			wantErrorCode:   "wise.server",
			wantFamily:      errorfamily.Transient,
			wantRetryable:   true,
			wantDetailCount: 1,
			wantIs:          []error{ErrSystemError},
			wantHasCodes:    []WiseErrorCode{WiseErrorCodeSystemError},
			wantContextKeys: []string{"status_code"},
		},
		{
			name:            "503 plain text maps to a retryable ServerError without details",
			statusCode:      http.StatusServiceUnavailable,
			body:            "upstream hiccup",
			wantType:        errorMatches[*ServerError],
			wantErrorCode:   "wise.server",
			wantFamily:      errorfamily.Transient,
			wantRetryable:   true,
			wantDetailCount: 0,
			wantNotIs:       []error{ErrSystemError},
			wantContextKeys: []string{"status_code"},
		},
		{
			name:            "422 multi-error envelope maps to bare APIError with both codes",
			statusCode:      http.StatusUnprocessableEntity,
			body:            `{"errors":[{"code":"parameter_missing","message":"a"},{"code":"not.valid","message":"b"}]}`,
			wantType:        errorMatches[*APIError],
			wantErrorCode:   "wise.api_error",
			wantFamily:      errorfamily.Rejection,
			wantDetailCount: 2,
			wantIs:          []error{ErrParameterMissing, ErrNotValid},
			wantHasCodes:    []WiseErrorCode{WiseErrorCodeParameterMissing, WiseErrorCodeNotValid},
			wantContextKeys: []string{"status_code"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := newAPIError(tt.statusCode, tt.body, tt.headers, tt.retryAfter, tt.rateLimitedBy)
			if err == nil {
				t.Fatal("newAPIError returned nil")
			}

			assertMatrixClassification(t, tt, err)

			apiErr := embeddedAPIError(err)
			if apiErr == nil {
				t.Fatalf("no embedded APIError reachable from %T", err)
			}

			assertMatrixMatching(t, tt, err, apiErr)
			assertMatrixContext(t, tt, err)
		})
	}
}

// errorMatrixRow is one status x envelope-shape expectation of
// TestErrorSurfaceMappingMatrix.
type errorMatrixRow struct {
	name            string
	statusCode      int
	body            string
	headers         http.Header
	retryAfter      time.Duration
	rateLimitedBy   string
	wantType        func(error) bool
	wantErrorCode   string
	wantFamily      errorfamily.Family
	wantRetryable   bool
	wantDetailCount int
	wantIs          []error
	wantNotIs       []error
	wantHasCodes    []WiseErrorCode
	wantContextKeys []string
}

// assertMatrixClassification checks the typed classification of one matrix
// row: concrete type, ErrorCode, ErrorFamily, and the IsRetryable contract.
func assertMatrixClassification(t *testing.T, tt errorMatrixRow, err error) {
	t.Helper()

	if !tt.wantType(err) {
		t.Errorf("type mismatch: got %T", err)
	}

	coder, ok := err.(interface{ ErrorCode() string })
	if !ok {
		t.Fatalf("%T does not implement ErrorCode", err)
	}

	if got := coder.ErrorCode(); got != tt.wantErrorCode {
		t.Errorf("ErrorCode() = %q, want %q", got, tt.wantErrorCode)
	}

	familier, ok := err.(interface{ ErrorFamily() errorfamily.Family })
	if !ok {
		t.Fatalf("%T does not implement ErrorFamily", err)
	}

	if got := familier.ErrorFamily(); got != tt.wantFamily {
		t.Errorf("ErrorFamily() = %v, want %v", got, tt.wantFamily)
	}

	retryable, implements := err.(interface{ IsRetryable() bool })
	if tt.wantRetryable && (!implements || !retryable.IsRetryable()) {
		t.Errorf("IsRetryable() = false/absent, want true")
	}

	if !tt.wantRetryable && implements {
		t.Errorf("IsRetryable implemented, want absence")
	}
}

// embeddedAPIError reaches the APIError value embedded in whichever typed
// wrapper newAPIError produced.
func embeddedAPIError(err error) *APIError {
	if typed, ok := errors.AsType[*APIError](err); ok {
		return typed
	}

	if typed, ok := errors.AsType[*RateLimitError](err); ok {
		return &typed.APIError
	}

	if typed, ok := errors.AsType[*AuthError](err); ok {
		return &typed.APIError
	}

	if typed, ok := errors.AsType[*SCAChallengeError](err); ok {
		return &typed.APIError
	}

	if typed, ok := errors.AsType[*NotFoundError](err); ok {
		return &typed.APIError
	}

	if typed, ok := errors.AsType[*BadRequestError](err); ok {
		return &typed.APIError
	}

	if typed, ok := errors.AsType[*ServerError](err); ok {
		return &typed.APIError
	}

	return nil
}

// assertMatrixMatching checks the structured-details and code-sentinel
// surface of one matrix row: Headers capture, Details count, errors.Is
// matches and non-matches, and HasCode.
func assertMatrixMatching(t *testing.T, tt errorMatrixRow, err error, apiErr *APIError) {
	t.Helper()

	if apiErr.Headers == nil {
		t.Errorf("Headers is nil, want non-nil (captured even when empty)")
	}

	if len(apiErr.Details) != tt.wantDetailCount {
		t.Errorf("len(Details) = %d, want %d", len(apiErr.Details), tt.wantDetailCount)
	}

	for _, sentinel := range tt.wantIs {
		if !errors.Is(err, sentinel) {
			t.Errorf("errors.Is(err, %v) = false, want true", sentinel)
		}
	}

	for _, sentinel := range tt.wantNotIs {
		if errors.Is(err, sentinel) {
			t.Errorf("errors.Is(err, %v) = true, want false", sentinel)
		}
	}

	for _, code := range tt.wantHasCodes {
		if !apiErr.HasCode(code) {
			t.Errorf("HasCode(%q) = false, want true", code)
		}
	}
}

// assertMatrixContext checks the OUTER error's ErrorContext keys (subtype
// overrides like the SCA headers and rate-limit fields matter here).
func assertMatrixContext(t *testing.T, tt errorMatrixRow, err error) {
	t.Helper()

	ctx, ok := err.(interface{ ErrorContext() map[string]string })
	if !ok {
		t.Fatalf("%T does not implement ErrorContext", err)
	}

	got := ctx.ErrorContext()
	for _, key := range tt.wantContextKeys {
		if _, present := got[key]; !present {
			t.Errorf("ErrorContext missing key %q (got %v)", key, got)
		}
	}
}

// FuzzParseErrorResponse hardens the envelope parser: for any input it must
// either return nil or a non-empty error list, and it must never panic.
func FuzzParseErrorResponse(f *testing.F) {
	f.Add(`{"errors":[{"code":"parameter_missing","message":"source is required"}]}`)
	f.Add(`{"errors":[]}`)
	f.Add(`{}`)
	f.Add(``)
	f.Add(`not json at all`)
	f.Add(`{"errors":[{"code":"","message":""}]}`)
	f.Add(`{"errors":[{"code":"a"},{"code":"b","message":"m"}]}`)
	f.Add(`[1,2,3]`)
	f.Add(`null`)

	f.Fuzz(func(t *testing.T, body string) {
		resp := parseErrorResponse(body)
		if resp == nil {
			return
		}

		if len(resp.Errors) == 0 {
			t.Errorf("parseErrorResponse(%q) returned non-nil with zero errors", body)
		}
	})
}

// FuzzNewAPIError pins the parser invariants over arbitrary input: no panic,
// deterministic classification, and byte-exact Details↔Message consistency.
func FuzzNewAPIError(f *testing.F) {
	f.Add(http.StatusBadRequest, `{"errors":[{"code":"parameter_missing","message":"source"}]}`)
	f.Add(http.StatusTooManyRequests, `{"errors":[{"code":"rate_limit_exceeded","message":"slow"}]}`)
	f.Add(http.StatusForbidden, "")
	f.Add(http.StatusInternalServerError, "plain text")
	f.Add(http.StatusNotFound, `{}`)
	f.Add(http.StatusUnauthorized, `{"errors":[]}`)

	f.Fuzz(func(t *testing.T, rawStatus int, body string) {
		statusCode := 100 + ((rawStatus%500)+500)%500

		err := newAPIError(statusCode, body, nil, 0, "")
		if err == nil {
			t.Fatal("newAPIError returned nil")
		}

		rebuilt := newAPIError(statusCode, body, nil, 0, "")

		if fmt.Sprintf("%T", err) != fmt.Sprintf("%T", rebuilt) {
			t.Errorf("classification nondeterministic: %T vs %T", err, rebuilt)
		}

		first, ok := err.(interface{ ErrorCode() string })

		second, ok2 := rebuilt.(interface{ ErrorCode() string })
		if !ok || !ok2 || first.ErrorCode() != second.ErrorCode() {
			t.Errorf("ErrorCode nondeterministic for status %d", statusCode)
		}

		apiErr := embeddedAPIError(err)
		if apiErr == nil {
			t.Fatalf("no embedded APIError reachable from %T", err)
		}

		if len(apiErr.Details) > 0 {
			msgs := make([]string, 0, len(apiErr.Details))
			for _, detail := range apiErr.Details {
				msgs = append(msgs, fmt.Sprintf("%s: %s", string(detail.Code), detail.Message))

				if !apiErr.HasCode(detail.Code) {
					t.Errorf("HasCode(%q) = false for a detail present in Details", detail.Code)
				}
			}

			if want := strings.Join(msgs, "; "); apiErr.Message != want {
				t.Errorf("Message %q does not join Details (want %q)", apiErr.Message, want)
			}
		} else if apiErr.Message != body {
			t.Errorf("Message %q = body %q expected when the envelope is absent", apiErr.Message, body)
		}
	})
}
