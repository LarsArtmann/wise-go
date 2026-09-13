package wise

import (
	"testing"

	"github.com/larsartmann/wise-go/internal/raw"
)

// Benchmarks for the hot parsing paths (statement imports process thousands
// of transactions per request). Run with:
//
//	GOEXPERIMENT=jsonv2 go test -bench . -benchmem -run ^$
//
// and compare across changes with benchstat (-count >= 6).

func BenchmarkBalanceAmountCents(b *testing.B) {
	amount := raw.BalanceAmount{Value: 1234.56, Currency: "EUR"}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if cents := amount.Cents(); cents != 123456 {
			b.Fatalf("Cents() = %d, want 123456", cents)
		}
	}
}

func BenchmarkClassifyTransactionType(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if got := classifyTransactionType(DetailTypeCardPayment, -1500); got != TransactionTypeCard {
			b.Fatalf("classify = %v, want card", got)
		}
	}
}

func BenchmarkMapTransaction(b *testing.B) {
	tx := raw.StatementTransaction{
		TransactionID: "tx-1",
		Date:          "2023-01-15 14:30:00",
		Amount:        raw.BalanceAmount{Value: -10.5, Currency: "EUR"},
		TotalFees:     raw.BalanceAmount{Value: 0, Currency: "EUR"},
		RunningBalance: raw.BalanceAmount{
			Value:    90,
			Currency: "EUR",
		},
		Details: raw.TransactionDetails{Type: "TRANSFER"},
	}

	profileID := NewProfileID(12345)
	balanceID := NewBalanceID(100)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := mapTransaction(tx, profileID, balanceID, "EUR"); err != nil {
			b.Fatalf("mapTransaction: %v", err)
		}
	}
}

func BenchmarkParseWebhookEvent(b *testing.B) {
	payload := []byte(`{
		"schema_version": "4.0.0",
		"subscription_id": "72195556-e5cb-495e-a010-b37a4f2a3043",
		"event_type": "transfers#state-change",
		"sent_at": "2020-01-01T12:34:56.123Z",
		"data": {"resource": {"id": 1, "profile_id": 2, "type": "transfer"},
			"current_state": "processing", "previous_state": "incoming_payment_waiting",
			"occurred_at": "2020-01-01T12:34:00Z"}
	}`)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		event, err := ParseWebhookEvent(payload)
		if err != nil {
			b.Fatalf("ParseWebhookEvent: %v", err)
		}

		if _, err := event.TransferStateChange(); err != nil {
			b.Fatalf("TransferStateChange: %v", err)
		}
	}
}

var _ = jsontext.Value{}  // jsontext referenced via json.RawMessage-free types above
var _ = json.Unmarshal // keep the json import honest if benchmarks shrink
var _ = fmt.Sprintf
