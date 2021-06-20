package internal

import (
	"google.golang.org/protobuf/compiler/protogen"
)

func generateLatestService(file *protogen.GeneratedFile, service, private *Service) error {
	packageContext := protogen.GoImportPath("context")
	packageValidation := protogen.GoImportPath("github.com/ozzo/ozzo-validation/v4")
	packageValidationIs := protogen.GoImportPath("github.com/ozzo/ozzo-validation/v4/is")
	packageCodes := protogen.GoImportPath("google.golang.org/grpc/codes")
	packageStatus := protogen.GoImportPath("google.golang.org/grpc/status")

	// Import all necessary files
	file.P("package ", service.GoPackageName)
	file.P("import (")
	file.P(packageContext)
	file.P("publicpb", service.GoImportPath)
	file.P("privatepb", private.GoImportPath)
	file.P("validation", packageValidation)
	file.P("is", packageValidationIs)
	file.P("codes", packageCodes)
	file.P("status", packageStatus)
	file.P(")")

	file.P("var _ = is.Int")

	if err := generateValidators(file, "publicpb", service); err != nil {
		return err
	}

	if err := generateLatestConverters(file, service, private); err != nil {
		return err
	}

	file.P("type Service struct {")
	file.P("Validator")
	file.P("Converter")
	file.P("Private *private.Service")
	file.P("}")

	for _, method := range service.Methods {
		publicInName := method.Input.GoIdent.GoName
		publicOutName := method.Output.GoIdent.GoName

		delegateMethod, err := findMethodDelegate(method, private)
		if err != nil {
			return err
		}

		privateInName := delegateMethod.Input.GoIdent.GoName
		privateOutName := delegateMethod.Output.GoIdent.GoName

		file.P("func(s *Service)", method.GoName, "(ctx context.Context, in *publicpb.", publicInName, ") (*publicpb.", publicOutName, ", error) {")
		file.P("if err := s.Validate", publicInName, "(in); err != nil {")
		file.P("return nil, err")
		file.P("}")

		file.P("out, _, err := s.", method.GoName, "Impl(ctx, in)")
		file.P("return out, err")
		file.P("}")

		file.P("func(s *Service)", method.GoName, "Impl(ctx context.Context, in *publicpb.", publicInName, ", mutators ...privatepb.", privateInName, "Mutator) (*publicpb.", publicOutName, ", *privatepb.", privateOutName, ", error) {")
		file.P("if err := s.Validate", publicInName, "(in); err != nil {")
		file.P("return nil, err")
		file.P("}")

		file.P("privIn := s.ToPrivate", privateInName, "(in)")
		file.P("privOut, err := s.Private.", delegateMethod.GoName, "(ctx, privIn)")
		file.P("if err != nil { return nil, err }")
		file.P("out, err := s.ToPublic", publicOutName, "(privOut)")
		file.P("if err != nil { return nil, err }")
		file.P("return out, err")
		file.P("}")
	}

	return nil
}
