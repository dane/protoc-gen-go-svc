package internal

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func generateServiceRegister(file *protogen.GeneratedFile, chain []*Service) error {
	private := chain[len(chain)-1]
	services := chain[:len(chain)-1]
	imports := []protogen.GoIdent{
		protogen.GoImportPath("google.golang.org/grpc").Ident("grpc"),
	}

	for _, service := range chain {
		packageName := string(service.GoPackageName)
		path := service.GoImportPath.Ident(fmt.Sprintf("%spb", packageName))
		imports = append(imports, path)
		imports = append(imports, service.GoServiceImportPath.Ident(packageName))
	}

	file.P("package service")
	file.P("import (")
	for _, ident := range imports {
		file.P(ident.GoName, ident.GoImportPath)
	}
	file.P(")")

	file.P("type Option interface { Name() string }")

	file.P("func RegisterServer(server *grpc.Server, impl privatepb.", private.GoName, "Server, options ...Option) {")
	file.P("servicePrivate := &", private.GoPackageName, ".Service{")
	file.P("Validator: ", private.GoPackageName, ".NewValidator(),")
	file.P("Impl: impl,")
	file.P("}")

	sort.Sort(sort.Reverse(byPackageName(services)))
	for i, service := range services {
		packageName := service.GoPackageName
		serviceName := service.GoName
		varName := fmt.Sprintf("service%s", packageName)
		file.P(varName, ":= &", packageName, ".Service{")
		file.P("Validator: ", packageName, ".NewValidator(),")
		file.P("Converter: ", packageName, ".NewConverter(),")
		file.P("Private: servicePrivate,")
		if i > 0 {
			nextVarName := fmt.Sprintf("service%s", services[i-1].GoPackageName)
			file.P("Next: ", nextVarName, ",")
		}
		file.P("}")

		file.P(packageName, "pb.Register", serviceName, "Server(server, ", varName, ")")
	}

	file.P("for _, opt := range options {")
	file.P("switch opt.Name() {")
	file.P("case private.ValidatorName:")
	file.P("servicePrivate.Validator = opt.(private.Validator)")

	for _, service := range services {
		packageName := service.GoPackageName
		varName := fmt.Sprintf("service%s", packageName)
		file.P("case ", packageName, ".ConverterName:")
		file.P(varName, ".Converter = opt.(", packageName, ".Converter)")
		file.P("case ", packageName, ".ValidatorName:")
		file.P(varName, ".Validator = opt.(", packageName, ".Validator)")
	}
	file.P("}")
	file.P("}")

	file.P("}")

	return nil
}

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

	generateImportUsage(file)
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

	generateImportUsage(file)
	generateServiceStruct(file,
		"Converter",
		"Private *private.Service",
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

	logger.Printf("package=%s at=import-usage", service.GoPackageName)
	generateImportUsage(file)
	logger.Printf("package=%s at=service-struct", service.GoPackageName)
	generateServiceStruct(file,
		"Converter",
		"Private *private.Service",
		"Next *next.Service",
		fmt.Sprintf("publicpb.%sServer", service.GoName),
	)

	logger.Printf("package=%s at=generate-validators", service.GoPackageName)
	if err := generateServiceValidators(file, "publicpb", service); err != nil {
		return err
	}

	logger.Printf("package=%s at=generate-converters", service.GoPackageName)
	if err := generateConverters(file, service, chain, PublicService); err != nil {
		return err
	}

	logger.Printf("package=%s at=generate-methods", service.GoPackageName)
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

	file.P(`const ConverterName = "`, service.Desc.FullName(), `.Converter"`)
	file.P("func NewConverter() Converter { return converter{} }")

	// Generate converter interface.
	file.P("type Converter interface {")
	file.P("Name() string")
	for _, method := range service.Methods {
		publicIn := method.Input
		publicOut := method.Output

		isDeprecated := deprecatedMethod(method)
		if isDeprecated || serviceType == LatestPublicService {
			if err := generateConverterToPrivateIface(file, publicIn, publicOut, isDeprecated, private); err != nil {
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
			if err := generateConverterToPrivateIface(file, message, message, false, private); err != nil {
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
			if err := generateConverterToPrivateIface(file, enum, enum, false, private); err != nil {
				return err
			}
			continue
		}

		if err := generateConverterToNextIface(file, enum, chain); err != nil {
			return err
		}
	}

	for _, message := range service.DeprecatedMessages {
		if err := generateConverterDeprecatedToPrivateIface(file, message, private); err != nil {
			return err
		}
	}

	for _, enum := range service.DeprecatedEnums {
		if err := generateConverterDeprecatedToPrivateIface(file, enum, private); err != nil {
			return err
		}
	}
	file.P("}")

	// Generate converter functions.
	file.P("type converter struct {}")
	file.P("func (c converter) Name() string { return ConverterName }")

	for _, method := range service.Methods {
		publicIn := method.Input
		publicOut := method.Output

		isDeprecated := deprecatedMethod(method)
		if isDeprecated || serviceType == LatestPublicService {
			if err := generateConverterToPrivateFunc(file, publicIn, private); err != nil {
				return err
			}

			if isDeprecated {
				continue
			}

			if err := generateConverterToPublicFromPrivateFunc(file, publicOut, false, service, private); err != nil {
				return err
			}
			continue
		}

		if err := generateConverterToNextFunc(file, publicIn, next); err != nil {
			return err
		}

		if err := generateConverterToPublicFuncFromNext(file, publicOut, service, chain); err != nil {
			return err
		}
	}

	for _, message := range service.Messages {
		_, isInput := inputs[message]
		_, isOutput := outputs[message]
		if isInput || isOutput {
			continue
		}

		if serviceType == LatestPublicService {
			if err := generateConverterToPrivateFunc(file, message, private); err != nil {
				return err
			}

			if err := generateConverterToPublicFromPrivateFunc(file, message, false, service, private); err != nil {
				return err
			}
			continue
		}

		if err := generateConverterToNextFunc(file, message, next); err != nil {
			return err
		}

		if err := generateConverterToPublicFuncFromNext(file, message, service, chain); err != nil {
			return err
		}
	}

	for _, enum := range service.Enums {
		if serviceType == LatestPublicService {
			if err := generateConverterToPrivateEnum(file, enum, private); err != nil {
				return err
			}

			if err := generateConverterToPublicEnum(file, enum, false, "privatepb", private); err != nil {
				return err
			}
			continue
		}

		if err := generateConverterToNextEnum(file, enum, next); err != nil {
			return err
		}

		if err := generateConverterToPublicEnum(file, enum, false, "nextpb", next); err != nil {
			return err
		}
	}

	for _, message := range service.DeprecatedMessages {
		if err := generateConverterToPublicFromPrivateFunc(file, message, true, service, private); err != nil {
			return err
		}
	}

	for _, enum := range service.DeprecatedEnums {
		if err := generateConverterToDeprecatedPublicEnum(file, enum, chain); err != nil {
			return err
		}
	}

	return nil
}

func generateConverterToNextEnum(file *protogen.GeneratedFile, enum *protogen.Enum, next *Service) error {
	return generateConverterToDstEnum(file, enum, "Next", "publicpb", "nextpb", next)
}

func generateConverterToPrivateEnum(file *protogen.GeneratedFile, enum *protogen.Enum, private *Service) error {
	return generateConverterToDstEnum(file, enum, "Private", "publicpb", "privatepb", private)
}

func generateConverterToPublicEnum(file *protogen.GeneratedFile, enum *protogen.Enum, isDeprecated bool, nextPackageName string, next *Service) error {
	nextEnum, err := findNextEnum(enum, next)
	if err != nil {
		return err
	}

	inEnumName := enum.GoIdent.GoName
	nextEnumName := nextEnum.GoIdent.GoName
	var deprecatedPrefix string
	if isDeprecated {
		deprecatedPrefix = "Deprecated"
	}

	file.P("func (c converter) To", deprecatedPrefix, "Public", inEnumName, "(in ", nextPackageName, ".", nextEnumName, ") (publicpb.", inEnumName, ", error) {")
	file.P("switch in {")
	for _, value := range enum.Values {
		receiveValues, err := findReceiveEnumValues(value, nextEnum)
		if err != nil {
			return err
		}

		nextEnumValue, err := findNextEnumValue(value, nextEnum)
		if err != nil {
			return err
		}

		receiveValues = append(receiveValues, nextEnumValue)
		valueName := value.GoIdent.GoName

		declared := make(map[*protogen.EnumValue]struct{})
		for _, receiveValue := range receiveValues {
			if _, ok := declared[receiveValue]; ok {
				continue
			}
			declared[receiveValue] = struct{}{}
			receiveValueName := receiveValue.GoIdent.GoName
			file.P("case ", nextPackageName, ".", receiveValueName, ":")
			file.P("return publicpb.", valueName, ", nil")
		}
	}
	file.P("}")

	defaultValue := enum.Values[0].GoIdent.GoName
	file.P("return publicpb.", defaultValue, `,fmt.Errorf("%q is not a supported value for this service version", in)`)
	file.P("}")

	return nil
}

func generateConverterToDeprecatedPublicEnum(file *protogen.GeneratedFile, enum *protogen.Enum, chain []*Service) error {
	privateEnum, err := findPrivateEnum(enum, chain)
	if err != nil {
		return err
	}

	inEnumName := enum.GoIdent.GoName
	privateEnumName := privateEnum.GoIdent.GoName

	file.P("func (c converter) ToDeprecatedPublic", inEnumName, "(in privatepb.", privateEnumName, ") (publicpb.", inEnumName, ", error) {")
	file.P("switch in {")
	for _, value := range enum.Values {
		receiveValues, err := findPrivateReceiveEnumValues(value, enum, chain)
		if err != nil {
			return err
		}

		privateEnumValue, err := findPrivateEnumValue(value, enum, chain)
		if err != nil {
			return err
		}

		receiveValues = append(receiveValues, privateEnumValue)
		valueName := value.GoIdent.GoName

		declared := make(map[*protogen.EnumValue]struct{})
		for _, receiveValue := range receiveValues {
			if _, ok := declared[receiveValue]; ok {
				continue
			}
			declared[receiveValue] = struct{}{}
			receiveValueName := receiveValue.GoIdent.GoName
			file.P("case privatepb.", receiveValueName, ":")
			file.P("return publicpb.", valueName, ", nil")
		}
	}
	file.P("}")

	defaultValue := enum.Values[0].GoIdent.GoName
	file.P("return publicpb.", defaultValue, `,fmt.Errorf("%q is not a supported value for this service version", in)`)
	file.P("}")

	return nil
}

func generateConverterToDstEnum(file *protogen.GeneratedFile, enum *protogen.Enum, dst, inPackageName, nextPackageName string, next *Service) error {
	nextEnum, err := findNextEnum(enum, next)
	if err != nil {
		return err
	}

	inEnumName := enum.GoIdent.GoName
	nextEnumName := nextEnum.GoIdent.GoName

	file.P("func (c converter) To", dst, nextEnumName, "(in ", inPackageName, ".", inEnumName, ") ", nextPackageName, ".", nextEnumName, "{")
	file.P("switch in {")
	for _, value := range enum.Values {
		nextEnumValue, err := findNextEnumValue(value, nextEnum)
		if err != nil {
			return err
		}

		valueName := value.GoIdent.GoName
		nextValueName := nextEnumValue.GoIdent.GoName

		file.P("case ", inPackageName, ".", valueName, ":")
		file.P("return ", nextPackageName, ".", nextValueName)
	}
	file.P("}")
	defaultNextValue := nextEnum.Values[0].GoIdent.GoName
	file.P("return ", nextPackageName, ".", defaultNextValue)
	file.P("}")

	return nil
}

func generateConverterToPrivateFunc(file *protogen.GeneratedFile, publicIn *protogen.Message, private *Service) error {
	return generateConverterMessageFunc(file, "Private", "privatepb", publicIn, private)
}

func generateConverterToNextFunc(file *protogen.GeneratedFile, publicIn *protogen.Message, next *Service) error {
	return generateConverterMessageFunc(file, "Next", "nextpb", publicIn, next)
}

func generateConverterMessageFunc(file *protogen.GeneratedFile, dst, packageName string, publicIn *protogen.Message, next *Service) error {
	nextIn, err := findNextMessage(publicIn, next)
	if err != nil {
		return err
	}

	publicInName := publicIn.GoIdent.GoName
	nextInName := nextIn.GoIdent.GoName

	file.P("func (c converter) To", dst, nextInName, "(in *publicpb.", publicInName, ") *", packageName, ".", nextInName, " {")
	file.P("var out ", packageName, ".", nextInName)
	for _, field := range publicIn.Fields {
		if deprecatedField(field) {
			continue
		}

		if field.Oneof != nil {
			continue
		}

		nextField, err := findNextField(field, nextIn)
		if err != nil {
			return fmt.Errorf("failed to generate converter function for %s: %w", publicInName, err)
		}

		publicFieldName := field.GoName
		nextFieldName := nextField.GoName

		if fieldMatch(field, nextField) {
			file.P("out.", nextFieldName, "= in.", publicFieldName)
		} else if nextField.Message != nil || nextField.Enum != nil {
			var name string
			if nextField.Message != nil {
				name = nextField.Message.GoIdent.GoName
			} else {
				name = nextField.Enum.GoIdent.GoName
			}
			file.P("out.", nextFieldName, "= c.To", dst, name, "(in.", publicFieldName, ")")
		}
	}

	for _, oneof := range publicIn.Oneofs {
		nextOneof, err := findNextOneof(oneof, nextIn)
		if err != nil {
			return err
		}

		fieldName := oneof.GoName
		nextFieldName := nextOneof.GoName

		file.P("switch in.", fieldName, ".(type) {")
		for _, field := range oneof.Fields {
			nextField, err := findNextOneofField(field, nextOneof)
			if err != nil {
				return err
			}

			typeName := field.GoIdent.GoName
			nextTypeName := nextField.GoIdent.GoName

			messageName := field.GoName
			nextMessageName := nextField.GoName

			file.P("case *publicpb.", typeName, ":")
			file.P("out.", nextFieldName, "= &", packageName, ".", nextTypeName, "{")
			file.P(nextMessageName, ": c.To", dst, nextMessageName, "(in.Get", messageName, "()),")
			file.P("}")
		}
		file.P("}")
	}
	file.P("return &out")
	file.P("}")

	return nil
}

func generateConverterToPublicFromPrivateFunc(file *protogen.GeneratedFile, publicIn *protogen.Message, isDeprecated bool, service, private *Service) error {
	privateIn, err := findNextMessage(publicIn, private)
	if err != nil {
		return err
	}

	publicInName := publicIn.GoIdent.GoName
	privateInName := privateIn.GoIdent.GoName

	if isDeprecated {
		file.P("func (c converter) ToDeprecatedPublic", publicInName, "(in *privatepb.", privateInName, ") (*publicpb.", publicInName, ", error) {")
	} else {
		file.P("func (c converter) ToPublic", publicInName, "(in *privatepb.", privateInName, ") (*publicpb.", publicInName, ", error) {")
	}
	file.P("var required validation.Errors")
	for _, field := range publicIn.Fields {
		if receiveRequired(field) {
			privateField, err := findNextField(field, privateIn)
			if err != nil {
				return fmt.Errorf("failed to generate converter function for %s: %w", publicInName, err)
			}

			privateFieldName := privateField.GoName
			file.P(`required["`, privateFieldName, `"] = validation.Validate(in.`, privateFieldName, `, validation.Required)`)
		}
	}

	file.P("if err := required.Filter(); err != nil { return nil, err }")
	file.P("var out publicpb.", publicInName)
	file.P("var err error")

	for _, field := range publicIn.Fields {
		if field.Oneof != nil {
			// TODO: Add oneof support
			continue
		}

		if err := generateConverterFieldToPublicFromPrivate(file, "in", field, privateIn, isDeprecated, service); err != nil {
			return err
		}
	}
	file.P("return &out, err")
	file.P("}")

	return nil
}

func generateConverterFieldToPublicFromPrivate(file *protogen.GeneratedFile, privateVarName string, field *protogen.Field, privateIn *protogen.Message, isDeprecated bool, service *Service) error {
	publicFieldName := field.GoName
	privateField, err := findNextField(field, privateIn)
	if err != nil {
		return fmt.Errorf("failed to generate converter field for %s: %w", publicFieldName, err)
	}

	privateFieldName := privateField.GoName

	var deprecatedPrefix string
	if isDeprecated {
		deprecatedPrefix = "Deprecated"
	}

	if fieldMatch(field, privateField) {
		file.P("out.", publicFieldName, "=", privateVarName, ".", privateFieldName)
	} else if field.Message != nil || field.Enum != nil {
		var name string
		if field.Message != nil {
			name = field.Message.GoIdent.GoName
		} else {
			name = field.Enum.GoIdent.GoName
		}

		if field.Desc.IsList() {
			file.P("for _, item := range ", privateVarName, ".", privateFieldName, "{")
			file.P("conv, err := c.To", deprecatedPrefix, "Public", name, "(item)")
			file.P("if err != nil { return nil, err }")
			file.P("out.", publicFieldName, "= append(out.", publicFieldName, ", conv)")
			file.P("}")
		} else {
			file.P("out.", publicFieldName, ", err = c.To", deprecatedPrefix, "Public", name, "(", privateVarName, ".", privateFieldName, ")")
			file.P("if err != nil { return nil, err }")
		}
	}

	return nil
}

func generateConverterFieldToPublicFromNext(file *protogen.GeneratedFile, field *protogen.Field, publicIn *protogen.Message, chain []*Service) error {
	next := chain[0]

	nextIn, err := findNextMessage(publicIn, next)
	if err != nil {
		return err
	}

	nextField, err := findNextField(field, nextIn)
	if err != nil {
		return err
	}

	publicFieldName := field.GoName
	nextFieldName := nextField.GoName

	if fieldMatch(field, nextField) {
		file.P("out.", publicFieldName, "= nextIn.", nextFieldName)
	} else if field.Message != nil || field.Enum != nil {
		var name, privateRef string
		if field.Message != nil {
			name = field.Message.GoIdent.GoName

			privateField, err := findPrivateField(field, publicIn, chain)
			if err != nil {
				return err
			}

			privateFieldName := privateField.GoName
			privateRef = fmt.Sprintf(", privateIn.%s", privateFieldName)
		} else {
			name = field.Enum.GoIdent.GoName
		}

		file.P("out.", publicFieldName, ", err = c.ToPublic", name, "(nextIn.", nextFieldName, privateRef, ")")
		file.P("if err != nil { return nil, err }")
	}

	return nil
}

func generateConverterToPublicFuncFromNext(file *protogen.GeneratedFile, publicIn *protogen.Message, service *Service, chain []*Service) error {
	next := chain[0]

	nextIn, err := findNextMessage(publicIn, next)
	if err != nil {
		return err
	}

	privateIn, err := findPrivateMessage(publicIn, chain)
	if err != nil {
		return err
	}

	publicInName := publicIn.GoIdent.GoName
	privateInName := privateIn.GoIdent.GoName
	nextInName := nextIn.GoIdent.GoName

	file.P("func (c converter) ToPublic", publicInName, "(nextIn *nextpb.", nextInName, ", privateIn *privatepb.", privateInName, ") (*publicpb.", publicInName, ", error) {")
	file.P("required := validation.Errors{}")
	for _, field := range publicIn.Fields {
		if receiveRequired(field) {
			privateField, err := findNextField(field, privateIn)
			if err != nil {
				return fmt.Errorf("failed to generate converter function for %s: %w", publicInName, err)
			}

			varName := "nextIn"
			if deprecatedField(field) {
				varName = "privateIn"
			}

			privateFieldName := privateField.GoName
			file.P(`required["`, privateFieldName, `"] = validation.Validate(`, varName, `.`, privateFieldName, `, validation.Required)`)
		}
	}

	file.P("if err := required.Filter(); err != nil { return nil, err }")
	file.P("var out publicpb.", publicInName)
	file.P("var err error")

	for _, field := range publicIn.Fields {
		if field.Oneof != nil {
			continue
		}

		if deprecatedField(field) {
			// TODO: Investigate why the isDeprecated parameter is false
			if err := generateConverterFieldToPublicFromPrivate(file, "privateIn", field, privateIn, false, service); err != nil {
				return err
			}
			continue
		}

		if err := generateConverterFieldToPublicFromNext(file, field, publicIn, chain); err != nil {
			return err
		}
	}

	for _, oneof := range publicIn.Oneofs {
		if deprecatedOneof(oneof) {
			continue
		}

		// convert to public from next
		nextOneof, err := findNextOneof(oneof, nextIn)
		if err != nil {
			return err
		}

		fieldName := oneof.GoName

		file.P("switch nextIn.", fieldName, ".(type) {")
		for _, field := range oneof.Fields {
			nextField, err := findNextOneofField(field, nextOneof)
			if err != nil {
				return err
			}

			privateField, err := findPrivateOneofField(field, oneof, publicIn, chain)
			if err != nil {
				return err
			}

			typeName := field.GoIdent.GoName
			nextTypeName := nextField.GoIdent.GoName

			messageName := field.GoName
			nextMessageName := nextField.GoName
			privateMessageName := privateField.GoName
			dst := "Public"

			file.P("case *nextpb.", nextTypeName, ":")
			file.P("var value publicpb.", typeName)
			file.P("value.", messageName, ", err = ", "c.To", dst, messageName, "(nextIn.Get", nextMessageName, "(), privateIn.Get", privateMessageName, "())")
			file.P("out.", fieldName, "= &value")
		}
		file.P("}")
	}

	file.P("return &out, err")
	file.P("}")

	return nil
}

func generateConverterToPrivateIface(file *protogen.GeneratedFile, publicIn, publicOut interface{}, isDeprecated bool, private *Service) error {
	publicInName, privateInName, publicInPointer, err := generateConverterMessageNamesAndPointers(publicIn, private)
	if err != nil {
		return err
	}

	publicOutName, privateOutName, publicOutPointer, err := generateConverterMessageNamesAndPointers(publicOut, private)
	if err != nil {
		return err
	}

	file.P("ToPrivate", privateInName, "(", publicInPointer, "publicpb.", publicInName, ") ", publicOutPointer, "privatepb.", privateInName)
	if !isDeprecated {
		file.P("ToPublic", publicOutName, "(", publicOutPointer, "privatepb.", privateOutName, ") (", publicOutPointer, "publicpb.", publicOutName, ", error)")
	}

	return nil
}

func generateConverterDeprecatedToPrivateIface(file *protogen.GeneratedFile, publicOut interface{}, private *Service) error {
	publicOutName, privateOutName, pointer, err := generateConverterMessageNamesAndPointers(publicOut, private)
	if err != nil {
		return err
	}

	file.P("ToDeprecatedPublic", publicOutName, "(", pointer, "privatepb.", privateOutName, ") (", pointer, "publicpb.", publicOutName, ", error)")
	return nil
}

func generateConverterMessageNamesAndPointers(pub interface{}, private *Service) (string, string, string, error) {
	var publicName, privateName, pointer string
	switch pub.(type) {
	case *protogen.Message:
		value := pub.(*protogen.Message)
		privateValue, err := findNextMessage(value, private)

		if err != nil {
			return "", "", "", err
		}

		publicName = value.GoIdent.GoName
		privateName = privateValue.GoIdent.GoName
		pointer = "*"
	case *protogen.Enum:
		value := pub.(*protogen.Enum)
		privateValue, err := findNextEnum(value, private)

		if err != nil {
			return "", "", "", err
		}

		publicName = value.GoIdent.GoName
		privateName = privateValue.GoIdent.GoName
	}

	return publicName, privateName, pointer, nil
}

func generateConverterToNextIface(file *protogen.GeneratedFile, v interface{}, chain []*Service) error {
	next := chain[0]
	var publicName, privateName, nextName, pointer, privateRef string
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
		privateRef = fmt.Sprintf(", *privatepb.%s", privateName)
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
	file.P("ToPublic", publicName, "(", pointer, "nextpb.", nextName, privateRef, ") (", pointer, "publicpb.", publicName, ", error)")

	return nil
}

func commonImports(imports ...protogen.GoIdent) []protogen.GoIdent {
	return append([]protogen.GoIdent{
		protogen.GoImportPath("fmt").Ident("fmt"),
		protogen.GoImportPath("context").Ident("context"),
		protogen.GoImportPath("github.com/go-ozzo/ozzo-validation/v4").Ident("validation"),
		protogen.GoImportPath("github.com/go-ozzo/ozzo-validation/v4/is").Ident("is"),
		protogen.GoImportPath("google.golang.org/grpc/codes").Ident("codes"),
		protogen.GoImportPath("google.golang.org/grpc/status").Ident("status"),
	}, imports...)
}

func generateImportUsage(file *protogen.GeneratedFile, refs ...string) {
	refs = append([]string{"is.Int", "validation.Validate", "fmt.Errorf"}, refs...)

	for _, ref := range refs {
		file.P("var _ =", ref)
	}
}

func generateServiceStruct(file *protogen.GeneratedFile, refs ...string) {
	file.P("type Service struct {")
	file.P("Validator")
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
	file.P("if err := s.Validate", inName, "(in); err != nil {")
	file.P(`return nil, status.Errorf(codes.InvalidArgument, "%s", err)`)
	file.P("}")
	file.P("out, _, err := s.", methodName, "Impl(ctx, in)")
	file.P("return out, err")
	file.P("}")
}

func generatePrivateServiceMethod(file *protogen.GeneratedFile, methodName, inName, outName string) {
	file.P("func (s *Service) ", methodName, "(ctx context.Context, in *privatepb.", inName, ") (*privatepb.", outName, ", error) {")
	file.P("if err := s.Validate", inName, "(in); err != nil {")
	file.P(`return nil, status.Errorf(codes.InvalidArgument, "%s", err)`)
	file.P("}")
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
	var deprecatedPrefix string
	if deprecatedMethod(method) {
		deprecatedPrefix = "Deprecated"
	}

	file.P("func (s *Service) ", publicMethodName, "Impl(ctx context.Context, in *publicpb.", publicInName, ", mutators ...private.", privateInName, "Mutator) (*publicpb.", publicOutName, ", *privatepb.", privateOutName, ", error) {")
	file.P("privateIn := s.ToPrivate", privateInName, "(in)")
	file.P("private.Apply", privateInName, "Mutators(privateIn, mutators)")
	file.P("privateOut, err := s.Private.", privateMethodName, "(ctx, privateIn)")
	file.P("if err != nil { return nil, nil, err }")
	file.P("out, err := s.To", deprecatedPrefix, "Public", publicOutName, "(privateOut)")
	file.P("if err != nil {")
	file.P(`return nil, nil, status.Errorf(codes.FailedPrecondition, "%s", err)`)
	file.P("}")
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

	for _, field := range method.Input.Fields {
		if deprecatedField(field) {
			privateMessage, err := findPrivateMessage(method.Input, chain)
			if err != nil {
				return err
			}

			privateField, err := findNextField(field, privateMessage)
			if err != nil {
				return err
			}

			privateMessageName := privateMessage.GoIdent.GoName
			privateFieldName := privateField.GoName
			file.P("mutators = append(mutators, private.Set", privateMessageName, "_", privateFieldName, "(in.", field.GoName, "))")
		}
	}

	file.P("nextIn := s.ToNext", nextInName, "(in)")
	file.P("nextOut, privateOut, err := s.Next.", nextMethodName, "Impl(ctx, nextIn, mutators...)")
	file.P("if err != nil { return nil, nil, err }")
	file.P("out, err := s.ToPublic", publicOutName, "(nextOut, privateOut)")
	file.P("if err != nil {")
	file.P(`return nil, nil, status.Errorf(codes.FailedPrecondition, "%s", err)`)
	file.P("}")
	file.P("return out, privateOut, nil")
	file.P("}")

	return nil
}

func generateServiceValidators(file *protogen.GeneratedFile, packageName string, service *Service) error {
	file.P(`const ValidatorName = "`, service.Desc.FullName(), `.Validator"`)
	file.P("func NewValidator() Validator { return validator{} }")

	file.P("type Validator interface {")
	file.P("Name() string")
	for _, message := range service.Messages {
		if _, ok := outputs[message]; ok {
			continue
		}
		messageName := message.GoIdent.GoName
		logger.Printf("package=%s at=generate-validator-interface message=%s", service.GoPackageName, messageName)
		file.P("Validate", messageName, "(*", packageName, ".", messageName, ") error")

		for _, oneof := range message.Oneofs {
			for _, field := range oneof.Fields {
				fieldName := field.GoIdent.GoName
				logger.Printf("package=%s at=generate-validator-interface oneof=%s", service.GoPackageName, fieldName)
				file.P("Validate", fieldName, "(*", packageName, ".", fieldName, ") error")
			}
		}
	}
	file.P("}")

	file.P("type validator struct {}")
	file.P("func (v validator) Name() string { return ValidatorName }")
	for _, message := range service.Messages {
		if _, ok := outputs[message]; ok {
			continue
		}

		messageName := message.GoIdent.GoName
		logger.Printf("package=%s at=generate-validator-function message=%s", service.GoPackageName, messageName)
		file.P("func(v validator) Validate", messageName, "(in *", packageName, ".", messageName, ") error {")
		if _, ok := inputs[message]; !ok {
			file.P("if in == nil { return nil }")
		}

		file.P("err := validation.ValidateStruct(in,")
		for _, field := range message.Fields {
			if field.Oneof != nil {
				continue
			}

			if field.Message != nil && !isServiceMessage(service, field.Message) {
				continue
			}

			fieldName := field.GoName
			file.P("validation.Field(&in.", fieldName, ",")

			if requiredField(field) {
				file.P("validation.Required,")
			}

			isName, err := is(field)
			if err != nil {
				return err
			}
			switch strings.ToLower(isName) {
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

			if field.Message != nil {
				messageName := field.Message.GoIdent.GoName
				file.P("validation.By(func(interface{}) error { return v.Validate", messageName, "(in.", fieldName, ") }),")
			}

			file.P("),")
		}

		for _, oneof := range message.Oneofs {
			oneofName := oneof.GoName
			for _, field := range oneof.Fields {
				messageName := field.GoName
				fieldName := field.GoIdent.GoName
				ref := fmt.Sprintf("*%s.%s", packageName, fieldName)
				file.P("validation.Field(&in.", oneofName, ",")
				file.P("validation.When(in.Get", messageName, "() != nil, validation.By(func(val interface{}) error { return v.Validate", fieldName, "(val.(", ref, ")) })),")
				file.P("),")
			}
		}
		file.P(")")
		file.P("if err != nil { return err }")
		file.P("return nil")
		file.P("}")

		for _, oneof := range message.Oneofs {
			for _, field := range oneof.Fields {
				typeName := field.GoIdent.GoName
				fieldName := field.GoName
				logger.Printf("package=%s at=generate-validator-function oneof=%s", service.GoPackageName, typeName)
				file.P("func(v validator) Validate", typeName, "(in *", packageName, ".", typeName, ") error {")
				file.P("if in == nil { return nil }")
				file.P("err := validation.ValidateStruct(in,")
				messageName := field.Message.GoIdent.GoName
				file.P("validation.Field(&in.", messageName, ",")

				if requiredOneof(oneof) {
					file.P("validation.Required,")
				}

				file.P("validation.By(func(interface{}) error { return v.Validate", messageName, "(in.", fieldName, ") }),")

				file.P("),")
				file.P(")")
				file.P("if err != nil { return err }")
				file.P("return nil")
				file.P("}")

			}
		}
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

func fieldMatch(a, b *protogen.Field) bool {
	if a.Desc.Kind() != b.Desc.Kind() {
		return false
	}

	if a.Desc.Kind() == protoreflect.MessageKind {
		return a.Message.Desc.FullName() == b.Message.Desc.FullName()
	}

	if a.Desc.Kind() == protoreflect.EnumKind {
		return a.Enum.Desc.FullName() == b.Enum.Desc.FullName()
	}

	return true
}

func isServiceMessage(service *Service, message *protogen.Message) bool {
	return service.GoImportPath == message.GoIdent.GoImportPath
}
