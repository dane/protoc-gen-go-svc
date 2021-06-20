package internal

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

type Generator struct {
	Verbose bool
}

type Service struct {
	protogen.GoPackageName
	protogen.GoIdent
	*protogen.Service
}

type Message struct {
	*protogen.Message
	Generated bool
	Skip      bool
}

type Enum struct {
	*protogen.Enum
	Generated bool
	Skip      bool
}

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
	messagesByImportPath map[protogen.GoImportPath]map[string]*Message
	enumsByImportPath    map[protogen.GoImportPath]map[string]*Enum
)

func init() {
	messagesByImportPath = make(map[protogen.GoImportPath]map[string]*Message)
	enumsByImportPath = make(map[protogen.GoImportPath]map[string]*Enum)
}

func (g Generator) Run(plugin *protogen.Plugin) error {
	var services []*Service
	var private *Service
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		for _, service := range file.Services {
			if !allowed(service.Comments) {
				continue
			}

			svc := &Service{
				GoIdent:       file.GoDescriptorIdent,
				GoPackageName: file.GoPackageName,
				Service:       service,
			}

			if PrivatePackage == svc.GoPackageName {
				private = svc
			} else {
				services = append(services, svc)
			}
		}

		for _, message := range file.Messages {
			path := message.GoIdent.GoImportPath
			name := message.GoIdent.GoName

			if _, ok := messagesByImportPath[path]; !ok {
				messagesByImportPath[path] = make(map[string]*Message)
			}

			addEnums(message.Enums)

			messagesByImportPath[path][name] = &Message{Message: message}
		}

		addEnums(file.Enums)
	}

	sort.Reverse(byPackageName(services))

	fileName := filepath.Join(ServiceDir, private.PackageName(), ServiceFileName)
	file := plugin.NewGeneratedFile(fileName, private.GoImportPath)
	if err := generatePrivateService(file, private); err != nil {
		return err
	}

	if len(services) > 0 {
		latest := services[0]
		fileName := filepath.Join(ServiceDir, latest.PackageName(), ServiceFileName)
		file := plugin.NewGeneratedFile(fileName, latest.GoImportPath)
		if err := generateLatestService(file, latest, private); err != nil {
			return err
		}
	}

	fileName = filepath.Join(ServiceDir, ServiceFileName)
	file = plugin.NewGeneratedFile(fileName, "")
	if err := generateRegister(file, services, private); err != nil {
		return err
	}

	return nil
}

func allowed(commentSet protogen.CommentSet) bool {
	return len(mergeComments(commentSet)) > 0
}

func mergeComments(commentSet protogen.CommentSet) []string {
	var annotations []string
	comments := strings.Split(string(commentSet.Leading), "\n")
	for _, comment := range commentSet.LeadingDetached {
		comments = append(comments, string(comment))
	}

	for _, comment := range comments {
		if strings.HasPrefix(comment, GenSvc) {
			annotations = append(annotations, comment)
		}
	}

	return annotations
}

type byPackageName []*Service

func (s byPackageName) Len() int {
	return len(s)
}

func (s byPackageName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byPackageName) Less(i, j int) bool {
	return s[i].GoPackageName < s[j].GoPackageName
}

func addEnums(enums []*protogen.Enum) {
	for _, enum := range enums {
		path := enum.GoIdent.GoImportPath
		name := enum.GoIdent.GoName

		if _, ok := enumsByImportPath[path]; !ok {
			enumsByImportPath[path] = make(map[string]*Enum)
		}

		enumsByImportPath[path][name] = &Enum{Enum: enum}
	}
}
