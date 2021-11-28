package internal

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/dane/protoc-gen-go-svc/internal/options"
)

func NewOneOf(svc *Service, msg *Message, oneof *protogen.Oneof) (*Field, error) {
	f := &Field{
		IsPrivate:    msg.IsPrivate,
		IsLatest:     msg.IsLatest,
		IsOneOf:      true,
		IsRequired:   options.IsRequiredOneOf(oneof),
		IsDeprecated: options.IsDeprecatedOneOf(oneof),
		Name:         oneof.GoName,
		Type:         OneOfType,
	}

	// Assign the private oneof and next oneof. This is only done if the oneof
	// isn't private since the private service, message, fields, etc. are the
	// first in the chain.
	if !f.IsPrivate {
		fieldName := options.OneOfName(oneof)
		var ok bool

		// Latest and deprecated oneofs chain directly to the private service.
		// They don't have a "next" service.
		if !f.IsLatest && !f.IsDeprecated && !msg.IsDeprecated {
			f.Next, ok = msg.Next.FieldByName[fieldName]
			if !ok {
				return nil, NewErrFieldNotFound(fieldName, msg.Next)
			}
		}

		f.Private, ok = msg.Private.FieldByName[fieldName]
		if !ok {
			return nil, NewErrFieldNotFound(fieldName, msg.Private)
		}

		for _, field := range oneof.Fields {
			msg := svc.MessageByName[messageKey(field.Message)]
			f.Messages = append(f.Messages, msg)
		}
	}

	rules, err := NewRules(f, options.OneOfValidate(oneof))
	if err != nil {
		return nil, NewErrCreateField(f, msg, err)
	}

	f.Rules = rules

	return f, nil
}
