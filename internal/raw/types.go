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

// User from Wise API (/v1/me, /v1/users/{userId}).
type User struct {
	ID     int64       `json:"id"`
	Name   string      `json:"name"`
	Email  string      `json:"email"`
	Active bool        `json:"active"`
	Detail *UserDetail `json:"details"`
}

// UserDetail is the personal-details block of a User; null on the wire when
// Wise has none on file.
type UserDetail struct {
	FirstName      string       `json:"firstName"`
	LastName       string       `json:"lastName"`
	PhoneNumber    string       `json:"phoneNumber"`
	DateOfBirth    string       `json:"dateOfBirth"` // "1977-01-01" or empty
	Occupation     string       `json:"occupation"`
	Avatar         string       `json:"avatar"`
	PrimaryAddress int64        `json:"primaryAddress"`
	Address        *UserAddress `json:"address"`
}

// UserAddress is the address on file for a user.
type UserAddress struct {
	CountryCode string `json:"countryCode"`
	FirstLine   string `json:"firstLine"`
	PostCode    string `json:"postCode"`
	City        string `json:"city"`
	State       string `json:"state"`
	Occupation  string `json:"occupation"`
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

// TotalFunds from Wise API (GET /v4/profiles/{id}/total-funds/{currency}).
type TotalFunds struct {
	TotalWorth     BalanceAmount `json:"totalWorth"`
	TotalAvailable BalanceAmount `json:"totalAvailable"`
}

// MultiCurrencyAccount from Wise API
// (GET /v1/profiles/{profileId}/multi-currency-account).
type MultiCurrencyAccount struct {
	ID               int64  `json:"id"`
	ProfileID        int64  `json:"profileId"`
	RecipientID      int64  `json:"recipientId"`
	CreationTime     string `json:"creationTime"`
	ModificationTime string `json:"modificationTime"`
	Active           bool   `json:"active"`
	Eligible         bool   `json:"eligible"`
}

// AccountDetails from Wise API
// (GET /{version}/balances/account-details/{id}, the 2027Q1 surface).
type AccountDetails struct {
	ID                    int64                 `json:"id"`
	ProfileID             int64                 `json:"profileId"`
	State                 string                `json:"state"` // "ACTIVE" or "INACTIVE"
	AccountHolderName     string                `json:"accountHolderName"`
	Country               string                `json:"country"` // ISO 3166-1 alpha-2
	AccountIdentification AccountIdentification `json:"accountIdentification"`
}

// AccountIdentification is the country/scheme-specific account number block
// of an account-details set: a spec oneOf, so exactly one member is present.
type AccountIdentification struct {
	BE *BEAccountIdentification `json:"beAccountIdentification,omitempty"`
	GB *GBAccountIdentification `json:"gbAccountIdentification,omitempty"`
	US *USAccountIdentification `json:"usAccountIdentification,omitempty"`
}

// BEAccountIdentification is the Belgian (SEPA) details format.
type BEAccountIdentification struct {
	BIC                string           `json:"bic,omitempty"`
	IBAN               string           `json:"iban"`
	InstitutionName    string           `json:"institutionName,omitempty"`
	InstitutionAddress *ISO20022Address `json:"institutionAddress,omitempty"`
}

// GBAccountIdentification is the UK details format.
type GBAccountIdentification struct {
	BIC                string           `json:"bic,omitempty"`
	IBAN               string           `json:"iban,omitempty"`
	SortCode           string           `json:"sortCode,omitempty"`
	AccountNumber      string           `json:"accountNumber"`
	InstitutionName    string           `json:"institutionName,omitempty"`
	InstitutionAddress *ISO20022Address `json:"institutionAddress,omitempty"`
}

// USAccountIdentification is the US (ACH) details format.
type USAccountIdentification struct {
	BIC                string           `json:"bic,omitempty"`
	RoutingNumber      string           `json:"routingNumber"`
	AccountNumber      string           `json:"accountNumber"`
	InstitutionName    string           `json:"institutionName,omitempty"`
	InstitutionAddress *ISO20022Address `json:"institutionAddress,omitempty"`
}

// ISO20022Address is the spec's structured postal address with optional
// free-form address lines.
type ISO20022Address struct {
	StreetName         string   `json:"streetName,omitempty"`
	BuildingNumber     string   `json:"buildingNumber,omitempty"`
	TownName           string   `json:"townName,omitempty"`
	PostCode           string   `json:"postCode,omitempty"`
	CountrySubDivision string   `json:"countrySubDivision,omitempty"`
	Country            string   `json:"country,omitempty"`
	AddressLines       []string `json:"addressLines,omitempty"`
}

// CurrencyInfo from Wise API (GET /v1/currencies).
type CurrencyInfo struct {
	Code             string   `json:"code"`
	Symbol           string   `json:"symbol"`
	Name             string   `json:"name"`
	CountryKeywords  []string `json:"countryKeywords"`
	SupportsDecimals bool     `json:"supportsDecimals"`
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
	ID                            string               `json:"id,omitempty"`
	SourceCurrency                string               `json:"sourceCurrency"`
	TargetCurrency                string               `json:"targetCurrency"`
	SourceAmount                  float64              `json:"sourceAmount"`
	TargetAmount                  float64              `json:"targetAmount"`
	PayIn                         string               `json:"payIn"`
	PayOut                        string               `json:"payOut"`
	Rate                          float64              `json:"rate"`
	CreatedTime                   string               `json:"createdTime"`
	ExpirationTime                string               `json:"expirationTime"`
	Status                        string               `json:"status"`
	Profile                       int64                `json:"profile"`
	RateType                      string               `json:"rateType"`
	ProvidedAmountType            string               `json:"providedAmountType"`
	GuaranteedTargetAmountAllowed bool                 `json:"guaranteedTargetAmountAllowed"`
	GuaranteedTargetAmount        bool                 `json:"guaranteedTargetAmount"`
	PaymentOptions                []QuotePaymentOption `json:"paymentOptions"`
	Notices                       []QuoteNotice        `json:"notices"`
}

// TransferPayment is one completed funding payment of a transfer
// (GET /v1/transfers/{transferId}/payments).
type TransferPayment struct {
	ID             int64   `json:"id"`
	Method         string  `json:"method"`
	PricingVariant *string `json:"pricingVariant"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	TimeCreated    string  `json:"timeCreated"`
	TimeUpdated    string  `json:"timeUpdated"`
}

// QuotePaymentOption is one payment method combination available for a quote.
type QuotePaymentOption struct {
	Disabled                   bool     `json:"disabled"`
	EstimatedDelivery          string   `json:"estimatedDelivery"`
	FormattedEstimatedDelivery string   `json:"formattedEstimatedDelivery"`
	Fee                        QuoteFee `json:"fee"`
	SourceAmount               float64  `json:"sourceAmount"`
	TargetAmount               float64  `json:"targetAmount"`
	SourceCurrency             string   `json:"sourceCurrency"`
	TargetCurrency             string   `json:"targetCurrency"`
	PayIn                      string   `json:"payIn"`
	PayOut                     string   `json:"payOut"`
	PayInProduct               string   `json:"payInProduct"`
	FeePercentage              float64  `json:"feePercentage"`
}

// QuoteFee is the fee breakdown for a quote payment option. All values are
// in major source-currency units, matching Wise's wire format; convert with
// Money conversions at the consumer boundary.
type QuoteFee struct {
	TransferWise float64 `json:"transferwise"`
	PayIn        float64 `json:"payIn"`
	Discount     float64 `json:"discount"`
	Partner      float64 `json:"partner"`
	Total        float64 `json:"total"`
}

// QuoteNotice is a message Wise wants shown to the user about a quote.
type QuoteNotice struct {
	Text string  `json:"text"`
	Link *string `json:"link"`
	Type string  `json:"type"`
}

// DeliveryEstimate from Wise API (/v1/delivery-estimates/{transferId}).
// EstimatedDeliveryDate uses Wise's zone-suffixed ISO-8601 layout
// ("2018-01-10T12:15:00.000+0000"), which differs from RFC3339.
type DeliveryEstimate struct {
	EstimatedDeliveryDate          string `json:"estimatedDeliveryDate"`
	FormattedEstimatedDeliveryDate string `json:"formattedEstimatedDeliveryDate"`
}

// TransferRequirement from Wise API (POST /v1/transfer-requirements).
// Fields is a dynamic form description: each entry groups one or more form
// controls (keyed by the JSON field name they map to) with validation
// metadata and, for selects, the list of allowed values.
//
// art-dupl:accept wire types intentionally mirror the parsed domain types in the root wise package; the tag-free domain layer decouples callers from the JSON shapes
type TransferRequirement struct {
	Type   string                    `json:"type"`
	Fields []TransferRequirementForm `json:"fields"`
}

// TransferRequirementForm is a labelled group of related form fields.
type TransferRequirementForm struct {
	Name  string                     `json:"name"`
	Group []TransferRequirementField `json:"group"`
}

// TransferRequirementField describes a single dynamic form control.
type TransferRequirementField struct {
	Key                         string                     `json:"key"`
	Name                        string                     `json:"name"`
	Type                        string                     `json:"type"`
	RefreshRequirementsOnChange bool                       `json:"refreshRequirementsOnChange"`
	Required                    bool                       `json:"required"`
	DisplayFormat               *string                    `json:"displayFormat,omitzero"`
	Example                     *string                    `json:"example,omitzero"`
	MinLength                   *int32                     `json:"minLength,omitzero"`
	MaxLength                   *int32                     `json:"maxLength,omitzero"`
	ValidationRegexp            *string                    `json:"validationRegexp,omitzero"`
	ValuesAllowed               []TransferRequirementValue `json:"valuesAllowed,omitzero"`
}

// TransferRequirementValue is one allowed value of a select field.
type TransferRequirementValue struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// AccountRequirement from the account-requirements endpoint
// (GET /v1/quotes/{quoteId}/account-requirements). One dynamic form per
// available payout route; Fields reuses the transfer-requirements form shape.
type AccountRequirement struct {
	Type      string                    `json:"type"`
	Title     string                    `json:"title"`
	UsageInfo string                    `json:"usageInfo"`
	Fields    []TransferRequirementForm `json:"fields"`
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
