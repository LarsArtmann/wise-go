module github.com/larsartmann/wise-go

// v0.6.0 was tagged on a pre-merge lineage that lacks the v0.5.2/v0.5.3
// fixes (balance listing types param, Corruption classification). Use v0.6.1+.
retract v0.6.0

go 1.27

require (
	github.com/getkin/kin-openapi v0.149.0
	github.com/larsartmann/go-branded-id v0.7.0
	github.com/larsartmann/go-error-family v0.11.0
	github.com/larsartmann/go-retry v0.7.1
	github.com/onsi/ginkgo/v2 v2.33.0
	github.com/onsi/gomega v1.44.0
)

require (
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-openapi/jsonpointer v1.0.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260926063103-aaccee046517 // indirect
	github.com/oasdiff/yaml v0.1.1 // indirect
	github.com/oasdiff/yaml3 v0.0.14 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	golang.org/x/tools v0.50.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
