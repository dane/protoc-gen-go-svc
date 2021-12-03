package internal

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/dane/protoc-gen-go-svc/internal/options"
)

type Method struct {
	IsPrivate    bool
	IsLatest     bool
	IsDeprecated bool
	Name         string
	Private      *Method
	Next         *Method
	Input        *Message
	Output       *Message
}

// NewMethod creates a `Method`. An error will be returned if the method
// cannot be created for any reason.
func NewMethod(svc *Service, method *protogen.Method) (*Method, error) {
	m := &Method{
		IsPrivate:    svc.IsPrivate,
		IsLatest:     svc.IsLatest,
		IsDeprecated: options.IsDeprecatedMethod(method),
		Name:         method.GoName,
	}

	var ok bool
	m.Input, ok = svc.MessageByName[messageKey(method.Input)]
	if !ok {
		return nil, NewErrMessageNotFound(messageKey(method.Input), svc)
	}

	m.Output, ok = svc.MessageByName[messageKey(method.Output)]
	if !ok {
		return nil, NewErrMessageNotFound(messageKey(method.Output), svc)
	}

	// Private methods are the last in the service chain.
	if m.IsPrivate {
		return m, nil
	}

	methodName := options.MethodName(method)

	// Methods of the latest service or deprecated methods chain directly to the
	// private service.
	if m.IsLatest || m.IsDeprecated {
		m.Private, ok = svc.Private.MethodByName[methodName]
		if !ok {
			return nil, NewErrMethodNotFound(methodName, svc.Private)
		}

		return m, nil
	}

	// All other methods will chain to a methods in the next service version.
	m.Next, ok = svc.Next.MethodByName[methodName]
	if !ok {
		return nil, NewErrMethodNotFound(methodName, svc.Next)
	}

	m.Private = m.Next.Private

	return m, nil
}
