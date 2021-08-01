package internal

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func serviceImportPath(service *Service) protogen.GoImportPath {
	prefix := strings.TrimSuffix(string(service.GoIdent.GoImportPath), fmt.Sprintf("/%s", service.GoPackageName))
	path := fmt.Sprintf("%s/service/%s", prefix, service.GoPackageName)
	return protogen.GoImportPath(path)
}
