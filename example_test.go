package wise_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/larsartmann/wise-go"
)

func ExampleNewCurrency() {
	currency, err := wise.NewCurrency("EUR")
	if err != nil {
		panic(err)
	}

	fmt.Println(currency)
	// Output: EUR
}

func ExampleMoney_String() {
	m := wise.Money{Cents: 1234, Currency: wise.Currency("EUR")}
	fmt.Println(m)
	// Output: EUR 12.34
}

func ExampleMoney_String_negative() {
	m := wise.Money{Cents: -5000, Currency: wise.Currency("USD")}
	fmt.Println(m)
	// Output: USD -50.00
}

// Cancel a transfer that has not been processed yet. Cancellation is final.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_CancelTransfer() {
	client := wise.New("your-api-key")

	transfer, err := client.CancelTransfer(context.Background(), wise.NewTransferID(16521634))
	if err != nil {
		if _, ok := errors.AsType[*wise.NotFoundError](err); ok {
			log.Fatal("transfer does not exist")
		}

		log.Fatal(err)
	}

	fmt.Println(transfer.Status)
}

// Fetch the expected delivery time for a transfer. The timezone parameter is
// optional; it only affects the pre-formatted text, not the timestamp.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_GetDeliveryEstimate() {
	client := wise.New("your-api-key")

	estimate, err := client.GetDeliveryEstimate(
		context.Background(), wise.NewTransferID(16521634), "Europe/Berlin",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("expected arrival: %s (%s)\n",
		estimate.EstimatedDeliveryDate, estimate.FormattedEstimatedDeliveryDate)
}

// Discover the dynamic transfer details a corridor requires and check a
// prepared request against them before spending a customerTransactionId.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_ValidateTransferRequirements() {
	client := wise.New("your-api-key")

	requirements, err := client.ValidateTransferRequirements(context.Background(),
		wise.ValidateTransferRequirementsRequest{
			TargetAccount: wise.NewRecipientID(98765432),
			QuoteID:       wise.NewQuoteID("11144c35-9fe8-4c32-b7fd-d05c2a7734bf"),
		})
	if err != nil {
		log.Fatal(err)
	}

	transferReq := wise.CreateTransferRequest{
		QuoteID:               wise.NewQuoteID("11144c35-9fe8-4c32-b7fd-d05c2a7734bf"),
		TargetAccount:         wise.NewRecipientID(98765432),
		CustomerTransactionID: "22244c35-9fe8-4c32-b7fd-d05c2a7734bf",
	}

	if missing := wise.MissingTransferDetails(requirements, transferReq); len(missing) > 0 {
		log.Fatalf("transfer details still missing: %v", missing)
	}
}

// Create an authenticated quote, guard against BLOCKED notices, and inspect
// the payment options before creating a transfer.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_CreateQuote() {
	client := wise.New("your-api-key")

	quote, err := client.CreateQuote(context.Background(), wise.NewProfileID(12345),
		wise.CreateQuoteRequest{
			SourceCurrency: wise.Currency("EUR"),
			TargetCurrency: wise.Currency("USD"),
			SourceAmount:   &wise.Money{Cents: 100_000, Currency: wise.Currency("EUR")},
			PreferredPayIn: wise.PayInBalance,
			PayOut:         wise.PayOutBankTransfer,
		})
	if err != nil {
		log.Fatal(err)
	}

	for _, notice := range quote.Notices {
		if notice.Type == wise.QuoteNoticeTypeBlocked {
			log.Fatalf("quote blocked: %s", notice.Text)
		}
	}

	for _, option := range quote.PaymentOptions {
		fmt.Printf("%s -> %s: fee %s, arrives %s\n",
			option.PayIn, option.PayOut, option.Source.Currency, option.FormattedEstimatedDelivery)
	}
}

// Fund a created transfer from the profile's Wise balance — the final step
// of the transfer flow. A rejected funding is a result, not an error.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_FundTransfer() {
	client := wise.New("your-api-key")

	funding, err := client.FundTransfer(
		context.Background(), wise.NewProfileID(12345), wise.NewTransferID(16521634),
	)
	if err != nil {
		log.Fatal(err)
	}

	if funding.Status == wise.FundingStatusRejected {
		if funding.ErrorCode == wise.FundingErrorCodeBalanceInsufficientFunds {
			log.Println("balance too low — top up and call FundTransfer again")

			return
		}

		log.Fatalf("funding rejected (%s): %s", funding.ErrorCode, funding.ErrorMessage)
	}

	fmt.Println(funding.Status)
}

// Download a balance statement as a file — here a PDF for archival. The
// interval must not exceed 469 days; use ListTransactions for the typed JSON
// statement instead.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_GetStatement() {
	client := wise.New("your-api-key")

	pdf, err := client.GetStatement(context.Background(), wise.GetStatementRequest{
		ProfileID: wise.NewProfileID(12345),
		BalanceID: wise.NewBalanceID(100),
		Currency:  wise.Currency("EUR"),
		From:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC),
		Format:    wise.StatementFormatPDF,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("statement-q1.pdf", pdf, 0o600); err != nil {
		log.Fatal(err)
	}
}

// Verify a webhook delivery before trusting its payload. Parse the
// subscription's public key once at startup; verify the raw body exactly as
// received against the X-Signature-SHA256 header on every delivery.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleVerifyWebhookSignature() {
	webhookKey, err := wise.ParseWebhookPublicKey([]byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA... (from the dashboard)
-----END PUBLIC KEY-----`))
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/webhooks/wise", func(w http.ResponseWriter, r *http.Request) {
		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			http.Error(w, "read body", http.StatusBadRequest)

			return
		}

		if !wise.VerifyWebhookSignature(body, r.Header.Get(wise.HeaderWebhookSignature), webhookKey) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)

			return
		}

		// Verified delivery: decode and process. Deduplicate on X-Delivery-Id —
		// Wise redelivers until acknowledged.
		w.WriteHeader(http.StatusOK)
	})
}

// List every currency Wise supports — public reference data, handy for
// populating currency pickers.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_ListCurrencies() {
	client := wise.New("your-api-key")

	currencies, err := client.ListCurrencies(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, currency := range currencies {
		if currency.SupportsDecimals {
			fmt.Println(currency.Code, currency.Symbol)
		}
	}
}

// Open a savings balance in USD. Savings balances require a name.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_CreateBalance() {
	client := wise.New("your-api-key")

	balance, err := client.CreateBalance(context.Background(), wise.CreateBalanceRequest{
		ProfileID: wise.NewProfileID(12345),
		Currency:  wise.Currency("USD"),
		Type:      wise.BalanceTypeSavings,
		Name:      "Vacation",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(balance.ID, balance.Name, balance.Amount)
}

// Overview of everything a profile owns, converted to one currency: Worth is
// cash plus invested portfolio valuation, Available is cash plus overdraft.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_GetTotalFunds() {
	client := wise.New("your-api-key")

	funds, err := client.GetTotalFunds(
		context.Background(), wise.NewProfileID(12345), wise.Currency("EUR"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("worth: %s, available: %s\n", funds.Worth, funds.Available)
}

// Read one balance by ID. Unlike ListBalances, the direct endpoint also
// returns hidden and invested balances.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_GetBalance() {
	client := wise.New("your-api-key")

	balance, err := client.GetBalance(
		context.Background(), wise.NewProfileID(12345), wise.NewBalanceID(100),
	)
	if err != nil {
		if _, ok := errors.AsType[*wise.NotFoundError](err); ok {
			log.Fatal("balance does not exist")
		}

		log.Fatal(err)
	}

	fmt.Println(balance.Currency, balance.Amount)
}

// Resolve the identity behind the API token — the user that owns the
// profiles ListProfiles returns.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_GetMe() {
	client := wise.New("your-api-key")

	me, err := client.GetMe(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(me.ID, me.Email)
}

// Read a user by ID. Personal API tokens can only read their own user.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_GetUser() {
	client := wise.New("your-api-key")

	user, err := client.GetUser(context.Background(), wise.NewUserID(101))
	if err != nil {
		log.Fatal(err)
	}

	if user.Details != nil && !user.Details.DateOfBirth.IsZero() {
		fmt.Println(user.Name, user.Details.DateOfBirth.Year())
	}
}

// Read the Multi-Currency Account — the holder of the currency balances. Its
// RecipientID is a real recipient: transfer to it to top up the account.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_GetMultiCurrencyAccount() {
	client := wise.New("your-api-key")

	account, err := client.GetMultiCurrencyAccount(
		context.Background(), wise.NewProfileID(12345),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("top up via recipient", account.RecipientID)
}

// Read every set of receiving bank details (IBANs, routing pairs, ...) and
// pull the labeled key values from the receive options' calls to action.
//
//nolint:testableexamples // documentation-only; runs against the live API
func ExampleClient_GetBankAccountDetails() {
	client := wise.New("your-api-key")

	details, err := client.GetBankAccountDetails(
		context.Background(), wise.NewProfileID(12345),
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, detail := range details {
		if detail.Deprecated {
			continue // Wise replaced these details; prefer the newer set
		}

		for _, option := range detail.ReceiveOptions {
			if option.Description != nil && option.Description.CTA != nil {
				fmt.Printf("%s %s: %s\n",
					detail.Currency, option.Description.CTA.Label, option.Description.CTA.Content)
			}
		}
	}
}
