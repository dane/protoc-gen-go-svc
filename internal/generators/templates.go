package generators

import _ "embed"

var (
	//go:embed templates/service_register.go.tmpl
	templateServiceRegister string

	//go:embed templates/partial_service_struct.go.tmpl
	templateServiceStruct string

	//go:embed templates/partial_service_method.go.tmpl
	templateServiceMethod string

	//go:embed templates/partial_service_method_impl_to_next.go.tmpl
	templateServiceMethodImplToNext string

	//go:embed templates/partial_service_method_impl_to_private.go.tmpl
	templateServiceMethodImplToPrivate string

	//go:embed templates/partial_service_mutators.go.tmpl
	templateServiceMutators string

	//go:embed templates/partial_service_validator_interface.go.tmpl
	templateServiceValidatorInterface string

	//go:embed templates/partial_service_validator_method.go.tmpl
	templateServiceValidatorMethod string
)
