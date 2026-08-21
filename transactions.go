package wise

import (
	"context"
	"fmt"
	"net/url"
	"time"

	id "github.com/larsartmann/go-branded-id"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
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
		v.Set("currency", string(req.Currency))
		v.Set("intervalStart", formatWiseTimestamp(req.From))
		v.Set("intervalEnd", formatWiseTimestamp(req.To))

		if req.Type != "" {
			v.Set("type", string(req.Type))
		}

		return v.Encode()
	}

	var statement raw.StatementResponse

	err := c.getWithQuery(ctx, path, query, &statement)
	if err != nil {
		return nil, fmt.Errorf("list transactions for profileID=%d balanceID=%d currency=%s: %w",
			req.ProfileID.Get(), req.BalanceID.Get(), req.Currency, err)
	}

	transactions := make([]Transaction, 0, len(statement.Transactions))
	for _, t := range statement.Transactions {
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

		transactions = append(transactions, tx)
	}

	endBalance, err := toMoney(statement.EndOfStatementBalance)
	if err != nil {
		return nil, fmt.Errorf("end of statement balance: %w", err)
	}

	return &ListTransactionsResponse{
		Transactions:          transactions,
		EndOfStatementBalance: endBalance,
	}, nil
}

func mapTransaction(
	t raw.StatementTransaction,
	profileID ProfileID,
	balanceID BalanceID,
	currency Currency,
) (Transaction, error) {
	date, err := parseWiseTimestamp(t.Date)
	if err != nil {
		return Transaction{}, errorfamily.WrapCorruption(err, "wise.transaction.parse_date",
			fmt.Sprintf("parse date %q for profileID=%d balanceID=%d currency=%s",
				t.Date, profileID.Get(), balanceID.Get(), currency))
	}

	txType := classifyTransactionType(DetailType(t.Details.Type), t.Amount.Value)

	total, err := toMoney(t.Amount)
	if err != nil {
		return Transaction{}, fmt.Errorf("total amount: %w", err)
	}

	amount := total
	if amount.Cents < 0 {
		amount.Cents = -amount.Cents
	}

	fees, err := toMoney(t.TotalFees)
	if err != nil {
		return Transaction{}, fmt.Errorf("fees amount: %w", err)
	}

	runningBalance, err := toMoney(t.RunningBalance)
	if err != nil {
		return Transaction{}, fmt.Errorf("running balance: %w", err)
	}

	exch, err := mapExchange(t.ExchangeDetails)
	if err != nil {
		return Transaction{}, fmt.Errorf("exchange: %w", err)
	}

	return Transaction{
		ID:             id.NewID[TransactionBrand](t.TransactionID),
		ProfileID:      profileID,
		BalanceID:      balanceID,
		Amount:         amount,
		Fees:           fees,
		Total:          total,
		RunningBalance: runningBalance,
		Exchange:       exch,
		Type:           txType,
		Description:    t.Details.Description,
		Reference:      t.ReferenceNumber,
		Category:       t.Details.Category,
		MerchantName:   t.Details.MerchantName,
		Date:           date,
	}, nil
}

// mapExchange converts raw Wise exchange details into the result type.
// Returns nil when there are no exchange details.
func mapExchange(ed *raw.ExchangeDetails) (*TransactionExchange, error) {
	if ed == nil {
		return nil, nil //nolint:nilnil // nil exchange details = nil result, no error
	}

	from, err := toMoney(ed.FromAmount)
	if err != nil {
		return nil, fmt.Errorf("from amount: %w", err)
	}

	to, err := toMoney(ed.ToAmount)
	if err != nil {
		return nil, fmt.Errorf("to amount: %w", err)
	}

	return &TransactionExchange{
		From: from,
		To:   to,
		Rate: ed.Rate,
	}, nil
}

// DetailType represents Wise's wire-format values for details.type.
// Use the DetailType* constants as ListTransactionsRequest.Type filter values.
type DetailType string

// DetailType constants are Wise's wire-format values for details.type.
const (
	DetailTypeCardPayment DetailType = "CARD_PAYMENT"
	DetailTypeCardRefund  DetailType = "CARD_REFUND"
	DetailTypeTransfer    DetailType = "TRANSFER"
	DetailTypePayment     DetailType = "PAYMENT"
	DetailTypeConversion  DetailType = "CONVERSION"
	DetailTypeExchange    DetailType = "EXCHANGE"
	DetailTypeFee         DetailType = "FEE"
)

// classifyTransactionType maps Wise detail types to SDK transaction types.
// CARD_PAYMENT is always a card transaction (typically a debit, but the type
// does not change with sign). CARD_REFUND is amount-dependent: positive amounts
// are classified as refunds, non-positive fall back to card. See README for the
// full contract.
func classifyTransactionType(wiseType DetailType, amount float64) TransactionType {
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

// StatementFormat identifies a balance-statement file format. The format
// selects the endpoint suffix; use ListTransactions for the parsed JSON form.
type StatementFormat string

// Statement file formats Wise serves for a balance.
const (
	StatementFormatCSV   StatementFormat = "csv"
	StatementFormatPDF   StatementFormat = "pdf"  // Includes Wise branding.
	StatementFormatXLSX  StatementFormat = "xlsx" // Excel.
	StatementFormatCAMT  StatementFormat = "xml"  // CAMT.053 XML.
	StatementFormatMT940 StatementFormat = "mt940"
	StatementFormatQIF   StatementFormat = "qif"
)

// isKnownStatementFormat reports whether format is one the SDK can request
// (kept in sync with the StatementFormat constants so an unknown format is
// rejected client-side instead of 404ing server-side).
func isKnownStatementFormat(format StatementFormat) bool {
	switch format {
	case StatementFormatCSV, StatementFormatPDF, StatementFormatXLSX,
		StatementFormatCAMT, StatementFormatMT940, StatementFormatQIF:
		return true
	default:
		return false
	}
}

// GetStatementRequest parameters for downloading a balance statement as a
// file. The interval From..To cannot exceed 469 days (Wise limit).
type GetStatementRequest struct {
	ProfileID ProfileID
	BalanceID BalanceID
	Currency  Currency
	From      time.Time
	To        time.Time
	Type      DetailType // Optional transaction-type filter.
	Format    StatementFormat
}

// GetStatement downloads a balance statement as a file in the requested
// format (CSV, PDF, XLSX, CAMT.053 XML, MT940, or QIF) and returns its raw
// bytes. Use ListTransactions for the typed JSON statement instead.
//
// The endpoint is SCA-protected for UK/EEA profiles.
func (c *Client) GetStatement(ctx context.Context, req GetStatementRequest) ([]byte, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/profiles/%d/balance-statements/%d/statement.%s",
		req.ProfileID.Get(), req.BalanceID.Get(), req.Format)

	query := func() string {
		v := url.Values{}
		v.Set("currency", string(req.Currency))
		v.Set("intervalStart", formatWiseTimestamp(req.From))
		v.Set("intervalEnd", formatWiseTimestamp(req.To))

		if req.Type != "" {
			v.Set("type", string(req.Type))
		}

		return v.Encode()
	}

	data, err := c.getRaw(ctx, path, query)
	if err != nil {
		return nil, fmt.Errorf("get %s statement for profileID=%d balanceID=%d: %w",
			req.Format, req.ProfileID.Get(), req.BalanceID.Get(), err)
	}

	return data, nil
}

func (r GetStatementRequest) validate() error {
	if r.ProfileID.Get() == 0 {
		return errorfamily.NewRejection(invalidRequestCode, "profileID is required")
	}

	if r.BalanceID.Get() == 0 {
		return errorfamily.NewRejection(invalidRequestCode, "balanceID is required")
	}

	if r.Currency == "" {
		return errorfamily.NewRejection(invalidRequestCode, "currency is required")
	}

	if r.From.After(r.To) {
		return errorfamily.NewRejection(
			invalidRequestCode,
			"intervalStart must not be after intervalEnd",
		)
	}

	if !isKnownStatementFormat(r.Format) {
		return errorfamily.NewRejection(
			invalidRequestCode,
			fmt.Sprintf("unknown statement format %q (want csv, pdf, xlsx, xml, mt940, or qif)", r.Format),
		)
	}

	return nil
}
