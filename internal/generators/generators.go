package generators

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

type Service struct {
	protogen.GoPackageName
	protogen.GoIdent
	*protogen.Service

	GoName              string
	GoServiceImportPath protogen.GoImportPath
	Messages            []*protogen.Message
	Enums               []*protogen.Enum
	DeprecatedMessages  []*protogen.Message
	DeprecatedEnums     []*protogen.Enum
}

func (s *Service) PackageName() string {
	return fmt.Sprintf("%s", s.GoPackageName)
}
