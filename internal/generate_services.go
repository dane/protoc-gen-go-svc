package internal

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/dane/protoc-gen-go-svc/internal/generators"
)

func generateServiceRegister(file *protogen.GeneratedFile, chain []*generators.Service) error {
	imports := []protogen.GoIdent{
		protogen.GoImportPath("google.golang.org/grpc").Ident("grpc"),
	}

	for _, service := range chain {
		packageName := string(service.GoPackageName)
		path := service.GoImportPath.Ident(fmt.Sprintf("%spb", packageName))
		imports = append(imports, path)
		imports = append(imports, service.GoServiceImportPath.Ident(packageName))
	}

	private := chain[len(chain)-1]
	services := chain[:len(chain)-1]
	sort.Sort(sort.Reverse(byPackageName(services)))

	return generators.NewServiceRegister(imports, services, private).Generate(file)
}

func generatePrivateService(file *protogen.GeneratedFile, service *generators.Service) error {
	imports := []protogen.GoIdent{service.GoImportPath.Ident("privatepb")}
	fields := []string{fmt.Sprintf("Impl privatepb.%sServer", service.GoName)}
	g := generators.NewServiceStruct(imports, service, fields)
	if err := g.Generate(file); err != nil {
		return err
	}

	if err := generateServiceMethods(file, service, "privatepb", true); err != nil {
		return err
	}

	if err := generateServiceValidators(file, "privatepb", service); err != nil {
		return err
	}

	if err := generateMutators(file, service); err != nil {
		return err
	}

	return nil
}

func generateLatestPublicService(file *protogen.GeneratedFile, service *generators.Service, chain []*generators.Service) error {
	private := chain[len(chain)-1]

	imports := []protogen.GoIdent{
		service.GoImportPath.Ident("publicpb"),
		private.GoImportPath.Ident("privatepb"),
		private.GoServiceImportPath.Ident("private"),
	}

	fields := []string{
		"Converter",
		"Private *private.Service",
		fmt.Sprintf("publicpb.%sServer", service.GoName),
	}

	g := generators.NewServiceStruct(imports, service, fields)
	if err := g.Generate(file); err != nil {
		return err
	}

	if err := generateServiceMethods(file, service, "publicpb", false); err != nil {
		return err
	}

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

func generatePublicService(file *protogen.GeneratedFile, service *generators.Service, chain []*generators.Service) error {
	next := chain[0]
	private := chain[len(chain)-1]

	imports := []protogen.GoIdent{
		service.GoImportPath.Ident("publicpb"),
		next.GoImportPath.Ident("nextpb"),
		private.GoImportPath.Ident("privatepb"),
		private.GoServiceImportPath.Ident("private"),
		next.GoServiceImportPath.Ident("next"),
	}

	fields := []string{
		"Converter",
		"Private *private.Service",
		"Next *next.Service",
		fmt.Sprintf("publicpb.%sServer", service.GoName),
	}

	g := generators.NewServiceStruct(imports, service, fields)
	if err := g.Generate(file); err != nil {
		return err
	}

	logger.Printf("package=%s at=generate-validators", service.GoPackageName)
	if err := generateServiceValidators(file, "publicpb", service); err != nil {
		return err
	}

	logger.Printf("package=%s at=generate-converters", service.GoPackageName)
	if err := generateConverters(file, service, chain, PublicService); err != nil {
		return err
	}

	logger.Printf("package=%s at=generate-methods", service.GoPackageName)
	if err := generateServiceMethods(file, service, "publicpb", false); err != nil {
		return err
	}

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

func generateConverters(file *protogen.GeneratedFile, service *generators.Service, chain []*generators.Service, serviceType ServiceType) error {
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
			logger.Printf("package=%s at=generate-converter-interface message=%s deprecated=%v", service.GoPackageName, publicIn.GoIdent.GoName, isDeprecated)
			logger.Printf("package=%s at=generate-converter-interface message=%s deprecated=%v", service.GoPackageName, publicOut.GoIdent.GoName, isDeprecated)

			if err := generateConverterToPrivateIface(file, publicIn, publicOut, isDeprecated, private); err != nil {
				return err
			}

			continue
		}

		logger.Printf("package=%s at=generate-converter-interface message=%s", service.GoPackageName, publicIn.GoIdent.GoName)
		logger.Printf("package=%s at=generate-converter-interface message=%s", service.GoPackageName, publicOut.GoIdent.GoName)

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

		logger.Printf("package=%s at=generate-converter-interface message=%s", service.GoPackageName, message.GoIdent.GoName)

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
		logger.Printf("package=%s at=generate-converter-interface enum=%s", service.GoPackageName, enum.GoIdent.GoName)
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
		logger.Printf("package=%s at=generate-converter-interface message=%s deprecated=%v", service.GoPackageName, message.GoIdent.GoName, true)
		if err := generateConverterDeprecatedToPrivateIface(file, message, private); err != nil {
			return err
		}
	}

	for _, enum := range service.DeprecatedEnums {
		logger.Printf("package=%s at=generate-converter-interface enum=%s deprecated=%v", service.GoPackageName, enum.GoIdent.GoName, true)
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
			logger.Printf("package=%s at=generate-converter-function message=%s deprecated=%v", service.GoPackageName, publicIn.GoIdent.GoName, isDeprecated)
			logger.Printf("package=%s at=generate-converter-function message=%s deprecated=%v", service.GoPackageName, publicOut.GoIdent.GoName, isDeprecated)

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

		logger.Printf("package=%s at=generate-converter-function message=%s", service.GoPackageName, publicIn.GoIdent.GoName)
		if err := generateConverterToNextFunc(file, publicIn, next); err != nil {
			return err
		}

		logger.Printf("package=%s at=generate-converter-function message=%s", service.GoPackageName, publicOut.GoIdent.GoName)
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

		logger.Printf("package=%s at=generate-converter-function message=%s", service.GoPackageName, message.GoIdent.GoName)

		if serviceType == LatestPublicService {
			logger.Printf("package=%s at=generate-converter-function message=%s from=%s", service.GoPackageName, message.GoIdent.GoName, private.GoPackageName)
			if err := generateConverterToPrivateFunc(file, message, private); err != nil {
				return err
			}

			if err := generateConverterToPublicFromPrivateFunc(file, message, false, service, private); err != nil {
				return err
			}
			continue
		}

		logger.Printf("package=%s at=generate-converter-function message=%s from=%s dst=%s", service.GoPackageName, message.GoIdent.GoName, next.GoPackageName, "ToNext")
		if err := generateConverterToNextFunc(file, message, next); err != nil {
			return err
		}

		logger.Printf("package=%s at=generate-converter-function message=%s from=%s dst=%s", service.GoPackageName, message.GoIdent.GoName, next.GoPackageName, "ToPublic")
		if err := generateConverterToPublicFuncFromNext(file, message, service, chain); err != nil {
			return err
		}
	}

	for _, enum := range service.Enums {
		logger.Printf("package=%s at=generate-converter-function enum=%s", service.GoPackageName, enum.GoIdent.GoName)
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
		logger.Printf("package=%s at=generate-converter-function message=%s deprecated=%v", service.GoPackageName, message.GoIdent.GoName, true)
		if err := generateConverterToPublicFromPrivateFunc(file, message, true, service, private); err != nil {
			return err
		}
	}

	for _, enum := range service.DeprecatedEnums {
		logger.Printf("package=%s at=generate-converter-function enum=%s deprecated=%v", service.GoPackageName, enum.GoIdent.GoName, true)
		if err := generateConverterToDeprecatedPublicEnum(file, enum, chain); err != nil {
			return err
		}
	}

	return nil
}

func generateConverterToNextEnum(file *protogen.GeneratedFile, enum *protogen.Enum, next *generators.Service) error {
	return generateConverterToDstEnum(file, enum, "Next", "publicpb", "nextpb", next)
}

func generateConverterToPrivateEnum(file *protogen.GeneratedFile, enum *protogen.Enum, private *generators.Service) error {
	return generateConverterToDstEnum(file, enum, "Private", "publicpb", "privatepb", private)
}

func generateConverterToPublicEnum(file *protogen.GeneratedFile, enum *protogen.Enum, isDeprecated bool, nextPackageName string, next *generators.Service) error {
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

func generateConverterToDeprecatedPublicEnum(file *protogen.GeneratedFile, enum *protogen.Enum, chain []*generators.Service) error {
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

func generateConverterToDstEnum(file *protogen.GeneratedFile, enum *protogen.Enum, dst, inPackageName, nextPackageName string, next *generators.Service) error {
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

func generateConverterToPrivateFunc(file *protogen.GeneratedFile, publicIn *protogen.Message, private *generators.Service) error {
	return generateConverterMessageFunc(file, "Private", "privatepb", publicIn, private)
}

func generateConverterToNextFunc(file *protogen.GeneratedFile, publicIn *protogen.Message, next *generators.Service) error {
	return generateConverterMessageFunc(file, "Next", "nextpb", publicIn, next)
}

func generateConverterMessageFunc(file *protogen.GeneratedFile, dst, packageName string, publicIn *protogen.Message, next *generators.Service) error {
	nextIn, err := findNextMessage(publicIn, next)
	if err != nil {
		return err
	}

	publicInName := publicIn.GoIdent.GoName
	nextInName := nextIn.GoIdent.GoName

	file.P("func (c converter) To", dst, nextInName, "(in *publicpb.", publicInName, ") *", packageName, ".", nextInName, " {")
	file.P("if in == nil { return nil }")
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

func generateConverterToPublicFromPrivateFunc(file *protogen.GeneratedFile, publicIn *protogen.Message, isDeprecated bool, service, private *generators.Service) error {
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
	file.P("if in == nil { return nil, nil }")
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
			continue
		}

		if err := generateConverterFieldToPublicFromPrivate(file, "in", field, privateIn, isDeprecated, service); err != nil {
			return err
		}
	}

	var deprecatedPrefix string
	if isDeprecated {
		deprecatedPrefix = "Deprecated"
	}

	for _, oneof := range publicIn.Oneofs {
		nextOneof, err := findNextOneof(oneof, privateIn)
		if err != nil {
			return err
		}

		fieldName := oneof.GoName

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
			dst := "Public"

			file.P("case *privatepb.", nextTypeName, ":")
			file.P("var value publicpb.", typeName)
			file.P("value.", messageName, ", err = ", "c.To", deprecatedPrefix, dst, messageName, "(in.Get", nextMessageName, "())")
			file.P("if err == nil { out.", fieldName, "= &value }")
		}
		file.P("}")
	}

	file.P("return &out, err")
	file.P("}")

	return nil
}

func generateConverterFieldToPublicFromPrivate(file *protogen.GeneratedFile, privateVarName string, field *protogen.Field, privateIn *protogen.Message, isDeprecated bool, service *generators.Service) error {
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

func generateConverterFieldToPublicFromNext(file *protogen.GeneratedFile, field *protogen.Field, publicIn *protogen.Message, chain []*generators.Service) error {
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

func generateConverterToPublicFuncFromNext(file *protogen.GeneratedFile, publicIn *protogen.Message, service *generators.Service, chain []*generators.Service) error {
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
	file.P("if nextIn == nil || privateIn == nil { return nil, nil }")
	file.P("required := validation.Errors{}")

	logger.Printf("package=%s at=generate-converter-function-validator message=%s", service.GoPackageName, publicInName)
	for _, field := range publicIn.Fields {
		if !receiveRequired(field) {
			continue
		}

		var varName string
		var nextField *protogen.Field
		var err error
		if deprecatedField(field) {
			logger.Printf("package=%s at=generate-converter-function-validator message=%s field=%s deprecated=%v", service.GoPackageName, publicInName, field.GoName, true)
			varName = "privateIn"
			nextField, err = findNextField(field, privateIn)
			if err != nil {
				return err
			}
		} else {
			logger.Printf("package=%s at=generate-converter-function-validator message=%s field=%s deprecated=%v", service.GoPackageName, publicInName, field.GoName, false)
			varName = "nextIn"
			nextField, err = findNextField(field, nextIn)
			if err != nil {
				return err
			}
		}

		fieldName := field.GoName
		nextFieldName := nextField.GoName
		file.P(`required["`, fieldName, `"] = validation.Validate(`, varName, `.`, nextFieldName, `, validation.Required)`)
	}

	file.P("if err := required.Filter(); err != nil { return nil, err }")
	file.P("var out publicpb.", publicInName)
	file.P("var err error")

	logger.Printf("package=%s at=generate-converter-function-assignment message=%s", service.GoPackageName, publicInName)
	for _, field := range publicIn.Fields {
		if field.Oneof != nil {
			continue
		}

		if deprecatedField(field) {
			if err := generateConverterFieldToPublicFromPrivate(file, "privateIn", field, privateIn, true, service); err != nil {
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
			nextOneof, err := findNextOneof(oneof, privateIn)
			if err != nil {
				return err
			}

			fieldName := oneof.GoName

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
				dst := "Public"

				file.P("case *privatepb.", nextTypeName, ":")
				file.P("out.", fieldName, "= &publicpb.", typeName, "{")
				file.P(messageName, ": c.To", dst, messageName, "(in.Get", nextMessageName, "()),")
				file.P("}")
			}
			file.P("}")

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

func generateConverterToPrivateIface(file *protogen.GeneratedFile, publicIn, publicOut interface{}, isDeprecated bool, private *generators.Service) error {
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

func generateConverterDeprecatedToPrivateIface(file *protogen.GeneratedFile, publicOut interface{}, private *generators.Service) error {
	publicOutName, privateOutName, pointer, err := generateConverterMessageNamesAndPointers(publicOut, private)
	if err != nil {
		return err
	}

	file.P("ToDeprecatedPublic", publicOutName, "(", pointer, "privatepb.", privateOutName, ") (", pointer, "publicpb.", publicOutName, ", error)")
	return nil
}

func generateConverterMessageNamesAndPointers(pub interface{}, private *generators.Service) (string, string, string, error) {
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

func generateConverterToNextIface(file *protogen.GeneratedFile, v interface{}, chain []*generators.Service) error {
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

func generateImportUsage(file *protogen.GeneratedFile, refs ...string) {
	refs = append([]string{"is.Int", "validation.Validate", "fmt.Errorf"}, refs...)

	for _, ref := range refs {
		file.P("var _ =", ref)
	}
}

func generateServiceMethods(file *protogen.GeneratedFile, service *generators.Service, packageName string, toPrivate bool) error {
	for _, method := range service.Methods {
		g := generators.NewServiceMethod(
			packageName, method.GoName,
			method.Input.GoIdent.GoName,
			method.Output.GoIdent.GoName,
			toPrivate,
		)

		if err := g.Generate(file); err != nil {
			return err
		}
	}

	return nil
}

func generateServiceMethodToPrivateImpl(file *protogen.GeneratedFile, method *protogen.Method, private *generators.Service) error {
	privateMethod, err := findNextMethod(method, private)
	if err != nil {
		return fmt.Errorf("failed to generate service %s method impl: %w", method.GoName, err)
	}

	privateIn, err := findNextMessage(method.Input, private)
	if err != nil {
		return err
	}

	privateOut, err := findNextMessage(method.Output, private)
	if err != nil {
		return err
	}

	var deprecatedPrefix string
	if deprecatedMethod(method) {
		deprecatedPrefix = "Deprecated"
	}

	g := ServiceMethodImplToPrivateGenerator{
		Prefix: deprecatedPrefix,

		MethodName: method.GoName,
		InputName:  method.Input.GoIdent.GoName,
		OutputName: method.Output.GoIdent.GoName,

		PrivateMethodName: privateMethod.GoName,
		PrivateInputName:  privateIn.GoIdent.GoName,
		PrivateOutputName: privateOut.GoIdent.GoName,
	}

	return g.Generate(file)
}

func generateServiceMethodToNextImpl(file *protogen.GeneratedFile, method *protogen.Method, chain []*generators.Service) error {
	next := chain[0]
	nextMethod, err := findNextMethod(method, next)
	if err != nil {
		return fmt.Errorf("failed to generate service %s method impl: %w", method.GoName, err)
	}

	nextIn, err := findNextMessage(method.Input, next)
	if err != nil {
		return err
	}

	privateIn, err := findPrivateMessage(method.Input, chain)
	if err != nil {
		return err
	}

	privateOut, err := findPrivateMessage(method.Output, chain)
	if err != nil {
		return err
	}

	g := ServiceMethodImplToNextGenerator{
		MethodName:        method.GoName,
		InputName:         method.Input.GoIdent.GoName,
		OutputName:        method.Output.GoIdent.GoName,
		PrivateInputName:  privateIn.GoIdent.GoName,
		PrivateOutputName: privateOut.GoIdent.GoName,
		NextMethodName:    nextMethod.GoName,
		NextInputName:     nextIn.GoIdent.GoName,
	}

	for _, field := range method.Input.Fields {
		if deprecatedField(field) {
			privateField, err := findNextField(field, privateIn)
			if err != nil {
				return err
			}

			g.DeprecatedFields = append(g.DeprecatedFields, DeprecatedField{
				FieldName:        field.GoName,
				PrivateFieldName: privateField.GoName,
			})
		}
	}

	return g.Generate(file)
}

func generateServiceValidators(file *protogen.GeneratedFile, packageName string, service *generators.Service) error {
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

func generateMutators(file *protogen.GeneratedFile, service *generators.Service) error {
	var serviceMutators []ServiceMutatorGenerator
	for _, method := range service.Methods {
		messageName := method.Input.GoIdent.GoName

		var fields []MutatorField
		for _, field := range method.Input.Fields {
			fieldType, err := findFieldType("privatepb", field)
			if err != nil {
				return fmt.Errorf("failed to generate mutator for %s: %w", messageName, err)
			}

			fields = append(fields, MutatorField{
				FieldName: field.GoName,
				FieldType: fieldType,
			})
		}

		serviceMutators = append(serviceMutators, ServiceMutatorGenerator{
			MessageName: messageName,
			Fields:      fields,
		})
	}

	return execute("service_mutators", templateServiceMutators, file, serviceMutators)
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

func isServiceMessage(service *generators.Service, message *protogen.Message) bool {
	return service.GoImportPath == message.GoIdent.GoImportPath
}
