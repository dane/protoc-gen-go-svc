package options

import (
	"github.com/dane/protoc-gen-go-svc/gen/svc"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func IsDeprecatedMethod(method *protogen.Method) bool {
	options := method.Desc.Options().(*descriptorpb.MethodOptions)
	annotation := proto.GetExtension(options, svc.E_Method).(*svc.MethodAnnotation)
	return annotation.GetDeprecated()
}

func MethodName(method *protogen.Method) string {
	options := method.Desc.Options().(*descriptorpb.MethodOptions)
	annotation := proto.GetExtension(options, svc.E_Method).(*svc.MethodAnnotation)
	if name := annotation.GetDelegate().GetName(); name != "" {
		return name
	}

	return string(method.Desc.Name())
}

func IsMethodConverterEmpty(method *protogen.Method) bool {
	options := method.Desc.Options().(*descriptorpb.MethodOptions)
	annotation := proto.GetExtension(options, svc.E_Method).(*svc.MethodAnnotation)
	return annotation.GetConverter().GetEmpty()
}
