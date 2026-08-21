package wise

import (
	"context"
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// ValidateTransferRequirements discovers the transfer-specific fields
// required for a quote and recipient combination. Requirements vary by
// currency route, transfer amount, and regulatory region; calling this
// before CreateTransfer avoids delays caused by missing details.
//
// The response is an array of dynamic forms. Fields flagged
// RefreshRequirementsOnChange mean the validation must be repeated once the
// field is populated to discover lower-level required fields. Cross-reference
// the response against a prepared transfer with MissingTransferDetails to
// learn which required fields are still unsatisfied.
func (c *Client) ValidateTransferRequirements(
	ctx context.Context,
	req ValidateTransferRequirementsRequest,
) ([]TransferRequirement, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	var requirements []raw.TransferRequirement

	err := c.post(ctx, "/v1/transfer-requirements", req.toWire(), &requirements)
	if err != nil {
		return nil, fmt.Errorf("validate transfer requirements for quote %s: %w", req.QuoteID.Get(), err)
	}

	result := make([]TransferRequirement, 0, len(requirements))
	for _, requirement := range requirements {
		result = append(result, mapTransferRequirement(requirement))
	}

	return result, nil
}

func (r ValidateTransferRequirementsRequest) validate() error {
	if r.TargetAccount.Get() == 0 {
		return errorfamily.NewRejection(
			"wise.transfer.invalid_request",
			"targetAccount is required",
		)
	}

	if r.QuoteID.Get() == "" {
		return errorfamily.NewRejection(
			"wise.transfer.invalid_request",
			"quoteUuid is required",
		)
	}

	return nil
}

// MissingTransferDetails cross-references a transfer-requirements response
// against a prepared CreateTransferRequest and returns the human-readable
// names of fields Wise marked required that the request does not satisfy.
// An empty result means the request carries every detail required for the
// corridor; call it after ValidateTransferRequirements and before
// CreateTransfer to fail before spending a customerTransactionId.
//
// Select fields (ValuesAllowed non-empty) are only satisfied by a value from
// their allowed list. Keys the SDK does not model as typed fields (rare,
// corridor-specific) are always reported — the corridor then needs a field
// beyond CreateTransferRequest's surface, and the response's dynamic form
// describes what Wise expects.
//
// Fields flagged RefreshRequirementsOnChange reveal lower-level requirements
// only after they are populated: re-run ValidateTransferRequirements with the
// populated value and call MissingTransferDetails on the fresh response.
func MissingTransferDetails(
	requirements []TransferRequirement,
	req CreateTransferRequest,
) []string {
	missing := make([]string, 0)

	for _, requirement := range requirements {
		for _, form := range requirement.Fields {
			for _, field := range form.Group {
				if !field.Required {
					continue
				}

				value, modeled := transferRequestDetailValue(req, field.Key)
				if !modeled {
					missing = append(missing, field.Name)

					continue
				}

				if value == "" || !selectValueAllowed(field, value) {
					missing = append(missing, field.Name)
				}
			}
		}
	}

	return missing
}

// selectValueAllowed reports whether value is an allowed choice for a select
// field. Fields without ValuesAllowed accept any non-empty value.
func selectValueAllowed(field TransferRequirementField, value string) bool {
	if len(field.ValuesAllowed) == 0 {
		return true
	}

	for _, allowed := range field.ValuesAllowed {
		if allowed.Key == value {
			return true
		}
	}

	return false
}

// transferRequestDetailValue maps a wire detail key onto the corresponding
// typed CreateTransferRequest field. The second result reports whether the
// key is modeled at all.
func transferRequestDetailValue(req CreateTransferRequest, key string) (string, bool) {
	switch key {
	case "reference":
		return req.Reference, true
	case "sourceOfFunds":
		return req.SourceOfFunds, true
	case "transferPurpose":
		return req.TransferPurpose, true
	case "transferPurposeInvoiceNumber":
		return req.TransferPurposeInvoiceNumber, true
	case "transferPurposeSubTransferPurpose":
		return req.TransferPurposeSubTransferPurpose, true
	default:
		return "", false
	}
}

func (r ValidateTransferRequirementsRequest) toWire() map[string]any {
	body := map[string]any{
		"targetAccount": r.TargetAccount.Get(),
		"quoteUuid":     r.QuoteID.Get(),
	}

	if r.CustomerTransactionID != "" {
		body["customerTransactionId"] = r.CustomerTransactionID
	}

	if r.OriginatorLegalEntityType != "" {
		body["originatorLegalEntityType"] = r.OriginatorLegalEntityType
	}

	if details := r.Details.toWire(); len(details) > 0 {
		body["details"] = details
	}

	return body
}

// toWire renders the optional details block, omitting empty fields.
func (d TransferRequirementsDetails) toWire() map[string]string {
	details := make(map[string]string)

	if d.Reference != "" {
		details["reference"] = d.Reference
	}

	if d.SourceOfFunds != "" {
		details["sourceOfFunds"] = d.SourceOfFunds
	}

	if d.SourceOfFundsOther != "" {
		details["sourceOfFundsOther"] = d.SourceOfFundsOther
	}

	if d.TransferPurpose != "" {
		details["transferPurpose"] = d.TransferPurpose
	}

	if d.TransferPurposeSubTransferPurpose != "" {
		details["transferPurposeSubTransferPurpose"] = d.TransferPurposeSubTransferPurpose
	}

	if d.TransferPurposeInvoiceNumber != "" {
		details["transferPurposeInvoiceNumber"] = d.TransferPurposeInvoiceNumber
	}

	if d.TransferNature != "" {
		details["transferNature"] = d.TransferNature
	}

	return details
}

func mapTransferRequirement(r raw.TransferRequirement) TransferRequirement {
	result := TransferRequirement{
		Type:   r.Type,
		Fields: make([]TransferRequirementForm, 0, len(r.Fields)),
	}

	for _, form := range r.Fields {
		result.Fields = append(result.Fields, mapTransferRequirementForm(form))
	}

	return result
}

func mapTransferRequirementForm(form raw.TransferRequirementForm) TransferRequirementForm {
	result := TransferRequirementForm{
		Name:  form.Name,
		Group: make([]TransferRequirementField, 0, len(form.Group)),
	}

	for _, field := range form.Group {
		result.Group = append(result.Group, mapTransferRequirementField(field))
	}

	return result
}

func mapTransferRequirementField(field raw.TransferRequirementField) TransferRequirementField {
	return TransferRequirementField{
		Key:                         field.Key,
		Name:                        field.Name,
		Type:                        field.Type,
		RefreshRequirementsOnChange: field.RefreshRequirementsOnChange,
		Required:                    field.Required,
		DisplayFormat:               field.DisplayFormat,
		Example:                     field.Example,
		MinLength:                   field.MinLength,
		MaxLength:                   field.MaxLength,
		ValidationRegexp:            field.ValidationRegexp,
		ValuesAllowed:               mapTransferRequirementValues(field.ValuesAllowed),
	}
}

func mapTransferRequirementValues(values []raw.TransferRequirementValue) []TransferRequirementValue {
	if values == nil {
		return nil
	}

	result := make([]TransferRequirementValue, 0, len(values))
	for _, value := range values {
		result = append(result, TransferRequirementValue{Key: value.Key, Name: value.Name})
	}

	return result
}
