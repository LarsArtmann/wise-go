package wise

import (
	"context"
	"errors"
	"io"
	"net/http"
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

func TestClassifyTransactionType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		wiseType DetailType
		amount   float64
		want     TransactionType
	}{
		{name: "card payment debit", wiseType: DetailTypeCardPayment, amount: -10, want: TransactionTypeCard},
		{
			name:     "card payment credit still card (not refund)",
			wiseType: DetailTypeCardPayment,
			amount:   25,
			want:     TransactionTypeCard,
		},
		{name: "card payment zero still card", wiseType: DetailTypeCardPayment, amount: 0, want: TransactionTypeCard},
		{name: "card refund positive", wiseType: DetailTypeCardRefund, amount: 25, want: TransactionTypeRefund},
		{name: "card refund zero", wiseType: DetailTypeCardRefund, amount: 0, want: TransactionTypeCard},
		{name: "transfer", wiseType: DetailTypeTransfer, amount: 100, want: TransactionTypeTransfer},
		{name: "payment", wiseType: DetailTypePayment, amount: -50, want: TransactionTypePayment},
		{name: "conversion", wiseType: DetailTypeConversion, amount: -100, want: TransactionTypeExchange},
		{name: "exchange alias", wiseType: DetailTypeExchange, amount: 50, want: TransactionTypeExchange},
		{name: "fee", wiseType: DetailTypeFee, amount: -0.5, want: TransactionTypeFee},
		{name: "unknown positive is credit", wiseType: "SOMETHING_NEW", amount: 10, want: TransactionTypeCredit},
		{name: "unknown negative is debit", wiseType: "SOMETHING_NEW", amount: -10, want: TransactionTypeDebit},
		{name: "unknown zero is debit", wiseType: "SOMETHING_NEW", amount: 0, want: TransactionTypeDebit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := classifyTransactionType(tt.wiseType, tt.amount)
			if got != tt.want {
				t.Errorf("classifyTransactionType(%q, %v) = %v, want %v",
					tt.wiseType, tt.amount, got, tt.want)
			}
		})
	}
}

func TestBalanceAmountCents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value float64
		want  int64
	}{
		{name: "zero", value: 0, want: 0},
		{name: "exact", value: 1234.56, want: 123456},
		{name: "negative", value: -50.00, want: -5000},
		{name: "float error rounded", value: 0.1 + 0.2, want: 30},
		{name: "rounds half up", value: 12.345, want: 1235},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := raw.BalanceAmount{Value: tt.value}.Cents()
			if got != tt.want {
				t.Errorf("Cents() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestToMoneyInvalidCurrency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		currency string
	}{
		{name: "empty currency", currency: ""},
		{name: "too short", currency: "EU"},
		{name: "too long", currency: "EURO"},
		{name: "lowercase", currency: "eur"},
		{name: "digits", currency: "EU1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := toMoney(raw.BalanceAmount{Value: 100, Currency: tt.currency})
			if err == nil {
				t.Fatalf("toMoney with currency %q expected error, got nil", tt.currency)
			}
		})
	}
}

func TestToMoneyValid(t *testing.T) {
	t.Parallel()

	got, err := toMoney(raw.BalanceAmount{Value: 12.34, Currency: "EUR"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Cents != 1234 {
		t.Errorf("Cents = %d, want 1234", got.Cents)
	}

	if got.Currency != Currency("EUR") {
		t.Errorf("Currency = %q, want %q", got.Currency, Currency("EUR"))
	}
}

func TestMapExchangeNil(t *testing.T) {
	t.Parallel()

	got, _ := mapExchange(nil)
	if got != nil {
		t.Errorf("mapExchange(nil) = %v, want nil", got)
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

func TestParseBalanceTypeError(t *testing.T) {
	t.Parallel()

	_, err := parseBalanceType("CRYPTO")
	if err == nil {
		t.Fatal("expected error for unknown balance type")
	}
}

func TestNewTransactionID(t *testing.T) {
	t.Parallel()

	id := NewTransactionID("tx-abc")
	if id.Get() != "tx-abc" {
		t.Errorf("Get() = %q, want %q", id.Get(), "tx-abc")
	}
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	custom := &http.Client{Timeout: 7 * time.Second}

	c := New("key", WithHTTPClient(custom))
	if c.httpClient != custom {
		t.Error("WithHTTPClient did not set the custom client")
	}
}

func TestParseWiseTimestamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "RFC3339 with Z suffix",
			input: "2023-01-15T10:30:00Z",
			want:  time.Date(2023, time.January, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:  "RFC3339 with offset",
			input: "2023-01-15T10:30:00+02:00",
			want:  time.Date(2023, time.January, 15, 8, 30, 0, 0, time.UTC),
		},
		{
			name:  "RFC3339 with fractional seconds",
			input: "2023-01-15T10:30:00.123Z",
			want:  time.Date(2023, time.January, 15, 10, 30, 0, 123000000, time.UTC),
		},
		{
			name: "zoneless with T separator (live /v2/profiles format)",
			// Exact value observed from the live Wise API on 2026-08-18.
			input: "2020-05-27T10:27:22",
			want:  time.Date(2020, time.May, 27, 10, 27, 22, 0, time.UTC),
		},
		{
			name:  "zoneless with space separator (statement date format)",
			input: "2023-01-15 14:30:00",
			want:  time.Date(2023, time.January, 15, 14, 30, 0, 0, time.UTC),
		},
		{
			name:  "milliseconds with numeric zone (delivery estimate format)",
			input: "2018-01-10T12:15:00.000+0000",
			want:  time.Date(2018, time.January, 10, 12, 15, 0, 0, time.UTC),
		},
	}

	assertTimestampCases(t, tests)
}

func TestFormatWiseTimestampNormalizesToUTC(t *testing.T) {
	t.Parallel()

	cest := time.FixedZone("CEST", 2*60*60)

	tests := []struct {
		name  string
		input time.Time
		want  string
	}{
		{
			name:  "UTC input",
			input: time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
			want:  "2023-12-31T23:59:59Z",
		},
		{
			name:  "local zone input (live regression: time.Now in +02:00)",
			input: time.Date(2024, 1, 1, 1, 59, 59, 0, cest),
			want:  "2023-12-31T23:59:59Z",
		},
		{
			name:  "fractional seconds dropped",
			input: time.Date(2023, 12, 31, 23, 59, 59, 500000000, time.UTC),
			want:  "2023-12-31T23:59:59Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := formatWiseTimestamp(tt.input); got != tt.want {
				t.Errorf("formatWiseTimestamp(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseWiseTimestampRejectsMalformed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "date only", input: "2023-01-15"},
		{name: "garbage", input: "not-a-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := parseWiseTimestamp(tt.input); err == nil {
				t.Fatalf("parseWiseTimestamp(%q) succeeded, want error", tt.input)
			}
		})
	}
}

func assertTimestampCases(t *testing.T, tests []struct {
	name    string
	input   string
	want    time.Time
	wantErr bool
},
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseWiseTimestamp(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseWiseTimestamp(%q) succeeded, want error", tt.input)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseWiseTimestamp(%q) failed: %v", tt.input, err)
			}

			if !got.Equal(tt.want) {
				t.Errorf("parseWiseTimestamp(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMapBalanceError(t *testing.T) {
	t.Parallel()

	_, err := mapBalance(raw.Balance{CreationTime: "bad", Type: "STANDARD"})
	if err == nil {
		t.Fatal("expected error for bad creation time")
	}
}

func TestMapProfileError(t *testing.T) {
	t.Parallel()

	_, err := mapProfile(raw.Profile{CreatedAt: "bad", Type: "PERSONAL"})
	if err == nil {
		t.Fatal("expected error for bad created_at")
	}
}

func TestNewCurrency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    Currency
		wantErr bool
	}{
		{name: "valid EUR", input: "EUR", want: Currency("EUR")},
		{name: "valid USD", input: "USD", want: Currency("USD")},
		{name: "valid GBP", input: "GBP", want: Currency("GBP")},
		{name: "empty rejected", input: "", wantErr: true},
		{name: "two letters rejected", input: "EU", wantErr: true},
		{name: "four letters rejected", input: "EURO", wantErr: true},
		{name: "lowercase rejected", input: "eur", wantErr: true},
		{name: "mixed case rejected", input: "Eur", wantErr: true},
		{name: "digits rejected", input: "EU1", wantErr: true},
		{name: "special chars rejected", input: "E-R", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewCurrency(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NewCurrency(%q) expected error, got nil", tt.input)
				}

				return
			}

			if err != nil {
				t.Fatalf("NewCurrency(%q) unexpected error: %v", tt.input, err)
			}

			if got != tt.want {
				t.Errorf("NewCurrency(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMoneyString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		money Money
		want  string
	}{
		{name: "zero", money: Money{Cents: 0, Currency: Currency("EUR")}, want: "EUR 0.00"},
		{name: "positive round", money: Money{Cents: 1234, Currency: Currency("USD")}, want: "USD 12.34"},
		{name: "positive with cents", money: Money{Cents: 12345, Currency: Currency("EUR")}, want: "EUR 123.45"},
		{name: "negative", money: Money{Cents: -5000, Currency: Currency("GBP")}, want: "GBP -50.00"},
		{name: "negative with cents", money: Money{Cents: -12345, Currency: Currency("EUR")}, want: "EUR -123.45"},
		{name: "single cent", money: Money{Cents: 1, Currency: Currency("USD")}, want: "USD 0.01"},
		{name: "large amount", money: Money{Cents: 999999999, Currency: Currency("EUR")}, want: "EUR 9999999.99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.money.String(); got != tt.want {
				t.Errorf("Money.String() = %q, want %q", got, tt.want)
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
