package internal

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/dane/protoc-gen-go-svc/internal/generators"
)

type Generator struct {
	Verbose bool
}

type ServiceType int

const (
	PublicService ServiceType = iota
	LatestPublicService
	PrivateService
)

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
	driver  Driver
	logger  *log.Logger
)

func init() {
	inputs = make(map[*protogen.Message]struct{})
	outputs = make(map[*protogen.Message]struct{})
	driver = NewOptionDriver(inputs, outputs)

}

func (g *Generator) Run(plugin *protogen.Plugin) error {
	var dst io.Writer = ioutil.Discard
	if g.Verbose {
		dst = os.Stderr
	}
	logger = log.New(dst, "", 0)

	var services []*generators.Service
	var private *generators.Service
	messages := make(map[protogen.GoImportPath]map[string]*protogen.Message)
	enums := make(map[protogen.GoImportPath]map[string]*protogen.Enum)

	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		for _, service := range file.Services {
			svc := &generators.Service{
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

				if deprecatedMethod(method) {
					svc.DeprecatedMessages = append(svc.DeprecatedMessages, method.Output)
					finder := newDeprecatedFinder(method.Output)
					for _, message := range finder.Messages() {
						if !containsMessage(message, svc.DeprecatedMessages) {
							svc.DeprecatedMessages = append(svc.DeprecatedMessages, message)
						}
					}

					for _, enum := range finder.Enums() {
						if !containsEnum(enum, svc.DeprecatedEnums) {
							svc.DeprecatedEnums = append(svc.DeprecatedEnums, enum)
						}
					}
				}
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

			// Group enums defined within messages by import path to set on each
			// service.
			for _, enum := range message.Enums {
				enumName := enum.GoIdent.GoName
				if _, ok := enums[importPath]; !ok {
					enums[importPath] = make(map[string]*protogen.Enum)
				}

				enums[importPath][enumName] = enum
			}
		}

		// Group enums by import path to set on each service.
		for _, enum := range file.Enums {
			importPath := enum.GoIdent.GoImportPath
			enumName := enum.GoIdent.GoName
			if _, ok := enums[importPath]; !ok {
				enums[importPath] = make(map[string]*protogen.Enum)
			}

			enums[importPath][enumName] = enum
		}
	}

	sort.Sort(byPackageName(services))
	services = append(services, private)

	for _, service := range services {
		// Set messages on service.
		for _, message := range messages[service.GoImportPath] {
			service.Messages = append(service.Messages, message)
		}

		// Set enums on service.
		for _, enum := range enums[service.GoImportPath] {
			service.Enums = append(service.Enums, enum)
		}
	}

	// Sort methods, messages, enums, fields and oneofs
	for _, service := range services {
		sort.Sort(byMethodName(service.Methods))
		sort.Sort(byMessageName(service.Messages))
		sort.Sort(byEnumName(service.Enums))

		for _, message := range service.Messages {
			sort.Sort(byFieldNumber(message.Fields))
			sort.Sort(byOneofName(message.Oneofs))

			for _, oneof := range message.Oneofs {
				sort.Sort(byFieldNumber(oneof.Fields))
			}
		}
	}

	serviceLen := len(services)
	for i, service := range services {
		logger.Printf("package=%s at=generate-service", service.GoPackageName)
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

	fileName := filepath.Join(ServiceDir, ServiceFileName)
	file := plugin.NewGeneratedFile(fileName, "")
	if err := generateServiceRegister(file, services); err != nil {
		return err
	}

	return nil
}

func containsMessage(message *protogen.Message, messages []*protogen.Message) bool {
	for _, exists := range messages {
		if message == exists {
			return true
		}
	}
	return false
}

func containsEnum(enum *protogen.Enum, enums []*protogen.Enum) bool {
	for _, exists := range enums {
		if enum == exists {
			return true
		}
	}
	return false
}
