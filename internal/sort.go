package internal

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/dane/protoc-gen-go-svc/internal/generators"
)

// byPackageName sort services by package name
type byPackageName []*generators.Service

func (s byPackageName) Len() int {
	return len(s)
}

func (s byPackageName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byPackageName) Less(i, j int) bool {
	return s[i].GoPackageName < s[j].GoPackageName
}

// byMethodName sort methods by name.
type byMethodName []*protogen.Method

func (s byMethodName) Len() int {
	return len(s)
}

func (s byMethodName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byMethodName) Less(i, j int) bool {
	return s[i].GoName < s[j].GoName
}

// byMessageName sort messages by name.
type byMessageName []*protogen.Message

func (s byMessageName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byMessageName) Less(i, j int) bool {
	return s[i].GoIdent.GoName < s[j].GoIdent.GoName
}

func (s byMessageName) Len() int {
	return len(s)
}

// byFieldNumber sort fields by position number.
type byFieldNumber []*protogen.Field

func (s byFieldNumber) Len() int {
	return len(s)
}

func (s byFieldNumber) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byFieldNumber) Less(i, j int) bool {
	return s[i].Desc.Number() < s[j].Desc.Number()
}

// byEnumName sort enums by name.
type byEnumName []*protogen.Enum

func (s byEnumName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byEnumName) Less(i, j int) bool {
	return s[i].GoIdent.GoName < s[j].GoIdent.GoName
}

func (s byEnumName) Len() int {
	return len(s)
}

// byEnumValueNumber sort enum values by position number.
type byEnumValueNumber []*protogen.EnumValue

func (s byEnumValueNumber) Len() int {
	return len(s)
}

func (s byEnumValueNumber) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byEnumValueNumber) Less(i, j int) bool {
	return s[i].Desc.Number() < s[j].Desc.Number()
}

// byOneofName sort oneofs by name.
type byOneofName []*protogen.Oneof

func (s byOneofName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byOneofName) Less(i, j int) bool {
	return s[i].GoIdent.GoName < s[j].GoIdent.GoName
}

func (s byOneofName) Len() int {
	return len(s)
}
