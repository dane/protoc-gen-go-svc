package generators

import (
	"io"

	"google.golang.org/protobuf/compiler/protogen"
)

func NewServiceRegister(imports []protogen.GoIdent, services []*Service, private *Service) Generator {
	return GeneratorFunc(func(w io.Writer) error {
		return execute("service_register", templateServiceRegister, w, serviceRegister{
			Imports:  imports,
			Services: services,
			Private:  private,
		})
	})
}

type serviceRegister struct {
	Imports  []protogen.GoIdent
	Services []*Service
	Private  *Service
}
