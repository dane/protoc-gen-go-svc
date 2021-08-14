package generators

import "io"

func NewServiceMutators(messageName string, fields []MutatorField) Generator {
	return GeneratorFunc(func(w io.Writer) error {
		return execute("service_mutators", templateServiceMutators, w, serviceMutators{
			MessageName: messageName,
			Fields:      fields,
		})
	})
}

type serviceMutators struct {
	MessageName string
	Fields      []MutatorField
}

type MutatorField struct {
	FieldName string
	FieldType string
}
