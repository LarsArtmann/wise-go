package raw

import (
	"encoding/json/v2"
	"testing"
)

// The wire-struct tests pin the raw types' JSON contract: Wise's wire keys
// decode into the tagged shapes (snake_case where Wise uses it, camelCase in
// the core API) and unknown fields are ignored, so Wise can add fields
// without breaking the SDK.

func TestProfileWireJSON(t *testing.T) {
	t.Parallel()

	var profile Profile

	body := []byte(`{
		"id": 1,
		"publicId": "pub-1",
		"userId": 77,
		"type": "PERSONAL",
		"firstName": "Jane",
		"lastName": "Doe",
		"email": "jane@example.com",
		"createdAt": "2023-01-15T10:30:00Z",
		"someFutureField": {"nested": true}
	}`)

	if err := json.Unmarshal(body, &profile); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if profile.ID != 1 || profile.PublicID != "pub-1" || profile.UserID != 77 {
		t.Fatalf("identity fields wrong: %+v", profile)
	}

	if profile.Type != "PERSONAL" || profile.Email != "jane@example.com" {
		t.Fatalf("fields wrong: %+v", profile)
	}
}

func TestSubscriptionWireJSON(t *testing.T) {
	t.Parallel()

	var subscription Subscription

	body := []byte(`{
		"id": "72195556-e5cb-495e-a010-b37a4f2a3043",
		"name": "Payout watcher",
		"trigger_on": "transfers#state-change",
		"delivery": {"version": "4.0.0", "url": "https://example.com/h"},
		"created_at": "2026-09-13T10:00:00Z",
		"created_by": {"id": "key-1", "type": "user"},
		"scope": {"domain": "profile", "id": "12345"}
	}`)

	if err := json.Unmarshal(body, &subscription); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if subscription.TriggerOn != "transfers#state-change" {
		t.Fatalf("trigger_on = %q", subscription.TriggerOn)
	}

	if subscription.CreatedBy.Type != "user" || subscription.Scope.Domain != "profile" {
		t.Fatalf("nested fields wrong: %+v", subscription)
	}
}

func TestWebhookEventEnvelopeWireJSON(t *testing.T) {
	t.Parallel()

	var envelope WebhookEventEnvelope

	body := []byte(`{
		"data": {"any": "shape"},
		"subscription_id": "72195556-e5cb-495e-a010-b37a4f2a3043",
		"event_type": "balances#credit",
		"schema_version": "4.0.0",
		"sent_at": "2026-09-13T10:00:00Z"
	}`)

	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if string(envelope.Data) != `{"any": "shape"}` {
		t.Fatalf("data = %s, want verbatim raw JSON", envelope.Data)
	}

	if envelope.EventType != "balances#credit" || envelope.SentAt != "2026-09-13T10:00:00Z" {
		t.Fatalf("envelope fields wrong: %+v", envelope)
	}
}

func TestTransferPayoutFailureDataWireJSON(t *testing.T) {
	t.Parallel()

	var payload TransferPayoutFailureData

	body := []byte(`{
		"transfer_id": 111,
		"profile_id": 222,
		"failure_reason_code": "TECHNICAL_ISSUE_RETRYABLE",
		"failure_description": "Wise will retry",
		"occurred_at": "2023-08-10T10:17:23.123Z"
	}`)

	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if payload.TransferID != 111 || payload.FailureReasonCode != "TECHNICAL_ISSUE_RETRYABLE" {
		t.Fatalf("fields wrong: %+v", payload)
	}
}
