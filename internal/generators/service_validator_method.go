package generators

import "io"

func NewServiceValidatorMethod(packageName string, messageName string, fields []ValidatorField) Generator {
	return GeneratorFunc(func(w io.Writer) error {
		return execute("service_validator_method", templateServiceValidatorMethod, w, ServiceValidatorMethod{
			PackageName: packageName,
			MessageName: messageName,
			Fields:      fields,
		})
	})
}

type ServiceValidatorMethod struct {
	PackageName string
	MessageName string
	Fields      []ValidatorField
}

type ValidatorField struct {
	FieldName string
	Rules     []string
}
