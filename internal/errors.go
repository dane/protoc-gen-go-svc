package internal

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

func NewErrCreateService(svc *Service, err error) error {
	return fmt.Errorf("failed to create service %s of package %s: %w", svc.Name, svc.ProtoPackageName, err)
}

func NewErrCreateField(f *Field, msg *Message, err error) error {
	return fmt.Errorf("failed to create field %s of message %s: %w", f.Name, msg.Name, err)
}

func NewErrMessageNotFound(messageName string, svc *Service) error {
	return fmt.Errorf("failed to find message %s in package %s", messageName, svc.ProtoPackageName)
}

func NewErrMethodNotFound(methodName string, svc *Service) error {
	return fmt.Errorf("failed to find method %s in service of package %s", methodName, svc.ProtoPackageName)
}

func NewErrFieldNotFound(fieldName string, msg *Message) error {
	return fmt.Errorf("failed to find field %s in message %s", fieldName, msg.Name)
}

func NewErrEnumValueNotFound(enumValueName string, f *Field) error {
	return fmt.Errorf("failed to find enum value %s in field %s", enumValueName, f.Name)
}

func NewErrBadServiceImportPath(file *protogen.File, importPath, fileImportPath string) error {
	return fmt.Errorf("file %s has service import path of %q, but expected %q", file.Desc.Path(), fileImportPath, importPath)
}

func NewErrInvalidServiceImportPath(file *protogen.File, importPath string) error {
	return fmt.Errorf("file %s has an invalid service import path %s", file.Desc.Path(), importPath)
}

func NewErrInvalidRuleIn(f *Field, value string) error {
	return fmt.Errorf("invalid value %q in `in` annotation in field %s", value, f.Name)
}

func NewErrInvalidRuleForField(f *Field, ruleName string) error {
	return fmt.Errorf("invalid rule %q for field %s", ruleName, f.Name)
}
