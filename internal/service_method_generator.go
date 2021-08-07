package internal

import (
	"io"
)

type ServiceMethodGenerator struct {
	MethodName  string
	InputName   string
	OutputName  string
	PackageName string
	ToPrivate   bool
}

func (g *ServiceMethodGenerator) Generate(w io.Writer) error {
	return execute("service_method", templateServiceMethod, w, g)
}
