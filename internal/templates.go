package internal

import (
	_ "embed"
	"fmt"
	"io"
	"text/template"
)

var (
	//go:embed templates/service_register.go.tmpl
	templateServiceRegister string

	//go:embed templates/partial_service_method.go.tmpl
	templateServiceMethod string
)

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
