package wise_test

import (
	"context"
	"errors"
	"fmt"
	"log"

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
