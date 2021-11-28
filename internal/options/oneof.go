package options

import (
	"github.com/dane/protoc-gen-go-svc/gen/svc"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func OneOfName(oneof *protogen.Oneof) string {
	options := oneof.Desc.Options().(*descriptorpb.OneofOptions)
	annotation := proto.GetExtension(options, svc.E_Oneof).(*svc.OneofAnnotation)
	if name := annotation.GetDelegate().GetName(); name != "" {
		return name
	}

	return string(oneof.Desc.Name())
}

func OneOfValidate(oneof *protogen.Oneof) *svc.Validate {
	options := oneof.Desc.Options().(*descriptorpb.OneofOptions)
	annotation := proto.GetExtension(options, svc.E_Oneof).(*svc.OneofAnnotation)
	return &svc.Validate{Required: annotation.GetValidate().GetRequired()}
}

func IsDeprecatedOneOf(oneof *protogen.Oneof) bool {
	options := oneof.Desc.Options().(*descriptorpb.OneofOptions)
	annotation := proto.GetExtension(options, svc.E_Oneof).(*svc.OneofAnnotation)
	return annotation.GetDeprecated()
}

func IsRequiredOneOf(oneof *protogen.Oneof) bool {
	options := oneof.Desc.Options().(*descriptorpb.OneofOptions)
	annotation := proto.GetExtension(options, svc.E_Oneof).(*svc.OneofAnnotation)
	return annotation.GetReceive().GetRequired()
}
