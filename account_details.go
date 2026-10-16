package wise

import (
	"context"
	"fmt"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// MultiCurrencyAccount is the profile's Wise account container — the holder
// of the currency balances and their receiving bank details.
//
// RecipientID is the account's own recipient: create a transfer to it to
// top up the matching balance.
type MultiCurrencyAccount struct {
	ID          AccountID
	ProfileID   ProfileID
	RecipientID RecipientID
	Created     time.Time
	Modified    time.Time
	Active      bool
	Eligible    bool
}

// GetMultiCurrencyAccount returns the profile's Multi-Currency Account
// (GET /v1/profiles/{profileId}/multi-currency-account).
func (c *Client) GetMultiCurrencyAccount(
	ctx context.Context,
	profileID ProfileID,
) (*MultiCurrencyAccount, error) {
	if err := requireID(profileID, errorCodeAccountInvalidRequest, "profileID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/profiles/%d/multi-currency-account", profileID.Get())

	var account raw.MultiCurrencyAccount

	if err := c.get(ctx, path, &account); err != nil {
		return nil, fmt.Errorf("get multi-currency account for profile %d: %w", profileID.Get(), err)
	}

	result, err := mapMultiCurrencyAccount(account)
	if err != nil {
		return nil, fmt.Errorf("map multi-currency account %d: %w", account.ID, err)
	}

	return &result, nil
}

func mapMultiCurrencyAccount(account raw.MultiCurrencyAccount) (MultiCurrencyAccount, error) {
	created, err := parseWiseTimestamp(account.CreationTime)
	if err != nil {
		return MultiCurrencyAccount{}, errorfamily.WrapCorruption(
			err,
			errorCodeAccountParseCreationTime,
			fmt.Sprintf("parse creationTime %q", account.CreationTime),
		)
	}

	modified, err := parseWiseTimestamp(account.ModificationTime)
	if err != nil {
		return MultiCurrencyAccount{}, errorfamily.WrapCorruption(
			err,
			errorCodeAccountParseModificationTime,
			fmt.Sprintf("parse modificationTime %q", account.ModificationTime),
		)
	}

	return MultiCurrencyAccount{
		ID:          NewAccountID(account.ID),
		ProfileID:   NewProfileID(account.ProfileID),
		RecipientID: NewRecipientID(account.RecipientID),
		Created:     created,
		Modified:    modified,
		Active:      account.Active,
		Eligible:    account.Eligible,
	}, nil
}

// AccountDetailsState reports whether a set of receiving bank details is
// usable. The 2027Q1 surface replaced the old AVAILABLE/ACTIVE statuses with
// this two-value state (an INACTIVE set exists but must not receive money).
type AccountDetailsState string

// Documented account-details states.
const (
	// AccountDetailsStateActive means the details are ready to receive money.
	AccountDetailsStateActive AccountDetailsState = "ACTIVE"
	// AccountDetailsStateInactive means the details exist but are switched off.
	AccountDetailsStateInactive AccountDetailsState = "INACTIVE"
)

// AccountDetails is one set of receiving bank details (an IBAN, a UK sort
// code/account pair, or a US routing/account pair) on the profile's
// Multi-Currency Account. Wise relocated the surface in the 2027Q1 roll: it
// is keyed by the details' own ID now, not by profile.
type AccountDetails struct {
	ID                AccountDetailsID
	ProfileID         ProfileID
	State             AccountDetailsState
	AccountHolderName string
	Country           string // ISO 3166-1 alpha-2
	Identification    AccountIdentification
}

// AccountIdentification is the country/scheme-specific account number block.
// Exactly one member is present; anything else is a corrupt response.
//
// art-dupl:accept parsed domain types intentionally mirror the tagged wire types in internal/raw; the tag-free domain layer decouples callers from the JSON shapes
type AccountIdentification struct {
	BE *BEAccountIdentification
	GB *GBAccountIdentification
	US *USAccountIdentification
}

// BEAccountIdentification is the Belgian (SEPA) details format.
type BEAccountIdentification struct {
	BIC                string
	IBAN               string
	InstitutionName    string
	InstitutionAddress *ISO20022Address
}

// GBAccountIdentification is the UK details format.
type GBAccountIdentification struct {
	BIC                string
	IBAN               string
	SortCode           string
	AccountNumber      string
	InstitutionName    string
	InstitutionAddress *ISO20022Address
}

// USAccountIdentification is the US (ACH) details format.
type USAccountIdentification struct {
	BIC                string
	RoutingNumber      string
	AccountNumber      string
	InstitutionName    string
	InstitutionAddress *ISO20022Address
}

// ISO20022Address is Wise's structured postal address for the instructing
// institution, with optional free-form address lines.
type ISO20022Address struct {
	StreetName         string
	BuildingNumber     string
	TownName           string
	PostCode           string
	CountrySubDivision string
	Country            string
	AddressLines       []string
}

// GetBankAccountDetails returns one set of receiving bank details by its own
// ID (GET /{version}/balances/account-details/{id}; the 2027Q1 shape that
// replaced GET /v1/profiles/{profileId}/account-details).
func (c *Client) GetBankAccountDetails(
	ctx context.Context,
	id AccountDetailsID,
) (*AccountDetails, error) {
	if err := requireID(id, errorCodeAccountDetailsInvalidRequest, "id"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/%s/balances/account-details/%d", accountDetailsAPIVersion, id.Get())

	var detail raw.AccountDetails

	if err := c.get(ctx, path, &detail); err != nil {
		return nil, fmt.Errorf("get bank account details %d: %w", id.Get(), err)
	}

	result, err := mapAccountDetails(detail)
	if err != nil {
		return nil, fmt.Errorf("map bank account details %d: %w", detail.ID, err)
	}

	return &result, nil
}

func mapAccountDetails(detail raw.AccountDetails) (AccountDetails, error) {
	members := countIdentificationMembers(detail.AccountIdentification)
	if members != 1 {
		return AccountDetails{}, errorfamily.WrapCorruption(
			ErrAccountIdentificationCardinality,
			errorCodeAccountDetailsParseIdentification,
			fmt.Sprintf("expected exactly one identification member, found %d (country %q)", members, detail.Country),
		)
	}

	return AccountDetails{
		ID:                NewAccountDetailsID(detail.ID),
		ProfileID:         NewProfileID(detail.ProfileID),
		State:             AccountDetailsState(detail.State),
		AccountHolderName: detail.AccountHolderName,
		Country:           detail.Country,
		Identification:    mapAccountIdentification(detail.AccountIdentification),
	}, nil
}

func countIdentificationMembers(ident raw.AccountIdentification) int {
	count := 0

	for _, present := range []bool{ident.BE != nil, ident.GB != nil, ident.US != nil} {
		if present {
			count++
		}
	}

	return count
}

func mapAccountIdentification(ident raw.AccountIdentification) AccountIdentification {
	result := AccountIdentification{BE: nil, GB: nil, US: nil}

	if ident.BE != nil {
		result.BE = &BEAccountIdentification{
			BIC:                ident.BE.BIC,
			IBAN:               ident.BE.IBAN,
			InstitutionName:    ident.BE.InstitutionName,
			InstitutionAddress: mapISO20022Address(ident.BE.InstitutionAddress),
		}
	}

	if ident.GB != nil {
		result.GB = &GBAccountIdentification{
			BIC:                ident.GB.BIC,
			IBAN:               ident.GB.IBAN,
			SortCode:           ident.GB.SortCode,
			AccountNumber:      ident.GB.AccountNumber,
			InstitutionName:    ident.GB.InstitutionName,
			InstitutionAddress: mapISO20022Address(ident.GB.InstitutionAddress),
		}
	}

	if ident.US != nil {
		result.US = &USAccountIdentification{
			BIC:                ident.US.BIC,
			RoutingNumber:      ident.US.RoutingNumber,
			AccountNumber:      ident.US.AccountNumber,
			InstitutionName:    ident.US.InstitutionName,
			InstitutionAddress: mapISO20022Address(ident.US.InstitutionAddress),
		}
	}

	return result
}

func mapISO20022Address(addr *raw.ISO20022Address) *ISO20022Address {
	if addr == nil {
		return nil
	}

	return &ISO20022Address{
		StreetName:         addr.StreetName,
		BuildingNumber:     addr.BuildingNumber,
		TownName:           addr.TownName,
		PostCode:           addr.PostCode,
		CountrySubDivision: addr.CountrySubDivision,
		Country:            addr.Country,
		AddressLines:       addr.AddressLines,
	}
}
