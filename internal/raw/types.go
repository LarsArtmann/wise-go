// Package raw contains Wise API wire-format types used for JSON deserialization.
// These types mirror Wise's JSON exactly and are not part of the public API.
package raw

import "math"

const centsPerUnit = 100

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
// Value is a float64 in major units (e.g., 1234.56 means 1234.56).
// Always convert to int64 cents via Cents() for precision-safe arithmetic.
type BalanceAmount struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

// Cents converts a BalanceAmount to int64 minor units (cents).
// Uses math.Round to handle IEEE 754 floating-point representation errors.
func (a BalanceAmount) Cents() int64 {
	return int64(math.Round(a.Value * centsPerUnit))
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

// ExchangeRate from Wise API (/v1/rates).
type ExchangeRate struct {
	Source string  `json:"source"`
	Target string  `json:"target"`
	Rate   float64 `json:"rate"`
	Time   string  `json:"time"`
}

// Recipient from Wise API (/v1/accounts, /v2/accounts).
type Recipient struct {
	ID                int64          `json:"id"`
	AccountHolderName string         `json:"accountHolderName"`
	Currency          string         `json:"currency"`
	Country           string         `json:"country"`
	Type              string         `json:"type"`
	Details           map[string]any `json:"details"`
	Active            bool           `json:"active"`
}

// Quote from Wise API (/v3/quotes).
// ID is a UUID string; it is omitted for unauthenticated quotes.
type Quote struct {
	ID             string  `json:"id,omitempty"`
	SourceCurrency string  `json:"sourceCurrency"`
	TargetCurrency string  `json:"targetCurrency"`
	SourceAmount   float64 `json:"sourceAmount"`
	TargetAmount   float64 `json:"targetAmount"`
	PayIn          string  `json:"payIn"`
	PayOut         string  `json:"payOut"`
	Rate           float64 `json:"rate"`
	CreatedTime    string  `json:"createdTime"`
	ExpirationTime string  `json:"expirationTime"`
	Status         string  `json:"status"`
}

// DeliveryEstimate from Wise API (/v1/delivery-estimates/{transferId}).
// EstimatedDeliveryDate uses Wise's zone-suffixed ISO-8601 layout
// ("2018-01-10T12:15:00.000+0000"), which differs from RFC3339.
type DeliveryEstimate struct {
	EstimatedDeliveryDate          string `json:"estimatedDeliveryDate"`
	FormattedEstimatedDeliveryDate string `json:"formattedEstimatedDeliveryDate"`
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
