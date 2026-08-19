package wise

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// transfersPageSize is the per-request page size used when paginating the
// transfers list. Wise does not document a maximum; 100 is the largest value
// observed to be honoured.
const transfersPageSize = 100

// ListTransfersRequest parameters for listing transfers.
//
// ProfileID is optional: when zero, Wise defaults to the user's personal
// profile. From/To filter on transfer creation time (createdDateStart/End,
// both inclusive) and are omitted when zero.
type ListTransfersRequest struct {
	ProfileID ProfileID
	From      time.Time
	To        time.Time
	Status    []TransferStatus // Optional filter: transfers in any of these statuses.
}

// ListTransfers returns the transfers for a profile, automatically
// paginating through the complete result set in transferPageSize chunks.
// Results are ordered by creation date, newest first (Wise API order).
//
// Unlike balance statements, the transfers endpoint is not SCA-protected and
// is available to personal API tokens in all regions.
func (c *Client) ListTransfers(ctx context.Context, req ListTransfersRequest) ([]Transfer, error) {
	var all []Transfer

	for offset := 0; ; offset += transfersPageSize {
		page, err := c.listTransfersPage(ctx, req, offset)
		if err != nil {
			return nil, err
		}

		all = append(all, page...)

		if len(page) < transfersPageSize {
			return all, nil
		}
	}
}

func (c *Client) listTransfersPage(
	ctx context.Context,
	req ListTransfersRequest,
	offset int,
) ([]Transfer, error) {
	query := func() string {
		v := url.Values{}
		v.Set("limit", strconv.Itoa(transfersPageSize))
		v.Set("offset", strconv.Itoa(offset))

		if req.ProfileID.Get() != 0 {
			v.Set("profile", strconv.FormatInt(req.ProfileID.Get(), 10))
		}

		if !req.From.IsZero() {
			v.Set("createdDateStart", req.From.Format(time.RFC3339))
		}

		if !req.To.IsZero() {
			v.Set("createdDateEnd", req.To.Format(time.RFC3339))
		}

		if len(req.Status) > 0 {
			statuses := make([]string, len(req.Status))
			for i, s := range req.Status {
				statuses[i] = string(s)
			}

			v.Set("status", strings.Join(statuses, ","))
		}

		return v.Encode()
	}

	var transfers []raw.Transfer

	err := c.getWithQuery(ctx, "/v1/transfers", query, &transfers)
	if err != nil {
		return nil, fmt.Errorf("list transfers (offset=%d): %w", offset, err)
	}

	result := make([]Transfer, 0, len(transfers))
	for _, t := range transfers {
		mapped, mapErr := mapTransfer(t)
		if mapErr != nil {
			return nil, errorfamily.WrapCorruption(mapErr, "wise.transfer.map",
				fmt.Sprintf("map transfer %d", t.ID))
		}

		result = append(result, mapped)
	}

	return result, nil
}

// mapTransfer converts a raw wire transfer into the parsed Transfer type.
func mapTransfer(t raw.Transfer) (Transfer, error) {
	created, err := parseWiseTimestamp(t.Created)
	if err != nil {
		return Transfer{}, fmt.Errorf("parse created %q: %w", t.Created, err)
	}

	source, err := toMoney(raw.BalanceAmount{Value: t.SourceValue, Currency: t.SourceCurrency})
	if err != nil {
		return Transfer{}, fmt.Errorf("source amount: %w", err)
	}

	target, err := toMoney(raw.BalanceAmount{Value: t.TargetValue, Currency: t.TargetCurrency})
	if err != nil {
		return Transfer{}, fmt.Errorf("target amount: %w", err)
	}

	reference := t.Details.Reference
	if reference == "" {
		reference = t.Reference // Wise deprecated the top-level field; use as fallback.
	}

	return Transfer{
		ID:                    NewTransferID(t.ID),
		RecipientID:           NewRecipientID(t.TargetAccount),
		Status:                TransferStatus(t.Status),
		Rate:                  t.Rate,
		Source:                source,
		Target:                target,
		Created:               created,
		Reference:             reference,
		CustomerTransactionID: t.CustomerTransactionID,
		HasActiveIssues:       t.HasActiveIssues,
	}, nil
}
