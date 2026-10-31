# Architecture Understanding — Series Index

Point-in-time architecture reports for wise-go, newest first. Reports are
frozen once written (docs-health ANNOTATE convention); the newest report on
a topic supersedes older ones on current-state claims only.

| Date       | Report                                                                                                                 | Verdict one-liner                                                                                                                                                                                                                                      |
| ---------- | ---------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 2026-10-31 | [samber/do + health recheck](2026-10-31_09-33_samber-do-health-architecture-recheck.md)                                | Zero new findings; samber/do still deliberately absent (correct for an SDK); scores 4.0 / 4.5 / 4.5 / 3.5 (self-healthy up from 3.0). Supersedes the 2026-10-16 review on current-state claims; F-1 code-half executed same day (`VerifyCredentials`). |
| 2026-10-16 | [samber/do + health review](2026-10-16_19-59_samber-do-health-architecture-review.md)                                  | First pass: DO-1..DO-6 counterfactual + ADR 002/003 policy say no DI container in the SDK; health-probe honesty gap found (F-1) → docs-half shipped same day.                                                                                          |
| 2026-07-18 | [modularity & composability](2026-07-18_10-08_modularity-and-composability.html)                                       | HTML report (generated artifact — leave-alone). Early structural assessment.                                                                                                                                                                           |
| 2026-07-18 | [current architecture](2026-07-18_10-15_current-architecture.d2) (+ [.svg](2026-07-18_10-15_current-architecture.svg)) | D2/SVG pair (generated — leave-alone). System-as-it-is diagram.                                                                                                                                                                                        |
| 2026-07-18 | [target architecture](2026-07-18_10-15_target-architecture.d2) (+ [.svg](2026-07-18_10-15_target-architecture.svg))    | D2/SVG pair (generated — leave-alone). System-as-it-should-be diagram.                                                                                                                                                                                 |

Recheck cadence: the samber/do + health recheck joins the ~8-week seasonal
sweep (next due ~2026-12-16; see the "Seasonal conventions" gotcha in
AGENTS.md) and re-triggers on samber/do lifecycle releases (>v2.1.0).

Related skills: `architecture-review` (scored assessment — produces the .md
reports above), `architecture-visualization` (D2/HTML artifacts).
