package internal

import (
	"io"

	"google.golang.org/protobuf/compiler/protogen"
)

type ServiceStructGenerator struct {
	PluginVersion string
	GoPackageName protogen.GoPackageName
	Imports       []protogen.GoIdent
	Fields        []string
}

func (g *ServiceStructGenerator) Generate(w io.Writer) error {
	return execute("service_struct", templateServiceStruct, w, g)
}
