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

// UserBrand is a phantom type for UserID.
type UserBrand struct{}

// ProfileID is a strongly-typed identifier for Wise profiles.
type ProfileID = id.ID[ProfileBrand, int64]

// BalanceID is a strongly-typed identifier for Wise balances.
type BalanceID = id.ID[BalanceBrand, int64]

// TransactionID is a strongly-typed identifier for Wise transactions.
type TransactionID = id.ID[TransactionBrand, string]

// UserID is a strongly-typed identifier for Wise users.
type UserID = id.ID[UserBrand, int64]

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

// NewUserID creates a new UserID from an int64 value.
func NewUserID(v int64) UserID {
	return id.NewID[UserBrand](v)
}
