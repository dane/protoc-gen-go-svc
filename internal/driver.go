package internal

import "google.golang.org/protobuf/compiler/protogen"

type Driver interface {
	DelegateMethodName(method *protogen.Method) (string, error)
	DelegateEnumName(enum *protogen.Enum) (string, error)
	DelegateEnumValueName(value *protogen.EnumValue) (string, error)
	DelegateMessageName(message *protogen.Message) (string, error)
	DelegateFieldName(field *protogen.Field) (string, error)
	DelegateOneofName(oneof *protogen.Oneof) (string, error)
	DeprecatedOneof(oneof *protogen.Oneof) bool
	DeprecatedField(field *protogen.Field) bool
	RequiredField(field *protogen.Field) bool
	RequiredOneof(oneof *protogen.Oneof) bool
	DeprecatedMethod(method *protogen.Method) bool
	ValidateMessage(message *protogen.Message) bool
	ValidateField(field *protogen.Field) bool
	ReceiveRequired(field *protogen.Field) bool
	ReceiveEnumValueNames(value *protogen.EnumValue) []string
	Is(field *protogen.Field) (string, error)
	Min(field *protogen.Field) (string, bool, error)
	Max(field *protogen.Field) (string, bool, error)
	In(packageName string, field *protogen.Field) ([]string, error)
}
