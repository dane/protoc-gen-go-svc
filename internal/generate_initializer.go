package internal

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func generateRegister(file *protogen.GeneratedFile, services []*Service, private *Service) error {
	packageGrpc := protogen.GoImportPath("google.golang.org/grpc")

	file.P("package service")
	file.P("import (")
	file.P("privatepb", private.GoIdent.GoImportPath)
	file.P(private.GoPackageName, serviceImportPath(private))
	file.P("grpc", packageGrpc)
	for _, service := range services {
		file.P(service.GoPackageName, serviceImportPath(service))
		file.P(service.GoPackageName, "pb", service.GoIdent.GoImportPath)
	}
	file.P(")")

	file.P("func RegisterServer(s *grpc.Server, impl privatepb.", private.Service.GoName, "Server) {")
	file.P("privateSvc := &private.Service{")
	file.P("Validator: private.NewValidator(),")
	file.P("Impl:      impl,")
	file.P("}")

	for i, service := range services {
		file.P(service.GoPackageName, "Svc := &", service.GoPackageName, ".Service{")
		file.P("Validator:", service.GoPackageName, ".NewValidator(),")
		file.P("Converter:", service.GoPackageName, ".NewConverter(),")
		file.P("Private: privateSvc,")
		if i > 0 {
			nextService := services[i-1]
			file.P("Converter:", service.GoPackageName, ".NewConverter(),")
			file.P("Next: ", nextService.GoPackageName, "Svc,")
		}
		file.P("}")
		file.P(service.GoPackageName, "pb.Register", service.Service.GoName, "Server(s, ", service.GoPackageName, "Svc)")
	}
	file.P("}")

	return nil
}

func serviceImportPath(service *Service) protogen.GoImportPath {
	prefix := strings.TrimSuffix(string(service.GoIdent.GoImportPath), fmt.Sprintf("/%s", service.GoPackageName))
	path := fmt.Sprintf("%s/service/%s", prefix, service.GoPackageName)
	return protogen.GoImportPath(path)
}

/*
package service

import (
	private "github.com/dane/protoc-gen-go-svc/example/proto/go/service/private"
	v1 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v1"
	v2 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v2"
)

func RegisterServer(s *grpc.Server, impl privatepb.PeopleServer) {
	privateSvc := &private.Service{
		Validator: private.NewValidator(),
		Impl:      impl,
	}

	v2Svc := &v2.Service{
		Validator: v2.NewValidator(),
		Private:   privateSvc,
		Next:      privateSvc,
	}
	v2pb.RegisterPeopleServer(s, v2Svc)
	v1Svc := &v1.Service{
		Validator: v1.NewValidator(),
		Private:   privateSvc,
		Next:      v2Svc,
	}
	v1pb.RegisterPeopleServer(s, v1Svc)
}

*/
