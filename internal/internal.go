package internal

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/dane/protoc-gen-go-svc/internal/generators"
)

func serviceImportPath(service *generators.Service) protogen.GoImportPath {
	prefix := strings.TrimSuffix(string(service.GoIdent.GoImportPath), fmt.Sprintf("/%s", service.GoPackageName))
	path := fmt.Sprintf("%s/service/%s", prefix, service.GoPackageName)
	return protogen.GoImportPath(path)
}
