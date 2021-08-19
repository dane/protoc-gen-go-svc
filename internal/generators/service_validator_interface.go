package generators

import (
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func NewServiceValidatorInterface(serviceFullName protoreflect.FullName, packageName string, messageNames []string) Generator {
	return GeneratorFunc(func(w io.Writer) error {
		return execute("service_validator_interface", templateServiceValidatorInterface, w, serviceValidatorInterface{
			ServiceFullName: serviceFullName,
			PackageName:     packageName,
			MessageNames:    messageNames,
		})
	})
}

type serviceValidatorInterface struct {
	ServiceFullName protoreflect.FullName
	PackageName     string
	MessageNames    []string
}
