package internal

import "io"

type ServiceMethodImplToPrivateGenerator struct {
	MethodName string
	InputName  string
	OutputName string

	PrivateMethodName string
	PrivateInputName  string
	PrivateOutputName string

	Prefix string
}

func (g *ServiceMethodImplToPrivateGenerator) Generate(w io.Writer) error {
	return execute("service_method_impl_to_private", templateServiceMethodImpToPrivate, w, g)
}
