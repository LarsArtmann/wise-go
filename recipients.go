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

// ListRecipients returns every recipient account of a profile, following the
// page envelope's seekPositionForNext cursor until the last page.
func (c *Client) ListRecipients(
	ctx context.Context,
	req ListRecipientsRequest,
) ([]Recipient, error) {
	if err := requireID(req.ProfileID, errorCodeRecipientInvalidRequest, "profileID"); err != nil {
		return nil, err
	}

	all := []Recipient{}

	seek := ""

	for {
		page, next, err := c.listRecipientsPage(ctx, req, seek)
		if err != nil {
			return nil, err
		}

		all = append(all, page...)

		if next == nil || len(page) == 0 {
			return all, nil
		}

		seek = strconv.FormatInt(*next, 10)
	}
}

// listRecipientsPage fetches one page for the seek position ("" = first page)
// and returns its mapped recipients plus the next seek position, nil on the
// last page.
func (c *Client) listRecipientsPage(
	ctx context.Context,
	req ListRecipientsRequest,
	seek string,
) ([]Recipient, *int64, error) {
	query := func() string {
		v := url.Values{}
		v.Set("profileId", strconv.FormatInt(req.ProfileID.Get(), 10))
		v.Set("size", strconv.Itoa(recipientsPageSize))

		if seek != "" {
			v.Set("seekPosition", seek)
		}

		if req.Currency != "" {
			v.Set(wireKeyCurrency, string(req.Currency))
		}

		return v.Encode()
	}

	var page raw.RecipientPage

	err := c.getWithQuery(ctx, "/v2/accounts", query, &page)
	if err != nil {
		return nil, nil, fmt.Errorf("list recipients for profile %d: %w", req.ProfileID.Get(), err)
	}

	result := make([]Recipient, 0, len(page.Content))
	for _, r := range page.Content {
		mapped, mapErr := mapRecipient(r)
		if mapErr != nil {
			return nil, nil, errorfamily.WrapCorruption(mapErr, errorCodeRecipientMap,
				fmt.Sprintf("map recipient %d", r.ID))
		}

		result = append(result, mapped)
	}

	return result, page.SeekPositionForNext, nil
}

// GetRecipient returns a single recipient account by ID.
func (c *Client) GetRecipient(ctx context.Context, recipientID RecipientID) (*Recipient, error) {
	path := fmt.Sprintf("/v1/accounts/%d", recipientID.Get())

	var recipient raw.Recipient

	if err := fetchByID(
		ctx, c, recipientID, "recipient", errorCodeRecipientInvalidRequest, path, &recipient,
	); err != nil {
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
	if err := requireID(r.ProfileID, errorCodeRecipientInvalidRequest, "profileID"); err != nil {
		return err
	}

	if r.Currency == "" {
		return errorfamily.NewRejection(
			errorCodeRecipientInvalidRequest,
			"currency is required",
		)
	}

	if r.AccountHolderName == "" {
		return errorfamily.NewRejection(
			errorCodeRecipientInvalidRequest,
			"accountHolderName is required",
		)
	}

	if r.Type == "" {
		return errorfamily.NewRejection(
			errorCodeRecipientInvalidRequest,
			"type is required",
		)
	}

	if len(r.Details) == 0 {
		return errorfamily.NewRejection(
			errorCodeRecipientInvalidRequest,
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

	// art-dupl:accept ownedByCustomer is a two-line optional wire field per request type; each body builder owns its own field set
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
			errorCodeRecipientParseCurrency,
			fmt.Sprintf("currency %q", r.Currency),
		)
	}

	details := make(map[string]string, len(r.Details))
	for k, v := range r.Details {
		details[k] = fmt.Sprintf("%v", v)
	}

	holder := r.AccountHolderName
	if holder == "" {
		holder = r.Name.FullName
	}

	return Recipient{
		ID:                id.NewID[RecipientBrand](r.ID),
		AccountHolderName: holder,
		Currency:          currency,
		Country:           r.Country,
		Type:              r.Type,
		Details:           details,
		Active:            r.Active,
	}, nil
}
