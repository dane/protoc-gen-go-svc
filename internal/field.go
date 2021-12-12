package internal

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/dane/protoc-gen-go-svc/internal/options"
)

type Field struct {
	IsPrivate       bool
	IsLatest        bool
	IsDeprecated    bool
	IsMessage       bool
	IsEnum          bool
	IsOneOf         bool
	IsMatch         bool
	IsRepeated      bool
	IsRequired      bool
	Name            string
	EnumName        string
	Type            Type
	Private         *Field
	Next            *Field
	Message         *Message
	Messages        []*Message
	EnumValues      []*EnumValue
	EnumValueByName map[string]*EnumValue
	Rules           []string
}

// NewField creates a `Field`. An error will be returned if the field cannot be
// created for any reason.
func NewField(svc *Service, msg *Message, field *protogen.Field) (*Field, error) {
	f := &Field{
		IsPrivate:       msg.IsPrivate,
		IsLatest:        msg.IsLatest,
		IsEnum:          (field.Enum != nil),
		IsMessage:       (field.Message != nil),
		IsRepeated:      field.Desc.IsList(),
		IsRequired:      options.IsRequiredField(field),
		IsDeprecated:    options.IsDeprecatedField(field),
		Name:            field.GoName,
		EnumValueByName: make(map[string]*EnumValue),
	}

	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		f.Type = BooleanType
	case protoreflect.Int64Kind:
		f.Type = Int64Type
	case protoreflect.Uint64Kind:
		f.Type = Uint64Type
	case protoreflect.DoubleKind:
		f.Type = Float64Type
	case protoreflect.StringKind:
		f.Type = StringType
	case protoreflect.BytesKind:
		f.Type = BytesType
	case protoreflect.MessageKind:
		f.Type = MessageType
	case protoreflect.EnumKind:
		f.Type = EnumType
	}

	// Assign the message that is populating the field. Messages are assigned
	// before the field is fully populated to allow the `isMatch` check to run.
	if f.IsMessage {
		// This guard protects from external messages that may not have been
		// added to the map.
		var ok bool
		f.Message, ok = svc.MessageByName[messageKey(field.Message)]
		if !ok {
			return nil, NewErrMessageNotFound(messageKey(field.Message), svc)
		}
	}

	// Assign the private field and next field. This is only done if the field
	// isn't private since the private service, message, fields, etc. are the
	// first in the chain.
	if !f.IsPrivate {
		fieldName := options.FieldName(field)
		var ok bool

		// Latest and deprecated fields chain directly to the private service.
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

		if f.IsLatest || f.IsDeprecated || msg.IsDeprecated {
			f.IsMatch = isMatch(f, f.Private)
		} else {
			f.IsMatch = isMatch(f, f.Next)
		}
	}

	// Enums are created after the private and next fields are assigned. This
	// makes it easier to find the enum values.
	if f.IsEnum {
		f.EnumName = field.Enum.GoIdent.GoName
		for _, value := range field.Enum.Values {
			v, err := NewEnumValue(f, value)
			if err != nil {
				return nil, NewErrCreateField(f, msg, err)
			}

			f.EnumValues = append(f.EnumValues, v)
			f.EnumValueByName[enumValueKey(value)] = v
		}
	}

	rules, err := NewRules(f, options.FieldValidate(field))
	if err != nil {
		return nil, NewErrCreateField(f, msg, err)
	}

	f.Rules = rules

	return f, nil
}

// isExternalFieldMessage checks if a field represents a message that is from an
// external package. An external package is not the private package or one of
// the public packages.
func isExternalFieldMessage(svc *Service, field *protogen.Field) bool {
	if field.Message == nil {
		return false
	}

	return isExternalMessage(svc, field.Message)
}

func isExternalMessage(svc *Service, message *protogen.Message) bool {
	// Cannot rely on proto package names because messages don't have access to
	// the proto package name.
	return svc.ImportPath != string(message.GoIdent.GoImportPath)
}

func isMatch(a, b *Field) bool {
	// Types must match in both fields.
	if a.Type != b.Type {
		return false
	}

	// Never assume enum or oneof comparisions are equal to each other.
	if a.Type == EnumType || a.Type == OneOfType {
		return false
	}

	// Types match and aren't messages, they're ints, floats, strings, bools,
	// etc.
	if a.Type != MessageType {
		return true
	}

	return isMessageMatch(a.Message, b.Message)
}
