package internal

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/dane/protoc-gen-go-svc/internal/options"
)

type Message struct {
	IsPrivate    bool
	IsLatest     bool
	IsDeprecated bool
	IsExternal   bool
	IsOneOf      bool
	//IsUsed       bool
	Name        string
	ImportPath  string
	PackageName string
	Private     *Message
	Next        *Message
	Fields      []*Field
	FieldByName map[string]*Field
}

func NewMessage(svc *Service, message *protogen.Message) (*Message, error) {
	msg := &Message{
		IsPrivate:    svc.IsPrivate,
		IsLatest:     svc.IsLatest,
		IsDeprecated: options.IsDeprecatedMessage(message),
		IsOneOf:      len(message.Oneofs) > 0,
		ImportPath:   svc.ImportPath,
		Name:         message.GoIdent.GoName,
		FieldByName:  make(map[string]*Field),
	}

	if msg.IsPrivate {
		return msg, nil
	}

	messageName := options.MessageName(message)
	var ok bool

	if msg.IsLatest || msg.IsDeprecated {
		msg.Private, ok = svc.Private.MessageByName[buildMesageKey(svc.Private, messageName)]
		if !ok {
			return nil, NewErrMessageNotFound(messageName, svc.Private)
		}

		return msg, nil
	}

	msg.Next, ok = svc.Next.MessageByName[buildMesageKey(svc.Next, messageName)]
	if !ok {
		return nil, NewErrMessageNotFound(messageName, svc.Next)
	}

	msg.Private = msg.Next.Private

	return msg, nil
}

func NewExternalMessage(message *protogen.Message) *Message {
	msg := &Message{
		IsExternal: true,
		ImportPath: string(message.GoIdent.GoImportPath),
		Name:       message.GoIdent.GoName,
	}

	importPath := strings.Split(msg.ImportPath, "/")
	msg.PackageName = fmt.Sprintf("ext%s", importPath[len(importPath)-1])

	return msg
}
