package internal

import (
	"google.golang.org/protobuf/compiler/protogen"
)

func generateValidators(file *protogen.GeneratedFile, pkgName string, service *Service) error {
	file.P("type Validator interface {")
	for _, method := range service.Methods {
		name := method.Input.GoIdent.GoName
		file.P("Validate", name, "(*", pkgName, ".", name, ") error")
	}
	file.P("}")

	file.P("type validator struct {}")

	for _, method := range service.Methods {
		name := method.Input.GoIdent.GoName
		file.P("func (v validator) Validate", name, "(in *", pkgName, ".", name, ") error {")
		file.P("err := validation.Validate(in,")
		for _, field := range method.Input.Fields {
			validations, err := generateFieldValidations(pkgName, field)
			if err != nil {
				return err
			}

			if len(validations) == 0 {
				continue
			}

			file.P("validation.Field(&in.", field.GoName, ",")
			for _, validationField := range validations {
				file.P(validationField, ",")
			}
			file.P("),")
		}
		file.P(")")
		file.P("if err != nil {")
		file.P("return status.Error(codes.InvalidArgument, err.Error())")
		file.P("}")
		file.P("return nil")
		file.P("}")
	}

	return nil
}
