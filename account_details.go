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
	if err := requireID(profileID, "wise.account.invalid_request", "profileID"); err != nil {
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
			"wise.account.parse_creation_time",
			fmt.Sprintf("parse creationTime %q", account.CreationTime),
		)
	}

	modified, err := parseWiseTimestamp(account.ModificationTime)
	if err != nil {
		return MultiCurrencyAccount{}, errorfamily.WrapCorruption(
			err,
			"wise.account.parse_modification_time",
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

// AccountDetailsStatus reports whether a set of bank account details exists
// and is usable.
type AccountDetailsStatus string

// Documented account-details statuses.
const (
	// AccountDetailsStatusAvailable means the details do not exist yet but
	// can be ordered via the account-details-orders endpoint.
	AccountDetailsStatusAvailable AccountDetailsStatus = "AVAILABLE"
	// AccountDetailsStatusActive means the details are ready to receive money.
	AccountDetailsStatusActive AccountDetailsStatus = "ACTIVE"
)

// ReceiveOptionType identifies a receive route of a details set.
type ReceiveOptionType string

// Documented receive-option types.
const (
	ReceiveOptionTypeLocal         ReceiveOptionType = "LOCAL"
	ReceiveOptionTypeInternational ReceiveOptionType = "INTERNATIONAL"
)

// BankAccountDetails is one set of receiving bank details (e.g. an EUR IBAN
// or USD routing pair) for the Multi-Currency Account. Deprecated sets are
// Wise-issued replacements of earlier details; prefer non-deprecated sets.
type BankAccountDetails struct {
	ID             *int64 // Nil for preview details that have not been issued.
	Currency       Currency
	CurrencyName   string
	Title          string
	Subtitle       string
	Status         AccountDetailsStatus
	Deprecated     bool
	ReceiveOptions []ReceiveOption
}

// ReceiveOption is one receive route (local or international) with the
// display text Wise provides for payment instructions.
type ReceiveOption struct {
	Type        ReceiveOptionType
	Title       string
	Description *ReceiveOptionDescription
}

// ReceiveOptionDescription carries the display text of a receive option.
type ReceiveOptionDescription struct {
	Title string
	Body  string
	CTA   *CallToAction
}

// CallToAction labels the key value of a receive option (e.g. "IBAN" → the
// actual IBAN string).
type CallToAction struct {
	Label   string
	Content string
}

// GetBankAccountDetails returns every set of receiving bank details for the
// profile's Multi-Currency Account
// (GET /v1/profiles/{profileId}/account-details).
func (c *Client) GetBankAccountDetails(
	ctx context.Context,
	profileID ProfileID,
) ([]BankAccountDetails, error) {
	if err := requireID(profileID, "wise.account_details.invalid_request", "profileID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/profiles/%d/account-details", profileID.Get())

	var details []raw.BankAccountDetails

	if err := c.get(ctx, path, &details); err != nil {
		return nil, fmt.Errorf("get bank account details for profile %d: %w", profileID.Get(), err)
	}

	result := make([]BankAccountDetails, 0, len(details))
	for _, detail := range details {
		mapped, mapErr := mapBankAccountDetails(detail)
		if mapErr != nil {
			return nil, fmt.Errorf("map bank account details %v: %w", detail.ID, mapErr)
		}

		result = append(result, mapped)
	}

	return result, nil
}

func mapBankAccountDetails(detail raw.BankAccountDetails) (BankAccountDetails, error) {
	currency, err := NewCurrency(detail.Currency.Code)
	if err != nil {
		return BankAccountDetails{}, errorfamily.WrapCorruption(
			err,
			"wise.account_details.parse_currency",
			fmt.Sprintf("currency %q", detail.Currency.Code),
		)
	}

	receiveOptions := make([]ReceiveOption, 0, len(detail.ReceiveOptions))
	for _, option := range detail.ReceiveOptions {
		receiveOptions = append(receiveOptions, mapReceiveOption(option))
	}

	return BankAccountDetails{
		ID:             detail.ID,
		Currency:       currency,
		CurrencyName:   detail.Currency.Name,
		Title:          detail.Title,
		Subtitle:       detail.Subtitle,
		Status:         AccountDetailsStatus(detail.Status),
		Deprecated:     detail.Deprecated,
		ReceiveOptions: receiveOptions,
	}, nil
}

func mapReceiveOption(option raw.BankReceiveOption) ReceiveOption {
	result := ReceiveOption{
		Type:        ReceiveOptionType(option.Type),
		Title:       option.Title,
		Description: nil,
	}

	if option.Description == nil {
		return result
	}

	var cta *CallToAction

	if option.Description.CTA != nil {
		cta = &CallToAction{
			Label:   option.Description.CTA.Label,
			Content: option.Description.CTA.Content,
		}
	}

	result.Description = &ReceiveOptionDescription{
		Title: option.Description.Title,
		Body:  option.Description.Body,
		CTA:   cta,
	}

	return result
}
