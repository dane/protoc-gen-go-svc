package options

import (
	"github.com/dane/protoc-gen-go-svc/gen/svc"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func FieldName(field *protogen.Field) string {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	if name := annotation.GetDelegate().GetName(); name != "" {
		return name
	}

	return string(field.Desc.Name())
}

func FieldValidate(field *protogen.Field) *svc.Validate {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	return annotation.GetValidate()
}

func IsDeprecatedField(field *protogen.Field) bool {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	return annotation.GetDeprecated()
}

func IsRequiredField(field *protogen.Field) bool {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	annotation := proto.GetExtension(options, svc.E_Field).(*svc.FieldAnnotation)
	return annotation.GetReceive().GetRequired()
}
