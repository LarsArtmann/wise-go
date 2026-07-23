# Twelve-Factor Compliance Review: wise-go

**Assessment date:** 2026-07-23  
**Repository:** `github.com/larsartmann/wise-go`  
**Reference:** [The Twelve-Factor App](https://12factor.net/)  
**Scope:** Repository structure, Go library implementation, development tooling, CI configuration, and documented operating model.

## Executive summary

`wise-go` is a Go SDK library, not a deployable web application or service. The Twelve-Factor methodology is explicitly aimed at software delivered as a service, so several factors are either **not applicable** or can only be assessed at the level of the library's integration contract. Applying the methodology mechanically would produce misleading findings, particularly for port binding, process formation, logs, and admin processes.

Within the factors that fit a reusable SDK, the project is strong: it is version-controlled, declares Go dependencies in `go.mod`, supports injected backing-service clients and URLs, keeps runtime state in memory only, and provides reproducible Nix-based development and checks. The main weaknesses are operational rather than algorithmic: there is no first-class release pipeline or immutable release model visible in the repository, the CI workflow does not currently execute the full Nix check, and configuration is partly represented by constructor arguments and compiled defaults rather than by an application-level environment contract.

### Overall result

| Dimension                         |                                                           Result | Interpretation                                                                                  |
| --------------------------------- | ---------------------------------------------------------------: | ----------------------------------------------------------------------------------------------- |
| Applicable factors                |                                          **5 strong, 2 partial** | Good library-level compliance for code, dependencies, resources, statelessness, and dev parity  |
| Not applicable to this repository |                                                    **5 factors** | This repo does not own a server process, deployment runtime, logging pipeline, or admin process |
| Recommended priority              | **P1: release/CI evidence; P2: consumer configuration guidance** | Improve the operational contract without turning the SDK into an application framework          |

**Bottom line:** The project is **well aligned with the spirit of Twelve-Factor where the methodology applies**, but it cannot be called a Twelve-Factor _app_ because it does not itself run an application process. It is best described as a Twelve-Factor-friendly client library.

## Assessment model

Each factor is classified using the following vocabulary:

- **Strong** — the repository or library design directly satisfies the relevant principle.
- **Partial** — the principle is supported, but evidence, automation, or a boundary is incomplete.
- **Not applicable** — responsibility belongs to an embedding application or deployment platform, not this library.
- **Inapplicable by design** — the factor assumes a service shape that would be inappropriate for this project.

The evidence below is based on the repository at the assessment date. A library review cannot infer how every downstream consumer deploys it.

## Factor-by-factor findings

### I. Codebase — Strong

**Principle:** One codebase tracked in revision control, many deploys.

**Evidence:** The project is a single Git repository and Go module, `github.com/larsartmann/wise-go`, with one package at the repository root (`go.mod:1-3`, `CONTRIBUTING.md:100-111`). The codebase is intentionally a library consumed by applications rather than a collection of independently deployed services.

**What is good:**

- Source, tests, dependency manifests, CI, and reproducible development configuration live together.
- The repository has a clear module identity and no duplicated application implementations.
- Sandbox and production behavior are selected by client configuration rather than separate codebases (`options.go:25-37`).

**Limitations:** The “many deploys” concept belongs to consuming applications. `wise-go` itself is published as a module version, not deployed as a service.

**Recommendation:** No code change. Keep one canonical module and publish versioned releases/tags consistently.

### II. Dependencies — Strong, with a reproducibility caveat

**Principle:** Explicitly declare and isolate dependencies.

**Evidence:** Runtime and test dependencies are declared in `go.mod`, with checksums in `go.sum` (`go.mod:5-29`). Nix provides pinned inputs through `flake.lock` and supplies Go, `gopls`, `golangci-lint`, and Go tooling (`flake.nix:14-26`, `flake.nix:50-72`). The required `GOEXPERIMENT=jsonv2` is documented and wired into Nix and BuildFlow (`flake.nix:59-61`, `.buildflow.yml:4-7`, `CONTRIBUTING.md:72-96`).

**What is good:**

- Go modules provide explicit dependency declaration.
- Nix supplies a reproducible toolchain and locked inputs.
- CI and local development document the same non-default Go experiment.
- The runtime dependency set is deliberately small: failsafe-go, go-branded-id, and go-error-family.

**Gap:** The ordinary manual workflow still relies on a developer-installed Go toolchain and separately installed linter. `go.mod` is explicit, but execution isolation is strongest only inside `nix develop` or the Nix checks.

**Recommendations:**

1. Make the Nix shell and `nix flake check` the canonical supported path in CI and contributor documentation.
2. Add a simple, documented dependency/toolchain verification command for non-Nix users, including the expected Go version and `GOEXPERIMENT`.
3. Ensure release builds use the same locked dependency/toolchain path as CI.

### III. Config — Partial, appropriate for a library

**Principle:** Store deploy-varying configuration in the environment.

**Evidence:** The client receives API credentials and deployment-specific values through constructor arguments and options (`client.go:36-87`, `options.go:25-63`). Production and sandbox URLs are selectable, custom URLs are supported, and timeout/retry behavior is configurable. The README explicitly demonstrates passing API keys at runtime (`README.md:52-60`, `README.md:105-128`).

**What is good:**

- Secrets are not hardcoded in the repository.
- The API key is supplied by the caller, not stored globally or read from a checked-in file (`client.go:36-40`, `client.go:177-180`).
- Per-deploy endpoint selection is supported with `WithBaseURL` and `WithSandbox`.
- Timeout and retry policy are injectable per client instance.

**Why this is only partial:** Twelve-Factor defines application config as environment variables. A library should not unilaterally read process environment variables because that hides configuration from callers and makes testing less explicit. In this project, the embedding application is responsible for translating environment variables such as `WISE_API_KEY` into `wise.New(...)` arguments.

**Recommendations:**

1. Keep the library API explicit; do not add implicit environment reads to `wise.New`.
2. Add consumer documentation showing a safe pattern such as reading `WISE_API_KEY` in the application boundary and passing it to the SDK.
3. Document that `WithBaseURL` is intended for deployment configuration and that custom URLs must be validated by the consuming application.
4. Consider validating an empty API key or documenting the intentional unauthenticated-request behavior more prominently. This is an API usability decision, not a direct Twelve-Factor requirement.

### IV. Backing services — Strong

**Principle:** Treat backing services as attached resources, replaceable through configuration.

**Evidence:** Wise is accessed over HTTP through a configurable base URL and an injectable HTTP client (`options.go:32-37`, `options.go:56-63`, `client.go:69-87`). The client uses the configured URL to issue requests (`client.go:131-155`) and supports sandbox, production, proxy, and test-server endpoints.

**What is good:**

- The external Wise API is treated as a network resource rather than embedded infrastructure.
- Tests replace the service with `net/http/httptest`, as documented in `CONTRIBUTING.md:129-140`.
- Transport behavior can be replaced for tracing, middleware, mTLS, or tests.
- Retry behavior is centralized and classifies rate-limit, server, and network failures (`client.go:76-80`, `client.go:90-98`).

**Caveat:** The default production URL is compiled into the library (`client.go:45-47`, `types.go`), but this is a sensible default, not an environment-specific secret or deployment binding. Consumers can override it.

**Recommendation:** No architectural change. Preserve the replaceable-resource boundary and add an explicit example for proxy/custom endpoint deployment configuration.

### V. Build, release, run — Partial

**Principle:** Strictly separate build, release, and run stages, with immutable releases.

**Evidence:** The repository has a declarative Nix development environment and checks, including formatting and a sandboxed Go test derivation (`flake.nix:74-123`). CI runs build, tests, lint, and vulnerability checks according to the feature inventory (`FEATURES.md:88-98`) and workflow inventory. The project is a library, so there is no executable runtime bundle owned by this repository.

**What is good:**

- Build/test inputs are declaratively described in `flake.nix`.
- The Nix check uses a source fileset, a pinned vendor hash, and a fixed Go package version (`flake.nix:45-47`, `flake.nix:88-114`).
- Tests run without a live Wise account or network access.
- There is a natural separation between source checkout, module build, and consumer runtime.

**Gap:** The repository does not visibly define a release stage for publishing an immutable module release, nor does CI currently prove the full `nix flake check` path. `FEATURES.md:97` explicitly records the full check as partially functional, and `TODO_LIST.md:27-31` tracks verification and CI integration.

**Recommendations:**

1. Add a CI job that runs `nix flake check` end to end, not only the lighter checks.
2. Define a release workflow that builds/tests a tagged commit and publishes the Go module from that immutable tag.
3. Record release identifiers and migration notes in `CHANGELOG.md`; never mutate a published release.
4. Keep runtime configuration in the consuming application, separate from the library artifact.

### VI. Processes — Strong for the library boundary; application-owned for runtime

**Principle:** Execute as stateless, share-nothing processes.

**Evidence:** `Client` stores only configuration, an HTTP client, and a retry executor (`client.go:28-34`). API calls construct requests from context and configuration, and response bodies are closed after use (`client.go:131-174`). The roadmap explicitly states that the SDK is stateless and that caching is the caller’s responsibility (`ROADMAP.md:136-144`).

**What is good:**

- No database, filesystem state, process-global cache, session store, or durable local state is introduced.
- Per-request context is used to control cancellation and deadlines (`client.go:145-149`).
- The client can be instantiated independently and safely injected into callers.
- The design does not require sticky sessions or process affinity.

**Boundary:** A Go SDK does not create or supervise the consuming application’s processes. The caller must ensure its own web handlers/workers remain stateless.

**Recommendation:** No code change. Preserve the stateless client contract and document concurrency expectations if the public API guarantees concurrent use.

### VII. Port binding — Not applicable by design

**Principle:** Export services by binding to a port rather than depending on runtime web-server injection.

`wise-go` is a client library and intentionally does not listen on a port or export an HTTP server. It consumes the Wise API through `net/http`; it does not provide a service to route traffic to. Adding a server solely to satisfy this factor would violate the project’s stated scope.

**Consumer responsibility:** Any application embedding `wise-go` should independently expose its own HTTP/gRPC service using its chosen process and port-binding model.

### VIII. Concurrency — Not applicable as an application process model; library-friendly

**Principle:** Scale workload by running more process instances and separating process types.

The repository defines no long-running process types, workers, or web server. It does provide a reusable client that performs request-scoped work and delegates concurrency to the embedding application and Go runtime. Retry execution is bounded by configured retry and backoff values (`client.go:14-19`, `options.go:46-53`).

**Consumer responsibility:** The consuming service should scale horizontally, avoid shared mutable process state, and configure Wise API rate limits appropriately. The SDK’s automatic retries can amplify load if consumers also retry without coordination; callers should treat the returned typed error and retry policy as the source of truth.

**Recommendation:** Document retry composition guidance for consumers, especially avoiding unbounded nested retries.

### IX. Disposability — Strong at request level; not applicable at process level

**Principle:** Fast startup, graceful shutdown, and robustness against sudden process death.

**Evidence:** Client construction is lightweight and performs no network call (`client.go:36-87`); authentication is explicit via `Authenticate` (`client.go:100-107`). Each request accepts a context and closes its response body (`client.go:145-174`). No background goroutine, daemon, PID file, or local durable state is visible in the client.

**What is good:**

- A client can be created quickly and discarded without cleanup APIs.
- In-flight requests can be canceled through context.
- The library does not own process shutdown or signal handling.

**Limitation:** Graceful SIGTERM handling belongs to the embedding application. The SDK cannot stop accepting requests or drain a server it does not own.

**Recommendation:** No library lifecycle framework. Add integration guidance showing callers should propagate request contexts and use an application-level `http.Server.Shutdown` path.

### X. Dev/prod parity — Strong for build and service shape, partial for external API access

**Principle:** Minimize time, personnel, and tooling gaps between development and production.

**Evidence:** Nix provides reproducible Go 1.26, Go tools, linting, and `GOEXPERIMENT=jsonv2` (`flake.nix:45-72`). CI documents the same experiment and test commands (`CONTRIBUTING.md:72-96`, `.buildflow.yml:4-7`). Tests use `httptest` and require no API key or live network (`CONTRIBUTING.md:129-140`). Sandbox support exists (`options.go:25-29`, `README.md:105-110`).

**Strengths:**

- Development and CI use the same declared Go experiment.
- The test suite exercises HTTP behavior through a local HTTP server rather than ad hoc mocks.
- Wise sandbox support gives consumers a closer pre-production endpoint.
- Formatting, linting, race tests, and coverage are documented as quality gates.

**Gaps:**

- The repository does not demonstrate a staging deployment or a live sandbox integration job.
- The full Nix check is not yet confirmed in CI (`FEATURES.md:97`).
- Manual development outside Nix can drift in Go/linter versions.

**Recommendations:**

1. Run the complete Nix check in CI.
2. Add an optional, credentialed sandbox integration workflow, never as a required fork-safe PR check.
3. Pin or verify tool versions for the manual path.
4. Keep test fixtures and sandbox behavior aligned with the documented wire formats.

### XI. Logs — Not applicable to the library; good non-interference

**Principle:** Treat logs as event streams written to stdout/stderr and managed by the environment.

The SDK does not create a logging subsystem, write log files, or route output. This is appropriate for a library: logging policy belongs to the consuming application, which can decide whether and how to log request failures. The client returns typed, contextual errors instead of silently emitting output (`client.go:157-170`, `errors.go`).

**Risk for consumers:** Applications should avoid logging API keys, bearer tokens, sensitive account data, or full response payloads. If request/response logging middleware is injected through `WithHTTPClient`, it must redact authorization and financial data.

**Recommendation:** Add a security-focused README note for custom HTTP middleware and redaction. Do not add default logging to the library.

### XII. Admin processes — Not applicable

**Principle:** Run administrative tasks as one-off processes using the same release and configuration as regular processes.

`wise-go` contains no executable admin commands, migrations, consoles, or one-off operational scripts. It is a library consumed by other applications. Any administrative operation, such as reconciliation or account verification, is performed by the consumer using the same module version and configuration boundary.

**Recommendation:** No change. If command-line tooling is added in the future, make it a separate executable with an explicit release/configuration contract rather than hiding admin behavior in package initialization.

## Cross-cutting observations

### Strong design decisions

- **Explicit external-resource boundary:** `WithBaseURL` and `WithHTTPClient` make the Wise API replaceable in tests and deployments.
- **Stateless request execution:** No local persistence or hidden process state is required.
- **Context propagation:** Callers control cancellation and deadlines.
- **Reproducible tooling:** Nix, Go modules, `go.sum`, and `GOEXPERIMENT=jsonv2` reduce environment drift.
- **Typed failure semantics:** Error families and retryability are more operationally useful than string matching.
- **No library-owned logs:** Consumers retain control of event streams and sensitive-data policy.

### Important non-findings

These are not compliance failures:

- No port binding: the project is not a service.
- No process formation: the project is not a deployable process.
- No admin command: the project is not an application runtime.
- No environment-variable reader inside the SDK: explicit constructor configuration is safer and more composable for a library.
- No database/cache: the SDK is intentionally a stateless API client.

## Prioritized action plan

### P1: Make release and CI evidence complete

1. Add a GitHub Actions job for full `nix flake check`.
2. Verify that the Nix test derivation, formatting check, and all expected lint checks pass in a clean CI environment.
3. Add a tagged-release workflow that validates the tag, runs the canonical checks, and publishes the module from that immutable revision.
4. Update `FEATURES.md` and `TODO_LIST.md` when the checks are actually proven, not merely configured.

**Why first:** This closes the largest observable gap between a reproducible development setup and a reproducible published artifact.

### P2: Clarify the library/application configuration boundary

1. Add README guidance for mapping application environment variables to `wise.New`.
2. Document the expected handling of `WISE_API_KEY`, sandbox selection, custom endpoint, timeout, and retry values at the application boundary.
3. Add redaction guidance for injected HTTP transports and middleware.
4. Document retry composition so consumer-level retries do not multiply SDK-level retries unexpectedly.

**Why second:** These changes improve real deployments without introducing hidden global configuration into the library.

### P3: Add optional sandbox verification

1. Create a manually triggered or protected sandbox integration workflow.
2. Keep live credentials out of pull requests and fork workflows.
3. Use the same module/toolchain setup as production checks.
4. Record which endpoints are covered and distinguish sandbox failures from unit-test failures.

**Why third:** It validates the external backing-service contract while preserving hermetic, credential-free PR tests.

### P4: Preserve application-owned operational concerns

Do not add a server, logger, daemon, process manager, database, cache, or environment reader to this SDK merely to increase a Twelve-Factor score. Those concerns belong to the consuming application and deployment platform.

## Verification performed

The review examined:

- `go.mod` and `go.sum` for dependency declaration.
- `flake.nix`, `flake.lock`, `.buildflow.yml`, and contributor instructions for build isolation and environment parity.
- `client.go` and `options.go` for configuration, resource attachment, retries, context propagation, response cleanup, and state management.
- `README.md`, `FEATURES.md`, `TODO_LIST.md`, `ROADMAP.md`, and `docs/DOMAIN_LANGUAGE.md` for the stated operating model and known gaps.
- The Twelve-Factor reference pages at [12factor.net](https://12factor.net/), including all twelve factor definitions.

This is a static repository review. It does not prove deployment behavior, production signal handling, release publication, CI execution in an external runner, or downstream consumer configuration.

## Conclusion

`wise-go` complies well with the parts of Twelve-Factor that make sense for a reusable SDK. Its strongest properties are explicit dependencies, replaceable HTTP resources, stateless request execution, context-driven cancellation, and reproducible development tooling. Its weaker area is not the library architecture but the surrounding delivery evidence: full Nix validation and immutable release automation are documented or planned rather than demonstrated here.

The correct target is not to make `wise-go` behave like a Twelve-Factor web app. The correct target is to keep the SDK **stateless, explicit, dependency-isolated, resource-configurable, and operationally quiet**, while making every consuming application responsible for its own environment configuration, process model, ports, logs, deployment stages, and administrative jobs.
