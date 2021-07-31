package internal

import (
	"fmt"
	"sort"

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

func findNextEnum(enum *protogen.Enum, next *Service) (*protogen.Enum, error) {
	enumName := enum.GoIdent.GoName
	name, err := delegateEnumName(enum)
	if err != nil {
		return nil, fmt.Errorf("failed to find delegate enum for %s: %w", enumName, err)
	}

	if name == "" {
		name = enumName
	}

	for _, enum := range next.Enums {
		if name == enum.GoIdent.GoName {
			return enum, nil
		}
	}

	return nil, fmt.Errorf("failed to find next enum for %s", enumName)
}

func findNextEnumValue(value *protogen.EnumValue, nextEnum *protogen.Enum) (*protogen.EnumValue, error) {
	enumValueName := string(value.Desc.Name())
	name, err := delegateEnumValueName(value)
	if err != nil {
		return nil, fmt.Errorf("failed to find delegate enum value for %s: %w", enumValueName, err)
	}

	if name == "" {
		name = enumValueName
	}

	for _, nextValue := range nextEnum.Values {
		if name == string(nextValue.Desc.Name()) {
			return nextValue, nil
		}
	}

	return nil, fmt.Errorf("failed to find next enum value for %s", enumValueName)
}

func findReceiveEnumValues(value *protogen.EnumValue, nextEnum *protogen.Enum) ([]*protogen.EnumValue, error) {
	var values []*protogen.EnumValue
	nextEnumName := nextEnum.GoIdent.GoName
	for _, name := range receiveEnumValueNames(value) {
		var matched bool
		for _, nextValue := range nextEnum.Values {
			if name == string(nextValue.Desc.Name()) {
				values = append(values, nextValue)
				matched = true
			}
		}
		if !matched {
			return nil, fmt.Errorf("failed to find %s in enum %s", name, nextEnumName)
		}
	}

	return values, nil
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

func findNextOneof(oneof *protogen.Oneof, next *protogen.Message) (*protogen.Oneof, error) {
	oneofName := string(oneof.Desc.Name())
	name, err := delegateOneofName(oneof)
	if err != nil {
		return nil, fmt.Errorf("failed to find delegate oneof for %s: %w", oneofName, err)
	}

	if name == "" {
		name = oneofName
	}

	for _, oneof := range next.Oneofs {
		if name == string(oneof.Desc.Name()) {
			return oneof, nil
		}
	}

	return nil, fmt.Errorf("failed to find next oneof for %s", oneofName)
}

func findNextOneofField(field *protogen.Field, next *protogen.Oneof) (*protogen.Field, error) {
	fieldName := string(field.Desc.Name())
	name, err := delegateFieldName(field)
	if err != nil {
		return nil, fmt.Errorf("failed to find delegate field for %s: %w", fieldName, err)
	}

	if name == "" {
		name = fieldName
	}

	for _, field := range next.Fields {
		if name == string(field.Desc.Name()) {
			return field, nil
		}
	}

	return nil, fmt.Errorf("failed to find next field for %s", fieldName)
}

func findNextField(field *protogen.Field, next *protogen.Message) (*protogen.Field, error) {
	fieldName := string(field.Desc.Name())
	name, err := delegateFieldName(field)
	if err != nil {
		return nil, fmt.Errorf("failed to find delegate field for %s: %w", fieldName, err)
	}

	if name == "" {
		name = fieldName
	}

	for _, field := range next.Fields {
		if name == string(field.Desc.Name()) {
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
		return "int64", nil
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

func findPrivateEnum(enum *protogen.Enum, chain []*Service) (*protogen.Enum, error) {
	targetEnum := enum
	var err error
	for _, next := range chain {
		targetEnum, err = findNextEnum(targetEnum, next)
		if err != nil {
			return nil, err
		}
	}
	return targetEnum, nil
}

func findPrivateEnumValue(value *protogen.EnumValue, enum *protogen.Enum, chain []*Service) (*protogen.EnumValue, error) {
	targetValue := value
	targetEnum := enum
	var err error

	for _, next := range chain {
		targetEnum, err = findNextEnum(targetEnum, next)
		if err != nil {
			return nil, err
		}

		targetValue, err = findNextEnumValue(targetValue, targetEnum)
		if err != nil {
			return nil, err
		}
	}

	return targetValue, nil
}

func findPrivateReceiveEnumValues(value *protogen.EnumValue, enum *protogen.Enum, chain []*Service) ([]*protogen.EnumValue, error) {
	privateEnum, err := findPrivateEnum(enum, chain)
	if err != nil {
		return nil, err
	}

	var values []*protogen.EnumValue
	privateEnumName := privateEnum.GoIdent.GoName
	for _, name := range receiveEnumValueNames(value) {
		var matched bool
		for _, privateValue := range privateEnum.Values {
			if name == string(privateValue.Desc.Name()) {
				values = append(values, privateValue)
				matched = true
			}
		}
		if !matched {
			return nil, fmt.Errorf("failed to find %s in enum %s", name, privateEnumName)
		}
	}

	return values, nil
}

func findPrivateOneof(oneof *protogen.Oneof, message *protogen.Message, chain []*Service) (*protogen.Oneof, error) {
	targetMessage := message
	targetOneof := oneof

	var err error
	for _, next := range chain {
		targetMessage, err = findNextMessage(targetMessage, next)
		if err != nil {
			return nil, err
		}

		targetOneof, err = findNextOneof(targetOneof, targetMessage)
		if err != nil {
			return nil, err
		}
	}
	return targetOneof, nil
}

func findPrivateOneofField(field *protogen.Field, oneof *protogen.Oneof, message *protogen.Message, chain []*Service) (*protogen.Field, error) {
	targetMessage := message
	targetOneof := oneof
	targetField := field

	var err error
	for _, next := range chain {
		targetMessage, err = findNextMessage(targetMessage, next)
		if err != nil {
			return nil, err
		}

		targetOneof, err = findNextOneof(targetOneof, targetMessage)
		if err != nil {
			return nil, err
		}

		targetField, err = findNextOneofField(targetField, targetOneof)
		if err != nil {
			return nil, err
		}
	}
	return targetField, nil
}

func findPrivateField(field *protogen.Field, message *protogen.Message, chain []*Service) (*protogen.Field, error) {
	targetMessage := message
	targetField := field

	var err error
	for _, next := range chain {
		targetMessage, err = findNextMessage(targetMessage, next)
		if err != nil {
			return nil, err
		}

		targetField, err = findNextField(targetField, targetMessage)
		if err != nil {
			return nil, err
		}
	}
	return targetField, nil
}

func newDeprecatedFinder(message *protogen.Message) deprecatedFinder {
	finder := deprecatedFinder{
		message:      message,
		goImportPath: message.GoIdent.GoImportPath,
		messages:     make(map[*protogen.Message]struct{}),
		enums:        make(map[*protogen.Enum]struct{}),
	}

	finder.build(message)
	return finder
}

type deprecatedFinder struct {
	message      *protogen.Message
	goImportPath protogen.GoImportPath
	messages     map[*protogen.Message]struct{}
	enums        map[*protogen.Enum]struct{}
}

func (d deprecatedFinder) Messages() []*protogen.Message {
	return d.sortedMessages()
}

func (d deprecatedFinder) Enums() []*protogen.Enum {
	return d.sortedEnums()
}

func (d deprecatedFinder) build(message *protogen.Message) {
	for _, field := range message.Fields {
		if field.Enum != nil {
			if _, ok := d.enums[field.Enum]; !ok {
				d.enums[field.Enum] = struct{}{}
			}
		}

		if field.Message == nil {
			continue
		}

		if field.Message.GoIdent.GoImportPath != d.goImportPath {
			continue
		}

		if _, ok := d.messages[field.Message]; ok {
			continue
		}

		d.messages[field.Message] = struct{}{}
		d.build(field.Message)
	}
}

func (d deprecatedFinder) sortedMessages() []*protogen.Message {
	var messages []*protogen.Message
	for message, _ := range d.messages {
		messages = append(messages, message)
	}

	sort.Sort(byMessageName(messages))
	return messages
}

func (d deprecatedFinder) sortedEnums() []*protogen.Enum {
	var enums []*protogen.Enum
	for enum, _ := range d.enums {
		enums = append(enums, enum)
	}

	sort.Sort(byEnumName(enums))
	return enums
}

type byMessageName []*protogen.Message

func (s byMessageName) Len() int {
	return len(s)
}

func (s byMessageName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byMessageName) Less(i, j int) bool {
	return s[i].GoIdent.GoName < s[j].GoIdent.GoName
}

type byEnumName []*protogen.Enum

func (s byEnumName) Len() int {
	return len(s)
}

func (s byEnumName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byEnumName) Less(i, j int) bool {
	return s[i].GoIdent.GoName < s[j].GoIdent.GoName
}
