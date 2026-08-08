# Public or Private? — Decision Document

**Status:** APPROVED — go public. Legal pre-flight nearly complete.
**Last updated:** 2026-08-08
**Current state:** Private repo, `Apache-2.0` license (converted 2026-08-08)

---

## TL;DR

Recommendation: **Go public.** The two blockers that held this back are now resolved — the license is Apache-2.0, and the `GOEXPERIMENT=jsonv2` friction is accepted as documented cost-of-doing-business (the maintainer uses it everywhere already; consumers who want this SDK are sophisticated enough to set one env var, and jsonv2 will graduate to default soon anyway).

All legal pre-flight items are complete: Apache-2.0 license, trademark disclaimer in README, Wise ToS verified (no prohibition found), and full-history gitleaks scan clean (0 leaks across 84 commits). Ready to flip. Sequenced plan in [Recommended path](#recommended-path) below.

---

## Critical fact that overrides everything

**All legal pre-flight is complete.** Apache-2.0 license in place, trademark disclaimer in README, Wise ToS verified (no prohibition found — see [Wise ToS verification](#wise-tos-verification)), and full-history secrets scan clean (0 leaks across 84 commits). The only remaining step is flipping the repo visibility toggle.

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
| 2   | **`GOEXPERIMENT=jsonv2` is friction for casual adopters** *(accepted, 2026-08-08)*. Every consumer must `export GOEXPERIMENT=jsonv2` or hit `build constraints exclude all Go files in encoding/json/v2`. Documented in README and AGENTS.md. **Decision: accept.** The maintainer already runs jsonv2 everywhere; consumers who want a fintech SDK are sophisticated enough to set one env var; and jsonv2 will graduate to default in Go (likely 1.27/1.28), at which point this vanishes. Not a blocker — just document it well. | Accepted |
| 3   | **Legal pre-flight work before flipping the switch.** (a) ~~Replace `LICENSE` with Apache-2.0~~ **Done (2026-08-08).** (b) Add trademark disclaimer to README — _"Wise" and "TransferWise" are registered trademarks of Wise Payments Limited. This project is not affiliated with or endorsed by Wise._ (c) Verify Wise's developer ToS explicitly permits third-party SDKs. (~~d~~) ~~Add `SECURITY.md`~~ **Dropped — maintainer opts out.** Remaining: (b) and (c) only. | Low    |
| 4   | **Reputational exposure on money-handling code.** If a consumer loses money due to a bug (e.g., the CARD_PAYMENT bug that got fixed in v0.3.0 — refunds were misclassified), blame follows. The Apache-2.0 `AS IS` clause is the legal shield; reputation is the real one. Public + pre-1.0 + breaking changes = users will get hurt eventually.                                                                            | Medium |
| 5   | **Future Wise official SDK conflict.** If Wise ships an official Go SDK later, the namespace (`wise-go`), naming, and consumer mindshare get awkward. Unlikely (they've had years), but possible. Low historical risk — unofficial SDKs usually coexist.                                                                                                                                                                        | Low    |
| 6   | **v0.5.0 is pre-1.0.** Semver permits breaking changes; reality is that public consumers still get hurt. Either commit to a 1.0 freeze soon after going public, or be very loud about the "early development" status (README already does this).                                                                                                                                                                                | Low    |
| 7   | **Maintenance scope creep pressure.** README/ROADMAP currently say "write operations not yet implemented." Public users will push for these. Either build them (scope expansion) or say no publicly (social cost).                                                                                                                                                                                                              | Low    |

---

## Recommended path

### Step 1 — Legal pre-flight: COMPLETE

All items done. Ready to flip.

- [x] **Change `LICENSE`** → Done: Apache-2.0 (2026-08-08). Patent grant is the right choice for fintech-adjacent code.
- [x] **Trademark disclaimer** in README — Done (2026-08-08). Added "Trademarks" section: _"Wise" and "TransferWise" are registered trademarks of Wise Payments Limited. This project is not affiliated with, endorsed by, or sponsored by Wise Payments Limited._
- [x] **Verify Wise developer ToS** permits third-party SDKs. See [Wise ToS verification](#wise-tos-verification) below. Finding: no explicit prohibition found, no explicit permission either — risk is low.
- [x] **Audit for secrets** — Done (2026-08-08). `gitleaks detect --source . --log-opts="--all"` scanned all 84 commits (950 KB). **0 leaks found.**

### Wise ToS verification

**Date:** 2026-08-08
**Finding:** No explicit prohibition of third-party SDKs. No explicit permission either.

**Sources reviewed:**

1. [Wise Intellectual Property](https://wise.com/help/articles/79CxCv9Qj1r7mDPJuIwUA3/wise-intellectual-property) — claims broad ownership of "API, developer tools, source code, code libraries" as Wise's exclusive property. Grants customers a "revocable, non-exclusive, non-sublicensable, non-transferable, royalty-free limited license" for personal use. States "Any use not specifically permitted is strictly prohibited."
2. [Wise Customer Agreement (UK)](https://wise.com/gb/legal/terms-of-use-personal) — defines "API Partner" as "a business we have partnered with," implying the formal API channel is partnership-based. Section 8.3(c) prohibits "Infringing Wise's Intellectual Property." Section 8.2(c) prohibits using robots/spiders to "monitor or copy our websites" without permission.
3. [Wise Platform](https://wise.com/platform/) — aimed at banks, financial institutions, and enterprises. No developer-specific terms of service found publicly.
4. [api-docs.wise.com](https://api-docs.wise.com) — JS-rendered SPA; no public developer agreement visible without authentication.

**Analysis:**

- The IP and customer terms are written for **customers using Wise's services** (sending money, holding accounts). They are not a developer/API-specific terms of service.
- Building an independent SDK that calls a publicly documented REST API is standard industry practice. The SDK does not reproduce Wise's code — it independently implements HTTP calls to documented endpoints.
- The Wise API is accessible to any Wise account holder who generates an API token. The SDK is read-only and does not interfere with Wise's services.
- The trademark disclaimer in the README protects against passing-off claims.
- The broad "Any use not specifically permitted is strictly prohibited" clause could theoretically be read to cover third-party SDKs, but no fintech company is known to enforce against community SDKs that wrap their public API.

**Risk assessment: Low.** No prohibition found. Standard practice. Trademark disclaimer in place. The remaining risk is that Wise could change their terms in the future — the revisit trigger for "Wise announces an official Go SDK" covers this.

### Step 2 — jsonv2 question: RESOLVED

**Decision (2026-08-08): Keep `GOEXPERIMENT=jsonv2`. Accept the friction.**

The original two options were (A) wait for jsonv2 to graduate, or (B) pin back deps. Chosen: neither — ship as-is. Rationale:

- The maintainer already runs jsonv2 across every project. It's not exotic in this workflow.
- A consumer who wants a Wise fintech SDK is sophisticated enough to set one env var.
- jsonv2 will graduate to default in Go (likely 1.27/1.28), at which point this vanishes with zero action.
- README and AGENTS.md already document the requirement clearly.

Not a blocker.

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
| 2026-08-08 | **Legal pre-flight complete.** Trademark disclaimer added, Wise ToS verified (no prohibition), gitleaks clean (0 leaks / 84 commits). | All Step 1 items checked off. Ready to flip visibility toggle. |
| 2026-08-08 | **Approved going public.** Converted LICENSE to Apache-2.0; accepted jsonv2 friction; dropped SECURITY.md requirement. | Both original blockers (license + jsonv2) resolved. Remaining work is trademark disclaimer + Wise ToS verification. See [Recommended path](#recommended-path). |
| 2026-07-18 | Document drafted; no flip yet. | The jsonv2 friction (Contra #2) and legal pre-flight (Contra #3) must be resolved first.                                            |

---

## Revisit triggers

Re-open this decision earlier than planned if any of these happen:

- Wise announces an official Go SDK (changes the ecosystem calculus — Contra #5 becomes live).
- A specific external user asks to consume wise-go (changes the cost/benefit of staying private).
- `go-branded-id` or `go-error-family` adoption stalls and wise-go's visibility would unblock it (amplifies Pro #2).
