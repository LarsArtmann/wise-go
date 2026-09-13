package wise

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/wise-go/internal/raw"
)

func TestClassifyTransactionType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		wiseType   DetailType
		totalCents int64
		want       TransactionType
	}{
		{name: "card payment debit", wiseType: DetailTypeCardPayment, totalCents: -1000, want: TransactionTypeCard},
		{
			name:       "card payment credit still card (not refund)",
			wiseType:   DetailTypeCardPayment,
			totalCents: 2500,
			want:       TransactionTypeCard,
		},
		{
			name:       "card payment zero still card",
			wiseType:   DetailTypeCardPayment,
			totalCents: 0,
			want:       TransactionTypeCard,
		},
		{name: "card refund positive", wiseType: DetailTypeCardRefund, totalCents: 2500, want: TransactionTypeRefund},
		{name: "card refund zero", wiseType: DetailTypeCardRefund, totalCents: 0, want: TransactionTypeCard},
		{name: "transfer", wiseType: DetailTypeTransfer, totalCents: 10000, want: TransactionTypeTransfer},
		{name: "payment", wiseType: DetailTypePayment, totalCents: -5000, want: TransactionTypePayment},
		{name: "conversion", wiseType: DetailTypeConversion, totalCents: -10000, want: TransactionTypeExchange},
		{name: "exchange alias", wiseType: DetailTypeExchange, totalCents: 5000, want: TransactionTypeExchange},
		{name: "fee", wiseType: DetailTypeFee, totalCents: -50, want: TransactionTypeFee},
		{name: "unknown positive is credit", wiseType: "SOMETHING_NEW", totalCents: 1000, want: TransactionTypeCredit},
		{name: "unknown negative is debit", wiseType: "SOMETHING_NEW", totalCents: -1000, want: TransactionTypeDebit},
		{name: "unknown zero is debit", wiseType: "SOMETHING_NEW", totalCents: 0, want: TransactionTypeDebit},
		{
			name:       "sub-cent amount is debit (cents floor)",
			wiseType:   "SOMETHING_NEW",
			totalCents: 0,
			want:       TransactionTypeDebit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := classifyTransactionType(tt.wiseType, tt.totalCents)
			if got != tt.want {
				t.Errorf("classifyTransactionType(%q, %d) = %v, want %v",
					tt.wiseType, tt.totalCents, got, tt.want)
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

func TestParseBalanceTypeError(t *testing.T) {
	t.Parallel()

	_, err := parseBalanceType("CRYPTO")
	if err == nil {
		t.Fatal("expected error for unknown balance type")
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

// FuzzParseWiseTimestamp pins the tolerant parser's invariants: it never
// panics, accepts every layout it documents (an RFC3339 input must parse,
// and a zoneless-layout input must come back as UTC).
func FuzzParseWiseTimestamp(f *testing.F) {
	f.Add("2020-05-27T10:27:22Z")
	f.Add("2018-01-10T12:15:00.000+0000")
	f.Add("2020-05-27T10:27:22")
	f.Add("2020-05-27 10:27:22")
	f.Add("")
	f.Add("garbage")

	f.Fuzz(func(t *testing.T, s string) {
		got, err := parseWiseTimestamp(s)
		if err != nil {
			return // unparseable input is fine; panics are not
		}

		// RFC3339 input must round-trip.
		if want, wantErr := time.Parse(time.RFC3339, s); wantErr == nil && !got.Equal(want) {
			t.Fatalf("parseWiseTimestamp(%q) = %v, want RFC3339 round-trip %v", s, got, want)
		}

		// Zoneless layout input must yield UTC.
		if _, wantErr := time.Parse("2006-01-02T15:04:05", s); wantErr == nil {
			if got.Location() != time.UTC {
				t.Fatalf("parseWiseTimestamp(%q) zoneless layout must be UTC, got %v", s, got.Location())
			}
		}
	})
}

// FuzzNewCurrency pins the currency validator's contract: it accepts exactly
// three uppercase ASCII letters and nothing else.
func FuzzNewCurrency(f *testing.F) {
	f.Add("EUR")
	f.Add("usd")
	f.Add("EURO")
	f.Add("E1R")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		got, err := NewCurrency(s)
		if err != nil {
			if got != "" {
				t.Fatalf("NewCurrency(%q) returned %q alongside an error", s, got)
			}

			return
		}

		if len(s) != 3 {
			t.Fatalf("NewCurrency(%q) accepted a non-3-letter code", s)
		}

		for _, c := range s {
			if c < 'A' || c > 'Z' {
				t.Fatalf("NewCurrency(%q) accepted a non-uppercase-ASCII letter", s)
			}
		}

		if string(got) != s {
			t.Fatalf("NewCurrency(%q) = %q, want the input verbatim", s, got)
		}
	})
}
