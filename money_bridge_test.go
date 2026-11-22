package wise

import "testing"

func TestMoneyAmountConvertsToCanonical(t *testing.T) {
	t.Parallel()

	m := Money{Cents: 123456, Currency: Currency("EUR")}

	amount, err := m.Amount()
	if err != nil {
		t.Fatalf("Amount() error = %v", err)
	}

	if got := amount.CurrencyCode(); got != "EUR" {
		t.Errorf("currency = %q, want EUR", got)
	}

	cents, err := amount.Int64()
	if err != nil {
		t.Fatalf("Int64() error = %v", err)
	}

	if cents != 123456 {
		t.Errorf("cents = %d, want 123456", cents)
	}
}

func TestMoneyAmountRejectsInvalidCurrency(t *testing.T) {
	t.Parallel()

	m := Money{Cents: 100, Currency: Currency("XXO")}

	if _, err := m.Amount(); err == nil {
		t.Fatal("Amount() with a non-ISO currency code succeeded, want error")
	}
}

func TestMoneyAmountNegative(t *testing.T) {
	t.Parallel()

	m := Money{Cents: -50, Currency: Currency("USD")}

	amount, err := m.Amount()
	if err != nil {
		t.Fatalf("Amount() error = %v", err)
	}

	cents, err := amount.Int64()
	if err != nil {
		t.Fatalf("Int64() error = %v", err)
	}

	if cents != -50 {
		t.Errorf("cents = %d, want -50", cents)
	}
}
