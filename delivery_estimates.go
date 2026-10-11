package wise

import (
	"context"
	"fmt"
	"net/url"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// rawDeliveryEstimate avoids exposing internal/raw types in this file's
// signature while keeping the wire decoding inside the package.
type rawDeliveryEstimate = raw.DeliveryEstimate

// DeliveryEstimate is the parsed representation of the estimated delivery
// time for a transfer. EstimatedDeliveryDate is the raw timestamp, and
// FormattedEstimatedDeliveryDate is a customer-friendly string suitable for
// passing through to a front-end.
type DeliveryEstimate struct {
	EstimatedDeliveryDate          time.Time
	FormattedEstimatedDeliveryDate string
}

// GetDeliveryEstimate returns the live delivery estimate for a transfer —
// the time at which Wise currently expects the transfer to arrive in the
// beneficiary's bank account. The estimate is not a guarantee, but Wise
// keeps it as accurate as possible as the transfer progresses.
//
// timezone optionally selects the IANA timezone used for the formatted
// estimate text; it defaults to UTC when empty.
func (c *Client) GetDeliveryEstimate(
	ctx context.Context,
	transferID TransferID,
	timezone string,
) (*DeliveryEstimate, error) {
	if err := requireID(transferID, "wise.transfer.invalid_request", "transferID"); err != nil {
		return nil, err
	}

	query := func() string {
		v := url.Values{}
		if timezone != "" {
			v.Set("timezone", timezone)
		}

		return v.Encode()
	}

	path := fmt.Sprintf("/v1/delivery-estimates/%d", transferID.Get())

	var estimate rawDeliveryEstimate

	if err := c.getWithQuery(ctx, path, query, &estimate); err != nil {
		return nil, fmt.Errorf("get delivery estimate for transfer %d: %w", transferID.Get(), err)
	}

	result, mapErr := mapDeliveryEstimate(estimate)
	if mapErr != nil {
		return nil, fmt.Errorf("map delivery estimate for transfer %d: %w", transferID.Get(), mapErr)
	}

	return result, nil
}

func mapDeliveryEstimate(estimate rawDeliveryEstimate) (*DeliveryEstimate, error) {
	deliveryTime, err := parseWiseTimestamp(estimate.EstimatedDeliveryDate)
	if err != nil {
		return nil, errorfamily.WrapCorruption(
			err,
			"wise.delivery_estimate.parse_date",
			fmt.Sprintf("parse estimatedDeliveryDate %q", estimate.EstimatedDeliveryDate),
		)
	}

	return &DeliveryEstimate{
		EstimatedDeliveryDate:          deliveryTime,
		FormattedEstimatedDeliveryDate: estimate.FormattedEstimatedDeliveryDate,
	}, nil
}
