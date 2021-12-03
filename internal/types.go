package internal

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

type Type int

const (
	Undefined Type = iota
	StringType
	Int64Type
	Uint64Type
	Float64Type
	BooleanType
	BytesType
	MessageType
	EnumType
	OneOfType
)

func methodKey(method *protogen.Method) string {
	return string(method.Desc.Name())
}

func buildMessageKey(svc *Service, messageName string) string {
	return fmt.Sprintf("%s.%s", svc.ProtoPackageName, messageName)
}

func messageKey(message *protogen.Message) string {
	return string(message.Desc.FullName())
}

func fieldKey(field *protogen.Field) string {
	return string(field.Desc.Name())
}

func enumValueKey(value *protogen.EnumValue) string {
	return string(value.Desc.Name())
}

func oneOfKey(oneof *protogen.Oneof) string {
	return string(oneof.Desc.Name())
}
