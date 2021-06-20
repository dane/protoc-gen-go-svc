package internal

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// findMethodDelegate finds the next method in the service chain.
func findMethodDelegate(method *protogen.Method, next *Service) (*protogen.Method, error) {
	targetName := method.GoName
	delegateName, err := delegateAnnotation(method.Comments)
	if err != nil {
		return nil, err
	}

	if delegateName != "" {
		targetName = delegateName
	}

	for _, method := range next.Methods {
		if method.GoName == targetName {
			return method, nil
		}
	}

	return nil, fmt.Errorf("failed to find delegate method %q in service package %q", targetName, next.GoPackageName)
}

// findMessageDelegate finds the next message in the service chain.
func findMessageDelegate(message *Message, next *Service) (*Message, error) {
	targetName := message.GoIdent.GoName
	delegateName, err := delegateAnnotation(message.Comments)
	if err != nil {
		return nil, err
	}

	if delegateName != "" {
		targetName = delegateName
	}

	if m, ok := messagesByImportPath[next.GoIdent.GoImportPath][targetName]; ok {
		return m, nil
	}

	return nil, fmt.Errorf("failed to find delegate message %q in service package %q", targetName, next.GoPackageName)
}

// findFieldDelegate finds the next field in the message chain.
func findFieldDelegate(field *protogen.Field, nextMessage *protogen.Message) (*protogen.Field, error) {
	targetName := field.GoName
	delegateName, err := delegateAnnotation(field.Comments)
	if err != nil {
		return nil, err
	}

	if delegateName != "" {
		targetName = delegateName
	}

	for _, nextField := range nextMessage.Fields {
		if nextField.GoName == targetName {
			return nextField, nil
		}
	}

	return nil, fmt.Errorf("failed to find delegate field %q in message %q", targetName, nextMessage.GoIdent.GoName)
}

// findEnumDelegate finds the next enum in the service chain.
func findEnumDelegate(enum *Enum, next *Service) (*Enum, error) {
	targetName := enum.GoIdent.GoName
	delegateName, err := delegateAnnotation(enum.Comments)
	if err != nil {
		return nil, err
	}

	if delegateName != "" {
		targetName = delegateName
	}

	if e, ok := enumsByImportPath[next.GoIdent.GoImportPath][targetName]; ok {
		return e, nil
	}

	return nil, fmt.Errorf("failed to find delegate enum %q in service package %q", targetName, next.GoPackageName)
}

// findEnumValueDelegate finds the next value in the enum chain.
func findEnumValueDelegate(value *protogen.EnumValue, next *protogen.Enum) (*protogen.EnumValue, error) {
	targetName := value.Desc.Name()
	delegateName, err := delegateAnnotation(value.Comments)
	if err != nil {
		return nil, err
	}

	if delegateName != "" {
		targetName = protoreflect.Name(delegateName)
	}

	for _, value := range next.Values {
		if targetName == value.Desc.Name() {
			return value, nil
		}
	}

	return nil, fmt.Errorf("failed to find delegate enum value %q in enum %q", targetName, next.Desc.Name())
}

// findEnumReceiveValues finds the next values in the enum chain that can
// populate the current value.
func findEnumReceiveValues(value *protogen.EnumValue, next *protogen.Enum) ([]*protogen.EnumValue, error) {
	receiveNames, err := receiveAnnotations(value.Comments)
	if err != nil {
		return nil, err
	}

	var values []*protogen.EnumValue
	for _, targetName := range receiveNames {
		for _, value := range next.Values {
			if protoreflect.Name(targetName) == value.Desc.Name() {
				values = append(values, value)
				continue
			}

			return nil, fmt.Errorf("failed to find receive enum value %q in enum %q", targetName, next.Desc.Name())
		}
	}

	return values, nil
}
