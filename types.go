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

// ExchangeRate is the parsed representation of a Wise exchange rate.
// Source and Target are ISO 4217 currency codes; Rate is the amount of
// Target currency per one unit of Source currency.
type ExchangeRate struct {
	Source Currency
	Target Currency
	Rate   float64
	Time   time.Time
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

// Quote is the parsed representation of a Wise quote — a locked exchange rate
// offer used to create a transfer. Authenticated quotes expire after 30 minutes.
//
// ID is empty for unauthenticated quotes, which cannot be used to create a
// transfer. PaymentOptions lists the available pay-in/pay-out combinations
// with their fees and estimated delivery times; Notices carries messages Wise
// requires to be shown to the user (a BLOCKED notice means the quote must not
// be used to create a transfer).
type Quote struct {
	ID                            QuoteID
	Source                        Money
	Target                        Money
	PayIn                         PayIn
	PayOut                        PayOut
	Rate                          float64
	Created                       time.Time
	ExpirationTime                time.Time
	Status                        QuoteStatus
	Profile                       ProfileID // Zero for unauthenticated quotes.
	RateType                      QuoteRateType
	ProvidedAmountType            QuoteProvidedAmountType
	GuaranteedTargetAmountAllowed bool
	GuaranteedTargetAmount        bool
	PaymentOptions                []QuotePaymentOption
	Notices                       []QuoteNotice
}

// QuotePaymentOption is one pay-in/pay-out combination available for a quote,
// with its fee breakdown and estimated delivery time. Source and Target are
// the amounts payable/receivable when using this option; EstimatedDelivery is
// parsed via the tolerant timestamp parser.
type QuotePaymentOption struct {
	Disabled                   bool
	EstimatedDelivery          time.Time
	FormattedEstimatedDelivery string
	Fee                        QuoteFee
	Source                     Money
	Target                     Money
	PayIn                      PayIn
	PayOut                     PayOut
	PayInProduct               string
	FeePercentage              float64
}

// QuoteFee is the fee breakdown Wise reports for a quote payment option, in
// source-currency major units (Wise's wire format). Total is the value to
// display when showing fees to the user.
type QuoteFee struct {
	TransferWise float64
	PayIn        float64
	Discount     float64
	Partner      float64
	Total        float64
}

// QuoteNotice is a message Wise requires to be shown to the user.
// A notice of type BLOCKED means the quote must not be used to create a
// transfer.
type QuoteNotice struct {
	Text string
	Link *string
	Type QuoteNoticeType
}

// QuoteRateType distinguishes a locked (guaranteed) rate from a floating one.
type QuoteRateType string

const (
	QuoteRateTypeFixed    QuoteRateType = "FIXED"
	QuoteRateTypeFloating QuoteRateType = "FLOATING"
)

// QuoteProvidedAmountType records whether the quote was created from a
// source or a target amount.
type QuoteProvidedAmountType string

const (
	QuoteProvidedAmountTypeSource QuoteProvidedAmountType = "SOURCE"
	QuoteProvidedAmountTypeTarget QuoteProvidedAmountType = "TARGET"
)

// QuoteNoticeType is the severity class of a quote notice.
type QuoteNoticeType string

const (
	QuoteNoticeTypeWarning QuoteNoticeType = "WARNING"
	QuoteNoticeTypeInfo    QuoteNoticeType = "INFO"
	QuoteNoticeTypeBlocked QuoteNoticeType = "BLOCKED"
)

// QuoteStatus is a Wise quote lifecycle status.
type QuoteStatus string

// Documented quote statuses.
const (
	QuoteStatusPending QuoteStatus = "pending"
	QuoteStatusActive  QuoteStatus = "active"
	QuoteStatusExpired QuoteStatus = "expired"
)

// PayIn identifies how a quote is funded.
type PayIn string

// PayIn values sent to the Wise API.
const (
	PayInBankTransfer PayIn = "BANK_TRANSFER"
	PayInBalance      PayIn = "BALANCE"
	PayInCard         PayIn = "CARD"
)

// PayOut identifies how a quote is paid out.
type PayOut string

// PayOut values sent to the Wise API.
const (
	PayOutBankTransfer PayOut = "BANK_TRANSFER"
	PayOutBalance      PayOut = "BALANCE"
	PayOutSwift        PayOut = "SWIFT"
	PayOutSwiftOur     PayOut = "SWIFT_OUR"
)

// CreateQuoteRequest parameters for creating an authenticated or
// unauthenticated quote. Exactly one of SourceAmount or TargetAmount must be
// set, and its currency must match the corresponding currency field.
type CreateQuoteRequest struct {
	SourceCurrency Currency
	TargetCurrency Currency
	SourceAmount   *Money // Optional; sets sourceAmount on the wire.
	TargetAmount   *Money // Optional; sets targetAmount on the wire.
	PreferredPayIn PayIn  // Optional; defaults to BANK_TRANSFER.
	PayOut         PayOut // Optional; defaults to BANK_TRANSFER for authenticated quotes.
	TargetAccount  RecipientID
}

// Recipient is the parsed representation of a Wise recipient account.
//
// Details contains the currency-specific account fields (e.g. sortCode,
// accountNumber, iban) as string values. Because the required fields differ by
// currency and route, consumers should use the account-requirements endpoints
// to discover the exact fields for their corridor.
type Recipient struct {
	ID                RecipientID
	AccountHolderName string
	Currency          Currency
	Country           string
	Type              string
	Details           map[string]string
	Active            bool
}

// CreateRecipientRequest parameters for creating a recipient account.
type CreateRecipientRequest struct {
	ProfileID         ProfileID
	Currency          Currency
	Type              string
	AccountHolderName string
	Details           map[string]string
	OwnedByCustomer   bool
}

// CreateTransferRequest parameters for creating a standard transfer.
//
// customerTransactionId is required for idempotency. QuoteID comes from an
// authenticated quote; TargetAccount is the recipient account ID returned by
// CreateRecipient or ListRecipients. SourceAccount is the optional refund
// recipient account ID.
type CreateTransferRequest struct {
	QuoteID                           QuoteID
	TargetAccount                     RecipientID
	SourceAccount                     RecipientID // Optional refund recipient account.
	CustomerTransactionID             string
	Reference                         string
	SourceOfFunds                     string
	TransferPurpose                   string
	TransferPurposeInvoiceNumber      string
	TransferPurposeSubTransferPurpose string
}

// ValidateTransferRequirementsRequest asks Wise which dynamic transfer
// details are required for a specific quote and recipient combination.
// Requirements vary by currency route, transfer amount, and regulatory
// region; calling this before CreateTransfer avoids delays caused by
// missing details.
//
// targetAccount and quoteUuid are required. customerTransactionId is
// optional; originatorLegalEntityType is required from March 2026 for
// Correspondent Send integrations (PRIVATE or BUSINESS).
type ValidateTransferRequirementsRequest struct {
	TargetAccount             RecipientID
	QuoteID                   QuoteID
	CustomerTransactionID     string
	OriginatorLegalEntityType string
	Details                   TransferRequirementsDetails // Optional; populated from the response.
}

// TransferRequirementsDetails mirrors the free-form details block accepted
// by the validation endpoint.
type TransferRequirementsDetails struct {
	Reference                         string
	SourceOfFunds                     string
	SourceOfFundsOther                string
	TransferPurpose                   string
	TransferPurposeSubTransferPurpose string
	TransferPurposeInvoiceNumber      string
	TransferNature                    string
}

// TransferRequirement is the parsed representation of one dynamic form from
// the transfer-requirements response.
type TransferRequirement struct {
	Type   string
	Fields []TransferRequirementForm
}

// TransferRequirementForm is a labelled group of related form fields.
type TransferRequirementForm struct {
	Name  string
	Group []TransferRequirementField
}

// TransferRequirementField describes a single dynamic form control.
// ValuesAllowed is populated only for select fields.
type TransferRequirementField struct {
	Key                         string
	Name                        string
	Type                        string
	RefreshRequirementsOnChange bool
	Required                    bool
	DisplayFormat               *string
	Example                     *string
	MinLength                   *int32
	MaxLength                   *int32
	ValidationRegexp            *string
	ValuesAllowed               []TransferRequirementValue
}

// TransferRequirementValue is one allowed value of a select field.
type TransferRequirementValue struct {
	Key  string
	Name string
}
