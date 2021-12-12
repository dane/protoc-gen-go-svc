package internal

import _ "embed"

var (
	//go:embed templates/register.pb.go.tmpl
	registerTemplate string

	//go:embed templates/service.pb.go.tmpl
	serviceTemplate string

	//go:embed templates/testing.go.tmpl
	testingTemplate string

	//go:embed templates/partials/converters.go.tmpl
	convertersPartial string

	//go:embed templates/partials/validators.go.tmpl
	validatorsPartial string

	//go:embed templates/partials/mutators.go.tmpl
	mutatorsPartial string

	//go:embed templates/partials/handlers.go.tmpl
	handlersPartial string

	//go:embed templates/partials/impls.go.tmpl
	implsPartial string
)

var Partials = []string{
	convertersPartial,
	validatorsPartial,
	mutatorsPartial,
	handlersPartial,
	implsPartial,
}
