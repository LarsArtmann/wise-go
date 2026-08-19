package raw

// Transfer from Wise API (/v1/transfers, /v1/transfers/{id}).
//
// The list and get endpoints return the same object shape. Field set mirrors
// the documented Standard Transfer response; unknown fields are ignored by
// the JSON decoder so Wise can add fields without breaking the SDK.
type Transfer struct {
	ID                    int64           `json:"id"`
	User                  int64           `json:"user"`
	TargetAccount         int64           `json:"targetAccount"`
	SourceAccount         *int64          `json:"sourceAccount"`
	Quote                 *int64          `json:"quote"`
	QuoteUUID             string          `json:"quoteUuid"`
	Status                string          `json:"status"`
	Rate                  float64         `json:"rate"`
	Created               string          `json:"created"`
	Business              *int64          `json:"business"`
	Details               TransferDetails `json:"details"`
	HasActiveIssues       bool            `json:"hasActiveIssues"`
	SourceCurrency        string          `json:"sourceCurrency"`
	SourceValue           float64         `json:"sourceValue"`
	TargetCurrency        string          `json:"targetCurrency"`
	TargetValue           float64         `json:"targetValue"`
	CustomerTransactionID string          `json:"customerTransactionId"`
	Reference             string          `json:"reference"` // Deprecated by Wise: prefer Details.Reference.
}

// TransferDetails is the details block of a Transfer.
type TransferDetails struct {
	Reference string `json:"reference"`
}
