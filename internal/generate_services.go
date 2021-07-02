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

	if err := generateConverters(file, service, chain, LatestPublicService); err != nil {
		return err
	}

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

	if err := generateConverters(file, service, chain, PublicService); err != nil {
		return err
	}

	generateServiceMethods(file, service, PublicService)

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

func generateConverters(file *protogen.GeneratedFile, service *Service, chain []*Service, serviceType ServiceType) error {
	next := chain[0]
	private := chain[len(chain)-1]

	file.P("type Converter interface {")
	for _, method := range service.Methods {
		publicIn := method.Input
		publicOut := method.Output

		if deprecatedMethod(method) || serviceType == LatestPublicService {
			if err := generateConverterToPrivateIface(file, publicIn, publicOut, private); err != nil {
				return err
			}
			continue
		}

		privateOut, err := findPrivateMessage(publicOut, chain)
		if err != nil {
			return err
		}

		nextIn, err := findNextMessage(publicIn, next)
		if err != nil {
			return err
		}

		nextOut, err := findNextMessage(publicOut, next)
		if err != nil {
			return err
		}

		publicInName := publicIn.GoIdent.GoName
		publicOutName := publicOut.GoIdent.GoName
		privateOutName := privateOut.GoIdent.GoName
		nextInName := nextIn.GoIdent.GoName
		nextOutName := nextOut.GoIdent.GoName

		file.P("ToNext", nextInName, "(*publicpb.", publicInName, ") *nextpb.", nextInName)
		file.P("ToPublic", publicOutName, "(*nextpb.", nextOutName, ", *privatepb.", privateOutName, ") (*publicpb.", publicOutName, ", error)")
	}

	for _, message := range service.Messages {
		_, isInput := inputs[message]
		_, isOutput := outputs[message]
		if isInput || isOutput {
			continue
		}

		if serviceType == LatestPublicService {
			if err := generateConverterToPrivateIface(file, message, message, private); err != nil {
				return err
			}
			continue
		}

		if err := generateConverterToNextIface(file, message, chain); err != nil {
			return err
		}
	}

	for _, enum := range service.Enums {
		if serviceType == LatestPublicService {
			if err := generateConverterToPrivateIface(file, enum, enum, private); err != nil {
				return err
			}
			continue
		}

		if err := generateConverterToNextIface(file, enum, chain); err != nil {
			return err
		}
	}
	file.P("}")

	return nil
}

func generateConverterToPrivateIface(file *protogen.GeneratedFile, publicIn, publicOut interface{}, private *Service) error {
	var publicInName, publicOutName, privateOutName, privateInName, pointer string
	switch publicIn.(type) {
	case *protogen.Message:
		value := publicIn.(*protogen.Message)
		privateValue, err := findNextMessage(value, private)
		if err != nil {
			return err
		}

		publicInName = value.GoIdent.GoName
		privateInName = privateValue.GoIdent.GoName
		pointer = "*"
	case *protogen.Enum:
		value := publicIn.(*protogen.Enum)
		privateValue, err := findNextEnum(value, private)
		if err != nil {
			return err
		}

		publicInName = value.GoIdent.GoName
		privateInName = privateValue.GoIdent.GoName
	}

	switch publicOut.(type) {
	case *protogen.Message:
		value := publicOut.(*protogen.Message)
		privateValue, err := findNextMessage(value, private)
		if err != nil {
			return err
		}

		publicOutName = value.GoIdent.GoName
		privateOutName = privateValue.GoIdent.GoName
		pointer = "*"
	case *protogen.Enum:
		value := publicOut.(*protogen.Enum)
		privateValue, err := findNextEnum(value, private)
		if err != nil {
			return err
		}

		publicOutName = value.GoIdent.GoName
		privateOutName = privateValue.GoIdent.GoName
	}

	file.P("ToPrivate", privateInName, "(", pointer, "publicpb.", publicInName, ") ", pointer, "privatepb.", privateInName)
	file.P("ToPublic", publicOutName, "(", pointer, "privatepb.", privateOutName, ") (", pointer, "publicpb.", publicOutName, ", error)")

	return nil
}

func generateConverterToNextIface(file *protogen.GeneratedFile, v interface{}, chain []*Service) error {
	next := chain[0]
	var publicName, privateName, nextName, pointer string
	switch v.(type) {
	case *protogen.Message:
		value := v.(*protogen.Message)
		privateMessage, err := findPrivateMessage(value, chain)
		if err != nil {
			return err
		}

		nextMessage, err := findNextMessage(value, next)
		if err != nil {
			return err
		}

		publicName = value.GoIdent.GoName
		privateName = privateMessage.GoIdent.GoName
		nextName = nextMessage.GoIdent.GoName
		pointer = "*"
	case *protogen.Enum:
		value := v.(*protogen.Enum)
		privateEnum, err := findPrivateEnum(value, chain)
		if err != nil {
			return err
		}

		nextEnum, err := findNextEnum(value, next)
		if err != nil {
			return err
		}

		publicName = value.GoIdent.GoName
		privateName = privateEnum.GoIdent.GoName
		nextName = nextEnum.GoIdent.GoName
	default:
		return fmt.Errorf("failed to generate converter interface for %T", v)
	}

	file.P("ToNext", nextName, "(", pointer, "publicpb.", publicName, ") ", pointer, "nextpb.", nextName)
	file.P("ToPublic", publicName, "(", pointer, "nextpb.", nextName, ", ", pointer, "privatepb.", privateName, ") (", pointer, "publicpb.", publicName, ", error)")

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
