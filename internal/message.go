package internal

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/dane/protoc-gen-go-svc/internal/options"
)

type Message struct {
	IsPrivate        bool
	IsLatest         bool
	IsDeprecated     bool
	IsExternal       bool
	IsOneOf          bool
	IsConverterEmpty bool
	Name             string
	ImportPath       string
	PackageName      string
	FullName         string
	Private          *Message
	Next             *Message
	Parent           *Message
	Fields           []*Field
	FieldByName      map[string]*Field
}

func (m *Message) Type() string {
	if m.IsExternal {
		return fmt.Sprintf("%s.%s", m.PackageName, m.Name)
	}

	if m.IsPrivate {
		return fmt.Sprintf("privatepb.%s", m.Name)
	}

	return fmt.Sprintf("publicpb.%s", m.Name)
}

func (m *Message) NextType() string {
	if m.Next == nil {
		return ""
	}

	if m.Next.IsExternal {
		return fmt.Sprintf("%s.%s", m.Next.PackageName, m.Next.Name)
	}

	return fmt.Sprintf("nextpb.%s", m.Next.Name)
}

func (m *Message) PrivateType() string {
	if m.Private == nil {
		return ""
	}

	if m.Private.IsExternal {
		return fmt.Sprintf("%s.%s", m.Private.PackageName, m.Private.Name)
	}

	return fmt.Sprintf("privatepb.%s", m.Private.Name)
}

func (m *Message) Ref() string {
	if m.IsExternal {
		return fmt.Sprintf("External%s", m.Name)
	}

	return m.Name
}

// NewMessage creates a `Message`. An error will be returned if the message
// cannot be created for any reason.
func NewMessage(svc *Service, message, parent *protogen.Message) (*Message, error) {
	var p *Message
	var ok bool

	if parent != nil {
		if p, ok = svc.MessageByName[messageKey(parent)]; !ok {
			return nil, NewErrMessageNotFound(messageKey(parent), svc)
		}
	}

	msg := &Message{
		IsPrivate:        svc.IsPrivate,
		IsLatest:         svc.IsLatest,
		IsDeprecated:     options.IsDeprecatedMessage(message),
		IsOneOf:          len(message.Oneofs) > 0,
		IsConverterEmpty: options.IsConverterEmpty(message),
		ImportPath:       svc.ImportPath,
		Name:             message.GoIdent.GoName,
		FieldByName:      make(map[string]*Field),
		Parent:           p,
		FullName:         string(message.Desc.FullName()),
	}

	// Private messages are the last in the service chain.
	if msg.IsPrivate {
		return msg, nil
	}

	messageName := options.MessageName(message)

	// Messages of the latest service or deprecated messages read/write directly
	// to the private service.
	if msg.IsLatest || msg.IsDeprecated {
		if msg.Parent == nil {
			messageName = buildMessageKey(svc.Private, messageName)
		} else {
			messageName = fmt.Sprintf("%s.%s", msg.Parent.Private.FullName, messageName)
		}

		msg.Private, ok = svc.Private.MessageByName[messageName]
		if !ok {
			return nil, NewErrMessageNotFound(messageName, svc.Private)
		}

		return msg, nil
	}

	// All other messages will chain to a message in the next service version.
	if msg.Parent == nil {
		messageName = buildMessageKey(svc.Next, messageName)
	} else {
		messageName = fmt.Sprintf("%s.%s", msg.Parent.Next.FullName, messageName)
	}
	msg.Next, ok = svc.Next.MessageByName[messageName]
	if !ok {
		return nil, NewErrMessageNotFound(messageName, svc.Next)
	}

	msg.Private = msg.Next.Private

	return msg, nil
}

// NewExternalMessage creates a `Message` for protobuf messages that are
// external to the public and private services. These are placeholder structures
// to make building validators and converters easier.
func NewExternalMessage(svc *Service, message *protogen.Message) (*Message, error) {
	msg := &Message{
		IsExternal: true,
		IsLatest:   svc.IsLatest,
		IsPrivate:  svc.IsPrivate,
		ImportPath: string(message.GoIdent.GoImportPath),
		Name:       message.GoIdent.GoName,
	}

	importPath := strings.Split(msg.ImportPath, "/")
	msg.PackageName = fmt.Sprintf("ext%s", importPath[len(importPath)-1])

	return msg
}
