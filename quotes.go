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

	body := req.toWire()

	var quote raw.Quote

	err := c.post(ctx, "/v3/quotes", body, &quote)
	if err != nil {
		return nil, fmt.Errorf("create unauthenticated quote %s-%s: %w",
			req.SourceCurrency, req.TargetCurrency, err)
	}

	return mapQuote(quote)
}

// CreateQuote creates an authenticated quote for a profile. The quote locks the
// mid-market rate for 30 minutes and can be used to create a transfer.
func (c *Client) CreateQuote(
	ctx context.Context,
	profileID ProfileID,
	req CreateQuoteRequest,
) (*Quote, error) {
	if profileID.Get() == 0 {
		return nil, errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"profileID is required",
		)
	}

	if err := req.validate(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v3/profiles/%d/quotes", profileID.Get())
	body := req.toWire()

	var quote raw.Quote

	err := c.post(ctx, path, body, &quote)
	if err != nil {
		return nil, fmt.Errorf("create quote for profile %d: %w", profileID.Get(), err)
	}

	result, mapErr := mapQuote(quote)
	if mapErr != nil {
		return nil, mapErr
	}

	result.Profile = profileID

	return result, nil
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

	path := fmt.Sprintf("/v3/profiles/%d/quotes/%d", profileID.Get(), quoteID.Get())

	var quote raw.Quote

	if err := fetchByID(ctx, c, quoteID.Get(), "quote", path, &quote); err != nil {
		return nil, err
	}

	result, mapErr := mapQuote(quote)
	if mapErr != nil {
		return nil, mapErr
	}

	result.Profile = profileID

	return result, nil
}

func (r CreateQuoteRequest) validate() error {
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

	if err := r.validateAmounts(); err != nil {
		return err
	}

	if r.PayOut == "" {
		return errorfamily.NewRejection(
			"wise.quote.invalid_request",
			"payOut is required",
		)
	}

	return nil
}

func (r CreateQuoteRequest) validateAmounts() error {
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

func (r CreateQuoteRequest) toWire() map[string]any {
	body := map[string]any{
		"sourceCurrency": string(r.SourceCurrency),
		"targetCurrency": string(r.TargetCurrency),
		"payIn":          string(r.PayIn),
		"payOut":         string(r.PayOut),
	}

	if r.SourceAmount != nil {
		body["sourceAmount"] = centsToMajor(r.SourceAmount.Cents)
	}

	if r.TargetAmount != nil {
		body["targetAmount"] = centsToMajor(r.TargetAmount.Cents)
	}

	return body
}

func mapQuote(q raw.Quote) (*Quote, error) {
	created, err := parseWiseTimestamp(q.CreatedTime)
	if err != nil {
		return nil, errorfamily.WrapCorruption(
			err,
			"wise.quote.parse_created",
			fmt.Sprintf("parse createdTime %q", q.CreatedTime),
		)
	}

	sourceCurrency, err := NewCurrency(q.SourceCurrency)
	if err != nil {
		return nil, errorfamily.WrapCorruption(
			err,
			"wise.quote.parse_source_currency",
			fmt.Sprintf("source currency %q", q.SourceCurrency),
		)
	}

	targetCurrency, err := NewCurrency(q.TargetCurrency)
	if err != nil {
		return nil, errorfamily.WrapCorruption(
			err,
			"wise.quote.parse_target_currency",
			fmt.Sprintf("target currency %q", q.TargetCurrency),
		)
	}

	return &Quote{
		ID:     id.NewID[QuoteBrand](q.ID),
		Source: Money{Cents: majorToCents(q.SourceAmount), Currency: sourceCurrency},
		Target: Money{Cents: majorToCents(q.TargetAmount), Currency: targetCurrency},
		PayIn:  PayIn(q.PayIn),
		PayOut: PayOut(q.PayOut),
		Rate:   q.Rate,
		Created: created,
		Status: QuoteStatus(q.Status),
	}, nil
}

func centsToMajor(cents int64) float64 {
	return float64(cents) / centsPerUnit
}

func majorToCents(major float64) int64 {
	return int64(major * centsPerUnit)
}
