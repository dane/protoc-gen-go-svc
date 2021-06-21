package internal

import (
	"google.golang.org/protobuf/compiler/protogen"
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

	file.P("type Service struct {")
	file.P("Validator")
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
