package options

import (
	"github.com/dane/protoc-gen-go-svc/gen/svc"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func MessageName(message *protogen.Message) string {
	options := message.Desc.Options().(*descriptorpb.MessageOptions)
	annotation := proto.GetExtension(options, svc.E_Message).(*svc.MessageAnnotation)
	if name := annotation.GetDelegate().GetName(); name != "" {
		return name
	}

	return string(message.Desc.Name())
}

func IsDeprecatedMessage(message *protogen.Message) bool {
	options := message.Desc.Options().(*descriptorpb.MessageOptions)
	annotation := proto.GetExtension(options, svc.E_Message).(*svc.MessageAnnotation)
	return annotation.GetDeprecated()
}
