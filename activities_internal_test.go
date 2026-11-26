package wise

import (
	"testing"
)

// TestParseActivityAmount pins the formatted-amount reader against the
// shapes Wise's feed is known to send and against the shapes it must NOT
// read as an amount.
func TestParseActivityAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		text    string
		present bool
		credit  bool
		cents   int64
	}{
		{"plain amount", "150 JPY", true, false, 15000},
		{"decimals", "1.50 SGD", true, false, 150},
		{"one decimal is tenths", "9.9 EUR", true, false, 990},
		{"comma grouping", "12,345.67 PLN", true, false, 1234567},
		{"two groups", "1,234,567.89 USD", true, false, 123456789},
		{"five cents", "0.05 GBP", true, false, 5},
		{"credit markup", "<positive>+ 9,999.99 USD</positive>", true, true, 999999},
		{"credit markup without sign", "<positive>100 EUR</positive>", true, true, 10000},
		{"leading plus", "+ 9 EUR", true, true, 900},
		{"leading minus is a debit", "- 9.99 USD", true, false, 999},
		{"empty", "", false, false, 0},
		{"currency pair", "USD, EUR", false, false, 0},
		{"german locale is not read as a thousandth", "1.234,56 EUR", false, false, 0},
		{"three decimals", "1.234 EUR", false, false, 0},
		{"malformed grouping", "12,34 EUR", false, false, 0},
		{"unrepresentable", "1234567890123456.00 EUR", false, false, 0},
		{"lowercase currency", "150 jpy", false, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := parseActivityAmount(tt.text)

			if got.Present != tt.present {
				t.Fatalf("parseActivityAmount(%q).Present = %t, want %t", tt.text, got.Present, tt.present)
			}

			if got.Credit != tt.credit {
				t.Errorf("parseActivityAmount(%q).Credit = %t, want %t", tt.text, got.Credit, tt.credit)
			}

			if got.Present && got.Amount.Cents != tt.cents {
				t.Errorf("parseActivityAmount(%q).Amount.Cents = %d, want %d", tt.text, got.Amount.Cents, tt.cents)
			}

			if got.Text != tt.text {
				t.Errorf("parseActivityAmount(%q).Text = %q, want the raw text kept", tt.text, got.Text)
			}
		})
	}
}
