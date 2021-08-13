package generators

import (
	"fmt"
	"io"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

type Generator interface {
	Generate(io.Writer) error
}

type GeneratorFunc func(io.Writer) error

func (g GeneratorFunc) Generate(w io.Writer) error {
	return g(w)
}

type Service struct {
	protogen.GoPackageName
	protogen.GoIdent
	*protogen.Service

	GoName              string
	GoServiceImportPath protogen.GoImportPath
	Messages            []*protogen.Message
	Enums               []*protogen.Enum
	DeprecatedMessages  []*protogen.Message
	DeprecatedEnums     []*protogen.Enum
}

func (s *Service) PackageName() string {
	return fmt.Sprintf("%s", s.GoPackageName)
}

func execute(name string, templateStr string, w io.Writer, params interface{}) error {
	tmpl := template.Must(
		template.
			New(name).
			Funcs(templateFuncs()).
			Parse(templateStr),
	)

	return tmpl.Execute(w, params)
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"sprintf": fmt.Sprintf,
		"previous": func(current *Service, services []*Service) *Service {
			for i, service := range services {
				if current == service {
					return services[i-1]
				}
			}
			return nil
		},
	}
}
