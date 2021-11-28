package options

import (
	"github.com/dane/protoc-gen-go-svc/gen/svc"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func GoPackage(file *protogen.File) string {
	options := file.Desc.Options().(*descriptorpb.FileOptions)
	return proto.GetExtension(options, svc.E_GoPackage).(string)
}
