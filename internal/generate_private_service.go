package internal

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func generatePrivateService(file *protogen.GeneratedFile, service *Service) error {
	packageContext := protogen.GoImportPath("context")
	packageValidation := protogen.GoImportPath("github.com/go-ozzo/ozzo-validation/v4")
	packageValidationIs := protogen.GoImportPath("github.com/go-ozzo/ozzo-validation/v4/is")
	packageCodes := protogen.GoImportPath("google.golang.org/grpc/codes")
	packageStatus := protogen.GoImportPath("google.golang.org/grpc/status")

	// Import all necessary files
	file.P("package ", service.GoPackageName)
	file.P("import (")
	file.P(packageContext)
	file.P("privatepb", service.GoImportPath)
	file.P("validation", packageValidation)
	file.P("is", packageValidationIs)
	file.P("codes", packageCodes)
	file.P("status", packageStatus)
	file.P(")")

	file.P("var _ = is.Int")

	if err := generateValidators(file, "privatepb", service); err != nil {
		return err
	}

	for _, method := range service.Methods {
		file.P("type ", method.Input.GoIdent.GoName, "Mutator func(*privatepb.", method.Input.GoIdent.GoName, ")")
	}

	for _, method := range service.Methods {
		for _, field := range method.Input.Fields {
			var fieldType string
			switch field.Desc.Kind() {
			case protoreflect.MessageKind:
				fieldType = fmt.Sprintf("*privatepb.%s", field.Message.GoIdent.GoName)
			case protoreflect.EnumKind:
				fieldType = fmt.Sprintf("privatepb.%s", field.Enum.GoIdent.GoName)
			case protoreflect.FloatKind:
				fieldType = "float64"
			default:
				fieldType = field.Desc.Kind().String()

			}
			file.P("func Set", field.GoIdent.GoName, "(value ", fieldType, ") ", method.Input.GoIdent.GoName, "Mutator {")
			file.P("return func(in *privatepb.", method.Input.GoIdent.GoName, ") {")
			file.P("in.", field.GoName, " = value")
			file.P("}")
			file.P("}")
		}
	}

	file.P("type Service struct {")
	file.P("Validator")
	file.P("privatepb.", service.Service.GoName, "Server")
	file.P("Impl privatepb.", service.Service.GoName, "Server")
	file.P("}")

	for _, method := range service.Methods {
		file.P("func(s *Service)", method.GoName, "(ctx context.Context, in *privatepb.", method.Input.GoIdent.GoName, ") (*privatepb.", method.Output.GoIdent.GoName, ", error) {")
		file.P("if err := s.Validate", method.Input.GoIdent.GoName, "(in); err != nil {")
		file.P("return nil, err")
		file.P("}")
		file.P("return s.Impl.", method.GoName, "(ctx, in)")
		file.P("}")
	}

	return nil
}
