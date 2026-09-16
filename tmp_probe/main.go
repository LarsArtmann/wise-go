package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

func main() {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("docs/reviews/wise-api-openapi.json")
	if err != nil {
		panic(err)
	}
	doc.Servers = []*openapi3.Server{{URL: "/"}}
	if doc.Info.Version == "" {
		doc.Info.Version = "wise-snapshot"
	}
	router, err := legacyrouter.NewRouter(doc, openapi3.DisableExamplesValidation(), openapi3.DisableSchemaFormatValidation())
	if err != nil {
		panic(err)
	}

	profilesJSON := `[
		{"id":12345,"type":"PERSONAL","firstName":"John","lastName":"Doe","email":"john@example.com","createdAt":"2023-01-15T10:30:00Z"},
		{"id":67890,"type":"BUSINESS","businessName":"Acme Corp","email":"billing@acme.com","createdAt":"2023-02-20T14:00:00Z"}
	]`

	req := &http.Request{Method: "GET", URL: &url.URL{Path: "/profiles"}, Header: http.Header{"Content-Type": []string{"application/json"}}}
	route, params, err := router.FindRoute(req)
	if err != nil {
		panic(err)
	}
	in := &openapi3filter.RequestValidationInput{Request: req, PathParams: params, Route: route, Options: &openapi3filter.Options{MultiError: true, AuthenticationFunc: openapi3filter.NoopAuthenticationFunc}}
	resp := &openapi3filter.ResponseValidationInput{RequestValidationInput: in, Status: 200, Header: http.Header{"Content-Type": []string{"application/json"}}}
	resp.SetBodyBytes([]byte(profilesJSON))
	if err := openapi3filter.ValidateResponse(context.Background(), resp); err != nil {
		fmt.Println("ERRORS:")
		for _, line := range strings.Split(err.Error(), "\n") {
			fmt.Println(" ", line)
		}
		return
	}
	fmt.Println("profiles conform (with format validation disabled)")
}
