package internal

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func findNextMethod(method *protogen.Method, next *Service) (*protogen.Method, error) {
	methodName := method.GoName
	name, err := delegateMethodName(method)
	if err != nil {
		return nil, fmt.Errorf("failed to find delegate method for %s: %w", methodName, err)
	}

	if name == "" {
		name = methodName
	}

	for _, method := range next.Methods {
		if name == method.GoName {
			return method, nil
		}
	}

	return nil, fmt.Errorf("failed to find next method for %s", methodName)
}

func findNextMessage(message *protogen.Message, next *Service) (*protogen.Message, error) {
	messageName := message.GoIdent.GoName
	name, err := delegateMessageName(message)
	if err != nil {
		return nil, fmt.Errorf("failed to find delegate message for %s: %w", messageName, err)
	}

	if name == "" {
		name = messageName
	}

	for _, message := range next.Messages {
		if name == message.GoIdent.GoName {
			return message, nil
		}
	}

	return nil, fmt.Errorf("failed to find next message for %s", messageName)
}

func findNextField(field *protogen.Field, next *protogen.Message) (*protogen.Field, error) {
	fieldName := field.GoIdent.GoName
	name, err := delegateFieldName(field)
	if err != nil {
		return nil, fmt.Errorf("failed to find delegate field for %s: %w", fieldName, err)
	}

	if name == "" {
		name = fieldName
	}

	for _, field := range next.Fields {
		if name == field.GoIdent.GoName {
			return field, nil
		}
	}

	return nil, fmt.Errorf("failed to find next field for %s", fieldName)
}

func findFieldType(packageName string, field *protogen.Field) (string, error) {
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return "string", nil
	case protoreflect.BoolKind:
		return "bool", nil
	case protoreflect.Int64Kind:
		return "in64", nil
	case protoreflect.FloatKind:
		return "float64", nil
	case protoreflect.EnumKind:
		enumName := field.Enum.GoIdent.GoName
		return fmt.Sprintf("%s.%s", packageName, enumName), nil
	case protoreflect.MessageKind:
		messageName := field.Message.GoIdent.GoName
		return fmt.Sprintf("*%s.%s", packageName, messageName), nil
	}

	return "", fmt.Errorf("unexpected field %s for type lookup", field.GoName)
}

func findPrivateMethod(method *protogen.Method, chain []*Service) (*protogen.Method, error) {
	targetMethod := method
	var err error
	for _, next := range chain {
		targetMethod, err = findNextMethod(targetMethod, next)
		if err != nil {
			return nil, err
		}
	}
	return targetMethod, nil
}

func findPrivateMessage(message *protogen.Message, chain []*Service) (*protogen.Message, error) {
	targetMessage := message
	var err error
	for _, next := range chain {
		targetMessage, err = findNextMessage(targetMessage, next)
		if err != nil {
			return nil, err
		}
	}
	return targetMessage, nil
}
