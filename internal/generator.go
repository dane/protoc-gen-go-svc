package internal

import (
	"fmt"
	"path/filepath"
	"sort"

	"google.golang.org/protobuf/compiler/protogen"
)

type Generator struct {
	Verbose bool
}

type Service struct {
	protogen.GoPackageName
	protogen.GoIdent
	*protogen.Service

	GoName              string
	GoServiceImportPath protogen.GoImportPath
	Messages            []*protogen.Message
}

type ServiceType int

const (
	PublicService ServiceType = iota
	LatestPublicService
	PrivateService
)

func (s *Service) PackageName() string {
	return fmt.Sprintf("%s", s.GoPackageName)
}

const (
	// GenSvc is the annotation prefix for protoc-gen-go-svc.
	GenSvc = "gen:svc"

	// PrivatePackage is the package name of the private gRPC service.
	PrivatePackage protogen.GoPackageName = "private"

	// ServiceFileName is the file name of all generated services.
	ServiceFileName = "service.go"

	// ServiceDir is the directory where generated services are stored within
	// the defined "go-svc_out" destination.
	ServiceDir = "service"
)

var (
	inputs  map[*protogen.Message]struct{}
	outputs map[*protogen.Message]struct{}
)

func init() {
	inputs = make(map[*protogen.Message]struct{})
	outputs = make(map[*protogen.Message]struct{})
}

func (g Generator) Run(plugin *protogen.Plugin) error {
	var services []*Service
	var private *Service
	messages := make(map[protogen.GoImportPath]map[string]*protogen.Message)

	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		for _, service := range file.Services {
			svc := &Service{
				GoIdent:       file.GoDescriptorIdent,
				GoPackageName: file.GoPackageName,
				GoName:        service.GoName,
				Service:       service,
			}
			svc.GoServiceImportPath = serviceImportPath(svc)

			if PrivatePackage == svc.GoPackageName {
				private = svc
			} else {
				services = append(services, svc)
			}

			for _, method := range service.Methods {
				inputs[method.Input] = struct{}{}
				outputs[method.Output] = struct{}{}
			}
		}

		// Group messages by import path to set on each service.
		for _, message := range file.Messages {
			importPath := message.GoIdent.GoImportPath
			messageName := message.GoIdent.GoName

			if _, ok := messages[importPath]; !ok {
				messages[importPath] = make(map[string]*protogen.Message)
			}

			messages[importPath][messageName] = message
		}
	}

	sort.Sort(byPackageName(services))
	services = append(services, private)

	// Set messages on service.
	for _, service := range services {
		for _, message := range messages[service.GoImportPath] {
			service.Messages = append(service.Messages, message)
		}
	}

	serviceLen := len(services)
	for i, service := range services {
		fileName := filepath.Join(ServiceDir, service.PackageName(), ServiceFileName)
		file := plugin.NewGeneratedFile(fileName, service.GoImportPath)
		chain := services[i+1:]

		var err error
		switch i {
		// Generate the private service.
		case serviceLen - 1:
			err = generatePrivateService(file, service)
			// Generate the latest service.
		case serviceLen - 2:
			err = generateLatestPublicService(file, service, chain)
			// Generate all other service versions.
		default:
			err = generatePublicService(file, service, chain)
		}

		if err != nil {
			return fmt.Errorf("failed to generate %s service: %w", service.GoPackageName, err)
		}
	}

	return nil
}
