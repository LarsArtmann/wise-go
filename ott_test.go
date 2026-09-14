package wise_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// testOTT is a synthetic one-time token for fixtures. It is NOT a credential;
// it exists so tests can assert the OTT never leaks into error strings.
const testOTT = "9f5f5812-2609-4e48-8418-b64437c0c7cd"

// errSCACodeAborted is a static sentinel (err113: never construct dynamic
// errors) standing in for an operator aborting the OTP prompt.
var errSCACodeAborted = errors.New("user aborted the prompt")

// ottStatusJSON builds a one-time-token status response body. challengesJSON
// must be a valid JSON array literal (may be empty).
func ottStatusJSON(challengesJSON string, validity int64, actionType string) string {
	return fmt.Sprintf(
		`{"oneTimeTokenProperties":{"oneTimeToken":%q,"challenges":%s,"validity":%d,"actionType":%q,"userId":6146956}}`,
		testOTT, challengesJSON, validity, actionType,
	)
}

// ottSMSChallengeJSON is one required, unpassed challenge with an SMS primary
// and phone-channel alternatives in the documented object form.
const ottSMSChallengeJSON = `[
	{
		"primaryChallenge": {"type": "SMS", "viewData": {"attributes": {"phone": "+49 *** 123"}}},
		"alternatives": [
			{"type": "WHATSAPP", "viewData": {"attributes": {"phone": "+49 *** 123"}}},
			{"type": "VOICE"}
		],
		"required": true,
		"passed": false
	}
]`

// ottPINOnlyChallengeJSON is a challenge no phone channel can clear.
const ottPINOnlyChallengeJSON = `[
	{
		"primaryChallenge": {"type": "PIN", "viewData": {"attributes": {"userId": 6146956}}},
		"alternatives": [],
		"required": true,
		"passed": false
	}
]`

// ottHandler serves one OTT endpoint and asserts the shared contract:
// method, the One-Time-Token header, and the bearer token.
func ottHandler(method string, serve func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Expect(r.Method).To(Equal(method))
		Expect(r.Header.Get("One-Time-Token")).To(Equal(testOTT))
		Expect(r.Header.Get("Authorization")).To(Equal("Bearer test-api-key"))
		serve(w, r)
	}
}

// ottPINWithSMSAlternativeJSON is a PIN-primary challenge an SMS
// alternative CAN clear.
const ottPINWithSMSAlternativeJSON = `[
	{
		"primaryChallenge": {"type": "PIN", "viewData": {"attributes": {"userId": 6146956}}},
		"alternatives": [{"type": "SMS"}],
		"required": true,
		"passed": false
	}
]`

var _ = Describe("OTT (SCA one-time-token endpoints)", func() {
	var (
		server *httptest.Server
		mux    *http.ServeMux
		client *wise.Client
	)

	BeforeEach(func() {
		mux = http.NewServeMux()
		server = httptest.NewServer(mux)
		client = wise.New("test-api-key", wise.WithBaseURL(server.URL))
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("GetOTTStatus", func() {
		It("should fetch and parse the challenge state", func() {
			mux.HandleFunc("/2026Q3/one-time-token/status", ottHandler(http.MethodGet, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(ottStatusJSON(ottSMSChallengeJSON, 3600, "BALANCE__GET_STATEMENT")))
			}))

			status, err := client.GetOTTStatus(context.Background(), testOTT)
			Expect(err).ToNot(HaveOccurred())

			Expect(status.ActionType).To(Equal("BALANCE__GET_STATEMENT"))
			Expect(status.UserID).To(Equal(int64(6146956)))
			Expect(status.Validity).To(Equal(time.Hour))

			Expect(status.Challenges).To(HaveLen(1))
			challenge := status.Challenges[0]
			Expect(challenge.Required).To(BeTrue())
			Expect(challenge.Passed).To(BeFalse())
			Expect(string(challenge.Primary.Type)).To(Equal("SMS"))
			Expect(challenge.Primary.ViewData).To(HaveKey("attributes"))
			Expect(challenge.Alternatives).To(HaveLen(2))
			Expect(string(challenge.Alternatives[0].Type)).To(Equal("WHATSAPP"))
			Expect(string(challenge.Alternatives[1].Type)).To(Equal("VOICE"))

			Expect(status.Cleared()).To(BeFalse())
			Expect(status.PendingRequired()).To(HaveLen(1))
		})

		It("should decode string-form alternatives leniently", func() {
			mux.HandleFunc("/2026Q3/one-time-token/status", ottHandler(http.MethodGet, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(ottStatusJSON(`[
					{"primaryChallenge": {"type": "PIN"}, "alternatives": ["SMS", "VOICE"], "required": true, "passed": false}
				]`, 60, "BALANCE__GET_STATEMENT")))
			}))

			status, err := client.GetOTTStatus(context.Background(), testOTT)
			Expect(err).ToNot(HaveOccurred())
			Expect(status.Challenges[0].Alternatives).To(HaveLen(2))
			Expect(string(status.Challenges[0].Alternatives[0].Type)).To(Equal("SMS"))
			Expect(string(status.Challenges[0].Alternatives[1].Type)).To(Equal("VOICE"))
			Expect(status.Challenges[0].Supports(wise.OTTChannelSMS)).To(BeTrue())
		})

		It("should treat an empty challenge list as cleared", func() {
			mux.HandleFunc("/2026Q3/one-time-token/status", ottHandler(http.MethodGet, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(ottStatusJSON(`[]`, 3600, "")))
			}))

			status, err := client.GetOTTStatus(context.Background(), testOTT)
			Expect(err).ToNot(HaveOccurred())
			Expect(status.Cleared()).To(BeTrue())
			Expect(status.PendingRequired()).To(BeEmpty())
		})

		It("should reject an empty OTT without calling the API", func() {
			_, err := client.GetOTTStatus(context.Background(), "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ott is required"))
			classified := errorfamily.Classify(err)
			Expect(classified).To(Equal(errorfamily.Rejection))
		})

		It("should surface API errors without leaking the OTT", func() {
			mux.HandleFunc("/2026Q3/one-time-token/status", ottHandler(http.MethodGet, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"errors":[{"code":"NOT_AUTHORISED","message":"nope"}]}`))
			}))

			_, err := client.GetOTTStatus(context.Background(), testOTT)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).ToNot(ContainSubstring(testOTT), "the OTT must never appear in an error string")
		})

		It("should classify a malformed response as corruption", func() {
			mux.HandleFunc("/2026Q3/one-time-token/status", ottHandler(http.MethodGet, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"oneTimeTokenProperties":{"challenges":`))
			}))

			_, err := client.GetOTTStatus(context.Background(), testOTT)
			Expect(err).To(HaveOccurred())
			Expect(errorfamily.Classify(err)).To(Equal(errorfamily.Corruption))
		})
	})

	Describe("TriggerOTT", func() {
		It("should POST the channel trigger and return the phone hint", func() {
			mux.HandleFunc("/2026Q3/one-time-token/sms/trigger", ottHandler(http.MethodPost, func(
				w http.ResponseWriter, r *http.Request,
			) {
				body, readErr := io.ReadAll(r.Body)
				Expect(readErr).ToNot(HaveOccurred())
				Expect(body).To(BeEmpty(), "the trigger endpoints take no body")

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"obfuscatedPhoneNo":"*********8888"}`))
			}))

			result, err := client.TriggerOTT(context.Background(), testOTT, wise.OTTChannelSMS)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.ObfuscatedPhoneNo).To(Equal("*********8888"))
		})

		It("should reject a non-phone channel without calling the API", func() {
			_, err := client.TriggerOTT(context.Background(), testOTT, wise.OTTChannel("pin"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("channel must be one of sms, whatsapp, voice"))
		})

		It("should reject an empty OTT without calling the API", func() {
			_, err := client.TriggerOTT(context.Background(), "", wise.OTTChannelSMS)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ott is required"))
		})
	})

	Describe("VerifyOTT", func() {
		It("should POST the OTP code and return the updated status", func() {
			mux.HandleFunc("/2026Q3/one-time-token/sms/verify", ottHandler(http.MethodPost, func(
				w http.ResponseWriter, r *http.Request,
			) {
				Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

				var body struct {
					OTPCode string `json:"otpCode"`
				}
				Expect(json.UnmarshalRead(r.Body, &body)).To(Succeed())
				Expect(body.OTPCode).To(Equal("111111"))

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(ottStatusJSON(`[]`, 3600, "")))
			}))

			status, err := client.VerifyOTT(context.Background(), testOTT, wise.OTTChannelSMS, "111111")
			Expect(err).ToNot(HaveOccurred())
			Expect(status.Cleared()).To(BeTrue())
		})

		It("should surface a wrong code as a typed API error", func() {
			mux.HandleFunc("/2026Q3/one-time-token/sms/verify", ottHandler(http.MethodPost, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"errors":[{"code":"OTPCODE_INVALID","message":"otp code didn't match"}]}`))
			}))

			_, err := client.VerifyOTT(context.Background(), testOTT, wise.OTTChannelSMS, "000000")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("verify sms challenge"))
			Expect(err.Error()).ToNot(ContainSubstring(testOTT))
		})

		It("should reject an empty OTP code without calling the API", func() {
			_, err := client.VerifyOTT(context.Background(), testOTT, wise.OTTChannelSMS, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("otpCode is required"))
		})
	})

	Describe("ClearSCAChallenge", func() {
		It("should trigger, prompt, and verify until cleared", func() {
			var (
				statusCalls atomic.Int64
				triggers    atomic.Int64
				verifies    atomic.Int64
			)

			mux.HandleFunc("/2026Q3/one-time-token/status", ottHandler(http.MethodGet, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				statusCalls.Add(1)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(ottStatusJSON(ottSMSChallengeJSON, 3600, "BALANCE__GET_STATEMENT")))
			}))
			mux.HandleFunc("/2026Q3/one-time-token/sms/trigger", ottHandler(http.MethodPost, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				triggers.Add(1)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"obfuscatedPhoneNo":"*********8888"}`))
			}))
			mux.HandleFunc("/2026Q3/one-time-token/sms/verify", ottHandler(http.MethodPost, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				verifies.Add(1)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(ottStatusJSON(`[]`, 3600, "")))
			}))

			var seenHint, seenChannel, seenType string

			status, err := client.ClearSCAChallenge(context.Background(), testOTT, wise.OTTChannelSMS,
				func(_ context.Context, challenge wise.OTTChallenge, channel wise.OTTChannel, phoneHint string) (string, error) {
					seenHint = phoneHint
					seenChannel = channel.String()
					seenType = string(challenge.Primary.Type)
					return "111111", nil
				})
			Expect(err).ToNot(HaveOccurred())
			Expect(status.Cleared()).To(BeTrue())
			Expect(statusCalls.Load()).To(Equal(int64(1)), "the verify response is the new status — no re-fetch needed")
			Expect(triggers.Load()).To(Equal(int64(1)))
			Expect(verifies.Load()).To(Equal(int64(1)))
			Expect(seenHint).To(Equal("*********8888"))
			Expect(seenChannel).To(Equal("sms"))
			Expect(seenType).To(Equal("SMS"))
		})

		It("should clear a second remaining challenge in a second round", func() {
			var verifies atomic.Int64

			mux.HandleFunc("/2026Q3/one-time-token/status", ottHandler(http.MethodGet, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(ottStatusJSON(`[
					{"primaryChallenge": {"type": "SMS"}, "alternatives": [], "required": true, "passed": true},
					{"primaryChallenge": {"type": "SMS"}, "alternatives": [], "required": true, "passed": false}
				]`, 3600, "BALANCE__GET_STATEMENT")))
			}))
			mux.HandleFunc("/2026Q3/one-time-token/sms/trigger", ottHandler(http.MethodPost, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"obfuscatedPhoneNo":"*********8888"}`))
			}))
			mux.HandleFunc("/2026Q3/one-time-token/sms/verify", ottHandler(http.MethodPost, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				round := verifies.Add(1)
				w.Header().Set("Content-Type", "application/json")
				if round == 1 {
					_, _ = w.Write([]byte(ottStatusJSON(`[
						{"primaryChallenge": {"type": "SMS"}, "alternatives": [], "required": true, "passed": true},
						{"primaryChallenge": {"type": "SMS"}, "alternatives": [], "required": true, "passed": false}
					]`, 3600, "BALANCE__GET_STATEMENT")))
					return
				}
				_, _ = w.Write([]byte(ottStatusJSON(`[]`, 3600, "")))
			}))

			status, err := client.ClearSCAChallenge(context.Background(), testOTT, wise.OTTChannelSMS,
				func(context.Context, wise.OTTChallenge, wise.OTTChannel, string) (string, error) {
					return "111111", nil
				})
			Expect(err).ToNot(HaveOccurred())
			Expect(status.Cleared()).To(BeTrue())
			Expect(verifies.Load()).To(Equal(int64(2)))
		})

		It("should clear via an alternative when the primary is not the channel", func() {
			mux.HandleFunc("/2026Q3/one-time-token/status", ottHandler(http.MethodGet, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(ottStatusJSON(ottPINWithSMSAlternativeJSON, 3600, "BALANCE__GET_STATEMENT")))
			}))
			mux.HandleFunc("/2026Q3/one-time-token/sms/trigger", ottHandler(http.MethodPost, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"obfuscatedPhoneNo":"*********8888"}`))
			}))
			mux.HandleFunc("/2026Q3/one-time-token/sms/verify", ottHandler(http.MethodPost, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(ottStatusJSON(`[]`, 3600, "")))
			}))

			status, err := client.ClearSCAChallenge(context.Background(), testOTT, wise.OTTChannelSMS,
				func(context.Context, wise.OTTChallenge, wise.OTTChannel, string) (string, error) {
					return "111111", nil
				})
			Expect(err).ToNot(HaveOccurred())
			Expect(status.Cleared()).To(BeTrue())
		})

		It("should reject when no phone channel can clear the challenge", func() {
			mux.HandleFunc("/2026Q3/one-time-token/status", ottHandler(http.MethodGet, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(ottStatusJSON(ottPINOnlyChallengeJSON, 3600, "BALANCE__GET_STATEMENT")))
			}))

			_, err := client.ClearSCAChallenge(context.Background(), testOTT, wise.OTTChannelSMS,
				func(context.Context, wise.OTTChallenge, wise.OTTChannel, string) (string, error) {
					return "111111", nil
				})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("channel_unsupported"))
			Expect(err.Error()).To(ContainSubstring("PIN"))
		})

		It("should abort when the code provider fails", func() {
			mux.HandleFunc("/2026Q3/one-time-token/status", ottHandler(http.MethodGet, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(ottStatusJSON(ottSMSChallengeJSON, 3600, "BALANCE__GET_STATEMENT")))
			}))
			mux.HandleFunc("/2026Q3/one-time-token/sms/trigger", ottHandler(http.MethodPost, func(
				w http.ResponseWriter, _ *http.Request,
			) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"obfuscatedPhoneNo":"*********8888"}`))
			}))

			providerErr := errSCACodeAborted
			_, err := client.ClearSCAChallenge(context.Background(), testOTT, wise.OTTChannelSMS,
				func(context.Context, wise.OTTChallenge, wise.OTTChannel, string) (string, error) {
					return "", providerErr
				})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("user aborted the prompt"))
		})

		It("should require a code provider", func() {
			_, err := client.ClearSCAChallenge(context.Background(), testOTT, wise.OTTChannelSMS, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("code provider is required"))
		})
	})

	Describe("OTTChallenge.Supports", func() {
		It("should match the primary and the alternatives", func() {
			challenge := wise.OTTChallenge{Primary: wise.OTTChallengeView{Type: wise.OTTChallengePIN}}
			Expect(challenge.Supports(wise.OTTChannelSMS)).To(BeFalse())

			challenge.Alternatives = []wise.OTTChallengeView{{Type: wise.OTTChallengeWhatsApp}}
			Expect(challenge.Supports(wise.OTTChannelWhatsApp)).To(BeTrue())
			Expect(challenge.Supports(wise.OTTChannelVoice)).To(BeFalse())
		})
	})
})
