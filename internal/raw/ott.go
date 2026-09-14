package raw

import "encoding/json/jsontext"

// OTTResponse is the wire representation of Wise's one-time-token (OTT)
// status, served by GET /2026Q3/one-time-token/status and as the 200 body of
// the per-channel verify endpoints. Unknown fields are ignored by the JSON
// decoder so Wise can add fields without breaking the SDK.
type OTTResponse struct {
	OneTimeTokenProperties OTTProperties `json:"oneTimeTokenProperties"`
}

// OTTProperties is the status block of an OTT: the challenge list, remaining
// validity, the action the token authorizes, and the creating user.
type OTTProperties struct {
	OneTimeToken string            `json:"oneTimeToken"`
	Challenges   []OTTWireChallenge `json:"challenges"`
	Validity     int64             `json:"validity"`
	ActionType   string            `json:"actionType"`
	UserID       int64             `json:"userId"`
}

// OTTWireChallenge is one challenge entry: the primary way to satisfy it,
// alternative ways, and whether it is required and already passed.
type OTTWireChallenge struct {
	PrimaryChallenge OTTWireChallengeView `json:"primaryChallenge"`
	// Alternatives stays raw (jsontext.Value): the OpenAPI spec types the
	// items only as "object" without properties, so the public layer decodes
	// leniently instead of the raw layer guessing one shape.
	Alternatives jsontext.Value `json:"alternatives"`
	Required     bool           `json:"required"`
	Passed       bool           `json:"passed"`
}

// OTTWireChallengeView is one concrete challenge (primary or alternative):
// its type and the render data Wise provides for it.
type OTTWireChallengeView struct {
	Type     string         `json:"type"`
	ViewData map[string]any `json:"viewData"`
}

// OTTTriggerResponse is the 200 body of the per-channel trigger endpoints.
type OTTTriggerResponse struct {
	ObfuscatedPhoneNo string `json:"obfuscatedPhoneNo"`
}

// OTTVerifyRequest is the request body of the per-channel verify endpoints.
type OTTVerifyRequest struct {
	OTPCode string `json:"otpCode"`
}
