package wise

import (
	"errors"
	"net/http"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
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
		wiseType string
		amount   float64
		want     TransactionType
	}{
		{name: "card payment debit", wiseType: "CARD_PAYMENT", amount: -10, want: TransactionTypeCard},
		{
			name:     "card payment credit still card (not refund)",
			wiseType: "CARD_PAYMENT",
			amount:   25,
			want:     TransactionTypeCard,
		},
		{name: "card payment zero still card", wiseType: "CARD_PAYMENT", amount: 0, want: TransactionTypeCard},
		{name: "card refund positive", wiseType: "CARD_REFUND", amount: 25, want: TransactionTypeRefund},
		{name: "card refund zero", wiseType: "CARD_REFUND", amount: 0, want: TransactionTypeCard},
		{name: "transfer", wiseType: "TRANSFER", amount: 100, want: TransactionTypeTransfer},
		{name: "payment", wiseType: "PAYMENT", amount: -50, want: TransactionTypePayment},
		{name: "conversion", wiseType: "CONVERSION", amount: -100, want: TransactionTypeExchange},
		{name: "exchange alias", wiseType: "EXCHANGE", amount: 50, want: TransactionTypeExchange},
		{name: "fee", wiseType: "FEE", amount: -0.5, want: TransactionTypeFee},
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

			got := BalanceAmount{Value: tt.value}.Cents()
			if got != tt.want {
				t.Errorf("Cents() = %d, want %d", got, tt.want)
			}
		})
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

	err := newAPIError(statusCode, body, time.Second)

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

	err := newAPIError(http.StatusTooManyRequests, "{}", 42*time.Second)

	rle, ok := errors.AsType[*RateLimitError](err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}

	if rle.RetryAfter != 42*time.Second {
		t.Errorf("RetryAfter = %v, want 42s", rle.RetryAfter)
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

func TestParseWiseDateError(t *testing.T) {
	t.Parallel()

	_, err := parseWiseDate("not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestMapBalanceError(t *testing.T) {
	t.Parallel()

	_, err := mapBalance(Balance{CreationTime: "bad", Type: "STANDARD"})
	if err == nil {
		t.Fatal("expected error for bad creation time")
	}
}

func TestMapProfileError(t *testing.T) {
	t.Parallel()

	_, err := mapProfile(Profile{CreatedAt: "bad", Type: "PERSONAL"})
	if err == nil {
		t.Fatal("expected error for bad created_at")
	}
}
