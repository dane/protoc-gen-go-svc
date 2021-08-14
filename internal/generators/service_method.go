package generators

import (
	"io"
)

func NewServiceMethod(packageName, methodName, inputName, outputName string, toPrivate bool) Generator {
	return GeneratorFunc(func(w io.Writer) error {
		return execute("service_method", templateServiceMethod, w, serviceMethod{
			PackageName: packageName,
			MethodName:  methodName,
			InputName:   inputName,
			OutputName:  outputName,
			ToPrivate:   toPrivate,
		})
	})
}

type serviceMethod struct {
	MethodName  string
	InputName   string
	OutputName  string
	PackageName string
	ToPrivate   bool
}
