// unlike the camelCase core API
//
//nolint:tagliatelle // Wise's webhook wire surface (this entire file) uses snake_case,
package raw

import "encoding/json/jsontext"

// Subscription is the wire representation of a webhook subscription
// (/2026Q3/profiles/{profileId}/subscriptions and the application-level
// equivalents). Field set mirrors the documented Subscription response;
// unknown fields are ignored by the JSON decoder so Wise can add fields
// without breaking the SDK.
type Subscription struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	TriggerOn string               `json:"trigger_on"`
	Delivery  SubscriptionDelivery `json:"delivery"`
	CreatedAt string               `json:"created_at"`
	CreatedBy SubscriptionCreator  `json:"created_by"`
	Scope     SubscriptionScope    `json:"scope"`
}

// SubscriptionDelivery is the delivery block of a Subscription: the HTTPS
// endpoint Wise POSTs event envelopes to, and the event-schema semantic
// version the payloads use (e.g. "4.0.0").
type SubscriptionDelivery struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

// SubscriptionCreator identifies what created a subscription: an API
// application (client key) or a user.
type SubscriptionCreator struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// SubscriptionScope names the owning surface of a subscription: the client
// key (application scope) or the profile ID (profile scope).
type SubscriptionScope struct {
	Domain string `json:"domain"`
	ID     string `json:"id"`
}

// WebhookEventEnvelope is the wire envelope Wise POSTs to a subscription's
// delivery URL. Data stays raw (jsontext.Value): per-event payload decoding
// is the public layer's job, so unknown event types never break parsing.
type WebhookEventEnvelope struct {
	Data           jsontext.Value `json:"data"`
	SubscriptionID string         `json:"subscription_id"`
	EventType      string         `json:"event_type"`
	SchemaVersion  string         `json:"schema_version"`
	SentAt         string         `json:"sent_at"`
}

// WebhookEventResource is the resource block carried by event payloads
// (transfers#state-change, balances#credit). AccountID is only present for
// transfer events.
type WebhookEventResource struct {
	Type      string `json:"type"`
	ID        int64  `json:"id"`
	ProfileID int64  `json:"profile_id"`
	AccountID *int64 `json:"account_id"`
}

// TransferStateChangeData is the wire payload of transfers#state-change
// events.
type TransferStateChangeData struct {
	Resource      WebhookEventResource `json:"resource"`
	CurrentState  string               `json:"current_state"`
	PreviousState string               `json:"previous_state"`
	OccurredAt    string               `json:"occurred_at"`
}

// TransferPayoutFailureData is the wire payload of transfers#payout-failure
// events (schema version 5.0.0). Wise recommends processing it alongside
// transfers#state-change: a payout can fail without the transfer state
// changing.
type TransferPayoutFailureData struct {
	TransferID         int64  `json:"transfer_id"`
	ProfileID          int64  `json:"profile_id"`
	FailureReasonCode  string `json:"failure_reason_code"`
	FailureDescription string `json:"failure_description"`
	OccurredAt         string `json:"occurred_at"`
}

// BalanceCreditData is the wire payload of balances#credit events. The
// transaction type is always "credit" (debits are their own event type), so
// the discriminator is not carried.
type BalanceCreditData struct {
	Resource                     WebhookEventResource `json:"resource"`
	Amount                       float64              `json:"amount"`
	Currency                     string               `json:"currency"`
	PostTransactionBalanceAmount float64              `json:"post_transaction_balance_amount"`
	OccurredAt                   string               `json:"occurred_at"`
}
