package options

import (
	"github.com/dane/protoc-gen-go-svc/gen/svc"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func EnumValueName(value *protogen.EnumValue) string {
	options := value.Desc.Options().(*descriptorpb.EnumValueOptions)
	annotation := proto.GetExtension(options, svc.E_EnumValue).(*svc.EnumValueAnnotation)
	if name := annotation.GetDelegate().GetName(); name != "" {
		return name
	}

	return string(value.Desc.Name())
}

func ReceiveEnumValueNames(value *protogen.EnumValue) []string {
	options := value.Desc.Options().(*descriptorpb.EnumValueOptions)
	annotation := proto.GetExtension(options, svc.E_EnumValue).(*svc.EnumValueAnnotation)
	if names := annotation.GetReceive().GetNames(); len(names) > 0 {
		return names
	}

	return []string{EnumValueName(value)}
}
