package internal

import (
	"fmt"
	"io"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

type ServiceRegisterGenerator struct {
	PluginVersion string
	Imports       []protogen.GoIdent
	Services      []*Service
	Private       *Service
}

func (g *ServiceRegisterGenerator) Generate(w io.Writer) error {
	tmpl := template.Must(
		template.
			New("service_register").
			Funcs(templateFuncs()).
			Parse(templateServiceRegister),
	)

	return tmpl.Execute(w, g)
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
