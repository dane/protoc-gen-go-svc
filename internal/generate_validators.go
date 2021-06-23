package internal

import (
	"google.golang.org/protobuf/compiler/protogen"
)

func generateValidators(file *protogen.GeneratedFile, pkgName string, service *Service) error {
	file.P(`const ValidatorName = "`, service.Service.Desc.FullName(), `.Validator"`)

	file.P("type Validator interface {")
	file.P("Name() string")
	for _, method := range service.Methods {
		name := messageName(method.Input)
		file.P("Validate", name, "(*", pkgName, ".", name, ") error")
		messagesByImportPath[method.Input.GoIdent.GoImportPath][name].Validated = true
	}

	for _, message := range messagesByImportPath[service.GoIdent.GoImportPath] {
		if message.Validated {
			continue
		}

		if !message.MustValidate {
			continue
		}

		name := messageName(message.Message)
		file.P("Validate", name, "(*", pkgName, ".", name, ") error")
	}
	file.P("}")

	// Reset validate generation tracking.
	for _, message := range messagesByImportPath[service.GoIdent.GoImportPath] {
		message.Validated = false
	}

	file.P("func NewValidator() Validator { return validator{} }")

	file.P("type validator struct {}")
	file.P("func (v validator) Name() string { return ValidatorName }")

	for _, method := range service.Methods {
		if err := generateMessageValidator(file, pkgName, method.Input); err != nil {
			return err
		}

		messagesByImportPath[method.Input.GoIdent.GoImportPath][messageName(method.Input)].Validated = true
	}

	for _, message := range messagesByImportPath[service.GoIdent.GoImportPath] {
		if message.MustValidate && !message.Validated {
			protoMessage := message.Message
			if err := generateMessageValidator(file, pkgName, protoMessage); err != nil {
				return err
			}
			messagesByImportPath[message.GoIdent.GoImportPath][messageName(protoMessage)].Validated = true
		}
	}

	return nil
}

func generateMessageValidator(file *protogen.GeneratedFile, pkgName string, message *protogen.Message) error {
	name := messageName(message)
	file.P("func (v validator) Validate", name, "(in *", pkgName, ".", name, ") error {")
	file.P("err := validation.ValidateStruct(in,")
	for _, field := range message.Fields {
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

	return nil
}
