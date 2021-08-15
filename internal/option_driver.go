package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dane/protoc-gen-go-svc/gen/svc"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type optionDriver struct {
	inputs  map[*protogen.Message]struct{}
	outputs map[*protogen.Message]struct{}
}

func NewOptionDriver(inputs, outputs map[*protogen.Message]struct{}) Driver {
	return optionDriver{
		inputs:  inputs,
		outputs: outputs,
	}
}

func (o optionDriver) FileGoPackage(file *protogen.File) (string, error) {
	options := file.Desc.Options().(*descriptorpb.FileOptions)
	annotation := proto.GetExtension(options, svc.E_GoPackage).(string)
	if annotation == "" {
		return "", fmt.Errorf("option (gen.svc.go_package) not defined in file %q", file.Desc.Path())
	}

	return annotation, nil
}

func (o optionDriver) DelegateMethodName(method *protogen.Method) (string, error) {
	options := method.Desc.Options().(*descriptorpb.MethodOptions)
	annotation := proto.GetExtension(options, svc.E_Method).(*svc.MethodAnnotation)
	return annotation.GetDelegate().GetName(), nil
}

func (o optionDriver) DelegateEnumName(enum *protogen.Enum) (string, error) {
	options := enum.Desc.Options().(*descriptorpb.EnumOptions)
	annotation := proto.GetExtension(options, svc.E_Enum).(*svc.EnumAnnotation)
	return annotation.GetDelegate().GetName(), nil
}

func (o optionDriver) DelegateEnumValueName(value *protogen.EnumValue) (string, error) {
	options := value.Desc.Options().(*descriptorpb.EnumValueOptions)
	annotation := proto.GetExtension(options, svc.E_EnumValue).(*svc.EnumValueAnnotation)
	return annotation.GetDelegate().GetName(), nil
}

func (o optionDriver) DelegateMessageName(message *protogen.Message) (string, error) {
	options := message.Desc.Options().(*descriptorpb.MessageOptions)
	annotation := proto.GetExtension(options, svc.E_Message).(*svc.MessageAnnotation)
	return annotation.GetDelegate().GetName(), nil
}

func (o optionDriver) DelegateFieldName(field *protogen.Field) (string, error) {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	return annotation.GetDelegate().GetName(), nil
}

func (o optionDriver) DelegateOneofName(oneof *protogen.Oneof) (string, error) {
	options := oneof.Desc.Options().(*descriptorpb.OneofOptions)
	annotation := proto.GetExtension(options, svc.E_Oneof).(*svc.OneofAnnotation)
	return annotation.GetDelegate().GetName(), nil
}

func (o optionDriver) DeprecatedField(field *protogen.Field) bool {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	return annotation.GetDeprecated()
}

func (o optionDriver) DeprecatedOneof(oneof *protogen.Oneof) bool {
	options := oneof.Desc.Options().(*descriptorpb.OneofOptions)
	annotation := proto.GetExtension(options, svc.E_Oneof).(*svc.OneofAnnotation)
	return annotation.GetDeprecated()
}

func (o optionDriver) DeprecatedMethod(method *protogen.Method) bool {
	options := method.Desc.Options().(*descriptorpb.MethodOptions)
	annotation := proto.GetExtension(options, svc.E_Method).(*svc.MethodAnnotation)
	return annotation.GetDeprecated()
}

func (o optionDriver) RequiredField(field *protogen.Field) bool {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	return annotation.GetValidate().GetRequired()
}

func (o optionDriver) RequiredOneof(Oneof *protogen.Oneof) bool {
	options := Oneof.Desc.Options().(*descriptorpb.OneofOptions)
	annotation := proto.GetExtension(options, svc.E_Oneof).(*svc.OneofAnnotation)
	return annotation.GetValidate().GetRequired()
}

func (o optionDriver) ValidateMessage(message *protogen.Message) bool {
	if _, ok := o.inputs[message]; ok {
		return true
	}

	for _, field := range message.Fields {
		if o.ValidateField(field) {
			return true
		}

		if field.Oneof != nil {
			for _, field := range field.Oneof.Fields {
				return o.ValidateMessage(field.Message)
			}
		}
	}

	return false
}

func (o optionDriver) ValidateField(field *protogen.Field) bool {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	return annotation.GetValidate() != nil
}

func (o optionDriver) ReceiveRequired(field *protogen.Field) bool {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	return annotation.GetReceive().GetRequired()
}

func (o optionDriver) ReceiveEnumValueNames(value *protogen.EnumValue) []string {
	options := value.Desc.Options().(*descriptorpb.EnumValueOptions)
	annotation := proto.GetExtension(options, svc.E_EnumValue).(*svc.EnumValueAnnotation)
	return annotation.GetReceive().GetNames()
}

func (o optionDriver) Min(field *protogen.Field) (string, bool, error) {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	return o.number(field, annotation.GetValidate().GetMin())
}

func (o optionDriver) Max(field *protogen.Field) (string, bool, error) {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	return o.number(field, annotation.GetValidate().GetMax())
}

func (o optionDriver) In(packageName string, field *protogen.Field) ([]string, error) {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	values := annotation.GetValidate().GetIn()
	if values == nil {
		return nil, nil
	}

	if o.builtins(values) {
		return values, nil
	}

	if field.Enum != nil {
		for i, value := range values {
			var matched bool
			enumName := field.Enum.GoIdent.GoName
			for _, ev := range field.Enum.Values {
				if string(ev.Desc.Name()) == value {
					valueName := ev.GoIdent.GoName
					values[i] = fmt.Sprintf("%s.%s", packageName, valueName)
					matched = true
				}
			}

			if !matched {
				return nil, fmt.Errorf("invalid value %s for enum %s", value, enumName)
			}
		}

		return values, nil
	}

	return nil, nil
}

func (o optionDriver) Is(field *protogen.Field) (string, error) {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	if annotation.GetValidate().GetIs() != svc.Validate_UNSPECIFIED {
		return annotation.GetValidate().GetIs().String(), nil
	}

	return "", nil
}

func (o optionDriver) number(field *protogen.Field, num *svc.Number) (string, bool, error) {
	isSet := num != nil

	var value string
	switch field.Desc.Kind() {
	case protoreflect.Int64Kind, protoreflect.StringKind:
		value = fmt.Sprintf("%d", num.GetInt64())
	case protoreflect.Uint64Kind:
		value = fmt.Sprintf("%d", num.GetUint64())
	case protoreflect.DoubleKind:
		value = fmt.Sprintf("%f", num.GetDouble())
	}

	return value, isSet, nil
}

func (o optionDriver) builtins(values []string) bool {
	var bools, ints, floats, strs int
	for _, value := range values {
		if o.contains([]string{"true", "false"}, value) {
			bools++
		}

		if len(value) >= 2 && strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			strs++
		}

		if _, err := strconv.ParseInt(value, 10, 64); err == nil {
			ints++
		}

		if _, err := strconv.ParseFloat(value, 64); err == nil {
			floats++
		}
	}

	count := len(values)
	if bools == count || ints == count || floats == count || strs == count {
		return true
	}
	return false
}

func (o optionDriver) contains(values []string, v string) bool {
	for _, value := range values {
		if v == value {
			return true
		}
	}
	return false
}
