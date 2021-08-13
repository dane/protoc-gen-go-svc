package internal

import (
	"io"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/dane/protoc-gen-go-svc/internal/generators"
)

type ServiceRegisterGenerator struct {
	PluginVersion string
	Imports       []protogen.GoIdent
	Services      []*generators.Service
	Private       *generators.Service
}

func (g *ServiceRegisterGenerator) Generate(w io.Writer) error {
	return execute("service_register", templateServiceRegister, w, g)
}
