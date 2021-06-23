package internal

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func generateLatestConverters(file *protogen.GeneratedFile, service, private *Service) error {
	file.P(`const ConverterName = "`, service.Service.Desc.FullName(), `.Converter"`)

	// Create Converter interface.
	file.P("func NewConverter() Converter { return converter{} }")

	file.P("type Converter interface {")
	file.P("Name() string")
	for _, method := range service.Methods {

		delegateMethod, err := findMethodDelegate(method, private)
		if err != nil {
			return err
		}

		generateToPrivateConverterIface(file, method.Input.GoIdent, delegateMethod.Input.GoIdent, true)
		generateToPublicConverterIface(file, method.Output.GoIdent, delegateMethod.Output.GoIdent, true)

		// These message do not need to be converted between the latest and
		// private service.
		publicInName := method.Input.GoIdent.GoName
		publicInPath := method.Input.GoIdent.GoImportPath

		privateOutName := delegateMethod.Output.GoIdent.GoName
		privateOutPath := delegateMethod.Output.GoIdent.GoImportPath

		messagesByImportPath[publicInPath][publicInName].Skip = true
		messagesByImportPath[privateOutPath][privateOutName].Skip = true
	}

	// Add messages to interface that are outside of top-level input and output
	// messages.
	for _, message := range messagesByImportPath[service.GoIdent.GoImportPath] {
		if message.Generated || message.Skip {
			continue
		}

		delegateMessage, err := findMessageDelegate(message, private)
		if err != nil {
			return err
		}

		generateToPrivateConverterIface(file, message.GoIdent, delegateMessage.GoIdent, true)
		generateToPublicConverterIface(file, message.GoIdent, delegateMessage.GoIdent, true)
	}

	// Add enums to interface that are outside of top-level input and output
	// messages.
	for _, enum := range enumsByImportPath[service.GoIdent.GoImportPath] {
		if enum.Generated || enum.Skip {
			continue
		}

		delegateEnum, err := findEnumDelegate(enum, private)
		if err != nil {
			return err
		}

		generateToPrivateConverterIface(file, enum.GoIdent, delegateEnum.GoIdent, false)
		generateToPublicConverterIface(file, enum.GoIdent, delegateEnum.GoIdent, false)
	}
	file.P("}")

	file.P("type converter struct {}")

	// Reset generated state of messages.
	for _, message := range messagesByImportPath[service.GoIdent.GoImportPath] {
		message.Generated = false
	}

	for _, enum := range enumsByImportPath[service.GoIdent.GoImportPath] {
		enum.Generated = false
	}

	// Create converter functions.
	file.P("func (c converter) Name() string { return ConverterName }")

	for _, method := range service.Methods {
		delegateMethod, err := findMethodDelegate(method, private)
		if err != nil {
			return err
		}

		if err := generateToPrivateConverterFunc(file, method.Input, delegateMethod.Input); err != nil {
			return err
		}

		if err := generateToPublicConverterFunc(file, method.Output, delegateMethod.Output); err != nil {
			return err
		}

		// These message do not need to be converted between the latest and
		// private service.
		publicInName := method.Input.GoIdent.GoName
		publicInPath := method.Input.GoIdent.GoImportPath

		privateOutName := delegateMethod.Output.GoIdent.GoName
		privateOutPath := delegateMethod.Output.GoIdent.GoImportPath

		messagesByImportPath[publicInPath][publicInName].Skip = true
		messagesByImportPath[privateOutPath][privateOutName].Skip = true
	}

	// Create converter functions that are outside of top-level input and output
	// messages.
	for _, message := range messagesByImportPath[service.GoIdent.GoImportPath] {
		if message.Generated || message.Skip {
			continue
		}

		delegateMessage, err := findMessageDelegate(message, private)
		if err != nil {
			return err
		}

		if err := generateToPrivateConverterFunc(file, message.Message, delegateMessage.Message); err != nil {
			return err
		}

		if err := generateToPublicConverterFunc(file, message.Message, delegateMessage.Message); err != nil {
			return err
		}
	}

	// Create converter functions that are outside of top-level input and output
	// messages.
	for _, enum := range enumsByImportPath[service.GoIdent.GoImportPath] {
		if enum.Generated || enum.Skip {
			continue
		}

		delegateEnum, err := findEnumDelegate(enum, private)
		if err != nil {
			return err
		}

		if err := generateToPrivateConverterEnumFunc(file, enum.Enum, delegateEnum.Enum); err != nil {
			return fmt.Errorf("failed to generate service %q: %w", service.GoPackageName, err)
		}

		if err := generateToPublicConverterEnumFunc(file, enum.Enum, delegateEnum.Enum); err != nil {
			return fmt.Errorf("failed to generate service %q: %w", service.GoPackageName, err)
		}
	}

	return nil
}

func fieldMatch(current, next *protogen.Field) bool {
	switch current.Desc.Kind() {
	case protoreflect.MessageKind:
		if current.Message.Desc.FullName() == next.Message.Desc.FullName() {
			return true
		}
	case protoreflect.EnumKind:
		if current.Enum.Desc.FullName() == next.Enum.Desc.FullName() {
			return true
		}
	default:
		if current.Desc.Kind() == next.Desc.Kind() {
			return true
		}
	}

	return false
}

func generateToPublicConverterIface(file *protogen.GeneratedFile, publicOut, privateOut protogen.GoIdent, isPointer bool) {
	publicOutName := publicOut.GoName
	publicOutPath := publicOut.GoImportPath

	privateOutName := privateOut.GoName

	var ptr string
	if isPointer {
		ptr = "*"
	}

	file.P("ToPublic", publicOutName, "(", ptr, "privatepb.", privateOutName, ") (", ptr, "publicpb.", publicOutName, ", error)")

	if messages, ok := messagesByImportPath[publicOutPath]; ok {
		if m, ok := messages[publicOutName]; ok {
			m.Generated = true
		}
	}

	if enums, ok := enumsByImportPath[publicOutPath]; ok {
		if m, ok := enums[publicOutName]; ok {
			m.Generated = true
		}
	}
}

func generateToPrivateConverterIface(file *protogen.GeneratedFile, publicIn, privateIn protogen.GoIdent, isPointer bool) {
	privateInName := privateIn.GoName
	privateInPath := privateIn.GoImportPath

	publicInName := publicIn.GoName

	var ptr string
	if isPointer {
		ptr = "*"
	}
	file.P("ToPrivate", privateInName, "(", ptr, "publicpb.", publicInName, ") ", ptr, "privatepb.", privateInName)

	if messages, ok := messagesByImportPath[privateInPath]; ok {
		if m, ok := messages[privateInName]; ok {
			m.Generated = true
		}
	}

	if enums, ok := enumsByImportPath[privateInPath]; ok {
		if m, ok := enums[privateInName]; ok {
			m.Generated = true
		}
	}
}

func generateToPrivateConverterFunc(file *protogen.GeneratedFile, publicIn, privateIn *protogen.Message) error {
	publicInName := publicIn.GoIdent.GoName

	privateInName := privateIn.GoIdent.GoName
	privateInPath := privateIn.GoIdent.GoImportPath

	file.P("func (c converter) ToPrivate", privateInName, "(in *publicpb.", publicInName, ") *privatepb.", privateInName, "{")
	file.P("var out privatepb.", privateInName)
	for _, field := range publicIn.Fields {
		delegateField, err := findFieldDelegate(field, privateIn)
		if err != nil {
			return err
		}

		if fieldMatch(field, delegateField) {
			file.P("out.", delegateField.GoName, "=", "in.", field.GoName)
		} else {
			file.P("out.", delegateField.GoName, "=", "c.ToPrivate", funcName(delegateField), "(in.", field.GoName, ")")
		}
	}
	file.P("return &out")
	file.P("}")

	messagesByImportPath[privateInPath][privateInName].Generated = true

	return nil
}

func generateToPublicConverterFunc(file *protogen.GeneratedFile, publicOut, privateOut *protogen.Message) error {
	publicOutName := publicOut.GoIdent.GoName
	publicOutPath := publicOut.GoIdent.GoImportPath

	privateOutName := privateOut.GoIdent.GoName

	file.P("func (c converter) ToPublic", publicOutName, "(in *privatepb.", privateOutName, ") (*publicpb.", publicOutName, ", error) {")
	file.P("var out publicpb.", publicOutName)
	file.P("var err error")

	for _, field := range publicOut.Fields {
		delegateField, err := findFieldDelegate(field, privateOut)
		if err != nil {
			return err
		}

		if fieldMatch(field, delegateField) {
			file.P("out.", field.GoName, "=", "in.", delegateField.GoName)
		} else {
			file.P("out.", field.GoName, ", err =", "c.ToPublic", funcName(field), "(in.", delegateField.GoName, ")")
			file.P("if err != nil { return nil, err }")
		}
	}

	file.P("return &out, err")
	file.P("}")

	messagesByImportPath[publicOutPath][publicOutName].Generated = true

	return nil
}

func generateToPrivateConverterEnumFunc(file *protogen.GeneratedFile, publicIn, privateIn *protogen.Enum) error {
	publicInName := publicIn.GoIdent.GoName

	privateInName := privateIn.GoIdent.GoName
	privateInPath := privateIn.GoIdent.GoImportPath

	file.P("func (c converter) ToPrivate", privateInName, "(in publicpb.", publicInName, ") privatepb.", privateInName, "{")
	file.P("switch in {")
	for _, value := range publicIn.Values {
		delegateValue, err := findEnumValueDelegate(value, privateIn)
		if err != nil {
			return err
		}
		file.P("case publicpb.", value.GoIdent.GoName, ":")
		file.P("return privatepb.", delegateValue.GoIdent.GoName)
	}
	file.P("}")

	if len(privateIn.Values) > 0 {
		file.P("return privatepb.", privateIn.Values[0].GoIdent.GoName)
	}
	file.P("}")

	enumsByImportPath[privateInPath][privateInName].Generated = true

	return nil
}

func generateToPublicConverterEnumFunc(file *protogen.GeneratedFile, publicOut, privateOut *protogen.Enum) error {
	publicOutName := publicOut.GoIdent.GoName
	publicOutPath := publicOut.GoIdent.GoImportPath

	privateOutName := privateOut.GoIdent.GoName

	file.P("func (c converter) ToPublic", publicOutName, "(in privatepb.", privateOutName, ") (publicpb.", publicOutName, ", error) {")
	file.P("switch in {")
	for _, value := range publicOut.Values {
		delegateValue, err := findEnumValueDelegate(value, privateOut)
		if err != nil {
			return err
		}
		file.P("case privatepb.", delegateValue.GoIdent.GoName, ":")
		file.P("return publicpb.", value.GoIdent.GoName, ", nil")

		receiveValues, err := findEnumReceiveValues(value, privateOut)
		if err != nil {
			return err
		}

		for _, receiveValue := range receiveValues {
			if receiveValue.GoIdent.GoName != delegateValue.GoIdent.GoName {
				file.P("case privatepb.", receiveValue.GoIdent.GoName, ":")
				file.P("return publicpb.", value.GoIdent.GoName)
			}
		}
	}
	file.P("}")

	if len(publicOut.Values) > 0 {
		file.P("return publicpb.", publicOut.Values[0].GoIdent.GoName, `, status.Errorf(codes.FailedPrecondition, "unexpected value %q", in)`)
	}
	file.P("}")

	enumsByImportPath[publicOutPath][publicOutName].Generated = true

	return nil
}

func funcName(field *protogen.Field) string {
	if field.Enum != nil {
		return field.Enum.GoIdent.GoName
	}
	return field.GoName
}
