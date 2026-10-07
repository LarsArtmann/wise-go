package wise

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json/jsontext"
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

// BenchmarkVerifyWebhookSignature measures the per-delivery verification hot
// path (RSA-SHA256 PKCS1v15 over the raw request bytes).
func BenchmarkVerifyWebhookSignature(b *testing.B) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatalf("generate key: %v", err)
	}

	payload := []byte(`{"schema_version":"4.0.0","event_type":"transfers#state-change","data":{"resource":{"id":1}}}`)

	digest := sha256.Sum256(payload)

	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		b.Fatalf("sign payload: %v", err)
	}

	signature := base64.StdEncoding.EncodeToString(sig)
	publicKey := &key.PublicKey

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if !VerifyWebhookSignature(payload, signature, publicKey) {
			b.Fatal("VerifyWebhookSignature(valid) = false, want true")
		}
	}
}

// BenchmarkDecodeExchangeRates measures the rates array-decode path (the
// live /v1/rates wire shape; several entries, requested pair not first).
func BenchmarkDecodeExchangeRates(b *testing.B) {
	payload := jsontext.Value(`[
		{"source":"GBP","target":"USD","rate":1.27,"time":"2026-10-07T00:17:01+0000"},
		{"source":"EUR","target":"GBP","rate":0.84,"time":"2026-10-07T00:17:01+0000"},
		{"source":"EUR","target":"USD","rate":1.0854,"time":"2026-10-07T00:17:01+0000"},
		{"source":"USD","target":"EUR","rate":0.92,"time":"2026-10-07T00:17:01+0000"}
	]`)

	source, target := Currency("EUR"), Currency("USD")

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := decodeExchangeRates(payload, source, target); err != nil {
			b.Fatalf("decodeExchangeRates: %v", err)
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
