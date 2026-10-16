package wise

import (
	"context"
	"fmt"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// OriginatorLegalEntityType discriminates the originator oneOf on the wire.
type OriginatorLegalEntityType string

// Documented originator legal-entity types.
const (
	// OriginatorLegalEntityTypePrivate is a natural person.
	OriginatorLegalEntityTypePrivate OriginatorLegalEntityType = "PRIVATE"
	// OriginatorLegalEntityTypeBusiness is a registered business.
	OriginatorLegalEntityTypeBusiness OriginatorLegalEntityType = "BUSINESS"
)

// OriginatorName captures the originator's name for compliance. For PRIVATE
// originators the given/family split is what Wise validates; for BUSINESS
// originators set FullName to the registered company name.
type OriginatorName struct {
	GivenName      string
	MiddleNames    []string
	FamilyName     string
	PatronymicName string
	FullName       string
}

// OriginatorAddress is the originator's postal address. StateCode is
// required for US/CA/BR/AU addresses.
type OriginatorAddress struct {
	FirstLine   string
	City        string
	StateCode   string
	CountryCode string
	PostCode    string
}

// Originator identifies who is really behind a third-party transfer. The
// same Reference must be reused for every transfer of one originator (Wise
// keys compliance on it). DateOfBirth applies to PRIVATE originators;
// BusinessRegistrationCode/Date to BUSINESS ones.
type Originator struct {
	LegalEntityType          OriginatorLegalEntityType
	Reference                string
	Name                     OriginatorName
	DateOfBirth              time.Time
	Nationality              string
	BusinessRegistrationCode string
	BusinessRegistrationDate time.Time
	AccountDetails           string
	Address                  OriginatorAddress
}

// CreateThirdPartyTransferRequest is the body of a third-party (originator)
// transfer: money moved on behalf of an originator, funded through your own
// quote. OriginalTransferID is the idempotency key (Wise deduplicates on
// it); QuoteID is a quoteUUID.
type CreateThirdPartyTransferRequest struct {
	ProfileID          ProfileID
	OriginalTransferID string
	TargetAccount      RecipientID
	SourceAccount      *RecipientID
	QuoteID            QuoteID
	Reference          string
	Originator         Originator
}

// ThirdPartyTransfer is the parsed originator transfer. The amounts are the
// wire's source/target values as Money.
type ThirdPartyTransfer struct {
	ID                 TransferID
	UserID             UserID
	TargetAccount      RecipientID
	SourceAccount      *RecipientID
	QuoteID            QuoteID
	Status             TransferStatus
	Rate               float64
	Reference          string
	HasActiveIssues    bool
	Source             Money
	Target             Money
	OriginalTransferID string
	Created            time.Time
}

// CreateThirdPartyTransfer creates an originator transfer
// (POST /v1/profiles/{profileId}/third-party-transfers). Requires the
// originator data block — a standard CreateTransfer cannot carry one.
func (c *Client) CreateThirdPartyTransfer(
	ctx context.Context,
	req CreateThirdPartyTransferRequest,
) (*ThirdPartyTransfer, error) {
	if err := requireID(req.ProfileID, errorCodeTransferInvalidRequest, "profileID"); err != nil {
		return nil, err
	}

	if err := requireID(req.TargetAccount, errorCodeTransferInvalidRequest, "targetAccount"); err != nil {
		return nil, err
	}

	if err := requireID(req.QuoteID, errorCodeTransferInvalidRequest, "quoteUuid"); err != nil {
		return nil, err
	}

	if req.OriginalTransferID == "" {
		return nil, errorfamily.NewRejection(errorCodeTransferInvalidRequest, "originalTransferId is required")
	}

	body := map[string]any{
		"originalTransferId": req.OriginalTransferID,
		"targetAccount":      req.TargetAccount.Get(),
		"quote":              req.QuoteID.Get(),
		"details":            map[string]string{"reference": req.Reference},
		"originator":         originatorToWire(req.Originator),
	}

	if req.SourceAccount != nil {
		body["sourceAccount"] = req.SourceAccount.Get()
	}

	path := fmt.Sprintf("/v1/profiles/%d/third-party-transfers", req.ProfileID.Get())

	var transfer raw.ThirdPartyTransfer

	if err := c.post(ctx, path, body, &transfer); err != nil {
		return nil, fmt.Errorf("create third-party transfer for profile %d: %w", req.ProfileID.Get(), err)
	}

	result, err := toThirdPartyTransfer("map created third-party transfer", transfer)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetThirdPartyTransfer returns an originator transfer by ID
// (GET /v1/profiles/{profileId}/third-party-transfers/{transferId}).
func (c *Client) GetThirdPartyTransfer(
	ctx context.Context,
	profileID ProfileID,
	transferID TransferID,
) (*ThirdPartyTransfer, error) {
	if err := requireID(profileID, errorCodeTransferInvalidRequest, "profileID"); err != nil {
		return nil, err
	}

	if err := requireID(transferID, errorCodeTransferInvalidRequest, "transferID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/profiles/%d/third-party-transfers/%d", profileID.Get(), transferID.Get())

	var transfer raw.ThirdPartyTransfer

	if err := c.get(ctx, path, &transfer); err != nil {
		return nil, fmt.Errorf("get third-party transfer %d: %w", transferID.Get(), err)
	}

	result, err := toThirdPartyTransfer("map third-party transfer", transfer)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func originatorToWire(originator Originator) map[string]any {
	wire := map[string]any{
		"legalEntityType": string(originator.LegalEntityType),
		"reference":       originator.Reference,
		"name":            originatorNameToWire(originator.Name),
		"address":         originatorAddressToWire(originator.Address),
	}

	if !originator.DateOfBirth.IsZero() {
		wire["dateOfBirth"] = originator.DateOfBirth.Format("2006-01-02")
	}

	if originator.Nationality != "" {
		wire["nationality"] = originator.Nationality
	}

	if originator.BusinessRegistrationCode != "" {
		wire["businessRegistrationCode"] = originator.BusinessRegistrationCode
	}

	if !originator.BusinessRegistrationDate.IsZero() {
		wire["businessRegistrationDate"] = originator.BusinessRegistrationDate.Format("2006-01-02")
	}

	if originator.AccountDetails != "" {
		wire["accountDetails"] = originator.AccountDetails
	}

	return wire
}

func originatorNameToWire(name OriginatorName) map[string]any {
	wire := map[string]any{}

	if name.GivenName != "" {
		wire["givenName"] = name.GivenName
	}

	if len(name.MiddleNames) > 0 {
		wire["middleNames"] = name.MiddleNames
	}

	if name.FamilyName != "" {
		wire["familyName"] = name.FamilyName
	}

	if name.PatronymicName != "" {
		wire["patronymicName"] = name.PatronymicName
	}

	if name.FullName != "" {
		wire["fullName"] = name.FullName
	}

	return wire
}

func originatorAddressToWire(address OriginatorAddress) map[string]any {
	wire := map[string]any{
		"firstLine":   address.FirstLine,
		"countryCode": address.CountryCode,
	}

	if address.City != "" {
		wire["city"] = address.City
	}

	if address.StateCode != "" {
		wire["stateCode"] = address.StateCode
	}

	if address.PostCode != "" {
		wire["postCode"] = address.PostCode
	}

	return wire
}

func toThirdPartyTransfer(label string, transfer raw.ThirdPartyTransfer) (ThirdPartyTransfer, error) {
	created, err := parseWiseTimestamp(transfer.Created)
	if err != nil {
		return ThirdPartyTransfer{}, errorfamily.WrapCorruption(
			err,
			errorCodeTransferParseCreated,
			fmt.Sprintf("%s: parse created %q", label, transfer.Created),
		)
	}

	sourceCurrency, err := NewCurrency(transfer.SourceCurrency)
	if err != nil {
		return ThirdPartyTransfer{}, errorfamily.WrapCorruption(
			err,
			errorCodeTransferParseCurrency,
			fmt.Sprintf("%s: source currency %q", label, transfer.SourceCurrency),
		)
	}

	targetCurrency, err := NewCurrency(transfer.TargetCurrency)
	if err != nil {
		return ThirdPartyTransfer{}, errorfamily.WrapCorruption(
			err,
			errorCodeTransferParseCurrency,
			fmt.Sprintf("%s: target currency %q", label, transfer.TargetCurrency),
		)
	}

	var sourceAccount *RecipientID

	if transfer.SourceAccount != nil {
		account := NewRecipientID(*transfer.SourceAccount)
		sourceAccount = &account
	}

	return ThirdPartyTransfer{
		ID:                 NewTransferID(transfer.ID),
		UserID:             NewUserID(transfer.User),
		TargetAccount:      NewRecipientID(transfer.TargetAccount),
		SourceAccount:      sourceAccount,
		QuoteID:            NewQuoteID(transfer.Quote),
		Status:             TransferStatus(transfer.Status),
		Rate:               transfer.Rate,
		Reference:          transfer.Details.Reference,
		HasActiveIssues:    transfer.HasActiveIssues,
		Source:             Money{Cents: majorToCents(transfer.SourceValue), Currency: sourceCurrency},
		Target:             Money{Cents: majorToCents(transfer.TargetValue), Currency: targetCurrency},
		OriginalTransferID: transfer.OriginalTransferID,
		Created:            created,
	}, nil
}
