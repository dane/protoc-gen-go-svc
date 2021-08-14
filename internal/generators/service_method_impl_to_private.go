package generators

import "io"

func NewServiceMethodImplToPrivate(prefix, methodName, inputName, outputName, privateMethodName, privteInputName, privateOutputName string) Generator {
	return GeneratorFunc(func(w io.Writer) error {
		return execute("service_method_impl_to_private", templateServiceMethodImplToPrivate, w, serviceMethodImplToPrivate{
			Prefix:            prefix,
			MethodName:        methodName,
			InputName:         inputName,
			OutputName:        outputName,
			PrivateMethodName: privateMethodName,
			PrivateInputName:  privteInputName,
			PrivateOutputName: privateOutputName,
		})
	})
}

type serviceMethodImplToPrivate struct {
	Prefix            string
	MethodName        string
	InputName         string
	OutputName        string
	PrivateMethodName string
	PrivateInputName  string
	PrivateOutputName string
}
