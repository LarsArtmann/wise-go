package raw

// Subscription is the wire representation of a webhook subscription
// (/2026Q3/profiles/{profileId}/subscriptions and the application-level
// equivalents). Field set mirrors the documented Subscription response;
// unknown fields are ignored by the JSON decoder so Wise can add fields
// without breaking the SDK.
type Subscription struct {
	ID string `json:"id"`
	//nolint:tagliatelle // Wise's webhook wire uses snake_case, unlike the camelCase core API
	Name string `json:"name"`
	//nolint:tagliatelle // Wise's webhook wire uses snake_case, unlike the camelCase core API
	TriggerOn string               `json:"trigger_on"`
	Delivery  SubscriptionDelivery `json:"delivery"`
	//nolint:tagliatelle // Wise's webhook wire uses snake_case, unlike the camelCase core API
	CreatedAt string              `json:"created_at"`
	CreatedBy SubscriptionCreator `json:"created_by"`
	Scope     SubscriptionScope   `json:"scope"`
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
