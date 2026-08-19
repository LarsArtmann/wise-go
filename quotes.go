package wise

import (
	"context"
	"fmt"

	id "github.com/larsartmann/go-branded-id"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// CreateUnauthenticatedQuote creates an illustrative quote without a user token.
// The returned quote has no ID and cannot be used to create a transfer.
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
	created, createdErr := parseQuoteCreated(quote)
	if createdErr != nil {
		return nil, createdErr
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
		ID:                           id.NewID[QuoteBrand](quote.ID),
		Source:                       source,
		Target:                       target,
		PayIn:                        PayIn(quote.PayIn),
		PayOut:                       PayOut(quote.PayOut),
		Rate:                         quote.Rate,
		Created:                      created,
		ExpirationTime:               expiration,
		Status:                       QuoteStatus(quote.Status),
		Profile:                      profileID,
		RateType:                     QuoteRateType(quote.RateType),
		ProvidedAmountType:           QuoteProvidedAmountType(quote.ProvidedAmountType),
		GuaranteedTargetAmountAllowed: quote.GuaranteedTargetAmountAllowed,
		GuaranteedTargetAmount:       quote.GuaranteedTargetAmount,
		PaymentOptions:               options,
		Notices:                      notices,
	}, nil
}

// parseQuoteCreated parses a quote's createdTime and expirationTime. Both
// must be present and parseable for the quote to be usable.
func parseQuoteCreated(quote raw.Quote) (created, expiration time.Time, err error) {
	created, err = parseWiseTimestamp(quote.CreatedTime)
	if err != nil {
		return time.Time{}, time.Time{}, errorfamily.WrapCorruption(
			err,
			"wise.quote.parse_created",
			fmt.Sprintf("parse createdTime %q", quote.CreatedTime),
		)
	}

	expiration, err = parseWiseTimestamp(quote.ExpirationTime)
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
func mapQuoteMonetary(quote raw.Quote) (source, target Money, err error) {
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
