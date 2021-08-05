package internal

import (
	"io"

	"google.golang.org/protobuf/compiler/protogen"
)

type ServiceRegisterGenerator struct {
	PluginVersion string
	Imports       []protogen.GoIdent
	Services      []*Service
	Private       *Service
}

func (g *ServiceRegisterGenerator) Generate(w io.Writer) error {
	return execute("service_register", templateServiceRegister, w, g)
}
