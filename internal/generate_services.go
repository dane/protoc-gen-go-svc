package internal

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func generatePrivateService(file *protogen.GeneratedFile, service *Service) error {
	imports := commonImports(
		service.GoImportPath.Ident("privatepb"),
	)

	file.P("package ", service.GoPackageName)
	file.P("import (")
	for _, ident := range imports {
		file.P(ident.GoName, ident.GoImportPath)
	}
	file.P(")")

	generateImportUsage(file,
		fmt.Sprintf("*publicpb.%sServer", service.GoName),
	)

	generateServiceStruct(file,
		fmt.Sprintf("Impl privatepb.%sServer", service.GoName),
	)

	generateServiceMethods(file, service, PrivateService)
	if err := generateServiceValidators(file, "privatepb", service); err != nil {
		return err
	}

	if err := generateMutators(file, service); err != nil {
		return err
	}

	if err := generateServiceValidators(file, "privatepb", service); err != nil {
		return err
	}

	return nil
}

func generateLatestPublicService(file *protogen.GeneratedFile, service *Service, chain []*Service) error {
	private := chain[len(chain)-1]
	imports := commonImports(
		service.GoImportPath.Ident("publicpb"),
		private.GoImportPath.Ident("privatepb"),
		private.GoServiceImportPath.Ident("private"),
	)

	file.P("package ", service.GoPackageName)
	file.P("import (")
	for _, ident := range imports {
		file.P(ident.GoName, ident.GoImportPath)
	}
	file.P(")")

	generateImportUsage(file,
		fmt.Sprintf("*publicpb.%sServer", service.GoName),
		fmt.Sprintf("*privatepb.%sServer", private.GoName),
		"*private.Service",
	)

	generateServiceStruct(file,
		"Converter Converter",
		"Private *privatepb.Service",
		fmt.Sprintf("publicpb.%sServer", service.GoName),
	)

	generateServiceMethods(file, service, LatestPublicService)
	for _, method := range service.Methods {
		if err := generateServiceMethodToPrivateImpl(file, method, private); err != nil {
			return err
		}
	}

	if err := generateServiceValidators(file, "publicpb", service); err != nil {
		return err
	}

	// TODO: generate converters

	return nil
}

func generatePublicService(file *protogen.GeneratedFile, service *Service, chain []*Service) error {
	next := chain[0]
	private := chain[len(chain)-1]
	imports := commonImports(
		service.GoImportPath.Ident("publicpb"),
		next.GoImportPath.Ident("nextpb"),
		private.GoImportPath.Ident("privatepb"),
		private.GoServiceImportPath.Ident("private"),
		next.GoServiceImportPath.Ident("next"),
	)

	file.P("package ", service.GoPackageName)
	file.P("import (")
	for _, ident := range imports {
		file.P(ident.GoName, ident.GoImportPath)
	}
	file.P(")")

	generateImportUsage(file,
		fmt.Sprintf("*publicpb.%sServer", service.GoName),
		fmt.Sprintf("*privatepb.%sServer", private.GoName),
		fmt.Sprintf("*nextpb.%sServer", next.GoName),
		"*private.Service",
		"*next.Service",
	)

	generateServiceStruct(file,
		"Converter Converter",
		"Private *privatepb.Service",
		"Next *nextpb.Service",
		fmt.Sprintf("publicpb.%sServer", service.GoName),
	)

	if err := generateServiceValidators(file, "publicpb", service); err != nil {
		return err
	}

	generateServiceMethods(file, service, PublicService)
	//if err := generatePublicServiceMethodImpls(file, service, private); err != nil {
	//	return err
	//}

	for _, method := range service.Methods {
		if deprecatedMethod(method) {
			if err := generateServiceMethodToPrivateImpl(file, method, private); err != nil {
				return err
			}
			continue
		}

		if err := generateServiceMethodToNextImpl(file, method, chain); err != nil {
			return err
		}
	}

	return nil
}

func commonImports(imports ...protogen.GoIdent) []protogen.GoIdent {
	return append([]protogen.GoIdent{
		protogen.GoImportPath("context").Ident("context"),
		protogen.GoImportPath("github.com/go-ozzo/ozzo-validation/v4").Ident("validation"),
		protogen.GoImportPath("github.com/go-ozzo/ozzo-validation/v4/is").Ident("is"),
		protogen.GoImportPath("google.golang.org/grpc/codes").Ident("codes"),
		protogen.GoImportPath("google.golang.org/grpc/status").Ident("status"),
	}, imports...)
}

func generateImportUsage(file *protogen.GeneratedFile, refs ...string) {
	refs = append([]string{
		"validation.Validatable",
		"is.Int",
		"codes.Code",
		"status.Status",
	}, refs...)

	for _, ref := range refs {
		file.P("var _ =", ref)
	}
}

func generateServiceStruct(file *protogen.GeneratedFile, refs ...string) {
	file.P("type Service struct {")
	file.P("Validator Validator")
	for _, ref := range refs {
		file.P(ref)
	}
	file.P("}")
}

func generateServiceMethods(file *protogen.GeneratedFile, service *Service, serviceType ServiceType) {
	for _, method := range service.Methods {
		methodName := method.GoName
		inName := method.Input.GoIdent.GoName
		outName := method.Output.GoIdent.GoName

		switch serviceType {
		case PublicService, LatestPublicService:
			generatePublicServiceMethod(file, methodName, inName, outName)
		case PrivateService:
			generatePrivateServiceMethod(file, methodName, inName, outName)
		}
	}
}

func generatePublicServiceMethod(file *protogen.GeneratedFile, methodName, inName, outName string) {
	file.P("func (s *Service) ", methodName, "(ctx context.Context, in *publicpb.", inName, ") (*publicpb.", outName, ", error) {")
	file.P("if err := s.Validate", inName, "(in); err != nil { return nil, nil, err }")
	file.P("out, _, err := s.", methodName, "Impl(ctx, in)")
	file.P("return out, err")
	file.P("}")
}

func generatePrivateServiceMethod(file *protogen.GeneratedFile, methodName, inName, outName string) {
	file.P("func (s *Service) ", methodName, "(ctx context.Context, in *privatepb.", inName, ") (*privatepb.", outName, ", error) {")
	file.P("if err := s.Validate", inName, "(in); err != nil { return nil, err }")
	file.P("return s.Impl.", methodName, "(ctx, in)")
	file.P("}")
}

func generateServiceMethodToPrivateImpl(file *protogen.GeneratedFile, method *protogen.Method, private *Service) error {
	publicMethodName := method.GoName
	publicInName := method.Input.GoIdent.GoName
	publicOutName := method.Output.GoIdent.GoName

	privateMethod, err := findNextMethod(method, private)
	if err != nil {
		return fmt.Errorf("failed to generate service %s method impl: %w", publicMethodName, err)
	}

	privateIn, err := findNextMessage(method.Input, private)
	if err != nil {
		return err
	}

	privateOut, err := findNextMessage(method.Output, private)
	if err != nil {
		return err
	}

	privateMethodName := privateMethod.GoName
	privateInName := privateIn.GoIdent.GoName
	privateOutName := privateOut.GoIdent.GoName

	file.P("func (s *Service) ", publicMethodName, "Impl(ctx context.Context, in *publicpb.", publicInName, ", mutators ...private.", privateInName, "Mutator) (*publicpb.", publicOutName, ", *privatepb.", privateOutName, ", error) {")
	file.P("privateIn := s.ToPrivate", privateInName, "(in)")
	file.P("private.Apply", privateInName, "Mutators(privateIn, mutators)")
	file.P("privateOut, err := s.Private.", privateMethodName, "(ctx, privateIn)")
	file.P("if err != nil { return nil, nil, err }")
	file.P("out, err := s.ToPublic", publicOutName, "(privateOut)")
	file.P("if err != nil { return nil, nil, err }")
	file.P("return out, privateOut, nil")
	file.P("}")

	return nil
}

func generateServiceMethodToNextImpl(file *protogen.GeneratedFile, method *protogen.Method, chain []*Service) error {
	publicMethodName := method.GoName
	publicInName := method.Input.GoIdent.GoName
	publicOutName := method.Output.GoIdent.GoName

	next := chain[0]
	nextMethod, err := findNextMethod(method, next)
	if err != nil {
		return fmt.Errorf("failed to generate service %s method impl: %w", publicMethodName, err)
	}

	nextIn, err := findNextMessage(method.Input, next)
	if err != nil {
		return err
	}

	nextMethodName := nextMethod.GoName
	nextInName := nextIn.GoIdent.GoName

	privateIn, err := findPrivateMessage(method.Input, chain)
	if err != nil {
		return err
	}

	privateOut, err := findPrivateMessage(method.Output, chain)
	if err != nil {
		return err
	}

	privateInName := privateIn.GoIdent.GoName
	privateOutName := privateOut.GoIdent.GoName

	file.P("func (s *Service) ", publicMethodName, "Impl(ctx context.Context, in *publicpb.", publicInName, ", mutators ...private.", privateInName, "Mutator) (*publicpb.", publicOutName, ", *privatepb.", privateOutName, ", error) {")
	file.P("nextIn := s.ToNext", nextInName, "(in)")
	file.P("nextOut, privateOut, err := s.Next.", nextMethodName, "Impl(ctx, nextIn, mutators...)")
	file.P("if err != nil { return nil, nil, err }")
	file.P("out, err := s.ToPublic", publicOutName, "(nextOut, privateOut)")
	file.P("if err != nil { return nil, nil, err }")
	file.P("return out, privateOut, nil")
	file.P("}")

	return nil
}

func generateServiceValidators(file *protogen.GeneratedFile, packageName string, service *Service) error {
	file.P("type Validator interface {")
	for _, message := range service.Messages {
		if !validateMessage(message) {
			continue
		}

		messageName := message.GoIdent.GoName
		file.P("Validate", messageName, "(*", packageName, ".", messageName, ") error")
	}
	file.P("}")

	file.P("type validator struct {}")
	for _, message := range service.Messages {
		if !validateMessage(message) {
			continue
		}

		messageName := message.GoIdent.GoName
		file.P("func(v validator) Validate", messageName, "(in *", packageName, ".", messageName, ") error {")
		file.P("err := validation.ValidateStruct(in,")
		for _, field := range message.Fields {
			if !validateField(field) {
				continue
			}

			fieldName := field.GoName
			file.P("validation.Field(&in.", fieldName, ",")

			if required(field) {
				file.P("validation.Required,")
			}

			isName, err := is(field)
			if err != nil {
				return err
			}
			switch isName {
			case "uuid":
				file.P("is.UUID,")
			case "url":
				file.P("is.URL,")
			case "email":
				file.P("is.Email,")
			}

			inValues, err := in(packageName, field)
			if err != nil {
				return err
			}

			if inValues != nil {
				file.P("validation.In(", strings.Join(inValues, ","), "),")
			}

			minValue, minSet, err := min(field)
			if err != nil {
				return err
			}

			maxValue, maxSet, err := max(field)
			if err != nil {
				return err
			}

			if minSet || maxSet {
				if !minSet {
					minValue = "0"
				}

				if !maxSet {
					maxValue = "0"
				}

				switch field.Desc.Kind() {
				case protoreflect.StringKind:
					file.P("validation.Length(", minValue, ",", maxValue, "),")
				case protoreflect.FloatKind, protoreflect.Uint64Kind, protoreflect.Int64Kind:
					if minSet {
						file.P("validation.Min(", minValue, "),")
					}

					if maxSet {
						file.P("validation.Max(", maxValue, "),")
					}
				default:
					return fmt.Errorf(`invalid field type for "min/max" validate annotations`)
				}
			}

			if field.Message != nil && validateMessage(message) {
				messageName := field.Message.GoIdent.GoName
				file.P("validation.By(func(interface{}) error { return v.Validate", messageName, "(in.", fieldName, ") }),")
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

func generateMutators(file *protogen.GeneratedFile, service *Service) error {
	for _, method := range service.Methods {
		messageName := method.Input.GoIdent.GoName
		file.P("type ", messageName, "Mutator func(*privatepb.", messageName, ")")

		for _, field := range method.Input.Fields {
			fieldName := field.GoName
			fieldType, err := findFieldType("privatepb", field)
			if err != nil {
				return fmt.Errorf("failed to generate mutator for %s: %w", messageName, err)
			}
			file.P("func Set", messageName, "_", fieldName, "(value ", fieldType, ") ", messageName, "Mutator {")
			file.P("return func(in *privatepb.", messageName, ") {")
			file.P("in.", fieldName, "= value")
			file.P("}")
			file.P("}")
		}

		file.P("func Apply", messageName, "Mutators(in *privatepb.", messageName, ", mutators []", messageName, "Mutator) {")
		file.P("for _, mutator := range mutators {")
		file.P("mutator(in)")
		file.P("}")
		file.P("}")
	}

	return nil
}
