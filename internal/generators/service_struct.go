package generators

import (
	"io"

	"google.golang.org/protobuf/compiler/protogen"
)

func NewServiceStruct(imports []protogen.GoIdent, service *Service, fields []string) Generator {
	return GeneratorFunc(func(w io.Writer) error {
		return execute("service_struct", templateServiceStruct, w, serviceStruct{
			GoPackageName: service.GoPackageName,
			Imports:       imports,
			Fields:        fields,
		})
	})
}

type serviceStruct struct {
	PluginVersion string
	GoPackageName protogen.GoPackageName
	Imports       []protogen.GoIdent
	Fields        []string
}
