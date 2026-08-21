package wise

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
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

// expectRejection asserts err is non-nil, contains wantSubstr, and classifies
// as a Rejection (the client-side validation contract: fail fast, never
// retryable, no network round-trip).
func expectRejection(t *testing.T, err error, wantSubstr string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), wantSubstr) {
		t.Errorf("error %q does not contain %q", err.Error(), wantSubstr)
	}

	if family := errorfamily.Classify(err); family != errorfamily.Rejection {
		t.Errorf("Classify() = %v, want Rejection (error: %v)", family, err)
	}
}

func TestCreateTransferRequestValidate(t *testing.T) {
	t.Parallel()

	const (
		quoteID   = "11144c35-9fe8-4c32-b7fd-d05c2a7734bf"
		txID      = "22244c35-9fe8-4c32-b7fd-d05c2a7734bf"
		accountID = int64(98765432)
	)

	valid := CreateTransferRequest{
		QuoteID:               NewQuoteID(quoteID),
		TargetAccount:         NewRecipientID(accountID),
		CustomerTransactionID: txID,
	}

	tests := []struct {
		name      string
		mutate    func(*CreateTransferRequest)
		wantSubst string
	}{
		{
			name:      "missing quoteID",
			mutate:    func(r *CreateTransferRequest) { r.QuoteID = QuoteID{} },
			wantSubst: "quoteID is required",
		},
		{
			name:      "missing targetAccount",
			mutate:    func(r *CreateTransferRequest) { r.TargetAccount = RecipientID{} },
			wantSubst: "targetAccount is required",
		},
		{
			name:      "missing customerTransactionId",
			mutate:    func(r *CreateTransferRequest) { r.CustomerTransactionID = "" },
			wantSubst: "customerTransactionId is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := valid
			tt.mutate(&req)
			expectRejection(t, req.validate(), tt.wantSubst)
		})
	}

	t.Run("complete request passes", func(t *testing.T) {
		t.Parallel()

		if err := valid.validate(); err != nil {
			t.Errorf("validate() = %v, want nil", err)
		}
	})
}

type quoteValidateCase struct {
	name      string
	mutate    func(*CreateQuoteRequest)
	wantSubst string
}

func quoteValidateCases() []quoteValidateCase {
	return []quoteValidateCase{
		{
			name:      "missing sourceCurrency",
			mutate:    func(r *CreateQuoteRequest) { r.SourceCurrency = "" },
			wantSubst: "sourceCurrency is required",
		},
		{
			name:      "missing targetCurrency",
			mutate:    func(r *CreateQuoteRequest) { r.TargetCurrency = "" },
			wantSubst: "targetCurrency is required",
		},
		{
			name:      "same currencies",
			mutate:    func(r *CreateQuoteRequest) { r.TargetCurrency = Currency("EUR") },
			wantSubst: "must be different",
		},
		{
			name:      "no amount set",
			mutate:    func(r *CreateQuoteRequest) { r.SourceAmount = nil },
			wantSubst: "either sourceAmount or targetAmount is required",
		},
		{
			name: "both amounts set",
			mutate: func(r *CreateQuoteRequest) {
				r.TargetAmount = &Money{Cents: 1086, Currency: Currency("USD")}
			},
			wantSubst: "only one of sourceAmount or targetAmount",
		},
		{
			name: "sourceAmount currency mismatch",
			mutate: func(r *CreateQuoteRequest) {
				r.SourceAmount = &Money{Cents: 1000, Currency: Currency("GBP")}
			},
			wantSubst: "sourceAmount currency must match sourceCurrency",
		},
		{
			name: "targetAmount currency mismatch",
			mutate: func(r *CreateQuoteRequest) {
				r.SourceAmount = nil
				r.TargetAmount = &Money{Cents: 1086, Currency: Currency("EUR")}
			},
			wantSubst: "targetAmount currency must match targetCurrency",
		},
	}
}

func TestCreateQuoteRequestValidate(t *testing.T) {
	t.Parallel()

	base := func() CreateQuoteRequest {
		return CreateQuoteRequest{
			SourceCurrency: Currency("EUR"),
			TargetCurrency: Currency("USD"),
			SourceAmount:   &Money{Cents: 1000, Currency: Currency("EUR")},
		}
	}

	for _, tt := range quoteValidateCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := base()
			tt.mutate(&req)
			expectRejection(t, req.validate(), tt.wantSubst)
		})
	}
}

func TestCreateQuoteRequestValidateAccepts(t *testing.T) {
	t.Parallel()

	t.Run("target-amount quote passes", func(t *testing.T) {
		t.Parallel()

		req := CreateQuoteRequest{
			SourceCurrency: Currency("EUR"),
			TargetCurrency: Currency("USD"),
			TargetAmount:   &Money{Cents: 1086, Currency: Currency("USD")},
		}

		if err := req.validate(); err != nil {
			t.Errorf("validate() = %v, want nil", err)
		}
	})

	t.Run("authenticated quote requires profileID", func(t *testing.T) {
		t.Parallel()

		req := CreateQuoteRequest{
			SourceCurrency: Currency("EUR"),
			TargetCurrency: Currency("USD"),
			SourceAmount:   &Money{Cents: 1000, Currency: Currency("EUR")},
		}

		expectRejection(t, req.validateAuthenticated(ProfileID{}), "profileID is required")
	})
}

func TestValidateTransferRequirementsRequestValidate(t *testing.T) {
	t.Parallel()

	valid := ValidateTransferRequirementsRequest{
		TargetAccount: NewRecipientID(98765432),
		QuoteID:       NewQuoteID("11144c35-9fe8-4c32-b7fd-d05c2a7734bf"),
	}

	tests := []struct {
		name      string
		mutate    func(*ValidateTransferRequirementsRequest)
		wantSubst string
	}{
		{
			name:      "missing targetAccount",
			mutate:    func(r *ValidateTransferRequirementsRequest) { r.TargetAccount = RecipientID{} },
			wantSubst: "targetAccount is required",
		},
		{
			name:      "missing quoteID",
			mutate:    func(r *ValidateTransferRequirementsRequest) { r.QuoteID = QuoteID{} },
			wantSubst: "quoteUuid is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := valid
			tt.mutate(&req)
			expectRejection(t, req.validate(), tt.wantSubst)
		})
	}

	t.Run("complete request passes", func(t *testing.T) {
		t.Parallel()

		if err := valid.validate(); err != nil {
			t.Errorf("validate() = %v, want nil", err)
		}
	})
}

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

// requirementField builds a single-field form requirement for tests.
func requirementField(
	key, name string,
	required bool,
	valuesAllowed ...TransferRequirementValue,
) TransferRequirement {
	return TransferRequirement{
		Type: "transfer",
		Fields: []TransferRequirementForm{{
			Name: name,
			Group: []TransferRequirementField{{
				Key:           key,
				Name:          name,
				Type:          "text",
				Required:      required,
				ValuesAllowed: valuesAllowed,
			}},
		}},
	}
}

type missingDetailsCase struct {
	name string
	reqs []TransferRequirement
	req  CreateTransferRequest
	want []string
}

func missingDetailsCases(purposeValues []TransferRequirementValue) []missingDetailsCase {
	return []missingDetailsCase{
		{
			name: "no requirements means nothing missing",
			reqs: nil,
			req:  CreateTransferRequest{},
			want: []string{},
		},
		{
			name: "optional fields are not reported",
			reqs: []TransferRequirement{
				requirementField("reference", "Transfer reference", false),
			},
			req:  CreateTransferRequest{},
			want: []string{},
		},
		{
			name: "required modeled field left empty is reported",
			reqs: []TransferRequirement{
				requirementField("reference", "Transfer reference", true),
			},
			req:  CreateTransferRequest{},
			want: []string{"Transfer reference"},
		},
		{
			name: "required modeled field set is satisfied",
			reqs: []TransferRequirement{
				requirementField("reference", "Transfer reference", true),
			},
			req:  CreateTransferRequest{Reference: "Invoice 2026-001"},
			want: []string{},
		},
		{
			name: "select field with disallowed value is reported",
			reqs: []TransferRequirement{
				requirementField("transferPurpose", "Transfer purpose", true, purposeValues...),
			},
			req:  CreateTransferRequest{TransferPurpose: "made.up.purpose"},
			want: []string{"Transfer purpose"},
		},
		{
			name: "select field with allowed value is satisfied",
			reqs: []TransferRequirement{
				requirementField("transferPurpose", "Transfer purpose", true, purposeValues...),
			},
			req:  CreateTransferRequest{TransferPurpose: "verification.transfers.purpose.other"},
			want: []string{},
		},
		{
			name: "unmodeled required key is always reported",
			reqs: []TransferRequirement{
				requirementField("legalEntityIdentifier", "Legal entity identifier", true),
			},
			req:  CreateTransferRequest{},
			want: []string{"Legal entity identifier"},
		},
	}
}

func TestMissingTransferDetails(t *testing.T) {
	t.Parallel()

	purposeValues := []TransferRequirementValue{
		{Key: "verification.transfers.purpose.pay.bills", Name: "Rent or other property expenses"},
		{Key: "verification.transfers.purpose.other", Name: "Other"},
	}

	for _, tt := range missingDetailsCases(purposeValues) {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := MissingTransferDetails(tt.reqs, tt.req)
			if len(got) != len(tt.want) {
				t.Fatalf("MissingTransferDetails() = %v, want %v", got, tt.want)
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("MissingTransferDetails()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestMissingTransferDetailsTwoPassFlow mirrors the documented
// RefreshRequirementsOnChange pattern: populating the refresh-triggering
// select reveals a lower-level required field on the second
// ValidateTransferRequirements round-trip, and MissingTransferDetails catches
// it before CreateTransfer spends the customerTransactionId.
func TestMissingTransferDetailsTwoPassFlow(t *testing.T) {
	t.Parallel()

	purposeValues := []TransferRequirementValue{
		{Key: "verification.transfers.purpose.intercompany", Name: "Intercompany payment"},
	}

	firstPass := make([]TransferRequirement, 0, 2)
	firstPass = append(firstPass,
		requirementField("transferPurpose", "Transfer purpose", true, purposeValues...),
	)

	firstPass[0].Fields[0].Group[0].RefreshRequirementsOnChange = true
	firstPass[0].Fields[0].Group[0].Type = "select"

	req := CreateTransferRequest{}

	if missing := MissingTransferDetails(firstPass, req); len(missing) != 1 {
		t.Fatalf("first pass: MissingTransferDetails() = %v, want [Transfer purpose]", missing)
	}

	req.TransferPurpose = "verification.transfers.purpose.intercompany"

	if missing := MissingTransferDetails(firstPass, req); len(missing) != 0 {
		t.Fatalf("after purpose set: MissingTransferDetails() = %v, want none", missing)
	}

	secondPass := make([]TransferRequirement, 0, len(firstPass)+1)
	secondPass = append(secondPass, firstPass...)
	secondPass = append(secondPass,
		requirementField("transferPurposeInvoiceNumber", "Invoice number", true),
	)

	missing := MissingTransferDetails(secondPass, req)
	if len(missing) != 1 || missing[0] != "Invoice number" {
		t.Errorf("second pass: MissingTransferDetails() = %v, want [Invoice number]", missing)
	}

	req.TransferPurposeInvoiceNumber = "INV-2026-001"

	if missing := MissingTransferDetails(secondPass, req); len(missing) != 0 {
		t.Errorf("after invoice set: MissingTransferDetails() = %v, want none", missing)
	}
}

// webhook verification tests use a locally generated RSA keypair to prove
// publicKeyPEM renders an RSA public key as a PKIX PEM block.
func publicKeyPEM(key *rsa.PublicKey) ([]byte, error) {
	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshal PKIX public key: %w", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}), nil
}

// webhook verification tests use a locally generated RSA keypair to prove
// webhookFixture is a payload signed with a locally generated RSA keypair, so
// accept/reject behavior is proven without Wise's real key.
type webhookFixture struct {
	payload   []byte
	validSig  string
	key       *rsa.PublicKey
	sourceKey *rsa.PrivateKey
	wrongKey  *rsa.PublicKey
}

func newWebhookFixture(t *testing.T) webhookFixture {
	t.Helper()

	payload := []byte(`{"event":{"type":"transfers#state-change","data":{"resource":{"id":16521632}}}}`)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	digest := sha256.Sum256(payload)

	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign payload: %v", err)
	}

	pemKey, err := publicKeyPEM(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}

	parsed, err := ParseWebhookPublicKey(pemKey)
	if err != nil {
		t.Fatalf("ParseWebhookPublicKey: %v", err)
	}

	wrongKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate wrong key: %v", err)
	}

	return webhookFixture{
		payload:   payload,
		validSig:  base64.StdEncoding.EncodeToString(sig),
		key:       parsed,
		sourceKey: key,
		wrongKey:  &wrongKey.PublicKey,
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	t.Parallel()

	fixture := newWebhookFixture(t)

	t.Run("valid signature verifies", func(t *testing.T) {
		t.Parallel()

		if !VerifyWebhookSignature(fixture.payload, fixture.validSig, fixture.key) {
			t.Error("VerifyWebhookSignature(valid) = false, want true")
		}
	})

	t.Run("tampered payload is rejected", func(t *testing.T) {
		t.Parallel()

		tampered := append([]byte{}, fixture.payload...)
		tampered[len(tampered)-3] = '9'

		if VerifyWebhookSignature(tampered, fixture.validSig, fixture.key) {
			t.Error("VerifyWebhookSignature(tampered) = true, want false")
		}
	})

	t.Run("wrong public key is rejected", func(t *testing.T) {
		t.Parallel()

		if VerifyWebhookSignature(fixture.payload, fixture.validSig, fixture.wrongKey) {
			t.Error("VerifyWebhookSignature(wrong key) = true, want false")
		}
	})

	t.Run("malformed signature is rejected", func(t *testing.T) {
		t.Parallel()

		if VerifyWebhookSignature(fixture.payload, "not-base64!!!", fixture.key) {
			t.Error("VerifyWebhookSignature(malformed) = true, want false")
		}
	})

	t.Run("empty signature and nil key are rejected", func(t *testing.T) {
		t.Parallel()

		if VerifyWebhookSignature(fixture.payload, "", fixture.key) {
			t.Error("VerifyWebhookSignature(empty sig) = true, want false")
		}

		if VerifyWebhookSignature(fixture.payload, fixture.validSig, nil) {
			t.Error("VerifyWebhookSignature(nil key) = true, want false")
		}
	})
}

func TestParseWebhookPublicKey(t *testing.T) {
	t.Parallel()

	fixture := newWebhookFixture(t)

	t.Run("PKCS#1 PEM keys are also accepted", func(t *testing.T) {
		t.Parallel()

		pkcs1 := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&fixture.sourceKey.PublicKey),
		})

		parsed1, err := ParseWebhookPublicKey(pkcs1)
		if err != nil {
			t.Fatalf("ParseWebhookPublicKey(PKCS1): %v", err)
		}

		if !VerifyWebhookSignature(fixture.payload, fixture.validSig, parsed1) {
			t.Error("VerifyWebhookSignature(PKCS1 key) = false, want true")
		}
	})

	t.Run("non-PEM input is a config error", func(t *testing.T) {
		t.Parallel()

		if _, err := ParseWebhookPublicKey([]byte("garbage")); err == nil {
			t.Error("ParseWebhookPublicKey(garbage) = nil error, want error")
		}
	})
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

// TestParseWiseDate exercises the date-only parser directly: Wise emits
// date-only strings with no timezone; empty means unset (zero time, not an
// error) and everything else must parse as a UTC-midnight time.
func TestParseWiseDate(t *testing.T) {
	t.Parallel()

	t.Run("empty string maps to the zero time", func(t *testing.T) {
		t.Parallel()

		got, err := parseWiseDate("")
		if err != nil {
			t.Fatalf("parseWiseDate(\"\") error: %v", err)
		}

		if !got.IsZero() {
			t.Errorf("parseWiseDate(\"\") = %v, want zero time", got)
		}
	})

	t.Run("valid date parses as UTC midnight", func(t *testing.T) {
		t.Parallel()

		got, err := parseWiseDate("1977-01-31")
		if err != nil {
			t.Fatalf("parseWiseDate error: %v", err)
		}

		want := time.Date(1977, time.January, 31, 0, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("parseWiseDate = %v, want %v", got, want)
		}
	})

	t.Run("garbage input errors", func(t *testing.T) {
		t.Parallel()

		_, err := parseWiseDate("not-a-date")
		if err == nil {
			t.Fatal("parseWiseDate(garbage) = nil error, want error")
		}

		if !strings.Contains(err.Error(), "not-a-date") {
			t.Errorf("error %q does not quote the input", err.Error())
		}
	})
}

// errTestPlain and friends are static sentinel errors for tests (err113:
// never construct dynamic errors).
var (
	errTestPlain      = errors.New("plain")
	errTestBoom       = errors.New("boom")
	errTestRateLimits = errors.New("rate limited")
)

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

// TestGetRawEdges covers the raw-response path's non-JSON arms: a transport
// failure and an empty body (a legitimately empty statement file).
func TestGetRawEdges(t *testing.T) {
	t.Parallel()

	t.Run("transport error surfaces", func(t *testing.T) {
		t.Parallel()

		closed := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		closed.Close()

		client := New("test-api-key", WithBaseURL(closed.URL))

		_, err := client.getRaw(context.Background(), "/x", nil)
		if err == nil {
			t.Fatal("getRaw against closed server = nil error, want transport error")
		}
	})

	t.Run("empty body is a valid result", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := New("test-api-key", WithBaseURL(server.URL))

		data, err := client.getRaw(context.Background(), "/statement.csv", func() string {
			return "currency=EUR"
		})
		if err != nil {
			t.Fatalf("getRaw empty body error: %v", err)
		}

		if len(data) != 0 {
			t.Errorf("getRaw empty body = %q, want empty", data)
		}
	})
}

// TestExecuteWithLoggingTransportError asserts the transport-error arm: the
// logger sees Status 0 and a non-nil Error, and the error is wrapped with the
// method and URL.
func TestExecuteWithLoggingTransportError(t *testing.T) {
	t.Parallel()

	closed := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	closed.Close()

	var entries []RequestLog

	client := New("test-api-key",
		WithBaseURL(closed.URL),
		WithLogger(RequestLogFunc(func(entry RequestLog) {
			entries = append(entries, entry)
		})),
	)

	_, err := client.ListCurrencies(context.Background())
	if err == nil {
		t.Fatal("ListCurrencies against closed server = nil error, want transport error")
	}

	if !strings.Contains(err.Error(), "/v1/currencies") {
		t.Errorf("error %q does not mention the URL", err.Error())
	}

	if len(entries) == 0 {
		t.Fatal("logger recorded no entries")
	}

	last := entries[len(entries)-1]
	if last.Status != 0 {
		t.Errorf("logged Status = %d, want 0 on transport failure", last.Status)
	}

	if last.Error == nil {
		t.Error("logged Error = nil, want the transport error")
	}

	if last.URL == "" || last.Method == "" {
		t.Errorf("logged entry missing method/URL: %+v", last)
	}
}

// TestVerifyWebhookSignatureEdges covers payload extremes: an empty body and
// a multi-megabyte body both verify exactly like normal ones (the signature
// is over the raw bytes, whatever they are).
func TestVerifyWebhookSignatureEdges(t *testing.T) {
	t.Parallel()

	sign := func(t *testing.T, key *rsa.PrivateKey, payload []byte) string {
		t.Helper()

		digest := sha256.Sum256(payload)

		sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
		if err != nil {
			t.Fatalf("sign: %v", err)
		}

		return base64.StdEncoding.EncodeToString(sig)
	}

	fixture := newWebhookFixture(t)

	t.Run("empty payload verifies against its own signature", func(t *testing.T) {
		t.Parallel()

		emptySig := sign(t, fixture.sourceKey, []byte{})
		if !VerifyWebhookSignature([]byte{}, emptySig, fixture.key) {
			t.Error("VerifyWebhookSignature(empty) = false, want true")
		}

		if VerifyWebhookSignature([]byte("x"), emptySig, fixture.key) {
			t.Error("empty signature must not verify different bytes")
		}
	})

	t.Run("multi-megabyte payload verifies", func(t *testing.T) {
		t.Parallel()

		huge := make([]byte, 0, 5<<20) // 5 MiB capacity, filled below (makezero)

		huge = append(huge, make([]byte, 5<<20)...)
		if _, err := rand.Read(huge); err != nil {
			t.Fatalf("rand: %v", err)
		}

		hugeSig := sign(t, fixture.sourceKey, huge)
		if !VerifyWebhookSignature(huge, hugeSig, fixture.key) {
			t.Error("VerifyWebhookSignature(5MiB) = false, want true")
		}

		huge[len(huge)-1] ^= 0xFF // flip the final byte
		if VerifyWebhookSignature(huge, hugeSig, fixture.key) {
			t.Error("tampered 5MiB payload verified, want reject")
		}
	})
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
}

// TestCreateRecipientRequestValidate covers the full missing-field matrix of
// the recipient request validator.
func TestCreateRecipientRequestValidate(t *testing.T) {
	t.Parallel()

	valid := CreateRecipientRequest{
		ProfileID:         NewProfileID(12345),
		Currency:          Currency("GBP"),
		Type:              "sort_code",
		AccountHolderName: "Jane Doe",
		Details:           map[string]string{"sortCode": "040075"},
	}

	tests := []struct {
		name      string
		mutate    func(*CreateRecipientRequest)
		wantCause string
	}{
		{"profileID", func(r *CreateRecipientRequest) { r.ProfileID = ProfileID{} }, "profileID is required"},
		{"empty currency", func(r *CreateRecipientRequest) { r.Currency = "" }, "currency is required"},
		{
			"accountHolderName", func(r *CreateRecipientRequest) { r.AccountHolderName = "" },
			"accountHolderName is required",
		},
		{"route type", func(r *CreateRecipientRequest) { r.Type = "" }, "type is required"},
		{"details", func(r *CreateRecipientRequest) { r.Details = nil }, "details are required"},
	}

	for _, tt := range tests {
		t.Run("missing "+tt.name, func(t *testing.T) {
			t.Parallel()

			req := valid
			tt.mutate(&req)
			expectRejection(t, req.validate(), tt.wantCause)
		})
	}

	t.Run("valid request passes", func(t *testing.T) {
		t.Parallel()

		if err := valid.validate(); err != nil {
			t.Errorf("validate(valid) = %v, want nil", err)
		}
	})
}

// TestTransferDetailsWire pins the wire rendering of optional transfer
// details: every field set lands under its documented key, empty fields are
// omitted.
func TestTransferDetailsWire(t *testing.T) {
	t.Parallel()

	t.Run("empty request omits all keys", func(t *testing.T) {
		t.Parallel()

		if got := (CreateTransferRequest{}).detailsWire(); len(got) != 0 {
			t.Errorf("detailsWire(empty) = %v, want empty", got)
		}
	})

	t.Run("all fields render under their wire keys", func(t *testing.T) {
		t.Parallel()

		req := CreateTransferRequest{
			Reference:                         "invoice-42",
			SourceOfFunds:                     "salary",
			TransferPurpose:                   "verification",
			TransferPurposeInvoiceNumber:      "INV-42",
			TransferPurposeSubTransferPurpose: "other",
		}

		got := req.detailsWire()

		want := map[string]string{
			"reference":                         "invoice-42",
			"sourceOfFunds":                     "salary",
			"transferPurpose":                   "verification",
			"transferPurposeInvoiceNumber":      "INV-42",
			"transferPurposeSubTransferPurpose": "other",
		}
		for key, value := range want {
			if got[key] != value {
				t.Errorf("detailsWire()[%q] = %q, want %q", key, got[key], value)
			}
		}

		if len(got) != len(want) {
			t.Errorf("detailsWire() = %v, want exactly %d keys", got, len(want))
		}
	})

	t.Run("transferRequestDetailValue mirrors the wire keys", func(t *testing.T) {
		t.Parallel()

		req := CreateTransferRequest{Reference: "invoice-42"}

		if value, ok := transferRequestDetailValue(req, "reference"); !ok || value != "invoice-42" {
			t.Errorf("transferRequestDetailValue(reference) = %q,%v; want invoice-42,true", value, ok)
		}

		if _, ok := transferRequestDetailValue(req, "notModeled"); ok {
			t.Error("transferRequestDetailValue(unknown) = ok, want false")
		}
	})
}

// TestTransferRequirementsDetailsToWire pins the optional-block rendering of
// the transfer-requirements request.
func TestTransferRequirementsDetailsToWire(t *testing.T) {
	t.Parallel()

	t.Run("zero details render nothing", func(t *testing.T) {
		t.Parallel()

		if got := (TransferRequirementsDetails{}).toWire(); len(got) != 0 {
			t.Errorf("toWire(zero) = %v, want empty", got)
		}
	})

	t.Run("set fields render under their wire keys", func(t *testing.T) {
		t.Parallel()

		got := (TransferRequirementsDetails{
			Reference:          "ref-1",
			SourceOfFunds:      "salary",
			SourceOfFundsOther: "dividends",
		}).toWire()

		if got["reference"] != "ref-1" || got["sourceOfFunds"] != "salary" ||
			got["sourceOfFundsOther"] != "dividends" {
			t.Errorf("toWire() = %v, want the three set keys", got)
		}
	})
}

func TestRefreshQuoteAccountRequirementsRequestValidate(t *testing.T) {
	t.Parallel()

	valid := RefreshQuoteAccountRequirementsRequest{
		QuoteID: NewQuoteID("11144c35-9fe8-4c32-b7fd-d05c2a7734bf"),
		Recipient: CreateRecipientRequest{
			Currency: Currency("USD"),
			Type:     "swift_code",
			Details:  map[string]string{"legalEntityType": "PRIVATE"},
		},
	}

	tests := []struct {
		name      string
		mutate    func(*RefreshQuoteAccountRequirementsRequest)
		wantCause string
	}{
		{
			"quoteID", func(r *RefreshQuoteAccountRequirementsRequest) { r.QuoteID = QuoteID{} },
			"quoteID is required",
		},
		{
			"recipient currency",
			func(r *RefreshQuoteAccountRequirementsRequest) { r.Recipient.Currency = "" },
			"recipient currency is required",
		},
		{
			"recipient type", func(r *RefreshQuoteAccountRequirementsRequest) { r.Recipient.Type = "" },
			"recipient type is required",
		},
		{
			"recipient details",
			func(r *RefreshQuoteAccountRequirementsRequest) { r.Recipient.Details = nil },
			"recipient details are required",
		},
	}

	for _, tt := range tests {
		t.Run("missing "+tt.name, func(t *testing.T) {
			t.Parallel()

			req := valid
			tt.mutate(&req)
			expectRejection(t, req.validate(), tt.wantCause)
		})
	}

	t.Run("valid partial form passes without holder name or profile", func(t *testing.T) {
		t.Parallel()

		if err := valid.validate(); err != nil {
			t.Errorf("validate(valid partial form) = %v, want nil", err)
		}
	})
}

// TestRefreshQuoteAccountRequirementsToWire pins the omit-empties contract of
// the refresh wire body: an in-progress form must not claim an empty
// accountHolderName or a zero profile, but carries them once set.
func TestRefreshQuoteAccountRequirementsToWire(t *testing.T) {
	t.Parallel()

	partial := RefreshQuoteAccountRequirementsRequest{
		Recipient: CreateRecipientRequest{
			Currency: Currency("USD"),
			Type:     "swift_code",
			Details:  map[string]string{"legalEntityType": "PRIVATE"},
		},
	}

	t.Run("partial form omits unset fields", func(t *testing.T) {
		t.Parallel()

		got := partial.toWire()
		if got["currency"] != "USD" || got["type"] != "swift_code" {
			t.Errorf("toWire() = %v, want currency and type", got)
		}
		if _, ok := got["profile"]; ok {
			t.Errorf("toWire() sent profile for a zero ProfileID: %v", got)
		}
		if _, ok := got["accountHolderName"]; ok {
			t.Errorf("toWire() sent accountHolderName for an empty name: %v", got)
		}
		if _, ok := got["ownedByCustomer"]; ok {
			t.Errorf("toWire() sent ownedByCustomer when false: %v", got)
		}
	})

	t.Run("set fields render under their wire keys", func(t *testing.T) {
		t.Parallel()

		complete := partial
		complete.Recipient.ProfileID = NewProfileID(12345)
		complete.Recipient.AccountHolderName = "Jane Doe"
		complete.Recipient.OwnedByCustomer = true

		got := complete.toWire()
		if got["profile"] != int64(12345) {
			t.Errorf("toWire()[profile] = %v (%T), want int64 12345", got["profile"], got["profile"])
		}
		if got["accountHolderName"] != "Jane Doe" {
			t.Errorf("toWire()[accountHolderName] = %v, want Jane Doe", got["accountHolderName"])
		}
		if got["ownedByCustomer"] != true {
			t.Errorf("toWire()[ownedByCustomer] = %v, want true", got["ownedByCustomer"])
		}
	})
}
