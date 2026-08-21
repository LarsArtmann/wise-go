package wise

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// HeaderWebhookSignature is the header Wise signs every webhook delivery
// with: a Base64-encoded RSA-SHA256 signature of the raw request body.
const HeaderWebhookSignature = "X-Signature-SHA256"

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
