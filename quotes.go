package wise

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	id "github.com/larsartmann/go-branded-id"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// CreateUnauthenticatedQuote creates a quote without an API token — a public
// rate preview. The returned quote has no ID and cannot be used to create a
// transfer; use CreateQuote for that.
func (c *Client) CreateUnauthenticatedQuote(
	ctx context.Context,
	req CreateQuoteRequest,
) (*Quote, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	var quote raw.Quote

	err := c.post(ctx, "/v3/quotes", req.toWire(false), &quote)
	if err != nil {
		return nil, fmt.Errorf("create unauthenticated quote %s-%s: %w",
			req.SourceCurrency, req.TargetCurrency, err)
	}

	return mapQuote(quote, ProfileID{})
}

// CreateQuote creates an authenticated quote for a profile. The quote locks the
// mid-market rate for 30 minutes and can be used to create a transfer.
func (c *Client) CreateQuote(
	ctx context.Context,
	profileID ProfileID,
	req CreateQuoteRequest,
) (*Quote, error) {
	if err := req.validateAuthenticated(profileID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v3/profiles/%d/quotes", profileID.Get())

	var quote raw.Quote

	err := c.post(ctx, path, req.toWire(true), &quote)
	if err != nil {
		return nil, fmt.Errorf("create quote for profile %d: %w", profileID.Get(), err)
	}

	return mapQuote(quote, profileID)
}

// GetQuote returns an existing authenticated quote by ID.
func (c *Client) GetQuote(
	ctx context.Context,
	profileID ProfileID,
	quoteID QuoteID,
) (*Quote, error) {
	if profileID.Get() == 0 {
		return nil, errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"profileID is required",
		)
	}

	if quoteID.Get() == "" {
		return nil, errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"quoteID is required",
		)
	}

	path := fmt.Sprintf("/v3/profiles/%d/quotes/%s", profileID.Get(), quoteID.Get())

	var quote raw.Quote

	if err := c.get(ctx, path, &quote); err != nil {
		return nil, fmt.Errorf("get quote %s for profile %d: %w", quoteID.Get(), profileID.Get(), err)
	}

	return mapQuote(quote, profileID)
}

// GetQuoteAccountRequirements returns the recipient-account fields required
// for the currency corridor of an authenticated quote, one dynamic form per
// available payout route. It bridges quotes to recipients: Type identifies
// the route (CreateRecipientRequest.Type), and the Fields forms describe
// which Details keys that route needs, with allowed select values and
// validation metadata.
//
// The SDK sends Accept-Minor-Version: 1, as Wise recommends for all new
// integrations, so recipient name and email fields are included.
func (c *Client) GetQuoteAccountRequirements(
	ctx context.Context,
	req QuoteAccountRequirementsRequest,
) ([]AccountRequirement, error) {
	if req.QuoteID.Get() == "" {
		return nil, errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"quoteID is required",
		)
	}

	query := func() string {
		if req.OriginatorLegalEntityType == "" {
			return ""
		}

		return url.Values{"originatorLegalEntityType": []string{req.OriginatorLegalEntityType}}.Encode()
	}

	path := fmt.Sprintf("/v1/quotes/%s/account-requirements", req.QuoteID.Get())

	var requirements []raw.AccountRequirement

	err := c.getWithQueryHeaders(ctx, path, query,
		map[string]string{"Accept-Minor-Version": "1"}, &requirements)
	if err != nil {
		return nil, fmt.Errorf("get account requirements for quote %s: %w", req.QuoteID.Get(), err)
	}

	result := make([]AccountRequirement, 0, len(requirements))
	for _, requirement := range requirements {
		result = append(result, mapAccountRequirement(requirement))
	}

	return result, nil
}

// mapAccountRequirement converts a raw account requirement into the parsed
// AccountRequirement type, reusing the shared dynamic-form mappers. The wire
// form contains only strings, so no parse failures are possible.
func mapAccountRequirement(requirement raw.AccountRequirement) AccountRequirement {
	fields := make([]TransferRequirementForm, 0, len(requirement.Fields))
	for _, form := range requirement.Fields {
		fields = append(fields, mapTransferRequirementForm(form))
	}

	return AccountRequirement{
		Type:      requirement.Type,
		Title:     requirement.Title,
		UsageInfo: requirement.UsageInfo,
		Fields:    fields,
	}
}

// RefreshQuoteAccountRequirements completes the two-pass recipient flow:
// after GetQuoteAccountRequirements returns a field with
// RefreshRequirementsOnChange=true (e.g. legalEntityType or a country
// selector), submit the recipient form with that field updated and Wise
// returns the revised requirements for the chosen route — for example
// selecting the US reveals the address.state field.
//
// The recipient payload does not need to be complete: Wise resolves the
// dependent form, not the account. Like the GET, the SDK sends
// Accept-Minor-Version: 1 so recipient name and email fields are included.
func (c *Client) RefreshQuoteAccountRequirements(
	ctx context.Context,
	req RefreshQuoteAccountRequirementsRequest,
) ([]AccountRequirement, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	query := func() string {
		if req.OriginatorLegalEntityType == "" {
			return ""
		}

		return url.Values{"originatorLegalEntityType": []string{req.OriginatorLegalEntityType}}.Encode()
	}

	path := fmt.Sprintf("/v1/quotes/%s/account-requirements", req.QuoteID.Get())

	var requirements []raw.AccountRequirement

	err := c.request(ctx, http.MethodPost, path, query, req.toWire(), &requirements,
		map[string]string{"Accept-Minor-Version": "1"})
	if err != nil {
		return nil, fmt.Errorf("refresh account requirements for quote %s: %w", req.QuoteID.Get(), err)
	}

	result := make([]AccountRequirement, 0, len(requirements))
	for _, requirement := range requirements {
		result = append(result, mapAccountRequirement(requirement))
	}

	return result, nil
}

func (r RefreshQuoteAccountRequirementsRequest) validate() error {
	if r.QuoteID.Get() == "" {
		return errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"quoteID is required",
		)
	}

	if r.Recipient.Currency == "" {
		return errorfamily.NewRejection(
			"wise.recipient.invalid_request",
			"recipient currency is required",
		)
	}

	if r.Recipient.Type == "" {
		return errorfamily.NewRejection(
			"wise.recipient.invalid_request",
			"recipient type is required",
		)
	}

	if len(r.Recipient.Details) == 0 {
		return errorfamily.NewRejection(
			"wise.recipient.invalid_request",
			"recipient details are required (at least the field that triggered the refresh)",
		)
	}

	return nil
}

// toWire renders the in-progress recipient form for the account-requirements
// refresh endpoint. Unlike CreateRecipientRequest.toWire it omits fields the
// caller has not filled in yet: an empty accountHolderName or zero profile
// would misrepresent the form's state to Wise's dependent-field resolution.
func (r RefreshQuoteAccountRequirementsRequest) toWire() map[string]any {
	body := map[string]any{
		"currency": string(r.Recipient.Currency),
		"type":     r.Recipient.Type,
		"details":  r.Recipient.Details,
	}

	if r.Recipient.ProfileID.Get() != 0 {
		body["profile"] = r.Recipient.ProfileID.Get()
	}

	if r.Recipient.AccountHolderName != "" {
		body["accountHolderName"] = r.Recipient.AccountHolderName
	}

	if r.Recipient.OwnedByCustomer {
		body["ownedByCustomer"] = true
	}

	return body
}

func (r CreateQuoteRequest) validate() error {
	return r.validateCommon()
}

func (r CreateQuoteRequest) validateAuthenticated(profileID ProfileID) error {
	if profileID.Get() == 0 {
		return errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"profileID is required",
		)
	}

	return r.validateCommon()
}

func (r CreateQuoteRequest) validateCommon() error {
	if r.SourceCurrency == "" {
		return errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"sourceCurrency is required",
		)
	}

	if r.TargetCurrency == "" {
		return errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"targetCurrency is required",
		)
	}

	if r.SourceCurrency == r.TargetCurrency {
		return errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"sourceCurrency and targetCurrency must be different",
		)
	}

	if r.SourceAmount == nil && r.TargetAmount == nil {
		return errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"either sourceAmount or targetAmount is required",
		)
	}

	if r.SourceAmount != nil && r.TargetAmount != nil {
		return errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"only one of sourceAmount or targetAmount may be set",
		)
	}

	if r.SourceAmount != nil && r.SourceAmount.Currency != r.SourceCurrency {
		return errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"sourceAmount currency must match sourceCurrency",
		)
	}

	if r.TargetAmount != nil && r.TargetAmount.Currency != r.TargetCurrency {
		return errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"targetAmount currency must match targetCurrency",
		)
	}

	return nil
}

func (r CreateQuoteRequest) toWire(authenticated bool) map[string]any {
	body := map[string]any{
		"sourceCurrency": string(r.SourceCurrency),
		"targetCurrency": string(r.TargetCurrency),
	}

	if r.SourceAmount != nil {
		body["sourceAmount"] = centsToMajor(r.SourceAmount.Cents)
	}

	if r.TargetAmount != nil {
		body["targetAmount"] = centsToMajor(r.TargetAmount.Cents)
	}

	if authenticated {
		if r.PreferredPayIn != "" {
			body["preferredPayIn"] = string(r.PreferredPayIn)
		}

		if r.PayOut != "" {
			body["payOut"] = string(r.PayOut)
		}

		if r.TargetAccount.Get() != 0 {
			body["targetAccount"] = r.TargetAccount.Get()
		}
	}

	return body
}

func mapQuote(quote raw.Quote, profileID ProfileID) (*Quote, error) {
	created, expiration, timeErr := parseQuoteCreated(quote)
	if timeErr != nil {
		return nil, timeErr
	}

	source, target, monetaryErr := mapQuoteMonetary(quote)
	if monetaryErr != nil {
		return nil, monetaryErr
	}

	options, optionErr := mapQuotePaymentOptions(quote.PaymentOptions)
	if optionErr != nil {
		return nil, optionErr
	}

	notices := mapQuoteNotices(quote.Notices)

	return &Quote{
		ID:                            id.NewID[QuoteBrand](quote.ID),
		Source:                        source,
		Target:                        target,
		PayIn:                         PayIn(quote.PayIn),
		PayOut:                        PayOut(quote.PayOut),
		Rate:                          quote.Rate,
		Created:                       created,
		ExpirationTime:                expiration,
		Status:                        QuoteStatus(quote.Status),
		Profile:                       profileID,
		RateType:                      QuoteRateType(quote.RateType),
		ProvidedAmountType:            QuoteProvidedAmountType(quote.ProvidedAmountType),
		GuaranteedTargetAmountAllowed: quote.GuaranteedTargetAmountAllowed,
		GuaranteedTargetAmount:        quote.GuaranteedTargetAmount,
		PaymentOptions:                options,
		Notices:                       notices,
	}, nil
}

// parseQuoteCreated parses a quote's createdTime and expirationTime. Both
// must be present and parseable for the quote to be usable.
func parseQuoteCreated(quote raw.Quote) (time.Time, time.Time, error) {
	created, err := parseWiseTimestamp(quote.CreatedTime)
	if err != nil {
		return time.Time{}, time.Time{}, errorfamily.WrapCorruption(
			err,
			"wise.quote.parse_created",
			fmt.Sprintf("parse createdTime %q", quote.CreatedTime),
		)
	}

	expiration, err := parseWiseTimestamp(quote.ExpirationTime)
	if err != nil {
		return time.Time{}, time.Time{}, errorfamily.WrapCorruption(
			err,
			"wise.quote.parse_expiration",
			fmt.Sprintf("parse expirationTime %q", quote.ExpirationTime),
		)
	}

	return created, expiration, nil
}

// mapQuoteMonetary converts a quote's source/target amounts and currencies
// into Money value objects with validated currencies.
func mapQuoteMonetary(quote raw.Quote) (Money, Money, error) {
	sourceCurrency, err := NewCurrency(quote.SourceCurrency)
	if err != nil {
		return Money{}, Money{}, errorfamily.WrapCorruption(
			err,
			"wise.quote.parse_source_currency",
			fmt.Sprintf("source currency %q", quote.SourceCurrency),
		)
	}

	targetCurrency, err := NewCurrency(quote.TargetCurrency)
	if err != nil {
		return Money{}, Money{}, errorfamily.WrapCorruption(
			err,
			"wise.quote.parse_target_currency",
			fmt.Sprintf("target currency %q", quote.TargetCurrency),
		)
	}

	return Money{Cents: majorToCents(quote.SourceAmount), Currency: sourceCurrency},
		Money{Cents: majorToCents(quote.TargetAmount), Currency: targetCurrency},
		nil
}

func mapQuotePaymentOptions(options []raw.QuotePaymentOption) ([]QuotePaymentOption, error) {
	if options == nil {
		return nil, nil
	}

	result := make([]QuotePaymentOption, 0, len(options))
	for _, option := range options {
		delivery, err := parseWiseTimestamp(option.EstimatedDelivery)
		if err != nil {
			return nil, errorfamily.WrapCorruption(
				err,
				"wise.quote.parse_payment_option_delivery",
				fmt.Sprintf("parse estimatedDelivery %q", option.EstimatedDelivery),
			)
		}

		sourceCurrency, err := NewCurrency(option.SourceCurrency)
		if err != nil {
			return nil, errorfamily.WrapCorruption(
				err,
				"wise.quote.parse_payment_option_source",
				fmt.Sprintf("payment option source currency %q", option.SourceCurrency),
			)
		}

		targetCurrency, err := NewCurrency(option.TargetCurrency)
		if err != nil {
			return nil, errorfamily.WrapCorruption(
				err,
				"wise.quote.parse_payment_option_target",
				fmt.Sprintf("payment option target currency %q", option.TargetCurrency),
			)
		}

		result = append(result, QuotePaymentOption{
			Disabled:                   option.Disabled,
			EstimatedDelivery:          delivery,
			FormattedEstimatedDelivery: option.FormattedEstimatedDelivery,
			Fee: QuoteFee{
				TransferWise: option.Fee.TransferWise,
				PayIn:        option.Fee.PayIn,
				Discount:     option.Fee.Discount,
				Partner:      option.Fee.Partner,
				Total:        option.Fee.Total,
			},
			Source:        Money{Cents: majorToCents(option.SourceAmount), Currency: sourceCurrency},
			Target:        Money{Cents: majorToCents(option.TargetAmount), Currency: targetCurrency},
			PayIn:         PayIn(option.PayIn),
			PayOut:        PayOut(option.PayOut),
			PayInProduct:  option.PayInProduct,
			FeePercentage: option.FeePercentage,
		})
	}

	return result, nil
}

func mapQuoteNotices(notices []raw.QuoteNotice) []QuoteNotice {
	if notices == nil {
		return nil
	}

	result := make([]QuoteNotice, 0, len(notices))
	for _, notice := range notices {
		result = append(result, QuoteNotice{
			Text: notice.Text,
			Link: notice.Link,
			Type: QuoteNoticeType(notice.Type),
		})
	}

	return result
}

func centsToMajor(cents int64) float64 {
	return float64(cents) / centsPerUnit
}

func majorToCents(major float64) int64 {
	return int64(major * centsPerUnit)
}
