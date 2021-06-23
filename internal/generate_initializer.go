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

	file.P("type Option interface {")
	file.P("Name() string")
	file.P("}")

	file.P("func RegisterServer(s *grpc.Server, impl privatepb.", private.Service.GoName, "Server, options ...Option) {")
	file.P("privateSvc := &private.Service{")
	file.P("Validator: private.NewValidator(),")
	file.P("Impl:      impl,")
	file.P("}")

	for i, service := range services {
		varName := fmt.Sprintf("%sSvc", service.GoPackageName)
		file.P(varName, " := &", service.GoPackageName, ".Service{")
		file.P("Validator:", service.GoPackageName, ".NewValidator(),")
		file.P("Converter:", service.GoPackageName, ".NewConverter(),")
		file.P("Private: privateSvc,")
		if i > 0 {
			nextService := services[i-1]
			file.P("Next: ", nextService.GoPackageName, "Svc,")
		}
		file.P("}")

		file.P("for _, opt := range options {")
		file.P("if ", service.GoPackageName, ".ConverterName == opt.Name() {")
		file.P(varName, ".Converter = opt.(", service.GoPackageName, ".Converter)")
		file.P("} else if ", service.GoPackageName, ".ValidatorName == opt.Name() {")
		file.P(varName, ".Validator = opt.(", service.GoPackageName, ".Validator)")
		file.P("}")
		file.P("}")
		file.P(service.GoPackageName, "pb.Register", service.Service.GoName, "Server(s, ", varName, ")")
	}
	file.P("}")

	return nil
}

func serviceImportPath(service *Service) protogen.GoImportPath {
	prefix := strings.TrimSuffix(string(service.GoIdent.GoImportPath), fmt.Sprintf("/%s", service.GoPackageName))
	path := fmt.Sprintf("%s/service/%s", prefix, service.GoPackageName)
	return protogen.GoImportPath(path)
}
