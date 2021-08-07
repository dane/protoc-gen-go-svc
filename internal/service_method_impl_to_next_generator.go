package internal

import "io"

type ServiceMethodImplToNextGenerator struct {
	MethodName string
	InputName  string
	OutputName string

	PrivateInputName  string
	PrivateOutputName string

	NextMethodName string
	NextInputName  string

	DeprecatedFields []DeprecatedField
}

type DeprecatedField struct {
	FieldName        string
	PrivateFieldName string
}

func (g *ServiceMethodImplToNextGenerator) Generate(w io.Writer) error {
	return execute("service_method_impl_to_next", templateServiceMethodImpToNext, w, g)
}
