package wise_test

// Spec conformance harness: the mock harnesses in wise_test.go and
// ott_test.go encode what the SDK BELIEVES the Wise API looks like. This
// harness records every HTTP exchange those harnesses serve and validates it
// against the machine-readable Wise OpenAPI spec vendored at
// docs/reviews/wise-api-openapi.json (the authoritative contract, see
// AGENTS.md). A mismatch means either the client/fixtures drifted from the
// documented API or the vendored spec snapshot is stale and needs a refresh.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

// specSnapshotPath is the vendored Wise OpenAPI 3.1 spec, relative to the
// package directory (the working directory when tests run).
const specSnapshotPath = "docs/reviews/wise-api-openapi.json"

// conformanceExchange is one recorded request/response pair served by a mock
// harness.
type conformanceExchange struct {
	method             string
	rawPath            string
	query              string
	reqHeader          http.Header
	reqBody            []byte
	status             int
	respBody           []byte
	respCT             string
	skipResponseSchema bool
}

// conformanceHarness wraps a mock handler, records every exchange, and
// validates the recorded set against the vendored spec on finish.
type conformanceHarness struct {
	inner     http.Handler
	mu        sync.Mutex
	exchanges []conformanceExchange
}

// attachConformance wraps inner so every served request is recorded for spec
// conformance. Use the returned value as the httptest.Server handler.
func attachConformance(inner http.Handler) *conformanceHarness {
	return &conformanceHarness{inner: inner}
}

// ServeHTTP records the exchange, forwards it to the wrapped handler, and
// relays the captured response to the real test client.
func (h *conformanceHarness) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var reqBody []byte

	if r.Body != nil {
		reqBody, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
	}

	r.Body = io.NopCloser(bytes.NewReader(reqBody))

	captured := &captureResponseWriter{header: make(http.Header)}
	h.inner.ServeHTTP(captured, r)

	for key, values := range captured.header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(captured.status)
	_, _ = w.Write(captured.body.Bytes())

	h.mu.Lock()
	defer h.mu.Unlock()

	h.exchanges = append(h.exchanges, conformanceExchange{
		method:             r.Method,
		rawPath:            r.URL.Path,
		query:              r.URL.RawQuery,
		reqHeader:          r.Header.Clone(),
		reqBody:            reqBody,
		status:             captured.status,
		respBody:           captured.body.Bytes(),
		respCT:             captured.header.Get("Content-Type"),
		skipResponseSchema: r.Header.Get(skipResponseSchemaHeader) == "1",
	})
}

// finish validates all exchanges recorded since the last finish and clears
// the buffer. fail receives one message per violation so the surrounding
// test framework attributes failures to the scenario that served them.
func (h *conformanceHarness) finish(fail func(msg string)) {
	h.mu.Lock()
	exchanges := h.exchanges
	h.exchanges = nil
	h.mu.Unlock()

	if len(exchanges) == 0 {
		return
	}

	specCtx, err := loadConformanceSpec()
	if err != nil {
		fail(fmt.Sprintf("load spec snapshot %s: %v", specSnapshotPath, err))

		return
	}

	for _, exchange := range exchanges {
		for _, problem := range validateExchange(specCtx, exchange) {
			fail(fmt.Sprintf(
				"does not conform to %s: %s %s: %s",
				specSnapshotPath, exchange.method, exchange.rawPath, problem,
			))
		}
	}
}

// conformanceSpecContext holds the parsed spec and its router. The server
// URLs are rewritten to a relative path so routing matches any mock host.
type conformanceSpecContext struct {
	doc    *openapi3.T
	router routers.Router
}

var (
	conformanceOnce   sync.Once
	conformanceLoaded *conformanceSpecContext
	conformanceErr    error
)

func loadConformanceSpec() (*conformanceSpecContext, error) {
	conformanceOnce.Do(func() {
		loader := openapi3.NewLoader()

		doc, loadErr := loader.LoadFromFile(specSnapshotPath)
		if loadErr != nil {
			conformanceErr = loadErr

			return
		}

		// Relative server URL: the router then matches any host (the mocks
		// serve from 127.0.0.1), per the kin-openapi routers docs.
		doc.Servers = []*openapi3.Server{{URL: "/"}}

		// Wise ships the snapshot with an empty info.version, which the
		// OpenAPI validator rejects; the quarterly surface is recorded in
		// the server URLs instead, so patch in a descriptive label.
		if doc.Info != nil && doc.Info.Version == "" {
			doc.Info.Version = "wise-snapshot"
		}

		router, routeErr := legacyrouter.NewRouter(doc, openapi3.DisableExamplesValidation())
		if routeErr != nil {
			conformanceErr = routeErr

			return
		}

		conformanceLoaded = &conformanceSpecContext{doc: doc, router: router}
	})

	return conformanceLoaded, conformanceErr
}

// permissiveDateTimeFormat accepts any date-time-shaped value. Wise's live
// responses deliberately use looser timestamp shapes than RFC3339
// (space-separated statement dates, millisecond+numeric-zone delivery
// estimates, zoneless createdAt — see AGENTS.md), which the spec idealizes
// as date-time; the fixtures mirror live output, so the format assertion is
// relaxed while required, type, enum, and shape checks stay enabled.
type permissiveDateTimeFormat struct{}

func (permissiveDateTimeFormat) Validate(string) error { return nil }

var conformanceValidationOptions = []openapi3.SchemaValidationOption{
	openapi3.WithStringFormatValidator("date-time", permissiveDateTimeFormat{}),
}

// conformanceOptions builds the shared request/response validation options.
// Multi errors report every violated schema constraint at once.
func conformanceOptions() *openapi3filter.Options {
	return &openapi3filter.Options{
		MultiError:              true,
		AuthenticationFunc:      openapi3filter.NoopAuthenticationFunc,
		SchemaValidationOptions: conformanceValidationOptions,
	}
}

// skipResponseSchemaHeader marks a fixture response as intentionally
// non-conformant (corruption and leniency tests feed malformed bodies on
// purpose); only response-body validation is skipped for those exchanges.
const skipResponseSchemaHeader = "X-Conformance-Skip-Response-Schema"

// exemptResponseSchema wraps a fixture handler whose response body
// intentionally violates the spec so the harness skips response-body
// validation for those exchanges. Path, method, query, and request
// validation still apply.
func exemptResponseSchema(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set(skipResponseSchemaHeader, "1")
		next(w, r)
	}
}

// versionedSegment matches the leading version segments the SDK puts into
// paths (/v1../v4 and the quarterly /2026Q3) while the spec templates carry
// no version prefix at all (versioning lives in the spec server URL).
var versionedSegment = regexp.MustCompile(`^(v[0-9]+|[0-9]{4}Q[1-4])$`)

// stripVersionedPrefix removes leading version segments so SDK wire paths
// line up with spec path templates (/v4/profiles/1/balances and
// /2026Q3/one-time-token/status both normalize into spec-template shape).
func stripVersionedPrefix(path string) string {
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")

	index := 0
	for index < len(segments) && versionedSegment.MatchString(segments[index]) {
		index++
	}

	return "/" + strings.Join(segments[index:], "/")
}

// statementVariantPattern covers the statement formats the SDK fetches but
// the OpenAPI bundle does not document (it only declares statement.json).
// The file formats (csv, pdf, xlsx, camt xml, mt940, qif) are documented in
// prose only, so they are exempt from operation matching and tallied instead.
var statementVariantPattern = regexp.MustCompile(
	`^/profiles/[^/]+/balance-statements/[^/]+/statement\.(csv|pdf|qif|xlsx|xml|mt940)$`,
)

// accountsListPath is the normalized template of the legacy recipient list.
// The bundled spec documents only the NEW paginated /accounts surface
// (a {content, seekPositionForNext, seekPositionForCurrent} envelope), while
// the legacy /v2/accounts wire this SDK targets returns a bare array.
// Response-body validation is skipped for it; the migration to the paginated
// surface is tracked in ROADMAP.md.
const accountsListPath = "/accounts"

// validateExchange checks one exchange against the spec and returns one
// message per violation (empty slice means conforming).
func validateExchange(specCtx *conformanceSpecContext, exchange conformanceExchange) []string {
	normalized := stripVersionedPrefix(exchange.rawPath)

	if statementVariantPattern.MatchString(normalized) {
		recordExemptStatementVariant(exchange.rawPath)

		return nil
	}

	specRequest, err := buildConformanceRequest(exchange, normalized)
	if err != nil {
		return []string{fmt.Sprintf("build validation request: %v", err)}
	}

	route, pathParams, findErr := specCtx.router.FindRoute(specRequest)
	if findErr != nil {
		return []string{fmt.Sprintf(
			"no operation matches the request (check for client/spec drift or a stale snapshot): %v",
			findErr,
		)}
	}

	recordConformingTemplate(exchange.method, route.Path)

	var problems []string

	requestInput := &openapi3filter.RequestValidationInput{
		Request:    specRequest,
		PathParams: pathParams,
		Route:      route,
		Options:    conformanceOptions(),
	}

	if validateErr := openapi3filter.ValidateRequest(context.Background(), requestInput); validateErr != nil {
		problems = append(problems, fmt.Sprintf("request: %v", validateErr))
	}

	if len(exchange.respBody) == 0 || exchange.skipResponseSchema {
		return problems
	}

	responseCT := exchange.respCT
	if responseCT == "" {
		responseCT = "application/json"
	}

	if !strings.Contains(responseCT, "json") {
		return problems
	}

	if normalized == accountsListPath && exchange.method == http.MethodGet {
		recordExemptLegacyAccountsList()

		return problems
	}

	responseInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestInput,
		Status:                 exchange.status,
		Header:                 http.Header{"Content-Type": []string{responseCT}},
		Options:                conformanceOptions(),
	}
	responseInput.SetBodyBytes(exchange.respBody)

	if validateErr := openapi3filter.ValidateResponse(context.Background(), responseInput); validateErr != nil {
		problems = append(problems, fmt.Sprintf("response %d: %v", exchange.status, validateErr))
	}

	return problems
}

// buildConformanceRequest reconstructs the exchange as an *http.Request the
// validator can route and decode. The URL stays path-relative so the
// relative spec server URL matches it.
func buildConformanceRequest(exchange conformanceExchange, normalizedPath string) (*http.Request, error) {
	parsed, err := url.Parse(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("normalize path: %w", err)
	}

	parsed.RawQuery = exchange.query

	body := exchange.reqBody

	request := &http.Request{
		Method: exchange.method,
		URL:    parsed,
		Header: exchange.reqHeader.Clone(),
		Body:   io.NopCloser(bytes.NewReader(body)),
		GetBody: func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(body)), nil
		},
		ContentLength: int64(len(body)),
	}

	if request.Header == nil {
		request.Header = make(http.Header)
	}

	return request, nil
}

// captureResponseWriter records what the mock handler wrote so the exchange
// can be both forwarded to the real test client and validated.
type captureResponseWriter struct {
	header     http.Header
	body       bytes.Buffer
	status     int
	wroteField bool
}

func (c *captureResponseWriter) Header() http.Header { return c.header }

func (c *captureResponseWriter) Write(data []byte) (int, error) {
	if !c.wroteField {
		c.status = http.StatusOK
		c.wroteField = true
	}

	return c.body.Write(data)
}

func (c *captureResponseWriter) WriteHeader(status int) {
	c.status = status
	c.wroteField = true
}

// Coverage bookkeeping: the coverage guard test asserts the harness saw a
// substantial slice of the SDK surface, so a broken recorder cannot produce
// a vacuous pass.

var (
	coverageMu              sync.Mutex
	conformingTemplates     = map[string]int{}
	exemptStatementPaths    = map[string]int{}
	exemptAccountsLists     int
	totalValidatedExchanges int
)

func recordConformingTemplate(method, template string) {
	coverageMu.Lock()
	defer coverageMu.Unlock()

	conformingTemplates[method+" "+template]++
	totalValidatedExchanges++
}

func recordExemptStatementVariant(rawPath string) {
	coverageMu.Lock()
	defer coverageMu.Unlock()

	exemptStatementPaths[rawPath]++
}

func recordExemptLegacyAccountsList() {
	coverageMu.Lock()
	defer coverageMu.Unlock()

	exemptAccountsLists++
}

func conformanceCoverageSnapshot() (templates []string, exemptPaths []string, exchanges, exemptAccountsListCount int) {
	coverageMu.Lock()
	defer coverageMu.Unlock()

	templates = make([]string, 0, len(conformingTemplates))
	for template := range conformingTemplates {
		templates = append(templates, template)
	}

	sort.Strings(templates)

	exemptPaths = make([]string, 0, len(exemptStatementPaths))
	for path := range exemptStatementPaths {
		exemptPaths = append(exemptPaths, path)
	}

	sort.Strings(exemptPaths)

	return templates, exemptPaths, totalValidatedExchanges, exemptAccountsLists
}
