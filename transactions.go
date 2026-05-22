package wise

import (
	"context"
	"fmt"
	"net/url"
	"time"

	id "github.com/larsartmann/go-branded-id"
)

// ListTransactions returns transactions for a balance within a time range.
// Wise returns all transactions in a single response (no pagination).
func (c *Client) ListTransactions(
	ctx context.Context,
	req ListTransactionsRequest,
) (*ListTransactionsResponse, error) {
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
			req.ProfileID, req.BalanceID, req.Currency, err)
	}

	transactions := make([]Transaction, len(statement.Transactions))
	for i, t := range statement.Transactions {
		tx, mapErr := mapTransaction(t, req.ProfileID, req.BalanceID, req.Currency, c.now)
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
	now func() time.Time, //nolint:revive // reserved for future clock injection
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

	// AmountCents is the absolute value of the transaction amount
	amountCents := totalCents
	if amountCents < 0 {
		amountCents = -amountCents
	}

	txType := classifyTransactionType(t.Details.Type, t.Amount.Value)

	return Transaction{
		ID:             id.NewID[TransactionBrand](t.TransactionID),
		ProfileID:      profileID,
		BalanceID:      balanceID,
		AmountCents:    amountCents,
		AmountCurrency: currency,
		FeesCents:      t.TotalFees.Cents(),
		FeesCurrency:   t.TotalFees.Currency,
		TotalCents:     totalCents,
		TotalCurrency:  t.Amount.Currency,
		Type:           txType,
		Description:    t.Details.Description,
		Reference:      t.ReferenceNumber,
		Category:       t.Details.Category,
		MerchantName:   t.Details.MerchantName,
		Date:           date,
	}, nil
}

// classifyTransactionType maps Wise detail types to SDK transaction types.
func classifyTransactionType(wiseType string, amount float64) TransactionType {
	switch wiseType {
	case "CARD_PAYMENT", "CARD_REFUND":
		if amount > 0 {
			return TransactionTypeRefund
		}

		return TransactionTypeCard
	case "TRANSFER":
		return TransactionTypeTransfer
	case "PAYMENT":
		return TransactionTypePayment
	case "CONVERSION", "EXCHANGE":
		return TransactionTypeExchange
	case "FEE":
		return TransactionTypeFee
	default:
		if amount > 0 {
			return TransactionTypeCredit
		}

		return TransactionTypeDebit
	}
}
