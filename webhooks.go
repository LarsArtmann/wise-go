package wise

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// HeaderWebhookSignature is the header Wise signs every webhook delivery
// with: a Base64-encoded RSA-SHA256 signature of the raw request body.
const HeaderWebhookSignature = "X-Signature-SHA256"

// HeaderDeliveryID is the header Wise stamps every webhook delivery with:
// an identifier that is unique per delivery attempt. Use it as the
// deduplication key when processing idempotently — Wise redelivers until
// acknowledged, and a verified signature does not make a delivery new.
const HeaderDeliveryID = "X-Delivery-Id"

// ParseWebhookPublicKey parses the PEM-encoded RSA public key Wise shows per
// webhook subscription. Parse it once at startup so a misconfigured key
// fails loudly there, not silently per delivery.
func ParseWebhookPublicKey(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		//nolint:err113 // config-validation error with fixed message; no sentinel to wrap
		return nil, errors.New("no PEM block found in webhook public key")
	}

	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			//nolint:err113 // dynamic type name in message; no sentinel to wrap
			return nil, fmt.Errorf("webhook public key is %T, want RSA", key)
		}

		return rsaKey, nil
	}

	rsaKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse webhook public key: %w", err)
	}

	return rsaKey, nil
}

// VerifyWebhookSignature reports whether the X-Signature-SHA256 header value
// of a webhook delivery is an authentic RSA-SHA256 signature of the raw
// request body. Verify the raw bytes exactly as received — read the body
// before any re-marshalling, which would change the signed input.
//
// Reject the delivery (HTTP 401/403) when this returns false.
func VerifyWebhookSignature(payload []byte, signatureB64 string, key *rsa.PublicKey) bool {
	if key == nil || signatureB64 == "" {
		return false
	}

	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return false
	}

	digest := sha256.Sum256(payload)

	// VerifyPKCS1v15 reports invalid signatures as an error only; a nil
	// error is a match, any error means reject.
	return rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature) == nil
}

// webhookAPIVersion is the quarterly versioned API surface that hosts the
// webhook subscription endpoints (/2026Q3/profiles/{profileId}/subscriptions
// and the application-level equivalents). The subscription endpoints exist
// only under the versioned surface — unlike the legacy /v1../v4 paths the
// rest of the SDK uses — per the OpenAPI spec's server URL and the live API
// reference.
const webhookAPIVersion = "2026Q3"

// CreateProfileWebhookSubscription registers a webhook subscription on a
// profile (POST /2026Q3/profiles/{profileId}/subscriptions): Wise will POST
// an event envelope to req.Delivery.URL whenever the req.TriggerOn event
// occurs for anything the profile owns.
//
// Name, TriggerOn, and Delivery (version + URL) are validated client-side;
// the URL must be HTTPS. The profile's user token owns the subscription —
// application-scoped subscriptions need a client-credentials token and are
// not covered by this SDK yet.
func (c *Client) CreateProfileWebhookSubscription(
	ctx context.Context,
	profileID ProfileID,
	req CreateWebhookSubscriptionRequest,
) (*WebhookSubscription, error) {
	if err := requireID(profileID, "wise.webhook.invalid_request", "profileID"); err != nil {
		return nil, err
	}

	if err := req.validate(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/%s/profiles/%d/subscriptions", webhookAPIVersion, profileID.Get())

	var subscription raw.Subscription

	if err := c.post(ctx, path, req.toWire(), &subscription); err != nil {
		return nil, fmt.Errorf("create webhook subscription for profile %d: %w", profileID.Get(), err)
	}

	return toWebhookSubscription("map created webhook subscription", subscription)
}

// ListProfileWebhookSubscriptions returns every webhook subscription
// registered on a profile (GET /2026Q3/profiles/{profileId}/subscriptions).
// Wise returns the complete set in a single response (no pagination).
func (c *Client) ListProfileWebhookSubscriptions(
	ctx context.Context,
	profileID ProfileID,
) ([]WebhookSubscription, error) {
	if err := requireID(profileID, "wise.webhook.invalid_request", "profileID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/%s/profiles/%d/subscriptions", webhookAPIVersion, profileID.Get())

	var subscriptions []raw.Subscription

	if err := c.get(ctx, path, &subscriptions); err != nil {
		return nil, fmt.Errorf("list webhook subscriptions for profile %d: %w", profileID.Get(), err)
	}

	results := make([]WebhookSubscription, 0, len(subscriptions))
	for i, subscription := range subscriptions {
		mapped, err := toWebhookSubscription(fmt.Sprintf("map webhook subscription %d", i), subscription)
		if err != nil {
			return nil, err
		}

		results = append(results, *mapped)
	}

	return results, nil
}

// GetProfileWebhookSubscription returns a single webhook subscription by ID
// (GET /2026Q3/profiles/{profileId}/subscriptions/{subscriptionId}). An
// unknown subscription ID is a 404, classified as *NotFoundError.
func (c *Client) GetProfileWebhookSubscription(
	ctx context.Context,
	profileID ProfileID,
	subscriptionID WebhookSubscriptionID,
) (*WebhookSubscription, error) {
	if err := requireID(profileID, "wise.webhook.invalid_request", "profileID"); err != nil {
		return nil, err
	}

	if err := requireID(subscriptionID, "wise.webhook.invalid_request", "subscriptionID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/%s/profiles/%d/subscriptions/%s", webhookAPIVersion, profileID.Get(), subscriptionID.Get())

	var subscription raw.Subscription

	if err := c.get(ctx, path, &subscription); err != nil {
		return nil, fmt.Errorf("get webhook subscription %s for profile %d: %w", subscriptionID.Get(), profileID.Get(), err)
	}

	return toWebhookSubscription("map webhook subscription", subscription)
}

// DeleteProfileWebhookSubscription removes a webhook subscription (DELETE
// /2026Q3/profiles/{profileId}/subscriptions/{subscriptionId}). Wise
// acknowledges with 204 No Content; deleting an unknown subscription ID is a
// 404, classified as *NotFoundError. Deleting stops deliveries immediately.
func (c *Client) DeleteProfileWebhookSubscription(
	ctx context.Context,
	profileID ProfileID,
	subscriptionID WebhookSubscriptionID,
) error {
	if err := requireID(profileID, "wise.webhook.invalid_request", "profileID"); err != nil {
		return err
	}

	if err := requireID(subscriptionID, "wise.webhook.invalid_request", "subscriptionID"); err != nil {
		return err
	}

	path := fmt.Sprintf("/%s/profiles/%d/subscriptions/%s", webhookAPIVersion, profileID.Get(), subscriptionID.Get())

	if err := c.delete(ctx, path); err != nil {
		return fmt.Errorf("delete webhook subscription %s for profile %d: %w", subscriptionID.Get(), profileID.Get(), err)
	}

	return nil
}

// validate mirrors the subscription-request schema's required fields so a
// malformed create never reaches the API.
func (r CreateWebhookSubscriptionRequest) validate() error {
	if r.Name == "" {
		return errorfamily.NewRejection("wise.webhook.invalid_request", "name is required")
	}

	if r.TriggerOn == "" {
		return errorfamily.NewRejection("wise.webhook.invalid_request", "triggerOn is required")
	}

	if r.Delivery.Version == "" {
		return errorfamily.NewRejection("wise.webhook.invalid_request", "delivery.version is required")
	}

	if !strings.HasPrefix(r.Delivery.URL, "https://") {
		return errorfamily.NewRejection(
			"wise.webhook.invalid_request",
			"delivery.url must be an HTTPS URL (Wise only delivers to HTTPS endpoints)",
		)
	}

	return nil
}

// toWire renders the subscription create body. Wise's webhook wire uses
// snake_case keys, unlike the camelCase core API.
func (r CreateWebhookSubscriptionRequest) toWire() map[string]any {
	return map[string]any{
		"name":       r.Name,
		"trigger_on": string(r.TriggerOn),
		"delivery": map[string]any{
			"version": r.Delivery.Version,
			"url":     r.Delivery.URL,
		},
	}
}

// toWebhookSubscription maps the wire subscription to its public type. The
// label is the full error-message prefix for mapper failures.
func toWebhookSubscription(label string, subscription raw.Subscription) (*WebhookSubscription, error) {
	createdAt, err := parseWiseTimestamp(subscription.CreatedAt)
	if err != nil {
		return nil, errorfamily.WrapCorruption(
			err,
			"wise.response.decode",
			label+" "+subscription.ID+": parse created_at",
		)
	}

	return &WebhookSubscription{
		ID:        NewWebhookSubscriptionID(subscription.ID),
		Name:      subscription.Name,
		TriggerOn: WebhookEventType(subscription.TriggerOn),
		Delivery: WebhookDelivery{
			Version: subscription.Delivery.Version,
			URL:     subscription.Delivery.URL,
		},
		CreatedAt: createdAt,
		CreatedBy: WebhookCreator{
			ID:   subscription.CreatedBy.ID,
			Type: WebhookCreatorType(subscription.CreatedBy.Type),
		},
		Scope: WebhookScope{
			Domain: WebhookScopeDomain(subscription.Scope.Domain),
			ID:     subscription.Scope.ID,
		},
	}, nil
}
