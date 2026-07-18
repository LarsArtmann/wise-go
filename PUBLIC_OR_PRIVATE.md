# Public or Private? — Decision Document

**Status:** DECISION PENDING (not blocking; private is the safe default)
**Last updated:** 2026-07-18
**Current state:** Private repo, `PROPRIETARY LICENSE` (all rights reserved)

---

## TL;DR

Recommendation: **Go public, but not today, and not in this shape.**

The asymmetry favors going public — low risk, high upside, mostly documentation work. The only genuine blocker is the `GOEXPERIMENT=jsonv2` friction. Solve that (by waiting for Go to graduate the experiment, or by pinning deps back), then flip.

Sequenced plan in [Recommended path](#recommended-path) below.

---

## Critical fact that overrides everything

**The current `LICENSE` makes going public pointless.** It says:

> Unauthorized copying, distribution, modification, or use of this Software … is strictly prohibited.

Going public without changing this means anyone can _read_ the code but nobody can legally `go get` it, use it, or contribute. The single most important decision is **which OSI license to adopt**, not whether to flip the visibility toggle.

---

## PRO

| #   | Argument                                                                                                                                                                                                                                                                                                                                                                                               | Weight |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ |
| 1   | **Fills a real ecosystem gap.** Wise publishes no official Go SDK and no complete OpenAPI spec. The Go ecosystem has nothing here. This is not a "yet another X" project.                                                                                                                                                                                                                              | High   |
| 2   | **Marketing for the whole library ecosystem.** `go-branded-id` and `go-error-family` are already public on the Go proxy. wise-go is the most compelling real-world showcase of both — phantom IDs across entities, behavioral error classification, the two-layer type design. Currently that showcase is invisible. Public wise-go is the funnel that makes the other two libs "click" for newcomers. | High   |
| 3   | **No IP or secret risk.** Verified: API keys are user-supplied, tests use `httptest` (no real Wise data), CI secrets live in GitHub Actions `${{ secrets }}` (stay private even when repo is public), no hardcoded credentials. The Wise REST API is itself public — wrapping it is not derivative IP.                                                                                                 | High   |
| 4   | **Standard Go ergonomics for consumers.** `go get github.com/larsartmann/wise-go` just works. pkg.go.dev auto-publishes docs. The Go proxy caches + serves. Sum DB verifies. Today every consumer needs `GOPRIVATE=github.com/larsartmann` + SSH access — friction you'd never impose on yourself for a public SDK.                                                                                    | High   |
| 5   | **External eyes surface bugs earlier.** The CARD_PAYMENT classification bug survived one private session (see `CHANGELOG.md` v0.3.0 Fixed). Different usage patterns stress-test the retry policy, error classification, and edge cases (`amount == 0`, `Retry-After` HTTP-date parsing, empty balance lists) that haven't been hit yet.                                                               | Medium |
| 6   | **Already production-shaped.** ~95% coverage, 0 lint issues, comprehensive CHANGELOG/FEATURES/ROADMAP/AGENTS, semver tags, hardened CI, BDD tests. This is not embarrassing to show the world. Most "I'll open source it when it's ready" projects never ship — this one is ready.                                                                                                                     | Medium |
| 7   | **Reputation + portfolio.** A shipped, well-engineered SDK with a real domain (fintech, money, type safety) is a stronger signal than any blog post. Demonstrates the "data models first" and "strong types over runtime checks" philosophy in production.                                                                                                                                             | Medium |
| 8   | **Community SDKs are normal in fintech.** Cf. `stripe-go` official, but countless unofficial AWS/Twilio SDKs coexisting. Wise's developer platform has a community category. No evidence Wise forbids third-party SDKs (to be verified in [Legal pre-flight](#step-1--legal-pre-flight)).                                                                                                              | Low    |

---

## CONTRA

| #   | Argument                                                                                                                                                                                                                                                                                                                                                                                                                        | Weight |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| 1   | **Implicit SLA.** Public = users will file issues, request write operations (transfers, recipients, quotes, webhooks), expect Wise API drift to be tracked. Solo maintainer with many projects. Private = no obligations.                                                                                                                                                                                                       | High   |
| 2   | **`GOEXPERIMENT=jsonv2` is a hard sell for casual adopters.** Every consumer must `export GOEXPERIMENT=jsonv2` or hit `build constraints exclude all Go files in encoding/json/v2`. Documented well, but for a public SDK this is friction at the front door. Solve by either waiting for jsonv2 to graduate (likely Go 1.27/1.28) or pinning back to `go-branded-id v0.3.1` + `go-error-family v0.6.x` for a public release.   | High   |
| 3   | **Legal pre-flight work before flipping the switch.** (a) Replace `LICENSE` with MIT or Apache-2.0. (b) Add trademark disclaimer to README — _"Wise" and "TransferWise" are registered trademarks of Wise Payments Limited. This project is not affiliated with or endorsed by Wise._ (c) Verify Wise's developer ToS explicitly permits third-party SDKs. (d) Add `SECURITY.md` for vuln disclosure. None of this is optional. | High   |
| 4   | **Reputational exposure on money-handling code.** If a consumer loses money due to a bug (e.g., the CARD_PAYMENT bug that just got fixed in v0.3.0 — refunds were misclassified), blame follows. The `LICENSE` "AS IS" clause is the legal shield; reputation is the real one. Public + pre-1.0 + breaking changes = users will get hurt eventually.                                                                            | Medium |
| 5   | **Future Wise official SDK conflict.** If Wise ships an official Go SDK later, the namespace (`wise-go`), naming, and consumer mindshare get awkward. Unlikely (they've had years), but possible. Low historical risk — unofficial SDKs usually coexist.                                                                                                                                                                        | Low    |
| 6   | **v0.3.0 is pre-1.0.** Semver permits breaking changes; reality is that public consumers still get hurt. Either commit to a 1.0 freeze soon after going public, or be very loud about the "early development" status (README already does this).                                                                                                                                                                                | Low    |
| 7   | **Maintenance scope creep pressure.** README/ROADMAP currently say "write operations not yet implemented." Public users will push for these. Either build them (scope expansion) or say no publicly (social cost).                                                                                                                                                                                                              | Low    |

---

## Recommended path

### Step 1 — Legal pre-flight (reversible, ~1–2 hours)

Non-negotiable prerequisites for any public release. Do these first; they cost nothing and unlock the option.

- [ ] **Change `LICENSE`** → Apache-2.0 (patent grant is the right choice for fintech-adjacent code; MIT is fine if minimalism matters more). Apache-2.0 is the recommendation.
- [ ] **Trademark disclaimer** in README — _"Wise" and "TransferWise" are registered trademarks of Wise Payments Limited. This project is not affiliated with or endorsed by Wise._
- [ ] **Verify Wise developer ToS** permits third-party SDKs before going public. Link the finding here.
- [ ] **Add `SECURITY.md`** with a vulnerability disclosure policy (private email → GitHub Security Advisory).
- [ ] **Audit for secrets** one more time — `git log -p` across history, not just HEAD. `gitleaks` is already in the buildflow pipeline; run a full-history scan.

### Step 2 — Resolve the jsonv2 question (blocking)

Two options, mutually exclusive:

- **Option A (recommended): Keep v0.3.x private.** Wait for `GOEXPERIMENT=jsonv2` to graduate to default in Go (likely Go 1.27 or 1.28). Then the public release has zero friction. Cleanest, most adoption-friendly.
- **Option B: Cut a public v0.4.0 that pins back.** `go-branded-id v0.3.1` + `go-error-family v0.6.x`. Removes the experiment requirement at the cost of newer features in both deps. Only choose this if there's external pressure to go public _now_.

Option A is the pick. The waiting cost is low (private repo keeps working); the adoption cost of Option B is permanent friction for any consumer who wants the newer dep features later.

### Step 3 — Flip the repo public

After steps 1 and 2:

- [ ] Toggle repo visibility in GitHub settings.
- [ ] Remove `GOPRIVATE=github.com/larsartmann` from `flake.nix` (no longer needed once public — actually, keep it if other `larsartmann/*` repos stay private; otherwise the proxy will be hit correctly without it).
- [ ] Verify `go list -m github.com/larsartmann/wise-go@latest` resolves via the public Go proxy.
- [ ] Verify pkg.go.dev picks up the docs (request indexing at https://pkg.go.dev/github.com/larsartmann/wise-go).
- [ ] Submit to: [awesome-go](https://github.com/avelino/awesome-go) (if it fits a category), the Wise developer community/Discord, a post on r/golang, and a blog post on lars.software.

### Step 4 — Post-public hygiene

- [ ] Commit to a 1.0 freeze timeline (e.g., "1.0 when write operations + webhooks land").
- [ ] Decide issue/PR triage SLA (e.g., "best effort, no guarantees, pre-1.0").
- [ ] Document the support policy in README and CONTRIBUTING.md.

---

## Decision log

| Date       | Decision                       | Rationale                                                                                                                           |
| ---------- | ------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------- |
| 2026-07-18 | Document drafted; no flip yet. | See [Recommended path](#recommended-path). The jsonv2 friction (Contra #2) and legal pre-flight (Contra #3) must be resolved first. |

---

## Revisit triggers

Re-open this decision earlier than planned if any of these happen:

- Wise announces an official Go SDK (changes the ecosystem calculus — Contra #5 becomes live).
- `GOEXPERIMENT=jsonv2` graduates to default in Go (removes Contra #2 — accelerates Step 2 Option A).
- A specific external user asks to consume wise-go (changes the cost/benefit of staying private).
- `go-branded-id` or `go-error-family` adoption stalls and wise-go's visibility would unblock it (amplifies Pro #2).
