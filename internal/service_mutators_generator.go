package internal

type ServiceMutatorGenerator struct {
	MessageName string
	Fields      []MutatorField
}

type MutatorField struct {
	FieldName string
	FieldType string
}
