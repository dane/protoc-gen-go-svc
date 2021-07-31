package internal

import "google.golang.org/protobuf/compiler/protogen"

func delegateMethodName(method *protogen.Method) (string, error) {
	return driver.DelegateMethodName(method)
}

func delegateEnumName(enum *protogen.Enum) (string, error) {
	return driver.DelegateEnumName(enum)
}

func delegateEnumValueName(value *protogen.EnumValue) (string, error) {
	return driver.DelegateEnumValueName(value)
}

func delegateMessageName(message *protogen.Message) (string, error) {
	return driver.DelegateMessageName(message)
}

func delegateFieldName(field *protogen.Field) (string, error) {
	return driver.DelegateFieldName(field)
}

func delegateOneofName(oneof *protogen.Oneof) (string, error) {
	return driver.DelegateOneofName(oneof)
}

func deprecatedOneof(oneof *protogen.Oneof) bool {
	return driver.DeprecatedOneof(oneof)
}

func deprecatedField(field *protogen.Field) bool {
	return driver.DeprecatedField(field)
}

func deprecatedMethod(method *protogen.Method) bool {
	return driver.DeprecatedMethod(method)
}

func validateMessage(message *protogen.Message) bool {
	return driver.ValidateMessage(message)
}

func validateField(field *protogen.Field) bool {
	return driver.ValidateField(field)
}

func receiveRequired(field *protogen.Field) bool {
	return driver.ReceiveRequired(field)
}

func receiveEnumValueNames(value *protogen.EnumValue) []string {
	return driver.ReceiveEnumValueNames(value)
}

func is(field *protogen.Field) (string, error) {
	return driver.Is(field)
}

func min(field *protogen.Field) (string, bool, error) {
	return driver.Min(field)
}

func max(field *protogen.Field) (string, bool, error) {
	return driver.Max(field)
}

func in(packageName string, field *protogen.Field) ([]string, error) {
	return driver.In(packageName, field)
}

func requiredField(field *protogen.Field) bool {
	return driver.RequiredField(field)
}

func requiredOneof(oneof *protogen.Oneof) bool {
	return driver.RequiredOneof(oneof)
}
