# Security Policy

## Supported versions

Only the latest tagged release receives security fixes. Older tags are
served by the Go module proxy but are not patched; upgrade before reporting
or deploying.

## Reporting a vulnerability

**Do not open a public issue for a security vulnerability.**

Use GitHub's private vulnerability reporting on this repository:
**Security → Report a vulnerability**
(<https://github.com/LarsArtmann/wise-go/security/advisories/new>).

Include what you can of:

- The affected module path and version (`github.com/larsartmann/wise-go@vX.Y.Z`)
- A minimal reproduction (code snippet, request/response pair, or failing test)
- The impact you see (what an attacker could do, what data is exposed)
- Any workaround you have verified

## Security Practices

- Secrets stay out of the repo and out of logs: the API key is supplied per
  client (`WithAPIKey`) and never rendered into error strings; SCA one-time
  tokens are exposed only via `TwoFAApprovalToken()`, and `SCAChallengeError`
  reports `token issued: true/false` — never the token itself.
- Webhook deliveries must be authenticated with `VerifyWebhookSignature`
  (RSA signature over the raw request bytes) before parsing; malformed
  envelopes are rejected as corruption-classified errors rather than trusted.
- Every change is gated by `go test -race`, `golangci-lint`, and the hermetic
  `nix flake check`; govulncheck runs in CI once re-enabled. Live-API tests
  execute only against the sandbox and only when a key is explicitly provided.

## What to report

Anything that would let an attacker or a malicious response compromise a
consumer of this SDK, for example:

- Response parsing that misclassifies or drops security-relevant API errors
  (auth, SCA challenges, rate limits)
- Signature verification flaws in `VerifyWebhookSignature` / webhook decoding
- Secret handling: anything that could leak the API key, SCA one-time tokens,
  or webhook payloads into logs or error strings
- Injection or unsafe deserialization in request/response paths

## What is out of scope

- Vulnerabilities in the Wise API itself — report those to Wise
  (<https://bugcrowd.com/wise>)
- Missing features, hardening suggestions without an attack path — those are
  regular issues
- Bypasses that require control of the consumer's own process or environment

## Handling commitment

Maintainers triage reports and respond within 7 days, publish a fix and a
patch release for accepted vulnerabilities, and credit reporters in the
release notes unless they prefer to stay anonymous.
