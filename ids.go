package wise

import (
	id "github.com/larsartmann/go-branded-id"
)

// ProfileBrand is a phantom type for ProfileID.
type ProfileBrand struct{}

// BalanceBrand is a phantom type for BalanceID.
type BalanceBrand struct{}

// TransactionBrand is a phantom type for TransactionID.
type TransactionBrand struct{}

// TransferBrand is a phantom type for TransferID.
type TransferBrand struct{}

// RecipientBrand is a phantom type for RecipientID.
type RecipientBrand struct{}

// QuoteBrand is a phantom type for QuoteID.
type QuoteBrand struct{}

// ProfileID is a strongly-typed identifier for Wise profiles.
type ProfileID = id.ID[ProfileBrand, int64]

// BalanceID is a strongly-typed identifier for Wise balances.
type BalanceID = id.ID[BalanceBrand, int64]

// TransactionID is a strongly-typed identifier for Wise transactions.
type TransactionID = id.ID[TransactionBrand, string]

// TransferID is a strongly-typed identifier for Wise transfers.
type TransferID = id.ID[TransferBrand, int64]

// RecipientID is a strongly-typed identifier for Wise recipient accounts.
type RecipientID = id.ID[RecipientBrand, int64]

// QuoteID is a strongly-typed identifier for Wise quotes.
type QuoteID = id.ID[QuoteBrand, int64]

// NewProfileID creates a new ProfileID from an int64 value.
func NewProfileID(v int64) ProfileID {
	return id.NewID[ProfileBrand](v)
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

// NewQuoteID creates a new QuoteID from an int64 value.
func NewQuoteID(v int64) QuoteID {
	return id.NewID[QuoteBrand](v)
}
