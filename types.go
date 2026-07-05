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
	"math"
	"time"
)

// API Base URLs.
const (
	ProductionURL = "https://api.wise.com"
	SandboxURL    = "https://api.sandbox.transferwise.tech"
)

// --- Raw API types (direct JSON deserialization from Wise API) ---

// Profile from Wise API (/v2/profiles).
type Profile struct {
	ID           int64  `json:"id"`
	PublicID     string `json:"publicId"`
	UserID       int64  `json:"userId"`
	Type         string `json:"type"` // "PERSONAL" or "BUSINESS"
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	BusinessName string `json:"businessName,omitempty"`
	Email        string `json:"email"`
	CreatedAt    string `json:"createdAt"`
}

// Balance from Wise API (/v4/profiles/{id}/balances).
type Balance struct {
	ID               int64         `json:"id"`
	Currency         string        `json:"currency"`
	Type             string        `json:"type"` // "STANDARD" or "SAVINGS"
	Name             string        `json:"name"`
	InvestmentState  string        `json:"investmentState"`
	Amount           BalanceAmount `json:"amount"`
	ReservedAmount   BalanceAmount `json:"reservedAmount"`
	CreationTime     string        `json:"creationTime"`
	ModificationTime string        `json:"modificationTime"`
	Visible          bool          `json:"visible"`
}

// BalanceAmount represents a monetary amount from the Wise API.
// Value is a float64 in major units (e.g., 1234.56 means €1234.56).
// Always convert to int64 cents via Cents() for precision-safe arithmetic.
type BalanceAmount struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

// Cents converts a BalanceAmount to int64 minor units (cents).
// Uses math.Round to handle IEEE 754 floating-point representation errors.
func (a BalanceAmount) Cents() int64 {
	return int64(math.Round(a.Value * 100))
}

// StatementResponse from balance-statements endpoint.
type StatementResponse struct {
	Transactions          []StatementTransaction `json:"transactions"`
	EndOfStatementBalance BalanceAmount          `json:"endOfStatementBalance"`
}

// StatementTransaction from Wise balance statements.
type StatementTransaction struct {
	TransactionID   string             `json:"transactionId"`
	Date            string             `json:"date"`
	Amount          BalanceAmount      `json:"amount"`
	TotalFees       BalanceAmount      `json:"totalFees"`
	Details         TransactionDetails `json:"details"`
	ExchangeDetails *ExchangeDetails   `json:"exchangeDetails,omitempty"`
	RunningBalance  BalanceAmount      `json:"runningBalance"`
	ReferenceNumber string             `json:"referenceNumber"`
}

// TransactionDetails contains transaction metadata.
type TransactionDetails struct {
	Type                 string `json:"type"`
	Description          string `json:"description"`
	Category             string `json:"category,omitempty"`
	MerchantName         string `json:"merchantName,omitempty"`
	MerchantCategoryCode int    `json:"merchantCategoryCode,omitempty"`
	PaymentReference     string `json:"paymentReference,omitempty"`
	TransferReference    string `json:"transferReference,omitempty"`
}

// ExchangeDetails for currency conversion transactions.
type ExchangeDetails struct {
	FromAmount   BalanceAmount `json:"fromAmount"`
	ToAmount     BalanceAmount `json:"toAmount"`
	Rate         float64       `json:"rate"`
	FromCurrency string        `json:"fromCurrency"`
	ToCurrency   string        `json:"toCurrency"`
}

// ErrorResponse from Wise API.
type ErrorResponse struct {
	Errors []ErrorDetail `json:"errors"`
}

// ErrorDetail contains error information from the Wise API.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// --- Parsed result types (strongly typed, int64 cents, time.Time) ---

// ProfileResult is the parsed representation of a Wise profile.
type ProfileResult struct {
	ID        ProfileID
	Type      ProfileType
	Name      string
	Email     string
	CreatedAt time.Time
}

// BalanceResult is the parsed representation of a Wise balance.
type BalanceResult struct {
	ID               BalanceID
	Currency         string
	Type             BalanceType
	Name             string
	AmountCents      int64
	AmountCurrency   string
	ReservedCents    int64
	ReservedCurrency string
	Visible          bool
	CreatedAt        time.Time
}

// Transaction is the parsed representation of a Wise transaction.
// All monetary amounts are in cents (int64) for precision-safe arithmetic.
type Transaction struct {
	ID                     TransactionID
	ProfileID              ProfileID
	BalanceID              BalanceID
	AmountCents            int64
	AmountCurrency         string
	FeesCents              int64
	FeesCurrency           string
	TotalCents             int64
	TotalCurrency          string
	RunningBalanceCents    int64
	RunningBalanceCurrency string
	Exchange               *TransactionExchange
	Type                   TransactionType
	Description            string
	Reference              string
	Category               string
	MerchantName           string
	Date                   time.Time
}

// TransactionExchange captures the currency-conversion details of an
// exchange transaction. nil when the transaction is not a conversion.
type TransactionExchange struct {
	FromCents    int64
	FromCurrency string
	ToCents      int64
	ToCurrency   string
	Rate         float64
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
	BalanceTypeStandard BalanceType = "STANDARD"
	BalanceTypeSavings  BalanceType = "SAVINGS"
)

// InvestmentState values reported by Wise balances (Balance.InvestmentState).
const (
	InvestmentStateNotInvested = "NOT_INVESTED"
	InvestmentStateInvested    = "INVESTED"
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
	TransactionTypeUnknown  TransactionType = "unknown"
)

// --- Request types ---

// ListTransactionsRequest parameters for listing transactions.
type ListTransactionsRequest struct {
	ProfileID ProfileID
	BalanceID BalanceID
	Currency  string
	From      time.Time
	To        time.Time
	Type      string // Optional filter by transaction type
}

// ListTransactionsResponse from listing transactions.
type ListTransactionsResponse struct {
	Transactions []Transaction
	HasMore      bool // Always false for Wise (returns all in one request)
}
