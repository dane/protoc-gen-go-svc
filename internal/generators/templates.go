package generators

import _ "embed"

var (
	//go:embed templates/service_register.go.tmpl
	templateServiceRegister string

	//go:embed templates/partial_service_struct.go.tmpl
	templateServiceStruct string

	//go:embed templates/partial_service_method.go.tmpl
	templateServiceMethod string
)
