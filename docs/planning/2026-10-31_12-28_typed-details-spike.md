# Typed recipient `Details` — design spike (2026-10-31 12:28 CEST)

Status: SPIKE COMPLETE, no code change. Answers TODO_LIST's
"Typed recipient `Details`" decision row (carried unanswered through seven
status reports) with a census, three options, and a recommendation.

## 1. What the domain actually is

`Recipient.Details` / `CreateRecipientRequest.Details`
(`types.go:500`/`types.go:510`) are `map[string]string` because Wise's
recipient fields are **route-polymorphic AND server-defined**: the required
keys depend on the (targetCurrency, source, route) triple, and Wise can
change them without an SDK-visible contract change. The
account-requirements endpoints (`TransferRequirement`/`TransferRequirementForm`,
`types.go:565`) exist precisely because the field tree is discovered at
runtime — `wise_test.go:2065` pins the shape: a recursive
`Field{Key, Name, Type, Group[]}` tree, not a static schema.

## 2. Census (what is derivable today)

| Source                                          | Finding                                                                                                                                  |
| ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| Vendored spec (`account-requirements-response`) | Dynamic field tree; the spec does NOT enumerate per-corridor keys — it cannot, by design.                                                |
| SDK tests/fixtures                              | Exactly one corridor exercised: GBP (`sortCode`, `accountNumber`) — `wise_test.go:2278` and 9 more sites. No other corridor keys appear. |
| SDK public surface                              | `TransferRequirementsDetails` (`types.go:551`) mirrors the accepted free-form block with named fields for the CREATE-side details.       |
| Consumers                                       | bank-sync (only known consumer) sends GBP corridors via the map; nobody depends on typed structs (v1.0 audit, 2026-08-21).               |

A full per-corridor census (40+ field sets across EUR/GBP/USD/AUD/...) would
require hundreds of credentialed account-requirements calls — that is the
data-gathering step of IMPLEMENTATION, not of this decision; it only becomes
worth doing when a demand signal exists.

## 3. Options

- **(a) Typed per-corridor structs** — `type GBPDetails struct{SortCode, AccountNumber string}` + a sum-type dispatch on route. Compile-time safety; but the guarantee is a LIE against a server-owned field set: when Wise adds/renames a required key (they do, without notice), the struct silently stops representing reality and the SDK — an anti-corruption layer — becomes the corruption. Also explodes the surface: ~40 structs, each a compatibility promise we cannot test offline (mock fixtures per corridor).
- **(b) Key-constants package** — `package detailskeys; const SortCode = "sortCode"` … String keys stay, typos become rarer. Cheap, additive, no false guarantees. Does nothing for required-vs-optional-per-corridor (the actual pain).
- **(c) Keep the map; type AROUND it (demand-gated)** — freeze the map at v1.0 (the audit's finding: it is the only shape consumers depend on), and if a demand signal arrives, add a helper `ValidateDetailsAgainstRequirements(details map[string]string, reqs []TransferRequirement) error` that checks the caller's map against the runtime requirements tree the SDK already decodes. Types the DECISION (missing/extra fields) instead of the VOCABULARY (field names) — the decision is stable, the vocabulary is not.

## 4. Recommendation

**(c)**, with (b) as a cheap garnish only if a second consumer appears.
Rationale: the field vocabulary is server-owned and route-polymorphic;
typing it in the SDK puts the compile-time promise where the SDK has no
authority. The validation helper composes with the existing
`TransferRequirement` decoding and is per-consumer-testable offline.
This resolves the carried decision: v1.0 freezes the map; typed accessors
are NOT planned; the validation helper sits in the parked demand-gated set
(M27).

## 5. Verification appendix

| Claim                                   | Command/evidence                                                          |
| --------------------------------------- | ------------------------------------------------------------------------- |
| Details is `map[string]string` in both  | `rg -n 'Details\s+map\[string\]string' types.go` → `:500`, `:510`         |
| Requirements tree is runtime-discovered | `types.go:565` + spec schema `account-requirements-response`              |
| Only GBP corridor in fixtures           | `rg '"sortCode"' wise_test.go internal_test.go` (11 sites, no other keys) |
| No consumer depends on typed structs    | v1.0 audit (`docs/reviews/archived/2026-08-21_v1.0-api-audit.md`)         |
