package internal

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

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

func serviceImportPath(service *Service) protogen.GoImportPath {
	prefix := strings.TrimSuffix(string(service.GoIdent.GoImportPath), fmt.Sprintf("/%s", service.GoPackageName))
	path := fmt.Sprintf("%s/service/%s", prefix, service.GoPackageName)
	return protogen.GoImportPath(path)
}
