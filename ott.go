package wise

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// HeaderOneTimeToken is the header Wise's one-time-token (OTT) endpoints
// require: the text value of the OTT whose challenges are being inspected,
// triggered, or verified. Values are in canonical MIME form (what net/http's
// Header.Get/Set use as map keys; on the wire HTTP header names are
// case-insensitive).
const HeaderOneTimeToken = "One-Time-Token"

// OTTChannel is the phone-based delivery channel of a one-time password for
// an OTT challenge. Wise serves a trigger and a verify endpoint per channel.
type OTTChannel string

// One-time-password delivery channels. These three channels share the same
// wire contract; PIN challenges use JOSE/JWE direct encryption and are not
// covered by this SDK.
const (
	OTTChannelSMS      OTTChannel = "sms"
	OTTChannelWhatsApp OTTChannel = "whatsapp"
	OTTChannelVoice    OTTChannel = "voice"
)

// isPhoneChannel reports whether channel is one of the OTP delivery channels
// the SDK can trigger and verify.
func isPhoneChannel(channel OTTChannel) bool {
	switch channel {
	case OTTChannelSMS, OTTChannelWhatsApp, OTTChannelVoice:
		return true
	default:
		return false
	}
}

// String returns the channel name.
func (c OTTChannel) String() string { return string(c) }

// challengeType maps a channel to the challenge type it satisfies: the wire
// channel paths are lowercase (sms/whatsapp/voice) while challenge types are
// uppercase (SMS/WHATSAPP/VOICE).
func (c OTTChannel) challengeType() OTTChallengeType {
	switch c {
	case OTTChannelSMS:
		return OTTChallengeSMS
	case OTTChannelWhatsApp:
		return OTTChallengeWhatsApp
	case OTTChannelVoice:
		return OTTChallengeVoice
	default:
		return OTTChallengeType(c)
	}
}

// OTTChallengeType is the type of a challenge Wise presents for an OTT
// (primaryChallenge.type). It is an open enum: Wise can add types, so unknown
// values pass through rather than failing to decode.
type OTTChallengeType string

// Challenge types documented for OTTs. Only the phone-based trio
// (SMS/WHATSAPP/VOICE) is clearable via the per-channel endpoints; PIN uses
// JOSE/JWE direct encryption and FACE_MAP/PARTNER_DEVICE_FINGERPRINT have no
// public API flow.
const (
	OTTChallengePIN                      OTTChallengeType = "PIN"
	OTTChallengeFaceMap                  OTTChallengeType = "FACE_MAP"
	OTTChallengeSMS                      OTTChallengeType = "SMS"
	OTTChallengeWhatsApp                 OTTChallengeType = "WHATSAPP"
	OTTChallengeVoice                    OTTChallengeType = "VOICE"
	OTTChallengePartnerDeviceFingerprint OTTChallengeType = "PARTNER_DEVICE_FINGERPRINT"
)

// OTTChallengeView is one concrete way to satisfy a challenge: its type plus
// the data Wise provides for rendering it (messages, IDs, or other
// attributes — e.g. an obfuscated phone number for SMS).
type OTTChallengeView struct {
	Type     OTTChallengeType
	ViewData map[string]any
}

// OTTChallenge is one entry of an OTT's challenge list: the primary way to
// clear it, alternative ways, and whether it is required to pass the OTT and
// already passed.
type OTTChallenge struct {
	Primary      OTTChallengeView
	Alternatives []OTTChallengeView
	Required     bool
	Passed       bool
}

// Supports reports whether the challenge can be cleared over the given
// phone channel: either the primary challenge is that channel's type, or one
// of the alternatives is.
func (c OTTChallenge) Supports(channel OTTChannel) bool {
	channelType := channel.challengeType()
	if c.Primary.Type == channelType {
		return true
	}

	for _, alt := range c.Alternatives {
		if alt.Type == channelType {
			return true
		}
	}

	return false
}

// OTTStatus is the parsed state of a one-time token: its challenges, the
// remaining validity, the action the token authorizes (e.g.
// BALANCE__GET_STATEMENT), and the user who created it.
type OTTStatus struct {
	Challenges []OTTChallenge
	// Validity is the time until the one-time token expires.
	Validity time.Duration
	// ActionType names the action the token authorizes, e.g.
	// "BALANCE__GET_STATEMENT" for a balance statement download.
	ActionType string
	UserID     int64
}

// PendingRequired returns the challenges that must still be passed to clear
// the OTT: required and not yet passed.
func (s *OTTStatus) PendingRequired() []OTTChallenge {
	if s == nil {
		return nil
	}

	pending := make([]OTTChallenge, 0, len(s.Challenges))
	for _, challenge := range s.Challenges {
		if challenge.Required && !challenge.Passed {
			pending = append(pending, challenge)
		}
	}

	return pending
}

// Cleared reports whether every required challenge has been passed — the
// state that unblocks the SCA-protected request.
func (s *OTTStatus) Cleared() bool {
	return len(s.PendingRequired()) == 0
}

// OTTTriggerResult is the outcome of triggering a challenge: the obfuscated
// phone number the one-time password was sent to, as a hint for the user.
type OTTTriggerResult struct {
	ObfuscatedPhoneNo string
}

// OTTCodeProvider supplies the one-time-password (OTP) for one challenge
// attempt — typically by prompting the user. phoneHint is the obfuscated
// phone number from the trigger call (may be empty). Returning an error
// aborts clearing.
type OTTCodeProvider func(
	ctx context.Context,
	challenge OTTChallenge,
	channel OTTChannel,
	phoneHint string,
) (string, error)

// maxOTTClearRounds bounds ClearSCAChallenge's loop: one round per required
// challenge plus slack. A token that never clears surfaces as a Rejection
// instead of an infinite trigger/verify cycle.
const maxOTTClearRounds = 5

// ottHeaders builds the extra-headers map the OTT endpoints require. The OTT
// travels in a header, never in a URL, body, or error message.
func ottHeaders(ott string) map[string]string {
	return map[string]string{HeaderOneTimeToken: ott}
}

// GetOTTStatus retrieves the challenge state of a one-time token
// (GET /2026Q3/one-time-token/status): which challenges must be passed to
// clear it, their types and alternatives, the remaining validity, and the
// authorized action. The personal API token authorizes the call; the OTT is
// the one issued in the x-2fa-approval header of an [SCAChallengeError].
func (c *Client) GetOTTStatus(ctx context.Context, ott string) (*OTTStatus, error) {
	if err := requireOTT(ott); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/%s/one-time-token/status", quarterlyAPIVersion)

	var response raw.OTTResponse

	if err := c.getWithQueryHeaders(ctx, path, nil, ottHeaders(ott), &response); err != nil {
		return nil, fmt.Errorf("get one-time-token status: %w", err)
	}

	return toOTTStatus(response)
}

// TriggerOTT sends the one-time password for the given channel to the
// user's verified phone number (POST /2026Q3/one-time-token/{channel}/trigger).
// The OTP arrives out of band; submit it with [Client.VerifyOTT]. Triggering
// is required before verifying. On the sandbox no real SMS is sent and the
// OTP is always "111111".
func (c *Client) TriggerOTT(ctx context.Context, ott string, channel OTTChannel) (*OTTTriggerResult, error) {
	if err := requireOTT(ott); err != nil {
		return nil, err
	}

	if !isPhoneChannel(channel) {
		return nil, errorfamily.NewRejection("wise.ott.invalid_request",
			fmt.Sprintf("channel must be one of sms, whatsapp, voice — got %q", channel))
	}

	path := fmt.Sprintf("/%s/one-time-token/%s/trigger", quarterlyAPIVersion, channel)

	var response raw.OTTTriggerResponse

	if err := c.postWithHeaders(ctx, path, nil, &response, ottHeaders(ott)); err != nil {
		return nil, fmt.Errorf("trigger %s challenge: %w", channel, err)
	}

	return &OTTTriggerResult{ObfuscatedPhoneNo: response.ObfuscatedPhoneNo}, nil
}

// VerifyOTT submits the one-time password for the given channel
// (POST /2026Q3/one-time-token/{channel}/verify) and returns the OTT's
// updated status: an empty pending-required set means the challenge is
// cleared and the SCA-protected request can be retried, optionally with the
// OTT in the x-2fa-approval header via [WithSCAApprovalToken].
func (c *Client) VerifyOTT(
	ctx context.Context,
	ott string,
	channel OTTChannel,
	otpCode string,
) (*OTTStatus, error) {
	if err := requireOTT(ott); err != nil {
		return nil, err //nolint:erraudit // otpCode is a secret — it must never appear in an error string
	}

	if !isPhoneChannel(channel) {
		return nil, errorfamily.NewRejection(
			"wise.ott.invalid_request",
			fmt.Sprintf(
				"channel must be one of sms, whatsapp, voice — got %q",
				channel,
			),
		) //nolint:erraudit // otpCode is a secret — it must never appear in an error string
	}

	if otpCode == "" {
		return nil, errorfamily.NewRejection("wise.ott.invalid_request",
			"otpCode is required — trigger the challenge first and submit the code it delivered")
	}

	path := fmt.Sprintf("/%s/one-time-token/%s/verify", quarterlyAPIVersion, channel)

	var response raw.OTTResponse

	if err := c.postWithHeaders(
		ctx, path, raw.OTTVerifyRequest{OTPCode: otpCode}, &response, ottHeaders(ott),
	); err != nil {
		return nil, fmt.Errorf(
			"verify %s challenge: %w",
			channel,
			err,
		) //nolint:erraudit // otpCode is a secret — it must never appear in an error string
	}

	return toOTTStatus(response)
}

// ClearSCAChallenge clears every required challenge of an OTT over one
// phone channel: it inspects the status, and per pending required challenge
// triggers the OTP, asks code for it, and verifies it, until the status
// reports cleared or an endpoint fails. Challenges that cannot be satisfied
// over the channel (e.g. a PIN-only challenge) surface as a Rejection naming
// the available types — clear those in the Wise web app by viewing a
// statement instead.
func (c *Client) ClearSCAChallenge(
	ctx context.Context,
	ott string,
	channel OTTChannel,
	code OTTCodeProvider,
) (*OTTStatus, error) {
	if code == nil {
		return nil, errorfamily.NewRejection("wise.ott.invalid_request",
			"code provider is required — ClearSCAChallenge needs the OTP only the user can supply")
	}

	status, err := c.GetOTTStatus(ctx, ott)
	if err != nil {
		return nil, fmt.Errorf("clear sca challenge: %w", err)
	}

	for range maxOTTClearRounds {
		pending := status.PendingRequired()
		if len(pending) == 0 {
			return status, nil
		}

		challenge := pending[0]
		if !challenge.Supports(channel) {
			return nil, rejectionUnsupportedChallenge(challenge, channel)
		}

		trigger, err := c.TriggerOTT(ctx, ott, channel)
		if err != nil {
			return nil, fmt.Errorf("clear sca challenge: %w", err)
		}

		phoneHint := ""
		if trigger != nil {
			phoneHint = trigger.ObfuscatedPhoneNo
		}

		otpCode, err := code(ctx, challenge, channel, phoneHint)
		if err != nil {
			return nil, fmt.Errorf("clear sca challenge: obtain otp code: %w", err)
		}

		status, err = c.VerifyOTT(ctx, ott, channel, otpCode)
		if err != nil {
			return nil, fmt.Errorf("clear sca challenge: %w", err)
		}
	}

	return nil, errorfamily.NewRejection("wise.ott.uncleared",
		fmt.Sprintf("one-time token still has required challenges after %d verify rounds — "+
			"inspect the status with GetOTTStatus and clear the rest manually", maxOTTClearRounds))
}

// rejectionUnsupportedChallenge builds the Rejection for a challenge the
// channel cannot satisfy, naming every way the challenge CAN be cleared so
// the operator can pick one.
func rejectionUnsupportedChallenge(challenge OTTChallenge, channel OTTChannel) *errorfamily.Error {
	types := make([]string, 0, 1+len(challenge.Alternatives))

	types = append(types, string(challenge.Primary.Type))
	for _, alt := range challenge.Alternatives {
		types = append(types, string(alt.Type))
	}

	return errorfamily.NewRejection("wise.ott.channel_unsupported",
		fmt.Sprintf("challenge of type %s cannot be cleared over the %s channel "+
			"(available for this challenge: %v) — clear it in the Wise web app by viewing a statement, "+
			"or verify a different channel",
			challenge.Primary.Type, channel, types))
}

// requireOTT rejects an empty one-time token before any request is made.
func requireOTT(ott string) error {
	if ott == "" {
		return errorfamily.NewRejection("wise.ott.invalid_request",
			"ott is required — the one-time token issued in the x-2fa-approval header of the SCA challenge")
	}

	return nil
}

// toOTTStatus maps the wire response to the parsed status. The wire's
// oneTimeToken echo is deliberately dropped: the caller passed the OTT in,
// and echoing it back into a printable struct only adds log-leak surface.
func toOTTStatus(response raw.OTTResponse) (*OTTStatus, error) {
	properties := response.OneTimeTokenProperties

	challenges := make([]OTTChallenge, 0, len(properties.Challenges))
	for _, wire := range properties.Challenges {
		alternatives, err := parseOTTAlternatives(wire.Alternatives)
		if err != nil {
			return nil, errorfamily.WrapCorruption(err, "wise.ott.decode",
				"decode one-time-token challenges")
		}

		challenges = append(challenges, OTTChallenge{
			Primary: OTTChallengeView{
				Type:     OTTChallengeType(wire.PrimaryChallenge.Type),
				ViewData: wire.PrimaryChallenge.ViewData,
			},
			Alternatives: alternatives,
			Required:     wire.Required,
			Passed:       wire.Passed,
		})
	}

	return &OTTStatus{
		Challenges: challenges,
		Validity:   time.Duration(properties.Validity) * time.Second,
		ActionType: properties.ActionType,
		UserID:     properties.UserID,
	}, nil
}

// errOTTAlternativesShape is the sentinel for alternatives payloads that are
// neither the documented object form nor a bare string array.
var errOTTAlternativesShape = errors.New("alternatives are neither objects nor strings")

// parseOTTAlternatives decodes a challenge's alternatives leniently: the
// OpenAPI spec types the items only as "object" with no properties, so both
// the documented object form ({"type": ...}) and a bare string form are
// accepted; anything else decodes to no alternatives rather than failing the
// whole status.
func parseOTTAlternatives(value jsontext.Value) ([]OTTChallengeView, error) {
	if len(value) == 0 {
		return nil, nil
	}

	var objectForm []struct {
		Type     string         `json:"type"`
		ViewData map[string]any `json:"viewData"`
	}

	if err := json.Unmarshal(value, &objectForm); err == nil {
		views := make([]OTTChallengeView, 0, len(objectForm))
		for _, alternative := range objectForm {
			views = append(views, OTTChallengeView{
				Type:     OTTChallengeType(alternative.Type),
				ViewData: alternative.ViewData,
			})
		}

		return views, nil
	}

	var stringForm []string
	if err := json.Unmarshal(value, &stringForm); err == nil {
		views := make([]OTTChallengeView, 0, len(stringForm))
		for _, alternative := range stringForm {
			views = append(views, OTTChallengeView{
				Type:     OTTChallengeType(alternative),
				ViewData: nil,
			})
		}

		return views, nil
	}

	return nil, fmt.Errorf("%w: %s", errOTTAlternativesShape, value.String())
}
