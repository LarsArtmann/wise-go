# Transfer Category Re-Audit vs Live Preview

**Date:** 2026-10-16 · **Source:** `docs.wise.com/api-reference/transfer` (live preview,
server version `2026Q4`) · **Baseline:** `docs/reviews/2026-08-08_api-docs-study.md`
(8/11 items covered) · **Feeds:** Pareto plan M5/M8/M9/M15.

## Category deltas vs the 2026-08-08 study

| # | Operation | Path | Status in SDK | Note |
|---|-----------|------|---------------|------|
| 1 | List transfers | `GET /transfers` | SHIPPED (`ListTransfers`, auto-paginates 100/page) | |
| 2 | Create transfer | `POST /transfers` | SHIPPED (`CreateTransfer`) | |
| 3 | Get transfer | `GET /transfers/{transferId}` | SHIPPED (`GetTransfer`) | no SCA needed (verified live 2026-10-16) |
| 4 | Cancel transfer | `PUT /transfers/{transferId}/cancel` | SHIPPED (`CancelTransfer`) | |
| 5 | Transfer requirements | `POST /transfer-requirements` | SHIPPED (`ValidateTransferRequirements` + refresh pair) | |
| 6 | Fund transfer | `POST /profiles/{profileId}/transfers/{transferId}/payments` | SHIPPED (`FundTransfer`) | |
| 7 | List transfer payments | `GET /transfers/{transferId}/payments` | **M5 — this plan** | response = array of `{id, method, pricingVariant, amount, currency, timeCreated, timeUpdated}`; "usually a single payment per transfer" |
| 8 | Receipt PDF | `GET /transfers/{transferId}/receipt.pdf` | SHIPPED (`GetTransferReceipt`) | 404 until `outgoing_payment_sent` |
| 9 | US combined receipt | `GET /v3/transfers/{transferId}/us-combined-receipt.pdf` | **M8 — this plan** | note the `/v3/` version prefix on an otherwise-unversioned surface; available from `incoming_payment_initiated` |
| 10 | NOC document | `GET /transfers/{transferId}/documents/noc` | **M9 — this plan** | INR transfers in `outgoing_payment_sent`; used to obtain a FIRC |
| 11 | Payout info | `GET /transfers/{transferId}/invoices...` family | SHIPPED (`GetTransferPayoutInfo`) | banking-partner reference tracking |
| 12 | **Banking partner invoice** | `GET /transfers/{transferId}/invoices/bankingpartner` | **UNTRACKED GAP — new finding** | not in the 2026-08-08 study's 11 items; add to TODO_LIST |
| 13 | **Deposit details (bank transfer pay-in)** | `GET /profiles/{profileId}/transfers/{transferId}/deposit-details/bank-transfer` | **UNTRACKED GAP — new finding** | pay-in instructions for funding; add to TODO_LIST |
| 14 | Third-party (originator) create | `POST /profiles/{profileId}/third-party-transfers` | **M15 — this plan** | originator = a SEPARATE endpoint family, not a `POST /transfers` variant; idempotency via `originalTransferId` |
| 15 | Third-party (originator) get | `GET /profiles/{profileId}/third-party-transfers/{transferId}` | **M15 — this plan** | |

## Corrections to earlier planning assumptions

- The 2026-10-16 session mapped "8/11 items covered" against a category inventory that
  did not include items 12-15. The true category is ~15 operations; 9 shipped,
  4 in this plan (M5/M8/M9/M15), 2 newly tracked gaps (items 12-13).
- Originator transfers are NOT a schema variant of `POST /transfers` (as the plan's
  M15 originally assumed) but dedicated `third-party-transfers` endpoints with an
  `Originator Transfer` schema (datablock `originator`, `originalTransferId`
  idempotency). M15 implements the create + get pair against that contract.
- The wire transfer schema keeps flat `sourceValue`/`targetValue` numbers (plus
  `sourceCurrency`/`targetCurrency`); the SDK's public `Transfer.Source`/`Target`
  `Money` pair is the parsed form. This settles the M22 history question at the
  contract level: the flat wire names always existed on the wire, never as public
  Go fields (full `git log -S` dig still pending, see plan M22).

## Conformance note for M5/M8/M9

The vendored spec (`docs/reviews/wise-api-openapi.json`, 173 paths) predates the
`2027Q1` server version; M6 (executed before M5 in this plan run) re-vendors the
fresh spec so the payments/receipt/NOC exchange templates validate against
authoritative schemas rather than the stale snapshot.
