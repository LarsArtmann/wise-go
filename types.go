// Package wise provides a Go client for the Wise (TransferWise) API.
//
// The client handles authentication, rate limiting, retries with exponential backoff,
// and returns strongly-typed Go structs for all API responses.
//
// Basic usage:
//
//	client := wise.New("your-api-key")
//	profiles, err := client.ListProfiles(ctx)
package wise

import (
	"fmt"
	"time"
)

// API Base URLs.
//
// Sandbox URLs were migrated in November 2025 from the V1 environment
// (api.sandbox.transferwise.tech) to the V2 environment (api.wise-sandbox.com).
// V1 was fully deprecated on June 30, 2026.
const (
	ProductionURL = "https://api.wise.com"
	SandboxURL    = "https://api.wise-sandbox.com"
)

const (
	centsPerUnit       = 100
	currencyCodeLength = 3
)

// --- Value objects ---

// Currency is an ISO 4217 currency code (e.g., "EUR", "USD", "GBP").
// Construct via NewCurrency for validation, or use Currency("EUR") directly
// when the value is known to be valid.
type Currency string

// NewCurrency validates and constructs a Currency from a raw string.
// Returns an error if the string is not exactly 3 uppercase ASCII letters.
//
//nolint:err113 // currency validation needs dynamic error messages with context
func NewCurrency(s string) (Currency, error) {
	if len(s) != currencyCodeLength {
		return "", fmt.Errorf("currency must be exactly 3 letters, got %d", len(s))
	}

	for _, c := range s {
		if c < 'A' || c > 'Z' {
			return "", fmt.Errorf("currency must be uppercase ASCII letters, got %q", s)
		}
	}

	return Currency(s), nil
}

// Money is a monetary amount in cents paired with its currency.
// The pairing makes mismatched currency/amount combinations unrepresentable.
type Money struct {
	Cents    int64
	Currency Currency
}

// String formats Money as "CUR DD.DD" (e.g., "EUR 12.34", "USD -50.00").
func (m Money) String() string {
	abs := m.Cents
	if abs < 0 {
		abs = -abs
	}

	major := abs / centsPerUnit
	minor := abs % centsPerUnit

	if m.Cents < 0 {
		return fmt.Sprintf("%s -%d.%02d", m.Currency, major, minor)
	}

	return fmt.Sprintf("%s %d.%02d", m.Currency, major, minor)
}

// --- Parsed result types (strongly typed, int64 cents, time.Time) ---

// Profile is the parsed representation of a Wise profile.
type Profile struct {
	ID        ProfileID
	Type      ProfileType
	Name      string
	Email     string
	CreatedAt time.Time
}

// Balance is the parsed representation of a Wise balance.
type Balance struct {
	ID        BalanceID
	Currency  Currency
	Type      BalanceType
	Name      string
	Amount    Money
	Reserved  Money
	Visible   bool
	CreatedAt time.Time
}

// Transaction is the parsed representation of a Wise transaction.
// All monetary amounts are Money values (int64 cents paired with Currency)
// for precision-safe arithmetic.
//
// Date is returned in UTC. Wise statement dates carry no timezone; the SDK
// interprets them as UTC via time.Parse. Convert explicitly before comparing
// against local-time values.
type Transaction struct {
	ID             TransactionID
	ProfileID      ProfileID
	BalanceID      BalanceID
	Amount         Money
	Fees           Money
	Total          Money
	RunningBalance Money
	Exchange       *TransactionExchange
	Type           TransactionType
	Description    string
	Reference      string
	Category       string
	MerchantName   string
	Date           time.Time
}

// TransactionExchange captures the currency-conversion details of an
// exchange transaction. nil when the transaction is not a conversion.
type TransactionExchange struct {
	From Money
	To   Money
	Rate float64
}

// --- Enum types ---

// ProfileType identifies the kind of Wise profile.
type ProfileType string

const (
	ProfileTypePersonal ProfileType = "personal"
	ProfileTypeBusiness ProfileType = "business"
)

// BalanceType identifies the kind of Wise balance.
type BalanceType string

const (
	BalanceTypeStandard BalanceType = "standard"
	BalanceTypeSavings  BalanceType = "savings"
)

// InvestmentState identifies whether a Wise balance is invested.
type InvestmentState string

// InvestmentState values reported by Wise balances (Balance.InvestmentState).
const (
	InvestmentStateNotInvested InvestmentState = "NOT_INVESTED"
	InvestmentStateInvested    InvestmentState = "INVESTED"
)

// TransactionType categorizes the transaction kind.
type TransactionType string

const (
	TransactionTypeCard     TransactionType = "card"
	TransactionTypeCredit   TransactionType = "credit"
	TransactionTypeDebit    TransactionType = "debit"
	TransactionTypeExchange TransactionType = "exchange"
	TransactionTypeFee      TransactionType = "fee"
	TransactionTypeRefund   TransactionType = "refund"
	TransactionTypeTransfer TransactionType = "transfer"
	TransactionTypePayment  TransactionType = "payment"
)

// --- Request types ---

// ListTransactionsRequest parameters for listing transactions.
type ListTransactionsRequest struct {
	ProfileID ProfileID
	BalanceID BalanceID
	Currency  Currency
	From      time.Time
	To        time.Time
	Type      DetailType // Optional filter by transaction type. See DetailType* constants.
}

// ListTransactionsResponse from listing transactions.
type ListTransactionsResponse struct {
	Transactions          []Transaction
	EndOfStatementBalance Money
}

// Transfer is the parsed representation of a Wise transfer — a payment order
// to a recipient account based on a quote. Unlike statement transactions,
// transfers are retrievable with a personal API token without SCA approval.
//
// Created is parsed via the SDK's tolerant timestamp handling; zoneless
// values are interpreted as UTC.
type Transfer struct {
	ID                    TransferID
	RecipientID           RecipientID
	Status                TransferStatus
	Rate                  float64
	Source                Money
	Target                Money
	Created               time.Time
	Reference             string
	CustomerTransactionID string
	HasActiveIssues       bool
}

// TransferStatus is a Wise transfer lifecycle status. Wise may add values;
// unrecognised statuses are preserved as-is, so comparisons should use the
// constants only for the states they handle.
type TransferStatus string

// Documented transfer lifecycle statuses (see Wise "Tracking transfers").
const (
	TransferStatusIncomingPaymentWaiting TransferStatus = "incoming_payment_waiting"
	TransferStatusWaitingForFunds        TransferStatus = "waiting_for_funds"
	TransferStatusProcessing             TransferStatus = "processing"
	TransferStatusFundsConverted         TransferStatus = "funds_converted"
	TransferStatusOutgoingPaymentSent    TransferStatus = "outgoing_payment_sent"
	TransferStatusBouncesCompleted       TransferStatus = "bounces_completed"
	TransferStatusDelivered              TransferStatus = "delivered"
	TransferStatusRefunded               TransferStatus = "refunded"
	TransferStatusCancelled              TransferStatus = "cancelled"
	TransferStatusUnsuccessful           TransferStatus = "unsuccessful"
	TransferStatusChargedBack            TransferStatus = "charged_back"
)
