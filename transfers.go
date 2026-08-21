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
// paginating through the complete result set in pages of 100 (Wise's
// per-request maximum). Results are ordered by creation date, newest first
// (Wise API order).
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
			v.Set("createdDateStart", formatWiseTimestamp(req.From))
		}

		if !req.To.IsZero() {
			v.Set("createdDateEnd", formatWiseTimestamp(req.To))
		}

		if len(req.Status) > 0 {
			v.Set("status", joinStatuses(req.Status))
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

// joinStatuses renders a status filter as Wise's comma-separated list.
func joinStatuses(statuses []TransferStatus) string {
	parts := make([]string, 0, len(statuses))
	for _, s := range statuses {
		parts = append(parts, string(s))
	}

	return strings.Join(parts, ",")
}

// GetTransfer returns a single transfer by ID.
func (c *Client) GetTransfer(ctx context.Context, transferID TransferID) (*Transfer, error) {
	path := fmt.Sprintf("/v1/transfers/%d", transferID.Get())

	var transfer raw.Transfer

	if err := fetchByID(ctx, c, transferID.Get(), "transfer", path, &transfer); err != nil {
		return nil, err
	}

	return toTransfer("map transfer", transfer)
}

func (r CreateTransferRequest) validate() error {
	if err := requireID(r.QuoteID, "wise.transfer.invalid_request", "quoteID"); err != nil {
		return err
	}

	if err := requireID(r.TargetAccount, "wise.transfer.invalid_request", "targetAccount"); err != nil {
		return err
	}

	if r.CustomerTransactionID == "" {
		return errorfamily.NewRejection(
			"wise.transfer.invalid_request",
			"customerTransactionId is required",
		)
	}

	return nil
}

// detailsWire renders the request's optional details block through
// TransferRequirementsDetails.toWire so the details wire keys have one
// spelling across every endpoint that accepts a details block.
func (r CreateTransferRequest) detailsWire() map[string]string {
	return TransferRequirementsDetails{ //nolint:exhaustruct // CreateTransferRequest carries neither sourceOfFundsOther nor transferNature
		Reference:                         r.Reference,
		SourceOfFunds:                     r.SourceOfFunds,
		TransferPurpose:                   r.TransferPurpose,
		TransferPurposeSubTransferPurpose: r.TransferPurposeSubTransferPurpose,
		TransferPurposeInvoiceNumber:      r.TransferPurposeInvoiceNumber,
	}.toWire()
}

// CreateTransfer creates a new transfer from a quote to a recipient account.
//
// customerTransactionId is required for idempotency and must be a UUID unique
// to this transfer attempt. Reusing the same value with the same targetAccount
// and quoteUuid will return the existing transfer instead of creating a
// duplicate.
func (c *Client) CreateTransfer(
	ctx context.Context,
	req CreateTransferRequest,
) (*Transfer, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	body := map[string]any{
		"targetAccount":         req.TargetAccount.Get(),
		"quoteUuid":             req.QuoteID.Get(),
		"customerTransactionId": req.CustomerTransactionID,
	}

	if req.SourceAccount.Get() != 0 {
		body["sourceAccount"] = req.SourceAccount.Get()
	}

	details := req.detailsWire()
	if len(details) > 0 {
		body[wireKeyDetails] = details
	}

	var transfer raw.Transfer

	err := c.post(ctx, "/v1/transfers", body, &transfer)
	if err != nil {
		return nil, fmt.Errorf("create transfer for quote %s: %w", req.QuoteID.Get(), err)
	}

	return toTransfer("map created transfer", transfer)
}

// CancelTransfer cancels a transfer by ID. A transfer can only be cancelled
// if it has not been processed (not in funds_converted or later state) and
// has no processing problems. Cancellation is final.
func (c *Client) CancelTransfer(ctx context.Context, transferID TransferID) (*Transfer, error) {
	if err := requireID(transferID, "wise.transfer.invalid_request", "transferID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/transfers/%d/cancel", transferID.Get())

	var transfer raw.Transfer

	if err := c.put(ctx, path, nil, &transfer); err != nil {
		return nil, fmt.Errorf("cancel transfer %d: %w", transferID.Get(), err)
	}

	return toTransfer("map cancelled transfer", transfer)
}

// FundTransfer funds a created transfer from the profile's Wise balance
// (POST /v1/profiles/{profileId}/transfers/{transferId}/payments), triggering
// processing of the payout. This is the final step of the transfer flow:
// quote, create recipient, create transfer, fund.
//
// Wise debits the transfer's source balance; the funding succeeds only if it
// holds the full amount. An underfunded balance is not an error return but a
// result with Status FundingStatusRejected and ErrorCode
// FundingErrorCodeBalanceInsufficientFunds — top up and call FundTransfer
// again.
//
// The endpoint is SCA-protected for profiles registered in the UK/EEA: without
// approval it fails with *SCAChallengeError (see WithSCAApprovalToken).
func (c *Client) FundTransfer(
	ctx context.Context,
	profileID ProfileID,
	transferID TransferID,
) (*FundTransferResult, error) {
	if err := requireID(profileID, "wise.transfer.invalid_request", "profileID"); err != nil {
		return nil, err
	}

	if err := requireID(transferID, "wise.transfer.invalid_request", "transferID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/profiles/%d/transfers/%d/payments", profileID.Get(), transferID.Get())

	var funding raw.FundingResponse

	// The request body is optional; an empty body selects default balance funding.
	if err := c.post(ctx, path, nil, &funding); err != nil {
		return nil, fmt.Errorf("fund transfer %d: %w", transferID.Get(), err)
	}

	result, mapErr := mapFundTransferResult(funding)
	if mapErr != nil {
		return nil, fmt.Errorf("map funding result for transfer %d: %w", transferID.Get(), mapErr)
	}

	return &result, nil
}

// mapFundTransferResult converts a raw funding response into the parsed
// FundTransferResult type.
func mapFundTransferResult(funding raw.FundingResponse) (FundTransferResult, error) {
	fundingType, err := parseEnum(map[string]FundingType{
		string(FundingTypeBalance):            FundingTypeBalance,
		string(FundingTypeTrustedPreFundBulk): FundingTypeTrustedPreFundBulk,
		string(FundingTypeTrustedPreFundTx):   FundingTypeTrustedPreFundTx,
	}, funding.Type, "funding type")
	if err != nil {
		return FundTransferResult{}, errorfamily.WrapCorruption(
			err,
			"wise.funding.parse_type",
			fmt.Sprintf("parse funding type %q", funding.Type),
		)
	}

	status, err := parseEnum(map[string]FundingStatus{
		string(FundingStatusCreated):   FundingStatusCreated,
		string(FundingStatusCompleted): FundingStatusCompleted,
		string(FundingStatusRejected):  FundingStatusRejected,
	}, funding.Status, "funding status")
	if err != nil {
		return FundTransferResult{}, errorfamily.WrapCorruption(
			err,
			"wise.funding.parse_status",
			fmt.Sprintf("parse funding status %q (type %q)", funding.Status, funding.Type),
		)
	}

	var balanceTransactionID *BalanceTransactionID

	if funding.BalanceTransactionID != nil {
		id := NewBalanceTransactionID(*funding.BalanceTransactionID)
		balanceTransactionID = &id
	}

	return FundTransferResult{
		Type:                 fundingType,
		Status:               status,
		ErrorCode:            FundingErrorCode(funding.ErrorCode),
		ErrorMessage:         funding.ErrorMessage,
		BalanceTransactionID: balanceTransactionID,
		PartnerReference:     funding.PartnerReference,
	}, nil
}

// toTransfer maps one raw transfer into the public type, wrapping mapping
// failures with the given error label and the transfer ID ("map cancelled
// transfer 42: ..."). Shared by the transfer endpoints so the map-and-wrap
// discipline has one spelling.
func toTransfer(label string, transfer raw.Transfer) (*Transfer, error) {
	result, err := mapTransfer(transfer)
	if err != nil {
		return nil, fmt.Errorf("%s %d: %w", label, transfer.ID, err)
	}

	return &result, nil
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

	var sourceAccount *BalanceID

	if t.SourceAccount != nil {
		account := NewBalanceID(*t.SourceAccount)
		sourceAccount = &account
	}

	return Transfer{
		ID:                    NewTransferID(t.ID),
		RecipientID:           NewRecipientID(t.TargetAccount),
		SourceAccount:         sourceAccount,
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
