package internal

import (
	"path"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Service struct {
	IsPrivate            bool
	IsLatest             bool
	ProtoPackageName     string
	PackageName          string
	ImportPath           string
	ServiceImportPath    string
	SubServiceImportPath string
	GeneratedFileName    string
	Name                 string
	Private              *Service
	Next                 *Service
	Messages             []*Message
	MessageByName        map[string]*Message
	Methods              []*Method
	MethodByName         map[string]*Method
}

// NewService creates a `Service`. An error will be returned if the service
// cannot be created for any reason.
func NewService(
	protoPackageName protoreflect.FullName,
	packageName protogen.GoPackageName,
	importPath protogen.GoImportPath,
	serviceImportPath protogen.GoImportPath,
	service *protogen.Service,
	messages []*protogen.Message,
	serviceChain []*Service,
) (*Service, error) {
	svc := &Service{
		IsPrivate:            len(serviceChain) == 0,
		IsLatest:             len(serviceChain) == 1,
		ProtoPackageName:     string(protoPackageName),
		PackageName:          string(packageName),
		ImportPath:           string(importPath),
		ServiceImportPath:    string(serviceImportPath),
		SubServiceImportPath: path.Join(string(serviceImportPath), string(packageName)),
		Name:                 service.GoName,
		MessageByName:        make(map[string]*Message),
		MethodByName:         make(map[string]*Method),
	}

	// The private service is the first entry in the chain. If the chain has a
	// length of 0, this function is constructing the private service.
	if len(serviceChain) >= 1 {
		svc.Private = serviceChain[0]
	}

	// Public services are appended to the chain. If there is more than one
	// entry in the chain, the last entry is the next service. For example:
	// - 1: private service
	// - 2: v2 service
	// - 3: v1 service
	//
	// Entry 1 and 2 would not have a next service, but entry 3 would.
	if len(serviceChain) >= 2 {
		svc.Next = serviceChain[len(serviceChain)-1]
	}

	// Create messages. Fields are created after all messages have been created
	// because oneofs and will reference messages.
	for _, message := range messages {
		msg, err := NewMessage(svc, message)
		if err != nil {
			return nil, NewErrCreateService(svc, err)
		}

		svc.Messages = append(svc.Messages, msg)
		svc.MessageByName[messageKey(message)] = msg
	}

	// Iterate through messages again to ensure all messages are present that a
	// field may reference.
	for _, message := range messages {
		for _, field := range message.Fields {
			// If the field references a message from an external package, most
			// likely `google.protobuf.Timestamp` or `Any`, build an "external"
			// message. This will be leveraged when comparing fields or building
			// validations and converters.
			if isExternalFieldMessage(svc, field) {
				if _, ok := svc.MessageByName[messageKey(field.Message)]; !ok {
					ext := NewExternalMessage(field.Message)
					svc.MessageByName[messageKey(field.Message)] = ext
					svc.Messages = append(svc.Messages, ext)
				}
			}

			// Skip fields that are part of a oneof. They will be constructed
			// later in the file. It's easier to create a oneof from the message
			// struct.
			if field.Oneof != nil {
				continue
			}

			// All package messages have been created already. There is no need
			// to guard against the message not being present.
			msg := svc.MessageByName[messageKey(message)]

			f, err := NewField(svc, msg, field)
			if err != nil {
				return nil, NewErrCreateService(svc, err)
			}

			msg.Fields = append(msg.Fields, f)
			msg.FieldByName[fieldKey(field)] = f
		}

		for _, oneof := range message.Oneofs {
			// All package messages have been created already. There is no need
			// to guard against the message not being present.
			msg := svc.MessageByName[messageKey(message)]

			f, err := NewOneOf(svc, msg, oneof)
			if err != nil {
				return nil, NewErrCreateService(svc, err)
			}

			msg.Fields = append(msg.Fields, f)
			msg.FieldByName[oneOfKey(oneof)] = f
		}
	}

	// Create methods. All messages will be present at this point.
	for _, method := range service.Methods {
		m, err := NewMethod(svc, method)
		if err != nil {
			return nil, NewErrCreateService(svc, err)
		}

		svc.Methods = append(svc.Methods, m)
		svc.MethodByName[methodKey(method)] = m
	}

	return svc, nil
}
