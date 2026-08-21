package wise

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	id "github.com/larsartmann/go-branded-id"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

const recipientsPageSize = 20

// ListRecipientsRequest parameters for listing recipient accounts.
type ListRecipientsRequest struct {
	ProfileID ProfileID
	Currency  Currency // Optional filter.
}

// ListRecipients returns the recipient accounts for a profile, paginating
// through the complete result set in recipientsPageSize chunks.
func (c *Client) ListRecipients(
	ctx context.Context,
	req ListRecipientsRequest,
) ([]Recipient, error) {
	if req.ProfileID.Get() == 0 {
		return nil, errorfamily.NewRejection(
			"wise.recipient.invalid_request",
			"profileID is required",
		)
	}

	var all []Recipient

	for offset := 0; ; offset += recipientsPageSize {
		page, err := c.listRecipientsPage(ctx, req, offset)
		if err != nil {
			return nil, err
		}

		all = append(all, page...)

		if len(page) < recipientsPageSize {
			return all, nil
		}
	}
}

func (c *Client) listRecipientsPage(
	ctx context.Context,
	req ListRecipientsRequest,
	offset int,
) ([]Recipient, error) {
	query := func() string {
		v := url.Values{}
		v.Set("profile", strconv.FormatInt(req.ProfileID.Get(), 10))
		v.Set("size", strconv.Itoa(recipientsPageSize))
		v.Set("offset", strconv.Itoa(offset))

		if req.Currency != "" {
			v.Set(wireKeyCurrency, string(req.Currency))
		}

		return v.Encode()
	}

	var recipients []raw.Recipient

	err := c.getWithQuery(ctx, "/v2/accounts", query, &recipients)
	if err != nil {
		return nil, fmt.Errorf("list recipients for profile %d: %w", req.ProfileID.Get(), err)
	}

	result := make([]Recipient, 0, len(recipients))
	for _, r := range recipients {
		mapped, mapErr := mapRecipient(r)
		if mapErr != nil {
			return nil, errorfamily.WrapCorruption(mapErr, "wise.recipient.map",
				fmt.Sprintf("map recipient %d", r.ID))
		}

		result = append(result, mapped)
	}

	return result, nil
}

// GetRecipient returns a single recipient account by ID.
func (c *Client) GetRecipient(ctx context.Context, recipientID RecipientID) (*Recipient, error) {
	path := fmt.Sprintf("/v1/accounts/%d", recipientID.Get())

	var recipient raw.Recipient

	if err := fetchByID(ctx, c, recipientID.Get(), "recipient", path, &recipient); err != nil {
		return nil, err
	}

	result, mapErr := mapRecipient(recipient)
	if mapErr != nil {
		return nil, fmt.Errorf("map recipient %d: %w", recipient.ID, mapErr)
	}

	return &result, nil
}

// CreateRecipient creates a new recipient account for a profile.
func (c *Client) CreateRecipient(
	ctx context.Context,
	req CreateRecipientRequest,
) (*Recipient, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	body := req.toWire()

	var recipient raw.Recipient

	err := c.post(ctx, "/v1/accounts", body, &recipient)
	if err != nil {
		return nil, fmt.Errorf("create recipient for profile %d: %w", req.ProfileID.Get(), err)
	}

	result, mapErr := mapRecipient(recipient)
	if mapErr != nil {
		return nil, fmt.Errorf("map created recipient %d: %w", recipient.ID, mapErr)
	}

	return &result, nil
}

func (r CreateRecipientRequest) validate() error {
	if r.ProfileID.Get() == 0 {
		return errorfamily.NewRejection(
			"wise.recipient.invalid_request",
			"profileID is required",
		)
	}

	if r.Currency == "" {
		return errorfamily.NewRejection(
			"wise.recipient.invalid_request",
			"currency is required",
		)
	}

	if r.AccountHolderName == "" {
		return errorfamily.NewRejection(
			"wise.recipient.invalid_request",
			"accountHolderName is required",
		)
	}

	if r.Type == "" {
		return errorfamily.NewRejection(
			"wise.recipient.invalid_request",
			"type is required",
		)
	}

	if len(r.Details) == 0 {
		return errorfamily.NewRejection(
			"wise.recipient.invalid_request",
			"details are required",
		)
	}

	return nil
}

func (r CreateRecipientRequest) toWire() map[string]any {
	body := map[string]any{
		"profile":           r.ProfileID.Get(),
		wireKeyCurrency:     string(r.Currency),
		wireKeyType:         r.Type,
		"accountHolderName": r.AccountHolderName,
		wireKeyDetails:      r.Details,
	}

	if r.OwnedByCustomer {
		body["ownedByCustomer"] = true
	}

	return body
}

func mapRecipient(r raw.Recipient) (Recipient, error) {
	currency, err := NewCurrency(r.Currency)
	if err != nil {
		return Recipient{}, errorfamily.WrapCorruption(
			err,
			"wise.recipient.parse_currency",
			fmt.Sprintf("currency %q", r.Currency),
		)
	}

	details := make(map[string]string, len(r.Details))
	for k, v := range r.Details {
		details[k] = fmt.Sprintf("%v", v)
	}

	return Recipient{
		ID:                id.NewID[RecipientBrand](r.ID),
		AccountHolderName: r.AccountHolderName,
		Currency:          currency,
		Country:           r.Country,
		Type:              r.Type,
		Details:           details,
		Active:            r.Active,
	}, nil
}
