package wise

import (
	"crypto/rand"
	"fmt"

	id "github.com/larsartmann/go-branded-id"
	errorfamily "github.com/larsartmann/go-error-family"
)

// ProfileBrand is a phantom type for ProfileID.
type ProfileBrand struct{}

// UserBrand is a phantom type for UserID.
type UserBrand struct{}

// BalanceBrand is a phantom type for BalanceID.
type BalanceBrand struct{}

// TransactionBrand is a phantom type for TransactionID.
type TransactionBrand struct{}

// TransferBrand is a phantom type for TransferID.
type TransferBrand struct{}

// RecipientBrand is a phantom type for RecipientID.
type RecipientBrand struct{}

// AccountBrand is a phantom type for AccountID.
type AccountBrand struct{}

// BalanceTransactionBrand is a phantom type for BalanceTransactionID.
type BalanceTransactionBrand struct{}

// QuoteBrand is a phantom type for QuoteID.
type QuoteBrand struct{}

// WebhookSubscriptionBrand is a phantom type for WebhookSubscriptionID.
type WebhookSubscriptionBrand struct{}

// ProfileID is a strongly-typed identifier for Wise profiles.
type ProfileID = id.ID[ProfileBrand, int64]

// UserID is a strongly-typed identifier for Wise user accounts.
type UserID = id.ID[UserBrand, int64]

// BalanceID is a strongly-typed identifier for Wise balances.
type BalanceID = id.ID[BalanceBrand, int64]

// TransactionID is a strongly-typed identifier for Wise transactions.
type TransactionID = id.ID[TransactionBrand, string]

// TransferID is a strongly-typed identifier for Wise transfers.
type TransferID = id.ID[TransferBrand, int64]

// RecipientID is a strongly-typed identifier for Wise recipient accounts.
type RecipientID = id.ID[RecipientBrand, int64]

// AccountID is a strongly-typed identifier for Wise multi-currency
// accounts — a distinct ID space from RecipientID even though both are
// int64 on the wire.
type AccountID = id.ID[AccountBrand, int64]

// BalanceTransactionID is a strongly-typed identifier for Wise balance
// transactions (the numeric ID of a debit or credit applied to a balance,
// e.g. the funding transaction created by FundTransfer). It is a distinct ID
// space from the composite string TransactionID of statement transactions.
type BalanceTransactionID = id.ID[BalanceTransactionBrand, int64]

// QuoteID is a strongly-typed identifier for Wise quotes.
// Note: Wise quote IDs are UUIDs (strings), unlike the integer IDs used for
// profiles, balances, transfers, and recipients.
type QuoteID = id.ID[QuoteBrand, string]

// WebhookSubscriptionID is a strongly-typed identifier for Wise webhook
// subscriptions. Like QuoteID it is a UUID string on the wire (spec:
// "UUID that uniquely identifies the subscription"), not an int64.
type WebhookSubscriptionID = id.ID[WebhookSubscriptionBrand, string]

// NewProfileID creates a new ProfileID from an int64 value.
func NewProfileID(v int64) ProfileID {
	return id.NewID[ProfileBrand](v)
}

// NewUserID creates a new UserID from an int64 value.
func NewUserID(v int64) UserID {
	return id.NewID[UserBrand](v)
}

// NewBalanceID creates a new BalanceID from an int64 value.
func NewBalanceID(v int64) BalanceID {
	return id.NewID[BalanceBrand](v)
}

// NewTransactionID creates a new TransactionID from a string value.
func NewTransactionID(v string) TransactionID {
	return id.NewID[TransactionBrand](v)
}

// NewTransferID creates a new TransferID from an int64 value.
func NewTransferID(v int64) TransferID {
	return id.NewID[TransferBrand](v)
}

// NewRecipientID creates a new RecipientID from an int64 value.
func NewRecipientID(v int64) RecipientID {
	return id.NewID[RecipientBrand](v)
}

// NewAccountID creates a new AccountID from an int64 value.
func NewAccountID(v int64) AccountID {
	return id.NewID[AccountBrand](v)
}

// NewBalanceTransactionID creates a new BalanceTransactionID from an int64
// value.
func NewBalanceTransactionID(v int64) BalanceTransactionID {
	return id.NewID[BalanceTransactionBrand](v)
}

// NewQuoteID creates a new QuoteID from a string UUID value.
func NewQuoteID(v string) QuoteID {
	return id.NewID[QuoteBrand](v)
}

// NewWebhookSubscriptionID creates a new WebhookSubscriptionID from a string
// UUID value.
func NewWebhookSubscriptionID(v string) WebhookSubscriptionID {
	return id.NewID[WebhookSubscriptionBrand](v)
}

// requireID turns a zero branded ID into a Rejection carrying the endpoint
// family's invalid-request code. field names the parameter as the Wise API
// spells it ("profileID" for most fields, but wire names like "targetAccount"
// and "quoteUuid" where the API uses them).
func requireID[B any, V comparable](identifier id.ID[B, V], code, field string) error {
	if identifier.IsZero() {
		return errorfamily.NewRejection(code, field+" is required")
	}

	return nil
}

// RFC 4122 version-4 bit layout: the version nibble and the two
// variant bits, plus their clear masks.
const (
	uuidVersion4    = 0x40
	uuidVariantRFC  = 0x80
	uuidVersionMask = 0x0f
	uuidVariantMask = 0x3f
)

// newRequestUUID generates a random RFC 4122 version-4 UUID for the
// X-idempotence-uuid request header (required by Wise on balance creation).
// Callers can pass their own key for cross-retry deduplication where the API
// accepts one.
func newRequestUUID() (string, error) {
	var uuid [16]byte
	if _, err := rand.Read(uuid[:]); err != nil {
		return "", fmt.Errorf("generate request UUID: %w", err)
	}

	uuid[6] = (uuid[6] & uuidVersionMask) | uuidVersion4
	uuid[8] = (uuid[8] & uuidVariantMask) | uuidVariantRFC

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}
