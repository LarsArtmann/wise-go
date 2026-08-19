# wise-go Session Status Report

**Snapshot:** 2026-08-08 05:15 CEST\
**Scope:** This report covers only the current session's investigation of Wise sandbox and integration testing, plus the TODO entry added as a result. It does not re-audit unrelated project work.

## Executive Summary

The repository has sandbox configuration and extensive HTTP-mock coverage, but no evidence of a real Wise sandbox integration test ever being implemented or run. The session correctly identified that distinction and recorded a high-priority TODO. The TODO is directionally correct, but it is not yet an execution-ready specification: it does not identify the exact endpoint, credential contract, command/build tag, assertions, CI secret names, or success evidence to record.

The working tree was clean after the TODO change was inspected, and the existing test suite passed with `GOEXPERIMENT=jsonv2 go test ./...`.

## A) FULLY DONE

1. **Repository inspection completed.** The session searched source, tests, documentation, CI configuration, and status artifacts for sandbox, live, integration, credential, and mock-test indicators.
2. **Sandbox option confirmed.** `WithSandbox()` exists in `options.go:25-29` and selects `SandboxURL` from `types.go:17-21`.
3. **Mock testing model confirmed.** `wise_test.go:116-139` constructs `httptest.Server`, uses a custom base URL, and supplies a fake API key. The tests exercise local handlers, not Wise infrastructure.
4. **No live integration-test files found.** No integration-test path or filename exists, and no `--live` flag or build-tagged live test was found.
5. **CI scope verified.** `.github/workflows/ci.yml:31-114` runs build, race tests, lint, vulnerability checking, and Nix checks, but contains no credentialed Wise sandbox job.
6. **Documentation corroborates the finding.** `CONTRIBUTING.md:129-140` explicitly states that tests use `httptest`, require no network access, and require no API key.
7. **Hermetic Nix sandbox was correctly distinguished from Wise sandbox.** Nix's sandboxed build/test environment is not a live Wise API environment.
8. **Existing tests passed.** `GOEXPERIMENT=jsonv2 go test ./...` passed for the root package and `internal/raw`.
9. **TODO recorded.** `TODO_LIST.md:9` now contains a P1 item for credentialed Wise sandbox integration tests, protected/manual execution, read-only endpoint coverage, and separation from mock tests.
10. **Diff hygiene checked.** `git diff --check` passed for the TODO change.
11. **Current timestamp captured.** CLI `date` returned `Sat Aug 8 05:15:41 AM CEST 2026`.

## B) PARTIALLY DONE

1. **Sandbox support is only configuration, not verification.** The SDK can select a sandbox URL, but the session did not verify that the configured URL is current or reachable using credentials.
2. **Integration-test requirement is documented but underspecified.** The TODO states the goal and broad safety boundary, but not enough implementation detail for another contributor to execute without making avoidable decisions.
3. **Test coverage is broad but synthetic.** Profiles, balances, transactions, authentication, parsing, error handling, retries, and query forwarding are covered through local fixtures. Wire compatibility against Wise's current sandbox responses remains unverified.
4. **Historical evidence was reviewed, not independently validated.** Older status documents consistently describe the live integration test as planned or absent. They are useful corroboration, but they are not proof of every historical execution environment.
5. **The current TODO priority is recorded as P1.** This matches the requested high priority, but the list has no explicit owner, acceptance checklist, credential prerequisites, or target milestone.
6. **The report was requested as Markdown and will be written as Markdown.** The status-report skill prefers HTML, but the user's explicit `.md` path requirement takes precedence.

## C) NOT STARTED

1. No live Wise sandbox test file.
2. No integration build tag such as `integration` or `live`.
3. No integration test command or script.
4. No environment-variable contract for sandbox credentials.
5. No protected or manually triggered integration workflow.
6. No GitHub Actions sandbox secrets.
7. No endpoint-by-endpoint sandbox acceptance assertions.
8. No test fixture strategy for stable Wise sandbox data.
9. No mechanism to prevent live tests from running accidentally on forks or ordinary pull requests.
10. No retry, timeout, or rate-limit policy specifically tuned for live sandbox verification.
11. No recording of a successful live run, timestamp, endpoint, API version, or covered operations.
12. No failure diagnostics that distinguish credential, endpoint, Wise availability, schema drift, and SDK defects.
13. No verification that the repository's `SandboxURL` matches Wise's current documented endpoint.
14. No simulation API support, which is outside the current read-only SDK scope but relevant to broader Wise sandbox testing.
15. No webhook or write-operation integration tests, consistent with those APIs not being implemented.

## D) TOTALLY FUCKED UP!

### Confirmed serious gaps

1. **We implicitly advertised sandbox support without proving interoperability.** `README.md:41` presents sandbox support as a feature, while no live sandbox test demonstrates that it works against Wise.
2. **The repository has a misleadingly strong testing impression.** The suite has high coverage and is described as integration-style in places, but its external boundary is fully mocked. This can catch client logic regressions while missing endpoint, authentication, schema, and environment incompatibilities.
3. **The earlier TODO context was too thin.** The added item was useful, but it omitted the operational details needed to execute it safely and reproducibly.

### Not found

- No evidence of destructive production activity.
- No evidence of leaked credentials.
- No evidence that the existing tests contacted Wise production or Wise sandbox.
- No test failure was introduced by the TODO documentation change.

## E) WHAT WE SHOULD IMPROVE

1. Define the integration contract before implementation: endpoint, API version, credential type, profile prerequisites, and supported calls.
2. Confirm the current Wise sandbox base URL before changing `SandboxURL`; do not rely on an unverified historical endpoint.
3. Add a build-tagged live test file so normal tests remain network-free.
4. Use explicit environment variables, for example `WISE_SANDBOX_API_KEY`, rather than hardcoded credentials.
5. Skip live tests with a clear message when credentials are absent.
6. Add a manually triggered GitHub Actions workflow, or a protected job that never runs for untrusted forks.
7. Never make live sandbox tests a required pull-request check unless credentials and sandbox stability are guaranteed.
8. Use short, bounded timeouts and avoid broad retry amplification against Wise sandbox.
9. Cover only stable, read-only operations initially: authentication/profiles, balances, and transactions.
10. Make assertions resilient to dynamic IDs, timestamps, balances, and transaction contents.
11. Assert response shape and semantic invariants rather than brittle exact sandbox data.
12. Capture endpoint and API version in test output without exposing secrets.
13. Classify failures into missing credentials, authentication failure, transport failure, schema decode failure, and assertion failure.
14. Add documentation describing how authorized maintainers run the test locally.
15. Record the first successful run as evidence in a dated status or release note.
16. Update `FEATURES.md` so sandbox support distinguishes selectable configuration from verified live interoperability.
17. Update `CONTRIBUTING.md` to distinguish mock tests from live integration tests.
18. Add acceptance criteria directly to `TODO_LIST.md`.
19. Review whether `SandboxURL` is stale before relying on it in automation.
20. Keep simulation API work separate from this first read-only connectivity check.

## F) UP TO 50 THINGS WE SHOULD GET DONE NEXT

|  # | Task                                                                     | Priority | Completion evidence                            |
| -: | ------------------------------------------------------------------------ | :------: | ---------------------------------------------- |
|  1 | Verify the current Wise sandbox endpoint from the approved documentation |    P1    | Endpoint decision recorded                     |
|  2 | Decide the supported credential type for this SDK's sandbox test         |    P1    | Credential contract documented                 |
|  3 | Define required sandbox environment variables                            |    P1    | Names and secret handling documented           |
|  4 | Define the first live test's supported read-only operations              |    P1    | Endpoint checklist approved                    |
|  5 | Add a build-tagged live integration test file                            |    P1    | `go test -tags=...` discovers it               |
|  6 | Add skip behavior when credentials are absent                            |    P1    | Ordinary test runs remain safe                 |
|  7 | Add bounded context timeout for live tests                               |    P1    | Timeout failure is deterministic               |
|  8 | Add live test assertions for profile retrieval                           |    P1    | Profile response validated                     |
|  9 | Add live test assertions for balance retrieval                           |    P1    | Balance response validated                     |
| 10 | Add live test assertions for transaction retrieval                       |    P1    | Transaction response validated                 |
| 11 | Make live assertions independent of fixed IDs                            |    P1    | Tests work with fresh accounts                 |
| 12 | Add diagnostic failure categories                                        |    P1    | Failures identify likely cause                 |
| 13 | Add a protected/manual GitHub Actions workflow                           |    P1    | Workflow cannot run on untrusted forks         |
| 14 | Configure sandbox credentials as repository/environment secrets          |    P1    | No secrets in source or logs                   |
| 15 | Keep live tests out of ordinary PR CI                                    |    P1    | Default CI remains credential-free             |
| 16 | Run the first authorized sandbox test locally                            |    P1    | Successful run log exists                      |
| 17 | Run the first authorized sandbox test in CI                              |    P1    | Successful workflow run exists                 |
| 18 | Record endpoint, date, and covered operations                            |    P1    | Evidence is reproducible                       |
| 19 | Reconcile `SandboxURL` with the verified endpoint                        |    P1    | SDK configuration is current                   |
| 20 | Update README sandbox documentation                                      |    P1    | User instructions match reality                |
| 21 | Update CONTRIBUTING testing documentation                                |    P1    | Mock/live distinction is explicit              |
| 22 | Update FEATURES sandbox status                                           |    P1    | Claim accurately reflects evidence             |
| 23 | Add an integration-test troubleshooting section                          |    P2    | Common failures documented                     |
| 24 | Add a CI concurrency policy for live tests                               |    P2    | Duplicate runs are limited                     |
| 25 | Add a live-test rate-limit budget                                        |    P2    | Tests stay within sandbox limits               |
| 26 | Add response redaction to any diagnostic logging                         |    P2    | Sensitive data is not emitted                  |
| 27 | Test expired or invalid credentials safely                               |    P2    | Authentication failure classification verified |
| 28 | Test sandbox-unavailable behavior without repeated retries               |    P2    | Failure is bounded and clear                   |
| 29 | Verify date and currency assumptions against sandbox responses           |    P2    | Parsing contract confirmed                     |
| 30 | Validate transaction date timezone handling with live data               |    P2    | UTC interpretation confirmed                   |
| 31 | Add schema-drift detection for observed response fields                  |    P2    | Unexpected shape fails clearly                 |
| 32 | Add a maintainer-only runbook                                            |    P2    | Authorized execution is repeatable             |
| 33 | Decide whether the test should use `WithSandbox()` or an explicit URL    |    P2    | Configuration path is tested                   |
| 34 | Test custom base URL override separately from live testing               |    P2    | Option precedence remains covered              |
| 35 | Pin any CI tooling introduced for live tests                             |    P2    | Workflow is reproducible                       |
| 36 | Set a maximum live-test duration                                         |    P2    | Workflow cannot hang indefinitely              |
| 37 | Set a retention policy for live-test logs                                |    P2    | Logs avoid unnecessary sensitive retention     |
| 38 | Add a scheduled sandbox verification, if Wise permits stable cadence     |    P2    | Drift is detected over time                    |
| 39 | Add manual approval before scheduled live runs                           |    P2    | External calls are deliberate                  |
| 40 | Define ownership for responding to sandbox failures                      |    P2    | Maintainer responsibility documented           |
| 41 | Add a changelog entry after the first verified run                       |    P2    | User-facing evidence recorded                  |
| 42 | Reassess the v1.0 gate against live verification                         |    P1    | Release checklist includes interoperability    |
| 43 | Keep write-operation tests blocked until write APIs exist                |    P2    | Scope remains honest                           |
| 44 | Keep webhook tests blocked until webhook support exists                  |    P2    | Scope remains honest                           |
| 45 | Evaluate Wise Simulation API separately                                  |    P3    | Separate proposal exists                       |
| 46 | Define whether mTLS is required for future partner testing               |    P3    | Authentication roadmap recorded                |
| 47 | Add compatibility notes for Wise API versioning                          |    P3    | Version behavior documented                    |
| 48 | Review sandbox endpoint changes on Wise releases                         |    P3    | Maintenance trigger exists                     |
| 49 | Add a release checklist item requiring live evidence                     |    P1    | Release process prevents unsupported claims    |
| 50 | Audit all documentation for “sandbox” wording after implementation       |    P2    | No mock/live ambiguity remains                 |

## G) QUESTIONS THAT CANNOT BE FIGURED OUT FROM THE REPOSITORY

1. Which Wise sandbox credential should authorized integration tests use: a personal API token, partner client credentials, or another Wise-provided credential flow?
2. Which Wise sandbox profile, balance, currency, and date-range fixtures are guaranteed to exist and remain stable for this project?
3. Should live sandbox verification be manually triggered only, or should maintainers also schedule it periodically?

## Session-Level Assessment

The session was successful at answering the original question and preventing a false claim that live integration tests exist. It was incomplete as implementation planning: the TODO was added before the operational contract was fully specified. No code or CI behavior was changed, and no external Wise request was made.

The next correct action is to resolve the three external prerequisites above, then implement the smallest safe read-only live test and protected execution path.
