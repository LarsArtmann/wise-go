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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
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

// TestRequireID pins the zero-ID guard's contract: every rejection carries
// the endpoint family's invalid-request code and the "<field> is required"
// message, for int64-branded and string-branded IDs alike.
func TestRequireID(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
		wantMsg  string
	}{
		{
			"zero int64 ID", requireID(NewProfileID(0), "wise.profile.invalid_request", "profileID"),
			"wise.profile.invalid_request",
			"[rejection:wise.profile.invalid_request] profileID is required",
		},
		{
			"nonzero int64 ID", requireID(NewProfileID(12345), "wise.profile.invalid_request", "profileID"),
			"", "",
		},
		{
			"empty string ID", requireID(NewQuoteID(""), "wise.quote.invalid_request", "quoteID"),
			"wise.quote.invalid_request",
			"[rejection:wise.quote.invalid_request] quoteID is required",
		},
		{
			"nonempty string ID", requireID(NewQuoteID("11114444-..."), "wise.quote.invalid_request", "quoteID"),
			"", "",
		},
		{
			"empty webhook subscription ID",
			requireID(NewWebhookSubscriptionID(""), "wise.webhook.invalid_request", "subscriptionID"),
			"wise.webhook.invalid_request",
			"[rejection:wise.webhook.invalid_request] subscriptionID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantCode == "" {
				if tt.err != nil {
					t.Fatalf("requireID() = %v, want nil", tt.err)
				}

				return
			}

			if tt.err == nil {
				t.Fatal("requireID() = nil, want rejection")
			}

			coder, ok := tt.err.(interface{ ErrorCode() string })
			if !ok {
				t.Fatalf("error does not implement ErrorCode: %T", tt.err)
			}

			if got := coder.ErrorCode(); got != tt.wantCode {
				t.Errorf("ErrorCode() = %q, want %q", got, tt.wantCode)
			}

			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
		})
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
