package wise

import (
	"context"
	"fmt"
	"net/url"
	"time"

	id "github.com/larsartmann/go-branded-id"
	errorfamily "github.com/larsartmann/go-error-family"
)

// ListTransactions returns transactions for a balance within a time range.
// Wise returns all transactions in a single response (no pagination).
func (c *Client) ListTransactions(
	ctx context.Context,
	req ListTransactionsRequest,
) (*ListTransactionsResponse, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/profiles/%d/balance-statements/%d/statement.json",
		req.ProfileID.Get(), req.BalanceID.Get())

	query := func() string {
		v := url.Values{}
		v.Set("currency", req.Currency)
		v.Set("intervalStart", req.From.Format(time.RFC3339))
		v.Set("intervalEnd", req.To.Format(time.RFC3339))

		if req.Type != "" {
			v.Set("type", req.Type)
		}

		return v.Encode()
	}

	var statement StatementResponse

	err := c.getWithQuery(ctx, path, query, &statement)
	if err != nil {
		return nil, fmt.Errorf("list transactions for profileID=%d balanceID=%d currency=%s: %w",
			req.ProfileID.Get(), req.BalanceID.Get(), req.Currency, err)
	}

	transactions := make([]Transaction, len(statement.Transactions))
	for i, t := range statement.Transactions {
		tx, mapErr := mapTransaction(t, req.ProfileID, req.BalanceID, req.Currency)
		if mapErr != nil {
			return nil, fmt.Errorf(
				"map transaction %s for profileID=%d balanceID=%d currency=%s: %w",
				t.TransactionID,
				req.ProfileID,
				req.BalanceID,
				req.Currency,
				mapErr,
			)
		}

		transactions[i] = tx
	}

	return &ListTransactionsResponse{
		Transactions: transactions,
		HasMore:      false, // Wise returns all in one request
	}, nil
}

func mapTransaction(
	t StatementTransaction,
	profileID ProfileID,
	balanceID BalanceID,
	currency string,
) (Transaction, error) {
	date, err := parseWiseDate(t.Date)
	if err != nil {
		return Transaction{}, fmt.Errorf(
			"parse date %q for profileID=%d balanceID=%d currency=%s: %w",
			t.Date,
			profileID.Get(),
			balanceID.Get(),
			currency,
			err,
		)
	}

	totalCents := t.Amount.Cents()

	// AmountCents is the absolute value of the transaction amount.
	amountCents := totalCents
	if amountCents < 0 {
		amountCents = -amountCents
	}

	txType := classifyTransactionType(t.Details.Type, t.Amount.Value)

	return Transaction{
		ID:                     id.NewID[TransactionBrand](t.TransactionID),
		ProfileID:              profileID,
		BalanceID:              balanceID,
		AmountCents:            amountCents,
		AmountCurrency:         t.Amount.Currency,
		FeesCents:              t.TotalFees.Cents(),
		FeesCurrency:           t.TotalFees.Currency,
		TotalCents:             totalCents,
		TotalCurrency:          t.Amount.Currency,
		RunningBalanceCents:    t.RunningBalance.Cents(),
		RunningBalanceCurrency: t.RunningBalance.Currency,
		Exchange:               mapExchange(t.ExchangeDetails),
		Type:                   txType,
		Description:            t.Details.Description,
		Reference:              t.ReferenceNumber,
		Category:               t.Details.Category,
		MerchantName:           t.Details.MerchantName,
		Date:                   date,
	}, nil
}

// mapExchange converts raw Wise exchange details into the result type.
// Returns nil when there are no exchange details.
func mapExchange(ed *ExchangeDetails) *TransactionExchange {
	if ed == nil {
		return nil
	}

	return &TransactionExchange{
		FromCents:    ed.FromAmount.Cents(),
		FromCurrency: ed.FromAmount.Currency,
		ToCents:      ed.ToAmount.Cents(),
		ToCurrency:   ed.ToAmount.Currency,
		Rate:         ed.Rate,
	}
}

// DetailType constants are Wise's wire-format values for details.type.
// Use these as ListTransactionsRequest.Type filter values.
const (
	DetailTypeCardPayment = "CARD_PAYMENT"
	DetailTypeCardRefund  = "CARD_REFUND"
	DetailTypeTransfer    = "TRANSFER"
	DetailTypePayment     = "PAYMENT"
	DetailTypeConversion  = "CONVERSION"
	DetailTypeExchange    = "EXCHANGE"
	DetailTypeFee         = "FEE"
)

// classifyTransactionType maps Wise detail types to SDK transaction types.
// CARD_PAYMENT is always a card transaction (typically a debit, but the type
// does not change with sign). CARD_REFUND is amount-dependent: positive amounts
// are classified as refunds, non-positive fall back to card. See README for the
// full contract.
func classifyTransactionType(wiseType string, amount float64) TransactionType {
	switch wiseType {
	case DetailTypeCardPayment:
		return TransactionTypeCard
	case DetailTypeCardRefund:
		if amount > 0 {
			return TransactionTypeRefund
		}

		return TransactionTypeCard
	case DetailTypeTransfer:
		return TransactionTypeTransfer
	case DetailTypePayment:
		return TransactionTypePayment
	case DetailTypeConversion, DetailTypeExchange:
		return TransactionTypeExchange
	case DetailTypeFee:
		return TransactionTypeFee
	default:
		if amount > 0 {
			return TransactionTypeCredit
		}

		return TransactionTypeDebit
	}
}

const invalidRequestCode = "wise.transactions.invalid_request"

// validate checks the request for client-side errors before hitting the API,
// failing fast with a clear rejection instead of an opaque API error.
func (r ListTransactionsRequest) validate() error {
	if r.Currency == "" {
		return errorfamily.NewRejection(invalidRequestCode, "currency is required")
	}

	if r.From.After(r.To) {
		return errorfamily.NewRejection(
			invalidRequestCode,
			"intervalStart must not be after intervalEnd",
		)
	}

	return nil
}
